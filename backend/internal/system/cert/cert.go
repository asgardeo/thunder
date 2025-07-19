/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

// Package cert provides utilities for managing TLS certificates.
package cert

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"os"
	"path"

	"github.com/asgardeo/thunder/internal/system/config"
)

// GetTLSConfig loads the TLS configuration from the certificate and key files.
func GetTLSConfig(cfg *config.Config, currentDirectory string) (*tls.Config, error) {
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
func GetCertificateKid(cfg *config.Config, currentDirectory string) (string, error) {
	tlsConfig, err := GetTLSConfig(cfg, currentDirectory)
	if err != nil {
		return "", err
	}

	if len(tlsConfig.Certificates) == 0 || len(tlsConfig.Certificates[0].Certificate) == 0 {
		return "", errors.New("no certificate found in TLS config")
	}

	certData := tlsConfig.Certificates[0].Certificate[0]
	parsedCert, err := x509.ParseCertificate(certData)
	if err != nil {
		return "", err
	}

	// Calculate SHA-256 thumbprint and use it as kid
	h := sha256.New()
	h.Write(parsedCert.Raw)
	x5tS256 := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return x5tS256, nil
}
