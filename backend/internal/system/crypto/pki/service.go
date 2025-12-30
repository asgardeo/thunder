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

// Package pki provides PKI related functionalities.
package pki

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"path"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/log"
)

var (
	logger = log.GetLogger().With(log.String(log.LoggerKeyComponentName, "PKIService"))
)

// PKIServiceInterface defines the interface for certificate services.
type PKIServiceInterface interface {
	GetPrivateKey(id string) (crypto.PrivateKey, error)
	GetCertThumbprint(id string) string
	GetX509Certificate(id string) (*x509.Certificate, error)
}

// pkiService stores loaded certificates indexed by their ID
type pkiService struct {
	certificates map[string]PKI
}

// newPKIService initializes and returns the cert service
func newPKIService() (PKIServiceInterface, error) {
	thunderRuntime := config.GetThunderRuntime()
	keyConfigs := thunderRuntime.Config.Crypto.Keys
	if len(keyConfigs) == 0 {
		logger.Warn("No key configurations found")
		return &pkiService{
			certificates: make(map[string]PKI),
		}, nil
	}

	certificates := make(map[string]PKI)
	for _, keyConfig := range keyConfigs {
		if keyConfig.ID == "" {
			return nil, errors.New("key configuration has empty ID")
		}

		certFilePath := path.Join(thunderRuntime.ThunderHome, keyConfig.CertFile)
		keyFilePath := path.Join(thunderRuntime.ThunderHome, keyConfig.KeyFile)

		// Check if the certificate and key files exist.
		if _, err := os.Stat(certFilePath); os.IsNotExist(err) {
			return nil, errors.New("certificate file not found at " + certFilePath)
		}
		if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
			return nil, errors.New("key file not found at " + keyFilePath)
		}

		// Load the certificate and key.
		tlsCert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
		if err != nil {
			return nil, err
		}
		algorithm, err := getAlgorithmFromKey(tlsCert.PrivateKey)
		if err != nil {
			return nil, err
		}
		thumbprint, err := getThumbprint(tlsCert)
		if err != nil {
			return nil, err
		}
		certificates[keyConfig.ID] = PKI{
			ID:          keyConfig.ID,
			Algorithm:   algorithm,
			PrivateKey:  tlsCert.PrivateKey,
			Certificate: tlsCert,
			ThumbPrint:  thumbprint,
		}
		logger.Debug("Loaded certificate",
			log.String("certFile", certFilePath),
			log.String("keyFile", keyFilePath),
			log.String("keyID", keyConfig.ID))
	}

	if len(certificates) == 0 {
		logger.Warn("No certificates were successfully loaded")
	}

	return &pkiService{
		certificates: certificates,
	}, nil
}

// GetPrivateKey retrieves the private key associated with the given ID.
func (s *pkiService) GetPrivateKey(id string) (crypto.PrivateKey, error) {
	cert, exists := s.certificates[id]
	if !exists {
		return nil, errors.New("certificate with ID " + id + " not found")
	}
	if cert.PrivateKey == nil {
		return nil, errors.New("no private key found for certificate with ID " + id)
	}
	return cert.PrivateKey, nil
}

// GetCertThumbprint retrieves the thumbprint of the certificate associated with the given ID.
func (s *pkiService) GetCertThumbprint(id string) string {
	cert, exists := s.certificates[id]
	if !exists {
		return ""
	}
	return cert.ThumbPrint
}

// GetX509Certificate retrieves the x509 certificate associated with the given ID.
func (s *pkiService) GetX509Certificate(id string) (*x509.Certificate, error) {
	cert, exists := s.certificates[id]
	if !exists {
		return nil, errors.New("certificate with ID " + id + " not found")
	}
	if len(cert.Certificate.Certificate) == 0 {
		return nil, errors.New("no certificate data found for certificate with ID " + id)
	}
	parsedCert, err := x509.ParseCertificate(cert.Certificate.Certificate[0])
	if err != nil {
		return nil, err
	}
	return parsedCert, nil
}

func getAlgorithmFromKey(key crypto.PrivateKey) (PKIAlgorithm, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return RSA, nil
	case *ecdsa.PrivateKey:
		// Determine ECDSA algorithm based on curve
		crvName := k.Curve.Params().Name
		switch crvName {
		case "P-256":
			return P256, nil
		case "P-384":
			return P384, nil
		case "P-521":
			return P521, nil
		default:
			return "", errors.New("unsupported ECDSA curve: " + crvName)
		}
	case ed25519.PrivateKey:
		return Ed25519, nil
	default:
		return "", errors.New("unsupported key type")
	}
}

func getThumbprint(cert tls.Certificate) (string, error) {
	certData := cert.Certificate[0]
	parsedCert, err := x509.ParseCertificate(certData)
	if err != nil {
		return "", err
	}

	x5tS256 := hash.GenerateThumbprint(parsedCert.Raw)
	return x5tS256, nil
}
