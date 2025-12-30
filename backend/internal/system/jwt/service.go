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
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/crypto/pki"
	"github.com/asgardeo/thunder/internal/system/crypto/sign"
	httpservice "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// JWTServiceInterface defines the interface for JWT operations.
type JWTServiceInterface interface {
	GetPublicKey() crypto.PublicKey
	GenerateJWT(sub, aud, iss string, validityPeriod int64, claims map[string]interface{}) (string, int64, error)
	VerifyJWT(jwtToken string, expectedAud, expectedIss string) error
	VerifyJWTWithPublicKey(jwtToken string, jwtPublicKey crypto.PublicKey, expectedAud, expectedIss string) error
	VerifyJWTWithJWKS(jwtToken, jwksURL, expectedAud, expectedIss string) error
	VerifyJWTSignature(jwtToken string) error
	VerifyJWTSignatureWithPublicKey(jwtToken string, jwtPublicKey crypto.PublicKey) error
	VerifyJWTSignatureWithJWKS(jwtToken string, jwksURL string) error
}

// JWTService implements the JWTServiceInterface for generating and managing JWT tokens.
type JWTService struct {
	privateKey crypto.PrivateKey
	signAlg    sign.SignAlgorithm
	jwsAlg     JWSAlgorithm
	kid        string
}

// GetJWTService returns a singleton instance of JWTService.
func newJWTService(pkiService pki.PKIServiceInterface) (JWTServiceInterface, error) {
	preferredKid := config.GetThunderRuntime().Config.JWT.PreferredKeyID

	privateKey, err := pkiService.GetPrivateKey(preferredKid)
	if err != nil {
		return nil, err
	}

	kid := pkiService.GetCertThumbprint(preferredKid)

	// Get algorithm based on the type of private key
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		return &JWTService{
			privateKey: k,
			signAlg:    sign.RSASHA256,
			jwsAlg:     RS256,
			kid:        kid,
		}, nil
	case *ecdsa.PrivateKey:
		// Determine ECDSA algorithm based on curve
		crvName := k.Curve.Params().Name
		switch crvName {
		case "P-256":
			return &JWTService{
				privateKey: k,
				signAlg:    sign.ECDSASHA256,
				jwsAlg:     ES256,
				kid:        kid,
			}, nil
		case "P-384":
			return &JWTService{
				privateKey: k,
				signAlg:    sign.ECDSASHA384,
				jwsAlg:     ES384,
				kid:        kid,
			}, nil
		case "P-521":
			return &JWTService{
				privateKey: k,
				signAlg:    sign.ECDSASHA512,
				jwsAlg:     ES512,
				kid:        kid,
			}, nil
		default:
			return nil, errors.New("unsupported EC curve: " + crvName + " only P-256, P-384 and P-521 are supported")
		}
	case ed25519.PrivateKey:
		return &JWTService{
			privateKey: k,
			signAlg:    sign.ED25519,
			jwsAlg:     EdDSA,
			kid:        kid,
		}, nil
	default:
		return nil, errors.New("unsupported private key type")
	}
}

// GetPublicKey returns the RSA public key corresponding to the server's private key.
func (js *JWTService) GetPublicKey() crypto.PublicKey {
	if js.privateKey == nil {
		return nil
	}
	switch k := js.privateKey.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public()
	default:
		return nil
	}
}

// GenerateJWT generates a standard JWT signed with the server's private key.
func (js *JWTService) GenerateJWT(sub, aud, iss string, validityPeriod int64, claims map[string]interface{}) (
	string, int64, error) {
	if js.privateKey == nil {
		return "", 0, errors.New("private key not loaded")
	}
	thunderRuntime := config.GetThunderRuntime()

	// Create the JWT header.
	header := map[string]string{
		"alg": string(js.jwsAlg),
		"typ": "JWT",
		"kid": js.kid,
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

	// Create the signing input and sign it with the private key.
	signingInput := headerBase64 + "." + payloadBase64
	signature, err := sign.Generate([]byte(signingInput), js.signAlg, js.privateKey)
	if err != nil {
		return "", 0, err
	}

	// Encode the signature in base64 URL format.
	signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureBase64, iat.Unix(), nil
}

// VerifyJWT verifies the JWT token using the server's public key.
func (js *JWTService) VerifyJWT(jwtToken string, expectedAud, expectedIss string) error {
	if js.privateKey == nil {
		return errors.New("private key not loaded")
	}

	// First verify signature using the configured server key and algorithm
	if err := js.VerifyJWTSignature(jwtToken); err != nil {
		return fmt.Errorf("invalid token signature: %w", err)
	}

	// Then verify claims
	return js.verifyJWTClaims(jwtToken, expectedAud, expectedIss)
}

// VerifyJWTWithPublicKey verifies the JWT token using the provided public key.
func (js *JWTService) VerifyJWTWithPublicKey(jwtToken string, jwtPublicKey crypto.PublicKey,
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

	// Derive public key from configured private key
	var pubKey crypto.PublicKey
	switch k := js.privateKey.(type) {
	case *rsa.PrivateKey:
		pubKey = &k.PublicKey
	case *ecdsa.PrivateKey:
		pubKey = &k.PublicKey
	case ed25519.PrivateKey:
		pubKey = k.Public()
	default:
		return errors.New("unsupported private key type for verification")
	}

	// Verify the signature using the configured algorithm
	return sign.Verify([]byte(signingInput), signature, js.signAlg, pubKey)
}

// VerifyJWTSignatureWithPublicKey verifies the signature of a JWT token using the provided public key.
func (js *JWTService) VerifyJWTSignatureWithPublicKey(jwtToken string, jwtPublicKey crypto.PublicKey) error {
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

	// Determine algorithm from JWT header
	header, err := DecodeJWTHeader(jwtToken)
	if err != nil {
		return fmt.Errorf("failed to parse JWT header: %w", err)
	}
	algStr, _ := header["alg"].(string)
	alg, err := mapJWSAlgToSignAlg(JWSAlgorithm(algStr))
	if err != nil {
		return err
	}

	// Verify the signature
	return sign.Verify([]byte(signingInput), signature, alg, jwtPublicKey)
}

// mapJWSAlgToSignAlg maps JWS alg header values to internal SignAlgorithm.
func mapJWSAlgToSignAlg(jwsAlg JWSAlgorithm) (sign.SignAlgorithm, error) {
	switch jwsAlg {
	case RS256:
		return sign.RSASHA256, nil
	case RS512:
		return sign.RSASHA512, nil
	case ES256:
		return sign.ECDSASHA256, nil
	case ES384:
		return sign.ECDSASHA384, nil
	case ES512:
		return sign.ECDSASHA512, nil
	case EdDSA:
		return sign.ED25519, nil
	default:
		return "", fmt.Errorf("unsupported JWS alg: %s", jwsAlg)
	}
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

	// Convert JWK to public key
	pubKey, err := jwkToPublicKey(jwk)
	if err != nil {
		return fmt.Errorf("failed to convert JWK to public key: %w", err)
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
