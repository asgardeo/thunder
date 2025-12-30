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
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
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

// jwkToPublicKey converts a JWK map to a crypto.PublicKey supporting RSA, EC, and Ed25519.
func jwkToPublicKey(jwk map[string]interface{}) (crypto.PublicKey, error) {
	kty, ok := jwk["kty"].(string)
	if !ok {
		return nil, errors.New("JWK missing kty")
	}

	switch kty {
	case "RSA":
		nStr, nOK := jwk["n"].(string)
		eStr, eOK := jwk["e"].(string)
		if !nOK || !eOK {
			return nil, errors.New("JWK missing RSA modulus or exponent")
		}

		nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode RSA modulus: %w", err)
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode RSA exponent: %w", err)
		}

		n := new(big.Int).SetBytes(nBytes)
		e := new(big.Int).SetBytes(eBytes).Int64()
		if e <= 0 {
			return nil, errors.New("invalid RSA exponent")
		}

		return &rsa.PublicKey{N: n, E: int(e)}, nil

	case "EC":
		crv, crvOK := jwk["crv"].(string)
		xStr, xOK := jwk["x"].(string)
		yStr, yOK := jwk["y"].(string)
		if !crvOK || !xOK || !yOK {
			return nil, errors.New("JWK missing EC parameters")
		}

		var curve elliptic.Curve
		var expectedKeySize int
		switch crv {
		case P256:
			curve = elliptic.P256()
			expectedKeySize = 32 // P-256 uses 256-bit (32-byte) coordinates
		case P384:
			curve = elliptic.P384()
			expectedKeySize = 48 // P-384 uses 384-bit (48-byte) coordinates
		case P521:
			curve = elliptic.P521()
			expectedKeySize = 66 // P-521 uses 521-bit (66-byte) coordinates
		default:
			return nil, fmt.Errorf("unsupported EC curve: %s", crv)
		}

		xBytes, err := base64.RawURLEncoding.DecodeString(xStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode EC x: %w", err)
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(yStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode EC y: %w", err)
		}

		// Validate coordinate byte lengths match the expected curve size
		if len(xBytes) != expectedKeySize || len(yBytes) != expectedKeySize {
			return nil, errors.New("invalid EC coordinate length")
		}

		x := new(big.Int).SetBytes(xBytes)
		y := new(big.Int).SetBytes(yBytes)
		if !curve.IsOnCurve(x, y) {
			return nil, errors.New("EC point not on curve")
		}

		return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil

	case "OKP":
		crv, crvOK := jwk["crv"].(string)
		xStr, xOK := jwk["x"].(string)
		if !crvOK || !xOK {
			return nil, errors.New("JWK missing OKP parameters")
		}

		switch crv {
		case "Ed25519":
			xBytes, err := base64.RawURLEncoding.DecodeString(xStr)
			if err != nil {
				return nil, fmt.Errorf("failed to decode Ed25519 x: %w", err)
			}
			if l := len(xBytes); l != ed25519.PublicKeySize {
				return nil, fmt.Errorf("invalid Ed25519 public key length: %d", l)
			}
			return ed25519.PublicKey(xBytes), nil
		default:
			return nil, fmt.Errorf("unsupported OKP curve: %s", crv)
		}
	default:
		return nil, fmt.Errorf("unsupported JWK kty: %s", kty)
	}
}
