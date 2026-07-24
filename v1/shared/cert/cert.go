package cert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

func DecodePrivKey(priv []byte) (*rsa.PrivateKey, error) {
	privPEM, _ := pem.Decode(priv)
	if privPEM.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("not an RSA private key")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(privPEM.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	return privKey, nil
}

func GeneratePrivKey() (*rsa.PrivateKey, error) {
	// Generate a new private key.
	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

func EncodePrivKey(privKey *rsa.PrivateKey) ([]byte, error) {
	f := bytes.NewBuffer(nil)

	// Encode the private key.
	err := pem.Encode(f, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})
	if err != nil {
		return nil, err
	}
	return f.Bytes(), nil
}

func GenerateSelfSignedCert(privKey *rsa.PrivateKey) ([]byte, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	// Set up our server certificate.
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"woodhouse"},
		},
		DNSNames:              []string{"woodhouse"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1000, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func EncodeCert(cert []byte) ([]byte, error) {
	f := bytes.NewBuffer(nil)

	// Encode the private key.
	err := pem.Encode(f, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	if err != nil {
		return nil, err
	}
	return f.Bytes(), nil
}
