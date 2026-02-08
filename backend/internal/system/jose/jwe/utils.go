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

package jwe

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	cryptohash "github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/jose/jws"
)

// encryptKey encrypts or derives the content encryption key (CEK) for a recipient.
// For ECDH-ES, the CEK is derived from the shared secret and written to the cek parameter
// slice in-place. For other algorithms, the CEK is treated as input and encrypted using
// the recipient's public key.
func encryptKey(cek []byte, alg KeyEncAlgorithm, recipientPubKey crypto.PublicKey,
	enc ContentEncAlgorithm) ([]byte, map[string]interface{}, error) {
	switch alg {
	case RSAOAEP256:
		encryptedKey, err := encryptWithRSAOAEP256(cek, recipientPubKey)
		return encryptedKey, nil, err

	case ECDHES:
		// For ECDH-ES, the CEK is directly derived from the shared secret
		return encryptWithECDHES(cek, recipientPubKey, enc)

	case ECDHESA128KW, ECDHESA256KW:
		// Derive KEK using ECDH-ES, then wrap CEK
		return encryptWithECDHESKW(cek, recipientPubKey, alg)

	default:
		return nil, nil, fmt.Errorf("unsupported JWE algorithm: %s", alg)
	}
}

// epkToMap converts an ephemeral public key to a JWK-like map representation.
func epkToMap(pub crypto.PublicKey) map[string]interface{} {
	ecdhPub, ok := pub.(*ecdh.PublicKey)
	if !ok {
		return nil
	}

	raw := ecdhPub.Bytes()
	var crv string
	var x, y []byte

	switch len(raw) {
	case 65:
		crv = jws.P256
		x = raw[1:33]
		y = raw[33:]
	case 97:
		crv = jws.P384
		x = raw[1:49]
		y = raw[49:]
	case 133:
		crv = jws.P521
		x = raw[1:67]
		y = raw[67:]
	default:
		return nil
	}

	return map[string]interface{}{
		"kty": "EC",
		"crv": crv,
		"x":   base64.RawURLEncoding.EncodeToString(x),
		"y":   base64.RawURLEncoding.EncodeToString(y),
	}
}

// decryptKey decrypts the encrypted content encryption key (CEK) using the recipient's private key.
func decryptKey(encryptedKey []byte, alg KeyEncAlgorithm, privateKey crypto.PrivateKey,
	header map[string]interface{}, enc ContentEncAlgorithm) ([]byte, error) {
	switch alg {
	case RSAOAEP256:
		return decryptWithRSAOAEP256(encryptedKey, privateKey)

	case ECDHES:
		return decryptWithECDHES(privateKey, header, enc)

	case ECDHESA128KW, ECDHESA256KW:
		return decryptWithECDHESKW(encryptedKey, privateKey, header, alg)

	default:
		return nil, fmt.Errorf("unsupported JWE algorithm: %s", alg)
	}
}

// computeSharedSecretForRecipient computes the ECDH shared secret Z from the recipient's private key.
func computeSharedSecretForRecipient(priv crypto.PrivateKey, pub crypto.PublicKey) ([]byte, error) {
	ecdsaPriv, ok := priv.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not an ECDSA key")
	}
	ecdhPriv, err := ecdsaPriv.ECDH()
	if err != nil {
		return nil, err
	}
	return computeSharedSecret(ecdhPriv, pub)
}

// encryptContent encrypts the payload using the content encryption key (CEK).
func encryptContent(payload []byte, cek []byte, enc ContentEncAlgorithm, aad []byte) ([]byte, []byte, []byte, error) {
	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, nil, nil, err
	}

	var gcm cipher.AEAD
	switch enc {
	case A128GCM, A192GCM, A256GCM:
		gcm, err = cipher.NewGCM(block)
	default:
		return nil, nil, nil, fmt.Errorf("unsupported encryption algorithm: %s", enc)
	}

	if err != nil {
		return nil, nil, nil, err
	}

	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, nil, nil, err
	}

	ciphertextWithTag := gcm.Seal(nil, iv, payload, aad)
	tagSize := gcm.Overhead()
	ciphertext := ciphertextWithTag[:len(ciphertextWithTag)-tagSize]
	tag := ciphertextWithTag[len(ciphertextWithTag)-tagSize:]

	return iv, ciphertext, tag, nil
}

// decryptContent decrypts the ciphertext using the content encryption key (CEK).
func decryptContent(ciphertext, iv, tag []byte, cek []byte, enc ContentEncAlgorithm, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, err
	}

	var gcm cipher.AEAD
	switch enc {
	case A128GCM, A192GCM, A256GCM:
		gcm, err = cipher.NewGCM(block)
	default:
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", enc)
	}

	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, iv, append(ciphertext, tag...), aad)
}

// DecodeJWE decodes a JWE compact serialization into its five parts.
func DecodeJWE(jweToken string) (header map[string]interface{}, headerBase64 string,
	encryptedKey, iv, ciphertext, tag []byte, err error) {
	parts := strings.Split(jweToken, ".")
	if len(parts) != 5 {
		return nil, "", nil, nil, nil, nil, errors.New("invalid JWE format")
	}

	headerBase64 = parts[0]
	headerBytes, err := base64.RawURLEncoding.DecodeString(headerBase64)
	if err != nil {
		return nil, "", nil, nil, nil, nil, fmt.Errorf("failed to decode header: %w", err)
	}

	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, "", nil, nil, nil, nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	encryptedKey, err = base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, "", nil, nil, nil, nil, fmt.Errorf("failed to decode encrypted key: %w", err)
	}

	iv, err = base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, "", nil, nil, nil, nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	ciphertext, err = base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil {
		return nil, "", nil, nil, nil, nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	tag, err = base64.RawURLEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, "", nil, nil, nil, nil, fmt.Errorf("failed to decode tag: %w", err)
	}

	return header, headerBase64, encryptedKey, iv, ciphertext, tag, nil
}

// concatKDF implements the Concat KDF function (RFC 7518 Section 4.6.2).
func concatKDF(z []byte, algID string, keyLen int) []byte {
	hasher, _ := cryptohash.GetHash(cryptohash.GenericSHA256)
	key := make([]byte, 0, keyLen)

	// SuppPubInfo is the key length in bits
	suppPubInfo := make([]byte, 4)
	binary.BigEndian.PutUint32(suppPubInfo, uint32(uint64(keyLen)*8)) // nolint:gosec // G115

	// OtherInfo = AlgorithmID || PartyUInfo || PartyVInfo || SuppPubInfo || SuppPrivInfo
	// For simplicity, we assume empty PartyUInfo, PartyVInfo and SuppPrivInfo as often done in JOSER
	algorithmID := lengthPrefixed([]byte(algID))
	partyUInfo := lengthPrefixed(nil)
	partyVInfo := lengthPrefixed(nil)
	suppPrivInfo := lengthPrefixed(nil)

	otherInfo := append(algorithmID, partyUInfo...) // nolint:gocritic
	otherInfo = append(otherInfo, partyVInfo...)
	otherInfo = append(otherInfo, suppPubInfo...)
	otherInfo = append(otherInfo, suppPrivInfo...)

	for counter := uint32(1); len(key) < keyLen; counter++ {
		hasher.Reset()
		counterBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(counterBuf, counter)

		hasher.Write(counterBuf)
		hasher.Write(z)
		hasher.Write(otherInfo)

		key = append(key, hasher.Sum(nil)...)
	}

	return key[:keyLen]
}

// lengthPrefixed returns the input data prefixed with its length as a 4-byte big-endian integer.
func lengthPrefixed(data []byte) []byte {
	res := make([]byte, 4+len(data))
	binary.BigEndian.PutUint32(res, uint32(uint64(len(data)))) // nolint:gosec // G115
	copy(res[4:], data)
	return res
}

// aesKeyWrap wraps the content encryption key (CEK) with the key encryption key (KEK).
// Implements RFC 3394 AES Key Wrap algorithm.
func aesKeyWrap(kek, cek []byte) ([]byte, error) {
	if len(cek)%8 != 0 {
		return nil, errors.New("CEK length must be a multiple of 8")
	}

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, err
	}

	n := len(cek) / 8
	r := make([]byte, (n+1)*8)
	copy(r[8:], cek)

	// Default IV for AES Key Wrap
	copy(r[:8], []byte{0xA6, 0xA6, 0xA6, 0xA6, 0xA6, 0xA6, 0xA6, 0xA6})

	for j := 0; j <= 5; j++ {
		for i := 1; i <= n; i++ {
			b := make([]byte, 16)
			copy(b[:8], r[:8])
			copy(b[8:], r[i*8:i*8+8])

			block.Encrypt(b, b)

			t := uint64(j)*uint64(n) + uint64(i) // nolint:gosec // G115
			for k := 0; k < 8; k++ {
				b[7-k] ^= byte(t >> (8 * k))
			}

			copy(r[:8], b[:8])
			copy(r[i*8:i*8+8], b[8:])
		}
	}

	return r, nil
}

// aesKeyUnwrap unwraps the wrapped key using the key encryption key (KEK).
// Implements RFC 3394 AES Key Wrap algorithm.
func aesKeyUnwrap(kek, wrapped []byte) ([]byte, error) {
	if len(wrapped)%8 != 0 || len(wrapped) < 16 {
		return nil, errors.New("invalid wrapped key length")
	}

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, err
	}

	n := (len(wrapped) / 8) - 1
	r := make([]byte, (n+1)*8)
	copy(r, wrapped)

	for j := 5; j >= 0; j-- {
		for i := n; i >= 1; i-- {
			t := uint64(j)*uint64(n) + uint64(i) // nolint:gosec // G115
			b := make([]byte, 16)
			copy(b[:8], r[:8])
			for k := 0; k < 8; k++ {
				b[7-k] ^= byte(t >> (8 * k))
			}
			copy(b[8:], r[i*8:i*8+8])

			block.Decrypt(b, b)

			copy(r[:8], b[:8])
			copy(r[i*8:i*8+8], b[8:])
		}
	}

	// Verify IV
	for i := 0; i < 8; i++ {
		if r[i] != 0xA6 {
			return nil, errors.New("IV mismatch during AES Key Unwrap")
		}
	}

	return r[8:], nil
}

// generateEphemeralKey generates an ephemeral EC key pair for the given public key's curve.
func generateEphemeralKey(recipientPubKey crypto.PublicKey) (crypto.PrivateKey, crypto.PublicKey, error) {
	ecdsaPub, ok := recipientPubKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, errors.New("recipient public key is not an ECDSA key")
	}

	var curve ecdh.Curve
	switch ecdsaPub.Curve.Params().Name {
	case jws.P256:
		curve = ecdh.P256()
	case jws.P384:
		curve = ecdh.P384()
	case jws.P521:
		curve = ecdh.P521()
	default:
		return nil, nil, fmt.Errorf("unsupported curve: %s", ecdsaPub.Curve.Params().Name)
	}

	priv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	pub := priv.PublicKey()
	return priv, pub, nil
}

// computeSharedSecret computes the ECDH shared secret Z.
func computeSharedSecret(privKey crypto.PrivateKey, pubKey crypto.PublicKey) ([]byte, error) {
	ecdhPriv, ok := privKey.(*ecdh.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not an ECDH private key")
	}

	var ecdhPub *ecdh.PublicKey
	switch p := pubKey.(type) {
	case *ecdh.PublicKey:
		ecdhPub = p
	case *ecdsa.PublicKey:
		var err error
		ecdhPub, err = p.ECDH()
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported public key type for ECDH")
	}

	return ecdhPriv.ECDH(ecdhPub)
}

// encryptWithRSAOAEP256 encrypts the CEK using RSA-OAEP-256 algorithm.
func encryptWithRSAOAEP256(cek []byte, recipientPubKey crypto.PublicKey) ([]byte, error) {
	rsaPub, ok := recipientPubKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("unsupported public key type for JWE key encryption")
	}
	h, err := cryptohash.GetHash(cryptohash.GenericSHA256)
	if err != nil {
		return nil, err
	}
	return rsa.EncryptOAEP(h, rand.Reader, rsaPub, cek, nil)
}

// encryptWithECDHES derives the CEK using ECDH-ES algorithm.
func encryptWithECDHES(cek []byte, recipientPubKey crypto.PublicKey,
	enc ContentEncAlgorithm) ([]byte, map[string]interface{}, error) {
	ephemeralPriv, ephemeralPub, err := generateEphemeralKey(recipientPubKey)
	if err != nil {
		return nil, nil, err
	}

	z, err := computeSharedSecret(ephemeralPriv, recipientPubKey)
	if err != nil {
		return nil, nil, err
	}

	// Derive key length based on enc
	keyLen := 0
	switch enc {
	case A128GCM:
		keyLen = 16
	case A192GCM:
		keyLen = 24
	case A256GCM:
		keyLen = 32
	default:
		return nil, nil, fmt.Errorf("unsupported encryption algorithm for ECDH-ES: %s", enc)
	}

	derivedKey := concatKDF(z, string(enc), keyLen)
	copy(cek, derivedKey)

	// Set epk in header
	headerExtras := map[string]interface{}{"epk": epkToMap(ephemeralPub)}
	return []byte{}, headerExtras, nil
}

// encryptWithECDHESKW encrypts the CEK using ECDH-ES with AES key wrap.
func encryptWithECDHESKW(cek []byte, recipientPubKey crypto.PublicKey,
	alg KeyEncAlgorithm) ([]byte, map[string]interface{}, error) {
	ephemeralPriv, ephemeralPub, err := generateEphemeralKey(recipientPubKey)
	if err != nil {
		return nil, nil, err
	}

	z, err := computeSharedSecret(ephemeralPriv, recipientPubKey)
	if err != nil {
		return nil, nil, err
	}

	kekLen := 16
	if alg == ECDHESA256KW {
		kekLen = 32
	}

	kek := concatKDF(z, string(alg), kekLen)
	wrappedKey, err := aesKeyWrap(kek, cek)
	if err != nil {
		return nil, nil, err
	}

	headerExtras := map[string]interface{}{"epk": epkToMap(ephemeralPub)}
	return wrappedKey, headerExtras, nil
}

// decryptWithRSAOAEP256 decrypts the encrypted key using RSA-OAEP-256 algorithm.
func decryptWithRSAOAEP256(encryptedKey []byte, privateKey crypto.PrivateKey) ([]byte, error) {
	rsaPriv, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("unsupported private key type for JWE key decryption")
	}
	h, err := cryptohash.GetHash(cryptohash.GenericSHA256)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptOAEP(h, rand.Reader, rsaPriv, encryptedKey, nil)
}

// decryptWithECDHES derives the CEK using ECDH-ES algorithm.
func decryptWithECDHES(privateKey crypto.PrivateKey, header map[string]interface{},
	enc ContentEncAlgorithm) ([]byte, error) {
	epkMap, ok := header["epk"].(map[string]interface{})
	if !ok {
		return nil, errors.New("missing epk in header")
	}
	ephemeralPub, err := jws.JWKToECPublicKey(epkMap)
	if err != nil {
		return nil, err
	}

	z, err := computeSharedSecretForRecipient(privateKey, ephemeralPub)
	if err != nil {
		return nil, err
	}

	keyLen := 0
	switch enc {
	case A128GCM:
		keyLen = 16
	case A192GCM:
		keyLen = 24
	case A256GCM:
		keyLen = 32
	default:
		return nil, fmt.Errorf("unsupported encryption algorithm for ECDH-ES: %s", enc)
	}

	return concatKDF(z, string(enc), keyLen), nil
}

// decryptWithECDHESKW decrypts the wrapped key using ECDH-ES with AES key wrap.
func decryptWithECDHESKW(encryptedKey []byte, privateKey crypto.PrivateKey,
	header map[string]interface{}, alg KeyEncAlgorithm) ([]byte, error) {
	epkMap, ok := header["epk"].(map[string]interface{})
	if !ok {
		return nil, errors.New("missing epk in header")
	}
	ephemeralPub, err := jws.JWKToECPublicKey(epkMap)
	if err != nil {
		return nil, err
	}

	z, err := computeSharedSecretForRecipient(privateKey, ephemeralPub)
	if err != nil {
		return nil, err
	}

	kekLen := 16
	if alg == ECDHESA256KW {
		kekLen = 32
	}

	kek := concatKDF(z, string(alg), kekLen)
	return aesKeyUnwrap(kek, encryptedKey)
}
