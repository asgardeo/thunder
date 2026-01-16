/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

// PKIServiceInterface defines the interface for certificate services.
type PKIServiceInterface interface {
	GetPrivateKey(id string) (crypto.PrivateKey, *serviceerror.ServiceError)
	GetCertThumbprint(id string) string
	GetX509Certificate(id string) (*x509.Certificate, *serviceerror.ServiceError)
}

// pkiService stores loaded certificates indexed by their ID
type pkiService struct {
	certificates map[string]PKI
	logger       *log.Logger
}

// newPKIService initializes and returns the cert service
func newPKIService() (PKIServiceInterface, error) {
	thunderRuntime := config.GetThunderRuntime()
	keyConfigs := thunderRuntime.Config.Crypto.Keys
	if len(keyConfigs) == 0 {
		return nil, errors.New("no key configurations found in the system configuration")
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
	}

	if len(certificates) == 0 {
		return nil, errors.New("no certificates loaded in PKI service")
	}

	return &pkiService{
		certificates: certificates,
		logger:       log.GetLogger().With(log.String(log.LoggerKeyComponentName, "PKIService")),
	}, nil
}

// GetPrivateKey retrieves the private key associated with the given ID.
func (s *pkiService) GetPrivateKey(id string) (crypto.PrivateKey, *serviceerror.ServiceError) {
	cert, exists := s.certificates[id]
	if !exists || cert.PrivateKey == nil {
		s.logger.Error("Private key not found for certificate ID: " + id)
		return nil, &serviceerror.InternalServerError
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
func (s *pkiService) GetX509Certificate(id string) (*x509.Certificate, *serviceerror.ServiceError) {
	cert, exists := s.certificates[id]
	if !exists {
		s.logger.Error("Certificate not found for certificate ID: " + id)
		return nil, &serviceerror.InternalServerError
	}
	if len(cert.Certificate.Certificate) == 0 {
		s.logger.Error("Certificate data is empty for certificate ID: " + id)
		return nil, &serviceerror.InternalServerError
	}
	parsedCert, err := x509.ParseCertificate(cert.Certificate.Certificate[0])
	if err != nil {
		s.logger.Error("Failed to parse x509 certificate for ID: " + id + " Error: " + err.Error())
		return nil, &serviceerror.InternalServerError
	}
	return parsedCert, nil
}

// getAlgorithmFromKey determines the PKIAlgorithm based on the type of the private key.
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

// getThumbprint computes the SHA-256 thumbprint of the given TLS certificate.
func getThumbprint(cert tls.Certificate) (string, error) {
	certData := cert.Certificate[0]
	parsedCert, err := x509.ParseCertificate(certData)
	if err != nil {
		return "", err
	}

	x5tS256 := hash.GenerateThumbprint(parsedCert.Raw)
	return x5tS256, nil
}
