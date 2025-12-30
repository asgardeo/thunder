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
)

type PKIUtilsTestSuite struct {
	suite.Suite
	testCertPath string
	testKeyPath  string
	tempFiles    []string
}

func TestPKIUtilsSuite(t *testing.T) {
	suite.Run(t, new(PKIUtilsTestSuite))
}

func (suite *PKIUtilsTestSuite) SetupSuite() {
	// Generate a test RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	// Create a self-signed certificate
	_, certPEM, err := createTestSelfSignedCert(privateKey)
	assert.NoError(suite.T(), err)

	// Create temporary certificate file
	certFile, err := os.CreateTemp("", "utils_test_cert_*.pem")
	assert.NoError(suite.T(), err)
	suite.testCertPath = certFile.Name()
	suite.tempFiles = append(suite.tempFiles, suite.testCertPath)

	_, err = certFile.Write(certPEM)
	assert.NoError(suite.T(), err)
	err = certFile.Close()
	assert.NoError(suite.T(), err)

	// Create temporary private key file
	keyFile, err := os.CreateTemp("", "utils_test_key_*.pem")
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
}

func (suite *PKIUtilsTestSuite) TearDownSuite() {
	for _, file := range suite.tempFiles {
		err := os.Remove(file)
		if err != nil {
			suite.T().Logf("Failed to remove temp file %s: %v", file, err)
		}
	}
}

func (suite *PKIUtilsTestSuite) TestLoadTLSConfig_Success() {
	tlsConfig, err := LoadTLSConfig(suite.testCertPath, suite.testKeyPath)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tlsConfig)
	assert.Equal(suite.T(), uint16(tls.VersionTLS12), tlsConfig.MinVersion)
	assert.Len(suite.T(), tlsConfig.Certificates, 1)
}

func (suite *PKIUtilsTestSuite) TestLoadTLSConfig_EmptyCertPath() {
	tlsConfig, err := LoadTLSConfig("", suite.testKeyPath)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tlsConfig)
	assert.Contains(suite.T(), err.Error(), "certificate file path is empty")
}

func (suite *PKIUtilsTestSuite) TestLoadTLSConfig_EmptyKeyPath() {
	tlsConfig, err := LoadTLSConfig(suite.testCertPath, "")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tlsConfig)
	assert.Contains(suite.T(), err.Error(), "key file path is empty")
}

func (suite *PKIUtilsTestSuite) TestLoadTLSConfig_MissingCertFile() {
	tlsConfig, err := LoadTLSConfig("non_existent_cert.pem", suite.testKeyPath)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tlsConfig)
	assert.Contains(suite.T(), err.Error(), "certificate file not found")
}

func (suite *PKIUtilsTestSuite) TestLoadTLSConfig_MissingKeyFile() {
	tlsConfig, err := LoadTLSConfig(suite.testCertPath, "non_existent_key.pem")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tlsConfig)
	assert.Contains(suite.T(), err.Error(), "key file not found")
}

func createTestSelfSignedCert(privateKey *rsa.PrivateKey) (*x509.Certificate, []byte, error) {
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
