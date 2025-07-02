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

// Package jwt provides functionality for generating and managing JWT tokens.
package jwt

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const defaultTokenValidity = 3600 // default validity period of 1 hour

var privateKey *rsa.PrivateKey

// LoadPrivateKey loads the private key from the specified file path in the configuration.
func LoadPrivateKey(cfg *config.Config, currentDirectory string) error {
	keyFilePath := path.Join(currentDirectory, cfg.Security.KeyFile)
	keyFilePath = filepath.Clean(keyFilePath)

	// Check if the key file exists.
	if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
		return errors.New("key file not found at " + keyFilePath)
	}

	// Read the key file.
	keyData, err := os.ReadFile(keyFilePath)
	if err != nil {
		return err
	}

	// Decode the PEM block.
	block, _ := pem.Decode(keyData)
	if block == nil {
		return errors.New("failed to decode PEM block containing private key")
	}

	// Handle PKCS1 and PKCS8 private keys.
	if block.Type == "RSA PRIVATE KEY" {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return err
		}
	} else if block.Type == "PRIVATE KEY" {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return err
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return errors.New("not an RSA private key")
		}
	} else {
		return errors.New("unsupported private key type: " + block.Type)
	}

	return nil
}

// GetPublicKey returns the RSA public key corresponding to the loaded private key.
func GetPublicKey() *rsa.PublicKey {
	if privateKey == nil {
		return nil
	}
	return &privateKey.PublicKey
}

// GenerateJWT generates a standard JWT signed with the server's private key.
func GenerateJWT(sub, aud string, validityPeriod int64, claims map[string]string) (string, int64, error) {
	if privateKey == nil {
		return "", 0, errors.New("private key not loaded")
	}

	config := config.GetThunderRuntime().Config

	// Create the JWT header.
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", 0, err
	}

	// Calculate the expiration time based on the validity period.
	if validityPeriod == 0 {
		validityPeriod = defaultTokenValidity
	}
	iat := time.Now()
	expirationTime := iat.Add(time.Duration(validityPeriod) * time.Second).Unix()

	// Create the JWT payload.
	payload := map[string]interface{}{
		"sub": sub,
		"iss": config.OAuth.JWT.Issuer,
		"aud": aud,
		"exp": expirationTime,
		"iat": iat.Unix(),
		"nbf": iat.Unix(),
		"jti": utils.GenerateUUID(),
	}

	// Add custom claims if provided.
	if len(claims) > 0 {
		for key, value := range claims {
			payload[key] = value
		}
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", 0, err
	}

	// Encode the header and payload in base64 URL format.
	headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create the signing input and hash it.
	signingInput := headerBase64 + "." + payloadBase64
	hashed := sha256.Sum256([]byte(signingInput))

	// Sign the hashed input with the private key.
	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", 0, err
	}

	// Encode the signature in base64 URL format.
	signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureBase64, iat.Unix(), nil
}
