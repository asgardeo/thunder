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
	"crypto/tls"
	"errors"
	"os"
	"path"

	"github.com/asgardeo/thunder/internal/system/log"
)

// LoadTLSConfig loads the TLS configuration from the certificate and key files.
func LoadTLSConfig(certFilePath, keyFilePath string) (*tls.Config, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "PKIUtils"))

	// Validate file paths
	if certFilePath == "" {
		return nil, errors.New("certificate file path is empty")
	}
	if keyFilePath == "" {
		return nil, errors.New("key file path is empty")
	}

	// Clean file paths
	certFilePath = path.Clean(certFilePath)
	keyFilePath = path.Clean(keyFilePath)

	// Check if the certificate and key files exist
	if _, err := os.Stat(certFilePath); os.IsNotExist(err) {
		return nil, errors.New("certificate file not found at " + certFilePath)
	}
	if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
		return nil, errors.New("key file not found at " + keyFilePath)
	}

	// Load the certificate and key
	cert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
	if err != nil {
		logger.Error("Failed to load X509 key pair", log.Error(err))
		return nil, err
	}

	logger.Debug("Successfully loaded TLS certificate",
		log.String("certFile", certFilePath),
		log.String("keyFile", keyFilePath))

	// Return the TLS configuration
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12, // Enforce minimum TLS version 1.2
	}, nil
}
