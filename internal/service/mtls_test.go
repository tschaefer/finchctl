/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/tschaefer/finchctl/internal/config"
)

func Test_GenerateCA(t *testing.T) {
	hostname := "finch." + gofakeit.DomainName()

	caCertPEM, caKeyPEM, err := __mtlsGenerateCA(hostname)
	assert.NoError(t, err, "generate CA should not error")
	assert.NotNil(t, caCertPEM, "CA cert should not be nil")
	assert.NotNil(t, caKeyPEM, "CA key should not be nil")

	block, _ := pem.Decode(caCertPEM)
	assert.NotNil(t, block, "CA cert PEM should decode")

	cert, err := x509.ParseCertificate(block.Bytes)
	assert.NoError(t, err, "CA cert should parse")
	assert.True(t, cert.IsCA, "cert should be CA")
	assert.Contains(t, cert.Subject.CommonName, hostname, "cert CN should contain hostname")
}

func Test_GenerateClientCert(t *testing.T) {
	hostname := "finch." + gofakeit.DomainName()

	caCertPEM, caKeyPEM, err := __mtlsGenerateCA(hostname)
	assert.NoError(t, err, "generate CA should not error")

	clientCertPEM, clientKeyPEM, err := __mtlsGenerateClientCert(hostname, caCertPEM, caKeyPEM)
	assert.NoError(t, err, "generate client cert should not error")
	assert.NotNil(t, clientCertPEM, "client cert should not be nil")
	assert.NotNil(t, clientKeyPEM, "client key should not be nil")

	block, _ := pem.Decode(clientCertPEM)
	assert.NotNil(t, block, "client cert PEM should decode")

	cert, err := x509.ParseCertificate(block.Bytes)
	assert.NoError(t, err, "client cert should parse")
	assert.False(t, cert.IsCA, "cert should not be CA")
	assert.Contains(t, cert.Subject.CommonName, hostname, "cert CN should contain hostname")
	assert.Contains(t, cert.ExtKeyUsage, x509.ExtKeyUsageClientAuth, "cert should have client auth usage")
}

func Test_IsCertificateExpired(t *testing.T) {
	hostname := "finch." + gofakeit.DomainName()

	caCertPEM, _, err := __mtlsGenerateCA(hostname)
	assert.NoError(t, err, "generate CA should not error")

	expired, err := __mtlsIsCertificateExpired(caCertPEM)
	assert.NoError(t, err, "check expiration should not error")
	assert.False(t, expired, "newly created cert should not be expired")
}

func Test_SaveAndLoadCertificate(t *testing.T) {
	hostname := "finch." + gofakeit.DomainName()

	caCertPEM, caKeyPEM, err := __mtlsGenerateCA(hostname)
	assert.NoError(t, err, "generate CA should not error")

	clientCertPEM, clientKeyPEM, err := __mtlsGenerateClientCert(hostname, caCertPEM, caKeyPEM)
	assert.NoError(t, err, "generate client cert should not error")

	tmpDir := t.TempDir()
	t.Setenv("FINCH_CONFIG", tmpDir)

	err = config.UpdateStack(hostname, clientCertPEM, clientKeyPEM)
	assert.NoError(t, err, "update stack certs should not error")

	loadedCertPEM, loadedKeyPEM, err := config.LookupStack(hostname)
	assert.NoError(t, err, "lookup stack certs should not error")
	assert.Equal(t, clientCertPEM, loadedCertPEM, "loaded cert should match saved cert")
	assert.Equal(t, clientKeyPEM, loadedKeyPEM, "loaded key should match saved key")
}

func Test_CertificateValidity(t *testing.T) {
	hostname := "finch." + gofakeit.DomainName()

	caCertPEM, caKeyPEM, err := __mtlsGenerateCA(hostname)
	assert.NoError(t, err, "generate CA should not error")

	clientCertPEM, _, err := __mtlsGenerateClientCert(hostname, caCertPEM, caKeyPEM)
	assert.NoError(t, err, "generate client cert should not error")

	block, _ := pem.Decode(clientCertPEM)
	cert, err := x509.ParseCertificate(block.Bytes)
	assert.NoError(t, err, "client cert should parse")

	assert.True(t, cert.NotBefore.Before(time.Now()), "cert should be valid from the past")
	assert.True(t, cert.NotAfter.After(time.Now()), "cert should be valid in the future")

	minValidity := time.Now().Add(87 * 24 * time.Hour)
	assert.True(t, cert.NotAfter.After(minValidity), "cert should be valid for at least 87 days")
}
