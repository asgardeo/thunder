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

package inboundclient

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/entityprovider"
	entitytypepkg "github.com/asgardeo/thunder/internal/entitytype"
	flowcommon "github.com/asgardeo/thunder/internal/flow/common"
	inboundmodel "github.com/asgardeo/thunder/internal/inboundclient/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	sysconfig "github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/i18n/core"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/transaction"
	"github.com/asgardeo/thunder/pkg/entityprovider"
	flowcommon "github.com/asgardeo/thunder/pkg/flow/common"
	inboundmodel "github.com/asgardeo/thunder/pkg/inboundclient/model"
	oauth2const "github.com/asgardeo/thunder/pkg/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/tests/mocks/certmock"
	"github.com/asgardeo/thunder/tests/mocks/design/layoutmock"
	"github.com/asgardeo/thunder/tests/mocks/design/thememock"
	"github.com/asgardeo/thunder/tests/mocks/entityprovidermock"
	"github.com/asgardeo/thunder/tests/mocks/entitytypemock"
	"github.com/asgardeo/thunder/tests/mocks/flow/flowmgtmock"
)

type InboundClientServiceTestSuite struct {
	suite.Suite
}

func TestInboundClientServiceTestSuite(t *testing.T) {
	suite.Run(t, new(InboundClientServiceTestSuite))
}

func (suite *InboundClientServiceTestSuite) SetupTest() {
	sysconfig.ResetServerRuntime()
	suite.Require().NoError(sysconfig.InitializeServerRuntime("/tmp/test", &sysconfig.Config{}))
}

func newServiceForTest(store inboundClientStoreInterface) InboundClientServiceInterface {
	return newInboundClientService(store, transaction.NewNoOpTransactioner(), nil, nil, nil, nil, nil, nil, nil)
}

func newServiceWithCert(certService cert.CertificateServiceInterface) *inboundClientService {
	svc := newInboundClientService(
		nil, transaction.NewNoOpTransactioner(), certService, nil, nil, nil, nil, nil, nil,
	)
	return svc.(*inboundClientService)
}

func validInboundClient(id string) inboundmodel.InboundClient {
	return inboundmodel.InboundClient{
		ID:                        id,
		AuthFlowID:                "flow-1",
		RegistrationFlowID:        "reg-1",
		IsRegistrationFlowEnabled: true,
	}
}

func ptrInboundClient() *inboundmodel.InboundClient {
	c := validInboundClient("p1")
	return &c
}

func validOAuthProfileData() *inboundmodel.OAuthProfileData {
	return &inboundmodel.OAuthProfileData{
		RedirectURIs:            []string{"https://app.example.com/cb"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "client_secret_basic",
	}
}

// ----- Inbound client CRUD -----

func (suite *InboundClientServiceTestSuite) TestCreateInboundClient_RunsValidationBeforePersist() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().IsDeclarative(mock.Anything, "p1").Return(false)
	svc := newServiceForTest(store)

	p := validOAuthProfileData()
	p.GrantTypes = []string{"not_a_real_grant"}

	err := svc.CreateInboundClient(context.Background(), ptrInboundClient(), nil, p, false, "")

	assert.ErrorIs(suite.T(), err, ErrOAuthInvalidGrantType)
}

func (suite *InboundClientServiceTestSuite) TestCreateInboundClient_PersistsBoth() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().IsDeclarative(mock.Anything, "p1").Return(false)
	store.EXPECT().CreateInboundClient(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().CreateOAuthProfile(mock.Anything, "p1", mock.Anything).Return(nil)

	svc := newServiceForTest(store)
	err := svc.CreateInboundClient(context.Background(), ptrInboundClient(),
		nil, validOAuthProfileData(), true, "")

	assert.NoError(suite.T(), err)
}

func (suite *InboundClientServiceTestSuite) TestCreateInboundClient_PersistsClientOnlyWhenOAuthNil() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().IsDeclarative(mock.Anything, "p1").Return(false)
	store.EXPECT().CreateInboundClient(mock.Anything, mock.Anything).Return(nil)

	svc := newServiceForTest(store)
	err := svc.CreateInboundClient(context.Background(), ptrInboundClient(), nil, nil, false, "")

	assert.NoError(suite.T(), err)
}

func (suite *InboundClientServiceTestSuite) TestCreateInboundClient_RefusesDeclarative() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().IsDeclarative(mock.Anything, "p1").Return(true)

	svc := newServiceForTest(store)
	err := svc.CreateInboundClient(context.Background(), ptrInboundClient(), nil, nil, false, "")

	assert.ErrorIs(suite.T(), err, ErrCannotModifyDeclarative)
}

func (suite *InboundClientServiceTestSuite) TestUpdateInboundClient_RefusesDeclarative() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().IsDeclarative(mock.Anything, "p1").Return(true)

	svc := newServiceForTest(store)
	err := svc.UpdateInboundClient(context.Background(), ptrInboundClient(), nil, nil, false, "", "")

	assert.ErrorIs(suite.T(), err, ErrCannotModifyDeclarative)
}

func (suite *InboundClientServiceTestSuite) TestDeleteInboundClient_RefusesDeclarative() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().IsDeclarative(mock.Anything, "p1").Return(true)

	svc := newServiceForTest(store)
	err := svc.DeleteInboundClient(context.Background(), "p1")

	assert.ErrorIs(suite.T(), err, ErrCannotModifyDeclarative)
}

func (suite *InboundClientServiceTestSuite) TestDelegatesPlainReads() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().GetInboundClientList(mock.Anything, mock.Anything).
		Return([]inboundmodel.InboundClient{validInboundClient("p1")}, nil)
	store.EXPECT().IsDeclarative(mock.Anything, "p1").Return(true)

	svc := newServiceForTest(store)
	list, err := svc.GetInboundClientList(context.Background())
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), list, 1)

	assert.True(suite.T(), svc.IsDeclarative(context.Background(), "p1"))
}

func (suite *InboundClientServiceTestSuite) TestDeleteInboundClient() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().IsDeclarative(mock.Anything, "p1").Return(false)
	store.EXPECT().DeleteInboundClient(mock.Anything, "p1").Return(nil)

	svc := newServiceForTest(store)
	assert.NoError(suite.T(), svc.DeleteInboundClient(context.Background(), "p1"))
}

func (suite *InboundClientServiceTestSuite) TestStorePropagatesErrors() {
	storeErr := errors.New("db error")
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().IsDeclarative(mock.Anything, "p1").Return(false)
	store.EXPECT().CreateInboundClient(mock.Anything, mock.Anything).Return(storeErr)

	svc := newServiceForTest(store)
	err := svc.CreateInboundClient(context.Background(), ptrInboundClient(), nil, nil, false, "")

	assert.ErrorIs(suite.T(), err, storeErr)
}

// ----- ValidateCertificateInput -----

func (suite *InboundClientServiceTestSuite) TestValidateCertificateInput_Empty() {
	c, err := validateCertificateInput(cert.CertificateReferenceTypeOAuthApp, "ref-1", "", nil)

	suite.Nil(c)
	suite.Nil(err)
}

func (suite *InboundClientServiceTestSuite) TestValidateCertificateInput_JWKS_Success() {
	c, err := validateCertificateInput(cert.CertificateReferenceTypeOAuthApp, "ref-1", "existing",
		&inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: `{"keys":[]}`})

	suite.Nil(err)
	suite.NotNil(c)
	suite.Equal("existing", c.ID)
	suite.Equal(cert.CertificateTypeJWKS, c.Type)
	suite.Equal(cert.CertificateReferenceTypeOAuthApp, c.RefType)
	suite.Equal("ref-1", c.RefID)
}

func (suite *InboundClientServiceTestSuite) TestValidateCertificateInput_JWKS_MissingValue() {
	c, err := validateCertificateInput(cert.CertificateReferenceTypeOAuthApp, "ref-1", "",
		&inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: ""})

	suite.Nil(c)
	suite.ErrorIs(err, ErrCertValueRequired)
}

func (suite *InboundClientServiceTestSuite) TestValidateCertificateInput_JWKSURI_Success() {
	c, err := validateCertificateInput(cert.CertificateReferenceTypeOAuthApp, "ref-1", "",
		&inboundmodel.Certificate{Type: cert.CertificateTypeJWKSURI, Value: "https://example.com/jwks"})

	suite.Nil(err)
	suite.Equal(cert.CertificateTypeJWKSURI, c.Type)
}

func (suite *InboundClientServiceTestSuite) TestValidateCertificateInput_JWKSURI_Invalid() {
	c, err := validateCertificateInput(cert.CertificateReferenceTypeOAuthApp, "ref-1", "",
		&inboundmodel.Certificate{Type: cert.CertificateTypeJWKSURI, Value: "not-a-uri"})

	suite.Nil(c)
	suite.ErrorIs(err, ErrCertInvalidJWKSURI)
}

func (suite *InboundClientServiceTestSuite) TestValidateCertificateInput_InvalidType() {
	c, err := validateCertificateInput(cert.CertificateReferenceTypeOAuthApp, "ref-1", "",
		&inboundmodel.Certificate{Type: "bogus", Value: "x"})

	suite.Nil(c)
	suite.ErrorIs(err, ErrCertInvalidType)
}

// ----- CreateCertificate -----

func (suite *InboundClientServiceTestSuite) TestCreateCertificate_Nil() {
	svc := newServiceWithCert(certmock.NewCertificateServiceInterfaceMock(suite.T()))

	out, vErr, opErr := svc.createCertificate(context.Background(),
		cert.CertificateReferenceTypeOAuthApp, "ref-1", nil)

	suite.Nil(out)
	suite.Nil(vErr)
	suite.Nil(opErr)
}

func (suite *InboundClientServiceTestSuite) TestCreateCertificate_Success() {
	mockCert := certmock.NewCertificateServiceInterfaceMock(suite.T())
	mockCert.EXPECT().CreateCertificate(mock.Anything, mock.Anything).
		Return(&cert.Certificate{}, nil)
	svc := newServiceWithCert(mockCert)

	in := &inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: `{}`}
	out, vErr, opErr := svc.createCertificate(context.Background(),
		cert.CertificateReferenceTypeOAuthApp, "ref-1", in)

	suite.Nil(vErr)
	suite.Nil(opErr)
	suite.Equal(cert.CertificateTypeJWKS, out.Type)
	suite.Equal(`{}`, out.Value)
}

func (suite *InboundClientServiceTestSuite) TestCreateCertificate_InvalidInput() {
	svc := newServiceWithCert(certmock.NewCertificateServiceInterfaceMock(suite.T()))

	in := &inboundmodel.Certificate{Type: cert.CertificateTypeJWKSURI, Value: "not-a-uri"}
	out, vErr, opErr := svc.createCertificate(context.Background(),
		cert.CertificateReferenceTypeOAuthApp, "ref-1", in)

	suite.Nil(out)
	suite.Nil(opErr)
	suite.ErrorIs(vErr, ErrCertInvalidJWKSURI)
}

func (suite *InboundClientServiceTestSuite) TestCreateCertificate_ServiceError() {
	mockCert := certmock.NewCertificateServiceInterfaceMock(suite.T())
	clientErr := &serviceerror.ServiceError{Type: serviceerror.ClientErrorType, Code: "C-1"}
	mockCert.EXPECT().CreateCertificate(mock.Anything, mock.Anything).Return(nil, clientErr)
	svc := newServiceWithCert(mockCert)

	in := &inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: `{}`}
	out, vErr, opErr := svc.createCertificate(context.Background(),
		cert.CertificateReferenceTypeOAuthApp, "ref-1", in)

	suite.Nil(out)
	suite.Nil(vErr)
	suite.Equal(CertOpCreate, opErr.Operation)
	suite.Same(clientErr, opErr.Underlying)
	suite.True(opErr.IsClientError())
}

// ----- GetCertificate -----

func (suite *InboundClientServiceTestSuite) TestGetCertificate_NotFound() {
	mockCert := certmock.NewCertificateServiceInterfaceMock(suite.T())
	mockCert.EXPECT().
		GetCertificateByReference(mock.Anything, cert.CertificateReferenceTypeApplication, "ref-1").
		Return(nil, &cert.ErrorCertificateNotFound)
	svc := newServiceWithCert(mockCert)

	out, err := svc.GetCertificate(context.Background(), cert.CertificateReferenceTypeApplication, "ref-1")

	suite.Nil(out)
	suite.Nil(err)
}

func (suite *InboundClientServiceTestSuite) TestGetCertificate_Success() {
	mockCert := certmock.NewCertificateServiceInterfaceMock(suite.T())
	mockCert.EXPECT().
		GetCertificateByReference(mock.Anything, cert.CertificateReferenceTypeApplication, "ref-1").
		Return(&cert.Certificate{Type: cert.CertificateTypeJWKS, Value: `{}`}, nil)
	svc := newServiceWithCert(mockCert)

	out, err := svc.GetCertificate(context.Background(), cert.CertificateReferenceTypeApplication, "ref-1")

	suite.Nil(err)
	suite.Equal(cert.CertificateTypeJWKS, out.Type)
}

func (suite *InboundClientServiceTestSuite) TestGetCertificate_ServerError() {
	mockCert := certmock.NewCertificateServiceInterfaceMock(suite.T())
	srvErr := &serviceerror.ServiceError{Type: serviceerror.ServerErrorType, Code: "S-1"}
	mockCert.EXPECT().
		GetCertificateByReference(mock.Anything, cert.CertificateReferenceTypeApplication, "ref-1").
		Return(nil, srvErr)
	svc := newServiceWithCert(mockCert)

	out, err := svc.GetCertificate(context.Background(), cert.CertificateReferenceTypeApplication, "ref-1")

	suite.Nil(out)
	suite.Equal(CertOpRetrieve, err.Operation)
	suite.False(err.IsClientError())
}

// ----- DeleteCertificate -----

func (suite *InboundClientServiceTestSuite) TestDeleteCertificate_Success() {
	mockCert := certmock.NewCertificateServiceInterfaceMock(suite.T())
	mockCert.EXPECT().
		DeleteCertificateByReference(mock.Anything, cert.CertificateReferenceTypeApplication, "ref-1").
		Return(nil)
	svc := newServiceWithCert(mockCert)

	err := svc.deleteCertificate(context.Background(), cert.CertificateReferenceTypeApplication, "ref-1")

	suite.Nil(err)
}

func (suite *InboundClientServiceTestSuite) TestDeleteCertificate_Error() {
	mockCert := certmock.NewCertificateServiceInterfaceMock(suite.T())
	clientErr := &serviceerror.ServiceError{Type: serviceerror.ClientErrorType, Code: "D-1"}
	mockCert.EXPECT().
		DeleteCertificateByReference(mock.Anything, cert.CertificateReferenceTypeOAuthApp, "ref-1").
		Return(clientErr)
	svc := newServiceWithCert(mockCert)

	err := svc.deleteCertificate(context.Background(), cert.CertificateReferenceTypeOAuthApp, "ref-1")

	suite.NotNil(err)
	suite.Equal(CertOpDelete, err.Operation)
	suite.Equal(cert.CertificateReferenceTypeOAuthApp, err.RefType)
}

// ----- SyncCertificate -----

func (suite *InboundClientServiceTestSuite) TestSyncCertificate_NoOp_NoExistingNoInput() {
	mockCert := certmock.NewCertificateServiceInterfaceMock(suite.T())
	mockCert.EXPECT().
		GetCertificateByReference(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, &cert.ErrorCertificateNotFound)
	svc := newServiceWithCert(mockCert)

	out, vErr, opErr := svc.syncCertificate(context.Background(),
		cert.CertificateReferenceTypeApplication, "ref-1", nil)

	suite.Nil(out)
	suite.Nil(vErr)
	suite.Nil(opErr)
}

func (suite *InboundClientServiceTestSuite) TestSyncCertificate_CreateWhenAbsent() {
	mockCert := certmock.NewCertificateServiceInterfaceMock(suite.T())
	mockCert.EXPECT().
		GetCertificateByReference(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, &cert.ErrorCertificateNotFound)
	mockCert.EXPECT().CreateCertificate(mock.Anything, mock.Anything).
		Return(&cert.Certificate{}, nil)
	svc := newServiceWithCert(mockCert)

	out, vErr, opErr := svc.syncCertificate(context.Background(),
		cert.CertificateReferenceTypeApplication, "ref-1",
		&inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: `{}`})

	suite.Nil(vErr)
	suite.Nil(opErr)
	suite.NotNil(out)
}

func (suite *InboundClientServiceTestSuite) TestSyncCertificate_UpdateWhenPresent() {
	mockCert := certmock.NewCertificateServiceInterfaceMock(suite.T())
	mockCert.EXPECT().
		GetCertificateByReference(mock.Anything, mock.Anything, mock.Anything).
		Return(&cert.Certificate{ID: "cert-1"}, nil)
	mockCert.EXPECT().UpdateCertificateByID(mock.Anything, "cert-1", mock.Anything).
		Return(&cert.Certificate{}, nil)
	svc := newServiceWithCert(mockCert)

	out, vErr, opErr := svc.syncCertificate(context.Background(),
		cert.CertificateReferenceTypeApplication, "ref-1",
		&inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: `{}`})

	suite.Nil(vErr)
	suite.Nil(opErr)
	suite.NotNil(out)
}

func (suite *InboundClientServiceTestSuite) TestSyncCertificate_DeleteWhenInputEmpty() {
	mockCert := certmock.NewCertificateServiceInterfaceMock(suite.T())
	mockCert.EXPECT().
		GetCertificateByReference(mock.Anything, mock.Anything, mock.Anything).
		Return(&cert.Certificate{ID: "cert-1"}, nil)
	mockCert.EXPECT().
		DeleteCertificateByReference(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	svc := newServiceWithCert(mockCert)

	out, vErr, opErr := svc.syncCertificate(context.Background(),
		cert.CertificateReferenceTypeApplication, "ref-1", nil)

	suite.Nil(out)
	suite.Nil(vErr)
	suite.Nil(opErr)
}

func (suite *InboundClientServiceTestSuite) TestSyncCertificate_ValidationError() {
	mockCert := certmock.NewCertificateServiceInterfaceMock(suite.T())
	mockCert.EXPECT().
		GetCertificateByReference(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, &cert.ErrorCertificateNotFound)
	svc := newServiceWithCert(mockCert)

	out, vErr, opErr := svc.syncCertificate(context.Background(),
		cert.CertificateReferenceTypeApplication, "ref-1",
		&inboundmodel.Certificate{Type: "bogus", Value: "x"})

	suite.Nil(out)
	suite.Nil(opErr)
	suite.ErrorIs(vErr, ErrCertInvalidType)
}

func (suite *InboundClientServiceTestSuite) TestGetInboundClientByEntityID_Delegates() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	want := &inboundmodel.InboundClient{ID: "p1"}
	store.EXPECT().GetInboundClientByEntityID(mock.Anything, "p1").Return(want, nil)

	svc := newServiceForTest(store)
	got, err := svc.GetInboundClientByEntityID(context.Background(), "p1")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "p1", got.ID)
}

func (suite *InboundClientServiceTestSuite) TestGetOAuthProfileByEntityID_Delegates() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	want := &inboundmodel.OAuthProfile{AppID: "p1"}
	store.EXPECT().GetOAuthProfileByEntityID(mock.Anything, "p1").Return(want, nil)

	svc := newServiceForTest(store)
	got, err := svc.GetOAuthProfileByEntityID(context.Background(), "p1")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "p1", got.AppID)
}

func (suite *InboundClientServiceTestSuite) TestUpdateInboundClient_ValidationFails() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().IsDeclarative(mock.Anything, "p1").Return(false)
	svc := newServiceForTest(store)

	p := validOAuthProfileData()
	p.GrantTypes = []string{"not_a_real_grant"}

	err := svc.UpdateInboundClient(context.Background(), ptrInboundClient(), nil, p, false, "", "")
	assert.ErrorIs(suite.T(), err, ErrOAuthInvalidGrantType)
}

func (suite *InboundClientServiceTestSuite) TestUpdateInboundClient_Succeeds() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().IsDeclarative(mock.Anything, "p1").Return(false)
	store.EXPECT().UpdateInboundClient(mock.Anything, mock.Anything).Return(nil)
	// syncOAuthProfile path: GetOAuthProfileByEntityID returns not found → CreateOAuthProfile
	store.EXPECT().GetOAuthProfileByEntityID(mock.Anything, "p1").Return(nil, ErrInboundClientNotFound)
	store.EXPECT().CreateOAuthProfile(mock.Anything, "p1", mock.Anything).Return(nil)

	mockCert := certmock.NewCertificateServiceInterfaceMock(suite.T())
	// syncCertificate for app cert (nil input): gets existing (not found), no update needed
	mockCert.EXPECT().GetCertificateByReference(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, &cert.ErrorCertificateNotFound)

	svc := newInboundClientService(store, transaction.NewNoOpTransactioner(), mockCert, nil, nil, nil, nil, nil, nil)
	err := svc.UpdateInboundClient(context.Background(), ptrInboundClient(), nil, validOAuthProfileData(), true, "", "")
	assert.NoError(suite.T(), err)
}

func (suite *InboundClientServiceTestSuite) TestValidate_ValidProfile() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	svc := newServiceForTest(store)

	err := svc.Validate(context.Background(), ptrInboundClient(), validOAuthProfileData(), true)
	assert.NoError(suite.T(), err)
}

func (suite *InboundClientServiceTestSuite) TestValidate_InvalidGrantType() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	svc := newServiceForTest(store)

	p := validOAuthProfileData()
	p.GrantTypes = []string{"bogus_grant"}

	err := svc.Validate(context.Background(), ptrInboundClient(), p, false)
	assert.ErrorIs(suite.T(), err, ErrOAuthInvalidGrantType)
}

func (suite *InboundClientServiceTestSuite) TestValidateRedirectURIs_WildcardInHost_Rejected() {
	p := &inboundmodel.OAuthProfileData{
		RedirectURIs: []string{"https://*.example.com/cb"},
		GrantTypes:   []string{"authorization_code"},
	}
	err := validateRedirectURIs(p)
	assert.ErrorIs(suite.T(), err, ErrOAuthInvalidRedirectURI)
}

func (suite *InboundClientServiceTestSuite) TestValidateRedirectURIs_WildcardInQuery_Rejected() {
	p := &inboundmodel.OAuthProfileData{
		RedirectURIs: []string{"https://app.example.com/cb?foo=*"},
		GrantTypes:   []string{"authorization_code"},
	}
	err := validateRedirectURIs(p)
	assert.ErrorIs(suite.T(), err, ErrOAuthInvalidRedirectURI)
}

func (suite *InboundClientServiceTestSuite) TestValidatePublicClient_PKCENotRequired_Fails() {
	p := &inboundmodel.OAuthProfileData{
		PublicClient:            true,
		PKCERequired:            false,
		TokenEndpointAuthMethod: "none",
	}
	err := validatePublicClient(p)
	assert.ErrorIs(suite.T(), err, ErrOAuthPublicClientMustHavePKCE)
}

func (suite *InboundClientServiceTestSuite) TestValidateTokenEndpointAuthMethod_InvalidMethod() {
	p := &inboundmodel.OAuthProfileData{
		TokenEndpointAuthMethod: "bogus_method",
	}
	err := validateTokenEndpointAuthMethod(p, false)
	assert.ErrorIs(suite.T(), err, ErrOAuthInvalidTokenEndpointAuthMethod)
}

func (suite *InboundClientServiceTestSuite) TestValidateTokenEndpoint_CertAllowedWhenUserInfoNeedsIt() {
	p := &inboundmodel.OAuthProfileData{
		TokenEndpointAuthMethod: "client_secret_basic",
		Certificate:             &inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: "{}"},
		UserInfo:                &inboundmodel.UserInfoConfig{EncryptionAlg: "RSA-OAEP-256"},
	}
	assert.NoError(suite.T(), validateTokenEndpointAuthMethod(p, true))
}

func (suite *InboundClientServiceTestSuite) TestValidateTokenEndpoint_CertRejectedWhenUserInfoDoesNotNeedIt() {
	p := &inboundmodel.OAuthProfileData{
		TokenEndpointAuthMethod: "client_secret_basic",
		Certificate:             &inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: "{}"},
	}
	err := validateTokenEndpointAuthMethod(p, true)
	assert.ErrorIs(suite.T(), err, ErrOAuthClientSecretCannotHaveCertificate)
}

func (suite *InboundClientServiceTestSuite) TestValidateTokenEndpointAuthMethod_PrivateKeyJWTHappy() {
	p := &inboundmodel.OAuthProfileData{
		TokenEndpointAuthMethod: "private_key_jwt",
		Certificate:             &inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: "{}"},
	}
	assert.NoError(suite.T(), validateTokenEndpointAuthMethod(p, false))
}

func (suite *InboundClientServiceTestSuite) TestValidateTokenEndpointAuthMethod_PrivateKeyJWTMissingCert() {
	p := &inboundmodel.OAuthProfileData{TokenEndpointAuthMethod: "private_key_jwt"}
	err := validateTokenEndpointAuthMethod(p, false)
	assert.ErrorIs(suite.T(), err, ErrOAuthPrivateKeyJWTRequiresCertificate)
}

func (suite *InboundClientServiceTestSuite) TestValidateTokenEndpointAuthMethod_PrivateKeyJWTWithSecret() {
	p := &inboundmodel.OAuthProfileData{
		TokenEndpointAuthMethod: "private_key_jwt",
		Certificate:             &inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: "{}"},
	}
	err := validateTokenEndpointAuthMethod(p, true)
	assert.ErrorIs(suite.T(), err, ErrOAuthPrivateKeyJWTCannotHaveClientSecret)
}

func (suite *InboundClientServiceTestSuite) TestValidateTokenEndpointAuthMethod_NoneRequiresPublicClient() {
	p := &inboundmodel.OAuthProfileData{TokenEndpointAuthMethod: "none"}
	err := validateTokenEndpointAuthMethod(p, false)
	assert.ErrorIs(suite.T(), err, ErrOAuthNoneAuthRequiresPublicClient)
}

func (suite *InboundClientServiceTestSuite) TestValidateTokenEndpointAuthMethod_NoneRejectsCertOrSecret() {
	p := &inboundmodel.OAuthProfileData{
		TokenEndpointAuthMethod: "none",
		PublicClient:            true,
		Certificate:             &inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: "{}"},
	}
	err := validateTokenEndpointAuthMethod(p, false)
	assert.ErrorIs(suite.T(), err, ErrOAuthNoneAuthCannotHaveCertOrSecret)
}

func (suite *InboundClientServiceTestSuite) TestValidateTokenEndpointAuthMethod_NoneClientCredentialsRejected() {
	p := &inboundmodel.OAuthProfileData{
		TokenEndpointAuthMethod: "none",
		PublicClient:            true,
		GrantTypes:              []string{"client_credentials"},
	}
	err := validateTokenEndpointAuthMethod(p, false)
	assert.ErrorIs(suite.T(), err, ErrOAuthClientCredentialsCannotUseNoneAuth)
}

// validateUserInfoConfig — happy paths

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_NilUserInfo() {
	assert.NoError(suite.T(), validateUserInfoConfig(&inboundmodel.OAuthProfileData{}))
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_PlainJSON() {
	p := &inboundmodel.OAuthProfileData{
		UserInfo: &inboundmodel.UserInfoConfig{ResponseType: inboundmodel.UserInfoResponseTypeJSON},
	}
	assert.NoError(suite.T(), validateUserInfoConfig(p))
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_JWSHappy() {
	p := &inboundmodel.OAuthProfileData{
		UserInfo: &inboundmodel.UserInfoConfig{
			ResponseType: inboundmodel.UserInfoResponseTypeJWS,
			SigningAlg:   "RS256",
		},
	}
	assert.NoError(suite.T(), validateUserInfoConfig(p))
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_JWEHappy() {
	p := &inboundmodel.OAuthProfileData{
		Certificate: &inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: "{}"},
		UserInfo: &inboundmodel.UserInfoConfig{
			ResponseType:  inboundmodel.UserInfoResponseTypeJWE,
			EncryptionAlg: "RSA-OAEP-256",
			EncryptionEnc: "A256GCM",
		},
	}
	assert.NoError(suite.T(), validateUserInfoConfig(p))
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_NestedJWTHappy() {
	p := &inboundmodel.OAuthProfileData{
		Certificate: &inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: "{}"},
		UserInfo: &inboundmodel.UserInfoConfig{
			ResponseType:  inboundmodel.UserInfoResponseTypeNESTEDJWT,
			SigningAlg:    "RS256",
			EncryptionAlg: "RSA-OAEP-256",
			EncryptionEnc: "A256GCM",
		},
	}
	assert.NoError(suite.T(), validateUserInfoConfig(p))
}

// validateUserInfoConfig — error paths

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_UnsupportedSigningAlg() {
	p := &inboundmodel.OAuthProfileData{
		UserInfo: &inboundmodel.UserInfoConfig{SigningAlg: "BOGUS"},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoUnsupportedSigningAlg)
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_EncryptionEncWithoutAlg() {
	p := &inboundmodel.OAuthProfileData{
		UserInfo: &inboundmodel.UserInfoConfig{EncryptionEnc: "A256GCM"},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoEncryptionEncRequiresAlg)
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_UnsupportedEncryptionAlg() {
	p := &inboundmodel.OAuthProfileData{
		UserInfo: &inboundmodel.UserInfoConfig{EncryptionAlg: "BOGUS", EncryptionEnc: "A256GCM"},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoUnsupportedEncryptionAlg)
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_EncryptionAlgWithoutEnc() {
	p := &inboundmodel.OAuthProfileData{
		UserInfo: &inboundmodel.UserInfoConfig{EncryptionAlg: "RSA-OAEP-256"},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoEncryptionAlgRequiresEnc)
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_UnsupportedEncryptionEnc() {
	p := &inboundmodel.OAuthProfileData{
		UserInfo: &inboundmodel.UserInfoConfig{EncryptionAlg: "RSA-OAEP-256", EncryptionEnc: "BOGUS"},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoUnsupportedEncryptionEnc)
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_EncryptionRequiresCertificate() {
	p := &inboundmodel.OAuthProfileData{
		UserInfo: &inboundmodel.UserInfoConfig{EncryptionAlg: "RSA-OAEP-256", EncryptionEnc: "A256GCM"},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoEncryptionRequiresCertificate)
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_JWKSURISSRFRejection() {
	p := &inboundmodel.OAuthProfileData{
		Certificate: &inboundmodel.Certificate{Type: cert.CertificateTypeJWKSURI, Value: "http://127.0.0.1/jwks"},
		UserInfo: &inboundmodel.UserInfoConfig{
			EncryptionAlg: "RSA-OAEP-256", EncryptionEnc: "A256GCM",
		},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoJWKSURINotSSRFSafe)
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_JWSMissingSigningAlg() {
	p := &inboundmodel.OAuthProfileData{
		UserInfo: &inboundmodel.UserInfoConfig{ResponseType: inboundmodel.UserInfoResponseTypeJWS},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoJWSRequiresSigningAlg)
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_JWEMissingEncryption() {
	p := &inboundmodel.OAuthProfileData{
		UserInfo: &inboundmodel.UserInfoConfig{ResponseType: inboundmodel.UserInfoResponseTypeJWE},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoJWERequiresEncryption)
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_NestedJWTMissingFields() {
	p := &inboundmodel.OAuthProfileData{
		UserInfo: &inboundmodel.UserInfoConfig{ResponseType: inboundmodel.UserInfoResponseTypeNESTEDJWT},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoNestedJWTRequiresAll)
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_UnsupportedResponseType() {
	p := &inboundmodel.OAuthProfileData{
		UserInfo: &inboundmodel.UserInfoConfig{ResponseType: "BOGUS"},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoUnsupportedResponseType)
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_SigningAlgRequiresResponseType() {
	p := &inboundmodel.OAuthProfileData{
		UserInfo: &inboundmodel.UserInfoConfig{SigningAlg: "RS256"},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoAlgRequiresResponseType)
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_EncryptionAlgRequiresResponseType() {
	p := &inboundmodel.OAuthProfileData{
		Certificate: &inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: "{}"},
		UserInfo: &inboundmodel.UserInfoConfig{
			EncryptionAlg: "RSA-OAEP-256", EncryptionEnc: "A256GCM",
		},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoAlgRequiresResponseType)
}

func (suite *InboundClientServiceTestSuite) TestValidateEntityInfoConfig_AllAlgsRequireResponseType() {
	p := &inboundmodel.OAuthProfileData{
		Certificate: &inboundmodel.Certificate{Type: cert.CertificateTypeJWKS, Value: "{}"},
		UserInfo: &inboundmodel.UserInfoConfig{
			SigningAlg: "RS256", EncryptionAlg: "RSA-OAEP-256", EncryptionEnc: "A256GCM",
		},
	}
	assert.ErrorIs(suite.T(), validateUserInfoConfig(p), ErrOAuthUserInfoAlgRequiresResponseType)
}

func (suite *InboundClientServiceTestSuite) TestResolveUserInfo_DefaultsResponseTypeToJSON() {
	out := resolveUserInfo(nil, nil)
	assert.Equal(suite.T(), inboundmodel.UserInfoResponseTypeJSON, out.ResponseType)
}

func (suite *InboundClientServiceTestSuite) TestResolveUserInfo_DefaultsResponseTypeToJSONForPartialConfig() {
	out := resolveUserInfo(&inboundmodel.UserInfoConfig{UserAttributes: []string{"email"}}, nil)
	assert.Equal(suite.T(), inboundmodel.UserInfoResponseTypeJSON, out.ResponseType)
}

func (suite *InboundClientServiceTestSuite) TestResolveUserInfo_PreservesExplicitResponseType() {
	in := &inboundmodel.UserInfoConfig{ResponseType: inboundmodel.UserInfoResponseTypeJWS, SigningAlg: "RS256"}
	out := resolveUserInfo(in, nil)
	assert.Equal(suite.T(), inboundmodel.UserInfoResponseTypeJWS, out.ResponseType)
}

func (suite *InboundClientServiceTestSuite) TestResolveUserInfo_FallsBackToIDTokenAttributes() {
	idToken := &inboundmodel.IDTokenConfig{UserAttributes: []string{"email"}}
	out := resolveUserInfo(&inboundmodel.UserInfoConfig{}, idToken)
	assert.Equal(suite.T(), []string{"email"}, out.UserAttributes)
	assert.Equal(suite.T(), inboundmodel.UserInfoResponseTypeJSON, out.ResponseType)
}

func (suite *InboundClientServiceTestSuite) TestResolveUserInfo_PreservesUserAttributesOverIDToken() {
	idToken := &inboundmodel.IDTokenConfig{UserAttributes: []string{"sub"}}
	out := resolveUserInfo(&inboundmodel.UserInfoConfig{UserAttributes: []string{"email"}}, idToken)
	assert.Equal(suite.T(), []string{"email"}, out.UserAttributes)
}

// validateOAuthProfile — verifies UserInfo validation is wired in.

func (suite *InboundClientServiceTestSuite) TestValidateOAuthProfile_PropagatesUserInfoErrors() {
	p := &inboundmodel.OAuthProfileData{
		RedirectURIs:            []string{"https://app.example.com/cb"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "client_secret_basic",
		UserInfo:                &inboundmodel.UserInfoConfig{SigningAlg: "BOGUS"},
	}
	assert.ErrorIs(suite.T(), validateOAuthProfile(p, true), ErrOAuthUserInfoUnsupportedSigningAlg)
}

func (suite *InboundClientServiceTestSuite) TestValidateOAuthProfile_NilProfile() {
	assert.NoError(suite.T(), validateOAuthProfile(nil, false))
}

// ----- BuildOAuthClient -----

func (suite *InboundClientServiceTestSuite) TestBuildOAuthClient_MapsAllFields() {
	dao := &inboundmodel.OAuthProfile{
		AppID: "app-1",
		OAuthProfile: &inboundmodel.OAuthProfileData{
			RedirectURIs:                       []string{"https://app/cb"},
			GrantTypes:                         []string{"authorization_code", "refresh_token"},
			ResponseTypes:                      []string{"code"},
			TokenEndpointAuthMethod:            "client_secret_basic",
			PKCERequired:                       true,
			PublicClient:                       false,
			RequirePushedAuthorizationRequests: true,
			Scopes:                             []string{"openid"},
			ScopeClaims:                        map[string][]string{"profile": {"name"}},
		},
	}
	client := BuildOAuthClient("entity-1", "client-1", "ou-1", dao)

	assert.Equal(suite.T(), "entity-1", client.AppID)
	assert.Equal(suite.T(), "client-1", client.ClientID)
	assert.Equal(suite.T(), "ou-1", client.OUID)
	assert.Equal(suite.T(), []string{"https://app/cb"}, client.RedirectURIs)
	assert.Equal(suite.T(), oauth2const.TokenEndpointAuthMethod("client_secret_basic"), client.TokenEndpointAuthMethod)
	assert.True(suite.T(), client.PKCERequired)
	assert.True(suite.T(), client.RequirePushedAuthorizationRequests)
	assert.Equal(suite.T(), []oauth2const.GrantType{"authorization_code", "refresh_token"}, client.GrantTypes)
	assert.Equal(suite.T(), []oauth2const.ResponseType{"code"}, client.ResponseTypes)
}

// ----- resolveAssertion -----

func (suite *InboundClientServiceTestSuite) TestResolveAssertion_NilInputUsesDefault() {
	out := resolveAssertion(nil, &inboundmodel.AssertionConfig{ValidityPeriod: 3600})
	assert.Equal(suite.T(), int64(3600), out.ValidityPeriod)
	assert.NotNil(suite.T(), out.UserAttributes)
}

func (suite *InboundClientServiceTestSuite) TestResolveAssertion_BothNilZeroValues() {
	out := resolveAssertion(nil, nil)
	assert.Equal(suite.T(), int64(0), out.ValidityPeriod)
	assert.NotNil(suite.T(), out.UserAttributes)
}

func (suite *InboundClientServiceTestSuite) TestResolveAssertion_InputZeroValidityFallsBack() {
	out := resolveAssertion(
		&inboundmodel.AssertionConfig{ValidityPeriod: 0, UserAttributes: []string{"sub"}},
		&inboundmodel.AssertionConfig{ValidityPeriod: 600},
	)
	assert.Equal(suite.T(), int64(600), out.ValidityPeriod)
	assert.Equal(suite.T(), []string{"sub"}, out.UserAttributes)
}

func (suite *InboundClientServiceTestSuite) TestResolveAssertion_InputOverridesDefault() {
	out := resolveAssertion(
		&inboundmodel.AssertionConfig{ValidityPeriod: 1200, UserAttributes: []string{"email"}},
		&inboundmodel.AssertionConfig{ValidityPeriod: 600},
	)
	assert.Equal(suite.T(), int64(1200), out.ValidityPeriod)
}

// ----- resolveOAuthTokens -----

func (suite *InboundClientServiceTestSuite) TestResolveOAuthTokens_NilInputUsesAssertion() {
	assertion := &inboundmodel.AssertionConfig{ValidityPeriod: 900, UserAttributes: []string{"email"}}
	at, idt := resolveOAuthTokens(nil, assertion)

	assert.Equal(suite.T(), int64(900), at.ValidityPeriod)
	assert.Equal(suite.T(), []string{"email"}, at.UserAttributes)
	assert.Equal(suite.T(), int64(900), idt.ValidityPeriod)
}

func (suite *InboundClientServiceTestSuite) TestResolveOAuthTokens_InputOverrides() {
	in := &inboundmodel.OAuthTokenConfig{
		AccessToken: &inboundmodel.AccessTokenConfig{ValidityPeriod: 60, UserAttributes: []string{"sub"}},
		IDToken:     &inboundmodel.IDTokenConfig{ValidityPeriod: 120, UserAttributes: []string{"email"}},
	}
	at, idt := resolveOAuthTokens(in, &inboundmodel.AssertionConfig{ValidityPeriod: 900})
	assert.Equal(suite.T(), int64(60), at.ValidityPeriod)
	assert.Equal(suite.T(), int64(120), idt.ValidityPeriod)
}

func (suite *InboundClientServiceTestSuite) TestResolveOAuthTokens_NilAssertionDoesNotPanic() {
	at, idt := resolveOAuthTokens(nil, nil)
	assert.NotNil(suite.T(), at)
	assert.NotNil(suite.T(), idt)
}

func (suite *InboundClientServiceTestSuite) TestResolveOAuthTokens_ZeroValidityFallsBackToAssertion() {
	in := &inboundmodel.OAuthTokenConfig{
		AccessToken: &inboundmodel.AccessTokenConfig{ValidityPeriod: 0},
		IDToken:     &inboundmodel.IDTokenConfig{ValidityPeriod: 0},
	}
	at, idt := resolveOAuthTokens(in, &inboundmodel.AssertionConfig{ValidityPeriod: 1800})
	assert.Equal(suite.T(), int64(1800), at.ValidityPeriod)
	assert.Equal(suite.T(), int64(1800), idt.ValidityPeriod)
}

// ----- resolveScopeClaims -----

func (suite *InboundClientServiceTestSuite) TestResolveScopeClaims_NilReturnsEmptyMap() {
	out := resolveScopeClaims(nil)
	assert.NotNil(suite.T(), out)
	assert.Empty(suite.T(), out)
}

func (suite *InboundClientServiceTestSuite) TestResolveScopeClaims_PassesThroughExistingMap() {
	in := map[string][]string{"profile": {"given_name"}}
	out := resolveScopeClaims(in)
	assert.Equal(suite.T(), in, out)
}

// ----- validateRedirectURIs error branches -----

func (suite *InboundClientServiceTestSuite) TestValidateRedirectURIs_SchemeWildcardRejected() {
	p := &inboundmodel.OAuthProfileData{
		RedirectURIs: []string{"htt*://app/cb"},
		GrantTypes:   []string{"authorization_code"},
	}
	assert.ErrorIs(suite.T(), validateRedirectURIs(p), ErrOAuthInvalidRedirectURI)
}

func (suite *InboundClientServiceTestSuite) TestValidateRedirectURIs_FragmentRejected() {
	p := &inboundmodel.OAuthProfileData{
		RedirectURIs: []string{"https://app/cb#frag"},
		GrantTypes:   []string{"authorization_code"},
	}
	assert.ErrorIs(suite.T(), validateRedirectURIs(p), ErrOAuthRedirectURIFragmentNotAllowed)
}

func (suite *InboundClientServiceTestSuite) TestValidateRedirectURIs_HostWildcardRejected() {
	p := &inboundmodel.OAuthProfileData{
		RedirectURIs: []string{"https://*.app.com/cb"},
		GrantTypes:   []string{"authorization_code"},
	}
	assert.ErrorIs(suite.T(), validateRedirectURIs(p), ErrOAuthInvalidRedirectURI)
}

func (suite *InboundClientServiceTestSuite) TestValidateRedirectURIs_QueryWildcardRejected() {
	p := &inboundmodel.OAuthProfileData{
		RedirectURIs: []string{"https://app/cb?x=*"},
		GrantTypes:   []string{"authorization_code"},
	}
	assert.ErrorIs(suite.T(), validateRedirectURIs(p), ErrOAuthInvalidRedirectURI)
}

func (suite *InboundClientServiceTestSuite) TestValidateRedirectURIs_MissingSchemeRejected() {
	p := &inboundmodel.OAuthProfileData{
		RedirectURIs: []string{"//app/cb"},
		GrantTypes:   []string{"authorization_code"},
	}
	assert.ErrorIs(suite.T(), validateRedirectURIs(p), ErrOAuthInvalidRedirectURI)
}

func (suite *InboundClientServiceTestSuite) TestValidateRedirectURIs_AuthCodeWithoutURIs() {
	p := &inboundmodel.OAuthProfileData{
		GrantTypes: []string{"authorization_code"},
	}
	assert.ErrorIs(suite.T(), validateRedirectURIs(p), ErrOAuthAuthCodeRequiresRedirectURIs)
}

// ----- containsInvalidWildcardSegment -----

func (suite *InboundClientServiceTestSuite) TestContainsInvalidWildcardSegment_PartialWildcard() {
	assert.True(suite.T(), containsInvalidWildcardSegment("/foo*"))
}

func (suite *InboundClientServiceTestSuite) TestContainsInvalidWildcardSegment_RegexMetachars() {
	assert.True(suite.T(), containsInvalidWildcardSegment("/[a-z]+"))
	assert.True(suite.T(), containsInvalidWildcardSegment("/foo|bar"))
	assert.True(suite.T(), containsInvalidWildcardSegment("/foo(x)"))
	assert.True(suite.T(), containsInvalidWildcardSegment("/foo$"))
}

func (suite *InboundClientServiceTestSuite) TestContainsInvalidWildcardSegment_Allowed() {
	assert.False(suite.T(), containsInvalidWildcardSegment("/foo/*/bar"))
	assert.False(suite.T(), containsInvalidWildcardSegment("/foo/**"))
	assert.False(suite.T(), containsInvalidWildcardSegment("/plain/path"))
}

// ----- FK validators -----

func (suite *InboundClientServiceTestSuite) TestValidateAuthFlowID_EmptyOrNoMgtIsNoOp() {
	svc := &inboundClientService{}
	assert.NoError(suite.T(), svc.validateAuthFlowID(context.Background(), ""))
}

func (suite *InboundClientServiceTestSuite) TestValidateAuthFlowID_InvalidReturnsError() {
	flowMgt := flowmgtmock.NewFlowMgtServiceInterfaceMock(suite.T())
	flowMgt.EXPECT().IsValidFlow(mock.Anything, "bad-flow", flowcommon.FlowTypeAuthentication).
		Return(false, nil)
	svc := &inboundClientService{flowMgt: flowMgt}
	assert.ErrorIs(suite.T(), svc.validateAuthFlowID(context.Background(), "bad-flow"), ErrFKInvalidAuthFlow)
}

func (suite *InboundClientServiceTestSuite) TestValidateAuthFlowID_ServerErrorPropagated() {
	flowMgt := flowmgtmock.NewFlowMgtServiceInterfaceMock(suite.T())
	flowMgt.EXPECT().IsValidFlow(mock.Anything, "fid", flowcommon.FlowTypeAuthentication).
		Return(false, &serviceerror.ServiceError{Code: "X"})
	svc := &inboundClientService{flowMgt: flowMgt}
	assert.ErrorIs(suite.T(), svc.validateAuthFlowID(context.Background(), "fid"), ErrFKFlowServerError)
}

func (suite *InboundClientServiceTestSuite) TestValidateAuthFlowID_ValidNoError() {
	flowMgt := flowmgtmock.NewFlowMgtServiceInterfaceMock(suite.T())
	flowMgt.EXPECT().IsValidFlow(mock.Anything, "good", flowcommon.FlowTypeAuthentication).
		Return(true, nil)
	svc := &inboundClientService{flowMgt: flowMgt}
	assert.NoError(suite.T(), svc.validateAuthFlowID(context.Background(), "good"))
}

func (suite *InboundClientServiceTestSuite) TestValidateRegistrationFlowID_AllBranches() {
	flowMgt := flowmgtmock.NewFlowMgtServiceInterfaceMock(suite.T())
	flowMgt.EXPECT().IsValidFlow(mock.Anything, "x", flowcommon.FlowTypeRegistration).Return(false, nil).Once()
	flowMgt.EXPECT().IsValidFlow(mock.Anything, "y", flowcommon.FlowTypeRegistration).
		Return(false, &serviceerror.ServiceError{Code: "E"}).Once()
	flowMgt.EXPECT().IsValidFlow(mock.Anything, "z", flowcommon.FlowTypeRegistration).Return(true, nil).Once()
	svc := &inboundClientService{flowMgt: flowMgt}
	assert.ErrorIs(suite.T(), svc.validateRegistrationFlowID(context.Background(), "x"), ErrFKInvalidRegistrationFlow)
	assert.ErrorIs(suite.T(), svc.validateRegistrationFlowID(context.Background(), "y"), ErrFKFlowServerError)
	assert.NoError(suite.T(), svc.validateRegistrationFlowID(context.Background(), "z"))
	assert.NoError(suite.T(), (&inboundClientService{}).validateRegistrationFlowID(context.Background(), ""))
}

func (suite *InboundClientServiceTestSuite) TestValidateThemeID_AllBranches() {
	tm := thememock.NewThemeMgtServiceInterfaceMock(suite.T())
	tm.EXPECT().IsThemeExist("missing").Return(false, nil).Once()
	tm.EXPECT().IsThemeExist("err").Return(false, &serviceerror.ServiceError{Code: "X"}).Once()
	tm.EXPECT().IsThemeExist("ok").Return(true, nil).Once()
	svc := &inboundClientService{themeMgt: tm}
	assert.ErrorIs(suite.T(), svc.validateThemeID("missing"), ErrFKThemeNotFound)
	assert.ErrorIs(suite.T(), svc.validateThemeID("err"), ErrFKThemeNotFound)
	assert.NoError(suite.T(), svc.validateThemeID("ok"))
	assert.NoError(suite.T(), (&inboundClientService{}).validateThemeID(""))
}

func (suite *InboundClientServiceTestSuite) TestValidateLayoutID_AllBranches() {
	lm := layoutmock.NewLayoutMgtServiceInterfaceMock(suite.T())
	lm.EXPECT().IsLayoutExist("missing").Return(false, nil).Once()
	lm.EXPECT().IsLayoutExist("err").Return(false, &serviceerror.ServiceError{Code: "X"}).Once()
	lm.EXPECT().IsLayoutExist("ok").Return(true, nil).Once()
	svc := &inboundClientService{layoutMgt: lm}
	assert.ErrorIs(suite.T(), svc.validateLayoutID("missing"), ErrFKLayoutNotFound)
	assert.ErrorIs(suite.T(), svc.validateLayoutID("err"), ErrFKLayoutNotFound)
	assert.NoError(suite.T(), svc.validateLayoutID("ok"))
	assert.NoError(suite.T(), (&inboundClientService{}).validateLayoutID(""))
}

func (suite *InboundClientServiceTestSuite) TestValidateAllowedUserTypes_NoOpWhenEmpty() {
	svc := &inboundClientService{}
	assert.NoError(suite.T(), svc.validateAllowedUserTypes(context.Background(), nil))
}

func (suite *InboundClientServiceTestSuite) TestValidateAllowedUserTypes_AllExist() {
	us := entitytypemock.NewEntityTypeServiceInterfaceMock(suite.T())
	us.EXPECT().GetEntityTypeList(mock.Anything, mock.Anything, mock.Anything, 0, false).Return(
		&entitytypepkg.EntityTypeListResponse{
			TotalResults: 1,
			Schemas:      []entitytypepkg.EntityTypeListItem{{Name: "person"}},
		}, nil)
	svc := &inboundClientService{entityType: us, logger: log.GetLogger()}
	assert.NoError(suite.T(), svc.validateAllowedUserTypes(context.Background(), []string{"person"}))
}

func (suite *InboundClientServiceTestSuite) TestValidateAllowedUserTypes_MissingType() {
	us := entitytypemock.NewEntityTypeServiceInterfaceMock(suite.T())
	us.EXPECT().GetEntityTypeList(mock.Anything, mock.Anything, mock.Anything, 0, false).Return(
		&entitytypepkg.EntityTypeListResponse{
			TotalResults: 1,
			Schemas:      []entitytypepkg.EntityTypeListItem{{Name: "person"}},
		}, nil)
	svc := &inboundClientService{entityType: us, logger: log.GetLogger()}
	err := svc.validateAllowedUserTypes(context.Background(), []string{"ghost"})
	assert.ErrorIs(suite.T(), err, ErrFKInvalidUserType)
}

func (suite *InboundClientServiceTestSuite) TestValidateAllowedUserTypes_EmptyTypeRejected() {
	us := entitytypemock.NewEntityTypeServiceInterfaceMock(suite.T())
	us.EXPECT().GetEntityTypeList(mock.Anything, mock.Anything, mock.Anything, 0, false).Return(
		&entitytypepkg.EntityTypeListResponse{TotalResults: 0}, nil)
	svc := &inboundClientService{entityType: us, logger: log.GetLogger()}
	assert.ErrorIs(suite.T(), svc.validateAllowedUserTypes(context.Background(), []string{""}), ErrFKInvalidUserType)
}

func (suite *InboundClientServiceTestSuite) TestValidateAllowedUserTypes_ServiceErrorPropagated() {
	us := entitytypemock.NewEntityTypeServiceInterfaceMock(suite.T())
	us.EXPECT().GetEntityTypeList(mock.Anything, mock.Anything, mock.Anything, 0, false).
		Return(nil, &serviceerror.ServiceError{Code: "ERR"})
	svc := &inboundClientService{entityType: us, logger: log.GetLogger()}
	assert.ErrorIs(suite.T(), svc.validateAllowedUserTypes(context.Background(), []string{"a"}), ErrFKInvalidUserType)
}

// ----- validateFKs aggregate -----

func (suite *InboundClientServiceTestSuite) TestValidateFKs_NilNoOp() {
	svc := &inboundClientService{}
	assert.NoError(suite.T(), svc.validateFKs(context.Background(), nil))
}

// ----- error wrappers -----

func TestCertOperationError_ErrorAndIsClientError(t *testing.T) {
	e := &CertOperationError{Underlying: &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		ErrorDescription: core.I18nMessage{DefaultValue: "bad cert"},
	}}
	assert.Equal(t, "bad cert", e.Error())
	assert.True(t, e.IsClientError())

	empty := &CertOperationError{}
	assert.Equal(t, "certificate operation failed", empty.Error())
	assert.False(t, empty.IsClientError())
}

func TestConsentSyncError_ErrorAndIsClientError(t *testing.T) {
	e := &ConsentSyncError{Underlying: &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		ErrorDescription: core.I18nMessage{DefaultValue: "consent down"},
	}}
	assert.Equal(t, "consent down", e.Error())
	assert.False(t, e.IsClientError())

	empty := &ConsentSyncError{}
	assert.Equal(t, "consent sync failed", empty.Error())
	assert.False(t, empty.IsClientError())
}

// ----- validateGrantAndResponseTypes branch coverage -----

func (suite *InboundClientServiceTestSuite) TestValidateGrantAndResponseTypes_InvalidResponseType() {
	p := &inboundmodel.OAuthProfileData{
		GrantTypes:    []string{"authorization_code"},
		ResponseTypes: []string{"bogus_rt"},
	}
	assert.ErrorIs(suite.T(), validateGrantAndResponseTypes(p), ErrOAuthInvalidResponseType)
}

func (suite *InboundClientServiceTestSuite) TestValidateGrantAndResponseTypes_ClientCredsWithResponseType() {
	p := &inboundmodel.OAuthProfileData{
		GrantTypes:    []string{"client_credentials"},
		ResponseTypes: []string{"code"},
	}
	assert.ErrorIs(suite.T(), validateGrantAndResponseTypes(p),
		ErrOAuthClientCredentialsCannotUseResponseTypes)
}

func (suite *InboundClientServiceTestSuite) TestValidateGrantAndResponseTypes_AuthCodeMissingCodeRT() {
	p := &inboundmodel.OAuthProfileData{
		GrantTypes:    []string{"authorization_code"},
		ResponseTypes: []string{},
	}
	assert.ErrorIs(suite.T(), validateGrantAndResponseTypes(p),
		ErrOAuthAuthCodeRequiresCodeResponseType)
}

func (suite *InboundClientServiceTestSuite) TestValidateGrantAndResponseTypes_RefreshTokenSole() {
	p := &inboundmodel.OAuthProfileData{
		GrantTypes: []string{"refresh_token"},
	}
	assert.ErrorIs(suite.T(), validateGrantAndResponseTypes(p),
		ErrOAuthRefreshTokenCannotBeSoleGrant)
}

func (suite *InboundClientServiceTestSuite) TestValidateGrantAndResponseTypes_PKCEWithoutAuthCode() {
	p := &inboundmodel.OAuthProfileData{
		GrantTypes:   []string{"client_credentials"},
		PKCERequired: true,
	}
	assert.ErrorIs(suite.T(), validateGrantAndResponseTypes(p), ErrOAuthPKCERequiresAuthCode)
}

func (suite *InboundClientServiceTestSuite) TestValidateGrantAndResponseTypes_ResponseTypeWithoutAuthCode() {
	p := &inboundmodel.OAuthProfileData{
		GrantTypes:    []string{"client_credentials"},
		ResponseTypes: []string{"code"},
	}
	// client_credentials + response_types triggers the earlier rule
	assert.Error(suite.T(), validateGrantAndResponseTypes(p))
}

func (suite *InboundClientServiceTestSuite) TestValidateGrantAndResponseTypes_HappyAuthCode() {
	p := &inboundmodel.OAuthProfileData{
		GrantTypes:    []string{"authorization_code"},
		ResponseTypes: []string{"code"},
	}
	assert.NoError(suite.T(), validateGrantAndResponseTypes(p))
}

func (suite *InboundClientServiceTestSuite) TestValidateGrantAndResponseTypes_HappyClientCredentials() {
	p := &inboundmodel.OAuthProfileData{
		GrantTypes: []string{"client_credentials"},
	}
	assert.NoError(suite.T(), validateGrantAndResponseTypes(p))
}

// ----- validatePublicClient branch coverage -----

func (suite *InboundClientServiceTestSuite) TestValidatePublicClient_NonNoneAuthMethod() {
	p := &inboundmodel.OAuthProfileData{
		TokenEndpointAuthMethod: "client_secret_basic",
		PKCERequired:            true,
	}
	assert.ErrorIs(suite.T(), validatePublicClient(p), ErrOAuthPublicClientMustUseNoneAuth)
}

func (suite *InboundClientServiceTestSuite) TestValidatePublicClient_HappyPath() {
	p := &inboundmodel.OAuthProfileData{
		TokenEndpointAuthMethod: "none",
		PKCERequired:            true,
	}
	assert.NoError(suite.T(), validatePublicClient(p))
}

// ----- validateFKs aggregate paths -----

func (suite *InboundClientServiceTestSuite) TestValidateFKs_AuthFlowErrorPropagated() {
	flowMgt := flowmgtmock.NewFlowMgtServiceInterfaceMock(suite.T())
	flowMgt.EXPECT().IsValidFlow(mock.Anything, "bad", flowcommon.FlowTypeAuthentication).Return(false, nil)
	svc := &inboundClientService{flowMgt: flowMgt}
	c := &inboundmodel.InboundClient{AuthFlowID: "bad"}
	assert.ErrorIs(suite.T(), svc.validateFKs(context.Background(), c), ErrFKInvalidAuthFlow)
}

func (suite *InboundClientServiceTestSuite) TestValidateFKs_AllPassWithEmptyOptionals() {
	svc := &inboundClientService{}
	c := &inboundmodel.InboundClient{}
	assert.NoError(suite.T(), svc.validateFKs(context.Background(), c))
}

// ----- consent helpers -----

func TestExtractRequestedAttributesFromInbound_AllNil(t *testing.T) {
	out := extractRequestedAttributesFromInbound(nil, nil)
	assert.Empty(t, out)
}

func TestExtractRequestedAttributesFromInbound_FromAssertionOnly(t *testing.T) {
	c := &inboundmodel.InboundClient{
		Assertion: &inboundmodel.AssertionConfig{UserAttributes: []string{"email", "sub"}},
	}
	out := extractRequestedAttributesFromInbound(c, nil)
	assert.Len(t, out, 2)
	assert.True(t, out["email"])
	assert.True(t, out["sub"])
}

func TestExtractRequestedAttributesFromInbound_DedupsAcrossSources(t *testing.T) {
	c := &inboundmodel.InboundClient{
		Assertion: &inboundmodel.AssertionConfig{UserAttributes: []string{"email"}},
	}
	p := &inboundmodel.OAuthProfileData{
		Token: &inboundmodel.OAuthTokenConfig{
			AccessToken: &inboundmodel.AccessTokenConfig{UserAttributes: []string{"email", "given_name"}},
			IDToken:     &inboundmodel.IDTokenConfig{UserAttributes: []string{"family_name"}},
		},
		UserInfo: &inboundmodel.UserInfoConfig{UserAttributes: []string{"email", "picture"}},
	}
	out := extractRequestedAttributesFromInbound(c, p)
	assert.Len(t, out, 4)
	assert.True(t, out["email"])
	assert.True(t, out["given_name"])
	assert.True(t, out["family_name"])
	assert.True(t, out["picture"])
}

func TestExtractRequestedAttributesFromInbound_NilSubFields(t *testing.T) {
	p := &inboundmodel.OAuthProfileData{
		Token:    &inboundmodel.OAuthTokenConfig{},
		UserInfo: nil,
	}
	out := extractRequestedAttributesFromInbound(nil, p)
	assert.Empty(t, out)
}

func TestAttributesToPurposeElements_EmptyMap(t *testing.T) {
	out := attributesToPurposeElements(map[string]bool{})
	assert.Empty(t, out)
}

func TestAttributesToPurposeElements_PopulatedMap(t *testing.T) {
	out := attributesToPurposeElements(map[string]bool{"email": true, "sub": true})
	assert.Len(t, out, 2)
	for _, el := range out {
		assert.False(t, el.IsMandatory)
	}
}

// ----- wrapConsentServiceError -----

func TestWrapConsentServiceError_NilReturnsNil(t *testing.T) {
	s := &inboundClientService{}
	assert.Nil(t, s.wrapConsentServiceError(nil))
}

func TestWrapConsentServiceError_WrapsServiceError(t *testing.T) {
	s := &inboundClientService{}
	se := &serviceerror.ServiceError{Code: "X", Type: serviceerror.ClientErrorType}
	wrapped := s.wrapConsentServiceError(se)
	var ce *ConsentSyncError
	assert.True(t, errors.As(wrapped, &ce))
	assert.Equal(t, se, ce.Underlying)
}

// ----- validateUniqueInboundClientID -----

func (suite *InboundClientServiceTestSuite) TestValidateUniqueInboundClientID_NotExisting() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().InboundClientExists(mock.Anything, "x").Return(false, nil)
	c := &inboundmodel.InboundClient{ID: "x"}
	assert.NoError(suite.T(), validateUniqueInboundClientID(context.Background(), store, c))
}

func (suite *InboundClientServiceTestSuite) TestValidateUniqueInboundClientID_DuplicateRejected() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().InboundClientExists(mock.Anything, "x").Return(true, nil)
	c := &inboundmodel.InboundClient{ID: "x"}
	err := validateUniqueInboundClientID(context.Background(), store, c)
	assert.ErrorContains(suite.T(), err, "duplicate entity ID")
}

func (suite *InboundClientServiceTestSuite) TestValidateUniqueInboundClientID_StoreErrorPropagated() {
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().InboundClientExists(mock.Anything, "x").Return(false, errors.New("db down"))
	c := &inboundmodel.InboundClient{ID: "x"}
	err := validateUniqueInboundClientID(context.Background(), store, c)
	assert.ErrorContains(suite.T(), err, "failed to check inbound client existence")
}

// ----- GetOAuthClientByClientID -----

func (suite *InboundClientServiceTestSuite) TestGetOAuthClientByClientID_NoEntityProvider() {
	svc := newServiceForTest(newInboundClientStoreInterfaceMock(suite.T())).(*inboundClientService)
	got, err := svc.GetOAuthClientByClientID(context.Background(), "client-1")
	assert.ErrorContains(suite.T(), err, "entity provider not configured")
	assert.Nil(suite.T(), got)
}

func (suite *InboundClientServiceTestSuite) TestGetOAuthClientByClientID_EmptyClientID() {
	ep := entityprovidermock.NewEntityProviderInterfaceMock(suite.T())
	svc := &inboundClientService{
		entityProvider: ep,
		store:          newInboundClientStoreInterfaceMock(suite.T()),
	}
	got, err := svc.GetOAuthClientByClientID(context.Background(), "")
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), got)
}

func (suite *InboundClientServiceTestSuite) TestGetOAuthClientByClientID_EntityNotFound() {
	ep := entityprovidermock.NewEntityProviderInterfaceMock(suite.T())
	ep.EXPECT().IdentifyEntity(mock.Anything).Return(nil, &entityprovider.EntityProviderError{
		Code: entityprovider.ErrorCodeEntityNotFound,
	})
	svc := &inboundClientService{entityProvider: ep, store: newInboundClientStoreInterfaceMock(suite.T())}
	got, err := svc.GetOAuthClientByClientID(context.Background(), "missing")
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), got)
}

func (suite *InboundClientServiceTestSuite) TestGetOAuthClientByClientID_IdentifyErrorPropagated() {
	ep := entityprovidermock.NewEntityProviderInterfaceMock(suite.T())
	ep.EXPECT().IdentifyEntity(mock.Anything).Return(nil, &entityprovider.EntityProviderError{
		Code: entityprovider.ErrorCodeSystemError, Message: "boom",
	})
	svc := &inboundClientService{entityProvider: ep, store: newInboundClientStoreInterfaceMock(suite.T())}
	got, err := svc.GetOAuthClientByClientID(context.Background(), "x")
	assert.ErrorContains(suite.T(), err, "failed to resolve client_id")
	assert.Nil(suite.T(), got)
}

func (suite *InboundClientServiceTestSuite) TestGetOAuthClientByClientID_NilEntityID() {
	ep := entityprovidermock.NewEntityProviderInterfaceMock(suite.T())
	ep.EXPECT().IdentifyEntity(mock.Anything).Return(nil, nil)
	svc := &inboundClientService{entityProvider: ep, store: newInboundClientStoreInterfaceMock(suite.T())}
	got, err := svc.GetOAuthClientByClientID(context.Background(), "x")
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), got)
}

const testServiceEntityID = "ent-1"

func (suite *InboundClientServiceTestSuite) TestGetOAuthClientByClientID_GetEntityNotFound() {
	id := testServiceEntityID
	ep := entityprovidermock.NewEntityProviderInterfaceMock(suite.T())
	ep.EXPECT().IdentifyEntity(mock.Anything).Return(&id, nil)
	ep.EXPECT().GetEntity(id).Return(nil, &entityprovider.EntityProviderError{
		Code: entityprovider.ErrorCodeEntityNotFound,
	})
	svc := &inboundClientService{entityProvider: ep, store: newInboundClientStoreInterfaceMock(suite.T())}
	got, err := svc.GetOAuthClientByClientID(context.Background(), "x")
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), got)
}

func (suite *InboundClientServiceTestSuite) TestGetOAuthClientByClientID_OAuthProfileNotFoundReturnsNil() {
	id := testServiceEntityID
	ep := entityprovidermock.NewEntityProviderInterfaceMock(suite.T())
	ep.EXPECT().IdentifyEntity(mock.Anything).Return(&id, nil)
	ep.EXPECT().GetEntity(id).Return(&entityprovider.Entity{ID: id, OUID: "ou-1"}, nil)

	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().GetOAuthProfileByEntityID(mock.Anything, id).Return(nil, ErrInboundClientNotFound)

	svc := &inboundClientService{entityProvider: ep, store: store}
	got, err := svc.GetOAuthClientByClientID(context.Background(), "x")
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), got)
}

func (suite *InboundClientServiceTestSuite) TestGetOAuthClientByClientID_StoreErrorPropagated() {
	id := testServiceEntityID
	ep := entityprovidermock.NewEntityProviderInterfaceMock(suite.T())
	ep.EXPECT().IdentifyEntity(mock.Anything).Return(&id, nil)
	ep.EXPECT().GetEntity(id).Return(&entityprovider.Entity{ID: id, OUID: "ou-1"}, nil)

	storeErr := errors.New("db down")
	store := newInboundClientStoreInterfaceMock(suite.T())
	store.EXPECT().GetOAuthProfileByEntityID(mock.Anything, id).Return(nil, storeErr)

	svc := &inboundClientService{entityProvider: ep, store: store}
	got, err := svc.GetOAuthClientByClientID(context.Background(), "x")
	assert.ErrorIs(suite.T(), err, storeErr)
	assert.Nil(suite.T(), got)
}
