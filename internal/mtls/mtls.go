/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package mtls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

const (
	CertValidityDays        = 90
	CertExpirationThreshold = 3 * 24 * time.Hour
)

func GenerateCA(hostname string) ([]byte, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA private key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Finch"},
			CommonName:   fmt.Sprintf("Finch CA - %s", hostname),
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(CertValidityDays * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal EC private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})

	return certPEM, keyPEM, nil
}

func GenerateClient(hostname string, caCertPEM, caKeyPEM []byte) ([]byte, []byte, error) {
	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode CA certificate PEM")
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caKeyBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode CA private key PEM")
	}
	caKey, err := x509.ParseECPrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA private key: %w", err)
	}

	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate client private key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate client serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Finch"},
			CommonName:   fmt.Sprintf("Finch Client - %s", hostname),
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(CertValidityDays * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	clientCertDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create client certificate: %w", err)
	}

	clientCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientCertDER,
	})

	keyBytes, err := x509.MarshalECPrivateKey(clientKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal EC private key: %w", err)
	}
	clientKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})

	return clientCertPEM, clientKeyPEM, nil
}
