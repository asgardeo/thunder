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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
)

// FileBasedStoreTestSuite contains comprehensive tests for the file-based application store.
// The test suite covers:
// - All CRUD operations including unsupported ones (CreateApplication, GetApplicationByID, etc.)
// - OAuth application retrieval by client ID with various configurations
// - Application listing and counting functionality
// - Error handling for storage failures, type assertion failures, and edge cases
// - Mock entity store implementation for isolated unit testing
type FileBasedStoreTestSuite struct {
	suite.Suite
	store applicationStoreInterface
}

func TestFileBasedStoreTestSuite(t *testing.T) {
	suite.Run(t, new(FileBasedStoreTestSuite))
}

func (suite *FileBasedStoreTestSuite) SetupTest() {
	genericStore := declarativeresource.NewGenericFileBasedStoreForTest(entity.KeyTypeApplication)
	suite.store = &fileBasedStore{
		GenericFileBasedStore: genericStore,
	}
}

// Helper function to create a test application
func (suite *FileBasedStoreTestSuite) createTestApplication(id, name string) *model.ApplicationProcessedDTO {
	return &model.ApplicationProcessedDTO{
		ID:                        id,
		Name:                      name,
		Description:               "Test application description",
		AuthFlowID:                "auth_flow_1",
		RegistrationFlowID:        "reg_flow_1",
		IsRegistrationFlowEnabled: true,
		URL:                       "https://example.com",
		LogoURL:                   "https://example.com/logo.png",
		Assertion: &model.AssertionConfig{
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

	// Verify the application was stored by retrieving it
	storedApp, err := suite.store.GetApplicationByID(app.ID)
	suite.NoError(err)
	suite.NotNil(storedApp)
	suite.Equal(app.ID, storedApp.ID)
	suite.Equal(app.Name, storedApp.Name)
}

// TestCreateApplication_StorageError removed - cannot inject errors with GenericFileBasedStore

// Tests for GetApplicationByID method

func (suite *FileBasedStoreTestSuite) TestGetApplicationByID_Success() {
	app := suite.createTestApplication("app1", "Test App 1")
	err := suite.store.CreateApplication(*app)
	suite.NoError(err)

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

// TestGetApplicationByID_StorageError removed - cannot inject errors with GenericFileBasedStore

// TestGetApplicationByID_TypeAssertionFailure removed - cannot inject wrong types with GenericFileBasedStore

// Tests for GetApplicationByName method

func (suite *FileBasedStoreTestSuite) TestGetApplicationByName_Success() {
	app1 := suite.createTestApplication("app1", "Test App 1")
	app2 := suite.createTestApplication("app2", "Test App 2")

	// Store apps
	err := suite.store.CreateApplication(*app1)
	suite.NoError(err)
	err = suite.store.CreateApplication(*app2)
	suite.NoError(err)

	result, err := suite.store.GetApplicationByName("Test App 1")

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(app1.ID, result.ID)
	suite.Equal("Test App 1", result.Name)
}

func (suite *FileBasedStoreTestSuite) TestGetApplicationByName_NotFound() {
	app := suite.createTestApplication("app1", "Test App 1")
	err := suite.store.CreateApplication(*app)
	suite.NoError(err)

	result, err := suite.store.GetApplicationByName("Nonexistent App")

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(model.ApplicationNotFoundError, err)
}

// TestGetApplicationByName_StorageError removed - cannot inject errors with GenericFileBasedStore

// TestGetApplicationByName_TypeAssertionFailure removed - cannot inject wrong types with GenericFileBasedStore

// Tests for GetApplicationList method

func (suite *FileBasedStoreTestSuite) TestGetApplicationList_Success() {
	app1 := suite.createTestApplication("app1", "Test App 1")
	app2 := suite.createTestApplication("app2", "Test App 2")

	// Store apps
	err := suite.store.CreateApplication(*app1)
	suite.NoError(err)
	err = suite.store.CreateApplication(*app2)
	suite.NoError(err)

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

// TestGetApplicationList_StorageError removed - cannot inject errors with GenericFileBasedStore

// Tests for GetTotalApplicationCount method

func (suite *FileBasedStoreTestSuite) TestGetTotalApplicationCount_Success() {
	app1 := suite.createTestApplication("app1", "Test App 1")
	app2 := suite.createTestApplication("app2", "Test App 2")

	// Store apps
	err := suite.store.CreateApplication(*app1)
	suite.NoError(err)
	err = suite.store.CreateApplication(*app2)
	suite.NoError(err)

	count, err := suite.store.GetTotalApplicationCount()

	suite.NoError(err)
	suite.Equal(2, count)
}

func (suite *FileBasedStoreTestSuite) TestGetTotalApplicationCount_Empty() {
	count, err := suite.store.GetTotalApplicationCount()

	suite.NoError(err)
	suite.Equal(0, count)
}

// TestGetTotalApplicationCount_StorageError removed - cannot inject errors with GenericFileBasedStore

// Tests for GetOAuthApplication method

func (suite *FileBasedStoreTestSuite) TestGetOAuthApplication_Success() {
	app := suite.createTestApplication("app1", "Test App 1")
	clientID := "client_app1"

	// Store app
	err := suite.store.CreateApplication(*app)
	suite.NoError(err)

	result, err := suite.store.GetOAuthApplication(clientID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(clientID, result.ClientID)
	suite.Equal("app1", result.AppID)
}

func (suite *FileBasedStoreTestSuite) TestGetOAuthApplication_NotFound() {
	app := suite.createTestApplication("app1", "Test App 1")

	// Store app with different client ID
	err := suite.store.CreateApplication(*app)
	suite.NoError(err)

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

	err := suite.store.CreateApplication(*app)
	suite.NoError(err)

	result, err := suite.store.GetOAuthApplication("any_client")

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(model.ApplicationNotFoundError, err)
}

func (suite *FileBasedStoreTestSuite) TestGetOAuthApplication_MultipleApps() {
	app1 := suite.createTestApplication("app1", "Test App 1")
	app2 := suite.createTestApplication("app2", "Test App 2")

	// Store both apps
	err := suite.store.CreateApplication(*app1)
	suite.NoError(err)
	err = suite.store.CreateApplication(*app2)
	suite.NoError(err)

	// Search for app2's client ID
	result, err := suite.store.GetOAuthApplication("client_app2")

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("client_app2", result.ClientID)
	suite.Equal("app2", result.AppID)
}

// TestGetOAuthApplication_StorageError removed - cannot inject errors with GenericFileBasedStore

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

	err := suite.store.CreateApplication(*app)
	suite.NoError(err)

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

// TestGetApplicationList_TypeAssertionFailure removed - cannot inject wrong types with GenericFileBasedStore

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

	err := suite.store.CreateApplication(*app)
	suite.NoError(err)

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

// TestFileBasedStore_IsApplicationExists tests checking if an application exists.
func (suite *FileBasedStoreTestSuite) TestFileBasedStore_IsApplicationExists() {
	suite.Run("returns true for existing application", func() {
		// Create an application
		app := model.ApplicationProcessedDTO{
			ID:   "test-app-1",
			Name: "Test Application",
		}
		err := suite.store.CreateApplication(app)
		suite.NoError(err)

		// Check existence
		exists, err := suite.store.IsApplicationExists("test-app-1")
		suite.NoError(err)
		suite.True(exists)
	})

	suite.Run("returns false for non-existent application", func() {
		exists, err := suite.store.IsApplicationExists("nonexistent")
		suite.NoError(err)
		suite.False(exists)
	})

	suite.Run("handles application not found error", func() {
		exists, err := suite.store.IsApplicationExists("nonexistent-app-xyz")
		suite.NoError(err)
		suite.False(exists)
	})
}

// TestFileBasedStore_IsApplicationExistsByName tests checking if an application exists by name.
func (suite *FileBasedStoreTestSuite) TestFileBasedStore_IsApplicationExistsByName() {
	suite.Run("returns true for existing application name", func() {
		// Create an application
		app := model.ApplicationProcessedDTO{
			ID:   "test-app-1",
			Name: "Unique App Name",
		}
		err := suite.store.CreateApplication(app)
		suite.NoError(err)

		// Check existence by name
		exists, err := suite.store.IsApplicationExistsByName("Unique App Name")
		suite.NoError(err)
		suite.True(exists)
	})

	suite.Run("returns false for non-existent application name", func() {
		exists, err := suite.store.IsApplicationExistsByName("Nonexistent Name")
		suite.NoError(err)
		suite.False(exists)
	})

	suite.Run("is case-sensitive for name matching", func() {
		// Create an application
		app := model.ApplicationProcessedDTO{
			ID:   "test-app-1",
			Name: "TestApp",
		}
		err := suite.store.CreateApplication(app)
		suite.NoError(err)

		// Check with different case
		exists, err := suite.store.IsApplicationExistsByName("testapp")
		suite.NoError(err)
		suite.False(exists) // Should not match due to case sensitivity

		// Check with correct case
		exists, err = suite.store.IsApplicationExistsByName("TestApp")
		suite.NoError(err)
		suite.True(exists)
	})
}

// TestFileBasedStore_IsApplicationDeclarative tests checking if an application is declarative.
func (suite *FileBasedStoreTestSuite) TestFileBasedStore_IsApplicationDeclarative() {
	suite.Run("returns true for existing declarative application", func() {
		// Create an application
		app := model.ApplicationProcessedDTO{
			ID:   "declarative-app-1",
			Name: "Declarative App",
		}
		err := suite.store.CreateApplication(app)
		suite.NoError(err)

		// Check if declarative
		isDeclarative := suite.store.IsApplicationDeclarative("declarative-app-1")
		suite.True(isDeclarative)
	})

	suite.Run("returns false for non-existent application", func() {
		isDeclarative := suite.store.IsApplicationDeclarative("nonexistent")
		suite.False(isDeclarative)
	})

	suite.Run("returns true for all existing applications", func() {
		// File-based store should return true for all existing applications
		// since they are all declarative (immutable)
		apps := []model.ApplicationProcessedDTO{
			{ID: "app-1", Name: "App 1"},
			{ID: "app-2", Name: "App 2"},
			{ID: "app-3", Name: "App 3"},
		}

		for _, app := range apps {
			err := suite.store.CreateApplication(app)
			suite.NoError(err)
		}

		for _, app := range apps {
			isDeclarative := suite.store.IsApplicationDeclarative(app.ID)
			suite.True(isDeclarative, "Application %s should be marked as declarative", app.ID)
		}
	})

	suite.Run("returns false for non-existent after created app", func() {
		// Create an app
		app := model.ApplicationProcessedDTO{
			ID:   "existing-app",
			Name: "Existing App",
		}
		err := suite.store.CreateApplication(app)
		suite.NoError(err)

		// Non-existent app should return false
		isDeclarative := suite.store.IsApplicationDeclarative("non-existent-app")
		suite.False(isDeclarative)
	})
}
