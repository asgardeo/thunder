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
	"context"
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
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	httpservice "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// JWTServiceInterface defines the interface for JWT operations.
type JWTServiceInterface interface {
	GetPublicKey() crypto.PublicKey
	GenerateJWT(ctx context.Context, sub, aud, iss string, validityPeriod int64,
		claims map[string]interface{}) (string, int64, *serviceerror.ServiceError)
	VerifyJWT(ctx context.Context, jwtToken string, expectedAud, expectedIss string) *serviceerror.ServiceError
	VerifyJWTWithPublicKey(ctx context.Context, jwtToken string, jwtPublicKey crypto.PublicKey, expectedAud,
		expectedIss string) *serviceerror.ServiceError
	VerifyJWTWithJWKS(ctx context.Context, jwtToken, jwksURL, expectedAud,
		expectedIss string) *serviceerror.ServiceError
	VerifyJWTSignature(ctx context.Context, jwtToken string) *serviceerror.ServiceError
	VerifyJWTSignatureWithPublicKey(ctx context.Context, jwtToken string,
		jwtPublicKey crypto.PublicKey) *serviceerror.ServiceError
	VerifyJWTSignatureWithJWKS(ctx context.Context, jwtToken string,
		jwksURL string) *serviceerror.ServiceError
}

// jwtService implements the JWTServiceInterface for generating and managing JWT tokens.
type jwtService struct {
	privateKey crypto.PrivateKey
	signAlg    sign.SignAlgorithm
	jwsAlg     JWSAlgorithm
	kid        string
	logger     *log.Logger
}

// GetJWTService returns a singleton instance of JWTService.
func newJWTService(pkiService pki.PKIServiceInterface) (JWTServiceInterface, error) {
	preferredKid := config.GetThunderRuntime().Config.JWT.PreferredKeyID

	privateKey, err := pkiService.GetPrivateKey(preferredKid)
	if err != nil {
		return nil, errors.New("failed to retrieve private key for the key id: " + preferredKid)
	}

	kid := pkiService.GetCertThumbprint(preferredKid)
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "JWTService"))

	// Get algorithm based on the type of private key
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		return &jwtService{
			privateKey: k,
			signAlg:    sign.RSASHA256,
			jwsAlg:     RS256,
			kid:        kid,
			logger:     logger,
		}, nil
	case *ecdsa.PrivateKey:
		// Determine ECDSA algorithm based on curve
		crvName := k.Curve.Params().Name
		switch crvName {
		case "P-256":
			return &jwtService{
				privateKey: k,
				signAlg:    sign.ECDSASHA256,
				jwsAlg:     ES256,
				kid:        kid,
				logger:     logger,
			}, nil
		case "P-384":
			return &jwtService{
				privateKey: k,
				signAlg:    sign.ECDSASHA384,
				jwsAlg:     ES384,
				kid:        kid,
				logger:     logger,
			}, nil
		case "P-521":
			return &jwtService{
				privateKey: k,
				signAlg:    sign.ECDSASHA512,
				jwsAlg:     ES512,
				kid:        kid,
				logger:     logger,
			}, nil
		default:
			return nil, errors.New("unsupported EC curve: " + crvName + " only P-256, P-384 and P-521 are supported")
		}
	case ed25519.PrivateKey:
		return &jwtService{
			privateKey: k,
			signAlg:    sign.ED25519,
			jwsAlg:     EdDSA,
			kid:        kid,
			logger:     logger,
		}, nil
	default:
		return nil, errors.New("unsupported private key type")
	}
}

// GetPublicKey returns the RSA public key corresponding to the server's private key.
func (js *jwtService) GetPublicKey() crypto.PublicKey {
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
func (js *jwtService) GenerateJWT(ctx context.Context, sub, aud, iss string,
	validityPeriod int64, claims map[string]interface{}) (string, int64, *serviceerror.ServiceError) {
	if js.privateKey == nil {
		js.logger.Error("Private key not found for JWT generation")
		return "", 0, &serviceerror.InternalServerError
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
		js.logger.Error("Failed to marshal JWT header: " + err.Error())
		return "", 0, &serviceerror.InternalServerError
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
		js.logger.Error("Failed to marshal JWT payload: " + err.Error())
		return "", 0, &serviceerror.InternalServerError
	}

	// Encode the header and payload in base64 URL format.
	headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create the signing input and sign it with the private key.
	signingInput := headerBase64 + "." + payloadBase64
	signature, err := sign.Generate([]byte(signingInput), js.signAlg, js.privateKey)
	if err != nil {
		js.logger.Error("Failed to sign JWT: " + err.Error())
		return "", 0, &serviceerror.InternalServerError
	}

	// Encode the signature in base64 URL format.
	signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureBase64, iat.Unix(), nil
}

// VerifyJWT verifies the JWT token using the server's public key.
func (js *jwtService) VerifyJWT(ctx context.Context, jwtToken string,
	expectedAud, expectedIss string) *serviceerror.ServiceError {
	if js.privateKey == nil {
		js.logger.Error("Private key not found for JWT verification")
		return &serviceerror.InternalServerError
	}

	// First verify signature using the configured server key and algorithm
	if err := js.VerifyJWTSignature(ctx, jwtToken); err != nil {
		return &ErrorInvalidTokenSignature
	}

	// Then verify claims
	return js.verifyJWTClaims(jwtToken, expectedAud, expectedIss)
}

// VerifyJWTWithPublicKey verifies the JWT token using the provided public key.
func (js *jwtService) VerifyJWTWithPublicKey(ctx context.Context, jwtToken string, jwtPublicKey crypto.PublicKey,
	expectedAud, expectedIss string) *serviceerror.ServiceError {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return &ErrorInvalidJWTFormat
	}

	if err := js.VerifyJWTSignatureWithPublicKey(ctx, jwtToken, jwtPublicKey); err != nil {
		return err
	}

	return js.verifyJWTClaims(jwtToken, expectedAud, expectedIss)
}

// VerifyJWTWithJWKS verifies the JWT token using a JWK Set (JWKS) endpoint.
func (js *jwtService) VerifyJWTWithJWKS(ctx context.Context, jwtToken, jwksURL,
	expectedAud, expectedIss string) *serviceerror.ServiceError {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return &ErrorInvalidJWTFormat
	}

	if err := js.VerifyJWTSignatureWithJWKS(ctx, jwtToken, jwksURL); err != nil {
		return &ErrorInvalidTokenSignature
	}

	return js.verifyJWTClaims(jwtToken, expectedAud, expectedIss)
}

// VerifyJWTSignature verifies the signature of a JWT token using the server's public key.
func (js *jwtService) VerifyJWTSignature(ctx context.Context, jwtToken string) *serviceerror.ServiceError {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return &ErrorInvalidJWTFormat
	}

	// Decode the signature
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return &ErrorInvalidTokenSignature
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
		return &ErrorUnsupportedJWSAlgorithm
	}

	// Verify the signature using the configured algorithm
	err = sign.Verify([]byte(signingInput), signature, js.signAlg, pubKey)
	if err != nil {
		return &ErrorInvalidTokenSignature
	}
	return nil
}

// VerifyJWTSignatureWithPublicKey verifies the signature of a JWT token using the provided public key.
func (js *jwtService) VerifyJWTSignatureWithPublicKey(ctx context.Context, jwtToken string,
	jwtPublicKey crypto.PublicKey) *serviceerror.ServiceError {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return &ErrorInvalidJWTFormat
	}

	// Decode the signature
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return &ErrorInvalidTokenSignature
	}

	// Create the signing input
	signingInput := parts[0] + "." + parts[1]

	// Determine algorithm from JWT header
	header, err := DecodeJWTHeader(jwtToken)
	if err != nil {
		return &ErrorDecodingJWTHeader
	}
	algStr, _ := header["alg"].(string)
	alg, err := mapJWSAlgToSignAlg(JWSAlgorithm(algStr))
	if err != nil {
		return &ErrorUnsupportedJWSAlgorithm
	}

	// Verify the signature
	err = sign.Verify([]byte(signingInput), signature, alg, jwtPublicKey)
	if err != nil {
		return &ErrorInvalidTokenSignature
	}
	return nil
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
func (js *jwtService) VerifyJWTSignatureWithJWKS(ctx context.Context, jwtToken string,
	jwksURL string) *serviceerror.ServiceError {
	// Get the key ID from the JWT header
	header, err := DecodeJWTHeader(jwtToken)
	if err != nil {
		return &ErrorDecodingJWTHeader
	}

	kid, ok := header["kid"].(string)
	if !ok {
		return &ErrorDecodingJWTHeader
	}

	// Fetch the JWK Set from the JWKS endpoint
	client := httpservice.NewHTTPClientWithTimeout(10 * time.Second)
	resp, err := client.Get(jwksURL)
	if err != nil {
		js.logger.Debug("Failed to fetch JWKS from URL: " + err.Error())
		return &ErrorFailedToGetJWKS
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			js.logger.Error("Failed to close response body", log.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		js.logger.Debug("Failed to fetch JWKS, HTTP status: " + resp.Status)
		return &ErrorFailedToGetJWKS
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		js.logger.Debug("Failed to read JWKS response body: " + err.Error())
		return &ErrorFailedToParseJWKS
	}

	var jwks struct {
		Keys []map[string]interface{} `json:"keys"`
	}
	if err := json.Unmarshal(body, &jwks); err != nil {
		js.logger.Debug("Failed to parse JWKS JSON: " + err.Error())
		return &ErrorFailedToParseJWKS
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
		return &ErrorNoMatchingJWKFound
	}

	// Convert JWK to public key
	pubKey, err := jwkToPublicKey(jwk)
	if err != nil {
		js.logger.Debug("Failed to convert JWK to public key: " + err.Error())
		return &ErrorFailedToParseJWKS
	}

	// Verify JWT signature
	if err := js.VerifyJWTSignatureWithPublicKey(ctx, jwtToken, pubKey); err != nil {
		return err
	}

	return nil
}

// verifyJWTClaims verifies the standard claims of a JWT token.
func (js *jwtService) verifyJWTClaims(jwtToken string, expectedAud, expectedIss string) *serviceerror.ServiceError {
	// Decode the JWT payload
	payload, err := DecodeJWTPayload(jwtToken)
	if err != nil {
		js.logger.Debug("Failed to decode JWT payload: " + err.Error())
		return &ErrorDecodingJWTPayload
	}

	// Validate standard claims (exp, nbf, aud, iss)
	now := time.Now().Unix()

	if exp, ok := payload["exp"].(float64); ok {
		if now > int64(exp) {
			js.logger.Debug("JWT token has expired")
			return &ErrorTokenExpired
		}
	} else {
		js.logger.Debug("JWT token missing 'exp' claim or it is not a number")
		return &ErrorInvalidJWTFormat
	}

	if nbf, ok := payload["nbf"].(float64); ok {
		if now < int64(nbf) {
			js.logger.Debug("JWT token is not valid yet (nbf claim)")
			return &ErrorInvalidJWTFormat
		}
	} else {
		js.logger.Debug("JWT token missing 'nbf' claim or it is not a number")
		return &ErrorInvalidJWTFormat
	}

	if expectedAud != "" {
		if aud, ok := payload["aud"].(string); ok {
			if aud != expectedAud {
				js.logger.Debug("Invalid audience: expected " + expectedAud + ", got " + aud)
				return &ErrorInvalidJWTFormat
			}
		} else {
			js.logger.Debug("Missing 'aud' claim or it is not a string")
			return &ErrorInvalidJWTFormat
		}
	}

	if expectedIss != "" {
		if iss, ok := payload["iss"].(string); ok {
			if iss != expectedIss {
				js.logger.Debug("Invalid issuer: expected " + expectedIss + ", got " + iss)
				return &ErrorInvalidJWTFormat
			}
		} else {
			js.logger.Debug("Missing 'iss' claim or it is not a string")
			return &ErrorInvalidJWTFormat
		}
	}

	return nil
}
