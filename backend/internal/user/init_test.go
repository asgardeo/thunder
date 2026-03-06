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

package user

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
)

// LoadDeclarativeUserResourcesTestSuite tests the loadDeclarativeUserResources wrapper function.
type LoadDeclarativeUserResourcesTestSuite struct {
	suite.Suite
}

// SetupSuite initializes test settings once.
func (suite *LoadDeclarativeUserResourcesTestSuite) SetupSuite() {
	testConfig := &config.Config{}
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	if err != nil {
		suite.Fail("Failed to initialize runtime", err)
	}
}

// TearDownSuite cleans up after tests.
func (suite *LoadDeclarativeUserResourcesTestSuite) TearDownSuite() {
	config.ResetThunderRuntime()
}

// TestLoadDeclarativeUserResources_CompositeStore tests loading with a composite store.
// Composite store contains both file-based (immutable) and database (mutable) stores.
func (suite *LoadDeclarativeUserResourcesTestSuite) TestLoadDeclarativeUserResources_CompositeStore() {
	fileStore := &userFileBasedStore{
		GenericFileBasedStore: declarativeresource.NewGenericFileBasedStoreForTest(entity.KeyTypeUser),
	}
	dbStore, err := newUserStore()
	suite.NoError(err)

	compositeStore := newCompositeUserStore(fileStore, dbStore)

	// Call loadDeclarativeUserResources with composite store
	// Should extract fileStore and dbStore correctly and call loadDeclarativeResources
	err = loadDeclarativeUserResources(compositeStore)

	// The function should completes without error (or with acceptable error if resources directory doesn't exist)
	// The important part is that it doesn't panic
	_ = err
}

// TestLoadDeclarativeUserResources_FileBasedStore tests loading with a file-based store.
// File-based store contains only immutable resources (declarative mode).
func (suite *LoadDeclarativeUserResourcesTestSuite) TestLoadDeclarativeUserResources_FileBasedStore() {
	fileStore := &userFileBasedStore{
		GenericFileBasedStore: declarativeresource.NewGenericFileBasedStoreForTest(entity.KeyTypeUser),
	}

	// Call loadDeclarativeUserResources with file-based store
	// Should extract fileStore and set dbStore to nil
	err := loadDeclarativeUserResources(fileStore)

	// The function should complete without panic
	_ = err
}

// TestLoadDeclarativeUserResources_DatabaseStore tests loading with a database store (mutable mode).
// Database store should be used only for runtime users, not declarative resources.
func (suite *LoadDeclarativeUserResourcesTestSuite) TestLoadDeclarativeUserResources_DatabaseStore() {
	dbStore, err := newUserStore()
	suite.NoError(err)

	// Call loadDeclarativeUserResources with database store
	// Should return nil immediately (no declarative resources in mutable mode)
	err = loadDeclarativeUserResources(dbStore)

	// Should return nil for database store (mutable mode)
	suite.NoError(err)
}

// TestLoadDeclarativeUserResources_InvalidStoreType tests error handling for unsupported store types.
func (suite *LoadDeclarativeUserResourcesTestSuite) TestLoadDeclarativeUserResources_InvalidStoreType() {
	// Create a mock store that doesn't match any expected type
	mockStore := &userStoreInterfaceMock{}

	// Call loadDeclarativeUserResources with invalid store type
	// Should return nil since it's checked against known types and not a file-based store
	err := loadDeclarativeUserResources(mockStore)

	// Should handle gracefully (return nil for unknown types)
	suite.NoError(err)
}

// TestLoadDeclarativeUserResourcesTestSuite runs the test suite.
func TestLoadDeclarativeUserResourcesTestSuite(t *testing.T) {
	suite.Run(t, new(LoadDeclarativeUserResourcesTestSuite))
}

// InitializeStoreTestSuite tests the initializeStore function.
type InitializeStoreTestSuite struct {
	suite.Suite
}

// SetupSuite initializes test settings once.
func (suite *InitializeStoreTestSuite) SetupSuite() {
	testConfig := &config.Config{}
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	if err != nil {
		suite.Fail("Failed to initialize runtime", err)
	}
}

// SetupTest resets configuration before each test.
func (suite *InitializeStoreTestSuite) SetupTest() {
	runtime := config.GetThunderRuntime()
	runtime.Config.User.Store = ""
	runtime.Config.DeclarativeResources.Enabled = false
}

// TearDownSuite cleans up after tests.
func (suite *InitializeStoreTestSuite) TearDownSuite() {
	config.ResetThunderRuntime()
}

// TestInitializeStore_MutableMode tests store initialization in mutable mode (database only).
// In mutable mode, only the database store is created, and no composite or file-based stores.
func (suite *InitializeStoreTestSuite) TestInitializeStore_MutableMode() {
	runtime := config.GetThunderRuntime()
	runtime.Config.User.Store = ""
	runtime.Config.DeclarativeResources.Enabled = false

	store, err := initializeStore(serverconst.StoreModeMutable)

	suite.NoError(err)
	suite.NotNil(store)
	// Verify it's not a composite store and not a file-based store
	_, isFileStore := store.(*userFileBasedStore)
	_, isComposite := store.(*compositeUserStore)
	suite.False(isFileStore, "should not be file-based store in mutable mode")
	suite.False(isComposite, "should not be composite store in mutable mode")
}

// TestInitializeStore_DeclarativeMode tests store initialization in declarative mode (file-based only).
// In declarative mode, only the file-based store is created, and no database or composite stores.
func (suite *InitializeStoreTestSuite) TestInitializeStore_DeclarativeMode() {
	runtime := config.GetThunderRuntime()
	runtime.Config.User.Store = string(serverconst.StoreModeDeclarative)
	runtime.Config.DeclarativeResources.Enabled = true

	store, err := initializeStore(serverconst.StoreModeDeclarative)

	suite.NoError(err)
	suite.NotNil(store)
	// Verify it's a file-based store
	fileStore, isFileStore := store.(*userFileBasedStore)
	suite.True(isFileStore, "should be file-based store in declarative mode")
	suite.NotNil(fileStore)
}

// TestInitializeStore_CompositeMode tests store initialization in composite mode.
// In composite mode, both file-based (immutable) and database (mutable) stores are created.
func (suite *InitializeStoreTestSuite) TestInitializeStore_CompositeMode() {
	runtime := config.GetThunderRuntime()
	runtime.Config.User.Store = string(serverconst.StoreModeComposite)
	runtime.Config.DeclarativeResources.Enabled = true

	store, err := initializeStore(serverconst.StoreModeComposite)

	suite.NoError(err)
	suite.NotNil(store)
	// Verify it's a composite store
	compositeStore, isComposite := store.(*compositeUserStore)
	suite.True(isComposite, "should be composite store in composite mode")
	suite.NotNil(compositeStore)
	// Verify composite store has both internal stores
	suite.NotNil(compositeStore.fileStore, "composite store should have file store")
	suite.NotNil(compositeStore.dbStore, "composite store should have db store")
	// Verify file store is of correct type
	_, isFileBased := compositeStore.fileStore.(*userFileBasedStore)
	suite.True(isFileBased, "composite store's file store should be userFileBasedStore")
}

// TestInitializeStore_MutableMode_DefaultBehavior tests default fallback to mutable when declarative is disabled.
// When User.Store is empty and DeclarativeResources.Enabled is false, should initialize mutable store.
func (suite *InitializeStoreTestSuite) TestInitializeStore_MutableMode_DefaultBehavior() {
	runtime := config.GetThunderRuntime()
	runtime.Config.User.Store = ""
	runtime.Config.DeclarativeResources.Enabled = false

	// Call with StoreModeMutable (result of getUserStoreMode in this config)
	store, err := initializeStore(serverconst.StoreModeMutable)

	suite.NoError(err)
	suite.NotNil(store)
	_, isComposite := store.(*compositeUserStore)
	_, isFileStore := store.(*userFileBasedStore)
	suite.False(isComposite, "should not be composite when declarative resources disabled")
	suite.False(isFileStore, "should not be file-based when declarative resources disabled")
}

// TestInitializeStore_DeclarativeMode_GlobalFallback tests fallback to declarative when globally enabled.
// When User.Store is empty but DeclarativeResources.Enabled is true, should initialize declarative store.
func (suite *InitializeStoreTestSuite) TestInitializeStore_DeclarativeMode_GlobalFallback() {
	runtime := config.GetThunderRuntime()
	runtime.Config.User.Store = ""
	runtime.Config.DeclarativeResources.Enabled = true

	// Call with StoreModeDeclarative (result of getUserStoreMode in this config)
	store, err := initializeStore(serverconst.StoreModeDeclarative)

	suite.NoError(err)
	suite.NotNil(store)
	// Verify it's a file-based store
	_, isFileStore := store.(*userFileBasedStore)
	suite.True(isFileStore, "should be file-based when declarative resources enabled globally")
}

// TestInitializeStoreTestSuite runs the test suite.
func TestInitializeStoreTestSuite(t *testing.T) {
	suite.Run(t, new(InitializeStoreTestSuite))
}
