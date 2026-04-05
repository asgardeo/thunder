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
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
)

// FileBasedStoreTestSuite contains comprehensive tests for the file-based application store.
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

func (suite *FileBasedStoreTestSuite) createTestApp(id string) applicationConfigDAO {
	return applicationConfigDAO{
		ID:                        id,
		AuthFlowID:                "auth-flow-1",
		RegistrationFlowID:        "reg-flow-1",
		IsRegistrationFlowEnabled: true,
	}
}

// Tests for CreateApplication method

func (suite *FileBasedStoreTestSuite) TestCreateApplication_Success() {
	app := suite.createTestApp("app1")

	err := suite.store.CreateApplication(context.Background(), app)
	suite.NoError(err)

	// Verify the application was stored by retrieving it
	storedApp, err := suite.store.GetApplicationByID(context.Background(), app.ID)
	suite.NoError(err)
	suite.NotNil(storedApp)
	suite.Equal(app.ID, storedApp.ID)
}

// Tests for GetApplicationByID method

func (suite *FileBasedStoreTestSuite) TestGetApplicationByID_Success() {
	app := suite.createTestApp("app1")
	err := suite.store.CreateApplication(context.Background(), app)
	suite.NoError(err)

	result, err := suite.store.GetApplicationByID(context.Background(), "app1")

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(app.ID, result.ID)
}

func (suite *FileBasedStoreTestSuite) TestGetApplicationByID_NotFound() {
	result, err := suite.store.GetApplicationByID(context.Background(), "nonexistent")

	suite.Error(err)
	suite.Nil(result)
}

// Tests for GetOAuthConfigByEntityID method

func (suite *FileBasedStoreTestSuite) TestGetOAuthConfigByEntityID_AlwaysNotFound() {
	// File-based store does not support OAuth config by entity ID
	result, err := suite.store.GetOAuthConfigByEntityID(context.Background(), "any-entity")

	suite.Error(err)
	suite.Nil(result)
	suite.Equal(model.ApplicationNotFoundError, err)
}

// Tests for GetApplicationList method

func (suite *FileBasedStoreTestSuite) TestGetApplicationList_Success() {
	app1 := suite.createTestApp("app1")
	app2 := suite.createTestApp("app2")

	err := suite.store.CreateApplication(context.Background(), app1)
	suite.NoError(err)
	err = suite.store.CreateApplication(context.Background(), app2)
	suite.NoError(err)

	result, err := suite.store.GetApplicationList(context.Background())

	suite.NoError(err)
	suite.Len(result, 2)

	var foundApp1, foundApp2 bool
	for _, app := range result {
		if app.ID == "app1" {
			foundApp1 = true
			suite.True(app.IsReadOnly)
		}
		if app.ID == "app2" {
			foundApp2 = true
			suite.True(app.IsReadOnly)
		}
	}
	suite.True(foundApp1)
	suite.True(foundApp2)
}

func (suite *FileBasedStoreTestSuite) TestGetApplicationList_EmptyList() {
	result, err := suite.store.GetApplicationList(context.Background())

	suite.NoError(err)
	suite.Len(result, 0)
}

// Tests for GetTotalApplicationCount method

func (suite *FileBasedStoreTestSuite) TestGetTotalApplicationCount_Success() {
	app1 := suite.createTestApp("app1")
	app2 := suite.createTestApp("app2")

	err := suite.store.CreateApplication(context.Background(), app1)
	suite.NoError(err)
	err = suite.store.CreateApplication(context.Background(), app2)
	suite.NoError(err)

	count, err := suite.store.GetTotalApplicationCount(context.Background())

	suite.NoError(err)
	suite.Equal(2, count)
}

func (suite *FileBasedStoreTestSuite) TestGetTotalApplicationCount_Empty() {
	count, err := suite.store.GetTotalApplicationCount(context.Background())

	suite.NoError(err)
	suite.Equal(0, count)
}

// Tests for unsupported operations

func (suite *FileBasedStoreTestSuite) TestUpdateApplication_NotSupported() {
	app := suite.createTestApp("app1")

	err := suite.store.UpdateApplication(context.Background(), app)

	suite.Error(err)
	suite.Contains(err.Error(), "UpdateApplication is not supported in file-based store")
}

func (suite *FileBasedStoreTestSuite) TestUpdateOAuthConfig_NotSupported() {
	rawJSON := json.RawMessage(`{}`)
	err := suite.store.UpdateOAuthConfig(context.Background(), "entity-1", rawJSON)

	suite.Error(err)
	suite.Contains(err.Error(), "UpdateOAuthConfig is not supported in file-based store")
}

func (suite *FileBasedStoreTestSuite) TestDeleteApplication_NotSupported() {
	err := suite.store.DeleteApplication(context.Background(), "app1")

	suite.Error(err)
	suite.Contains(err.Error(), "DeleteApplication is not supported in file-based store")
}

func (suite *FileBasedStoreTestSuite) TestDeleteOAuthConfig_NotSupported() {
	err := suite.store.DeleteOAuthConfig(context.Background(), "entity-1")

	suite.Error(err)
	suite.Contains(err.Error(), "DeleteOAuthConfig is not supported in file-based store")
}

func (suite *FileBasedStoreTestSuite) TestCreateOAuthConfig_NotSupported() {
	rawJSON := json.RawMessage(`{}`)
	err := suite.store.CreateOAuthConfig(context.Background(), "entity-1", rawJSON)

	suite.Error(err)
	suite.Contains(err.Error(), "CreateOAuthConfig is not supported in file-based store")
}

// Tests for IsApplicationExists

func (suite *FileBasedStoreTestSuite) TestIsApplicationExists_ExistingApp() {
	app := applicationConfigDAO{ID: "test-app-1"}
	err := suite.store.CreateApplication(context.Background(), app)
	suite.NoError(err)

	exists, err := suite.store.IsApplicationExists(context.Background(), "test-app-1")
	suite.NoError(err)
	suite.True(exists)
}

func (suite *FileBasedStoreTestSuite) TestIsApplicationExists_NotFound() {
	exists, err := suite.store.IsApplicationExists(context.Background(), "nonexistent")
	suite.NoError(err)
	suite.False(exists)
}

// Tests for IsApplicationDeclarative

func (suite *FileBasedStoreTestSuite) TestIsApplicationDeclarative_ExistingApp() {
	app := applicationConfigDAO{ID: "declarative-app-1"}
	err := suite.store.CreateApplication(context.Background(), app)
	suite.NoError(err)

	isDeclarative := suite.store.IsApplicationDeclarative(context.Background(), "declarative-app-1")
	suite.True(isDeclarative)
}

func (suite *FileBasedStoreTestSuite) TestIsApplicationDeclarative_NotFound() {
	isDeclarative := suite.store.IsApplicationDeclarative(context.Background(), "nonexistent")
	suite.False(isDeclarative)
}

func (suite *FileBasedStoreTestSuite) TestIsApplicationDeclarative_MultipleApps() {
	apps := []applicationConfigDAO{
		{ID: "app-1"},
		{ID: "app-2"},
		{ID: "app-3"},
	}

	for _, app := range apps {
		err := suite.store.CreateApplication(context.Background(), app)
		suite.NoError(err)
	}

	for _, app := range apps {
		isDeclarative := suite.store.IsApplicationDeclarative(context.Background(), app.ID)
		suite.True(isDeclarative, "Application %s should be marked as declarative", app.ID)
	}
}

func (suite *FileBasedStoreTestSuite) TestNewFileBasedStore() {
	store, _ := newFileBasedStore()

	suite.NotNil(store)
	suite.IsType(&fileBasedStore{}, store)
}
