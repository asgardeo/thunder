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
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/asgardeo/thunder/internal/system/config"
	httpservice "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

var (
	instance *JWTService
	once     sync.Once
)

// JWTServiceInterface defines the interface for JWT operations.
type JWTServiceInterface interface {
	Init() error
	GetPublicKey() *rsa.PublicKey
	GenerateJWT(sub, aud, iss string, validityPeriod int64, claims map[string]interface{}) (string, int64, error)
	VerifyJWT(jwtToken string, expectedAud, expectedIss string) error
	VerifyJWTWithPublicKey(jwtToken string, jwtPublicKey *rsa.PublicKey, expectedAud, expectedIss string) error
	VerifyJWTWithJWKS(jwtToken, jwksURL, expectedAud, expectedIss string) error
	VerifyJWTSignature(jwtToken string) error
	VerifyJWTSignatureWithPublicKey(jwtToken string, jwtPublicKey *rsa.PublicKey) error
	VerifyJWTSignatureWithJWKS(jwtToken string, jwksURL string) error
}

// JWTService implements the JWTServiceInterface for generating and managing JWT tokens.
type JWTService struct {
	privateKey *rsa.PrivateKey
}

// GetJWTService returns a singleton instance of JWTService.
func GetJWTService() JWTServiceInterface {
	once.Do(func() {
		instance = &JWTService{}
	})
	return instance
}

// Init loads the private key from the configured file path.
func (js *JWTService) Init() error {
	thunderRuntime := config.GetThunderRuntime()
	keyFilePath := path.Join(thunderRuntime.ThunderHome, thunderRuntime.Config.Security.KeyFile)
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
		js.privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return err
		}
	} else if block.Type == "PRIVATE KEY" {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return err
		}
		var ok bool
		js.privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return errors.New("not an RSA private key")
		}
	} else {
		return errors.New("unsupported private key type: " + block.Type)
	}

	return nil
}

// GetPublicKey returns the RSA public key corresponding to the server's private key.
func (js *JWTService) GetPublicKey() *rsa.PublicKey {
	if js.privateKey == nil {
		return nil
	}
	return &js.privateKey.PublicKey
}

// GenerateJWT generates a standard JWT signed with the server's private key.
func (js *JWTService) GenerateJWT(sub, aud, iss string, validityPeriod int64, claims map[string]interface{}) (
	string, int64, error) {
	if js.privateKey == nil {
		return "", 0, errors.New("private key not loaded")
	}

	thunderRuntime := config.GetThunderRuntime()

	// Get the certificate kid (Key ID) for the JWT header.
	kid := thunderRuntime.CertConfig.CertKid
	if kid == "" {
		return "", 0, errors.New("certificate Key ID (kid) not found")
	}

	// Create the JWT header.
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
		"kid": kid,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", 0, err
	}

	tokenIssuer := iss
	if tokenIssuer == "" {
		tokenIssuer = thunderRuntime.Config.JWT.Issuer
	}

	// Calculate the expiration time based on the validity period.
	if validityPeriod == 0 {
		validityPeriod = thunderRuntime.Config.JWT.ValidityPeriod
	}
	iat := time.Now()
	expirationTime := iat.Add(time.Duration(validityPeriod) * time.Second).Unix()

	// Create the JWT payload.
	payload := map[string]interface{}{
		"sub": sub,
		"iss": tokenIssuer,
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
	signature, err := rsa.SignPKCS1v15(nil, js.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", 0, err
	}

	// Encode the signature in base64 URL format.
	signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureBase64, iat.Unix(), nil
}

// VerifyJWT verifies the JWT token using the server's public key.
func (js *JWTService) VerifyJWT(jwtToken string, expectedAud, expectedIss string) error {
	publicKey := js.GetPublicKey()
	if publicKey == nil {
		return errors.New("public key not available")
	}
	return js.VerifyJWTWithPublicKey(jwtToken, publicKey, expectedAud, expectedIss)
}

// VerifyJWTWithPublicKey verifies the JWT token using the provided public key.
func (js *JWTService) VerifyJWTWithPublicKey(jwtToken string, jwtPublicKey *rsa.PublicKey,
	expectedAud, expectedIss string) error {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return errors.New("invalid JWT token format")
	}

	if err := js.VerifyJWTSignatureWithPublicKey(jwtToken, jwtPublicKey); err != nil {
		return fmt.Errorf("invalid token signature: %w", err)
	}

	return js.verifyJWTClaims(jwtToken, expectedAud, expectedIss)
}

// VerifyJWTWithJWKS verifies the JWT token using a JWK Set (JWKS) endpoint.
func (js *JWTService) VerifyJWTWithJWKS(jwtToken, jwksURL, expectedAud, expectedIss string) error {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return errors.New("invalid JWT token format")
	}

	if err := js.VerifyJWTSignatureWithJWKS(jwtToken, jwksURL); err != nil {
		return fmt.Errorf("invalid token signature: %w", err)
	}

	return js.verifyJWTClaims(jwtToken, expectedAud, expectedIss)
}

// VerifyJWTSignature verifies the signature of a JWT token using the server's public key.
func (js *JWTService) VerifyJWTSignature(jwtToken string) error {
	publicKey := js.GetPublicKey()
	if publicKey == nil {
		return errors.New("public key not available")
	}
	return js.VerifyJWTSignatureWithPublicKey(jwtToken, publicKey)
}

// VerifyJWTSignatureWithPublicKey verifies the signature of a JWT token using the provided public key.
func (js *JWTService) VerifyJWTSignatureWithPublicKey(jwtToken string, jwtPublicKey *rsa.PublicKey) error {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return errors.New("invalid JWT token format")
	}

	// Decode the signature
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("failed to decode JWT signature: %w", err)
	}

	// Create the signing input
	signingInput := parts[0] + "." + parts[1]

	// Hash the signing input
	hashed := sha256.Sum256([]byte(signingInput))

	// Verify the signature
	return rsa.VerifyPKCS1v15(jwtPublicKey, crypto.SHA256, hashed[:], signature)
}

// VerifyJWTSignatureWithJWKS verifies the signature of a JWT token using a JWK Set (JWKS) endpoint.
func (js *JWTService) VerifyJWTSignatureWithJWKS(jwtToken string, jwksURL string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "JWTService"))

	// Get the key ID from the JWT header
	header, err := DecodeJWTHeader(jwtToken)
	if err != nil {
		return fmt.Errorf("failed to parse JWT header: %w", err)
	}

	kid, ok := header["kid"].(string)
	if !ok {
		return fmt.Errorf("JWT header missing 'kid' claim")
	}

	// Fetch the JWK Set from the JWKS endpoint
	client := httpservice.NewHTTPClientWithTimeout(10 * time.Second)
	resp, err := client.Get(jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Error("Failed to close response body", log.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks struct {
		Keys []map[string]interface{} `json:"keys"`
	}
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("failed to parse JWKS: %w", err)
	}

	// Find the key with matching kid
	var jwk map[string]interface{}
	for _, key := range jwks.Keys {
		if keyID, ok := key["kid"].(string); ok && keyID == kid {
			jwk = key
			break
		}
	}
	if jwk == nil {
		return fmt.Errorf("no matching key found in JWKS")
	}

	// Convert JWK to RSA public key
	pubKey, err := jwkToRSAPublicKey(jwk)
	if err != nil {
		return fmt.Errorf("failed to convert JWK to RSA public key: %w", err)
	}

	// Verify JWT signature
	if err := js.VerifyJWTSignatureWithPublicKey(jwtToken, pubKey); err != nil {
		return fmt.Errorf("invalid token signature: %w", err)
	}

	return nil
}

// verifyJWTClaims verifies the standard claims of a JWT token.
func (js *JWTService) verifyJWTClaims(jwtToken string, expectedAud, expectedIss string) error {
	// Decode the JWT payload
	payload, err := DecodeJWTPayload(jwtToken)
	if err != nil {
		return fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	// Validate standard claims (exp, nbf, aud, iss)
	now := time.Now().Unix()

	if exp, ok := payload["exp"].(float64); ok {
		if now > int64(exp) {
			return errors.New("token has expired")
		}
	} else {
		return errors.New("missing or invalid 'exp' claim")
	}

	if nbf, ok := payload["nbf"].(float64); ok {
		if now < int64(nbf) {
			return errors.New("token not valid yet (nbf)")
		}
	} else {
		return errors.New("missing or invalid 'nbf' claim")
	}

	if expectedAud != "" {
		if aud, ok := payload["aud"].(string); ok {
			if aud != expectedAud {
				return fmt.Errorf("invalid audience: expected %s, got %s", expectedAud, aud)
			}
		} else {
			return errors.New("missing or invalid 'aud' claim")
		}
	}

	if expectedIss != "" {
		if iss, ok := payload["iss"].(string); ok {
			if iss != expectedIss {
				return fmt.Errorf("invalid issuer: expected %s, got %s", expectedIss, iss)
			}
		} else {
			return errors.New("missing or invalid 'iss' claim")
		}
	}

	return nil
}
