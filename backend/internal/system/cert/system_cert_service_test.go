/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/asgardeo/thunder/internal/system/config"

	"github.com/stretchr/testify/suite"
)

type SystemCertificateServiceTestSuite struct {
	suite.Suite
	service    SystemCertificateServiceInterface
	tempDir    string
	certFile   string
	keyFile    string
	testConfig *config.Config
}

func TestSystemCertificateServiceSuite(t *testing.T) {
	suite.Run(t, new(SystemCertificateServiceTestSuite))
}

func (suite *SystemCertificateServiceTestSuite) SetupSuite() {
	suite.service = NewSystemCertificateService()

	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "cert_test_")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	suite.certFile = "test.crt"
	suite.keyFile = "test.key"

	// Generate test certificate and key
	suite.generateTestCertificate()

	// Create test config
	suite.testConfig = &config.Config{
		Security: config.SecurityConfig{
			CertFile: suite.certFile,
			KeyFile:  suite.keyFile,
		},
	}
}

func (suite *SystemCertificateServiceTestSuite) TearDownSuite() {
	if suite.tempDir != "" {
		err := os.RemoveAll(suite.tempDir)
		suite.Require().NoError(err)
	}
}

func (suite *SystemCertificateServiceTestSuite) generateTestCertificate() {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	suite.Require().NoError(err)

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test Org"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: nil,
	}

	// Generate certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	suite.Require().NoError(err)

	// Write certificate file
	certPath := filepath.Join(suite.tempDir, suite.certFile)
	certFile, err := os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) // #nosec G304 - controlled path
	suite.Require().NoError(err)
	defer func() {
		cerr := certFile.Close()
		suite.Require().NoError(cerr)
	}()

	err = pem.Encode(certFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})
	suite.Require().NoError(err)

	// Write key file
	keyPath := filepath.Join(suite.tempDir, suite.keyFile)
	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) // #nosec G304 - controlled path
	suite.Require().NoError(err)
	defer func() {
		kerr := keyFile.Close()
		suite.Require().NoError(kerr)
	}()

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	suite.Require().NoError(err)

	err = pem.Encode(keyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyDER,
	})
	suite.Require().NoError(err)
}

func (suite *SystemCertificateServiceTestSuite) TestNewSystemCertificateService() {
	service := NewSystemCertificateService()
	suite.NotNil(service)
	suite.Implements((*SystemCertificateServiceInterface)(nil), service)
}

func (suite *SystemCertificateServiceTestSuite) TestGetTLSConfig_ValidCertificates() {
	tlsConfig, err := suite.service.GetTLSConfig(suite.testConfig, suite.tempDir)

	suite.NoError(err)
	suite.NotNil(tlsConfig)
	suite.Len(tlsConfig.Certificates, 1)
	suite.Equal(uint16(tls.VersionTLS12), tlsConfig.MinVersion)
}

func (suite *SystemCertificateServiceTestSuite) TestGetTLSConfig_CertificateFileNotFound() {
	config := &config.Config{
		Security: config.SecurityConfig{
			CertFile: "nonexistent.crt",
			KeyFile:  suite.keyFile,
		},
	}

	tlsConfig, err := suite.service.GetTLSConfig(config, suite.tempDir)

	suite.Error(err)
	suite.Nil(tlsConfig)
	suite.Contains(err.Error(), "certificate file not found")
}

func (suite *SystemCertificateServiceTestSuite) TestGetTLSConfig_KeyFileNotFound() {
	config := &config.Config{
		Security: config.SecurityConfig{
			CertFile: suite.certFile,
			KeyFile:  "nonexistent.key",
		},
	}

	tlsConfig, err := suite.service.GetTLSConfig(config, suite.tempDir)

	suite.Error(err)
	suite.Nil(tlsConfig)
	suite.Contains(err.Error(), "key file not found")
}

func (suite *SystemCertificateServiceTestSuite) TestGetTLSConfig_InvalidCertificateFile() {
	// Create invalid certificate file
	invalidCertPath := filepath.Join(suite.tempDir, "invalid.crt")
	err := os.WriteFile(invalidCertPath, []byte("invalid certificate content"), 0600)
	suite.Require().NoError(err)

	config := &config.Config{
		Security: config.SecurityConfig{
			CertFile: "invalid.crt",
			KeyFile:  suite.keyFile,
		},
	}

	tlsConfig, err := suite.service.GetTLSConfig(config, suite.tempDir)

	suite.Error(err)
	suite.Nil(tlsConfig)
}

func (suite *SystemCertificateServiceTestSuite) TestGetTLSConfig_InvalidKeyFile() {
	// Create invalid key file
	invalidKeyPath := filepath.Join(suite.tempDir, "invalid.key")
	err := os.WriteFile(invalidKeyPath, []byte("invalid key content"), 0600)
	suite.Require().NoError(err)

	config := &config.Config{
		Security: config.SecurityConfig{
			CertFile: suite.certFile,
			KeyFile:  "invalid.key",
		},
	}

	tlsConfig, err := suite.service.GetTLSConfig(config, suite.tempDir)

	suite.Error(err)
	suite.Nil(tlsConfig)
}

func (suite *SystemCertificateServiceTestSuite) TestGetTLSConfig_DifferentDirectory() {
	// Test with relative paths - this is how the method is intended to work
	// Copy the test certificates to the different directory for this test
	differentDir, err := os.MkdirTemp("", "cert_test_different_")
	suite.Require().NoError(err)
	defer func() {
		err := os.RemoveAll(differentDir)
		suite.Require().NoError(err)
	}()

	// Copy certificate and key files to the different directory
	srcCertPath := filepath.Join(suite.tempDir, suite.certFile)
	srcKeyPath := filepath.Join(suite.tempDir, suite.keyFile)
	destCertPath := filepath.Join(differentDir, suite.certFile)
	destKeyPath := filepath.Join(differentDir, suite.keyFile)

	// Validate source paths are within expected directory using proper path validation
	cleanTempDir := filepath.Clean(suite.tempDir)
	cleanSrcCert := filepath.Clean(srcCertPath)
	cleanSrcKey := filepath.Clean(srcKeyPath)
	expectedPrefix := cleanTempDir + string(filepath.Separator)
	suite.Require().True(strings.HasPrefix(cleanSrcCert, expectedPrefix))
	suite.Require().True(strings.HasPrefix(cleanSrcKey, expectedPrefix))

	certData, err := os.ReadFile(srcCertPath) // #nosec G304 - validated path within tempDir
	suite.Require().NoError(err)
	err = os.WriteFile(destCertPath, certData, 0600)
	suite.Require().NoError(err)

	keyData, err := os.ReadFile(srcKeyPath) // #nosec G304 - validated path within tempDir
	suite.Require().NoError(err)
	err = os.WriteFile(destKeyPath, keyData, 0600)
	suite.Require().NoError(err)

	config := &config.Config{
		Security: config.SecurityConfig{
			CertFile: suite.certFile, // relative path
			KeyFile:  suite.keyFile,  // relative path
		},
	}

	tlsConfig, err := suite.service.GetTLSConfig(config, differentDir)

	suite.NoError(err)
	suite.NotNil(tlsConfig)
	suite.Len(tlsConfig.Certificates, 1)
}

func (suite *SystemCertificateServiceTestSuite) TestGetTLSConfig_RelativePaths() {
	// Test with relative paths from the temp directory
	relativeConfig := &config.Config{
		Security: config.SecurityConfig{
			CertFile: "./test.crt",
			KeyFile:  "./test.key",
		},
	}

	tlsConfig, err := suite.service.GetTLSConfig(relativeConfig, suite.tempDir)

	suite.NoError(err)
	suite.NotNil(tlsConfig)
	suite.Len(tlsConfig.Certificates, 1)
}

func (suite *SystemCertificateServiceTestSuite) TestGetCertificateKid_ValidTLSConfig() {
	tlsConfig, err := suite.service.GetTLSConfig(suite.testConfig, suite.tempDir)
	suite.Require().NoError(err)

	kid, err := suite.service.GetCertificateKid(tlsConfig)

	suite.NoError(err)
	suite.NotEmpty(kid)
	// Kid should be a base64 encoded SHA-256 hash
	suite.Regexp(`^[A-Za-z0-9+/]+=*$`, kid)
}

func (suite *SystemCertificateServiceTestSuite) TestGetCertificateKid_NilTLSConfig() {
	kid, err := suite.service.GetCertificateKid(nil)

	suite.Error(err)
	suite.Empty(kid)
	suite.Equal("TLS configuration is not set", err.Error())
}

func (suite *SystemCertificateServiceTestSuite) TestGetCertificateKid_EmptyCertificates() {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{},
		MinVersion:   tls.VersionTLS12, // #nosec G402 - MinVersion is set for test config
	}

	kid, err := suite.service.GetCertificateKid(tlsConfig)

	suite.Error(err)
	suite.Empty(kid)
	suite.Equal("no certificate found in TLS config", err.Error())
}

func (suite *SystemCertificateServiceTestSuite) TestGetCertificateKid_EmptyCertificateData() {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{},
			},
		},
		MinVersion: tls.VersionTLS12, // #nosec G402 - MinVersion is set for test config
	}

	kid, err := suite.service.GetCertificateKid(tlsConfig)

	suite.Error(err)
	suite.Empty(kid)
	suite.Equal("no certificate found in TLS config", err.Error())
}

func (suite *SystemCertificateServiceTestSuite) TestGetCertificateKid_InvalidCertificateData() {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{
					[]byte("invalid certificate data"),
				},
			},
		},
		MinVersion: tls.VersionTLS12, // #nosec G402 - MinVersion is set for test config
	}

	kid, err := suite.service.GetCertificateKid(tlsConfig)

	suite.Error(err)
	suite.Empty(kid)
}

func (suite *SystemCertificateServiceTestSuite) TestGetCertificateKid_Consistency() {
	// Test that the same certificate always generates the same kid
	tlsConfig, err := suite.service.GetTLSConfig(suite.testConfig, suite.tempDir)
	suite.Require().NoError(err)

	kid1, err1 := suite.service.GetCertificateKid(tlsConfig)
	suite.Require().NoError(err1)

	kid2, err2 := suite.service.GetCertificateKid(tlsConfig)
	suite.Require().NoError(err2)

	suite.Equal(kid1, kid2, "Same certificate should always generate the same kid")
}

func (suite *SystemCertificateServiceTestSuite) TestIntegration_GetTLSConfigAndKid() {
	// Integration test: Get TLS config and then extract kid
	tlsConfig, err := suite.service.GetTLSConfig(suite.testConfig, suite.tempDir)
	suite.Require().NoError(err)
	suite.NotNil(tlsConfig)

	kid, err := suite.service.GetCertificateKid(tlsConfig)
	suite.Require().NoError(err)
	suite.NotEmpty(kid)

	// Verify the TLS config is actually usable
	suite.Len(tlsConfig.Certificates, 1)
	suite.Equal(uint16(tls.VersionTLS12), tlsConfig.MinVersion)

	// Verify certificate data is accessible
	certData := tlsConfig.Certificates[0].Certificate[0]
	suite.NotEmpty(certData)

	// Verify the certificate can be parsed
	parsedCert, err := x509.ParseCertificate(certData)
	suite.NoError(err)
	suite.NotNil(parsedCert)
}

// Benchmark tests for performance verification
func (suite *SystemCertificateServiceTestSuite) TestGetTLSConfig_Performance() {
	// This is a simple performance check to ensure loading certificates is reasonable
	start := time.Now()
	for i := 0; i < 10; i++ {
		_, err := suite.service.GetTLSConfig(suite.testConfig, suite.tempDir)
		suite.NoError(err)
	}
	duration := time.Since(start)

	// Should be able to load certificates 10 times in under 1 second
	suite.True(duration < time.Second, "Certificate loading should be fast")
}

func (suite *SystemCertificateServiceTestSuite) TestGetCertificateKid_Performance() {
	tlsConfig, err := suite.service.GetTLSConfig(suite.testConfig, suite.tempDir)
	suite.Require().NoError(err)

	start := time.Now()
	for i := 0; i < 100; i++ {
		_, err := suite.service.GetCertificateKid(tlsConfig)
		suite.NoError(err)
	}
	duration := time.Since(start)

	// Should be able to generate kid 100 times in under 1 second
	suite.True(duration < time.Second, "Kid generation should be fast")
}
