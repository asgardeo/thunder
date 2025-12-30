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

package pki

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/asgardeo/thunder/internal/system/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PKIInitTestSuite struct {
	suite.Suite
	tempFiles []string
}

func TestPKIInitSuite(t *testing.T) {
	suite.Run(t, new(PKIInitTestSuite))
}

func (suite *PKIInitTestSuite) TearDownSuite() {
	for _, file := range suite.tempFiles {
		err := os.Remove(file)
		if err != nil {
			suite.T().Logf("Failed to remove temp file %s: %v", file, err)
		}
	}
}

func (suite *PKIInitTestSuite) SetupTest() {
	config.ResetThunderRuntime()
}

func (suite *PKIInitTestSuite) TestInitialize_Success() {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	// Create self-signed certificate with absolute paths
	certPath, keyPath := suite.createAndWriteFilesAbsolute(privateKey)
	defer suite.cleanupFiles(certPath, keyPath)

	// Create test config - use empty ThunderHome so absolute paths work
	cfg := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "test-key-1",
					CertFile: certPath,
					KeyFile:  keyPath,
				},
			},
		},
	}

	err = config.InitializeThunderRuntime("", cfg)
	assert.NoError(suite.T(), err)

	// Initialize PKI service
	pkiService, err := Initialize()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), pkiService)
}

func (suite *PKIInitTestSuite) createAndWriteFilesAbsolute(privateKey *rsa.PrivateKey) (string, string) {
	// Create self-signed certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	assert.NoError(suite.T(), err)

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	// Create temporary certificate file
	certFile, err := os.CreateTemp("", "init_test_cert_abs_*.pem")
	assert.NoError(suite.T(), err)
	suite.tempFiles = append(suite.tempFiles, certFile.Name())

	_, err = certFile.Write(certPEM)
	assert.NoError(suite.T(), err)
	err = certFile.Close()
	assert.NoError(suite.T(), err)

	// Create temporary private key file
	keyFile, err := os.CreateTemp("", "init_test_key_abs_*.pem")
	assert.NoError(suite.T(), err)
	suite.tempFiles = append(suite.tempFiles, keyFile.Name())

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	_, err = keyFile.Write(privateKeyPEM)
	assert.NoError(suite.T(), err)
	err = keyFile.Close()
	assert.NoError(suite.T(), err)

	return certFile.Name(), keyFile.Name()
}

func (suite *PKIInitTestSuite) TestInitialize_NoKeyConfigs() {
	cfg := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{},
		},
	}

	err := config.InitializeThunderRuntime("", cfg)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), pkiService)
}

func (suite *PKIInitTestSuite) TestInitialize_MissingCertFile() {
	cfg := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "test-key-1",
					CertFile: "/non/existent/cert.pem",
					KeyFile:  "/non/existent/key.pem",
				},
			},
		},
	}

	err := config.InitializeThunderRuntime("", cfg)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), pkiService)
}

func (suite *PKIInitTestSuite) TestInitialize_MissingKeyFile() {
	// Create a valid cert file with absolute path
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	certPath, _ := suite.createAndWriteFilesAbsolute(privateKey)
	defer suite.cleanupFiles(certPath, "")

	cfg := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "test-key-1",
					CertFile: certPath,
					KeyFile:  "/non/existent/key.pem",
				},
			},
		},
	}

	err = config.InitializeThunderRuntime("", cfg)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), pkiService)
}

func (suite *PKIInitTestSuite) TestInitialize_WithECDSAKey() {
	// Generate ECDSA key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(suite.T(), err)

	certPath, keyPath := suite.createAndWriteFilesECAbsolute(privateKey)
	defer suite.cleanupFiles(certPath, keyPath)

	cfg := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "test-ec-key",
					CertFile: certPath,
					KeyFile:  keyPath,
				},
			},
		},
	}

	err = config.InitializeThunderRuntime("", cfg)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), pkiService)
}

func (suite *PKIInitTestSuite) createAndWriteFilesECAbsolute(privateKey *ecdsa.PrivateKey) (string, string) {
	// Create self-signed certificate with ECDSA key
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test-ec.example.com",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	assert.NoError(suite.T(), err)

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	// Create temporary certificate file
	certFile, err := os.CreateTemp("", "init_test_ec_cert_abs_*.pem")
	assert.NoError(suite.T(), err)
	suite.tempFiles = append(suite.tempFiles, certFile.Name())

	_, err = certFile.Write(certPEM)
	assert.NoError(suite.T(), err)
	err = certFile.Close()
	assert.NoError(suite.T(), err)

	// Create temporary private key file
	keyFile, err := os.CreateTemp("", "init_test_ec_key_abs_*.pem")
	assert.NoError(suite.T(), err)
	suite.tempFiles = append(suite.tempFiles, keyFile.Name())

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	assert.NoError(suite.T(), err)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	_, err = keyFile.Write(privateKeyPEM)
	assert.NoError(suite.T(), err)
	err = keyFile.Close()
	assert.NoError(suite.T(), err)

	return certFile.Name(), keyFile.Name()
}

func (suite *PKIInitTestSuite) cleanupFiles(certPath, keyPath string) {
	if certPath != "" {
		err := os.Remove(certPath)
		if err != nil {
			suite.T().Logf("Failed to remove cert file %s: %v", certPath, err)
		}
	}
	if keyPath != "" {
		err := os.Remove(keyPath)
		if err != nil {
			suite.T().Logf("Failed to remove key file %s: %v", keyPath, err)
		}
	}
}
