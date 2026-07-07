package sas

import (
	"strings"
	"testing"
)

const clientID = "kitchen-bridge"

// exchange runs an honest client/server SAS agreement and returns both sides'
// derived (sas, key) pairs.
func exchange(t *testing.T) (clientSAS, serverSAS string, clientKey, serverKey []byte) {
	t.Helper()

	clientPriv, err := GenerateKey()
	if err != nil {
		t.Fatalf("client key: %v", err)
	}
	serverPriv, err := GenerateKey()
	if err != nil {
		t.Fatalf("server key: %v", err)
	}
	pka := clientPriv.PublicKey().Bytes()
	pkb := serverPriv.PublicKey().Bytes()

	na, err := Nonce()
	if err != nil {
		t.Fatalf("client nonce: %v", err)
	}
	nb, err := Nonce()
	if err != nil {
		t.Fatalf("server nonce: %v", err)
	}

	// Server commits to nb; client must be able to verify it once revealed.
	commitment := Commit(pkb, pka, clientID, nb)
	if !VerifyCommit(commitment, pkb, pka, clientID, nb) {
		t.Fatalf("honest commitment did not verify")
	}

	serverPubForClient, err := ParsePublicKey(pkb)
	if err != nil {
		t.Fatalf("parse pkb: %v", err)
	}
	clientPubForServer, err := ParsePublicKey(pka)
	if err != nil {
		t.Fatalf("parse pka: %v", err)
	}

	clientSAS, clientKey, err = Derive(clientPriv, serverPubForClient, pka, pkb, clientID, na, nb)
	if err != nil {
		t.Fatalf("client derive: %v", err)
	}
	serverSAS, serverKey, err = Derive(serverPriv, clientPubForServer, pka, pkb, clientID, na, nb)
	if err != nil {
		t.Fatalf("server derive: %v", err)
	}
	return clientSAS, serverSAS, clientKey, serverKey
}

func TestHonestExchangeAgrees(t *testing.T) {
	clientSAS, serverSAS, clientKey, serverKey := exchange(t)

	if clientSAS != serverSAS {
		t.Errorf("SAS mismatch on honest exchange: client=%q server=%q", clientSAS, serverSAS)
	}
	if len(clientSAS) != Digits {
		t.Errorf("SAS length = %d, want %d (%q)", len(clientSAS), Digits, clientSAS)
	}
	if strings.Trim(clientSAS, "0123456789") != "" {
		t.Errorf("SAS contains non-digits: %q", clientSAS)
	}
	if len(clientKey) != KeySize {
		t.Errorf("key length = %d, want %d", len(clientKey), KeySize)
	}
	if string(clientKey) != string(serverKey) {
		t.Errorf("session key mismatch on honest exchange")
	}
}

func TestTamperedCommitmentRejected(t *testing.T) {
	clientPriv, _ := GenerateKey()
	serverPriv, _ := GenerateKey()
	pka := clientPriv.PublicKey().Bytes()
	pkb := serverPriv.PublicKey().Bytes()
	nb, _ := Nonce()

	commitment := Commit(pkb, pka, clientID, nb)

	// Flip a bit in the revealed nonce: verification must fail.
	badNb := append([]byte(nil), nb...)
	badNb[0] ^= 0x01
	if VerifyCommit(commitment, pkb, pka, clientID, badNb) {
		t.Errorf("commitment verified against tampered nonce")
	}
	// A different client_id must also fail (client_id is bound in).
	if VerifyCommit(commitment, pkb, pka, "attacker", nb) {
		t.Errorf("commitment verified against wrong client_id")
	}
}

// TestSimulatedMITM models an attacker relaying between client and server with
// its own ephemeral key on each leg. The two legs derive independent SAS values
// which, overwhelmingly, do not match — so the user rejects the pairing.
func TestSimulatedMITM(t *testing.T) {
	// Honest client <-> attacker leg.
	clientPriv, _ := GenerateKey()
	mitmToClientPriv, _ := GenerateKey()
	pkaClient := clientPriv.PublicKey().Bytes()
	pkbMitm := mitmToClientPriv.PublicKey().Bytes()
	naClient, _ := Nonce()
	nbMitm, _ := Nonce()

	// Attacker <-> honest server leg.
	mitmToServerPriv, _ := GenerateKey()
	serverPriv, _ := GenerateKey()
	pkaMitm := mitmToServerPriv.PublicKey().Bytes()
	pkbServer := serverPriv.PublicKey().Bytes()
	naMitm, _ := Nonce()
	nbServer, _ := Nonce()

	serverPubForClient, _ := ParsePublicKey(pkbMitm)
	clientLegSAS, _, err := Derive(clientPriv, serverPubForClient, pkaClient, pkbMitm, clientID, naClient, nbMitm)
	if err != nil {
		t.Fatalf("client leg derive: %v", err)
	}

	clientPubForServer, _ := ParsePublicKey(pkaMitm)
	serverLegSAS, _, err := Derive(serverPriv, clientPubForServer, pkaMitm, pkbServer, clientID, naMitm, nbServer)
	if err != nil {
		t.Fatalf("server leg derive: %v", err)
	}

	if clientLegSAS == serverLegSAS {
		t.Errorf("MITM legs produced matching SAS (%q) — attacker would go undetected", clientLegSAS)
	}
}

func TestDeriveDeterministic(t *testing.T) {
	clientPriv, _ := GenerateKey()
	serverPriv, _ := GenerateKey()
	pka := clientPriv.PublicKey().Bytes()
	pkb := serverPriv.PublicKey().Bytes()
	na, _ := Nonce()
	nb, _ := Nonce()

	serverPub, _ := ParsePublicKey(pkb)
	sas1, key1, err := Derive(clientPriv, serverPub, pka, pkb, clientID, na, nb)
	if err != nil {
		t.Fatalf("derive 1: %v", err)
	}
	sas2, key2, err := Derive(clientPriv, serverPub, pka, pkb, clientID, na, nb)
	if err != nil {
		t.Fatalf("derive 2: %v", err)
	}
	if sas1 != sas2 || string(key1) != string(key2) {
		t.Errorf("Derive not deterministic")
	}
}

func TestGrouped(t *testing.T) {
	if got := Grouped("12345678"); got != "1234 5678" {
		t.Errorf("Grouped = %q, want %q", got, "1234 5678")
	}
	if got := Grouped("123"); got != "123" {
		t.Errorf("Grouped of short string = %q, want passthrough", got)
	}
}
