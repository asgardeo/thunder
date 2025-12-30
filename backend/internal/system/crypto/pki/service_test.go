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
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
)

type PKIServiceTestSuite struct {
	suite.Suite
	testPrivateKey *rsa.PrivateKey
	testKeyPath    string
	testCertPath   string
	tempFiles      []string
}

func TestPKIServiceSuite(t *testing.T) {
	suite.Run(t, new(PKIServiceTestSuite))
}

func (suite *PKIServiceTestSuite) SetupSuite() {
	// Generate a test RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)
	suite.testPrivateKey = privateKey

	// Create a self-signed certificate
	cert, certPEM, err := createSelfSignedCert(privateKey)
	assert.NoError(suite.T(), err)

	// Create temporary certificate file
	certFile, err := os.CreateTemp("", "test_cert_*.pem")
	assert.NoError(suite.T(), err)
	suite.testCertPath = certFile.Name()
	suite.tempFiles = append(suite.tempFiles, suite.testCertPath)

	_, err = certFile.Write(certPEM)
	assert.NoError(suite.T(), err)
	err = certFile.Close()
	assert.NoError(suite.T(), err)

	// Create temporary private key file
	keyFile, err := os.CreateTemp("", "test_key_*.pem")
	assert.NoError(suite.T(), err)
	suite.testKeyPath = keyFile.Name()
	suite.tempFiles = append(suite.tempFiles, suite.testKeyPath)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	_, err = keyFile.Write(privateKeyPEM)
	assert.NoError(suite.T(), err)
	err = keyFile.Close()
	assert.NoError(suite.T(), err)

	_ = cert // silence unused
}

func (suite *PKIServiceTestSuite) TearDownSuite() {
	for _, file := range suite.tempFiles {
		err := os.Remove(file)
		if err != nil {
			suite.T().Logf("Failed to remove temp file %s: %v", file, err)
		}
	}
}

func (suite *PKIServiceTestSuite) SetupTest() {
	config.ResetThunderRuntime()
}

func (suite *PKIServiceTestSuite) TestGetX509Certificate_Success() {
	// Build a pkiService with an in-memory certificate
	cert, _, err := createSelfSignedCert(suite.testPrivateKey)
	assert.NoError(suite.T(), err)

	tlsCert := tls.Certificate{Certificate: [][]byte{cert.Raw}}
	pkiSvc := &pkiService{certificates: map[string]PKI{
		"id-1": {
			Certificate: tlsCert,
		},
	}}

	parsed, err := pkiSvc.GetX509Certificate("id-1")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), cert.Raw, parsed.Raw)
}

func (suite *PKIServiceTestSuite) TestGetX509Certificate_NoData() {
	pkiSvc := &pkiService{certificates: map[string]PKI{
		"id-1": {
			Certificate: tls.Certificate{Certificate: [][]byte{}},
		},
	}}

	parsed, err := pkiSvc.GetX509Certificate("id-1")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), parsed)
	assert.Contains(suite.T(), err.Error(), "no certificate data")
}

func (suite *PKIServiceTestSuite) TestGetX509Certificate_ParseError() {
	pkiSvc := &pkiService{certificates: map[string]PKI{
		"id-1": {
			Certificate: tls.Certificate{Certificate: [][]byte{[]byte("bad-cert-bytes")}},
		},
	}}

	parsed, err := pkiSvc.GetX509Certificate("id-1")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), parsed)
}

func (suite *PKIServiceTestSuite) TestGetX509Certificate_NotFound() {
	pkiSvc := &pkiService{certificates: map[string]PKI{}}
	parsed, err := pkiSvc.GetX509Certificate("missing")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not found")
	assert.Nil(suite.T(), parsed)
}

func (suite *PKIServiceTestSuite) TestInitialize_Success() {
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "test-key",
					CertFile: suite.testCertPath,
					KeyFile:  suite.testKeyPath,
				},
			},
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), pkiService)
	assert.Implements(suite.T(), (*PKIServiceInterface)(nil), pkiService)
}

func (suite *PKIServiceTestSuite) TestInitialize_NoKeyConfigs() {
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{},
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), pkiService)
}

func (suite *PKIServiceTestSuite) TestInitialize_MissingCertFile() {
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "test-key",
					CertFile: "non_existent_cert.pem",
					KeyFile:  suite.testKeyPath,
				},
			},
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), pkiService)
	assert.Contains(suite.T(), err.Error(), "certificate file not found")
}

func (suite *PKIServiceTestSuite) TestInitialize_MissingKeyFile() {
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "test-key",
					CertFile: suite.testCertPath,
					KeyFile:  "non_existent_key.pem",
				},
			},
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), pkiService)
	assert.Contains(suite.T(), err.Error(), "key file not found")
}

func (suite *PKIServiceTestSuite) TestGetPrivateKey_Success() {
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "test-key",
					CertFile: suite.testCertPath,
					KeyFile:  suite.testKeyPath,
				},
			},
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.NoError(suite.T(), err)

	privateKey, err := pkiService.GetPrivateKey("test-key")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), privateKey)
	assert.IsType(suite.T(), &rsa.PrivateKey{}, privateKey)
}

func (suite *PKIServiceTestSuite) TestGetPrivateKey_NotFound() {
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "test-key",
					CertFile: suite.testCertPath,
					KeyFile:  suite.testKeyPath,
				},
			},
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.NoError(suite.T(), err)

	privateKey, err := pkiService.GetPrivateKey("non-existent-key")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), privateKey)
	assert.Contains(suite.T(), err.Error(), "not found")
}

func (suite *PKIServiceTestSuite) TestGetCertThumbprint_Success() {
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "test-key",
					CertFile: suite.testCertPath,
					KeyFile:  suite.testKeyPath,
				},
			},
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.NoError(suite.T(), err)

	thumbprint := pkiService.GetCertThumbprint("test-key")
	assert.NotEmpty(suite.T(), thumbprint)
}

func (suite *PKIServiceTestSuite) TestGetCertThumbprint_NotFound() {
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "test-key",
					CertFile: suite.testCertPath,
					KeyFile:  suite.testKeyPath,
				},
			},
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.NoError(suite.T(), err)

	thumbprint := pkiService.GetCertThumbprint("non-existent-key")
	assert.Empty(suite.T(), thumbprint)
}

func (suite *PKIServiceTestSuite) TestInitialize_WithECDSAKey() {
	// Generate ECDSA P-256 key
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(suite.T(), err)

	// Create self-signed certificate with ECDSA key
	_, certPEM, err := createSelfSignedCertEC(ecKey)
	assert.NoError(suite.T(), err)

	// Write certificate to temp file
	certFile, err := os.CreateTemp("", "test_ec_cert_*.pem")
	assert.NoError(suite.T(), err)
	defer func() {
		err := os.Remove(certFile.Name())
		if err != nil {
			suite.T().Logf("Failed to remove cert file %s: %v", certFile.Name(), err)
		}
	}()

	_, err = certFile.Write(certPEM)
	assert.NoError(suite.T(), err)
	err = certFile.Close()
	assert.NoError(suite.T(), err)

	// Write EC private key to temp file
	ecBytes, err := x509.MarshalECPrivateKey(ecKey)
	assert.NoError(suite.T(), err)

	ecPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: ecBytes,
	})

	keyFile, err := os.CreateTemp("", "test_ec_key_*.pem")
	assert.NoError(suite.T(), err)
	defer func() {
		err := os.Remove(keyFile.Name())
		if err != nil {
			suite.T().Logf("Failed to remove key file %s: %v", keyFile.Name(), err)
		}
	}()

	_, err = keyFile.Write(ecPEM)
	assert.NoError(suite.T(), err)
	err = keyFile.Close()
	assert.NoError(suite.T(), err)

	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "ec-key",
					CertFile: certFile.Name(),
					KeyFile:  keyFile.Name(),
				},
			},
		},
	}
	err = config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), pkiService)

	privKey, err := pkiService.GetPrivateKey("ec-key")
	assert.NoError(suite.T(), err)
	assert.IsType(suite.T(), &ecdsa.PrivateKey{}, privKey)
}

func (suite *PKIServiceTestSuite) TestInitialize_SkipsEmptyIDButLoadsOthers() {
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{ID: "", CertFile: suite.testCertPath, KeyFile: suite.testKeyPath},
				{ID: "valid", CertFile: suite.testCertPath, KeyFile: suite.testKeyPath},
			},
		},
	}
	assert.NoError(suite.T(), config.InitializeThunderRuntime("", testConfig))

	pkiService, err := Initialize()
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), pkiService)
	assert.Contains(suite.T(), err.Error(), "empty ID")
}

func (suite *PKIServiceTestSuite) TestInitialize_WithEd25519Key() {
	// Generate Ed25519 key
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(suite.T(), err)

	// Create self-signed certificate with Ed25519 key
	_, certPEM, err := createSelfSignedCertEd25519(privKey)
	assert.NoError(suite.T(), err)

	// Write certificate to temp file
	certFile, err := os.CreateTemp("", "test_ed_cert_*.pem")
	assert.NoError(suite.T(), err)
	defer func() {
		err := os.Remove(certFile.Name())
		if err != nil {
			suite.T().Logf("Failed to remove cert file %s: %v", certFile.Name(), err)
		}
	}()

	_, err = certFile.Write(certPEM)
	assert.NoError(suite.T(), err)
	err = certFile.Close()
	assert.NoError(suite.T(), err)

	// Write Ed25519 private key to temp file
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	assert.NoError(suite.T(), err)

	edPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	keyFile, err := os.CreateTemp("", "test_ed_key_*.pem")
	assert.NoError(suite.T(), err)
	defer func() {
		err := os.Remove(keyFile.Name())
		if err != nil {
			suite.T().Logf("Failed to remove key file %s: %v", keyFile.Name(), err)
		}
	}()

	_, err = keyFile.Write(edPEM)
	assert.NoError(suite.T(), err)
	err = keyFile.Close()
	assert.NoError(suite.T(), err)

	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "ed-key",
					CertFile: certFile.Name(),
					KeyFile:  keyFile.Name(),
				},
			},
		},
	}
	err = config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	pkiService, err := Initialize()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), pkiService)

	privKeyRetrieved, err := pkiService.GetPrivateKey("ed-key")
	assert.NoError(suite.T(), err)
	assert.IsType(suite.T(), ed25519.PrivateKey{}, privKeyRetrieved)

	_ = pubKey // silence unused
}

func (suite *PKIServiceTestSuite) TestGetAlgorithmFromKey_VariantsAndErrors() {
	// P-384
	ec384, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	alg, err := getAlgorithmFromKey(ec384)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), P384, alg)

	// P-521
	ec521, _ := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	alg, err = getAlgorithmFromKey(ec521)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), P521, alg)

	// Unsupported curve
	ec224, _ := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	alg, err = getAlgorithmFromKey(ec224)
	assert.Error(suite.T(), err)
	assert.Empty(suite.T(), alg)

	// Unsupported type
	alg, err = getAlgorithmFromKey(123)
	assert.Error(suite.T(), err)
	assert.Empty(suite.T(), alg)
}

// Helper function to create a self-signed certificate
func createSelfSignedCert(privateKey *rsa.PrivateKey) (*x509.Certificate, []byte, error) {
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	parsedCert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, nil, err
	}

	return parsedCert, certPEM, nil
}

// Helper function to create a self-signed certificate with ECDSA key
func createSelfSignedCertEC(privateKey *ecdsa.PrivateKey) (*x509.Certificate, []byte, error) {
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	parsedCert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, nil, err
	}

	return parsedCert, certPEM, nil
}

// Helper function to create a self-signed certificate with Ed25519 key
func createSelfSignedCertEd25519(privateKey ed25519.PrivateKey) (*x509.Certificate, []byte, error) {
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, privateKey.Public(), privateKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	parsedCert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, nil, err
	}

	return parsedCert, certPEM, nil
}
