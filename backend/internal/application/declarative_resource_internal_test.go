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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
)

// ValidateApplicationWrapperTestSuite tests the validateApplicationWrapper function.
type ValidateApplicationWrapperTestSuite struct {
	suite.Suite
	mockFileStore *applicationStoreInterfaceMock
	mockDBStore   *applicationStoreInterfaceMock
}

func TestValidateApplicationWrapperTestSuite(t *testing.T) {
	suite.Run(t, new(ValidateApplicationWrapperTestSuite))
}

func (s *ValidateApplicationWrapperTestSuite) SetupTest() {
	s.mockFileStore = newApplicationStoreInterfaceMock(s.T())
	s.mockDBStore = newApplicationStoreInterfaceMock(s.T())
}

func (s *ValidateApplicationWrapperTestSuite) TestValidateApplicationWrapper_Success_DeclarativeMode() {
	// Test successful validation in declarative-only mode (no DB store)
	app := &model.ApplicationProcessedDTO{
		ID:   "app123",
		Name: "Test Application",
	}

	// Mock file store check - app does not exist
	s.mockFileStore.EXPECT().IsApplicationExists("app123").Return(false, nil)

	err := validateApplicationWrapper(app, s.mockFileStore, nil)

	assert.Nil(s.T(), err)
}

func (s *ValidateApplicationWrapperTestSuite) TestValidateApplicationWrapper_Success_CompositeMode() {
	// Test successful validation in composite mode (both file and DB stores)
	app := &model.ApplicationProcessedDTO{
		ID:   "app456",
		Name: "Another Application",
	}

	// Mock file store check - app does not exist
	s.mockFileStore.EXPECT().IsApplicationExists("app456").Return(false, nil)

	// Mock DB store check - app does not exist
	s.mockDBStore.EXPECT().IsApplicationExists("app456").Return(false, nil)

	err := validateApplicationWrapper(app, s.mockFileStore, s.mockDBStore)

	assert.Nil(s.T(), err)
}

func (s *ValidateApplicationWrapperTestSuite) TestValidateApplicationWrapper_InvalidType() {
	// Test with invalid data type
	invalidData := "not an ApplicationProcessedDTO"

	err := validateApplicationWrapper(invalidData, s.mockFileStore, nil)

	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid type: expected *ApplicationProcessedDTO")
}

func (s *ValidateApplicationWrapperTestSuite) TestValidateApplicationWrapper_EmptyName() {
	// Test with empty application name
	app := &model.ApplicationProcessedDTO{
		ID:   "app789",
		Name: "",
	}

	err := validateApplicationWrapper(app, s.mockFileStore, nil)

	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "application name cannot be empty")
}

func (s *ValidateApplicationWrapperTestSuite) TestValidateApplicationWrapper_DuplicateInFileStore() {
	// Test duplicate ID in file store
	app := &model.ApplicationProcessedDTO{
		ID:   "duplicate123",
		Name: "Duplicate App",
	}

	// Mock file store check - app already exists
	s.mockFileStore.EXPECT().IsApplicationExists("duplicate123").Return(true, nil)

	err := validateApplicationWrapper(app, s.mockFileStore, nil)

	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "duplicate application ID 'duplicate123'")
	assert.Contains(s.T(), err.Error(), "already exists in declarative resources")
}

func (s *ValidateApplicationWrapperTestSuite) TestValidateApplicationWrapper_DuplicateInDBStore() {
	// Test duplicate ID in database store (composite mode)
	app := &model.ApplicationProcessedDTO{
		ID:   "duplicate456",
		Name: "Another Duplicate",
	}

	// Mock file store check - app does not exist in file store
	s.mockFileStore.EXPECT().IsApplicationExists("duplicate456").Return(false, nil)

	// Mock DB store check - app already exists in DB
	s.mockDBStore.EXPECT().IsApplicationExists("duplicate456").Return(true, nil)

	err := validateApplicationWrapper(app, s.mockFileStore, s.mockDBStore)

	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "duplicate application ID 'duplicate456'")
	assert.Contains(s.T(), err.Error(), "already exists in the database store")
}

func (s *ValidateApplicationWrapperTestSuite) TestValidateApplicationWrapper_FileStoreCheckError() {
	// Test when file store check returns an error
	app := &model.ApplicationProcessedDTO{
		ID:   "app999",
		Name: "Error Test App",
	}

	// Mock file store check returns an error
	s.mockFileStore.EXPECT().IsApplicationExists("app999").Return(false, assert.AnError)

	err := validateApplicationWrapper(app, s.mockFileStore, nil)

	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "failed to check application existence")
}

func (s *ValidateApplicationWrapperTestSuite) TestValidateApplicationWrapper_DBStoreCheckError() {
	// Test when DB store check returns an error (composite mode)
	app := &model.ApplicationProcessedDTO{
		ID:   "app888",
		Name: "DB Error Test App",
	}

	// Mock file store check - app does not exist
	s.mockFileStore.EXPECT().IsApplicationExists("app888").Return(false, nil)

	// Mock DB store check returns an error
	s.mockDBStore.EXPECT().IsApplicationExists("app888").Return(false, assert.AnError)

	err := validateApplicationWrapper(app, s.mockFileStore, s.mockDBStore)

	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "failed to check application existence")
}

func (s *ValidateApplicationWrapperTestSuite) TestValidateApplicationWrapper_DeclarativeMode_NoDBStoreCheck() {
	// Test that DB store is not checked when it's nil (declarative-only mode)
	app := &model.ApplicationProcessedDTO{
		ID:   "app777",
		Name: "Declarative Only App",
	}

	// Mock file store check - app does not exist
	s.mockFileStore.EXPECT().IsApplicationExists("app777").Return(false, nil)

	// DB store should not be called (nil check prevents it)
	// No expectations set for mockDBStore

	err := validateApplicationWrapper(app, s.mockFileStore, nil)

	assert.Nil(s.T(), err)
}

func (s *ValidateApplicationWrapperTestSuite) TestValidateApplicationWrapper_MultipleValidations() {
	// Test multiple validations to ensure store mocks are reusable
	testCases := []struct {
		name          string
		app           *model.ApplicationProcessedDTO
		fileExists    bool
		dbExists      bool
		useDBStore    bool
		expectedError bool
		errorContains string
	}{
		{
			name:          "Valid app in declarative mode",
			app:           &model.ApplicationProcessedDTO{ID: "app1", Name: "App 1"},
			fileExists:    false,
			useDBStore:    false,
			expectedError: false,
		},
		{
			name:          "Valid app in composite mode",
			app:           &model.ApplicationProcessedDTO{ID: "app2", Name: "App 2"},
			fileExists:    false,
			dbExists:      false,
			useDBStore:    true,
			expectedError: false,
		},
		{
			name:          "Duplicate in file store",
			app:           &model.ApplicationProcessedDTO{ID: "app3", Name: "App 3"},
			fileExists:    true,
			useDBStore:    false,
			expectedError: true,
			errorContains: "already exists in declarative resources",
		},
		{
			name:          "Duplicate in DB store",
			app:           &model.ApplicationProcessedDTO{ID: "app4", Name: "App 4"},
			fileExists:    false,
			dbExists:      true,
			useDBStore:    true,
			expectedError: true,
			errorContains: "already exists in the database store",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Create fresh mocks for each test case
			mockFileStore := newApplicationStoreInterfaceMock(s.T())
			var mockDBStore applicationStoreInterface
			if tc.useDBStore {
				mockDBStore = newApplicationStoreInterfaceMock(s.T())
			}

			// Setup expectations
			mockFileStore.EXPECT().IsApplicationExists(tc.app.ID).Return(tc.fileExists, nil)
			if tc.useDBStore && !tc.fileExists {
				dbStoreMock := mockDBStore.(*applicationStoreInterfaceMock)
				dbStoreMock.EXPECT().IsApplicationExists(tc.app.ID).Return(tc.dbExists, nil)
			}

			// Execute
			err := validateApplicationWrapper(tc.app, mockFileStore, mockDBStore)

			// Assert
			if tc.expectedError {
				assert.NotNil(s.T(), err)
				if tc.errorContains != "" {
					assert.Contains(s.T(), err.Error(), tc.errorContains)
				}
			} else {
				assert.Nil(s.T(), err)
			}
		})
	}
}
