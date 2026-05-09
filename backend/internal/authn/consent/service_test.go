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

package consent

import (
	"github.com/asgardeo/thunder/internal/system/i18n/core"

	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	authnprovidercm "github.com/asgardeo/thunder/internal/authnprovider/common"
	"github.com/asgardeo/thunder/internal/consent"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/tests/mocks/consentmock"
	"github.com/asgardeo/thunder/tests/mocks/jose/jwtmock"
)

type ConsentEnforcerServiceTestSuite struct {
	suite.Suite
	mockConsentSvc *consentmock.ConsentServiceInterfaceMock
	mockJWTSvc     *jwtmock.JWTServiceInterfaceMock
	service        *consentEnforcerService
}

func TestConsentEnforcerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ConsentEnforcerServiceTestSuite))
}

func (s *ConsentEnforcerServiceTestSuite) SetupSuite() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://auth.example.com",
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeServerRuntime("/tmp/test", testConfig)
}

func (s *ConsentEnforcerServiceTestSuite) SetupTest() {
	s.mockConsentSvc = consentmock.NewConsentServiceInterfaceMock(s.T())
	s.mockJWTSvc = jwtmock.NewJWTServiceInterfaceMock(s.T())
	s.service = &consentEnforcerService{
		consentService: s.mockConsentSvc,
		jwtService:     s.mockJWTSvc,
		logger:         log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ConsentEnforcerService")),
	}
}

func (s *ConsentEnforcerServiceTestSuite) TestNewConsentEnforcerService() {
	svc := newConsentEnforcerService(s.mockConsentSvc, s.mockJWTSvc)
	s.NotNil(svc)
}

// ResolveConsent tests

func (s *ConsentEnforcerServiceTestSuite) TestResolveConsent_ConsentDisabled() {
	s.mockConsentSvc.On("IsEnabled").Return(false)

	result, svcErr := s.service.ResolveConsent(context.Background(), "ou1", "app1", "user1",
		[]string{"email"}, nil, nil)

	s.Nil(result)
	s.Nil(svcErr)
}

func (s *ConsentEnforcerServiceTestSuite) TestResolveConsent_ListPurposesClientError() {
	clientErr := &serviceerror.ServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "CONSENT-4001",
	}

	s.mockConsentSvc.On("IsEnabled").Return(true)
	s.mockConsentSvc.On("ListConsentPurposes", mock.Anything, "ou1", "app1").
		Return(nil, clientErr)

	result, svcErr := s.service.ResolveConsent(context.Background(), "ou1", "app1", "user1",
		[]string{"email"}, nil, nil)

	s.Nil(result)
	s.NotNil(svcErr)
	s.Equal(ErrorConsentPurposeFetchFailed.Code, svcErr.Code)
}

func (s *ConsentEnforcerServiceTestSuite) TestResolveConsent_ListPurposesServerError() {
	serverErr := &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "CONSENT-5001",
	}

	s.mockConsentSvc.On("IsEnabled").Return(true)
	s.mockConsentSvc.On("ListConsentPurposes", mock.Anything, "ou1", "app1").
		Return(nil, serverErr)

	result, svcErr := s.service.ResolveConsent(context.Background(), "ou1", "app1", "user1",
		[]string{"email"}, nil, nil)

	s.Nil(result)
	s.NotNil(svcErr)
	s.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (s *ConsentEnforcerServiceTestSuite) TestResolveConsent_NoPurposesConfigured() {
	s.mockConsentSvc.On("IsEnabled").Return(true)
	s.mockConsentSvc.On("ListConsentPurposes", mock.Anything, "ou1", "app1").
		Return([]consent.ConsentPurpose{}, nil)

	result, svcErr := s.service.ResolveConsent(context.Background(), "ou1", "app1", "user1",
		[]string{"email"}, nil, nil)

	s.Nil(result)
	s.Nil(svcErr)
}

func (s *ConsentEnforcerServiceTestSuite) TestResolveConsent_SearchConsentsClientError() {
	purposes := []consent.ConsentPurpose{
		{
			ID:   "purpose-1",
			Name: "app:app1:attrs",
			Elements: []consent.PurposeElement{
				{Name: "email", IsMandatory: true},
			},
		},
	}
	clientErr := &serviceerror.ServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "CONSENT-4002",
	}

	s.mockConsentSvc.On("IsEnabled").Return(true)
	s.mockConsentSvc.On("ListConsentPurposes", mock.Anything, "ou1", "app1").
		Return(purposes, nil)
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return(nil, clientErr)

	result, svcErr := s.service.ResolveConsent(context.Background(), "ou1", "app1", "user1",
		[]string{"email"}, nil, nil)

	s.Nil(result)
	s.NotNil(svcErr)
	s.Equal(ErrorConsentSearchFailed.Code, svcErr.Code)
}

func (s *ConsentEnforcerServiceTestSuite) TestResolveConsent_SearchConsentsServerError() {
	purposes := []consent.ConsentPurpose{
		{
			ID:   "purpose-1",
			Name: "app:app1:attrs",
			Elements: []consent.PurposeElement{
				{Name: "email", IsMandatory: true},
			},
		},
	}
	serverErr := &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "CONSENT-5002",
	}

	s.mockConsentSvc.On("IsEnabled").Return(true)
	s.mockConsentSvc.On("ListConsentPurposes", mock.Anything, "ou1", "app1").
		Return(purposes, nil)
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return(nil, serverErr)

	result, svcErr := s.service.ResolveConsent(context.Background(), "ou1", "app1", "user1",
		[]string{"email"}, nil, nil)

	s.Nil(result)
	s.NotNil(svcErr)
	s.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (s *ConsentEnforcerServiceTestSuite) TestResolveConsent_AllConsentsActive() {
	purposes := []consent.ConsentPurpose{
		{
			ID:   "purpose-1",
			Name: "app:app1:attrs",
			Elements: []consent.PurposeElement{
				{Name: "email", IsMandatory: true},
			},
		},
	}
	existingConsents := []consent.Consent{
		{
			ID:      "consent-1",
			GroupID: "app1",
			Purposes: []consent.ConsentPurposeItem{
				{
					Name: "app:app1:attrs",
					Elements: []consent.ConsentElementApproval{
						{Name: "email", IsUserApproved: true},
					},
				},
			},
		},
	}

	s.mockConsentSvc.On("IsEnabled").Return(true)
	s.mockConsentSvc.On("ListConsentPurposes", mock.Anything, "ou1", "app1").
		Return(purposes, nil)
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return(existingConsents, nil)

	result, svcErr := s.service.ResolveConsent(context.Background(), "ou1", "app1", "user1",
		[]string{"email"}, nil, nil)

	s.Nil(result)
	s.Nil(svcErr)
}

func (s *ConsentEnforcerServiceTestSuite) TestResolveConsent_PromptNeeded() {
	purposes := []consent.ConsentPurpose{
		{
			ID:          "purpose-1",
			Name:        "app:app1:attrs",
			Description: "Test purpose",
			Elements: []consent.PurposeElement{
				{Name: "email", IsMandatory: true},
				{Name: "phone", IsMandatory: false},
			},
		},
	}

	s.mockConsentSvc.On("IsEnabled").Return(true)
	s.mockConsentSvc.On("ListConsentPurposes", mock.Anything, "ou1", "app1").
		Return(purposes, nil)
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{}, nil)
	s.mockJWTSvc.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("test-session-token", int64(0), nil)

	result, svcErr := s.service.ResolveConsent(context.Background(), "ou1", "app1", "user1",
		[]string{"email"}, []string{"phone"}, nil)

	s.Nil(svcErr)
	s.NotNil(result)
	s.Len(result.Purposes, 1)
	s.Equal("app:app1:attrs", result.Purposes[0].PurposeName)
	s.Equal([]string{"email"}, result.Purposes[0].Essential)
	s.Equal([]string{"phone"}, result.Purposes[0].Optional)
	s.NotEmpty(result.SessionToken)
}

func (s *ConsentEnforcerServiceTestSuite) TestResolveConsent_RequiredAttributesFilter() {
	purposes := []consent.ConsentPurpose{
		{
			ID:   "purpose-1",
			Name: "app:app1:attrs",
			Elements: []consent.PurposeElement{
				{Name: "email", IsMandatory: true},
				{Name: "phone", IsMandatory: false},
				{Name: "address", IsMandatory: false},
			},
		},
	}

	s.mockConsentSvc.On("IsEnabled").Return(true)
	s.mockConsentSvc.On("ListConsentPurposes", mock.Anything, "ou1", "app1").
		Return(purposes, nil)
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{}, nil)
	s.mockJWTSvc.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("test-session-token", int64(0), nil)

	// Only request "email" — "phone" and "address" should be filtered out
	result, svcErr := s.service.ResolveConsent(context.Background(), "ou1", "app1", "user1",
		[]string{"email"}, nil, nil)

	s.Nil(svcErr)
	s.NotNil(result)
	s.Len(result.Purposes, 1)
	s.Equal([]string{"email"}, result.Purposes[0].Essential)
	s.Empty(result.Purposes[0].Optional)
}

func (s *ConsentEnforcerServiceTestSuite) TestResolveConsent_UserProfileFilter() {
	purposes := []consent.ConsentPurpose{
		{
			ID:   "purpose-1",
			Name: "app:app1:attrs",
			Elements: []consent.PurposeElement{
				{Name: "email", IsMandatory: true},
				{Name: "phone", IsMandatory: false},
			},
		},
	}
	availableAttributes := &authnprovidercm.AttributesResponse{
		Attributes: map[string]*authnprovidercm.AttributeResponse{
			"email": {},
		},
	}

	s.mockConsentSvc.On("IsEnabled").Return(true)
	s.mockConsentSvc.On("ListConsentPurposes", mock.Anything, "ou1", "app1").
		Return(purposes, nil)
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{}, nil)
	s.mockJWTSvc.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("test-session-token", int64(0), nil)

	result, svcErr := s.service.ResolveConsent(context.Background(), "ou1", "app1", "user1",
		nil, nil, availableAttributes)

	s.Nil(svcErr)
	s.NotNil(result)
	s.Len(result.Purposes, 1)
	s.Empty(result.Purposes[0].Essential)
	s.Equal([]string{"email"}, result.Purposes[0].Optional)
}

func (s *ConsentEnforcerServiceTestSuite) TestResolveConsent_PartialConsentsExist() {
	purposes := []consent.ConsentPurpose{
		{
			ID:   "purpose-1",
			Name: "app:app1:attrs",
			Elements: []consent.PurposeElement{
				{Name: "email", IsMandatory: true},
				{Name: "phone", IsMandatory: false},
			},
		},
	}
	existingConsents := []consent.Consent{
		{
			ID: "consent-1",
			Purposes: []consent.ConsentPurposeItem{
				{
					Name: "app:app1:attrs",
					Elements: []consent.ConsentElementApproval{
						{Name: "email", IsUserApproved: true},
					},
				},
			},
		},
	}

	s.mockConsentSvc.On("IsEnabled").Return(true)
	s.mockConsentSvc.On("ListConsentPurposes", mock.Anything, "ou1", "app1").
		Return(purposes, nil)
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return(existingConsents, nil)
	s.mockJWTSvc.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("test-session-token", int64(0), nil)

	result, svcErr := s.service.ResolveConsent(context.Background(), "ou1", "app1", "user1",
		[]string{"email"}, []string{"phone"}, nil)

	s.Nil(svcErr)
	s.NotNil(result)
	s.Len(result.Purposes, 1)
	s.Empty(result.Purposes[0].Essential)
	s.Equal([]string{"phone"}, result.Purposes[0].Optional)
}

// RecordConsent tests

// buildTestSessionToken creates a fake JWT with the given consent session payload embedded.
// The token is structured as a valid 3-part JWT so DecodeJWTPayload can parse it.
// VerifyJWT is mocked to pass, so the signature is a placeholder.
func buildTestSessionToken(purposes []consentSessionPurpose) string {
	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)

	sessionData := consentSessionData{Purposes: purposes}
	sessionJSON, _ := json.Marshal(sessionData)
	payload := map[string]interface{}{
		consentSessionClaimKey: json.RawMessage(sessionJSON),
	}
	payloadJSON, _ := json.Marshal(payload)

	return base64.RawURLEncoding.EncodeToString(headerJSON) + "." +
		base64.RawURLEncoding.EncodeToString(payloadJSON) + ".fake-sig"
}

func buildSessionTokenWithPayload(payload map[string]interface{}) string {
	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	return base64.RawURLEncoding.EncodeToString(headerJSON) + "." +
		base64.RawURLEncoding.EncodeToString(payloadJSON) + ".fake-sig"
}

func (s *ConsentEnforcerServiceTestSuite) TestResolveConsent_CreateConsentSessionTokenFails() {
	purposes := []consent.ConsentPurpose{
		{
			ID:   "purpose-1",
			Name: "app:app1:attrs",
			Elements: []consent.PurposeElement{
				{Name: "email", IsMandatory: true},
			},
		},
	}

	s.mockConsentSvc.On("IsEnabled").Return(true)
	s.mockConsentSvc.On("ListConsentPurposes", mock.Anything, "ou1", "app1").
		Return(purposes, nil)
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{}, nil)
	s.mockJWTSvc.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", int64(0), &serviceerror.ServiceError{
			Error: core.I18nMessage{Key: "error.test.jwt_generation_failed", DefaultValue: "JWT generation failed"},
		})

	result, svcErr := s.service.ResolveConsent(context.Background(), "ou1", "app1", "user1",
		[]string{"email"}, nil, nil)

	s.Nil(result)
	s.NotNil(svcErr)
	s.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (s *ConsentEnforcerServiceTestSuite) TestCreateConsentSessionToken_GenerateJWTFails() {
	promptData := &ConsentPromptData{
		Purposes: []ConsentPurposePrompt{{PurposeName: "purpose-1", Essential: []string{"email"}}},
	}

	s.mockJWTSvc.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", int64(0), &serviceerror.ServiceError{
			Error: core.I18nMessage{Key: "error.test.jwt_generation_failed", DefaultValue: "JWT generation failed"},
		})

	token, err := s.service.createConsentSessionToken(context.Background(), promptData)

	s.Empty(token)
	s.Error(err)
	s.Contains(err.Error(), "failed to generate consent session token")
}

func (s *ConsentEnforcerServiceTestSuite) TestVerifyAndDecodeConsentSession_DecodePayloadFails() {
	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	token := base64.RawURLEncoding.EncodeToString(headerJSON) + ".invalid-payload.signature"

	s.mockJWTSvc.On("VerifyJWT", token, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))

	result, err := s.service.verifyAndDecodeConsentSession(token)

	s.Nil(result)
	s.Error(err)
}

func (s *ConsentEnforcerServiceTestSuite) TestVerifyAndDecodeConsentSession_MissingClaim() {
	token := buildSessionTokenWithPayload(map[string]interface{}{"sub": "user1"})

	s.mockJWTSvc.On("VerifyJWT", token, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))

	result, err := s.service.verifyAndDecodeConsentSession(token)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "missing consent session claim")
}

func (s *ConsentEnforcerServiceTestSuite) TestVerifyAndDecodeConsentSession_InvalidClaimFormat() {
	token := buildSessionTokenWithPayload(map[string]interface{}{consentSessionClaimKey: "invalid"})

	s.mockJWTSvc.On("VerifyJWT", token, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))

	result, err := s.service.verifyAndDecodeConsentSession(token)

	s.Nil(result)
	s.Error(err)
}

func (s *ConsentEnforcerServiceTestSuite) TestRecordConsent_SessionTokenInvalid() {
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{PurposeName: "purpose1", Approved: true},
		},
	}
	s.mockJWTSvc.On("VerifyJWT", "bad-token", consentSessionTokenAudience, mock.Anything).
		Return(&serviceerror.ServiceError{
			Code:  "JWT-5001",
			Error: core.I18nMessage{Key: "error.test.invalid_token", DefaultValue: "Invalid token"},
		})

	result, svcErr := s.service.RecordConsent(context.Background(), "ou1", "app1", "user1",
		decisions, "bad-token", 0)

	s.Nil(result)
	s.NotNil(svcErr)
	s.Equal(ErrorConsentSessionInvalid.Code, svcErr.Code)
}

func (s *ConsentEnforcerServiceTestSuite) TestRecordConsent_MissingPurpose_TreatedAsDenied() {
	sessionToken := buildTestSessionToken([]consentSessionPurpose{
		{PurposeName: "purpose1", Essential: []string{"email"}},
		{PurposeName: "purpose2", Essential: []string{"phone"}},
	})
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{PurposeName: "purpose1", Approved: true, Elements: []ElementDecision{
				{Name: "email", Approved: true},
			}},
			// purpose2 is missing — should be filled in as denied
		},
	}
	createdConsent := &consent.Consent{
		ID: "consent-filled",
		Purposes: []consent.ConsentPurposeItem{
			{Name: "purpose1", Elements: []consent.ConsentElementApproval{
				{Name: "email", IsUserApproved: true},
			}},
			{Name: "purpose2", Elements: []consent.ConsentElementApproval{
				{Name: "phone", IsUserApproved: true}, // essential overridden to approved
			}},
		},
	}

	s.mockJWTSvc.On("VerifyJWT", sessionToken, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{}, nil)
	s.mockConsentSvc.On("CreateConsent", mock.Anything, "ou1",
		mock.MatchedBy(func(req *consent.ConsentRequest) bool {
			// Verify purpose2 was added with phone element
			for _, p := range req.Purposes {
				if p.Name == "purpose2" {
					for _, e := range p.Elements {
						if e.Name == "phone" {
							return true
						}
					}
				}
			}
			return false
		})).Return(createdConsent, nil)

	result, svcErr := s.service.RecordConsent(context.Background(), "ou1", "app1", "user1",
		decisions, sessionToken, 0)

	s.Nil(result)
	s.NotNil(svcErr)
	// HasEssentialDenial: phone (essential in purpose2) was implicitly denied
	s.Equal(ErrorEssentialConsentDenied.Code, svcErr.Code)
}

func (s *ConsentEnforcerServiceTestSuite) TestRecordConsent_SearchFails_ClientError() {
	sessionToken := buildTestSessionToken([]consentSessionPurpose{
		{PurposeName: "purpose1", Essential: []string{"email"}},
	})
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{PurposeName: "purpose1", Approved: true, Elements: []ElementDecision{
				{Name: "email", Approved: true},
			}},
		},
	}
	clientErr := &serviceerror.ServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "CONSENT-4002",
	}

	s.mockJWTSvc.On("VerifyJWT", sessionToken, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return(nil, clientErr)

	result, svcErr := s.service.RecordConsent(context.Background(), "ou1", "app1", "user1",
		decisions, sessionToken, 0)

	s.Nil(result)
	s.NotNil(svcErr)
	s.Equal(ErrorConsentSearchFailed.Code, svcErr.Code)
}

func (s *ConsentEnforcerServiceTestSuite) TestRecordConsent_SearchFails_ServerError() {
	sessionToken := buildTestSessionToken([]consentSessionPurpose{
		{PurposeName: "purpose1", Essential: []string{"email"}},
	})
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{PurposeName: "purpose1", Approved: true, Elements: []ElementDecision{
				{Name: "email", Approved: true},
			}},
		},
	}
	serverErr := &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "CONSENT-5002",
	}

	s.mockJWTSvc.On("VerifyJWT", sessionToken, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return(nil, serverErr)

	result, svcErr := s.service.RecordConsent(context.Background(), "ou1", "app1", "user1",
		decisions, sessionToken, 0)

	s.Nil(result)
	s.NotNil(svcErr)
	s.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (s *ConsentEnforcerServiceTestSuite) TestRecordConsent_NoExisting_CreateSuccess() {
	sessionToken := buildTestSessionToken([]consentSessionPurpose{
		{PurposeName: "app:app1:attrs", Essential: []string{"email"}},
	})
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{
				PurposeName: "app:app1:attrs",
				Approved:    true,
				Elements: []ElementDecision{
					{Name: "email", Approved: true},
				},
			},
		},
	}
	createdConsent := &consent.Consent{
		ID:      "consent-new",
		GroupID: "app1",
		Purposes: []consent.ConsentPurposeItem{
			{
				Name: "app:app1:attrs",
				Elements: []consent.ConsentElementApproval{
					{Name: "email", IsUserApproved: true},
				},
			},
		},
	}

	s.mockJWTSvc.On("VerifyJWT", sessionToken, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{}, nil)
	s.mockConsentSvc.On("CreateConsent", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentRequest")).Return(createdConsent, nil)

	result, svcErr := s.service.RecordConsent(context.Background(), "ou1", "app1", "user1",
		decisions, sessionToken, 0)

	s.Nil(svcErr)
	s.NotNil(result)
	s.Equal("consent-new", result.ID)
}

func (s *ConsentEnforcerServiceTestSuite) TestRecordConsent_NoExisting_CreateFails_ClientError() {
	sessionToken := buildTestSessionToken([]consentSessionPurpose{
		{PurposeName: "purpose1", Essential: []string{"email"}},
	})
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{PurposeName: "purpose1", Approved: true, Elements: []ElementDecision{
				{Name: "email", Approved: true},
			}},
		},
	}
	clientErr := &serviceerror.ServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "CONSENT-4003",
	}

	s.mockJWTSvc.On("VerifyJWT", sessionToken, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{}, nil)
	s.mockConsentSvc.On("CreateConsent", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentRequest")).Return(nil, clientErr)

	result, svcErr := s.service.RecordConsent(context.Background(), "ou1", "app1", "user1",
		decisions, sessionToken, 0)

	s.Nil(result)
	s.NotNil(svcErr)
	s.Equal(ErrorConsentCreateFailed.Code, svcErr.Code)
}

func (s *ConsentEnforcerServiceTestSuite) TestRecordConsent_NoExisting_CreateFails_ServerError() {
	sessionToken := buildTestSessionToken([]consentSessionPurpose{
		{PurposeName: "purpose1", Essential: []string{"email"}},
	})
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{PurposeName: "purpose1", Approved: true, Elements: []ElementDecision{
				{Name: "email", Approved: true},
			}},
		},
	}
	serverErr := &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "CONSENT-5003",
	}

	s.mockJWTSvc.On("VerifyJWT", sessionToken, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{}, nil)
	s.mockConsentSvc.On("CreateConsent", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentRequest")).Return(nil, serverErr)

	result, svcErr := s.service.RecordConsent(context.Background(), "ou1", "app1", "user1",
		decisions, sessionToken, 0)

	s.Nil(result)
	s.NotNil(svcErr)
	s.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (s *ConsentEnforcerServiceTestSuite) TestRecordConsent_ExistingConsent_UpdateSuccess() {
	sessionToken := buildTestSessionToken([]consentSessionPurpose{
		{PurposeName: "app:app1:attrs", Essential: []string{"email"}, Optional: []string{"phone"}},
	})
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{
				PurposeName: "app:app1:attrs",
				Approved:    true,
				Elements: []ElementDecision{
					{Name: "email", Approved: true},
					{Name: "phone", Approved: true},
				},
			},
		},
	}
	existingConsent := consent.Consent{
		ID:      "consent-existing",
		GroupID: "app1",
		Purposes: []consent.ConsentPurposeItem{
			{
				Name: "app:app1:attrs",
				Elements: []consent.ConsentElementApproval{
					{Name: "email", IsUserApproved: true},
				},
			},
		},
	}
	updatedConsent := &consent.Consent{
		ID:      "consent-existing",
		GroupID: "app1",
		Purposes: []consent.ConsentPurposeItem{
			{
				Name: "app:app1:attrs",
				Elements: []consent.ConsentElementApproval{
					{Name: "email", IsUserApproved: true},
					{Name: "phone", IsUserApproved: true},
				},
			},
		},
	}

	s.mockJWTSvc.On("VerifyJWT", sessionToken, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{existingConsent}, nil)
	s.mockConsentSvc.On("UpdateConsent", mock.Anything, "ou1", "consent-existing",
		mock.AnythingOfType("*consent.ConsentRequest")).Return(updatedConsent, nil)

	result, svcErr := s.service.RecordConsent(context.Background(), "ou1", "app1", "user1",
		decisions, sessionToken, 0)

	s.Nil(svcErr)
	s.NotNil(result)
	s.Equal("consent-existing", result.ID)
	s.Len(result.Purposes[0].Elements, 2)
}

func (s *ConsentEnforcerServiceTestSuite) TestRecordConsent_ExistingConsent_UpdateFails_ClientError() {
	sessionToken := buildTestSessionToken([]consentSessionPurpose{
		{PurposeName: "purpose1", Essential: []string{"email"}},
	})
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{PurposeName: "purpose1", Approved: true, Elements: []ElementDecision{
				{Name: "email", Approved: true},
			}},
		},
	}
	existingConsent := consent.Consent{ID: "consent-existing"}
	clientErr := &serviceerror.ServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "CONSENT-4004",
	}

	s.mockJWTSvc.On("VerifyJWT", sessionToken, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{existingConsent}, nil)
	s.mockConsentSvc.On("UpdateConsent", mock.Anything, "ou1", "consent-existing",
		mock.AnythingOfType("*consent.ConsentRequest")).Return(nil, clientErr)

	result, svcErr := s.service.RecordConsent(context.Background(), "ou1", "app1", "user1",
		decisions, sessionToken, 0)

	s.Nil(result)
	s.NotNil(svcErr)
	s.Equal(ErrorConsentUpdateFailed.Code, svcErr.Code)
}

func (s *ConsentEnforcerServiceTestSuite) TestRecordConsent_ExistingConsent_UpdateFails_ServerError() {
	sessionToken := buildTestSessionToken([]consentSessionPurpose{
		{PurposeName: "purpose1", Essential: []string{"email"}},
	})
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{PurposeName: "purpose1", Approved: true, Elements: []ElementDecision{
				{Name: "email", Approved: true},
			}},
		},
	}
	existingConsent := consent.Consent{ID: "consent-existing"}
	serverErr := &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "CONSENT-5004",
	}

	s.mockJWTSvc.On("VerifyJWT", sessionToken, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{existingConsent}, nil)
	s.mockConsentSvc.On("UpdateConsent", mock.Anything, "ou1", "consent-existing",
		mock.AnythingOfType("*consent.ConsentRequest")).Return(nil, serverErr)

	result, svcErr := s.service.RecordConsent(context.Background(), "ou1", "app1", "user1",
		decisions, sessionToken, 0)

	s.Nil(result)
	s.NotNil(svcErr)
	s.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}

func (s *ConsentEnforcerServiceTestSuite) TestRecordConsent_WithValidityPeriod() {
	sessionToken := buildTestSessionToken([]consentSessionPurpose{
		{PurposeName: "purpose1", Essential: []string{"email"}},
	})
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{PurposeName: "purpose1", Approved: true, Elements: []ElementDecision{
				{Name: "email", Approved: true},
			}},
		},
	}
	createdConsent := &consent.Consent{ID: "consent-timed"}

	s.mockJWTSvc.On("VerifyJWT", sessionToken, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{}, nil)
	s.mockConsentSvc.On("CreateConsent", mock.Anything, "ou1",
		mock.MatchedBy(func(req *consent.ConsentRequest) bool {
			return req.ValidityTime > 0
		})).Return(createdConsent, nil)

	result, svcErr := s.service.RecordConsent(context.Background(), "ou1", "app1", "user1",
		decisions, sessionToken, 3600)

	s.Nil(svcErr)
	s.NotNil(result)
	s.Equal("consent-timed", result.ID)
}

func (s *ConsentEnforcerServiceTestSuite) TestRecordConsent_ZeroValidityPeriod() {
	sessionToken := buildTestSessionToken([]consentSessionPurpose{
		{PurposeName: "purpose1", Essential: []string{"email"}},
	})
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{PurposeName: "purpose1", Approved: true, Elements: []ElementDecision{
				{Name: "email", Approved: true},
			}},
		},
	}
	createdConsent := &consent.Consent{ID: "consent-no-expiry"}

	s.mockJWTSvc.On("VerifyJWT", sessionToken, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{}, nil)
	s.mockConsentSvc.On("CreateConsent", mock.Anything, "ou1",
		mock.MatchedBy(func(req *consent.ConsentRequest) bool {
			return req.ValidityTime == 0
		})).Return(createdConsent, nil)

	result, svcErr := s.service.RecordConsent(context.Background(), "ou1", "app1", "user1",
		decisions, sessionToken, 0)

	s.Nil(svcErr)
	s.NotNil(result)
}

func (s *ConsentEnforcerServiceTestSuite) TestRecordConsent_EssentialDenied_ReturnsError() {
	sessionToken := buildTestSessionToken([]consentSessionPurpose{
		{PurposeName: "app:app1:attrs", Essential: []string{"email"}, Optional: []string{"phone"}},
	})
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{
				PurposeName: "app:app1:attrs",
				Approved:    true,
				Elements: []ElementDecision{
					{Name: "email", Approved: false}, // user denies essential
					{Name: "phone", Approved: false},
				},
			},
		},
	}
	// The consent record should reflect the user's actual decisions (email denied)
	createdConsent := &consent.Consent{
		ID:      "consent-essential-deny",
		GroupID: "app1",
		Purposes: []consent.ConsentPurposeItem{
			{
				Name: "app:app1:attrs",
				Elements: []consent.ConsentElementApproval{
					{Name: "email", IsUserApproved: false},
					{Name: "phone", IsUserApproved: false},
				},
			},
		},
	}

	s.mockJWTSvc.On("VerifyJWT", sessionToken, consentSessionTokenAudience, mock.Anything).
		Return((*serviceerror.ServiceError)(nil))
	s.mockConsentSvc.On("SearchConsents", mock.Anything, "ou1",
		mock.AnythingOfType("*consent.ConsentSearchFilter")).Return([]consent.Consent{}, nil)
	s.mockConsentSvc.On("CreateConsent", mock.Anything, "ou1",
		mock.MatchedBy(func(req *consent.ConsentRequest) bool {
			// Verify that email element is NOT overridden — stays denied
			for _, p := range req.Purposes {
				for _, e := range p.Elements {
					if e.Name == "email" {
						return !e.IsUserApproved
					}
				}
			}
			return false
		})).Return(createdConsent, nil)

	result, svcErr := s.service.RecordConsent(context.Background(), "ou1", "app1", "user1",
		decisions, sessionToken, 0)

	s.Nil(result)
	s.NotNil(svcErr)
	s.Equal(ErrorEssentialConsentDenied.Code, svcErr.Code)
	// Verify consent was still persisted (CreateConsent was called)
	s.mockConsentSvc.AssertCalled(s.T(), "CreateConsent", mock.Anything, "ou1", mock.Anything)
}

// buildConsentedElementSet tests

func (s *ConsentEnforcerServiceTestSuite) TestBuildConsentedElementSet_Empty() {
	result := buildConsentedElementSet([]consent.Consent{})
	s.Empty(result)
}

func (s *ConsentEnforcerServiceTestSuite) TestBuildConsentedElementSet_ApprovedElements() {
	consents := []consent.Consent{
		{
			Purposes: []consent.ConsentPurposeItem{
				{
					Name: "purpose1",
					Elements: []consent.ConsentElementApproval{
						{Name: "email", IsUserApproved: true},
						{Name: "phone", IsUserApproved: false},
					},
				},
			},
		},
	}

	result := buildConsentedElementSet(consents)

	s.True(result["purpose1:email"])
	s.False(result["purpose1:phone"])
}

func (s *ConsentEnforcerServiceTestSuite) TestBuildConsentedElementSet_MultipleConsents() {
	consents := []consent.Consent{
		{
			Purposes: []consent.ConsentPurposeItem{
				{Name: "purpose1", Elements: []consent.ConsentElementApproval{
					{Name: "email", IsUserApproved: true},
				}},
			},
		},
		{
			Purposes: []consent.ConsentPurposeItem{
				{Name: "purpose2", Elements: []consent.ConsentElementApproval{
					{Name: "phone", IsUserApproved: true},
				}},
			},
		},
	}

	result := buildConsentedElementSet(consents)

	s.True(result["purpose1:email"])
	s.True(result["purpose2:phone"])
	s.Len(result, 2)
}

// buildUserAttributeSet tests

func (s *ConsentEnforcerServiceTestSuite) TestBuildUserAttributeSet_Nil() {
	result := buildUserAttributeSet(nil)
	s.Nil(result)
}

func (s *ConsentEnforcerServiceTestSuite) TestBuildUserAttributeSet_Empty() {
	available := &authnprovidercm.AttributesResponse{
		Attributes: map[string]*authnprovidercm.AttributeResponse{},
	}

	result := buildUserAttributeSet(available)
	s.Nil(result)
}

func (s *ConsentEnforcerServiceTestSuite) TestBuildUserAttributeSet_WithAttributes() {
	available := &authnprovidercm.AttributesResponse{
		Attributes: map[string]*authnprovidercm.AttributeResponse{
			"email": {},
			"phone": {},
		},
	}

	result := buildUserAttributeSet(available)

	s.NotNil(result)
	s.True(result["email"])
	s.True(result["phone"])
	s.Len(result, 2)
}

// buildPurposePrompts tests

func (s *ConsentEnforcerServiceTestSuite) TestBuildPurposePrompts_AllNeedConsent() {
	purposes := []consent.ConsentPurpose{
		{
			ID:          "p1",
			Name:        "purpose1",
			Description: "Test purpose",
			Elements: []consent.PurposeElement{
				{Name: "email", IsMandatory: true},
				{Name: "phone", IsMandatory: false},
			},
		},
	}

	result := buildPurposePrompts(purposes, nil, nil, map[string]bool{}, nil)

	s.Len(result, 1)
	s.Equal("purpose1", result[0].PurposeName)
	s.Equal("p1", result[0].PurposeID)
	s.Equal("Test purpose", result[0].Description)
	s.Empty(result[0].Essential)
	s.Equal([]string{"email", "phone"}, result[0].Optional)
}

func (s *ConsentEnforcerServiceTestSuite) TestBuildPurposePrompts_AllAlreadyConsented() {
	purposes := []consent.ConsentPurpose{
		{
			Name: "purpose1",
			Elements: []consent.PurposeElement{
				{Name: "email", IsMandatory: true},
			},
		},
	}
	consentedElements := map[string]bool{"purpose1:email": true}

	result := buildPurposePrompts(purposes, nil, nil, consentedElements, nil)

	s.Empty(result)
}

func (s *ConsentEnforcerServiceTestSuite) TestBuildPurposePrompts_RequiredAttributesFilter() {
	purposes := []consent.ConsentPurpose{
		{
			Name: "purpose1",
			Elements: []consent.PurposeElement{
				{Name: "email", IsMandatory: true},
				{Name: "phone", IsMandatory: false},
			},
		},
	}

	result := buildPurposePrompts(purposes, []string{"email"}, nil, map[string]bool{}, nil)

	s.Len(result, 1)
	s.Equal([]string{"email"}, result[0].Essential)
	s.Empty(result[0].Optional)
}

func (s *ConsentEnforcerServiceTestSuite) TestBuildPurposePrompts_UserProfileFilter() {
	purposes := []consent.ConsentPurpose{
		{
			Name: "purpose1",
			Elements: []consent.PurposeElement{
				{Name: "email", IsMandatory: true},
				{Name: "phone", IsMandatory: false},
			},
		},
	}
	userAttributeSet := map[string]bool{"email": true}

	result := buildPurposePrompts(purposes, nil, nil, map[string]bool{}, userAttributeSet)

	s.Len(result, 1)
	s.Empty(result[0].Essential)
	s.Equal([]string{"email"}, result[0].Optional)
}

func (s *ConsentEnforcerServiceTestSuite) TestBuildPurposePrompts_NoMatchingElements() {
	purposes := []consent.ConsentPurpose{
		{
			Name: "purpose1",
			Elements: []consent.PurposeElement{
				{Name: "email", IsMandatory: true},
			},
		},
	}

	// email is filtered out by required attributes
	result := buildPurposePrompts(purposes, []string{"phone"}, nil, map[string]bool{}, nil)

	s.Empty(result)
}

// mergeConsentPurposes tests

func (s *ConsentEnforcerServiceTestSuite) TestMergeConsentPurposes_NoExisting() {
	incoming := []consent.ConsentPurposeItem{
		{
			Name: "purpose1",
			Elements: []consent.ConsentElementApproval{
				{Name: "email", IsUserApproved: true},
			},
		},
	}

	result := mergeConsentPurposes(nil, incoming)

	s.Len(result, 1)
	s.Equal("purpose1", result[0].Name)
}

func (s *ConsentEnforcerServiceTestSuite) TestMergeConsentPurposes_NewElementAddedToExistingPurpose() {
	existing := []consent.ConsentPurposeItem{
		{
			Name: "purpose1",
			Elements: []consent.ConsentElementApproval{
				{Name: "email", IsUserApproved: true},
			},
		},
	}
	incoming := []consent.ConsentPurposeItem{
		{
			Name: "purpose1",
			Elements: []consent.ConsentElementApproval{
				{Name: "phone", IsUserApproved: true},
			},
		},
	}

	result := mergeConsentPurposes(existing, incoming)

	s.Len(result, 1)
	s.Len(result[0].Elements, 2)
}

func (s *ConsentEnforcerServiceTestSuite) TestMergeConsentPurposes_NewDecisionOverridesExisting() {
	existing := []consent.ConsentPurposeItem{
		{
			Name: "purpose1",
			Elements: []consent.ConsentElementApproval{
				{Name: "email", IsUserApproved: false},
			},
		},
	}
	incoming := []consent.ConsentPurposeItem{
		{
			Name: "purpose1",
			Elements: []consent.ConsentElementApproval{
				{Name: "email", IsUserApproved: true},
			},
		},
	}

	result := mergeConsentPurposes(existing, incoming)

	s.Len(result, 1)
	s.Len(result[0].Elements, 1)
	s.True(result[0].Elements[0].IsUserApproved)
}

func (s *ConsentEnforcerServiceTestSuite) TestMergeConsentPurposes_ExistingPurposePreserved() {
	existing := []consent.ConsentPurposeItem{
		{Name: "purpose1", Elements: []consent.ConsentElementApproval{
			{Name: "email", IsUserApproved: true},
		}},
		{Name: "purpose2", Elements: []consent.ConsentElementApproval{
			{Name: "address", IsUserApproved: true},
		}},
	}
	incoming := []consent.ConsentPurposeItem{
		{Name: "purpose1", Elements: []consent.ConsentElementApproval{
			{Name: "email", IsUserApproved: true},
		}},
	}

	result := mergeConsentPurposes(existing, incoming)

	s.Len(result, 2)
	purposeNames := make([]string, 0, 2)
	for _, p := range result {
		purposeNames = append(purposeNames, p.Name)
	}
	s.Contains(purposeNames, "purpose1")
	s.Contains(purposeNames, "purpose2")
}

func (s *ConsentEnforcerServiceTestSuite) TestMergeConsentPurposes_NewPurposeAdded() {
	existing := []consent.ConsentPurposeItem{
		{Name: "purpose1", Elements: []consent.ConsentElementApproval{
			{Name: "email", IsUserApproved: true},
		}},
	}
	incoming := []consent.ConsentPurposeItem{
		{Name: "purpose1", Elements: []consent.ConsentElementApproval{
			{Name: "email", IsUserApproved: true},
		}},
		{Name: "purpose-new", Elements: []consent.ConsentElementApproval{
			{Name: "phone", IsUserApproved: true},
		}},
	}

	result := mergeConsentPurposes(existing, incoming)

	s.Len(result, 2)
	purposeNames := make([]string, 0, 2)
	for _, p := range result {
		purposeNames = append(purposeNames, p.Name)
	}
	s.Contains(purposeNames, "purpose1")
	s.Contains(purposeNames, "purpose-new")
}

// buildConsentElementApprovals tests

func (s *ConsentEnforcerServiceTestSuite) TestBuildConsentElementApprovals_Empty() {
	decisions := &ConsentDecisions{Purposes: []PurposeDecision{}}

	result := buildConsentElementApprovals(decisions)

	s.Empty(result)
}

func (s *ConsentEnforcerServiceTestSuite) TestBuildConsentElementApprovals_MultipleDecisions() {
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{
				PurposeName: "purpose1",
				Approved:    true,
				Elements: []ElementDecision{
					{Name: "email", Approved: true},
					{Name: "phone", Approved: false},
				},
			},
		},
	}

	result := buildConsentElementApprovals(decisions)

	s.Len(result, 1)
	s.Equal("purpose1", result[0].Name)
	s.Len(result[0].Elements, 2)
	s.True(result[0].Elements[0].IsUserApproved)
	s.False(result[0].Elements[1].IsUserApproved)
	s.Equal(consent.NamespaceAttribute, result[0].Elements[0].Namespace)
}

// purposeElementKey tests

func (s *ConsentEnforcerServiceTestSuite) TestPurposeElementKey() {
	key := purposeElementKey("purpose1", "email")
	s.Equal("purpose1:email", key)
}

// fillMissingDecisions tests

func (s *ConsentEnforcerServiceTestSuite) TestFillMissingDecisions_AllPresent() {
	session := &consentSessionData{
		Purposes: []consentSessionPurpose{
			{PurposeName: "p1", Essential: []string{"email"}},
		},
	}
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{PurposeName: "p1", Approved: true, Elements: []ElementDecision{{Name: "email", Approved: true}}},
		},
	}
	fillMissingDecisions(session, decisions)
	s.Len(decisions.Purposes, 1) // no change
}

func (s *ConsentEnforcerServiceTestSuite) TestFillMissingDecisions_MissingPurposeAdded() {
	session := &consentSessionData{
		Purposes: []consentSessionPurpose{
			{PurposeName: "p1", Essential: []string{"email"}},
			{PurposeName: "p2", Essential: []string{"phone"}, Optional: []string{"address"}},
		},
	}
	decisions := &ConsentDecisions{
		Purposes: []PurposeDecision{
			{PurposeName: "p1", Approved: true, Elements: []ElementDecision{{Name: "email", Approved: true}}},
		},
	}
	fillMissingDecisions(session, decisions)

	s.Len(decisions.Purposes, 2)
	added := decisions.Purposes[1]
	s.Equal("p2", added.PurposeName)
	s.False(added.Approved)
	s.Len(added.Elements, 2)
	s.Equal("phone", added.Elements[0].Name)
	s.False(added.Elements[0].Approved)
	s.Equal("address", added.Elements[1].Name)
	s.False(added.Elements[1].Approved)
}
