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
	"encoding/json"

	"github.com/stretchr/testify/mock"

	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
)

// CompositeStoreTestSuite tests the composite application store functionality.
type CompositeStoreTestSuite struct {
	suite.Suite
	fileStore      applicationStoreInterface
	dbStoreMock    *applicationStoreInterfaceMock
	compositeStore *compositeApplicationStore
}

// SetupTest sets up the test environment.
func (suite *CompositeStoreTestSuite) SetupTest() {
	// Clear the singleton entity store to avoid state leakage between tests
	_ = entity.GetInstance().Clear()

	// Create NEW file-based store for each test to avoid state leakage
	suite.fileStore, _ = newFileBasedStore()

	// Create mock DB store
	suite.dbStoreMock = newApplicationStoreInterfaceMock(suite.T())

	// Create composite store
	suite.compositeStore = newCompositeApplicationStore(suite.fileStore, suite.dbStoreMock)
}

// TestCompositeStore_GetApplicationByID tests retrieving applications from composite store.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetApplicationByID() {
	testCases := []struct {
		name           string
		appID          string
		setupFileStore func()
		setupDBStore   func()
		want           *applicationConfigDAO
		wantErr        bool
	}{
		{
			name:  "retrieves from DB store",
			appID: "db-app-1",
			setupFileStore: func() {
				// File store doesn't have this app
			},
			setupDBStore: func() {
				suite.dbStoreMock.On("GetApplicationByID", mock.Anything, "db-app-1").
					Return(&applicationConfigDAO{ID: "db-app-1"}, nil).
					Once()
			},
			want: &applicationConfigDAO{ID: "db-app-1"},
		},
		{
			name:  "retrieves from file store when not in DB",
			appID: "file-app-1",
			setupFileStore: func() {
				err := suite.fileStore.CreateApplication(context.Background(), applicationConfigDAO{
					ID: "file-app-1",
				})
				suite.NoError(err)
			},
			setupDBStore: func() {
				suite.dbStoreMock.On("GetApplicationByID", mock.Anything, "file-app-1").
					Return(nil, model.ApplicationNotFoundError).
					Once()
			},
			want: &applicationConfigDAO{ID: "file-app-1"},
		},
		{
			name:  "not found in either store",
			appID: "nonexistent",
			setupFileStore: func() {
				// App not in file store
			},
			setupDBStore: func() {
				suite.dbStoreMock.On("GetApplicationByID", mock.Anything, "nonexistent").
					Return(nil, model.ApplicationNotFoundError).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // Fresh setup for each test
			tc.setupFileStore()
			tc.setupDBStore()

			got, err := suite.compositeStore.GetApplicationByID(context.Background(), tc.appID)

			if tc.wantErr {
				suite.Error(err)
				suite.True(errors.Is(err, model.ApplicationNotFoundError))
			} else {
				suite.NoError(err)
				suite.Equal(tc.want.ID, got.ID)
			}
		})
	}
}

// TestCompositeStore_GetOAuthConfigByEntityID tests retrieving OAuth configs.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetOAuthConfigByEntityID() {
	suite.Run("retrieves from DB store", func() {
		suite.dbStoreMock.On("GetOAuthConfigByEntityID", mock.Anything, "entity-123").
			Return(&oauthConfigDAO{EntityID: "entity-123"}, nil).
			Once()

		got, err := suite.compositeStore.GetOAuthConfigByEntityID(context.Background(), "entity-123")
		suite.NoError(err)
		suite.Equal("entity-123", got.EntityID)
	})

	suite.Run("not found in either store", func() {
		suite.dbStoreMock.On("GetOAuthConfigByEntityID", mock.Anything, "entity-456").
			Return(nil, model.ApplicationNotFoundError).
			Once()

		got, err := suite.compositeStore.GetOAuthConfigByEntityID(context.Background(), "entity-456")
		suite.Error(err)
		suite.Nil(got)
	})
}

// TestCompositeStore_CreateApplication tests creating applications.
func (suite *CompositeStoreTestSuite) TestCompositeStore_CreateApplication() {
	suite.Run("creates in DB store only", func() {
		app := applicationConfigDAO{ID: "new-app-1"}

		suite.dbStoreMock.On("CreateApplication", mock.Anything, app).
			Return(nil).
			Once()

		err := suite.compositeStore.CreateApplication(context.Background(), app)
		suite.NoError(err)
	})

	suite.Run("propagates DB store error", func() {
		app := applicationConfigDAO{ID: "new-app-2"}

		dbErr := errors.New("database error")
		suite.dbStoreMock.On("CreateApplication", mock.Anything, app).
			Return(dbErr).
			Once()

		err := suite.compositeStore.CreateApplication(context.Background(), app)
		suite.Error(err)
		suite.Equal(dbErr, err)
	})
}

// TestCompositeStore_CreateOAuthConfig tests creating OAuth configs.
func (suite *CompositeStoreTestSuite) TestCompositeStore_CreateOAuthConfig() {
	rawJSON := json.RawMessage(`{"redirect_uris":["https://example.com/cb"]}`)

	suite.Run("delegates to DB store", func() {
		suite.dbStoreMock.On("CreateOAuthConfig", mock.Anything, "entity-1", rawJSON).
			Return(nil).
			Once()

		err := suite.compositeStore.CreateOAuthConfig(context.Background(), "entity-1", rawJSON)
		suite.NoError(err)
	})
}

// TestCompositeStore_UpdateApplication tests updating applications.
func (suite *CompositeStoreTestSuite) TestCompositeStore_UpdateApplication() {
	suite.Run("updates DB app successfully", func() {
		app := applicationConfigDAO{ID: "app-1", AuthFlowID: "new-flow"}

		suite.dbStoreMock.On("UpdateApplication", mock.Anything, app).
			Return(nil).
			Once()

		err := suite.compositeStore.UpdateApplication(context.Background(), app)
		suite.NoError(err)
	})

	suite.Run("propagates DB store error", func() {
		app := applicationConfigDAO{ID: "app-2"}

		dbErr := errors.New("update failed")
		suite.dbStoreMock.On("UpdateApplication", mock.Anything, app).
			Return(dbErr).
			Once()

		err := suite.compositeStore.UpdateApplication(context.Background(), app)
		suite.Error(err)
		suite.Equal(dbErr, err)
	})
}

// TestCompositeStore_UpdateOAuthConfig tests updating OAuth configs.
func (suite *CompositeStoreTestSuite) TestCompositeStore_UpdateOAuthConfig() {
	rawJSON := json.RawMessage(`{"redirect_uris":["https://new.example.com/cb"]}`)

	suite.Run("delegates to DB store", func() {
		suite.dbStoreMock.On("UpdateOAuthConfig", mock.Anything, "entity-1", rawJSON).
			Return(nil).
			Once()

		err := suite.compositeStore.UpdateOAuthConfig(context.Background(), "entity-1", rawJSON)
		suite.NoError(err)
	})
}

// TestCompositeStore_DeleteApplication tests deleting applications.
func (suite *CompositeStoreTestSuite) TestCompositeStore_DeleteApplication() {
	suite.Run("deletes DB app successfully", func() {
		suite.dbStoreMock.On("DeleteApplication", mock.Anything, "app-1").
			Return(nil).
			Once()

		err := suite.compositeStore.DeleteApplication(context.Background(), "app-1")
		suite.NoError(err)
	})

	suite.Run("propagates DB store error", func() {
		dbErr := errors.New("delete failed")
		suite.dbStoreMock.On("DeleteApplication", mock.Anything, "app-2").
			Return(dbErr).
			Once()

		err := suite.compositeStore.DeleteApplication(context.Background(), "app-2")
		suite.Error(err)
		suite.Equal(dbErr, err)
	})
}

// TestCompositeStore_DeleteOAuthConfig tests deleting OAuth configs.
func (suite *CompositeStoreTestSuite) TestCompositeStore_DeleteOAuthConfig() {
	suite.Run("delegates to DB store", func() {
		suite.dbStoreMock.On("DeleteOAuthConfig", mock.Anything, "entity-1").
			Return(nil).
			Once()

		err := suite.compositeStore.DeleteOAuthConfig(context.Background(), "entity-1")
		suite.NoError(err)
	})
}

// TestCompositeStore_IsApplicationExists tests existence checks across both stores.
func (suite *CompositeStoreTestSuite) TestCompositeStore_IsApplicationExists() {
	suite.Run("exists in DB store", func() {
		suite.dbStoreMock.On("IsApplicationExists", mock.Anything, "db-app-1").
			Return(true, nil).
			Once()

		exists, err := suite.compositeStore.IsApplicationExists(context.Background(), "db-app-1")
		suite.NoError(err)
		suite.True(exists)
		suite.dbStoreMock.AssertExpectations(suite.T())
	})

	suite.Run("exists in file store", func() {
		suite.SetupTest() // Fresh setup

		err := suite.fileStore.CreateApplication(context.Background(), applicationConfigDAO{
			ID: "file-app-1",
		})
		suite.NoError(err)

		// DB mock should NOT be called since file store has it
		exists, err := suite.compositeStore.IsApplicationExists(context.Background(), "file-app-1")
		suite.NoError(err)
		suite.True(exists)
	})

	suite.Run("not found in either store", func() {
		suite.dbStoreMock.On("IsApplicationExists", mock.Anything, "nonexistent").
			Return(false, nil).
			Once()

		exists, err := suite.compositeStore.IsApplicationExists(context.Background(), "nonexistent")
		suite.NoError(err)
		suite.False(exists)
		suite.dbStoreMock.AssertExpectations(suite.T())
	})

	suite.Run("propagates DB error", func() {
		dbErr := errors.New("db error")
		suite.dbStoreMock.On("IsApplicationExists", mock.Anything, "error-app").
			Return(false, dbErr).
			Once()

		exists, err := suite.compositeStore.IsApplicationExists(context.Background(), "error-app")
		suite.Error(err)
		suite.Equal(dbErr, err)
		suite.False(exists)
		suite.dbStoreMock.AssertExpectations(suite.T())
	})
}

// TestCompositeStore_IsApplicationDeclarative tests checking if an application is declarative.
func (suite *CompositeStoreTestSuite) TestCompositeStore_IsApplicationDeclarative() {
	suite.Run("returns true for app in file store", func() {
		suite.SetupTest() // Fresh setup

		err := suite.fileStore.CreateApplication(context.Background(), applicationConfigDAO{
			ID: "declarative-app-1",
		})
		suite.NoError(err)

		isDeclarative := suite.compositeStore.IsApplicationDeclarative(context.Background(), "declarative-app-1")
		suite.True(isDeclarative)
	})

	suite.Run("returns false for app not in file store", func() {
		isDeclarative := suite.compositeStore.IsApplicationDeclarative(context.Background(), "db-app-1")
		suite.False(isDeclarative)
	})

	suite.Run("returns false for non-existent app", func() {
		isDeclarative := suite.compositeStore.IsApplicationDeclarative(context.Background(), "nonexistent")
		suite.False(isDeclarative)
	})
}

// TestCompositeStore_GetTotalApplicationCount tests counting applications from both stores.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetTotalApplicationCount() {
	suite.Run("returns DB count when no file apps", func() {
		suite.dbStoreMock.On("GetTotalApplicationCount", mock.Anything).
			Return(5, nil).
			Once()

		count, err := suite.compositeStore.GetTotalApplicationCount(context.Background())
		suite.NoError(err)
		suite.Equal(5, count)
	})

	suite.Run("includes file store count", func() {
		suite.SetupTest() // Fresh setup

		_ = suite.fileStore.CreateApplication(context.Background(), applicationConfigDAO{ID: "file-app-1"})
		_ = suite.fileStore.CreateApplication(context.Background(), applicationConfigDAO{ID: "file-app-2"})

		suite.dbStoreMock.On("GetTotalApplicationCount", mock.Anything).
			Return(3, nil).
			Once()

		count, err := suite.compositeStore.GetTotalApplicationCount(context.Background())
		suite.NoError(err)
		suite.Equal(5, count) // 3 from DB + 2 from file
	})

	suite.Run("propagates DB error", func() {
		dbErr := errors.New("database error")
		suite.dbStoreMock.On("GetTotalApplicationCount", mock.Anything).
			Return(0, dbErr).
			Once()

		count, err := suite.compositeStore.GetTotalApplicationCount(context.Background())
		suite.Error(err)
		suite.Equal(dbErr, err)
		suite.Equal(0, count)
	})
}

// TestCompositeStore_GetApplicationList tests retrieving the application list.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetApplicationList() {
	suite.Run("merges applications from both stores", func() {
		suite.SetupTest() // Fresh setup

		dbApps := []applicationConfigDAO{
			{ID: "db-app-1"},
			{ID: "db-app-2"},
		}

		suite.dbStoreMock.On("GetTotalApplicationCount", mock.Anything).
			Return(2, nil).
			Once()
		suite.dbStoreMock.On("GetApplicationList", mock.Anything).
			Return(dbApps, nil).
			Once()

		_ = suite.fileStore.CreateApplication(context.Background(), applicationConfigDAO{ID: "file-app-1"})

		list, err := suite.compositeStore.GetApplicationList(context.Background())
		suite.NoError(err)
		suite.Len(list, 3)

		// Verify IsReadOnly flags
		for _, app := range list {
			if app.ID == "db-app-1" || app.ID == "db-app-2" {
				suite.False(app.IsReadOnly, "DB app %s should have IsReadOnly=false", app.ID)
			} else if app.ID == "file-app-1" {
				suite.True(app.IsReadOnly, "File app %s should have IsReadOnly=true", app.ID)
			}
		}
	})

	suite.Run("removes duplicates with DB precedence", func() {
		suite.SetupTest() // Fresh setup

		dbApps := []applicationConfigDAO{{ID: "app-1"}}

		suite.dbStoreMock.On("GetTotalApplicationCount", mock.Anything).
			Return(1, nil).
			Once()
		suite.dbStoreMock.On("GetApplicationList", mock.Anything).
			Return(dbApps, nil).
			Once()

		// Add to file store with same ID
		_ = suite.fileStore.CreateApplication(context.Background(), applicationConfigDAO{ID: "app-1"})

		list, err := suite.compositeStore.GetApplicationList(context.Background())
		suite.NoError(err)
		suite.Len(list, 1)
		suite.Equal("app-1", list[0].ID)
		suite.False(list[0].IsReadOnly) // DB version takes precedence
	})

	suite.Run("propagates DB error", func() {
		dbErr := errors.New("database error")
		suite.dbStoreMock.On("GetTotalApplicationCount", mock.Anything).
			Return(0, dbErr).
			Once()

		list, err := suite.compositeStore.GetApplicationList(context.Background())
		suite.Error(err)
		suite.Nil(list)
	})
}

// TestCompositeStore_MergeAndDeduplicate tests the merge and deduplication logic.
func (suite *CompositeStoreTestSuite) TestCompositeStore_MergeAndDeduplicate() {
	suite.Run("db apps get precedence over file apps", func() {
		dbApps := []applicationConfigDAO{
			{ID: "app-1"},
			{ID: "app-2"},
		}
		fileApps := []applicationConfigDAO{
			{ID: "app-1"}, // Duplicate ID
			{ID: "app-3"},
		}

		result := mergeAndDeduplicateAppConfigs(dbApps, fileApps)

		suite.Len(result, 3)
		suite.Equal("app-1", result[0].ID)
		suite.False(result[0].IsReadOnly)
		suite.Equal("app-3", result[2].ID)
		suite.True(result[2].IsReadOnly)
	})

	suite.Run("marks DB apps as mutable", func() {
		dbApps := []applicationConfigDAO{{ID: "db-app-1"}}
		fileApps := []applicationConfigDAO{}

		result := mergeAndDeduplicateAppConfigs(dbApps, fileApps)

		suite.Len(result, 1)
		suite.False(result[0].IsReadOnly)
	})

	suite.Run("marks file apps as immutable", func() {
		dbApps := []applicationConfigDAO{}
		fileApps := []applicationConfigDAO{{ID: "file-app-1"}}

		result := mergeAndDeduplicateAppConfigs(dbApps, fileApps)

		suite.Len(result, 1)
		suite.True(result[0].IsReadOnly)
	})

	suite.Run("handles empty lists", func() {
		result := mergeAndDeduplicateAppConfigs([]applicationConfigDAO{}, []applicationConfigDAO{})
		suite.Empty(result)
	})
}

// TestCompositeStoreTestSuite runs the composite store test suite.
func TestCompositeStoreTestSuite(t *testing.T) {
	// Initialize entity store for file-based store
	_ = entity.GetInstance()
	suite.Run(t, new(CompositeStoreTestSuite))
}
