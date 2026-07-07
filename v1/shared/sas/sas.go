// Package sas implements a commitment-based Short Authentication String (SAS)
// key agreement used to authenticate the woodhouse pairing bootstrap.
//
// The pairing transport is unauthenticated (TLS with InsecureSkipVerify), so
// the channel is authenticated instead by a human comparing a short code shown
// on both the client CLI and the server web UI. An active man-in-the-middle
// running two independent legs cannot force both codes to match: the server
// commits to its nonce before learning the client's nonce, so neither leg can
// steer the derived code, and the codes collide only with probability ~10^-8.
//
// Flow (see AuthService.Pair):
//  1. Client sends client_id + ephemeral X25519 public key PKa.
//  2. Server sends PKb and Commit(PKb, PKa, clientID, Nb).
//  3. Client sends its nonce Na.
//  4. Server reveals Nb.
//  5. Client verifies the commitment; both call Derive to obtain the same SAS
//     (compared by the user) and AES-256 session key (used to deliver the cert
//     and refresh token once the user confirms).
package sas

import (
	"crypto/ecdh"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// NonceSize is the length of the SAS nonces in bytes.
	NonceSize = 32
	// KeySize is the derived AES-256 session key length in bytes.
	KeySize = 32
	// Digits is the number of decimal digits in the SAS.
	Digits = 8

	sasModulus = 100_000_000 // 10^Digits
	labelSAS   = "woodhouse-pairing-sas"
	labelKey   = "woodhouse-pairing-key"
)

// curve is the ephemeral key-agreement curve.
func curve() ecdh.Curve { return ecdh.X25519() }

// GenerateKey returns a fresh ephemeral X25519 key pair.
func GenerateKey() (*ecdh.PrivateKey, error) {
	return curve().GenerateKey(rand.Reader)
}

// ParsePublicKey parses and validates a peer's X25519 public key. It rejects
// malformed keys; ECDH additionally rejects low-order points.
func ParsePublicKey(b []byte) (*ecdh.PublicKey, error) {
	return curve().NewPublicKey(b)
}

// Nonce returns a cryptographically random NonceSize-byte nonce.
func Nonce() ([]byte, error) {
	n := make([]byte, NonceSize)
	if _, err := rand.Read(n); err != nil {
		return nil, err
	}
	return n, nil
}

// transcript returns a length-delimited concatenation of the fields so that no
// combination of field contents can be confused with another.
func transcript(fields ...[]byte) []byte {
	var b []byte
	var l [4]byte
	for _, f := range fields {
		binary.BigEndian.PutUint32(l[:], uint32(len(f)))
		b = append(b, l[:]...)
		b = append(b, f...)
	}
	return b
}

// Commit returns SHA-256 over a length-delimited transcript binding the server
// public key, client public key, client id and the server nonce nb. The server
// sends this before revealing nb, committing it and preventing an adaptive
// choice of nonce once the client's nonce is known.
func Commit(serverPub, clientPub []byte, clientID string, nb []byte) []byte {
	sum := sha256.Sum256(transcript(serverPub, clientPub, []byte(clientID), nb))
	return sum[:]
}

// VerifyCommit reports whether commitment matches Commit(...) in constant time.
func VerifyCommit(commitment, serverPub, clientPub []byte, clientID string, nb []byte) bool {
	return subtle.ConstantTimeCompare(commitment, Commit(serverPub, clientPub, clientID, nb)) == 1
}

// Derive computes the shared ECDH secret and derives the SAS and session key.
// clientPub/serverPub are PKa/PKb, clientNonce/serverNonce are Na/Nb, all bound
// into the derivation transcript along with clientID. Both parties call this
// with identical inputs and obtain the same SAS and key.
func Derive(priv *ecdh.PrivateKey, peer *ecdh.PublicKey, clientPub, serverPub []byte, clientID string, clientNonce, serverNonce []byte) (sas string, key []byte, err error) {
	shared, err := priv.ECDH(peer)
	if err != nil {
		return "", nil, err
	}
	// Defence in depth: X25519 ECDH already rejects low-order points, but never
	// proceed with an all-zero shared secret.
	if isZero(shared) {
		return "", nil, errors.New("sas: degenerate shared secret")
	}

	// Bind the full transcript into every derived output; distinct labels give
	// domain separation between the SAS and the session key.
	t := transcript(clientPub, serverPub, []byte(clientID), clientNonce, serverNonce)
	salt := append(append(make([]byte, 0, len(clientNonce)+len(serverNonce)), clientNonce...), serverNonce...)

	sasBytes, err := hkdf.Key(sha256.New, shared, salt, labelSAS+string(t), 8)
	if err != nil {
		return "", nil, err
	}
	key, err = hkdf.Key(sha256.New, shared, salt, labelKey+string(t), KeySize)
	if err != nil {
		return "", nil, err
	}

	n := binary.BigEndian.Uint64(sasBytes) % sasModulus
	return fmt.Sprintf("%0*d", Digits, n), key, nil
}

// Grouped renders a SAS for display as two space-separated 4-digit groups,
// e.g. "1234 5678", which is easier to compare by eye.
func Grouped(sas string) string {
	if len(sas) != Digits {
		return sas
	}
	return sas[:4] + " " + sas[4:]
}

func isZero(b []byte) bool {
	var v byte
	for _, x := range b {
		v |= x
	}
	return v == 0
}
