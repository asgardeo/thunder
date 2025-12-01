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

package branding

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/tests/mocks/database/clientmock"
	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
)

type BrandingStoreTestSuite struct {
	suite.Suite
	mockDBProvider *providermock.DBProviderInterfaceMock
	mockDBClient   *clientmock.DBClientInterfaceMock
	store          *brandingStore
}

func TestBrandingStoreTestSuite(t *testing.T) {
	suite.Run(t, new(BrandingStoreTestSuite))
}

func (suite *BrandingStoreTestSuite) SetupTest() {
	suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
	suite.mockDBClient = clientmock.NewDBClientInterfaceMock(suite.T())
	suite.store = &brandingStore{
		dbProvider:   suite.mockDBProvider,
		deploymentID: "test-deployment-id",
	}
}

// GetBrandingListCount Tests
func (suite *BrandingStoreTestSuite) TestGetBrandingListCount_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetBrandingListCount, mock.Anything).Return([]map[string]interface{}{
		{"total": int64(10)},
	}, nil)

	count, err := suite.store.GetBrandingListCount()

	suite.NoError(err)
	suite.Equal(10, count)
}

func (suite *BrandingStoreTestSuite) TestGetBrandingListCount_EmptyResult() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetBrandingListCount, mock.Anything).Return([]map[string]interface{}{}, nil)

	count, err := suite.store.GetBrandingListCount()

	suite.NoError(err)
	suite.Equal(0, count)
}

func (suite *BrandingStoreTestSuite) TestGetBrandingListCount_DBClientError() {
	dbError := errors.New("database connection error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	count, err := suite.store.GetBrandingListCount()

	suite.Error(err)
	suite.Equal(0, count)
	suite.Contains(err.Error(), "failed to get database client")
}

func (suite *BrandingStoreTestSuite) TestGetBrandingListCount_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetBrandingListCount, mock.Anything).Return(nil, queryError)

	count, err := suite.store.GetBrandingListCount()

	suite.Error(err)
	suite.Equal(0, count)
	suite.Contains(err.Error(), "failed to execute count query")
}

func (suite *BrandingStoreTestSuite) TestGetBrandingListCount_InvalidCountResult() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetBrandingListCount, mock.Anything).Return([]map[string]interface{}{
		{"invalid": "value"},
	}, nil)

	count, err := suite.store.GetBrandingListCount()

	suite.Error(err)
	suite.Equal(0, count)
	suite.Contains(err.Error(), "failed to parse total/count from query result")
}

// GetBrandingList Tests
func (suite *BrandingStoreTestSuite) TestGetBrandingList_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetBrandingList, 10, 0, mock.Anything).Return([]map[string]interface{}{
		{
			"branding_id":  "brand1",
			"display_name": "Application 1 Branding",
		},
		{
			"branding_id":  "brand2",
			"display_name": "Application 2 Branding",
		},
	}, nil)

	brandings, err := suite.store.GetBrandingList(10, 0)

	suite.NoError(err)
	suite.Len(brandings, 2)
	suite.Equal("brand1", brandings[0].ID)
	suite.Equal("Application 1 Branding", brandings[0].DisplayName)
	suite.Equal("brand2", brandings[1].ID)
	suite.Equal("Application 2 Branding", brandings[1].DisplayName)
}

func (suite *BrandingStoreTestSuite) TestGetBrandingList_EmptyResult() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetBrandingList, 10, 0, mock.Anything).Return([]map[string]interface{}{}, nil)

	brandings, err := suite.store.GetBrandingList(10, 0)

	suite.NoError(err)
	suite.Len(brandings, 0)
}

func (suite *BrandingStoreTestSuite) TestGetBrandingList_DBClientError() {
	dbError := errors.New("database connection error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	brandings, err := suite.store.GetBrandingList(10, 0)

	suite.Error(err)
	suite.Nil(brandings)
	suite.Contains(err.Error(), "failed to get database client")
}

func (suite *BrandingStoreTestSuite) TestGetBrandingList_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetBrandingList, 10, 0, mock.Anything).Return(nil, queryError)

	brandings, err := suite.store.GetBrandingList(10, 0)

	suite.Error(err)
	suite.Nil(brandings)
	suite.Contains(err.Error(), "failed to execute branding list query")
}

func (suite *BrandingStoreTestSuite) TestGetBrandingList_InvalidBrandingID() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetBrandingList, 10, 0, mock.Anything).Return([]map[string]interface{}{
		{"branding_id": 123, "preferences": `{}`}, // Invalid type
	}, nil)

	brandings, err := suite.store.GetBrandingList(10, 0)

	suite.Error(err)
	suite.Nil(brandings)
	suite.Contains(err.Error(), "failed to build branding from result row")
}

func (suite *BrandingStoreTestSuite) TestGetBrandingList_PreferencesAsMap() {
	// Note: GetBrandingList now only returns id and displayName (no preferences)
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetBrandingList, 10, 0, mock.Anything).Return([]map[string]interface{}{
		{
			"branding_id":  "brand1",
			"display_name": "Application 1 Branding",
		},
	}, nil)

	brandings, err := suite.store.GetBrandingList(10, 0)

	suite.NoError(err)
	suite.Len(brandings, 1)
	suite.Equal("brand1", brandings[0].ID)
	suite.Equal("Application 1 Branding", brandings[0].DisplayName)
}

// CreateBranding Tests
func (suite *BrandingStoreTestSuite) TestCreateBranding_Success() {
	preferencesJSON := json.RawMessage(`{"theme":{"activeColorScheme":"dark"}}`)
	request := CreateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: preferencesJSON,
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateBranding, "brand1",
		"Application 1 Branding", mock.Anything, mock.Anything).Return(int64(1), nil)

	err := suite.store.CreateBranding("brand1", request)

	suite.NoError(err)
}

func (suite *BrandingStoreTestSuite) TestCreateBranding_DBClientError() {
	dbError := errors.New("database connection error")
	preferencesJSON := json.RawMessage(`{}`)
	request := CreateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: preferencesJSON,
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	err := suite.store.CreateBranding("brand1", request)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to get database client")
}

func (suite *BrandingStoreTestSuite) TestCreateBranding_ExecuteError() {
	preferencesJSON := json.RawMessage(`{}`)
	request := CreateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: preferencesJSON,
	}
	executeError := errors.New("execute error")

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryCreateBranding, "brand1",
		"Application 1 Branding", mock.Anything, mock.Anything).Return(int64(0), executeError)

	err := suite.store.CreateBranding("brand1", request)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to execute query")
}

// GetBranding Tests
func (suite *BrandingStoreTestSuite) TestGetBranding_Success() {
	preferencesJSON := json.RawMessage(`{"theme":{"activeColorScheme":"dark"}}`)
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetBrandingByID, "brand1", mock.Anything).Return([]map[string]interface{}{
		{
			"branding_id":  "brand1",
			"display_name": "Application 1 Branding",
			"preferences":  string(preferencesJSON),
		},
	}, nil)

	branding, err := suite.store.GetBranding("brand1")

	suite.NoError(err)
	suite.Equal("brand1", branding.ID)
	suite.Equal("Application 1 Branding", branding.DisplayName)
}

func (suite *BrandingStoreTestSuite) TestGetBranding_NotFound() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetBrandingByID, "brand1", mock.Anything).
		Return([]map[string]interface{}{}, nil)

	branding, err := suite.store.GetBranding("brand1")

	suite.Error(err)
	suite.Equal(Branding{}, branding)
	suite.Equal(ErrBrandingNotFound, err)
}

func (suite *BrandingStoreTestSuite) TestGetBranding_MultipleResults() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetBrandingByID, "brand1", mock.Anything).Return([]map[string]interface{}{
		{"branding_id": "brand1", "display_name": "Application 1 Branding", "preferences": `{}`},
		{"branding_id": "brand2", "display_name": "Application 2 Branding", "preferences": `{}`},
	}, nil)

	branding, err := suite.store.GetBranding("brand1")

	suite.Error(err)
	suite.Equal(Branding{}, branding)
	suite.Contains(err.Error(), "unexpected number of results")
}

func (suite *BrandingStoreTestSuite) TestGetBranding_DBClientError() {
	dbError := errors.New("database connection error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	branding, err := suite.store.GetBranding("brand1")

	suite.Error(err)
	suite.Equal(Branding{}, branding)
	suite.Contains(err.Error(), "failed to get database client")
}

func (suite *BrandingStoreTestSuite) TestGetBranding_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetBrandingByID, "brand1", mock.Anything).Return(nil, queryError)

	branding, err := suite.store.GetBranding("brand1")

	suite.Error(err)
	suite.Equal(Branding{}, branding)
	suite.Contains(err.Error(), "failed to execute query")
}

// IsBrandingExist Tests
func (suite *BrandingStoreTestSuite) TestIsBrandingExist_True() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckBrandingExists, "brand1", mock.Anything).Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.IsBrandingExist("brand1")

	suite.NoError(err)
	suite.True(exists)
}

func (suite *BrandingStoreTestSuite) TestIsBrandingExist_False() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckBrandingExists, "brand1", mock.Anything).Return([]map[string]interface{}{
		{"count": int64(0)},
	}, nil)

	exists, err := suite.store.IsBrandingExist("brand1")

	suite.NoError(err)
	suite.False(exists)
}

func (suite *BrandingStoreTestSuite) TestIsBrandingExist_EmptyResult() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckBrandingExists, "brand1", mock.Anything).
		Return([]map[string]interface{}{}, nil)

	exists, err := suite.store.IsBrandingExist("brand1")

	suite.NoError(err)
	suite.False(exists)
}

func (suite *BrandingStoreTestSuite) TestIsBrandingExist_InvalidCount() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckBrandingExists, "brand1", mock.Anything).Return([]map[string]interface{}{
		{"invalid": "value"},
	}, nil)

	exists, err := suite.store.IsBrandingExist("brand1")

	suite.Error(err)
	suite.False(exists)
	suite.Contains(err.Error(), "failed to parse count from query result")
}

func (suite *BrandingStoreTestSuite) TestIsBrandingExist_DBClientError() {
	dbError := errors.New("database connection error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	exists, err := suite.store.IsBrandingExist("brand1")

	suite.Error(err)
	suite.False(exists)
	suite.Contains(err.Error(), "failed to get database client")
}

func (suite *BrandingStoreTestSuite) TestIsBrandingExist_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckBrandingExists, "brand1", mock.Anything).Return(nil, queryError)

	exists, err := suite.store.IsBrandingExist("brand1")

	suite.Error(err)
	suite.False(exists)
	suite.Contains(err.Error(), "failed to check branding existence")
}

// UpdateBranding Tests
func (suite *BrandingStoreTestSuite) TestUpdateBranding_Success() {
	preferencesJSON := json.RawMessage(`{"theme":{"activeColorScheme":"light"}}`)
	request := UpdateBrandingRequest{
		DisplayName: "Application 2 Branding",
		Preferences: preferencesJSON,
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateBranding, "Application 2 Branding",
		mock.Anything, "brand1", mock.Anything).Return(int64(1), nil)

	err := suite.store.UpdateBranding("brand1", request)

	suite.NoError(err)
}

func (suite *BrandingStoreTestSuite) TestUpdateBranding_NotFound() {
	preferencesJSON := json.RawMessage(`{}`)
	request := UpdateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: preferencesJSON,
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateBranding, "Application 1 Branding",
		mock.Anything, "brand1", mock.Anything).Return(int64(0), nil)

	err := suite.store.UpdateBranding("brand1", request)

	suite.Error(err)
	suite.Equal(ErrBrandingNotFound, err)
}

func (suite *BrandingStoreTestSuite) TestUpdateBranding_DBClientError() {
	dbError := errors.New("database connection error")
	preferencesJSON := json.RawMessage(`{}`)
	request := UpdateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: preferencesJSON,
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	err := suite.store.UpdateBranding("brand1", request)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to get database client")
}

func (suite *BrandingStoreTestSuite) TestUpdateBranding_ExecuteError() {
	preferencesJSON := json.RawMessage(`{}`)
	request := UpdateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: preferencesJSON,
	}
	executeError := errors.New("execute error")

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryUpdateBranding, "Application 1 Branding",
		mock.Anything, "brand1", mock.Anything).Return(int64(0), executeError)

	err := suite.store.UpdateBranding("brand1", request)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to execute query")
}

// DeleteBranding Tests
func (suite *BrandingStoreTestSuite) TestDeleteBranding_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteBranding, "brand1", mock.Anything).Return(int64(1), nil)

	err := suite.store.DeleteBranding("brand1")

	suite.NoError(err)
}

func (suite *BrandingStoreTestSuite) TestDeleteBranding_NotFound() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteBranding, "brand1", mock.Anything).Return(int64(0), nil)

	err := suite.store.DeleteBranding("brand1")

	suite.NoError(err) // DeleteBranding doesn't return error for not found, just logs
}

func (suite *BrandingStoreTestSuite) TestDeleteBranding_DBClientError() {
	dbError := errors.New("database connection error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	err := suite.store.DeleteBranding("brand1")

	suite.Error(err)
	suite.Contains(err.Error(), "failed to get database client")
}

func (suite *BrandingStoreTestSuite) TestDeleteBranding_ExecuteError() {
	executeError := errors.New("execute error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteBranding, "brand1", mock.Anything).Return(int64(0), executeError)

	err := suite.store.DeleteBranding("brand1")

	suite.Error(err)
	suite.Contains(err.Error(), "failed to execute query")
}

// GetApplicationsCountByBrandingID Tests
func (suite *BrandingStoreTestSuite) TestGetApplicationsCountByBrandingID_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetApplicationsCountByBrandingID, "brand1", mock.Anything).
		Return([]map[string]interface{}{{"count": int64(3)}}, nil)

	count, err := suite.store.GetApplicationsCountByBrandingID("brand1")

	suite.NoError(err)
	suite.Equal(3, count)
}

func (suite *BrandingStoreTestSuite) TestGetApplicationsCountByBrandingID_Zero() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetApplicationsCountByBrandingID, "brand1", mock.Anything).
		Return([]map[string]interface{}{{"count": int64(0)}}, nil)

	count, err := suite.store.GetApplicationsCountByBrandingID("brand1")

	suite.NoError(err)
	suite.Equal(0, count)
}

func (suite *BrandingStoreTestSuite) TestGetApplicationsCountByBrandingID_EmptyResult() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetApplicationsCountByBrandingID, "brand1", mock.Anything).
		Return([]map[string]interface{}{}, nil)

	count, err := suite.store.GetApplicationsCountByBrandingID("brand1")

	suite.NoError(err)
	suite.Equal(0, count)
}

func (suite *BrandingStoreTestSuite) TestGetApplicationsCountByBrandingID_DBClientError() {
	dbError := errors.New("database connection error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	count, err := suite.store.GetApplicationsCountByBrandingID("brand1")

	suite.Error(err)
	suite.Equal(0, count)
	suite.Contains(err.Error(), "failed to get database client")
}

func (suite *BrandingStoreTestSuite) TestGetApplicationsCountByBrandingID_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetApplicationsCountByBrandingID, "brand1", mock.Anything).Return(nil, queryError)

	count, err := suite.store.GetApplicationsCountByBrandingID("brand1")

	suite.Error(err)
	suite.Equal(0, count)
	suite.Contains(err.Error(), "failed to get applications count")
}

// Helper function tests
func (suite *BrandingStoreTestSuite) TestParseCountResult_WithTotal() {
	results := []map[string]interface{}{
		{"total": int64(5)},
	}
	count, err := parseCountResult(results)
	suite.NoError(err)
	suite.Equal(5, count)
}

func (suite *BrandingStoreTestSuite) TestParseCountResult_WithCount() {
	results := []map[string]interface{}{
		{"count": int64(7)},
	}
	count, err := parseCountResult(results)
	suite.NoError(err)
	suite.Equal(7, count)
}

func (suite *BrandingStoreTestSuite) TestParseCountResult_Empty() {
	results := []map[string]interface{}{}
	count, err := parseCountResult(results)
	suite.NoError(err)
	suite.Equal(0, count)
}

func (suite *BrandingStoreTestSuite) TestParseBoolFromCount_True() {
	results := []map[string]interface{}{
		{"count": int64(1)},
	}
	exists, err := parseBoolFromCount(results)
	suite.NoError(err)
	suite.True(exists)
}

func (suite *BrandingStoreTestSuite) TestParseBoolFromCount_False() {
	results := []map[string]interface{}{
		{"count": int64(0)},
	}
	exists, err := parseBoolFromCount(results)
	suite.NoError(err)
	suite.False(exists)
}

func (suite *BrandingStoreTestSuite) TestParseBoolFromCount_Empty() {
	results := []map[string]interface{}{}
	exists, err := parseBoolFromCount(results)
	suite.NoError(err)
	suite.False(exists)
}

func (suite *BrandingStoreTestSuite) TestBuildBrandingFromResultRow_PreferencesAsBytes() {
	preferencesBytes := []byte(`{"theme":"dark"}`)
	row := map[string]interface{}{
		"branding_id":  "brand1",
		"display_name": "Application 1 Branding",
		"preferences":  preferencesBytes,
	}
	branding, err := buildBrandingFromResultRow(row)
	suite.NoError(err)
	suite.Equal("brand1", branding.ID)
	suite.Equal("Application 1 Branding", branding.DisplayName)
}

func (suite *BrandingStoreTestSuite) TestBuildBrandingFromResultRow_PreferencesAsMap() {
	preferencesMap := map[string]interface{}{"theme": "dark"}
	row := map[string]interface{}{
		"branding_id":  "brand1",
		"display_name": "Application 1 Branding",
		"preferences":  preferencesMap,
	}
	branding, err := buildBrandingFromResultRow(row)
	suite.NoError(err)
	suite.Equal("brand1", branding.ID)
	suite.Equal("Application 1 Branding", branding.DisplayName)
}

func (suite *BrandingStoreTestSuite) TestBuildBrandingFromResultRow_InvalidID() {
	row := map[string]interface{}{
		"branding_id":  123, // Invalid type
		"display_name": "Application 1 Branding",
		"preferences":  `{}`,
	}
	branding, err := buildBrandingFromResultRow(row)
	suite.Error(err)
	suite.Equal(Branding{}, branding)
	suite.Contains(err.Error(), "failed to parse branding_id as string")
}

func (suite *BrandingStoreTestSuite) TestBuildBrandingFromResultRow_NilPreferences() {
	row := map[string]interface{}{
		"branding_id":  "brand1",
		"display_name": "Application 1 Branding",
		"preferences":  nil,
	}
	branding, err := buildBrandingFromResultRow(row)
	suite.NoError(err)
	suite.Equal("brand1", branding.ID)
	suite.Equal("Application 1 Branding", branding.DisplayName)
	suite.Nil(branding.Preferences)
}

func (suite *BrandingStoreTestSuite) TestBuildBrandingFromResultRow_MarshalError() {
	// Create a preferences value that will fail to marshal
	// Using a channel which cannot be marshaled
	preferencesChan := make(chan int)
	row := map[string]interface{}{
		"branding_id":  "brand1",
		"display_name": "Application 1 Branding",
		"preferences":  preferencesChan,
	}
	branding, err := buildBrandingFromResultRow(row)
	suite.Error(err)
	suite.Equal(Branding{}, branding)
	suite.Contains(err.Error(), "failed to marshal preferences")
}
