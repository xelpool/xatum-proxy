package server

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

func GenCertificate() ([]byte, []byte, error) {
	pubkey, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	keyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "OPENSSH PRIVATE KEY",
			Bytes: keyBytes,
		},
	)

	notBefore := time.Now()
	notAfter := notBefore.Add(15 * 365 * 24 * time.Hour) // certificate expires in 15 years

	template := x509.Certificate{
		SerialNumber:          big.NewInt(0),
		Subject:               pkix.Name{CommonName: "mining pool"},
		SignatureAlgorithm:    x509.PureEd25519,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, pubkey, key)
	if err != nil {
		return []byte{}, []byte{}, err

	}
	certPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: derBytes,
		},
	)
	err = os.WriteFile("key.pem", keyPem, 0o600)
	if err != nil {
		return []byte{}, []byte{}, err
	}
	return certPem, keyPem, os.WriteFile("cert.pem", certPem, 0o600)
}
