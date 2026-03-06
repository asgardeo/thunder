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

package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/consent"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/certmock"
	"github.com/asgardeo/thunder/tests/mocks/consentmock"
)

type ApplicationServiceConsentTestSuite struct {
	suite.Suite
}

func TestApplicationServiceConsentTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationServiceConsentTestSuite))
}

// newTestApplicationServiceWithConsent creates a minimal applicationService with only the consentService field set.
func newTestApplicationServiceWithConsent(consentSvc consent.ConsentServiceInterface) *applicationService {
	return &applicationService{
		consentService: consentSvc,
	}
}

// ----- validateConsentConfig -----

func (s *ApplicationServiceConsentTestSuite) TestValidateConsentConfig_NilLoginConsent_SetsDefaults() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	app := &model.ApplicationDTO{}
	svcErr := svc.validateConsentConfig(app)

	s.Nil(svcErr)
	s.NotNil(app.LoginConsent)
	s.False(app.LoginConsent.Enabled)
	s.Equal(int64(0), app.LoginConsent.ValidityPeriod)
}

func (s *ApplicationServiceConsentTestSuite) TestValidateConsentConfig_EnabledWhenConsentServiceDisabled() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	cMock.EXPECT().IsEnabled().Return(false)
	svc := newTestApplicationServiceWithConsent(cMock)

	app := &model.ApplicationDTO{
		LoginConsent: &model.LoginConsentConfig{Enabled: true},
	}
	svcErr := svc.validateConsentConfig(app)

	s.NotNil(svcErr)
	s.Equal(ErrorConsentServiceNotEnabled.Code, svcErr.Code)
}

func (s *ApplicationServiceConsentTestSuite) TestValidateConsentConfig_EnabledWhenConsentServiceEnabled() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	cMock.EXPECT().IsEnabled().Return(true)
	svc := newTestApplicationServiceWithConsent(cMock)

	app := &model.ApplicationDTO{
		LoginConsent: &model.LoginConsentConfig{Enabled: true, ValidityPeriod: 3600},
	}
	svcErr := svc.validateConsentConfig(app)

	s.Nil(svcErr)
}

func (s *ApplicationServiceConsentTestSuite) TestValidateConsentConfig_NegativeValidityPeriodResetToZero() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	cMock.EXPECT().IsEnabled().Return(true)
	svc := newTestApplicationServiceWithConsent(cMock)

	app := &model.ApplicationDTO{
		LoginConsent: &model.LoginConsentConfig{Enabled: true, ValidityPeriod: -100},
	}
	svcErr := svc.validateConsentConfig(app)

	s.Nil(svcErr)
	s.Equal(int64(0), app.LoginConsent.ValidityPeriod)
}

func (s *ApplicationServiceConsentTestSuite) TestValidateConsentConfig_DisabledConsentWithoutConsentService() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	app := &model.ApplicationDTO{
		LoginConsent: &model.LoginConsentConfig{Enabled: false, ValidityPeriod: 100},
	}
	svcErr := svc.validateConsentConfig(app)

	s.Nil(svcErr)
}

// ----- extractRequestedAttributes -----

func (s *ApplicationServiceConsentTestSuite) TestExtractRequestedAttributes_NilApp_ReturnsNil() {
	result := extractRequestedAttributes(nil)

	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestExtractRequestedAttributes_NoAttributes_ReturnsEmpty() {
	app := &model.ApplicationProcessedDTO{
		ID:   "app-1",
		Name: "App",
	}

	result := extractRequestedAttributes(app)

	s.Empty(result)
}

func (s *ApplicationServiceConsentTestSuite) TestExtractRequestedAttributes_FromAssertion() {
	app := &model.ApplicationProcessedDTO{
		Assertion: &model.AssertionConfig{
			UserAttributes: []string{"email", "username"},
		},
	}

	result := extractRequestedAttributes(app)

	s.Len(result, 2)
	s.True(result["email"])
	s.True(result["username"])
}

func (s *ApplicationServiceConsentTestSuite) TestExtractRequestedAttributes_FromAccessToken() {
	app := &model.ApplicationProcessedDTO{
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
					Token: &model.OAuthTokenConfig{
						AccessToken: &model.AccessTokenConfig{
							UserAttributes: []string{"email"},
						},
					},
				},
			},
		},
	}

	result := extractRequestedAttributes(app)

	s.Len(result, 1)
	s.True(result["email"])
}

func (s *ApplicationServiceConsentTestSuite) TestExtractRequestedAttributes_FromIDToken() {
	app := &model.ApplicationProcessedDTO{
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
					Token: &model.OAuthTokenConfig{
						IDToken: &model.IDTokenConfig{
							UserAttributes: []string{"phone"},
						},
					},
				},
			},
		},
	}

	result := extractRequestedAttributes(app)

	s.Len(result, 1)
	s.True(result["phone"])
}

func (s *ApplicationServiceConsentTestSuite) TestExtractRequestedAttributes_FromUserInfo() {
	app := &model.ApplicationProcessedDTO{
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
					UserInfo: &model.UserInfoConfig{
						UserAttributes: []string{"address"},
					},
				},
			},
		},
	}

	result := extractRequestedAttributes(app)

	s.Len(result, 1)
	s.True(result["address"])
}

func (s *ApplicationServiceConsentTestSuite) TestExtractRequestedAttributes_Deduplicated() {
	app := &model.ApplicationProcessedDTO{
		Assertion: &model.AssertionConfig{
			UserAttributes: []string{"email"},
		},
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
					Token: &model.OAuthTokenConfig{
						AccessToken: &model.AccessTokenConfig{
							UserAttributes: []string{"email", "phone"},
						},
					},
				},
			},
		},
	}

	result := extractRequestedAttributes(app)

	// "email" appears in both assertion and access token, should be unique
	s.Len(result, 2)
	s.True(result["email"])
	s.True(result["phone"])
}

// ----- attributesToPurposeElements -----

func (s *ApplicationServiceConsentTestSuite) TestAttributesToPurposeElements() {
	attrs := map[string]bool{
		"email": true,
		"phone": true,
	}

	elements := attributesToPurposeElements(attrs)

	s.Len(elements, 2)
	for _, el := range elements {
		s.True(attrs[el.Name])
		s.False(el.IsMandatory)
		s.Equal(consent.NamespaceAttribute, el.Namespace)
	}
}

func (s *ApplicationServiceConsentTestSuite) TestAttributesToPurposeElements_EmptyMap() {
	elements := attributesToPurposeElements(map[string]bool{})

	s.Empty(elements)
}

// ----- wrapConsentServiceError -----

func (s *ApplicationServiceConsentTestSuite) TestWrapConsentServiceError_NilReturnsNil() {
	result := wrapConsentServiceError(nil)

	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestWrapConsentServiceError_ClientError_ReturnsConsentSyncFailed() {
	clientErr := &serviceerror.I18nServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "CSE-1007",
	}

	result := wrapConsentServiceError(clientErr)

	s.NotNil(result)
	s.Equal(serviceerror.ClientErrorType, result.Type)
	s.Equal(ErrorConsentSyncFailed.Code, result.Code)
}

func (s *ApplicationServiceConsentTestSuite) TestWrapConsentServiceError_ServerError_ReturnsInternalServerError() {
	serverErr := &serviceerror.I18nServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "CSE-500",
	}

	result := wrapConsentServiceError(serverErr)

	s.NotNil(result)
	s.Equal(serviceerror.ServerErrorType, result.Type)
}

// ----- createMissingConsentElements -----

func (s *ApplicationServiceConsentTestSuite) TestCreateMissingConsentElements_EmptyNames_ReturnsNil() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	result := svc.createMissingConsentElements(context.TODO(), "default", []string{})

	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestCreateMissingConsentElements_AllExist_NoCreate() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	names := []string{"email", "phone"}
	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", names).
		Return([]string{"email", "phone"}, nil)

	result := svc.createMissingConsentElements(context.TODO(), "default", names)

	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestCreateMissingConsentElements_SomeMissing_CreatesMissing() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	names := []string{"email", "phone"}
	// Only "email" exists; "phone" is missing
	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", names).
		Return([]string{"email"}, nil)

	expectedInput := []consent.ConsentElementInput{
		{Name: "phone", Namespace: consent.NamespaceAttribute},
	}
	cMock.EXPECT().CreateConsentElements(mock.Anything, "default", expectedInput).
		Return([]consent.ConsentElement{{ID: "e1", Name: "phone"}}, nil)

	result := svc.createMissingConsentElements(context.TODO(), "default", names)

	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestCreateMissingConsentElements_ValidateError_ReturnsError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	names := []string{"email"}
	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", names).
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	result := svc.createMissingConsentElements(context.TODO(), "default", names)

	s.NotNil(result)
	s.Equal(serviceerror.ServerErrorType, result.Type)
}

func (s *ApplicationServiceConsentTestSuite) TestCreateMissingConsentElements_CreateError_ReturnsError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	names := []string{"email"}
	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", names).
		Return([]string{}, nil)

	cMock.EXPECT().CreateConsentElements(mock.Anything, "default", mock.Anything).
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	result := svc.createMissingConsentElements(context.TODO(), "default", names)

	s.NotNil(result)
	s.Equal(serviceerror.ServerErrorType, result.Type)
}

// ----- syncConsentPurposeOnCreate -----

func (s *ApplicationServiceConsentTestSuite) TestSyncConsentPurposeOnCreate_NoAttributes_SkipsCreation() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	// No UserAttributes configured on the app
	appDTO := &model.ApplicationProcessedDTO{ID: "app-1", Name: "Test App"}

	result := svc.syncConsentPurposeOnCreate(appDTO)

	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestSyncConsentPurposeOnCreate_WithAttributes() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	appDTO := &model.ApplicationProcessedDTO{
		ID:   "app-1",
		Name: "Test App",
		Assertion: &model.AssertionConfig{
			UserAttributes: []string{"email"},
		},
	}

	// Validate returns empty → createMissingConsentElements creates "email"
	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", mock.Anything).
		Return([]string{}, nil)
	cMock.EXPECT().CreateConsentElements(mock.Anything, "default", mock.Anything).
		Return([]consent.ConsentElement{{ID: "e1", Name: "email"}}, nil)
	cMock.EXPECT().CreateConsentPurpose(mock.Anything, "default", mock.Anything).
		Return(&consent.ConsentPurpose{ID: "p1", Name: "Test App"}, nil)

	result := svc.syncConsentPurposeOnCreate(appDTO)

	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestSyncConsentPurposeOnCreate_CreateElementsError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	appDTO := &model.ApplicationProcessedDTO{
		ID:        "app-1",
		Name:      "Test App",
		Assertion: &model.AssertionConfig{UserAttributes: []string{"email"}},
	}

	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", mock.Anything).
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	result := svc.syncConsentPurposeOnCreate(appDTO)

	s.NotNil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestSyncConsentPurposeOnCreate_CreatePurposeError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	appDTO := &model.ApplicationProcessedDTO{
		ID:        "app-1",
		Name:      "Test App",
		Assertion: &model.AssertionConfig{UserAttributes: []string{"email"}},
	}

	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", mock.Anything).
		Return([]string{"email"}, nil)
	cMock.EXPECT().CreateConsentPurpose(mock.Anything, "default", mock.Anything).
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	result := svc.syncConsentPurposeOnCreate(appDTO)

	s.NotNil(result)
}

// ----- updateConsentPurpose -----

func (s *ApplicationServiceConsentTestSuite) TestUpdateConsentPurpose_NoPurposes_NoNewAttrs() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	existingApp := &model.ApplicationProcessedDTO{ID: "app-1", Name: "App"}
	updatedApp := &model.ApplicationProcessedDTO{ID: "app-1", Name: "App"}

	cMock.EXPECT().ListConsentPurposes(mock.Anything, "default", "app-1").
		Return([]consent.ConsentPurpose{}, nil)

	result := svc.updateConsentPurpose(existingApp, updatedApp)

	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestUpdateConsentPurpose_NoPurposes_NewAttrs() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	existingApp := &model.ApplicationProcessedDTO{ID: "app-1", Name: "App"}
	updatedApp := &model.ApplicationProcessedDTO{
		ID:        "app-1",
		Name:      "App",
		Assertion: &model.AssertionConfig{UserAttributes: []string{"email"}},
	}

	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", mock.Anything).
		Return([]string{"email"}, nil)
	cMock.EXPECT().ListConsentPurposes(mock.Anything, "default", "app-1").
		Return([]consent.ConsentPurpose{}, nil)
	cMock.EXPECT().CreateConsentPurpose(mock.Anything, "default", mock.Anything).
		Return(&consent.ConsentPurpose{ID: "p1"}, nil)

	result := svc.updateConsentPurpose(existingApp, updatedApp)

	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestUpdateConsentPurpose_ExistingPurpose_NoNewAttrs() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	existingApp := &model.ApplicationProcessedDTO{
		ID:        "app-1",
		Name:      "App",
		Assertion: &model.AssertionConfig{UserAttributes: []string{"email"}},
	}
	updatedApp := &model.ApplicationProcessedDTO{ID: "app-1", Name: "App"} // no attributes now

	cMock.EXPECT().ListConsentPurposes(mock.Anything, "default", "app-1").
		Return([]consent.ConsentPurpose{{ID: "p1", Name: "App"}}, nil)
	cMock.EXPECT().DeleteConsentPurpose(mock.Anything, "default", "p1").
		Return((*serviceerror.I18nServiceError)(nil))

	result := svc.updateConsentPurpose(existingApp, updatedApp)

	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestUpdateConsentPurpose_ExistingPurpose_NewAttrs() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	existingApp := &model.ApplicationProcessedDTO{
		ID:        "app-1",
		Name:      "App",
		Assertion: &model.AssertionConfig{UserAttributes: []string{"email"}},
	}
	updatedApp := &model.ApplicationProcessedDTO{
		ID:        "app-1",
		Name:      "App Updated",
		Assertion: &model.AssertionConfig{UserAttributes: []string{"email", "phone"}},
	}

	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", mock.Anything).
		Return([]string{"email", "phone"}, nil)
	cMock.EXPECT().ListConsentPurposes(mock.Anything, "default", "app-1").
		Return([]consent.ConsentPurpose{{ID: "p1", Name: "App"}}, nil)
	cMock.EXPECT().UpdateConsentPurpose(mock.Anything, "default", "p1", mock.Anything).
		Return(&consent.ConsentPurpose{ID: "p1", Name: "App Updated"}, nil)

	result := svc.updateConsentPurpose(existingApp, updatedApp)

	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestUpdateConsentPurpose_ListPurposesError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	existingApp := &model.ApplicationProcessedDTO{ID: "app-1"}
	updatedApp := &model.ApplicationProcessedDTO{ID: "app-1"}

	cMock.EXPECT().ListConsentPurposes(mock.Anything, "default", "app-1").
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	result := svc.updateConsentPurpose(existingApp, updatedApp)

	s.NotNil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestUpdateConsentPurpose_UpdatePurposeError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	existingApp := &model.ApplicationProcessedDTO{
		ID:   "app-1",
		Name: "App",
	}
	updatedApp := &model.ApplicationProcessedDTO{
		ID:        "app-1",
		Name:      "App",
		Assertion: &model.AssertionConfig{UserAttributes: []string{"email"}},
	}

	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", mock.Anything).
		Return([]string{"email"}, nil)
	cMock.EXPECT().ListConsentPurposes(mock.Anything, "default", "app-1").
		Return([]consent.ConsentPurpose{{ID: "p1"}}, nil)
	cMock.EXPECT().UpdateConsentPurpose(mock.Anything, "default", "p1", mock.Anything).
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	result := svc.updateConsentPurpose(existingApp, updatedApp)

	s.NotNil(result)
}

// ----- deleteConsentPurposes -----

func (s *ApplicationServiceConsentTestSuite) TestDeleteConsentPurposes_NoPurposesFound() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	cMock.EXPECT().ListConsentPurposes(mock.Anything, "default", "app-1").
		Return([]consent.ConsentPurpose{}, nil)

	result := svc.deleteConsentPurposes("app-1")

	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestDeleteConsentPurposes_Success() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	cMock.EXPECT().ListConsentPurposes(mock.Anything, "default", "app-1").
		Return([]consent.ConsentPurpose{{ID: "p1", Name: "Login"}}, nil)
	cMock.EXPECT().DeleteConsentPurpose(mock.Anything, "default", "p1").
		Return((*serviceerror.I18nServiceError)(nil))

	result := svc.deleteConsentPurposes("app-1")

	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestDeleteConsentPurposes_AssociatedRecordsError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	cMock.EXPECT().ListConsentPurposes(mock.Anything, "default", "app-1").
		Return([]consent.ConsentPurpose{{ID: "p1"}}, nil)
	cMock.EXPECT().DeleteConsentPurpose(mock.Anything, "default", "p1").
		Return(&consent.ErrorDeletingConsentPurposeWithAssociatedRecords)

	result := svc.deleteConsentPurposes("app-1")

	// Should return nil — associated records error is treated as a warning, not a fatal error
	s.Nil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestDeleteConsentPurposes_OtherDeleteError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	cMock.EXPECT().ListConsentPurposes(mock.Anything, "default", "app-1").
		Return([]consent.ConsentPurpose{{ID: "p1"}}, nil)
	cMock.EXPECT().DeleteConsentPurpose(mock.Anything, "default", "p1").
		Return(&serviceerror.InternalServerErrorWithI18n)

	result := svc.deleteConsentPurposes("app-1")

	s.NotNil(result)
}

func (s *ApplicationServiceConsentTestSuite) TestDeleteConsentPurposes_ListError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsent(cMock)

	cMock.EXPECT().ListConsentPurposes(mock.Anything, "default", "app-1").
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	result := svc.deleteConsentPurposes("app-1")

	s.NotNil(result)
}

// newTestApplicationServiceWithConsentStoreAndCert creates a minimal applicationService
// with consentService, appStore and certService set for testing syncConsentPurposeOnUpdate.
func newTestApplicationServiceWithConsentStoreAndCert(
	consentSvc consent.ConsentServiceInterface,
	store applicationStoreInterface,
	certSvc cert.CertificateServiceInterface,
) *applicationService {
	return &applicationService{
		consentService: consentSvc,
		appStore:       store,
		certService:    certSvc,
	}
}

// ----- syncConsentPurposeOnUpdate compensation paths -----

// TestSyncConsentPurposeOnUpdate_ConsentFails_RevertStoreFails verifies that when the consent
// sync fails and the compensating app-revert also fails, the original consent error is returned.
func (s *ApplicationServiceConsentTestSuite) TestSyncConsentPurposeOnUpdate_ConsentFails_RevertStoreFails() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	storeMock := newApplicationStoreInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsentStoreAndCert(cMock, storeMock, nil)

	existingApp := &model.ApplicationProcessedDTO{
		ID:           "app-1",
		Name:         "App",
		LoginConsent: &model.LoginConsentConfig{Enabled: false},
	}
	updatedApp := &model.ApplicationProcessedDTO{
		ID:           "app-1",
		Name:         "App",
		LoginConsent: &model.LoginConsentConfig{Enabled: false},
	}

	// deleteConsentPurposes fails → triggers compensation
	cMock.EXPECT().ListConsentPurposes(mock.Anything, "default", "app-1").
		Return(nil, &serviceerror.InternalServerErrorWithI18n)
	// Compensation: app revert also fails (logged, not propagated)
	storeMock.On("UpdateApplication", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("revert store error"))

	svcErr := svc.syncConsentPurposeOnUpdate(context.Background(), "app-1", existingApp, updatedApp, nil, nil)

	// Returns the original consent error even though the revert also failed
	s.NotNil(svcErr)
}

// TestSyncConsentPurposeOnUpdate_ConsentFails_WithCert_CertRollbackFails verifies that when
// the consent sync fails with an updatedCert present and cert rollback fails, the original
// consent error is still returned (rollback failure is only logged).
func (s *ApplicationServiceConsentTestSuite) TestSyncConsentPurposeOnUpdate_ConsentFails_WithCert_CertRollbackFail() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	storeMock := newApplicationStoreInterfaceMock(s.T())
	certMock := certmock.NewCertificateServiceInterfaceMock(s.T())
	svc := newTestApplicationServiceWithConsentStoreAndCert(cMock, storeMock, certMock)

	existingApp := &model.ApplicationProcessedDTO{
		ID:           "app-1",
		Name:         "App",
		LoginConsent: &model.LoginConsentConfig{Enabled: false},
	}
	updatedApp := &model.ApplicationProcessedDTO{
		ID:           "app-1",
		Name:         "App",
		LoginConsent: &model.LoginConsentConfig{Enabled: false},
	}
	// updatedCert != nil (existingCert == nil) → rollback calls DeleteCertificateByReference
	updatedCert := &cert.Certificate{
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}

	// deleteConsentPurposes fails → triggers compensation
	cMock.EXPECT().ListConsentPurposes(mock.Anything, "default", "app-1").
		Return(nil, &serviceerror.InternalServerErrorWithI18n)
	// Compensation: revert app succeeds
	storeMock.On("UpdateApplication", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// Cert rollback fails (existingCert=nil, updatedCert!=nil → DeleteCertificateByReference)
	certMock.EXPECT().
		DeleteCertificateByReference(mock.Anything, cert.CertificateReferenceTypeApplication, "app-1").
		Return(&serviceerror.ServiceError{Type: serviceerror.ServerErrorType})

	svcErr := svc.syncConsentPurposeOnUpdate(context.Background(), "app-1", existingApp, updatedApp, nil, updatedCert)

	// Returns the original consent error even though cert rollback also failed
	s.NotNil(svcErr)
}
