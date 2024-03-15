package cert

import (
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"os"

	"github.com/jimjibone/woodhouse-4/log"
	"github.com/jimjibone/woodhouse-4/shared/paths"
)

type CertManager struct {
	certPath string
	keyPath  string
	cert     tls.Certificate
	certPEM  []byte
}

func NewCertManager(certPath, keyPath string) (*CertManager, error) {
	keyPath = paths.AbsPathify(keyPath)
	certPath = paths.AbsPathify(certPath)
	cm := &CertManager{
		certPath: certPath,
		keyPath:  keyPath,
	}

	// Check for the cert and key files.
	genKey := false
	genCert := false
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		genKey = true
		genCert = true
	}
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		genCert = true
	}

	// Generate a new certificate and private key if they don't already exist.
	var err error
	var privKey *rsa.PrivateKey
	var certBytes []byte
	if genKey {
		log.Infof("generating new private key %s", keyPath)
		if privKey, err = GeneratePrivKey(); err != nil {
			return nil, fmt.Errorf("failed to generate private key: %w", err)
		}
		if err = SavePrivKey(privKey, keyPath); err != nil {
			return nil, fmt.Errorf("failed to save private key: %w", err)
		}
	}
	if genCert {
		if privKey == nil {
			log.Debugf("loading existing private key %s", keyPath)
			if privKey, err = LoadPrivKey(keyPath); err != nil {
				return nil, fmt.Errorf("failed to load private key: %w", err)
			}
		}

		log.Infof("generating new certificate %s", certPath)
		if certBytes, err = GenerateSelfSignedCert(privKey); err != nil {
			return nil, fmt.Errorf("failed to generate cert: %w", err)
		}
		if err = SaveCert(certBytes, certPath); err != nil {
			return nil, fmt.Errorf("failed to save cert: %w", err)
		}
	}

	// Load the final cert and key.
	cm.cert, err = tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	cm.certPEM, err = os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	return cm, nil
}

func (cm *CertManager) Cert() *tls.Certificate {
	return &cm.cert
}

func (cm *CertManager) CertPEM() []byte {
	return cm.certPEM
}
