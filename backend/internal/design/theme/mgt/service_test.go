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

package thememgt

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
)

// Test Suite
type ThemeServiceTestSuite struct {
	suite.Suite
	mockStore *themeMgtStoreInterfaceMock
	service   ThemeMgtServiceInterface
}

func TestThemeServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ThemeServiceTestSuite))
}

func (suite *ThemeServiceTestSuite) SetupTest() {
	// Initialize config runtime with default values
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	if err != nil {
		suite.Fail("Failed to initialize runtime", err)
	}

	suite.mockStore = newThemeMgtStoreInterfaceMock(suite.T())
	suite.service = newThemeMgtService(suite.mockStore)
}

// Test GetThemeList - Success
func (suite *ThemeServiceTestSuite) TestGetThemeList_Success() {
	themes := []Theme{
		{
			ID:          "theme-1",
			DisplayName: "Classic Theme",
			Description: "A classic theme",
			Theme:       json.RawMessage(`{"colors": {"primary": "#007bff"}}`),
		},
		{
			ID:          "theme-2",
			DisplayName: "Dark Theme",
			Description: "A dark theme",
			Theme:       json.RawMessage(`{"colors": {"primary": "#000000"}}`),
		},
	}

	suite.mockStore.On("GetThemeListCount").Return(2, nil)
	suite.mockStore.On("GetThemeList", 10, 0).Return(themes, nil)

	result, err := suite.service.GetThemeList(10, 0)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 2, result.TotalResults)
	assert.Equal(suite.T(), 2, result.Count)
	assert.Equal(suite.T(), 1, result.StartIndex)
	assert.Len(suite.T(), result.Themes, 2)
}

// Test GetThemeList - Store Count Error
func (suite *ThemeServiceTestSuite) TestGetThemeList_CountError() {
	suite.mockStore.On("GetThemeListCount").Return(0, errors.New("database error"))

	result, err := suite.service.GetThemeList(10, 0)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
}

// Test GetThemeList - Store Error
func (suite *ThemeServiceTestSuite) TestGetThemeList_StoreError() {
	suite.mockStore.On("GetThemeListCount").Return(2, nil)
	suite.mockStore.On("GetThemeList", 10, 0).Return(nil, errors.New("database error"))

	result, err := suite.service.GetThemeList(10, 0)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
}

// Test GetThemeList - Invalid Pagination
func (suite *ThemeServiceTestSuite) TestGetThemeList_InvalidLimit() {
	result, err := suite.service.GetThemeList(-1, 0)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "THM-1008", err.Code)
}

func (suite *ThemeServiceTestSuite) TestGetThemeList_InvalidOffset() {
	result, err := suite.service.GetThemeList(10, -1)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "THM-1009", err.Code)
}

// Test CreateTheme - Success
func (suite *ThemeServiceTestSuite) TestCreateTheme_Success() {
	themeRequest := CreateThemeRequest{
		DisplayName: "New Theme",
		Description: "A new theme",
		Theme:       json.RawMessage(`{"colors": {"primary": "#ff0000"}}`),
	}

	suite.mockStore.On("CreateTheme", mock.AnythingOfType("string"), themeRequest).Return(nil)

	result, err := suite.service.CreateTheme(themeRequest)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "New Theme", result.DisplayName)
	assert.Equal(suite.T(), "A new theme", result.Description)
	assert.NotEmpty(suite.T(), result.ID)
}

// Test CreateTheme - Missing Display Name
func (suite *ThemeServiceTestSuite) TestCreateTheme_MissingDisplayName() {
	themeRequest := CreateThemeRequest{
		DisplayName: "",
		Description: "A theme without name",
		Theme:       json.RawMessage(`{"colors": {"primary": "#ff0000"}}`),
	}

	result, err := suite.service.CreateTheme(themeRequest)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "THM-1005", err.Code)
}

// Test CreateTheme - Declarative mode enabled
func (suite *ThemeServiceTestSuite) TestCreateTheme_DeclarativeModeEnabled() {
	runtime := config.GetThunderRuntime()
	runtime.Config.Theme.Store = "declarative"

	themeRequest := CreateThemeRequest{
		DisplayName: "Declarative Theme",
		Description: "Should be blocked",
		Theme:       json.RawMessage(`{"colors": {"primary": "#ff0000"}}`),
	}

	result, err := suite.service.CreateTheme(themeRequest)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "THM-1014", err.Code)
}

// Test CreateTheme - Invalid Theme JSON
func (suite *ThemeServiceTestSuite) TestCreateTheme_InvalidJSON() {
	themeRequest := CreateThemeRequest{
		DisplayName: "Theme",
		Description: "Invalid JSON theme",
		Theme:       json.RawMessage(`{invalid json}`),
	}

	result, err := suite.service.CreateTheme(themeRequest)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "THM-1007", err.Code)
}

// Test CreateTheme - Store Error
func (suite *ThemeServiceTestSuite) TestCreateTheme_StoreError() {
	themeRequest := CreateThemeRequest{
		DisplayName: "Theme",
		Description: "A theme",
		Theme:       json.RawMessage(`{"colors": {"primary": "#ff0000"}}`),
	}

	suite.mockStore.On("CreateTheme", mock.AnythingOfType("string"), themeRequest).Return(errors.New("database error"))

	result, err := suite.service.CreateTheme(themeRequest)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
}

// Test GetTheme - Success
func (suite *ThemeServiceTestSuite) TestGetTheme_Success() {
	theme := Theme{
		ID:          "theme-123",
		DisplayName: "Test Theme",
		Description: "A test theme",
		Theme:       json.RawMessage(`{"colors": {"primary": "#007bff"}}`),
	}

	suite.mockStore.On("GetTheme", "theme-123").Return(theme, nil)

	result, err := suite.service.GetTheme("theme-123")

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "theme-123", result.ID)
	assert.Equal(suite.T(), "Test Theme", result.DisplayName)
}

// Test GetTheme - Invalid ID
func (suite *ThemeServiceTestSuite) TestGetTheme_InvalidID() {
	result, err := suite.service.GetTheme("")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "THM-1002", err.Code)
}

// Test GetTheme - Not Found
func (suite *ThemeServiceTestSuite) TestGetTheme_NotFound() {
	suite.mockStore.On("GetTheme", "non-existent").Return(Theme{}, errThemeNotFound)

	result, err := suite.service.GetTheme("non-existent")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "THM-1003", err.Code)
}

// Test GetTheme - Store Error
func (suite *ThemeServiceTestSuite) TestGetTheme_StoreError() {
	suite.mockStore.On("GetTheme", "theme-123").Return(Theme{}, errors.New("database error"))

	result, err := suite.service.GetTheme("theme-123")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
}

// Test UpdateTheme - Success
func (suite *ThemeServiceTestSuite) TestUpdateTheme_Success() {
	updateRequest := UpdateThemeRequest{
		DisplayName: "Updated Theme",
		Description: "An updated theme",
		Theme:       json.RawMessage(`{"colors": {"primary": "#00ff00"}}`),
	}

	suite.mockStore.On("IsThemeDeclarative", "theme-123").Return(false)
	suite.mockStore.On("IsThemeExist", "theme-123").Return(true, nil)
	suite.mockStore.On("UpdateTheme", "theme-123", updateRequest).Return(nil)

	result, err := suite.service.UpdateTheme("theme-123", updateRequest)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "theme-123", result.ID)
	assert.Equal(suite.T(), "Updated Theme", result.DisplayName)
}

// Test UpdateTheme - Invalid ID
func (suite *ThemeServiceTestSuite) TestUpdateTheme_InvalidID() {
	updateRequest := UpdateThemeRequest{
		DisplayName: "Theme",
		Description: "A theme",
		Theme:       json.RawMessage(`{"colors": {}}`),
	}

	result, err := suite.service.UpdateTheme("", updateRequest)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "THM-1002", err.Code)
}

// Test UpdateTheme - Missing Display Name
func (suite *ThemeServiceTestSuite) TestUpdateTheme_MissingDisplayName() {
	updateRequest := UpdateThemeRequest{
		DisplayName: "",
		Description: "A theme",
		Theme:       json.RawMessage(`{"colors": {}}`),
	}

	result, err := suite.service.UpdateTheme("theme-123", updateRequest)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "THM-1005", err.Code)
}

// Test UpdateTheme - Not Found
func (suite *ThemeServiceTestSuite) TestUpdateTheme_NotFound() {
	updateRequest := UpdateThemeRequest{
		DisplayName: "Theme",
		Description: "A theme",
		Theme:       json.RawMessage(`{"colors": {}}`),
	}

	suite.mockStore.On("IsThemeDeclarative", "non-existent").Return(false)
	suite.mockStore.On("IsThemeExist", "non-existent").Return(false, nil)

	result, err := suite.service.UpdateTheme("non-existent", updateRequest)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "THM-1003", err.Code)
}

// Test UpdateTheme - Invalid JSON
func (suite *ThemeServiceTestSuite) TestUpdateTheme_InvalidJSON() {
	updateRequest := UpdateThemeRequest{
		DisplayName: "Theme",
		Description: "A theme",
		Theme:       json.RawMessage(`{invalid}`),
	}
	suite.mockStore.On("IsThemeDeclarative", "theme-123").Return(false)
	result, err := suite.service.UpdateTheme("theme-123", updateRequest)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "THM-1007", err.Code)
}

// Test DeleteTheme - Success
func (suite *ThemeServiceTestSuite) TestDeleteTheme_Success() {
	suite.mockStore.On("IsThemeDeclarative", "theme-123").Return(false)
	suite.mockStore.On("IsThemeExist", "theme-123").Return(true, nil)
	suite.mockStore.On("GetApplicationsCountByThemeID", "theme-123").Return(0, nil)
	suite.mockStore.On("DeleteTheme", "theme-123").Return(nil)

	err := suite.service.DeleteTheme("theme-123")

	assert.Nil(suite.T(), err)
}

// Test DeleteTheme - Invalid ID
func (suite *ThemeServiceTestSuite) TestDeleteTheme_InvalidID() {
	err := suite.service.DeleteTheme("")

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "THM-1002", err.Code)
}

// Test DeleteTheme - Not Found (idempotent delete returns success)
func (suite *ThemeServiceTestSuite) TestDeleteTheme_NotFound() {
	suite.mockStore.On("IsThemeDeclarative", "non-existent").Return(false)
	suite.mockStore.On("IsThemeExist", "non-existent").Return(false, nil)

	err := suite.service.DeleteTheme("non-existent")

	assert.Nil(suite.T(), err)
}

// Test DeleteTheme - Theme In Use
func (suite *ThemeServiceTestSuite) TestDeleteTheme_InUse() {
	suite.mockStore.On("IsThemeDeclarative", "theme-123").Return(false)
	suite.mockStore.On("IsThemeExist", "theme-123").Return(true, nil)
	suite.mockStore.On("GetApplicationsCountByThemeID", "theme-123").Return(3, nil)

	err := suite.service.DeleteTheme("theme-123")

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "THM-1004", err.Code)
}

// Test DeleteTheme - Store Error
func (suite *ThemeServiceTestSuite) TestDeleteTheme_StoreError() {
	suite.mockStore.On("IsThemeDeclarative", "theme-123").Return(false)
	suite.mockStore.On("IsThemeExist", "theme-123").Return(true, nil)
	suite.mockStore.On("GetApplicationsCountByThemeID", "theme-123").Return(0, nil)
	suite.mockStore.On("DeleteTheme", "theme-123").Return(errors.New("database error"))

	err := suite.service.DeleteTheme("theme-123")

	assert.NotNil(suite.T(), err)
}

// Test IsThemeExist - Exists
func (suite *ThemeServiceTestSuite) TestIsThemeExist_True() {
	suite.mockStore.On("IsThemeExist", "theme-123").Return(true, nil)

	exists, err := suite.service.IsThemeExist("theme-123")

	assert.Nil(suite.T(), err)
	assert.True(suite.T(), exists)
}

// Test IsThemeExist - Not Exists
func (suite *ThemeServiceTestSuite) TestIsThemeExist_False() {
	suite.mockStore.On("IsThemeExist", "non-existent").Return(false, nil)

	exists, err := suite.service.IsThemeExist("non-existent")

	assert.Nil(suite.T(), err)
	assert.False(suite.T(), exists)
}

// Test IsThemeExist - Store Error
func (suite *ThemeServiceTestSuite) TestIsThemeExist_StoreError() {
	suite.mockStore.On("IsThemeExist", "theme-123").Return(false, errors.New("database error"))

	exists, err := suite.service.IsThemeExist("theme-123")

	assert.NotNil(suite.T(), err)
	assert.False(suite.T(), exists)
}
