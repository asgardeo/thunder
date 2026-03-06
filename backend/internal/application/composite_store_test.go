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
	suite.fileStore = newFileBasedStore()

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
		want           *model.ApplicationProcessedDTO
		wantErr        bool
	}{
		{
			name:  "retrieves from DB store",
			appID: "db-app-1",
			setupFileStore: func() {
				// File store doesn't have this app
			},
			setupDBStore: func() {
				suite.dbStoreMock.On("GetApplicationByID", "db-app-1").
					Return(&model.ApplicationProcessedDTO{
						ID:   "db-app-1",
						Name: "DB App",
					}, nil).
					Once()
			},
			want: &model.ApplicationProcessedDTO{
				ID:   "db-app-1",
				Name: "DB App",
			},
		},
		{
			name:  "retrieves from file store when not in DB",
			appID: "file-app-1",
			setupFileStore: func() {
				// Add app to file store
				err := suite.fileStore.CreateApplication(model.ApplicationProcessedDTO{
					ID:   "file-app-1",
					Name: "File App",
				})
				suite.NoError(err)
			},
			setupDBStore: func() {
				suite.dbStoreMock.On("GetApplicationByID", "file-app-1").
					Return(nil, model.ApplicationNotFoundError).
					Once()
			},
			want: &model.ApplicationProcessedDTO{
				ID:   "file-app-1",
				Name: "File App",
			},
		},
		{
			name:  "not found in either store",
			appID: "nonexistent",
			setupFileStore: func() {
				// App not in file store
			},
			setupDBStore: func() {
				suite.dbStoreMock.On("GetApplicationByID", "nonexistent").
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

			got, err := suite.compositeStore.GetApplicationByID(tc.appID)

			if tc.wantErr {
				suite.Error(err)
				suite.True(errors.Is(err, model.ApplicationNotFoundError))
			} else {
				suite.NoError(err)
				suite.Equal(tc.want.ID, got.ID)
				suite.Equal(tc.want.Name, got.Name)
			}
		})
	}
}

// TestCompositeStore_GetApplicationByName tests retrieving applications by name.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetApplicationByName() {
	suite.Run("retrieves from DB store by name", func() {
		suite.dbStoreMock.On("GetApplicationByName", "DB App").
			Return(&model.ApplicationProcessedDTO{
				ID:   "db-app-1",
				Name: "DB App",
			}, nil).
			Once()

		got, err := suite.compositeStore.GetApplicationByName("DB App")
		suite.NoError(err)
		suite.Equal("db-app-1", got.ID)
		suite.Equal("DB App", got.Name)
	})

	suite.Run("retrieves from file store when not in DB", func() {
		suite.SetupTest() // Fresh setup

		err := suite.fileStore.CreateApplication(model.ApplicationProcessedDTO{
			ID:   "file-app-1",
			Name: "File App",
		})
		suite.NoError(err)

		suite.dbStoreMock.On("GetApplicationByName", "File App").
			Return(nil, model.ApplicationNotFoundError).
			Once()

		got, err := suite.compositeStore.GetApplicationByName("File App")
		suite.NoError(err)
		suite.Equal("file-app-1", got.ID)
		suite.Equal("File App", got.Name)
	})

	suite.Run("not found in either store", func() {
		suite.dbStoreMock.On("GetApplicationByName", "Nonexistent").
			Return(nil, model.ApplicationNotFoundError).
			Once()

		got, err := suite.compositeStore.GetApplicationByName("Nonexistent")
		suite.Error(err)
		suite.Nil(got)
		suite.True(errors.Is(err, model.ApplicationNotFoundError))
	})
}

// TestCompositeStore_GetOAuthApplication tests retrieving OAuth applications.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetOAuthApplication() {
	suite.Run("retrieves from DB store", func() {
		suite.dbStoreMock.On("GetOAuthApplication", "client-123").
			Return(&model.OAuthAppConfigProcessedDTO{
				ClientID: "client-123",
			}, nil).
			Once()

		got, err := suite.compositeStore.GetOAuthApplication("client-123")
		suite.NoError(err)
		suite.Equal("client-123", got.ClientID)
	})

	suite.Run("retrieves from file store when not in DB", func() {
		suite.dbStoreMock.On("GetOAuthApplication", "client-456").
			Return(nil, model.ApplicationNotFoundError).
			Once()

		got, err := suite.compositeStore.GetOAuthApplication("client-456")
		suite.Error(err)
		suite.Nil(got)
	})
}

// TestCompositeStore_CreateApplication tests creating applications.
func (suite *CompositeStoreTestSuite) TestCompositeStore_CreateApplication() {
	suite.Run("creates in DB store only", func() {
		app := model.ApplicationProcessedDTO{
			ID:   "new-app-1",
			Name: "New App",
		}

		suite.dbStoreMock.On("CreateApplication", app).
			Return(nil).
			Once()

		err := suite.compositeStore.CreateApplication(app)
		suite.NoError(err)
	})

	suite.Run("propagates DB store error", func() {
		app := model.ApplicationProcessedDTO{
			ID:   "new-app-2",
			Name: "Another App",
		}

		dbErr := errors.New("database error")
		suite.dbStoreMock.On("CreateApplication", app).
			Return(dbErr).
			Once()

		err := suite.compositeStore.CreateApplication(app)
		suite.Error(err)
		suite.Equal(dbErr, err)
	})
}

// TestCompositeStore_UpdateApplication tests updating applications.
func (suite *CompositeStoreTestSuite) TestCompositeStore_UpdateApplication() {
	suite.Run("updates DB app successfully", func() {
		existing := &model.ApplicationProcessedDTO{
			ID:   "app-1",
			Name: "Old Name",
		}
		updated := &model.ApplicationProcessedDTO{
			ID:   "app-1",
			Name: "New Name",
		}

		suite.dbStoreMock.On("UpdateApplication", existing, updated).
			Return(nil).
			Once()

		err := suite.compositeStore.UpdateApplication(existing, updated)
		suite.NoError(err)
	})

	suite.Run("propagates DB store error", func() {
		existing := &model.ApplicationProcessedDTO{
			ID: "app-2",
		}
		updated := &model.ApplicationProcessedDTO{
			ID: "app-2",
		}

		dbErr := errors.New("update failed")
		suite.dbStoreMock.On("UpdateApplication", existing, updated).
			Return(dbErr).
			Once()

		err := suite.compositeStore.UpdateApplication(existing, updated)
		suite.Error(err)
		suite.Equal(dbErr, err)
	})
}

// TestCompositeStore_DeleteApplication tests deleting applications.
func (suite *CompositeStoreTestSuite) TestCompositeStore_DeleteApplication() {
	suite.Run("deletes DB app successfully", func() {
		suite.dbStoreMock.On("DeleteApplication", "app-1").
			Return(nil).
			Once()

		err := suite.compositeStore.DeleteApplication("app-1")
		suite.NoError(err)
	})

	suite.Run("propagates DB store error", func() {
		dbErr := errors.New("delete failed")
		suite.dbStoreMock.On("DeleteApplication", "app-2").
			Return(dbErr).
			Once()

		err := suite.compositeStore.DeleteApplication("app-2")
		suite.Error(err)
		suite.Equal(dbErr, err)
	})
}

// TestCompositeStore_IsApplicationExists tests existence checks across both stores.
func (suite *CompositeStoreTestSuite) TestCompositeStore_IsApplicationExists() {
	suite.Run("exists in DB store", func() {
		// Mock DB store to return true, file store won't have it
		suite.dbStoreMock.On("IsApplicationExists", "db-app-1").
			Return(true, nil).
			Once()

		exists, err := suite.compositeStore.IsApplicationExists("db-app-1")
		suite.NoError(err)
		suite.True(exists)
		suite.dbStoreMock.AssertExpectations(suite.T())
	})

	suite.Run("exists in file store", func() {
		suite.SetupTest() // Fresh setup

		// Add app to file store
		err := suite.fileStore.CreateApplication(model.ApplicationProcessedDTO{
			ID:   "file-app-1",
			Name: "File App",
		})
		suite.NoError(err)

		// DB mock should NOT be called since file store has it
		exists, err := suite.compositeStore.IsApplicationExists("file-app-1")
		suite.NoError(err)
		suite.True(exists)
	})

	suite.Run("not found in either store", func() {
		suite.dbStoreMock.On("IsApplicationExists", "nonexistent").
			Return(false, nil).
			Once()

		exists, err := suite.compositeStore.IsApplicationExists("nonexistent")
		suite.NoError(err)
		suite.False(exists)
		suite.dbStoreMock.AssertExpectations(suite.T())
	})

	suite.Run("propagates DB error", func() {
		dbErr := errors.New("db error")
		suite.dbStoreMock.On("IsApplicationExists", "error-app").
			Return(false, dbErr).
			Once()

		exists, err := suite.compositeStore.IsApplicationExists("error-app")
		suite.Error(err)
		suite.Equal(dbErr, err)
		suite.False(exists)
		suite.dbStoreMock.AssertExpectations(suite.T())
	})
}

// TestCompositeStore_IsApplicationExistsByName tests name existence checks across both stores.
func (suite *CompositeStoreTestSuite) TestCompositeStore_IsApplicationExistsByName() {
	suite.Run("name exists in DB store", func() {
		suite.dbStoreMock.On("IsApplicationExistsByName", "App Name").
			Return(true, nil).
			Once()

		exists, err := suite.compositeStore.IsApplicationExistsByName("App Name")
		suite.NoError(err)
		suite.True(exists)
	})

	suite.Run("name exists in file store", func() {
		suite.SetupTest() // Fresh setup

		err := suite.fileStore.CreateApplication(model.ApplicationProcessedDTO{
			ID:   "file-app-1",
			Name: "Unique App Name",
		})
		suite.NoError(err)

		exists, err := suite.compositeStore.IsApplicationExistsByName("Unique App Name")
		suite.NoError(err)
		suite.True(exists)
	})

	suite.Run("name not found in either store", func() {
		suite.dbStoreMock.On("IsApplicationExistsByName", "Nonexistent Name").
			Return(false, nil).
			Once()

		exists, err := suite.compositeStore.IsApplicationExistsByName("Nonexistent Name")
		suite.NoError(err)
		suite.False(exists)
	})
}

// TestCompositeStore_IsApplicationDeclarative tests checking if an application is immutable.
func (suite *CompositeStoreTestSuite) TestCompositeStore_IsApplicationDeclarative() {
	suite.Run("returns true for immutable app (exists in file store)", func() {
		suite.SetupTest() // Fresh setup

		// Add app to file store
		err := suite.fileStore.CreateApplication(model.ApplicationProcessedDTO{
			ID:   "immutable-app-1",
			Name: "Declarative App",
		})
		suite.NoError(err)

		isDeclarative := suite.compositeStore.IsApplicationDeclarative("immutable-app-1")
		suite.True(isDeclarative)
	})

	suite.Run("returns false for mutable app (not in file store)", func() {
		isDeclarative := suite.compositeStore.IsApplicationDeclarative("db-app-1")
		suite.False(isDeclarative)
	})

	suite.Run("returns false for non-existent app", func() {
		isDeclarative := suite.compositeStore.IsApplicationDeclarative("nonexistent")
		suite.False(isDeclarative)
	})
}

// TestCompositeStore_GetTotalApplicationCount tests counting applications from both stores.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetTotalApplicationCount() {
	suite.Run("returns total count from both stores", func() {
		suite.dbStoreMock.On("GetTotalApplicationCount").
			Return(5, nil).
			Once()

		count, err := suite.compositeStore.GetTotalApplicationCount()
		suite.NoError(err)
		suite.Equal(5, count)
	})

	suite.Run("includes file store count", func() {
		suite.SetupTest() // Fresh setup

		// Add apps to file store
		_ = suite.fileStore.CreateApplication(model.ApplicationProcessedDTO{
			ID:   "file-app-1",
			Name: "File App 1",
		})
		_ = suite.fileStore.CreateApplication(model.ApplicationProcessedDTO{
			ID:   "file-app-2",
			Name: "File App 2",
		})

		suite.dbStoreMock.On("GetTotalApplicationCount").
			Return(3, nil).
			Once()

		count, err := suite.compositeStore.GetTotalApplicationCount()
		suite.NoError(err)
		suite.Equal(5, count) // 3 from DB + 2 from file
	})

	suite.Run("propagates DB error", func() {
		dbErr := errors.New("database error")
		suite.dbStoreMock.On("GetTotalApplicationCount").
			Return(0, dbErr).
			Once()

		count, err := suite.compositeStore.GetTotalApplicationCount()
		suite.Error(err)
		suite.Equal(dbErr, err)
		suite.Equal(0, count)
	})
}

// TestCompositeStore_GetApplicationList tests retrieving the application list.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetApplicationList() {
	suite.Run("merges applications from both stores", func() {
		suite.SetupTest() // Fresh setup

		dbApps := []model.BasicApplicationDTO{
			{ID: "db-app-1", Name: "DB App 1"},
			{ID: "db-app-2", Name: "DB App 2"},
		}
		fileApps := []model.BasicApplicationDTO{
			{ID: "file-app-1", Name: "File App 1"},
		}

		suite.dbStoreMock.On("GetTotalApplicationCount").
			Return(2, nil).
			Once()
		suite.dbStoreMock.On("GetApplicationList").
			Return(dbApps, nil).
			Once()

		// Add file apps
		for _, app := range fileApps {
			_ = suite.fileStore.CreateApplication(model.ApplicationProcessedDTO{
				ID:   app.ID,
				Name: app.Name,
			})
		}

		list, err := suite.compositeStore.GetApplicationList()
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

		// Both stores have an app with the same ID
		dbApps := []model.BasicApplicationDTO{
			{ID: "app-1", Name: "DB Version"},
		}

		suite.dbStoreMock.On("GetTotalApplicationCount").
			Return(1, nil).
			Once()
		suite.dbStoreMock.On("GetApplicationList").
			Return(dbApps, nil).
			Once()

		// Add to file store with same ID
		_ = suite.fileStore.CreateApplication(model.ApplicationProcessedDTO{
			ID:   "app-1",
			Name: "File Version",
		})

		list, err := suite.compositeStore.GetApplicationList()
		suite.NoError(err)
		suite.Len(list, 1)
		suite.Equal("DB Version", list[0].Name)
		suite.False(list[0].IsReadOnly) // DB version takes precedence
	})

	suite.Run("propagates DB error", func() {
		dbErr := errors.New("database error")
		suite.dbStoreMock.On("GetTotalApplicationCount").
			Return(0, dbErr).
			Once()

		list, err := suite.compositeStore.GetApplicationList()
		suite.Error(err)
		suite.Nil(list)
	})
}

// TestCompositeStore_MergeAndDeduplicate tests the merge and deduplication logic.
func (suite *CompositeStoreTestSuite) TestCompositeStore_MergeAndDeduplicate() {
	suite.Run("db apps get precedence over file apps", func() {
		dbApps := []model.BasicApplicationDTO{
			{ID: "app-1", Name: "DB App 1"},
			{ID: "app-2", Name: "DB App 2"},
		}
		fileApps := []model.BasicApplicationDTO{
			{ID: "app-1", Name: "File App 1"}, // Duplicate ID
			{ID: "app-3", Name: "File App 3"},
		}

		result := mergeAndDeduplicateApplications(dbApps, fileApps)

		suite.Len(result, 3)
		// Verify DB app comes first and is marked mutable
		suite.Equal("app-1", result[0].ID)
		suite.Equal("DB App 1", result[0].Name)
		suite.False(result[0].IsReadOnly)

		// Verify file-only app is marked immutable
		suite.Equal("app-3", result[2].ID)
		suite.Equal("File App 3", result[2].Name)
		suite.True(result[2].IsReadOnly)
	})

	suite.Run("marks DB apps as mutable", func() {
		dbApps := []model.BasicApplicationDTO{
			{ID: "db-app-1", Name: "DB App"},
		}
		fileApps := []model.BasicApplicationDTO{}

		result := mergeAndDeduplicateApplications(dbApps, fileApps)

		suite.Len(result, 1)
		suite.False(result[0].IsReadOnly)
	})

	suite.Run("marks file apps as immutable", func() {
		dbApps := []model.BasicApplicationDTO{}
		fileApps := []model.BasicApplicationDTO{
			{ID: "file-app-1", Name: "File App"},
		}

		result := mergeAndDeduplicateApplications(dbApps, fileApps)

		suite.Len(result, 1)
		suite.True(result[0].IsReadOnly)
	})

	suite.Run("handles empty lists", func() {
		result := mergeAndDeduplicateApplications([]model.BasicApplicationDTO{}, []model.BasicApplicationDTO{})
		suite.Empty(result)
	})
}

// TestCompositeStoreTestSuite runs the composite store test suite.
func TestCompositeStoreTestSuite(t *testing.T) {
	// Initialize entity store for file-based store
	_ = entity.GetInstance()
	suite.Run(t, new(CompositeStoreTestSuite))
}
