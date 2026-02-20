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

package tokenservice

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/jose/jwtmock"
)

const (
	testJWTTokenString = "test.jwt.token"     //nolint:gosec // Test token, not a real credential
	invalidJWTFormat   = "invalid.jwt.format" //nolint:gosec // Test token, not a real credential
	testClientID       = "client123"
)

type TokenValidatorTestSuite struct {
	suite.Suite
	mockJWTService *jwtmock.JWTServiceInterfaceMock
	validator      *tokenValidator
	oauthApp       *appmodel.OAuthAppConfigProcessedDTO
}

func TestTokenValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(TokenValidatorTestSuite))
}

func (suite *TokenValidatorTestSuite) SetupTest() {
	// Initialize Thunder Runtime for tests
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
			Audience:       "application", // Default audience for tests
			Leeway:         30,            // 30 seconds leeway for clock skew
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.validator = &tokenValidator{
		jwtService: suite.mockJWTService,
	}

	suite.oauthApp = &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
	}
}

// Helper function to create a test JWT token
func (suite *TokenValidatorTestSuite) createTestJWT(claims map[string]interface{}) string {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}

	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	return fmt.Sprintf("%s.%s.signature", headerB64, claimsB64)
}

// getDefaultAudience is a helper function to get the configured default audience from runtime.
// It skips the test if the runtime is not initialized or the audience is not configured.
func (suite *TokenValidatorTestSuite) getDefaultAudience() string {
	runtime := config.GetThunderRuntime()
	if runtime == nil {
		suite.T().Skip("ThunderRuntime not initialized")
		return ""
	}
	defaultAudience := runtime.Config.JWT.Audience
	if defaultAudience == "" {
		suite.T().Skip("Default audience not configured in runtime")
		return ""
	}
	return defaultAudience
}

// ============================================================================
// ValidateSubjectToken Tests - Success Cases
// ============================================================================

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Success_BasicToken() {
	defaultAudience := suite.getDefaultAudience()

	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":   "user123",
		"iss":   "https://thunder.io",
		"aud":   defaultAudience, // Use default audience for Thunder issuer
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"scope": "read write",
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user123", result.Sub)
	assert.Equal(suite.T(), "https://thunder.io", result.Iss)
	assert.Equal(suite.T(), []string{"read", "write"}, result.Scopes)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Success_WithTokenConfig() {
	// App with token config should still validate using Thunder-level issuer from config
	customOAuthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token:    &appmodel.OAuthTokenConfig{},
	}

	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"exp": float64(now + 3600),
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, customOAuthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "https://thunder.io", result.Iss)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Success_WithoutNbfClaim() {
	defaultAudience := suite.getDefaultAudience()

	// nbf is optional, should succeed without it
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"aud": defaultAudience, // Use default audience for Thunder issuer
		"exp": float64(now + 3600),
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Success_WithEmptyScopes() {
	defaultAudience := suite.getDefaultAudience()

	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"aud": defaultAudience, // Use default audience for Thunder issuer
		"exp": float64(now + 3600),
		// No scope claim
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Empty(suite.T(), result.Scopes)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Error_InvalidJWTFormat() {
	token := invalidJWTFormat

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to decode token")
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Error_MalformedJWT() {
	token := "not-a-jwt-at-all" //nolint:gosec // Test token, not a real credential

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to decode token")
}

// ============================================================================
// ValidateSubjectToken Tests - Issuer Validation Errors
// ============================================================================

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Error_MissingIssuerClaim() {
	claims := map[string]interface{}{
		"sub": "user123",
		// Missing iss claim
	}
	token := suite.createTestJWT(claims)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "missing 'iss' claim")
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Error_InvalidIssuerType() {
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": 12345, // Wrong type
	}
	token := suite.createTestJWT(claims)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "missing 'iss' claim")
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Error_UntrustedIssuer() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://evil-issuer.com",
		"exp": float64(now + 3600),
	}
	token := suite.createTestJWT(claims)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "not supported")
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Error_InvalidSignature() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"exp": float64(now + 3600),
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).
		Return(&serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "SIGNATURE_VERIFICATION_FAILED",
			Error:            "Signature verification failed",
			ErrorDescription: "The JWT signature verification failed",
		})

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid subject token signature")
	suite.mockJWTService.AssertExpectations(suite.T())
}

// ============================================================================
// ValidateSubjectToken Tests - Claims Validation Errors
// ============================================================================

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Error_MissingSubClaim() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"iss": "https://thunder.io",
		"exp": float64(now + 3600),
		// Missing sub claim
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "missing or invalid 'sub' claim")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Error_InvalidSubType() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": 12345, // Wrong type
		"iss": "https://thunder.io",
		"exp": float64(now + 3600),
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "missing or invalid 'sub' claim")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Error_ExpiredToken() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"exp": float64(now - 3600), // Expired
		"nbf": float64(now - 7200),
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "token has expired")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Error_NotYetValid() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"exp": float64(now + 3600),
		"nbf": float64(now + 1800), // Not yet valid
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "token not yet valid")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestVerifyTokenSignatureByIssuer_Success_ThunderIssuer() {
	token := testJWTTokenString

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	err := suite.validator.verifyTokenSignatureByIssuer(token, "https://thunder.io", suite.oauthApp)

	assert.NoError(suite.T(), err)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestVerifyTokenSignatureByIssuer_Success_WithTokenConfig() {
	// App with token config should still use Thunder-level issuer for signature verification
	customApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.AccessTokenConfig{},
		},
	}
	token := testJWTTokenString

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	err := suite.validator.verifyTokenSignatureByIssuer(token, "https://thunder.io", customApp)

	assert.NoError(suite.T(), err)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestVerifyTokenSignatureByIssuer_Error_SignatureFailure() {
	token := testJWTTokenString

	suite.mockJWTService.On("VerifyJWTSignature", token).
		Return(&serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "SIGNATURE_MISMATCH",
			Error:            "Signature mismatch",
			ErrorDescription: "The JWT signature does not match",
		})

	err := suite.validator.verifyTokenSignatureByIssuer(token, "https://thunder.io", suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to verify token signature")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestVerifyTokenSignatureByIssuer_Error_ExternalIssuerNotSupported() {
	// External issuer (not in trusted Thunder issuers)
	token := testJWTTokenString

	err := suite.validator.verifyTokenSignatureByIssuer(token, "https://external-idp.com", suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "no verification method configured for issuer")
	assert.Contains(suite.T(), err.Error(), "https://external-idp.com")
}

func (suite *TokenValidatorTestSuite) TestFederationScenario_DecodeBeforeVerify() {
	defaultAudience := suite.getDefaultAudience()

	// This test verifies the decode-first approach for federation
	// Token with valid issuer should pass issuer check before signature verification
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"aud": defaultAudience, // Use default audience for Thunder issuer
		"exp": float64(now + 3600),
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	// Signature verification should be called AFTER issuer validation
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestFederationScenario_FailFastOnUntrustedIssuer() {
	// Token with untrusted issuer should fail before signature verification
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://untrusted-issuer.com",
		"exp": float64(now + 3600),
	}
	token := suite.createTestJWT(claims)

	// Should not call VerifyJWTSignature because issuer check fails first
	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "not supported")
	// VerifyJWTSignature should NOT have been called
	suite.mockJWTService.AssertNotCalled(suite.T(), "VerifyJWTSignature")
}

func (suite *TokenValidatorTestSuite) TestFederationScenario_OnlyThunderIssuerIsValid() {
	// Only the Thunder-level issuer from config is accepted; app-level issuers are no longer supported
	appWithTokenConfig := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.AccessTokenConfig{},
		},
	}

	now := time.Now().Unix()

	// Test token from Thunder issuer (matches config-level issuer)
	claimsValid := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"exp": float64(now + 3600),
	}
	tokenValid := suite.createTestJWT(claimsValid)
	suite.mockJWTService.On("VerifyJWTSignature", tokenValid).Return(nil)

	resultValid, errValid := suite.validator.ValidateSubjectToken(
		tokenValid, appWithTokenConfig)
	assert.NoError(suite.T(), errValid)
	assert.NotNil(suite.T(), resultValid)

	// Test token from unknown issuer (not in valid issuers - should fail)
	claimsInvalid := map[string]interface{}{
		"sub": "user456",
		"iss": "https://unknown-issuer.com",
		"exp": float64(now + 3600),
	}
	tokenInvalid := suite.createTestJWT(claimsInvalid)

	resultInvalid, errInvalid := suite.validator.ValidateSubjectToken(
		tokenInvalid, appWithTokenConfig)
	assert.Error(suite.T(), errInvalid)
	assert.Nil(suite.T(), resultInvalid)
	assert.Contains(suite.T(), errInvalid.Error(), "not supported")
	// VerifyJWTSignature should NOT have been called for untrusted issuer
	suite.mockJWTService.AssertNotCalled(suite.T(), "VerifyJWTSignature", tokenInvalid)

	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestFederationScenario_FutureExternalIssuerSupport() {
	// This test documents the intended behavior for future external issuer support
	// When JWKS support is added, verifyTokenSignatureByIssuer should use JWKS endpoint

	token := testJWTTokenString
	externalIssuer := "https://external-idp.com"

	// Currently returns error because no JWKS support yet
	err := suite.validator.verifyTokenSignatureByIssuer(token, externalIssuer, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "no verification method configured")

	// TODO: When JWKS support is added, this should:
	// 1. Fetch JWKS from external issuer's .well-known endpoint
	// 2. Verify signature using public key from JWKS
	// 3. Return nil on success
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Security_RejectsTokenWithoutExp() {
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		// Missing exp claim - security risk
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	// Should reject tokens without expiration
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_EdgeCase_VeryLongToken() {
	// Get the configured default audience from runtime
	runtime := config.GetThunderRuntime()
	if runtime == nil {
		suite.T().Skip("ThunderRuntime not initialized")
		return
	}
	defaultAudience := runtime.Config.JWT.Audience
	if defaultAudience == "" {
		suite.T().Skip("Default audience not configured in runtime")
		return
	}

	// Test with token containing large claims
	now := time.Now().Unix()
	largeClaims := map[string]interface{}{
		"sub":   "user123",
		"iss":   "https://thunder.io",
		"aud":   defaultAudience, // Use default audience for Thunder issuer
		"exp":   float64(now + 3600),
		"large": string(make([]byte, 10000)), // 10KB of data
	}
	token := suite.createTestJWT(largeClaims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Success_Basic() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":                          "test-client",
		"iss":                          "https://thunder.io",
		"aud":                          "test-client",
		"exp":                          float64(now + 3600),
		"iat":                          float64(now),
		"scope":                        "read write",
		"access_token_sub":             "user123",
		"access_token_aud":             testAppID,
		"grant_type":                   "authorization_code",
		"access_token_user_attributes": map[string]interface{}{"name": "John Doe"},
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user123", result.Sub)
	assert.Equal(suite.T(), testAppID, result.Aud)
	assert.Equal(suite.T(), "authorization_code", result.GrantType)
	assert.Equal(suite.T(), []string{"read", "write"}, result.Scopes)
	assert.Equal(suite.T(), map[string]interface{}{"name": "John Doe"}, result.UserAttributes)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Success_WithoutUserAttributes() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":              "test-client",
		"iss":              "https://thunder.io",
		"aud":              "test-client",
		"exp":              float64(now + 3600),
		"iat":              float64(now),
		"scope":            "read write",
		"access_token_sub": "user123",
		"access_token_aud": testAppID,
		"grant_type":       "authorization_code",
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), result.UserAttributes)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Success_EmptyScopes() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":              "test-client",
		"iss":              "https://thunder.io",
		"aud":              "test-client",
		"exp":              float64(now + 3600),
		"iat":              float64(now),
		"access_token_sub": "user123",
		"access_token_aud": testAppID,
		"grant_type":       "authorization_code",
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Empty(suite.T(), result.Scopes)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Error_InvalidSignature() {
	token := "invalid.token.signature"

	suite.mockJWTService.On("VerifyJWT", token, "", "").
		Return(&serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "SIGNATURE_VERIFICATION_FAILED",
			Error:            "Signature verification failed",
			ErrorDescription: "The JWT signature verification failed",
		})

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid refresh token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Error_InvalidJWTFormat() {
	token := invalidJWTFormat

	// VerifyJWT is called first and should fail for invalid format
	suite.mockJWTService.On("VerifyJWT", token, "", "").
		Return(&serviceerror.ServiceError{
			Type:             serviceerror.ClientErrorType,
			Code:             "INVALID_JWT_FORMAT",
			Error:            "Invalid JWT format",
			ErrorDescription: "The JWT format is invalid",
		})

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid refresh token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Error_DecodeFailure() {
	// Invalid base64 in payload
	//nolint:gosec // Test token, not a real credential
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.invalid-base64.signature"

	// VerifyJWT is called first and should fail for invalid base64
	suite.mockJWTService.On("VerifyJWT", token, "", "").
		Return(&serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "INVALID_JWT_SIGNATURE",
			Error:            "Invalid JWT signature",
			ErrorDescription: "The JWT signature is invalid",
		})

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid refresh token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Success_MissingIat() {
	// iat is optional per RFC 7519, so refresh token should work without it
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "test-client",
		"iss": "https://thunder.io",
		"aud": "test-client",
		"exp": float64(now + 3600),
		// Missing iat - should be allowed
		"access_token_sub": "user123",
		"access_token_aud": testAppID,
		"grant_type":       "authorization_code",
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user123", result.Sub)
	assert.Equal(suite.T(), int64(0), result.Iat) // iat should be 0 when missing
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Error_ExpiredToken() {
	// VerifyJWT validates exp claim, so it should return an error for expired tokens
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":              "test-client",
		"iss":              "https://thunder.io",
		"aud":              "test-client",
		"exp":              float64(now - 3600), // Expired
		"nbf":              float64(now - 7200), // Required by VerifyJWT
		"iat":              float64(now - 7200),
		"access_token_sub": "user123",
		"access_token_aud": testAppID,
		"grant_type":       "authorization_code",
	}
	token := suite.createTestJWT(claims)

	// VerifyJWT should catch expired tokens
	suite.mockJWTService.On("VerifyJWT", token, "", "").
		Return(&serviceerror.ServiceError{
			Type:             serviceerror.ClientErrorType,
			Code:             "TOKEN_EXPIRED",
			Error:            "Token has expired",
			ErrorDescription: "The token has expired",
		})

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid refresh token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Error_NotYetValid() {
	// VerifyJWT validates nbf claim, so it should return an error for not yet valid tokens
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":              "test-client",
		"iss":              "https://thunder.io",
		"aud":              "test-client",
		"exp":              float64(now + 3600),
		"nbf":              float64(now + 1800), // Not yet valid
		"iat":              float64(now),
		"access_token_sub": "user123",
		"access_token_aud": testAppID,
		"grant_type":       "authorization_code",
	}
	token := suite.createTestJWT(claims)

	// VerifyJWT should catch not yet valid tokens
	suite.mockJWTService.On("VerifyJWT", token, "", "").
		Return(&serviceerror.ServiceError{
			Type:             serviceerror.ClientErrorType,
			Code:             "TOKEN_NOT_VALID_YET",
			Error:            "Token not valid yet",
			ErrorDescription: "Token not valid yet (nbf)",
		})

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid refresh token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Error_MissingSub() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"iss":              "https://thunder.io",
		"aud":              "test-client",
		"exp":              float64(now + 3600),
		"iat":              float64(now),
		"access_token_sub": "user123",
		"access_token_aud": testAppID,
		"grant_type":       "authorization_code",
		// Missing sub
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "missing or invalid 'sub' claim")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Error_WrongClientID() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":              "wrong-client",
		"iss":              "https://thunder.io",
		"aud":              "wrong-client",
		"exp":              float64(now + 3600),
		"iat":              float64(now),
		"access_token_sub": "user123",
		"access_token_aud": testAppID,
		"grant_type":       "authorization_code",
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "refresh token does not belong to the requesting client")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Error_MissingAccessTokenSub() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":              "test-client",
		"iss":              "https://thunder.io",
		"aud":              "test-client",
		"exp":              float64(now + 3600),
		"iat":              float64(now),
		"access_token_aud": testAppID,
		"grant_type":       "authorization_code",
		// Missing access_token_sub
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "missing or invalid 'access_token_sub' claim")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Error_MissingAccessTokenAud() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":              "test-client",
		"iss":              "https://thunder.io",
		"aud":              "test-client",
		"exp":              float64(now + 3600),
		"iat":              float64(now),
		"access_token_sub": "user123",
		"grant_type":       "authorization_code",
		// Missing access_token_aud
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "missing or invalid 'access_token_aud' claim")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Error_MissingGrantType() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":              "test-client",
		"iss":              "https://thunder.io",
		"aud":              "test-client",
		"exp":              float64(now + 3600),
		"iat":              float64(now),
		"access_token_sub": "user123",
		"access_token_aud": testAppID,
		// Missing grant_type
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "missing or invalid 'grant_type' claim")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateRefreshToken_Success_WithClaimsLocales() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":                          "test-client",
		"iss":                          "https://thunder.io",
		"aud":                          "test-client",
		"exp":                          float64(now + 3600),
		"iat":                          float64(now),
		"scope":                        "read write",
		"access_token_sub":             "user123",
		"access_token_aud":             testAppID,
		"grant_type":                   "authorization_code",
		"access_token_user_attributes": map[string]interface{}{"name": "John Doe"},
		"access_token_claims_locales":  "en-US fr-CA ja",
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWT", token, "", "").Return(nil)

	result, err := suite.validator.ValidateRefreshToken(token, "test-client")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user123", result.Sub)
	assert.Equal(suite.T(), testAppID, result.Aud)
	assert.Equal(suite.T(), "authorization_code", result.GrantType)
	assert.Equal(suite.T(), []string{"read", "write"}, result.Scopes)
	assert.Equal(suite.T(), map[string]interface{}{"name": "John Doe"}, result.UserAttributes)
	assert.Equal(suite.T(), "en-US fr-CA ja", result.ClaimsLocales)
	suite.mockJWTService.AssertExpectations(suite.T())
}

// ============================================================================
// ValidateAuthAssertion Tests - Success Cases
// ============================================================================

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Success_WithAppID() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":                    "user123",
		"iss":                    "https://thunder.io",
		"aud":                    testAppID, // Matches client app_id
		"exp":                    float64(now + 3600),
		"nbf":                    float64(now - 60),
		"assurance":              map[string]interface{}{"aal": "AAL1", "ial": "IAL1"}, // Make it an auth assertion
		"authorized_permissions": "read:documents write:documents",
		"userType":               "person",
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID
	suite.oauthApp.ClientID = testClientID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user123", result.Sub)
	assert.Equal(suite.T(), "https://thunder.io", result.Iss)
	assert.Equal(suite.T(), testAppID, result.Aud)
	assert.Equal(suite.T(), []string{"read:documents", "write:documents"}, result.Scopes)
	assert.Equal(suite.T(), "person", result.UserAttributes["userType"])
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Success_WithEmptyScopes() {
	defaultAudience := suite.getDefaultAudience()

	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"aud": defaultAudience, // Use default audience
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
		// No authorized_permissions or scope
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user123", result.Sub)
	assert.Equal(suite.T(), []string{}, result.Scopes)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Success_WithScopeClaim() {
	defaultAudience := suite.getDefaultAudience()

	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":                    "user123",
		"iss":                    "https://thunder.io",
		"aud":                    defaultAudience, // Use default audience
		"exp":                    float64(now + 3600),
		"nbf":                    float64(now - 60),
		"scope":                  "openid profile", // Standard scope claim takes priority
		"authorized_permissions": "read:documents write:documents",
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), []string{"openid", "profile"}, result.Scopes) // scope claim takes priority
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Success_WithUserAttributes() {
	defaultAudience := suite.getDefaultAudience()

	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":                    "user123",
		"iss":                    "https://thunder.io",
		"aud":                    defaultAudience, // Use default audience
		"exp":                    float64(now + 3600),
		"nbf":                    float64(now - 60),
		"authorized_permissions": "read write",
		"userType":               "person",
		"email":                  "user@example.com",
		"customAttr":             "customValue",
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "person", result.UserAttributes["userType"])
	assert.Equal(suite.T(), "user@example.com", result.UserAttributes["email"])
	assert.Equal(suite.T(), "customValue", result.UserAttributes["customAttr"])
	suite.mockJWTService.AssertExpectations(suite.T())
}

// ============================================================================
// ValidateAuthAssertion Tests - Error Cases
// ============================================================================

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Error_MissingAudience() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		// Missing aud claim
		"exp":       float64(now + 3600),
		"nbf":       float64(now - 60),
		"assurance": map[string]interface{}{"aal": "AAL1", "ial": "IAL1"}, // Make it an auth assertion
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "missing 'aud' claim")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Error_AudienceMismatch() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":       "user123",
		"iss":       "https://thunder.io",
		"aud":       "different-app-id", // Doesn't match default audience or client app_id
		"exp":       float64(now + 3600),
		"nbf":       float64(now - 60),
		"assurance": map[string]interface{}{"aal": "AAL1", "ial": "IAL1"}, // Make it an auth assertion
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID
	suite.oauthApp.ClientID = testClientID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), "auth assertion audience mismatch", err.Error())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Success_WithDefaultAudience() {
	defaultAudience := suite.getDefaultAudience()

	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"aud": defaultAudience, // Matches configured default audience
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID // Different from audience
	suite.oauthApp.ClientID = testClientID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user123", result.Sub)
	assert.Equal(suite.T(), "https://thunder.io", result.Iss)
	assert.Equal(suite.T(), defaultAudience, result.Aud)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Error_InvalidJWTFormat() {
	token := invalidJWTFormat

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to decode token")
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Error_MissingIssuer() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		// Missing iss claim
		"aud": "app123",
		"exp": float64(now + 3600),
	}
	token := suite.createTestJWT(claims)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "missing 'iss' claim")
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Error_ExpiredToken() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"aud": "app123",
		"exp": float64(now - 3600), // Expired
		"nbf": float64(now - 7200),
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "token has expired")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Error_InvalidIssuer() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://invalid-issuer.com", // Not in valid issuers
		"aud": "app123",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "is not supported")
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Error_InvalidSignature() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"aud": "app123",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(&serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "INVALID_SIGNATURE",
		Error:            "Invalid signature",
		ErrorDescription: "The JWT signature is invalid",
	})

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid subject token signature")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Error_InvalidSubType() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": 12345, // Invalid type - should be string
		"iss": "https://thunder.io",
		"aud": "app123",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "missing or invalid 'sub' claim")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Error_TokenNotYetValid() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"aud": "app123",
		"exp": float64(now + 3600),
		"nbf": float64(now + 60), // Not yet valid - nbf is in the future
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "token not yet valid")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Error_MissingExp() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"aud": "app123",
		// Missing exp claim
		"nbf": float64(now - 60),
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "missing or invalid 'exp' claim")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Error_InvalidAudType() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":       "user123",
		"iss":       "https://thunder.io",
		"aud":       12345, // Invalid type - should be string
		"exp":       float64(now + 3600),
		"nbf":       float64(now - 60),
		"assurance": map[string]interface{}{"aal": "AAL1", "ial": "IAL1"}, // Make it an auth assertion
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "claim aud is not a string")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateAuthAssertion_Success_WithEmptyAuthorizedPermissions() {
	defaultAudience := suite.getDefaultAudience()

	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":                    "user123",
		"iss":                    "https://thunder.io",
		"aud":                    defaultAudience, // Use default audience
		"exp":                    float64(now + 3600),
		"nbf":                    float64(now - 60),
		"authorized_permissions": "", // Empty string
	}
	token := suite.createTestJWT(claims)

	suite.oauthApp.AppID = testAppID

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), []string{}, result.Scopes)
	suite.mockJWTService.AssertExpectations(suite.T())
}

// ============================================================================
// Leeway Tests - Time-based claim validation with clock skew tolerance
// ============================================================================

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Leeway_ExpiredWithinLeeway_ShouldPass() {
	defaultAudience := suite.getDefaultAudience()

	now := time.Now().Unix()
	// Token expired 10 seconds ago, but leeway is 30 seconds - should pass
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"aud": defaultAudience,
		"exp": float64(now - 10), // Expired 10 seconds ago
		"nbf": float64(now - 3600),
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user123", result.Sub)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Leeway_ExpiredBeyondLeeway_ShouldFail() {
	now := time.Now().Unix()
	// Token expired 60 seconds ago, leeway is 30 seconds - should fail
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"exp": float64(now - 60), // Expired 60 seconds ago
		"nbf": float64(now - 3600),
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "token has expired")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Leeway_NbfInFutureWithinLeeway_ShouldPass() {
	defaultAudience := suite.getDefaultAudience()

	now := time.Now().Unix()
	// Token nbf is 10 seconds in future, but leeway is 30 seconds - should pass
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"aud": defaultAudience,
		"exp": float64(now + 3600),
		"nbf": float64(now + 10), // Not valid for another 10 seconds
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user123", result.Sub)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Leeway_NbfInFutureBeyondLeeway_ShouldFail() {
	now := time.Now().Unix()
	// Token nbf is 60 seconds in future, leeway is 30 seconds - should fail
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"exp": float64(now + 3600),
		"nbf": float64(now + 60), // Not valid for another 60 seconds
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "token not yet valid")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Leeway_ExpirationBoundary_ShouldFail() {
	testCases := []struct {
		name      string
		leeway    int64
		expOffset int64 // offset from now in seconds (negative = expired)
		desc      string
	}{
		{
			name:      "ZeroLeeway_ExpiredOneSecondAgo",
			leeway:    0,
			expOffset: -1,
			desc:      "Token expired 1 second ago with zero leeway should fail",
		},
		{
			name:      "ExactlyAtBoundary",
			leeway:    30,
			expOffset: -30,
			desc:      "Token exp exactly at leeway boundary (now >= exp + leeway) should fail",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			config.ResetThunderRuntime()
			testConfig := &config.Config{
				JWT: config.JWTConfig{
					Issuer:         "https://thunder.io",
					ValidityPeriod: 3600,
					Audience:       "application",
					Leeway:         tc.leeway,
				},
			}
			_ = config.InitializeThunderRuntime("test", testConfig)

			now := time.Now().Unix()
			claims := map[string]interface{}{
				"sub": "user123",
				"iss": "https://thunder.io",
				"exp": float64(now + tc.expOffset),
				"nbf": float64(now - 3600),
			}
			token := suite.createTestJWT(claims)

			suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil).Once()

			result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

			assert.Error(suite.T(), err, tc.desc)
			assert.Nil(suite.T(), result)
			assert.Contains(suite.T(), err.Error(), "token has expired")
		})
	}
}

func (suite *TokenValidatorTestSuite) TestValidateSubjectToken_Leeway_ExpJustInsideBoundary_ShouldPass() {
	// Reset and test with 30 second leeway
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
			Audience:       "application",
			Leeway:         30, // 30 seconds leeway
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	defaultAudience := suite.getDefaultAudience()

	now := time.Now().Unix()
	// Token exp is just inside leeway boundary (now - 29 seconds)
	// Condition: now >= exp + leeway
	// = now >= (now - 29) + 30 = now >= now + 1 = FALSE (should pass)
	claims := map[string]interface{}{
		"sub": "user123",
		"iss": "https://thunder.io",
		"aud": defaultAudience,
		"exp": float64(now - 29), // Just inside boundary
		"nbf": float64(now - 3600),
	}
	token := suite.createTestJWT(claims)

	suite.mockJWTService.On("VerifyJWTSignature", token).Return(nil)

	result, err := suite.validator.ValidateSubjectToken(token, suite.oauthApp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user123", result.Sub)
	suite.mockJWTService.AssertExpectations(suite.T())
}
