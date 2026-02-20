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

package idp

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CompositeIDPStoreTestSuite struct {
	suite.Suite
	fileStore      *idpStoreInterfaceMock
	dbStore        *idpStoreInterfaceMock
	compositeStore *compositeIDPStore
}

const (
	compositeStoreTestIDPName = "Test IDP"
	compositeStoreTestIDPID   = "idp-123"
)

func TestCompositeIDPStoreTestSuite(t *testing.T) {
	suite.Run(t, new(CompositeIDPStoreTestSuite))
}

func (s *CompositeIDPStoreTestSuite) SetupTest() {
	s.fileStore = newIdpStoreInterfaceMock(s.T())
	s.dbStore = newIdpStoreInterfaceMock(s.T())
	s.compositeStore = newCompositeIDPStore(s.fileStore, s.dbStore)
}

// TestCreateIdentityProvider_CreatesInDBStoreOnly verifies that Create only goes to DB store
func (s *CompositeIDPStoreTestSuite) TestCreateIdentityProvider_CreatesInDBStoreOnly() {
	idp := IDPDTO{
		ID:          "test-id",
		Name:        compositeStoreTestIDPName,
		Description: "Test Description",
		Type:        IDPTypeOIDC,
	}

	s.dbStore.On("CreateIdentityProvider", idp).Return(nil)

	err := s.compositeStore.CreateIdentityProvider(idp)

	s.NoError(err)
	s.dbStore.AssertCalled(s.T(), "CreateIdentityProvider", idp)
	s.fileStore.AssertNotCalled(s.T(), "CreateIdentityProvider")
}

// TestGetIdentityProvider_ReturnsFromDBFirst verifies DB store is queried first
func (s *CompositeIDPStoreTestSuite) TestGetIdentityProvider_ReturnsFromDBFirst() {
	idpID := compositeStoreTestIDPID
	dbIDP := &IDPDTO{
		ID:   idpID,
		Name: "DB IDP",
		Type: IDPTypeOIDC,
	}

	s.dbStore.On("GetIdentityProvider", idpID).Return(dbIDP, nil)

	result, err := s.compositeStore.GetIdentityProvider(idpID)

	s.NoError(err)
	s.Equal(dbIDP, result)
	s.dbStore.AssertCalled(s.T(), "GetIdentityProvider", idpID)
	// File store should not be called since DB returned successfully
	s.fileStore.AssertNotCalled(s.T(), "GetIdentityProvider")
}

// TestGetIdentityProvider_FallsBackToFileStore verifies fallback when DB store fails
func (s *CompositeIDPStoreTestSuite) TestGetIdentityProvider_FallsBackToFileStore() {
	idpID := compositeStoreTestIDPID
	fileIDP := &IDPDTO{
		ID:   idpID,
		Name: "File IDP",
		Type: IDPTypeOIDC,
	}

	s.dbStore.On("GetIdentityProvider", idpID).Return((*IDPDTO)(nil), ErrIDPNotFound)
	s.fileStore.On("GetIdentityProvider", idpID).Return(fileIDP, nil)

	result, err := s.compositeStore.GetIdentityProvider(idpID)

	s.NoError(err)
	s.Equal(fileIDP, result)
	s.dbStore.AssertCalled(s.T(), "GetIdentityProvider", idpID)
	s.fileStore.AssertCalled(s.T(), "GetIdentityProvider", idpID)
}

// TestGetIdentityProvider_ReturnsErrorWhenNotInBothStores verifies error when not found
func (s *CompositeIDPStoreTestSuite) TestGetIdentityProvider_ReturnsErrorWhenNotInBothStores() {
	idpID := compositeStoreTestIDPID

	s.dbStore.On("GetIdentityProvider", idpID).Return((*IDPDTO)(nil), ErrIDPNotFound)
	s.fileStore.On("GetIdentityProvider", idpID).Return((*IDPDTO)(nil), ErrIDPNotFound)

	result, err := s.compositeStore.GetIdentityProvider(idpID)

	s.Error(err)
	s.Nil(result)
	s.True(errors.Is(err, ErrIDPNotFound))
}

// TestGetIdentityProviderByName_ReturnsFromDBFirst verifies DB store is queried first for name lookup
func (s *CompositeIDPStoreTestSuite) TestGetIdentityProviderByName_ReturnsFromDBFirst() {
	idpName := compositeStoreTestIDPName
	dbIDP := &IDPDTO{
		ID:   "db-123",
		Name: idpName,
		Type: IDPTypeOIDC,
	}

	s.dbStore.On("GetIdentityProviderByName", idpName).Return(dbIDP, nil)

	result, err := s.compositeStore.GetIdentityProviderByName(idpName)

	s.NoError(err)
	s.Equal(dbIDP, result)
	s.dbStore.AssertCalled(s.T(), "GetIdentityProviderByName", idpName)
	s.fileStore.AssertNotCalled(s.T(), "GetIdentityProviderByName")
}

// TestGetIdentityProviderByName_FallsBackToFileStore verifies fallback for name lookup
func (s *CompositeIDPStoreTestSuite) TestGetIdentityProviderByName_FallsBackToFileStore() {
	idpName := compositeStoreTestIDPName
	fileIDP := &IDPDTO{
		ID:   "file-123",
		Name: idpName,
		Type: IDPTypeOIDC,
	}

	s.dbStore.On("GetIdentityProviderByName", idpName).Return((*IDPDTO)(nil), ErrIDPNotFound)
	s.fileStore.On("GetIdentityProviderByName", idpName).Return(fileIDP, nil)

	result, err := s.compositeStore.GetIdentityProviderByName(idpName)

	s.NoError(err)
	s.Equal(fileIDP, result)
}

// TestGetIdentityProviderList_MergesAndDeduplicates verifies list merging from both stores
func (s *CompositeIDPStoreTestSuite) TestGetIdentityProviderList_MergesAndDeduplicates() {
	dbIDPs := []BasicIDPDTO{
		{ID: "db-1", Name: "DB IDP 1", Type: IDPTypeOIDC},
		{ID: "db-2", Name: "DB IDP 2", Type: IDPTypeOIDC},
	}
	fileIDPs := []BasicIDPDTO{
		{ID: "file-1", Name: "File IDP 1", Type: IDPTypeOIDC},
		{ID: "file-2", Name: "File IDP 2", Type: IDPTypeOIDC},
	}

	s.dbStore.On("GetIdentityProviderList").Return(dbIDPs, nil)
	s.fileStore.On("GetIdentityProviderList").Return(fileIDPs, nil)

	result, err := s.compositeStore.GetIdentityProviderList()

	s.NoError(err)
	s.Len(result, 4)
	resultByID := make(map[string]BasicIDPDTO, len(result))
	for _, idp := range result {
		resultByID[idp.ID] = idp
	}

	// Verify DB IDPs are present and marked as mutable
	for _, expectedID := range []string{"db-1", "db-2"} {
		idp, ok := resultByID[expectedID]
		s.True(ok, "Expected DB IDP missing", expectedID)
		if ok {
			s.False(idp.IsReadOnly, "DB IDP should be marked as mutable")
		}
	}

	// Verify file IDPs are present and marked as immutable
	for _, expectedID := range []string{"file-1", "file-2"} {
		idp, ok := resultByID[expectedID]
		s.True(ok, "Expected file IDP missing", expectedID)
		if ok {
			s.True(idp.IsReadOnly, "File IDP should be marked as immutable")
		}
	}
}

// TestGetIdentityProviderList_DeduplicatesOnIDConflict verifies deduplication of IDs
func (s *CompositeIDPStoreTestSuite) TestGetIdentityProviderList_DeduplicatesOnIDConflict() {
	dbIDPs := []BasicIDPDTO{
		{ID: "shared-id", Name: "DB IDP", Type: IDPTypeOIDC},
		{ID: "db-2", Name: "DB IDP 2", Type: IDPTypeOIDC},
	}
	fileIDPs := []BasicIDPDTO{
		{ID: "shared-id", Name: "File IDP", Type: IDPTypeOIDC},
		{ID: "file-2", Name: "File IDP 2", Type: IDPTypeOIDC},
	}

	s.dbStore.On("GetIdentityProviderList").Return(dbIDPs, nil)
	s.fileStore.On("GetIdentityProviderList").Return(fileIDPs, nil)

	result, err := s.compositeStore.GetIdentityProviderList()

	s.NoError(err)
	s.Len(result, 3) // Only 3 unique IDs
	resultByID := make(map[string]BasicIDPDTO, len(result))
	for _, idp := range result {
		resultByID[idp.ID] = idp
	}

	// Verify shared ID takes DB version (first added)
	sharedIDP, ok := resultByID["shared-id"]
	s.True(ok, "Expected shared IDP missing")
	if ok {
		s.Equal("DB IDP", sharedIDP.Name)
		s.False(sharedIDP.IsReadOnly)
	}

	dbIDP, ok := resultByID["db-2"]
	s.True(ok, "Expected DB IDP missing")
	if ok {
		s.False(dbIDP.IsReadOnly, "DB IDP should be marked as mutable")
	}

	fileIDP, ok := resultByID["file-2"]
	s.True(ok, "Expected file IDP missing")
	if ok {
		s.True(fileIDP.IsReadOnly, "File IDP should be marked as immutable")
	}
}

// TestGetIdentityProviderList_HandlesEmptyStores verifies empty results are handled
func (s *CompositeIDPStoreTestSuite) TestGetIdentityProviderList_HandlesEmptyStores() {
	s.dbStore.On("GetIdentityProviderList").Return([]BasicIDPDTO{}, nil)
	s.fileStore.On("GetIdentityProviderList").Return([]BasicIDPDTO{}, nil)

	result, err := s.compositeStore.GetIdentityProviderList()

	s.NoError(err)
	s.Len(result, 0)
}

// TestGetIdentityProviderList_HandlesDBStoreError verifies error handling
func (s *CompositeIDPStoreTestSuite) TestGetIdentityProviderList_HandlesDBStoreError() {
	s.dbStore.On("GetIdentityProviderList").Return([]BasicIDPDTO{}, errors.New("DB error"))

	result, err := s.compositeStore.GetIdentityProviderList()

	s.Error(err)
	s.Nil(result)
}

// TestGetIdentityProviderList_HandlesFileStoreError verifies error handling
func (s *CompositeIDPStoreTestSuite) TestGetIdentityProviderList_HandlesFileStoreError() {
	dbIDPs := []BasicIDPDTO{
		{ID: "db-1", Name: "DB IDP 1", Type: IDPTypeOIDC},
	}
	s.dbStore.On("GetIdentityProviderList").Return(dbIDPs, nil)
	s.fileStore.On("GetIdentityProviderList").Return([]BasicIDPDTO{}, errors.New("File error"))

	result, err := s.compositeStore.GetIdentityProviderList()

	s.Error(err)
	s.Nil(result)
}

// TestUpdateIdentityProvider_UpdatesInDBStoreOnly verifies Update only goes to DB store
func (s *CompositeIDPStoreTestSuite) TestUpdateIdentityProvider_UpdatesInDBStoreOnly() {
	idp := &IDPDTO{
		ID:          "idp-123",
		Name:        "Updated IDP",
		Description: "Updated Description",
		Type:        IDPTypeOIDC,
	}

	s.dbStore.On("UpdateIdentityProvider", idp).Return(nil)

	err := s.compositeStore.UpdateIdentityProvider(idp)

	s.NoError(err)
	s.dbStore.AssertCalled(s.T(), "UpdateIdentityProvider", idp)
	s.fileStore.AssertNotCalled(s.T(), "UpdateIdentityProvider")
}

// TestDeleteIdentityProvider_DeletesFromDBStoreOnly verifies Delete only goes to DB store
func (s *CompositeIDPStoreTestSuite) TestDeleteIdentityProvider_DeletesFromDBStoreOnly() {
	idpID := "idp-123"

	s.dbStore.On("DeleteIdentityProvider", idpID).Return(nil)

	err := s.compositeStore.DeleteIdentityProvider(idpID)

	s.NoError(err)
	s.dbStore.AssertCalled(s.T(), "DeleteIdentityProvider", idpID)
	s.fileStore.AssertNotCalled(s.T(), "DeleteIdentityProvider")
}

// TestMergeAndDeduplicateIDPs_CorrectlyMarksReadOnlyFlags verifies read-only flag assignment
func (s *CompositeIDPStoreTestSuite) TestMergeAndDeduplicateIDPs_CorrectlyMarksReadOnlyFlags() {
	dbIDPs := []BasicIDPDTO{
		{ID: "db-1", Name: "DB IDP", Type: IDPTypeOIDC},
	}
	fileIDPs := []BasicIDPDTO{
		{ID: "file-1", Name: "File IDP", Type: IDPTypeOIDC},
	}

	result := mergeAndDeduplicateIDPs(dbIDPs, fileIDPs)

	s.Len(result, 2)

	// Find and verify DB IDP
	for _, idp := range result {
		if idp.ID == "db-1" {
			s.False(idp.IsReadOnly, "DB IDP should be mutable")
		}
		if idp.ID == "file-1" {
			s.True(idp.IsReadOnly, "File IDP should be immutable")
		}
	}
}

// TestMergeAndDeduplicateIDPs_PreservesDuplicatesPreference verifies DB precedence over file
func (s *CompositeIDPStoreTestSuite) TestMergeAndDeduplicateIDPs_PreservesDuplicatesPreference() {
	dbIDPs := []BasicIDPDTO{
		{ID: "shared", Name: "DB Name", Type: IDPTypeOIDC},
	}
	fileIDPs := []BasicIDPDTO{
		{ID: "shared", Name: "File Name", Type: IDPTypeOIDC},
	}

	result := mergeAndDeduplicateIDPs(dbIDPs, fileIDPs)

	s.Len(result, 1)
	s.Equal("DB Name", result[0].Name)
	s.False(result[0].IsReadOnly)
}
