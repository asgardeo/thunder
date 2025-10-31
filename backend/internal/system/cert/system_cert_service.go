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

// Package cert provides functionality for managing system certificates.
package cert

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"path"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/hash"
)

// SystemCertificateServiceInterface defines the interface for system certificate operations.
type SystemCertificateServiceInterface interface {
	GetTLSConfig(cfg *config.Config, currentDirectory string) (*tls.Config, error)
	GetCertificateKid(tlsConfig *tls.Config) (string, error)
}

// SystemCertificateService implements the SystemCertificateServiceInterface for managing system certificates.
type systemCertificateService struct{}

// NewSystemCertificateService creates a new instance of SystemCertificateService.
func NewSystemCertificateService() SystemCertificateServiceInterface {
	return &systemCertificateService{}
}

// GetTLSConfig loads the TLS configuration from the certificate and key files.
func (c *systemCertificateService) GetTLSConfig(cfg *config.Config, currentDirectory string) (*tls.Config, error) {
	certFilePath := path.Join(currentDirectory, cfg.Security.CertFile)
	keyFilePath := path.Join(currentDirectory, cfg.Security.KeyFile)

	// Check if the certificate and key files exist.
	if _, err := os.Stat(certFilePath); os.IsNotExist(err) {
		return nil, errors.New("certificate file not found at " + certFilePath)
	}
	if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
		return nil, errors.New("key file not found at " + keyFilePath)
	}

	// Load the certificate and key.
	cert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
	if err != nil {
		return nil, err
	}

	// Return the TLS configuration.
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12, // Enforce minimum TLS version 1.2
	}, nil
}

// GetCertificateKid extracts the Key ID (kid) from the TLS certificate using SHA-256 thumbprint.
func (c *systemCertificateService) GetCertificateKid(tlsConfig *tls.Config) (string, error) {
	if tlsConfig == nil {
		return "", errors.New("TLS configuration is not set")
	}
	if len(tlsConfig.Certificates) == 0 || len(tlsConfig.Certificates[0].Certificate) == 0 {
		return "", errors.New("no certificate found in TLS config")
	}

	certData := tlsConfig.Certificates[0].Certificate[0]
	parsedCert, err := x509.ParseCertificate(certData)
	if err != nil {
		return "", err
	}

	x5tS256 := hash.GenerateThumbprint(parsedCert.Raw)
	return x5tS256, nil
}
