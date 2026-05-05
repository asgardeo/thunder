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

package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/agent/model"
	"github.com/asgardeo/thunder/internal/entity"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/pkg/inboundclient"
	inboundmodel "github.com/asgardeo/thunder/pkg/inboundclient/model"
	oauth2const "github.com/asgardeo/thunder/pkg/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/tests/mocks/entitymock"
	"github.com/asgardeo/thunder/tests/mocks/inboundclientmock"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
)

const (
	testAgentID   = "agent-id-123"
	testAgentName = "test-agent"
	testAgentType = "employee"
	testOUID      = "ou-id-abc"
)

// AgentServiceTestSuite groups all agent service unit tests.
type AgentServiceTestSuite struct {
	suite.Suite
}

func TestAgentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AgentServiceTestSuite))
}

// setupService wires a service with permissive default mocks. Tests override specific
// expectations as needed.
func (suite *AgentServiceTestSuite) setupService() (
	*agentService,
	*entitymock.EntityServiceInterfaceMock,
	*inboundclientmock.InboundClientServiceInterfaceMock,
	*oumock.OrganizationUnitServiceInterfaceMock,
) {
	mockEntity := entitymock.NewEntityServiceInterfaceMock(suite.T())
	mockInbound := inboundclientmock.NewInboundClientServiceInterfaceMock(suite.T())
	mockOU := oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())

	// Permissive defaults — tests narrow these as needed.
	mockEntity.On("GetEntity", mock.Anything, mock.Anything).
		Maybe().Return((*entity.Entity)(nil), entity.ErrEntityNotFound)
	mockEntity.On("CreateEntity", mock.Anything, mock.Anything, mock.Anything).
		Maybe().Return(&entity.Entity{ID: testAgentID}, nil)
	mockEntity.On("DeleteEntity", mock.Anything, mock.Anything).
		Maybe().Return(nil)
	mockEntity.On("IdentifyEntity", mock.Anything, mock.Anything).
		Maybe().Return((*string)(nil), entity.ErrEntityNotFound)
	mockEntity.On("UpdateEntity", mock.Anything, mock.Anything, mock.Anything).
		Maybe().Return(&entity.Entity{}, nil)
	mockEntity.On("UpdateSystemCredentials", mock.Anything, mock.Anything, mock.Anything).
		Maybe().Return(nil)
	mockEntity.On("GetEntityList", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Maybe().Return([]entity.Entity{}, nil)
	mockEntity.On("GetEntityListCount", mock.Anything, mock.Anything, mock.Anything).
		Maybe().Return(0, nil)
	mockEntity.On("GetEntityListByOUIDs", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).
		Maybe().Return([]entity.Entity{}, nil)
	mockEntity.On("GetEntityListCountByOUIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Maybe().Return(0, nil)
	mockEntity.On("GetEntityGroups", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Maybe().Return([]entity.EntityGroup{}, nil)
	mockEntity.On("GetGroupCountForEntity", mock.Anything, mock.Anything).
		Maybe().Return(0, nil)

	mockInbound.On("GetInboundClientByEntityID", mock.Anything, mock.Anything).
		Maybe().Return((*inboundmodel.InboundClient)(nil), inboundclient.ErrInboundClientNotFound)
	mockInbound.On("GetOAuthProfileByEntityID", mock.Anything, mock.Anything).
		Maybe().Return((*inboundmodel.OAuthProfile)(nil), inboundclient.ErrInboundClientNotFound)
	mockInbound.On("GetCertificate", mock.Anything, mock.Anything, mock.Anything).
		Maybe().Return((*inboundmodel.Certificate)(nil), (*inboundclient.CertOperationError)(nil))
	mockInbound.On("CreateInboundClient", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).
		Maybe().Return(nil)
	mockInbound.On("UpdateInboundClient", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Maybe().Return(nil)
	mockInbound.On("DeleteInboundClient", mock.Anything, mock.Anything).
		Maybe().Return(nil)

	mockOU.On("IsOrganizationUnitExists", mock.Anything, mock.Anything).
		Maybe().Return(true, (*serviceerror.ServiceError)(nil))
	mockOU.On("GetOrganizationUnitByPath", mock.Anything, mock.Anything).
		Maybe().Return(oupkg.OrganizationUnit{ID: testOUID}, (*serviceerror.ServiceError)(nil))
	mockOU.On("GetOrganizationUnitHandlesByIDs", mock.Anything, mock.Anything).
		Maybe().Return(map[string]string{}, (*serviceerror.ServiceError)(nil))

	svc := &agentService{
		logger:               log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AgentService")),
		entityService:        mockEntity,
		inboundClientService: mockInbound,
		ouService:            mockOU,
	}
	return svc, mockEntity, mockInbound, mockOU
}

// buildAgentEntityFixture returns an entity.Entity with system attributes for the given fields.
func buildAgentEntityFixture(name, description, owner, clientID string) *entity.Entity {
	attrs := map[string]interface{}{}
	if name != "" {
		attrs[fieldName] = name
	}
	if description != "" {
		attrs[fieldDescription] = description
	}
	if owner != "" {
		attrs[fieldOwner] = owner
	}
	if clientID != "" {
		attrs[fieldClientID] = clientID
	}
	sysAttrs, _ := json.Marshal(attrs)
	return &entity.Entity{
		ID:               testAgentID,
		Category:         entity.EntityCategoryAgent,
		Type:             testAgentType,
		State:            entity.EntityStateActive,
		OUID:             testOUID,
		SystemAttributes: sysAttrs,
	}
}

// --- pure helper tests ---

func (suite *AgentServiceTestSuite) TestNeedsInboundClient_NilRequest() {
	assert.False(suite.T(), needsInboundClient(nil))
}

func (suite *AgentServiceTestSuite) TestNeedsInboundClient_EmptyRequest() {
	assert.False(suite.T(), needsInboundClient(&model.CreateAgentRequest{}))
}

func (suite *AgentServiceTestSuite) TestNeedsInboundClient_WithAuthFlowID() {
	req := &model.CreateAgentRequest{AuthFlowID: "flow-1"}
	assert.True(suite.T(), needsInboundClient(req))
}

func (suite *AgentServiceTestSuite) TestNeedsInboundClient_WithInboundAuthConfig() {
	req := &model.CreateAgentRequest{
		InboundAuthConfig: []model.InboundAuthConfig{
			{Type: model.OAuthInboundAuthType, Config: &model.OAuthAgentConfig{}},
		},
	}
	assert.True(suite.T(), needsInboundClient(req))
}

func (suite *AgentServiceTestSuite) TestNeedsInboundClient_WithAllowedUserTypes() {
	req := &model.CreateAgentRequest{AllowedUserTypes: []string{"employee"}}
	assert.True(suite.T(), needsInboundClient(req))
}

func (suite *AgentServiceTestSuite) TestUpdateNeedsInboundClient_NilRequest() {
	assert.False(suite.T(), updateNeedsInboundClient(nil))
}

func (suite *AgentServiceTestSuite) TestUpdateNeedsInboundClient_EmptyRequest() {
	assert.False(suite.T(), updateNeedsInboundClient(&model.UpdateAgentRequest{}))
}

func (suite *AgentServiceTestSuite) TestUpdateNeedsInboundClient_WithThemeID() {
	req := &model.UpdateAgentRequest{ThemeID: "theme-abc"}
	assert.True(suite.T(), updateNeedsInboundClient(req))
}

func (suite *AgentServiceTestSuite) TestRequiresClientSecret_NilConfig() {
	assert.False(suite.T(), requiresClientSecret(nil))
}

func (suite *AgentServiceTestSuite) TestRequiresClientSecret_PublicClient() {
	cfg := &model.OAuthAgentConfig{PublicClient: true}
	assert.False(suite.T(), requiresClientSecret(cfg))
}

func (suite *AgentServiceTestSuite) TestRequiresClientSecret_ClientSecretBasic() {
	cfg := &model.OAuthAgentConfig{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
	}
	assert.True(suite.T(), requiresClientSecret(cfg))
}

func (suite *AgentServiceTestSuite) TestRequiresClientSecret_NoneMethod() {
	cfg := &model.OAuthAgentConfig{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
	}
	assert.False(suite.T(), requiresClientSecret(cfg))
}

func (suite *AgentServiceTestSuite) TestRequiresClientSecret_DefaultIsTrue() {
	cfg := &model.OAuthAgentConfig{}
	assert.True(suite.T(), requiresClientSecret(cfg))
}

// --- readSystemAttributes / buildSystemAttributesJSON ---

func (suite *AgentServiceTestSuite) TestReadSystemAttributes_Empty() {
	name, desc, owner, clientID := readSystemAttributes(nil)
	assert.Empty(suite.T(), name)
	assert.Empty(suite.T(), desc)
	assert.Empty(suite.T(), owner)
	assert.Empty(suite.T(), clientID)
}

func (suite *AgentServiceTestSuite) TestReadSystemAttributes_AllFields() {
	raw, _ := json.Marshal(map[string]interface{}{
		"name":        "my-agent",
		"description": "desc",
		"owner":       "alice",
		"clientId":    "cid-123",
	})
	name, desc, owner, clientID := readSystemAttributes(raw)
	assert.Equal(suite.T(), "my-agent", name)
	assert.Equal(suite.T(), "desc", desc)
	assert.Equal(suite.T(), "alice", owner)
	assert.Equal(suite.T(), "cid-123", clientID)
}

func (suite *AgentServiceTestSuite) TestBuildSystemAttributesJSON_AllFields() {
	raw, err := buildSystemAttributesJSON("n", "d", "o", "c")
	suite.Require().NoError(err)
	suite.Require().NotNil(raw)
	name, desc, owner, clientID := readSystemAttributes(raw)
	assert.Equal(suite.T(), "n", name)
	assert.Equal(suite.T(), "d", desc)
	assert.Equal(suite.T(), "o", owner)
	assert.Equal(suite.T(), "c", clientID)
}

func (suite *AgentServiceTestSuite) TestBuildSystemAttributesJSON_EmptyFields() {
	raw, err := buildSystemAttributesJSON("", "", "", "")
	suite.Require().NoError(err)
	assert.Nil(suite.T(), raw)
}

// --- removeSecrets ---

func (suite *AgentServiceTestSuite) TestRemoveSecrets_Nil() {
	assert.Nil(suite.T(), removeSecrets(nil))
}

func (suite *AgentServiceTestSuite) TestRemoveSecrets_RemovesClientSecret() {
	in := []model.InboundAuthConfig{
		{
			Type: model.OAuthInboundAuthType,
			Config: &model.OAuthAgentConfig{
				ClientID:     "cid",
				ClientSecret: "secret",
			},
		},
	}
	out := removeSecrets(in)
	suite.Require().Len(out, 1)
	assert.Equal(suite.T(), "cid", out[0].Config.ClientID)
	assert.Empty(suite.T(), out[0].Config.ClientSecret)
}

// --- validateBaseFields ---

func (suite *AgentServiceTestSuite) TestValidateBaseFields_MissingName() {
	svcErr := validateBaseFields("", "type")
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorInvalidAgentName.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestValidateBaseFields_MissingType() {
	svcErr := validateBaseFields("name", "")
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorInvalidAgentType.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestValidateBaseFields_Valid() {
	svcErr := validateBaseFields("name", "type")
	assert.Nil(suite.T(), svcErr)
}

// --- validatePaginationParams ---

func (suite *AgentServiceTestSuite) TestValidatePaginationParams_NegativeLimit() {
	svcErr := validatePaginationParams(-1, 0)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorInvalidLimit.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestValidatePaginationParams_LimitOver100() {
	svcErr := validatePaginationParams(101, 0)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorInvalidLimit.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestValidatePaginationParams_NegativeOffset() {
	svcErr := validatePaginationParams(10, -1)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorInvalidOffset.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestValidatePaginationParams_Valid() {
	svcErr := validatePaginationParams(10, 0)
	assert.Nil(suite.T(), svcErr)
}

// --- CreateAgent ---

func (suite *AgentServiceTestSuite) TestCreateAgent_NilRequest() {
	svc, _, _, _ := suite.setupService()
	resp, svcErr := svc.CreateAgent(context.Background(), nil)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorInvalidRequestFormat.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestCreateAgent_MissingName() {
	svc, _, _, _ := suite.setupService()
	req := &model.CreateAgentRequest{Type: testAgentType, OUID: testOUID}
	resp, svcErr := svc.CreateAgent(context.Background(), req)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorInvalidAgentName.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestCreateAgent_MissingType() {
	svc, _, _, _ := suite.setupService()
	req := &model.CreateAgentRequest{Name: testAgentName, OUID: testOUID}
	resp, svcErr := svc.CreateAgent(context.Background(), req)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorInvalidAgentType.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestCreateAgent_OUNotFound() {
	svc, _, _, mockOU := suite.setupService()
	clearMockCalls(mockOU, "IsOrganizationUnitExists")
	mockOU.On("IsOrganizationUnitExists", mock.Anything, testOUID).Return(false, (*serviceerror.ServiceError)(nil))

	req := &model.CreateAgentRequest{Name: testAgentName, Type: testAgentType, OUID: testOUID}
	resp, svcErr := svc.CreateAgent(context.Background(), req)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorOrganizationUnitNotFound.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestCreateAgent_NameAlreadyExists() {
	svc, mockEntity, _, _ := suite.setupService()
	existingID := "existing-agent-id"
	clearMockCalls(mockEntity, "IdentifyEntity")
	mockEntity.On("IdentifyEntity", mock.Anything, mock.Anything).Return(&existingID, nil)
	clearMockCalls(mockEntity, "GetEntity")
	mockEntity.On("GetEntity", mock.Anything, existingID).Return(
		&entity.Entity{ID: existingID, Category: entity.EntityCategoryAgent}, nil)

	req := &model.CreateAgentRequest{Name: testAgentName, Type: testAgentType, OUID: testOUID}
	resp, svcErr := svc.CreateAgent(context.Background(), req)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorAgentAlreadyExistsWithName.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestCreateAgent_EntityOnly_Success() {
	svc, mockEntity, mockInbound, _ := suite.setupService()

	createdEntity := buildAgentEntityFixture(testAgentName, "", "", "")
	clearMockCalls(mockEntity, "CreateEntity")
	mockEntity.On("CreateEntity", mock.Anything, mock.Anything, mock.Anything).
		Return(createdEntity, nil)

	req := &model.CreateAgentRequest{
		Name: testAgentName,
		Type: testAgentType,
		OUID: testOUID,
	}
	resp, svcErr := svc.CreateAgent(context.Background(), req)
	suite.Require().Nil(svcErr)
	suite.Require().NotNil(resp)
	assert.Equal(suite.T(), testAgentName, resp.Name)
	assert.Equal(suite.T(), testAgentType, resp.Type)

	// No inbound client should be created for entity-only agents.
	mockInbound.AssertNotCalled(suite.T(), "CreateInboundClient")
}

func (suite *AgentServiceTestSuite) TestCreateAgent_WithInboundAuth_Success() {
	svc, mockEntity, mockInbound, _ := suite.setupService()

	createdEntity := buildAgentEntityFixture(testAgentName, "", "", "")
	clearMockCalls(mockEntity, "CreateEntity")
	mockEntity.On("CreateEntity", mock.Anything, mock.Anything, mock.Anything).
		Return(createdEntity, nil)

	clearMockCalls(mockInbound, "CreateInboundClient")
	mockInbound.On("CreateInboundClient", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(nil)

	req := &model.CreateAgentRequest{
		Name:       testAgentName,
		Type:       testAgentType,
		OUID:       testOUID,
		AuthFlowID: "flow-1",
	}
	resp, svcErr := svc.CreateAgent(context.Background(), req)
	suite.Require().Nil(svcErr)
	suite.Require().NotNil(resp)
	assert.Equal(suite.T(), "flow-1", resp.AuthFlowID)
	mockInbound.AssertCalled(suite.T(), "CreateInboundClient", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (suite *AgentServiceTestSuite) TestCreateAgent_FlowIDResolvedToDefault() {
	svc, mockEntity, mockInbound, _ := suite.setupService()

	createdEntity := buildAgentEntityFixture(testAgentName, "", "", "")
	clearMockCalls(mockEntity, "CreateEntity")
	mockEntity.On("CreateEntity", mock.Anything, mock.Anything, mock.Anything).
		Return(createdEntity, nil)

	clearMockCalls(mockInbound, "CreateInboundClient")
	mockInbound.On("CreateInboundClient", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			client := args.Get(1).(*inboundmodel.InboundClient)
			client.AuthFlowID = "default-flow-id"
			client.RegistrationFlowID = "default-reg-flow-id"
		}).Return(nil)

	req := &model.CreateAgentRequest{
		Name: testAgentName,
		Type: testAgentType,
		OUID: testOUID,
		InboundAuthConfig: []model.InboundAuthConfig{
			{
				Type: model.OAuthInboundAuthType,
				Config: &model.OAuthAgentConfig{
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}
	resp, svcErr := svc.CreateAgent(context.Background(), req)
	suite.Require().Nil(svcErr)
	suite.Require().NotNil(resp)
	assert.Equal(suite.T(), "default-flow-id", resp.AuthFlowID)
	assert.Equal(suite.T(), "default-reg-flow-id", resp.RegistrationFlowID)
}

func (suite *AgentServiceTestSuite) TestCreateAgent_WithOAuth_Success() {
	svc, mockEntity, mockInbound, _ := suite.setupService()

	createdEntity := buildAgentEntityFixture(testAgentName, "", "", "cid-xxx")
	clearMockCalls(mockEntity, "CreateEntity")
	mockEntity.On("CreateEntity", mock.Anything, mock.Anything, mock.Anything).
		Return(createdEntity, nil)

	clearMockCalls(mockInbound, "CreateInboundClient")
	mockInbound.On("CreateInboundClient", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(nil)

	req := &model.CreateAgentRequest{
		Name:       testAgentName,
		Type:       testAgentType,
		OUID:       testOUID,
		AuthFlowID: "flow-1",
		InboundAuthConfig: []model.InboundAuthConfig{
			{
				Type: model.OAuthInboundAuthType,
				Config: &model.OAuthAgentConfig{
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}
	resp, svcErr := svc.CreateAgent(context.Background(), req)
	suite.Require().Nil(svcErr)
	suite.Require().NotNil(resp)
	suite.Require().Len(resp.InboundAuthConfig, 1)
	assert.Equal(suite.T(), model.OAuthInboundAuthType, resp.InboundAuthConfig[0].Type)
	assert.NotEmpty(suite.T(), resp.InboundAuthConfig[0].Config.ClientID)
	assert.NotEmpty(suite.T(), resp.InboundAuthConfig[0].Config.ClientSecret)
}

func (suite *AgentServiceTestSuite) TestCreateAgent_EntityCreationFails() {
	svc, mockEntity, _, _ := suite.setupService()

	clearMockCalls(mockEntity, "CreateEntity")
	mockEntity.On("CreateEntity", mock.Anything, mock.Anything, mock.Anything).
		Return((*entity.Entity)(nil), entity.ErrSchemaValidationFailed)

	req := &model.CreateAgentRequest{Name: testAgentName, Type: testAgentType, OUID: testOUID}
	resp, svcErr := svc.CreateAgent(context.Background(), req)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorSchemaValidationFailed.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestCreateAgent_InboundCreationFails_CompensatesEntity() {
	svc, mockEntity, mockInbound, _ := suite.setupService()

	createdEntity := buildAgentEntityFixture(testAgentName, "", "", "")
	clearMockCalls(mockEntity, "CreateEntity")
	mockEntity.On("CreateEntity", mock.Anything, mock.Anything, mock.Anything).
		Return(createdEntity, nil)

	clearMockCalls(mockInbound, "CreateInboundClient")
	mockInbound.On("CreateInboundClient", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).
		Return(inboundclient.ErrOAuthInvalidGrantType)

	clearMockCalls(mockEntity, "DeleteEntity")
	mockEntity.On("DeleteEntity", mock.Anything, mock.Anything).Return(nil)

	req := &model.CreateAgentRequest{
		Name:       testAgentName,
		Type:       testAgentType,
		OUID:       testOUID,
		AuthFlowID: "flow-1",
	}
	resp, svcErr := svc.CreateAgent(context.Background(), req)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorInvalidGrantType.Code, svcErr.Code)
	mockEntity.AssertCalled(suite.T(), "DeleteEntity", mock.Anything, mock.Anything)
}

// --- GetAgent ---

func (suite *AgentServiceTestSuite) TestGetAgent_EmptyID() {
	svc, _, _, _ := suite.setupService()
	resp, svcErr := svc.GetAgent(context.Background(), "", false)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorMissingAgentID.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestGetAgent_NotFound() {
	svc, _, _, _ := suite.setupService()
	// Default mock returns ErrEntityNotFound.
	resp, svcErr := svc.GetAgent(context.Background(), testAgentID, false)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorAgentNotFound.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestGetAgent_WrongCategory() {
	svc, mockEntity, _, _ := suite.setupService()

	wrongCatEntity := buildAgentEntityFixture(testAgentName, "", "", "")
	wrongCatEntity.Category = entity.EntityCategoryUser

	clearMockCalls(mockEntity, "GetEntity")
	mockEntity.On("GetEntity", mock.Anything, testAgentID).Return(wrongCatEntity, nil)

	resp, svcErr := svc.GetAgent(context.Background(), testAgentID, false)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorAgentNotFound.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestGetAgent_Success_NoInbound() {
	svc, mockEntity, _, _ := suite.setupService()

	agentEntity := buildAgentEntityFixture(testAgentName, "desc", "alice", "")

	clearMockCalls(mockEntity, "GetEntity")
	mockEntity.On("GetEntity", mock.Anything, testAgentID).Return(agentEntity, nil)

	resp, svcErr := svc.GetAgent(context.Background(), testAgentID, false)
	suite.Require().Nil(svcErr)
	suite.Require().NotNil(resp)
	assert.Equal(suite.T(), testAgentID, resp.ID)
	assert.Equal(suite.T(), testAgentName, resp.Name)
	assert.Equal(suite.T(), "desc", resp.Description)
	assert.Equal(suite.T(), "alice", resp.Owner)
	assert.Nil(suite.T(), resp.InboundAuthConfig)
}

func (suite *AgentServiceTestSuite) TestGetAgent_Success_WithOAuth() {
	svc, mockEntity, mockInbound, _ := suite.setupService()

	agentEntity := buildAgentEntityFixture(testAgentName, "", "", "cid-123")

	clearMockCalls(mockEntity, "GetEntity")
	mockEntity.On("GetEntity", mock.Anything, testAgentID).Return(agentEntity, nil)

	inboundRec := &inboundmodel.InboundClient{ID: testAgentID, AuthFlowID: "flow-1"}
	clearMockCalls(mockInbound, "GetInboundClientByEntityID")
	mockInbound.On("GetInboundClientByEntityID", mock.Anything, testAgentID).Return(inboundRec, nil)

	oauthProfile := &inboundmodel.OAuthProfile{
		AppID: testAgentID,
		OAuthProfile: &inboundmodel.OAuthProfileData{
			GrantTypes: []string{"client_credentials"},
		},
	}
	clearMockCalls(mockInbound, "GetOAuthProfileByEntityID")
	mockInbound.On("GetOAuthProfileByEntityID", mock.Anything, testAgentID).Return(oauthProfile, nil)

	resp, svcErr := svc.GetAgent(context.Background(), testAgentID, false)
	suite.Require().Nil(svcErr)
	suite.Require().NotNil(resp)
	assert.Equal(suite.T(), "flow-1", resp.AuthFlowID)
	suite.Require().Len(resp.InboundAuthConfig, 1)
	// clientSecret must be scrubbed on GET.
	assert.Empty(suite.T(), resp.InboundAuthConfig[0].Config.ClientSecret)
	assert.Equal(suite.T(), "cid-123", resp.InboundAuthConfig[0].Config.ClientID)
}

// --- DeleteAgent ---

func (suite *AgentServiceTestSuite) TestDeleteAgent_EmptyID() {
	svc, _, _, _ := suite.setupService()
	svcErr := svc.DeleteAgent(context.Background(), "")
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorMissingAgentID.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestDeleteAgent_NotFound() {
	svc, _, _, _ := suite.setupService()
	svcErr := svc.DeleteAgent(context.Background(), testAgentID)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorAgentNotFound.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestDeleteAgent_Success_NoInboundClient() {
	svc, mockEntity, mockInbound, _ := suite.setupService()

	agentEntity := buildAgentEntityFixture(testAgentName, "", "", "")
	clearMockCalls(mockEntity, "GetEntity")
	mockEntity.On("GetEntity", mock.Anything, testAgentID).Return(agentEntity, nil)

	clearMockCalls(mockInbound, "DeleteInboundClient")
	mockInbound.On("DeleteInboundClient", mock.Anything, testAgentID).
		Return(inboundclient.ErrInboundClientNotFound)

	clearMockCalls(mockEntity, "DeleteEntity")
	mockEntity.On("DeleteEntity", mock.Anything, testAgentID).Return(nil)

	svcErr := svc.DeleteAgent(context.Background(), testAgentID)
	assert.Nil(suite.T(), svcErr)
	mockEntity.AssertCalled(suite.T(), "DeleteEntity", mock.Anything, testAgentID)
}

func (suite *AgentServiceTestSuite) TestDeleteAgent_Success_WithInboundClient() {
	svc, mockEntity, mockInbound, _ := suite.setupService()

	agentEntity := buildAgentEntityFixture(testAgentName, "", "", "cid-abc")
	clearMockCalls(mockEntity, "GetEntity")
	mockEntity.On("GetEntity", mock.Anything, testAgentID).Return(agentEntity, nil)

	clearMockCalls(mockInbound, "DeleteInboundClient")
	mockInbound.On("DeleteInboundClient", mock.Anything, testAgentID).Return(nil)

	clearMockCalls(mockEntity, "DeleteEntity")
	mockEntity.On("DeleteEntity", mock.Anything, testAgentID).Return(nil)

	svcErr := svc.DeleteAgent(context.Background(), testAgentID)
	assert.Nil(suite.T(), svcErr)
	mockInbound.AssertCalled(suite.T(), "DeleteInboundClient", mock.Anything, testAgentID)
	mockEntity.AssertCalled(suite.T(), "DeleteEntity", mock.Anything, testAgentID)
}

// --- GetAgentList ---

func (suite *AgentServiceTestSuite) TestGetAgentList_InvalidLimit() {
	svc, _, _, _ := suite.setupService()
	resp, svcErr := svc.GetAgentList(context.Background(), -1, 0, nil, false)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorInvalidLimit.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestGetAgentList_Success() {
	svc, mockEntity, _, _ := suite.setupService()

	agentEntity := buildAgentEntityFixture(testAgentName, "desc", "alice", "")
	clearMockCalls(mockEntity, "GetEntityList")
	mockEntity.On("GetEntityList", mock.Anything, entity.EntityCategoryAgent, 30, 0, mock.Anything).
		Return([]entity.Entity{*agentEntity}, nil)
	clearMockCalls(mockEntity, "GetEntityListCount")
	mockEntity.On("GetEntityListCount", mock.Anything, entity.EntityCategoryAgent, mock.Anything).
		Return(1, nil)

	resp, svcErr := svc.GetAgentList(context.Background(), 0, 0, nil, false)
	suite.Require().Nil(svcErr)
	suite.Require().NotNil(resp)
	assert.Equal(suite.T(), 1, resp.TotalResults)
	suite.Require().Len(resp.Agents, 1)
	assert.Equal(suite.T(), testAgentName, resp.Agents[0].Name)
}

// --- GetAgentGroups ---

func (suite *AgentServiceTestSuite) TestGetAgentGroups_EmptyID() {
	svc, _, _, _ := suite.setupService()
	resp, svcErr := svc.GetAgentGroups(context.Background(), "", 10, 0)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorMissingAgentID.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestGetAgentGroups_AgentNotFound() {
	svc, _, _, _ := suite.setupService()
	resp, svcErr := svc.GetAgentGroups(context.Background(), testAgentID, 10, 0)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorAgentNotFound.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestGetAgentGroups_Success() {
	svc, mockEntity, _, _ := suite.setupService()

	agentEntity := buildAgentEntityFixture(testAgentName, "", "", "")
	clearMockCalls(mockEntity, "GetEntity")
	mockEntity.On("GetEntity", mock.Anything, testAgentID).Return(agentEntity, nil)

	clearMockCalls(mockEntity, "GetGroupCountForEntity")
	mockEntity.On("GetGroupCountForEntity", mock.Anything, testAgentID).Return(2, nil)

	clearMockCalls(mockEntity, "GetEntityGroups")
	mockEntity.On("GetEntityGroups", mock.Anything, testAgentID, 10, 0).
		Return([]entity.EntityGroup{
			{ID: "g1", Name: "group-one", OUID: testOUID},
			{ID: "g2", Name: "group-two", OUID: testOUID},
		}, nil)

	resp, svcErr := svc.GetAgentGroups(context.Background(), testAgentID, 10, 0)
	suite.Require().Nil(svcErr)
	suite.Require().NotNil(resp)
	assert.Equal(suite.T(), 2, resp.TotalResults)
	assert.Len(suite.T(), resp.Groups, 2)
}

// --- UpdateAgent ---

func (suite *AgentServiceTestSuite) TestUpdateAgent_EmptyID() {
	svc, _, _, _ := suite.setupService()
	resp, svcErr := svc.UpdateAgent(context.Background(), "", &model.UpdateAgentRequest{
		Name: testAgentName, Type: testAgentType,
	})
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorMissingAgentID.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestUpdateAgent_NilRequest() {
	svc, _, _, _ := suite.setupService()
	resp, svcErr := svc.UpdateAgent(context.Background(), testAgentID, nil)
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorInvalidRequestFormat.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestUpdateAgent_AgentNotFound() {
	svc, _, _, _ := suite.setupService()
	resp, svcErr := svc.UpdateAgent(context.Background(), testAgentID, &model.UpdateAgentRequest{
		Name: testAgentName, Type: testAgentType,
	})
	assert.Nil(suite.T(), resp)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorAgentNotFound.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestUpdateAgent_Success_EntityOnly() {
	svc, mockEntity, _, _ := suite.setupService()

	agentEntity := buildAgentEntityFixture("old-name", "", "", "")
	clearMockCalls(mockEntity, "GetEntity")
	mockEntity.On("GetEntity", mock.Anything, testAgentID).Return(agentEntity, nil)

	clearMockCalls(mockEntity, "UpdateEntity")
	mockEntity.On("UpdateEntity", mock.Anything, testAgentID, mock.Anything).
		Return(&entity.Entity{}, nil)

	resp, svcErr := svc.UpdateAgent(context.Background(), testAgentID, &model.UpdateAgentRequest{
		Name: testAgentName, Type: testAgentType,
	})
	suite.Require().Nil(svcErr)
	suite.Require().NotNil(resp)
	assert.Equal(suite.T(), testAgentName, resp.Name)
}

func (suite *AgentServiceTestSuite) TestUpdateAgent_FlowIDResolvedToDefault() {
	svc, mockEntity, mockInbound, _ := suite.setupService()

	agentEntity := buildAgentEntityFixture("old-name", "", "", "")
	clearMockCalls(mockEntity, "GetEntity")
	mockEntity.On("GetEntity", mock.Anything, testAgentID).Return(agentEntity, nil)

	clearMockCalls(mockEntity, "UpdateEntity")
	mockEntity.On("UpdateEntity", mock.Anything, testAgentID, mock.Anything).
		Return(&entity.Entity{}, nil)

	clearMockCalls(mockInbound, "GetInboundClientByEntityID")
	mockInbound.On("GetInboundClientByEntityID", mock.Anything, mock.Anything).
		Return(&inboundmodel.InboundClient{}, nil)

	clearMockCalls(mockInbound, "UpdateInboundClient")
	mockInbound.On("UpdateInboundClient", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			client := args.Get(1).(*inboundmodel.InboundClient)
			client.AuthFlowID = "default-flow-id"
			client.RegistrationFlowID = "default-reg-flow-id"
		}).Return(nil)

	resp, svcErr := svc.UpdateAgent(context.Background(), testAgentID, &model.UpdateAgentRequest{
		Name: testAgentName,
		Type: testAgentType,
		InboundAuthConfig: []model.InboundAuthConfig{
			{
				Type: model.OAuthInboundAuthType,
				Config: &model.OAuthAgentConfig{
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	})
	suite.Require().Nil(svcErr)
	suite.Require().NotNil(resp)
	assert.Equal(suite.T(), "default-flow-id", resp.AuthFlowID)
	assert.Equal(suite.T(), "default-reg-flow-id", resp.RegistrationFlowID)
}

// --- mapEntityError ---

func (suite *AgentServiceTestSuite) TestMapEntityError_NotFound() {
	svcErr := mapEntityError(entity.ErrEntityNotFound)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorAgentNotFound.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestMapEntityError_SchemaValidation() {
	svcErr := mapEntityError(entity.ErrSchemaValidationFailed)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorSchemaValidationFailed.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestMapEntityError_Unknown() {
	svcErr := mapEntityError(entity.ErrAmbiguousEntity)
	assert.Nil(suite.T(), svcErr)
}

// --- translateInboundClientError ---

func (suite *AgentServiceTestSuite) TestTranslateInboundClientError_InvalidRedirectURI() {
	svc, _, _, _ := suite.setupService()
	svcErr := svc.translateInboundClientError(inboundclient.ErrOAuthInvalidRedirectURI)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorInvalidRedirectURI.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestTranslateInboundClientError_InvalidGrantType() {
	svc, _, _, _ := suite.setupService()
	svcErr := svc.translateInboundClientError(inboundclient.ErrOAuthInvalidGrantType)
	suite.Require().NotNil(svcErr)
	assert.Equal(suite.T(), ErrorInvalidGrantType.Code, svcErr.Code)
}

func (suite *AgentServiceTestSuite) TestTranslateInboundClientError_Unknown() {
	svc, _, _, _ := suite.setupService()
	svcErr := svc.translateInboundClientError(inboundclient.ErrInboundClientNotFound)
	assert.Nil(suite.T(), svcErr)
}

// --- helpers ---

// clearMockCalls removes all expectations for the named method from the mock, so a test
// can register a more specific expectation without conflicting with the permissive default.
func clearMockCalls(m any, method string) {
	var mockObj *mock.Mock
	switch v := m.(type) {
	case *entitymock.EntityServiceInterfaceMock:
		mockObj = &v.Mock
	case *inboundclientmock.InboundClientServiceInterfaceMock:
		mockObj = &v.Mock
	case *oumock.OrganizationUnitServiceInterfaceMock:
		mockObj = &v.Mock
	}
	if mockObj == nil {
		return
	}
	var kept []*mock.Call
	for _, c := range mockObj.ExpectedCalls {
		if c.Method != method {
			kept = append(kept, c)
		}
	}
	mockObj.ExpectedCalls = kept
}
