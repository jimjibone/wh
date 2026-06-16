package cert

import (
	"crypto/rsa"
	"crypto/tls"
	"fmt"

	"github.com/jimjibone/woodhouse-4/log"
	"github.com/jimjibone/woodhouse-4/shared/stores"
)

var (
	storeCertPath = "woodhouse.crt"
	storeKeyPath  = "woodhouse.key"
)

type CertManager struct {
	store   stores.Store
	cert    tls.Certificate
	certPEM []byte
}

func NewCertManager(store stores.Store) (*CertManager, error) {
	cm := &CertManager{
		store: store,
	}

	// Check for the cert and key files.
	genKey := false
	genCert := false
	if !store.Has(storeKeyPath) {
		genKey = true
		genCert = true
	}
	if !store.Has(storeCertPath) {
		genCert = true
	}

	// Generate a new certificate and private key if they don't already exist.
	var err error
	var privKey *rsa.PrivateKey
	var certBytes []byte
	if genKey {
		log.Infof("generating new private key")
		if privKey, err = GeneratePrivKey(); err != nil {
			return nil, fmt.Errorf("failed to generate private key: %w", err)
		}
		data, err := EncodePrivKey(privKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encode private key: %w", err)
		}
		err = store.Set(storeKeyPath, data)
		if err != nil {
			return nil, fmt.Errorf("failed to save private key: %w", err)
		}
	}
	if genCert {
		if privKey == nil {
			log.Debugf("loading existing private key")
			data, err := store.Get(storeKeyPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load private key: %w", err)
			}
			privKey, err = DecodePrivKey(data)
			if err != nil {
				return nil, fmt.Errorf("failed to decode private key: %w", err)
			}
		}

		log.Infof("generating new certificate")
		if certBytes, err = GenerateSelfSignedCert(privKey); err != nil {
			return nil, fmt.Errorf("failed to generate cert: %w", err)
		}
		data, err := EncodeCert(certBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to encode cert: %w", err)
		}
		err = store.Set(storeCertPath, data)
		if err != nil {
			return nil, fmt.Errorf("failed to save cert: %w", err)
		}
	}

	// Load the final cert and key.
	certData, err := store.Get(storeCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load cert: %w", err)
	}
	keyData, err := store.Get(storeKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}
	cm.cert, err = tls.X509KeyPair(certData, keyData)
	if err != nil {
		return nil, err
	}
	cm.certPEM = certData

	return cm, nil
}

func (cm *CertManager) Cert() *tls.Certificate {
	return &cm.cert
}

func (cm *CertManager) CertPEM() []byte {
	return cm.certPEM
}
