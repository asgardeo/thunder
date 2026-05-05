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

package userinfo

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	certmodel "github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jose/jwe"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	appmodel "github.com/asgardeo/thunder/pkg/application/model"
	inboundmodel "github.com/asgardeo/thunder/pkg/inboundclient/model"
	"github.com/asgardeo/thunder/tests/mocks/httpmock"
	"github.com/asgardeo/thunder/tests/mocks/jose/jwemock"
	"github.com/asgardeo/thunder/tests/mocks/jose/jwtmock"
)

const testJWKSURI = "https://rp.example.com/jwks"

// JWEUserInfoTestSuite defines the test suite for JWE/JWS userinfo generation.
type JWEUserInfoTestSuite struct {
	suite.Suite
}

// TestJWEUserInfoSuite runs the JWE userinfo test suite.
func TestJWEUserInfoSuite(t *testing.T) {
	suite.Run(t, new(JWEUserInfoTestSuite))
}

func (s *JWEUserInfoTestSuite) SetupTest() {
	config.ResetServerRuntime()
	_ = config.InitializeServerRuntime("test-home", &config.Config{
		JWT: config.JWTConfig{Issuer: "test-issuer", ValidityPeriod: 600},
	})
}

func (s *JWEUserInfoTestSuite) TearDownTest() {
	config.ResetServerRuntime()
}

// TestGenerateJWEUserInfo_Success verifies a JWE response from an inline JWKS.
func (s *JWEUserInfoTestSuite) TestGenerateJWEUserInfo_Success() {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubJWKS := rsaPublicKeyToJWKS(&privateKey.PublicKey, "enc")

	mockJWE := jwemock.NewJWEServiceInterfaceMock(s.T())
	mockJWE.On("Encrypt",
		mock.Anything, mock.Anything,
		jwe.KeyEncAlgorithm("RSA-OAEP-256"),
		jwe.ContentEncAlgorithm("A256GCM"),
		"json",
		"",
	).Return("compact.jwe.token", (*serviceerror.ServiceError)(nil))

	svc := &userInfoService{jweService: mockJWE, logger: log.GetLogger()}
	cfg := &inboundmodel.UserInfoConfig{EncryptionAlg: "RSA-OAEP-256", EncryptionEnc: "A256GCM"}
	cert := &appmodel.ApplicationCertificate{Type: certmodel.CertificateTypeJWKS, Value: pubJWKS}

	result, svcErr := svc.generateJWEUserInfo(context.Background(), map[string]interface{}{"sub": "user1"}, cfg, cert)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), inboundmodel.UserInfoResponseTypeJWE, result.Type)
	assert.Equal(s.T(), "compact.jwe.token", result.JWTBody)
}

// TestGenerateJWEUserInfo_NoCert verifies missing cert returns server error.
func (s *JWEUserInfoTestSuite) TestGenerateJWEUserInfo_NoCert() {
	svc := &userInfoService{jweService: jwemock.NewJWEServiceInterfaceMock(s.T()), logger: log.GetLogger()}
	cfg := &inboundmodel.UserInfoConfig{EncryptionAlg: "RSA-OAEP-256", EncryptionEnc: "A256GCM"}

	result, svcErr := svc.generateJWEUserInfo(context.Background(), map[string]interface{}{"sub": "user1"}, cfg, nil)
	assert.Nil(s.T(), result)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), serviceerror.InternalServerError.Code, svcErr.Code)
}

// TestGenerateJWEUserInfo_EncryptFailure verifies JWE encryption failure returns server error.
func (s *JWEUserInfoTestSuite) TestGenerateJWEUserInfo_EncryptFailure() {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubJWKS := rsaPublicKeyToJWKS(&privateKey.PublicKey, "enc")

	mockJWE := jwemock.NewJWEServiceInterfaceMock(s.T())
	mockJWE.On("Encrypt",
		mock.Anything, mock.Anything,
		jwe.KeyEncAlgorithm("RSA-OAEP-256"),
		jwe.ContentEncAlgorithm("A256GCM"),
		"json",
		"",
	).Return("", &serviceerror.InternalServerError)

	svc := &userInfoService{jweService: mockJWE, logger: log.GetLogger()}
	cfg := &inboundmodel.UserInfoConfig{EncryptionAlg: "RSA-OAEP-256", EncryptionEnc: "A256GCM"}
	cert := &appmodel.ApplicationCertificate{Type: certmodel.CertificateTypeJWKS, Value: pubJWKS}

	result, svcErr := svc.generateJWEUserInfo(context.Background(), map[string]interface{}{"sub": "user1"}, cfg, cert)
	assert.Nil(s.T(), result)
	assert.NotNil(s.T(), svcErr)
}

// TestResolveRPPublicKey_InlineJWKS verifies key resolution from inline JWKS with use=enc.
func (s *JWEUserInfoTestSuite) TestResolveRPPublicKey_InlineJWKS() {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubJWKS := rsaPublicKeyToJWKS(&privateKey.PublicKey, "enc")

	svc := &userInfoService{logger: log.GetLogger()}
	cert := &appmodel.ApplicationCertificate{Type: certmodel.CertificateTypeJWKS, Value: pubJWKS}

	pub, _, svcErr := svc.resolveRPPublicKey(context.Background(), cert, "RSA-OAEP-256")
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), pub)
}

// TestResolveRPPublicKey_NoUseField verifies a key without 'use' field is rejected
// (RFC 7517 §4.2: keys without use must not be assumed encryption-capable).
func (s *JWEUserInfoTestSuite) TestResolveRPPublicKey_NoUseField() {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubJWKS := rsaPublicKeyToJWKS(&privateKey.PublicKey, "")

	svc := &userInfoService{logger: log.GetLogger()}
	cert := &appmodel.ApplicationCertificate{Type: certmodel.CertificateTypeJWKS, Value: pubJWKS}

	pub, _, svcErr := svc.resolveRPPublicKey(context.Background(), cert, "RSA-OAEP-256")
	assert.NotNil(s.T(), svcErr)
	assert.Nil(s.T(), pub)
}

// TestResolveRPPublicKey_JWKSURI verifies key resolution by fetching from a URI.
func (s *JWEUserInfoTestSuite) TestResolveRPPublicKey_JWKSURI() {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubJWKS := rsaPublicKeyToJWKS(&privateKey.PublicKey, "enc")

	mockHTTP := httpmock.NewHTTPClientInterfaceMock(s.T())
	mockHTTP.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == testJWKSURI && req.Method == http.MethodGet
	})).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(pubJWKS)),
		}, nil,
	)

	svc := &userInfoService{httpClient: mockHTTP, logger: log.GetLogger()}
	cert := &appmodel.ApplicationCertificate{
		Type:  certmodel.CertificateTypeJWKSURI,
		Value: testJWKSURI,
	}

	pub, _, svcErr := svc.resolveRPPublicKey(context.Background(), cert, "RSA-OAEP-256")
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), pub)
}

// TestResolveRPPublicKey_URIFetchError verifies HTTP error returns server error.
func (s *JWEUserInfoTestSuite) TestResolveRPPublicKey_URIFetchError() {
	mockHTTP := httpmock.NewHTTPClientInterfaceMock(s.T())
	mockHTTP.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == testJWKSURI
	})).Return(nil, errors.New("connection refused"))

	svc := &userInfoService{httpClient: mockHTTP, logger: log.GetLogger()}
	cert := &appmodel.ApplicationCertificate{
		Type:  certmodel.CertificateTypeJWKSURI,
		Value: testJWKSURI,
	}

	pub, _, svcErr := svc.resolveRPPublicKey(context.Background(), cert, "RSA-OAEP-256")
	assert.Nil(s.T(), pub)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), serviceerror.InternalServerError.Code, svcErr.Code)
}

// TestResolveRPPublicKey_JWKSURINon200 verifies that a non-200 HTTP response is rejected.
func (s *JWEUserInfoTestSuite) TestResolveRPPublicKey_JWKSURINon200() {
	mockHTTP := httpmock.NewHTTPClientInterfaceMock(s.T())
	mockHTTP.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == testJWKSURI
	})).Return(
		&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil,
	)

	svc := &userInfoService{httpClient: mockHTTP, logger: log.GetLogger()}
	cert := &appmodel.ApplicationCertificate{
		Type:  certmodel.CertificateTypeJWKSURI,
		Value: testJWKSURI,
	}

	pub, _, svcErr := svc.resolveRPPublicKey(context.Background(), cert, "RSA-OAEP-256")
	assert.Nil(s.T(), pub)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), serviceerror.InternalServerError.Code, svcErr.Code)
}

// TestResolveRPPublicKey_JWKSURITooLarge verifies that an oversized JWKS response is rejected
// specifically by the size gate and not by a JSON parse failure on the truncated body.
func (s *JWEUserInfoTestSuite) TestResolveRPPublicKey_JWKSURITooLarge() {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	validJWKS := rsaPublicKeyToJWKS(&privateKey.PublicKey, "enc")
	// Pad the valid JWKS with a large custom field so its total size exceeds 1 MB.
	// This ensures the test exercises the size-cap path, not a JSON parse error.
	padding := strings.Repeat("a", (1<<20)+1)
	oversizedBody := validJWKS[:len(validJWKS)-1] + `,"x-padding":"` + padding + `"}`
	mockHTTP := httpmock.NewHTTPClientInterfaceMock(s.T())
	mockHTTP.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == testJWKSURI
	})).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(oversizedBody)),
		}, nil,
	)

	svc := &userInfoService{httpClient: mockHTTP, logger: log.GetLogger()}
	cert := &appmodel.ApplicationCertificate{
		Type:  certmodel.CertificateTypeJWKSURI,
		Value: testJWKSURI,
	}

	pub, _, svcErr := svc.resolveRPPublicKey(context.Background(), cert, "RSA-OAEP-256")
	assert.Nil(s.T(), pub)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), serviceerror.InternalServerError.Code, svcErr.Code)
}

// TestResolveRPPublicKey_NoCert verifies nil cert returns server error.
func (s *JWEUserInfoTestSuite) TestResolveRPPublicKey_NoCert() {
	svc := &userInfoService{logger: log.GetLogger()}

	pub, _, svcErr := svc.resolveRPPublicKey(context.Background(), nil, "RSA-OAEP-256")
	assert.Nil(s.T(), pub)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), serviceerror.InternalServerError.Code, svcErr.Code)
}

// TestResolveRPPublicKey_SigOnlyKey verifies a sig-only JWKS returns server error.
func (s *JWEUserInfoTestSuite) TestResolveRPPublicKey_SigOnlyKey() {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubJWKS := rsaPublicKeyToJWKS(&privateKey.PublicKey, "sig")

	svc := &userInfoService{logger: log.GetLogger()}
	cert := &appmodel.ApplicationCertificate{Type: certmodel.CertificateTypeJWKS, Value: pubJWKS}

	pub, _, svcErr := svc.resolveRPPublicKey(context.Background(), cert, "RSA-OAEP-256")
	assert.Nil(s.T(), pub)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), serviceerror.InternalServerError.Code, svcErr.Code)
}

// TestGenerateNestedJWTUserInfo_Success verifies a sign-then-encrypt nested JWT.
func (s *JWEUserInfoTestSuite) TestGenerateNestedJWTUserInfo_Success() {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubJWKS := rsaPublicKeyToJWKS(&privateKey.PublicKey, "enc")

	mockJWT := jwtmock.NewJWTServiceInterfaceMock(s.T())
	mockJWT.On("GenerateJWT",
		"user1", "test-issuer", int64(600),
		mock.Anything, mock.Anything, "RS256",
	).Return("signed.jwt.token", int64(0), (*serviceerror.ServiceError)(nil))

	mockJWE := jwemock.NewJWEServiceInterfaceMock(s.T())
	mockJWE.On("Encrypt",
		mock.Anything, mock.Anything,
		jwe.KeyEncAlgorithm("RSA-OAEP-256"),
		jwe.ContentEncAlgorithm("A256GCM"),
		"JWT",
		"",
	).Return("nested.jwe.token", (*serviceerror.ServiceError)(nil))

	svc := &userInfoService{
		jwtService: mockJWT,
		jweService: mockJWE,
		logger:     log.GetLogger(),
	}

	cfg := &inboundmodel.UserInfoConfig{SigningAlg: "RS256", EncryptionAlg: "RSA-OAEP-256", EncryptionEnc: "A256GCM"}
	cert := &appmodel.ApplicationCertificate{Type: certmodel.CertificateTypeJWKS, Value: pubJWKS}

	result, svcErr := svc.generateNestedJWTUserInfo(
		context.Background(),
		"user1",
		map[string]interface{}{"client_id": "client1"},
		map[string]interface{}{"sub": "user1"},
		cfg,
		cert,
	)
	assert.Nil(s.T(), svcErr)
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), inboundmodel.UserInfoResponseTypeNESTEDJWT, result.Type)
	assert.Equal(s.T(), "nested.jwe.token", result.JWTBody)
}

// TestGenerateJWEUserInfo_EncryptErrorPropagated verifies that the exact error from Encrypt is returned,
// not a generic InternalServerError.
func (s *JWEUserInfoTestSuite) TestGenerateJWEUserInfo_EncryptErrorPropagated() {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubJWKS := rsaPublicKeyToJWKS(&privateKey.PublicKey, "enc")

	mockJWE := jwemock.NewJWEServiceInterfaceMock(s.T())
	unsupportedErr := &serviceerror.ServiceError{Code: "JWE-1003", Type: serviceerror.ClientErrorType}
	mockJWE.On("Encrypt",
		mock.Anything, mock.Anything,
		jwe.KeyEncAlgorithm("RSA-OAEP-256"),
		jwe.ContentEncAlgorithm("A256GCM"),
		"json",
		"",
	).Return("", unsupportedErr)

	svc := &userInfoService{jweService: mockJWE, logger: log.GetLogger()}
	cfg := &inboundmodel.UserInfoConfig{EncryptionAlg: "RSA-OAEP-256", EncryptionEnc: "A256GCM"}
	cert := &appmodel.ApplicationCertificate{Type: certmodel.CertificateTypeJWKS, Value: pubJWKS}

	result, svcErr := svc.generateJWEUserInfo(context.Background(), map[string]interface{}{"sub": "user1"}, cfg, cert)
	assert.Nil(s.T(), result)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), "JWE-1003", svcErr.Code)
}

// TestGenerateJWSUserInfo_UnsupportedAlg verifies that an algorithm incompatible with the server key
// returns InternalServerError (server misconfiguration, not a client auth error).
func (s *JWEUserInfoTestSuite) TestGenerateJWSUserInfo_UnsupportedAlg() {
	mockJWT := jwtmock.NewJWTServiceInterfaceMock(s.T())
	mockJWT.On("GenerateJWT",
		"user1", "test-issuer", int64(600),
		mock.Anything, mock.Anything, "ES256",
	).Return("", int64(0), &jwt.ErrorUnsupportedJWSAlgorithm)

	svc := &userInfoService{jwtService: mockJWT, logger: log.GetLogger()}
	cfg := &inboundmodel.UserInfoConfig{SigningAlg: "ES256"}

	result, svcErr := svc.generateJWSUserInfo(
		"user1",
		map[string]interface{}{"client_id": "client1"},
		map[string]interface{}{"sub": "user1"},
		cfg,
	)
	assert.Nil(s.T(), result)
	assert.NotNil(s.T(), svcErr)
	assert.Equal(s.T(), serviceerror.InternalServerError.Code, svcErr.Code)
}

// rsaPublicKeyToJWKS builds a minimal RSA JWKS JSON for tests.
// Pass use="" to omit the 'use' field.
func rsaPublicKeyToJWKS(pub *rsa.PublicKey, use string) string {
	eBytes := big.NewInt(int64(pub.E)).Bytes()
	key := map[string]interface{}{
		"kty": "RSA",
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(eBytes),
	}
	if use != "" {
		key["use"] = use
	}
	b, _ := json.Marshal(map[string]interface{}{"keys": []interface{}{key}})
	return string(b)
}
