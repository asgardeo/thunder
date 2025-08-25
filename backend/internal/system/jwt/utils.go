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

package jwt

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// DecodeJWT decodes a JWT string and returns its header and payload as maps.
func DecodeJWT(token string) (map[string]interface{}, map[string]interface{}, error) {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return nil, nil, errors.New("invalid JWT format")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, errors.New("failed to decode JWT header: " + err.Error())
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, errors.New("failed to decode JWT payload: " + err.Error())
	}

	header := make(map[string]interface{})
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, nil, errors.New("failed to unmarshal JWT header: " + err.Error())
	}
	payload := make(map[string]interface{})
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, nil, errors.New("failed to unmarshal JWT payload: " + err.Error())
	}

	return header, payload, nil
}

// DecodeJWTPayload decodes the payload of a JWT token and returns it as a map.
func DecodeJWTPayload(jwtToken string) (map[string]interface{}, error) {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWT token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims map[string]interface{}
	if err = json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT claims: %w", err)
	}

	return claims, nil
}

// DecodeJWTHeader decodes the header of a JWT token and returns it as a map.
func DecodeJWTHeader(jwtToken string) (map[string]interface{}, error) {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWT token format")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT header: %w", err)
	}

	var header map[string]interface{}
	if err = json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT header: %w", err)
	}

	return header, nil
}

// jwkToRSAPublicKey converts a JWK (JSON Web Key) to an RSA Public Key.
func jwkToRSAPublicKey(jwk map[string]interface{}) (*rsa.PublicKey, error) {
	// Extract modulus (n) and exponent (e) from JWK
	n, ok := jwk["n"].(string)
	if !ok {
		return nil, errors.New("invalid JWK format, missing 'n' parameter")
	}

	e, ok := jwk["e"].(string)
	if !ok {
		return nil, errors.New("invalid JWK format, missing 'e' parameter")
	}

	// Decode modulus and exponent from Base64URL
	nBytes, err := base64.RawURLEncoding.DecodeString(n)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(e)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert exponent bytes to int
	var eInt int
	for i := 0; i < len(eBytes); i++ {
		eInt = eInt<<8 + int(eBytes[i])
	}

	// Create RSA public key
	pubKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}

	return pubKey, nil
}
