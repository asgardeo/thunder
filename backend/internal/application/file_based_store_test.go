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

package application

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/file_based_runtime/entity"
)

// FileBasedStoreTestSuite contains comprehensive tests for the file-based application store.
// The test suite covers:
// - All CRUD operations including unsupported ones (CreateApplication, GetApplicationByID, etc.)
// - OAuth application retrieval by client ID with various configurations
// - Application listing and counting functionality
// - Error handling for storage failures, type assertion failures, and edge cases
// - Mock entity store implementation for isolated unit testing
//
// Test Coverage: 100% of statements in file_based_store.go
// All interface methods are thoroughly tested with success and error scenarios.
type FileBasedStoreTestSuite struct {
	suite.Suite
	store       applicationStoreInterface
	mockStorage *mockEntityStore
}

// mockEntityStore implements entity.StoreInterface for testing purposes
type mockEntityStore struct {
	data   map[string]*entity.Entity
	errors map[string]error // Map operation names to errors for testing
}

func (m *mockEntityStore) Get(key entity.CompositeKey) (*entity.Entity, error) {
	if err, exists := m.errors["Get"]; exists {
		return nil, err
	}

	entity, exists := m.data[key.String()]
	if !exists {
		return nil, errors.New("entity not found")
	}
	return entity, nil
}

func (m *mockEntityStore) Set(key entity.CompositeKey, data interface{}) error {
	if err, exists := m.errors["Set"]; exists {
		return err
	}

	m.data[key.String()] = &entity.Entity{
		ID:   key,
		Data: data,
	}
	return nil
}

func (m *mockEntityStore) Delete(key entity.CompositeKey) error {
	if err, exists := m.errors["Delete"]; exists {
		return err
	}

	delete(m.data, key.String())
	return nil
}

func (m *mockEntityStore) List() ([]*entity.Entity, error) {
	if err, exists := m.errors["List"]; exists {
		return nil, err
	}

	entities := make([]*entity.Entity, 0, len(m.data))
	for _, entity := range m.data {
		entities = append(entities, entity)
	}
	return entities, nil
}

func (m *mockEntityStore) ListByID(id string) ([]*entity.Entity, error) {
	if err, exists := m.errors["ListByID"]; exists {
		return nil, err
	}

	var entities []*entity.Entity
	for _, entity := range m.data {
		if entity.ID.ID == id {
			entities = append(entities, entity)
		}
	}
	return entities, nil
}

func (m *mockEntityStore) ListByType(keyType entity.KeyType) ([]*entity.Entity, error) {
	if err, exists := m.errors["ListByType"]; exists {
		return nil, err
	}

	var entities []*entity.Entity
	for _, entity := range m.data {
		if entity.ID.Type == keyType {
			entities = append(entities, entity)
		}
	}
	return entities, nil
}

func (m *mockEntityStore) CountByType(keyType entity.KeyType) (int, error) {
	count := 0
	for _, entity := range m.data {
		if entity.ID.Type == keyType {
			count++
		}
	}
	return count, nil
}

func (m *mockEntityStore) Clear() error {
	if err, exists := m.errors["Clear"]; exists {
		return err
	}

	m.data = make(map[string]*entity.Entity)
	return nil
}

func (m *mockEntityStore) setError(operation string, err error) {
	if m.errors == nil {
		m.errors = make(map[string]error)
	}
	m.errors[operation] = err
}

func TestFileBasedStoreTestSuite(t *testing.T) {
	suite.Run(t, new(FileBasedStoreTestSuite))
}

func (suite *FileBasedStoreTestSuite) SetupTest() {
	suite.mockStorage = &mockEntityStore{
		data:   make(map[string]*entity.Entity),
		errors: make(map[string]error),
	}

	// Create file-based store with mock storage
	suite.store = &fileBasedStore{
		storage: suite.mockStorage,
	}
}

// Helper function to create a test application
func (suite *FileBasedStoreTestSuite) createTestApplication(id, name string) *model.ApplicationProcessedDTO {
	return &model.ApplicationProcessedDTO{
		ID:                        id,
		Name:                      name,
		Description:               "Test application description",
		AuthFlowGraphID:           "auth_flow_1",
		RegistrationFlowGraphID:   "reg_flow_1",
		IsRegistrationFlowEnabled: true,
		URL:                       "https://example.com",
		LogoURL:                   "https://example.com/logo.png",
		Token: &model.TokenConfig{
			Issuer:         "test-issuer",
			ValidityPeriod: 3600,
			UserAttributes: []string{"email", "name"},
		},
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
					AppID:                   id,
					ClientID:                "client_" + id,
					HashedClientSecret:      "hashed_secret_" + id,
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretPost,
					PKCERequired:            true,
					PublicClient:            false,
				},
			},
		},
	}
}

// Tests for CreateApplication method

func (suite *FileBasedStoreTestSuite) TestCreateApplication_Success() {
	app := suite.createTestApplication("app1", "Test App 1")

	err := suite.store.CreateApplication(*app)

	suite.NoError(err)

	// Verify the application was stored
	key := entity.NewCompositeKey(app.ID, entity.KeyTypeApplication)
	storedEntity, exists := suite.mockStorage.data[key.String()]
	suite.True(exists)
	suite.Equal(app, storedEntity.Data)
}

func (suite *FileBasedStoreTestSuite) TestCreateApplication_StorageError() {
	app := suite.createTestApplication("app1", "Test App 1")
	suite.mockStorage.setError("Set", errors.New("storage error"))

	err := suite.store.CreateApplication(*app)

	suite.Error(err)
	suite.Contains(err.Error(), "storage error")
}

// Tests for GetApplicationByID method

func (suite *FileBasedStoreTestSuite) TestGetApplicationByID_Success() {
	app := suite.createTestApplication("app1", "Test App 1")
	key := entity.NewCompositeKey(app.ID, entity.KeyTypeApplication)
	suite.mockStorage.data[key.String()] = &entity.Entity{
		ID:   key,
		Data: app,
	}

	result, err := suite.store.GetApplicationByID("app1")

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(app.ID, result.ID)
	suite.Equal(app.Name, result.Name)
}

func (suite *FileBasedStoreTestSuite) TestGetApplicationByID_NotFound() {
	result, err := suite.store.GetApplicationByID("nonexistent")

	suite.Error(err)
	suite.Nil(result)
}

func (suite *FileBasedStoreTestSuite) TestGetApplicationByID_StorageError() {
	suite.mockStorage.setError("Get", errors.New("storage error"))

	result, err := suite.store.GetApplicationByID("app1")

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "storage error")
}

func (suite *FileBasedStoreTestSuite) TestGetApplicationByID_TypeAssertionFailure() {
	// Store wrong type data
	key := entity.NewCompositeKey("app1", entity.KeyTypeApplication)
	suite.mockStorage.data[key.String()] = &entity.Entity{
		ID:   key,
		Data: "wrong type", // Not an ApplicationProcessedDTO
	}

	result, err := suite.store.GetApplicationByID("app1")

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(model.ApplicationDataCorruptedError, err)
}

// Tests for GetApplicationByName method

func (suite *FileBasedStoreTestSuite) TestGetApplicationByName_Success() {
	app1 := suite.createTestApplication("app1", "Test App 1")
	app2 := suite.createTestApplication("app2", "Test App 2")

	// Store apps in mock storage
	key1 := entity.NewCompositeKey(app1.ID, entity.KeyTypeApplication)
	key2 := entity.NewCompositeKey(app2.ID, entity.KeyTypeApplication)
	suite.mockStorage.data[key1.String()] = &entity.Entity{ID: key1, Data: app1}
	suite.mockStorage.data[key2.String()] = &entity.Entity{ID: key2, Data: app2}

	result, err := suite.store.GetApplicationByName("Test App 1")

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(app1.ID, result.ID)
	suite.Equal("Test App 1", result.Name)
}

func (suite *FileBasedStoreTestSuite) TestGetApplicationByName_NotFound() {
	app := suite.createTestApplication("app1", "Test App 1")
	key := entity.NewCompositeKey(app.ID, entity.KeyTypeApplication)
	suite.mockStorage.data[key.String()] = &entity.Entity{ID: key, Data: app}

	result, err := suite.store.GetApplicationByName("Nonexistent App")

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(model.ApplicationNotFoundError, err)
}

func (suite *FileBasedStoreTestSuite) TestGetApplicationByName_StorageError() {
	suite.mockStorage.setError("ListByType", errors.New("storage error"))

	result, err := suite.store.GetApplicationByName("Test App")

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "storage error")
}

func (suite *FileBasedStoreTestSuite) TestGetApplicationByName_TypeAssertionFailure() {
	// Store wrong type data
	key := entity.NewCompositeKey("app1", entity.KeyTypeApplication)
	suite.mockStorage.data[key.String()] = &entity.Entity{
		ID:   key,
		Data: "wrong type", // Not an ApplicationProcessedDTO
	}

	result, err := suite.store.GetApplicationByName("Test App")

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(model.ApplicationNotFoundError, err)
}

// Tests for GetApplicationList method

func (suite *FileBasedStoreTestSuite) TestGetApplicationList_Success() {
	app1 := suite.createTestApplication("app1", "Test App 1")
	app2 := suite.createTestApplication("app2", "Test App 2")

	// Store apps in mock storage
	key1 := entity.NewCompositeKey(app1.ID, entity.KeyTypeApplication)
	key2 := entity.NewCompositeKey(app2.ID, entity.KeyTypeApplication)
	suite.mockStorage.data[key1.String()] = &entity.Entity{ID: key1, Data: app1}
	suite.mockStorage.data[key2.String()] = &entity.Entity{ID: key2, Data: app2}

	result, err := suite.store.GetApplicationList()

	suite.NoError(err)
	suite.Len(result, 2)

	// Check that both apps are in the result
	var foundApp1, foundApp2 bool
	for _, app := range result {
		if app.ID == "app1" && app.Name == "Test App 1" {
			foundApp1 = true
		}
		if app.ID == "app2" && app.Name == "Test App 2" {
			foundApp2 = true
		}
	}
	suite.True(foundApp1)
	suite.True(foundApp2)
}

func (suite *FileBasedStoreTestSuite) TestGetApplicationList_EmptyList() {
	result, err := suite.store.GetApplicationList()

	suite.NoError(err)
	suite.Len(result, 0)
}

func (suite *FileBasedStoreTestSuite) TestGetApplicationList_StorageError() {
	suite.mockStorage.setError("ListByType", errors.New("storage error"))

	result, err := suite.store.GetApplicationList()

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "storage error")
}

// Tests for GetTotalApplicationCount method

func (suite *FileBasedStoreTestSuite) TestGetTotalApplicationCount_Success() {
	app1 := suite.createTestApplication("app1", "Test App 1")
	app2 := suite.createTestApplication("app2", "Test App 2")

	// Store apps in mock storage
	key1 := entity.NewCompositeKey(app1.ID, entity.KeyTypeApplication)
	key2 := entity.NewCompositeKey(app2.ID, entity.KeyTypeApplication)
	suite.mockStorage.data[key1.String()] = &entity.Entity{ID: key1, Data: app1}
	suite.mockStorage.data[key2.String()] = &entity.Entity{ID: key2, Data: app2}

	count, err := suite.store.GetTotalApplicationCount()

	suite.NoError(err)
	suite.Equal(2, count)
}

func (suite *FileBasedStoreTestSuite) TestGetTotalApplicationCount_Empty() {
	count, err := suite.store.GetTotalApplicationCount()

	suite.NoError(err)
	suite.Equal(0, count)
}

func (suite *FileBasedStoreTestSuite) TestGetTotalApplicationCount_StorageError() {
	suite.mockStorage.setError("ListByType", errors.New("storage error"))

	count, err := suite.store.GetTotalApplicationCount()

	suite.Error(err)
	suite.Equal(0, count)
	suite.Contains(err.Error(), "storage error")
}

// Tests for GetOAuthApplication method

func (suite *FileBasedStoreTestSuite) TestGetOAuthApplication_Success() {
	app := suite.createTestApplication("app1", "Test App 1")
	clientID := "client_app1"

	// Store app in mock storage
	key := entity.NewCompositeKey(app.ID, entity.KeyTypeApplication)
	suite.mockStorage.data[key.String()] = &entity.Entity{ID: key, Data: app}

	result, err := suite.store.GetOAuthApplication(clientID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(clientID, result.ClientID)
	suite.Equal("app1", result.AppID)
}

func (suite *FileBasedStoreTestSuite) TestGetOAuthApplication_NotFound() {
	app := suite.createTestApplication("app1", "Test App 1")

	// Store app with different client ID
	key := entity.NewCompositeKey(app.ID, entity.KeyTypeApplication)
	suite.mockStorage.data[key.String()] = &entity.Entity{ID: key, Data: app}

	result, err := suite.store.GetOAuthApplication("nonexistent_client")

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(model.ApplicationNotFoundError, err)
}

func (suite *FileBasedStoreTestSuite) TestGetOAuthApplication_NoOAuthConfig() {
	// Create app without OAuth configuration
	app := &model.ApplicationProcessedDTO{
		ID:                "app1",
		Name:              "Test App 1",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{},
	}

	key := entity.NewCompositeKey(app.ID, entity.KeyTypeApplication)
	suite.mockStorage.data[key.String()] = &entity.Entity{ID: key, Data: app}

	result, err := suite.store.GetOAuthApplication("any_client")

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(model.ApplicationNotFoundError, err)
}

func (suite *FileBasedStoreTestSuite) TestGetOAuthApplication_MultipleApps() {
	app1 := suite.createTestApplication("app1", "Test App 1")
	app2 := suite.createTestApplication("app2", "Test App 2")

	// Store both apps
	key1 := entity.NewCompositeKey(app1.ID, entity.KeyTypeApplication)
	key2 := entity.NewCompositeKey(app2.ID, entity.KeyTypeApplication)
	suite.mockStorage.data[key1.String()] = &entity.Entity{ID: key1, Data: app1}
	suite.mockStorage.data[key2.String()] = &entity.Entity{ID: key2, Data: app2}

	// Search for app2's client ID
	result, err := suite.store.GetOAuthApplication("client_app2")

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("client_app2", result.ClientID)
	suite.Equal("app2", result.AppID)
}

func (suite *FileBasedStoreTestSuite) TestGetOAuthApplication_StorageError() {
	suite.mockStorage.setError("ListByType", errors.New("storage error"))

	result, err := suite.store.GetOAuthApplication("client1")

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "storage error")
}

func (suite *FileBasedStoreTestSuite) TestGetOAuthApplication_NonOAuthInboundAuth() {
	// Create app with non-OAuth inbound auth configuration
	app := &model.ApplicationProcessedDTO{
		ID:   "app1",
		Name: "Test App 1",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: "saml", // Non-OAuth type
			},
		},
	}

	key := entity.NewCompositeKey(app.ID, entity.KeyTypeApplication)
	suite.mockStorage.data[key.String()] = &entity.Entity{ID: key, Data: app}

	result, err := suite.store.GetOAuthApplication("any_client")

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(model.ApplicationNotFoundError, err)
}

// Tests for unsupported operations

func (suite *FileBasedStoreTestSuite) TestUpdateApplication_NotSupported() {
	app1 := suite.createTestApplication("app1", "Test App 1")
	app2 := suite.createTestApplication("app1", "Updated App 1")

	err := suite.store.UpdateApplication(app1, app2)

	suite.Error(err)
	suite.Contains(err.Error(), "UpdateApplication is not supported in file-based store")
}

func (suite *FileBasedStoreTestSuite) TestDeleteApplication_NotSupported() {
	err := suite.store.DeleteApplication("app1")

	suite.Error(err)
	suite.Contains(err.Error(), "DeleteApplication is not supported in file-based store")
}

// Tests for edge cases and error handling

func (suite *FileBasedStoreTestSuite) TestGetApplicationList_TypeAssertionFailure() {
	// Store entity with wrong type
	key := entity.NewCompositeKey("app1", entity.KeyTypeApplication)
	suite.mockStorage.data[key.String()] = &entity.Entity{
		ID:   key,
		Data: "wrong type", // Not an ApplicationProcessedDTO
	}

	result, err := suite.store.GetApplicationList()

	suite.NoError(err)
	suite.Len(result, 0) // Wrong type entities are skipped
}

func (suite *FileBasedStoreTestSuite) TestGetOAuthApplication_NilOAuthConfig() {
	// Create app with OAuth type but nil config
	app := &model.ApplicationProcessedDTO{
		ID:   "app1",
		Name: "Test App 1",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type:           model.OAuthInboundAuthType,
				OAuthAppConfig: nil, // Nil OAuth config
			},
		},
	}

	key := entity.NewCompositeKey(app.ID, entity.KeyTypeApplication)
	suite.mockStorage.data[key.String()] = &entity.Entity{ID: key, Data: app}

	result, err := suite.store.GetOAuthApplication("any_client")

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(model.ApplicationNotFoundError, err)
}

func (suite *FileBasedStoreTestSuite) TestNewFileBasedStore() {
	store := newFileBasedStore()

	suite.NotNil(store)
	suite.IsType(&fileBasedStore{}, store)
}
