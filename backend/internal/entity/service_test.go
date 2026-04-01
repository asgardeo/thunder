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

package entity

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/transaction"
)

type ServiceTestSuite struct {
	suite.Suite
	store   *entityStoreInterfaceMock
	svc     EntityServiceInterface
	ctx     context.Context
	testErr error
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) SetupTest() {
	s.store = newEntityStoreInterfaceMock(s.T())
	s.svc = newEntityService(s.store, transaction.NewNoOpTransactioner())
	s.ctx = context.Background()
	s.testErr = errors.New("store error")
}

func testEntity(id string) *Entity {
	attrs, _ := json.Marshal(map[string]interface{}{"username": "user-" + id})
	return &Entity{
		ID:                 id,
		Category:           EntityCategoryUser,
		Type:               "employee",
		State:              EntityStateActive,
		OrganizationUnitID: "ou-1",
		Attributes:         json.RawMessage(attrs),
	}
}

// CreateEntity

func (s *ServiceTestSuite) TestCreateEntity_NilEntity() {
	_, err := s.svc.CreateEntity(s.ctx, nil, nil, nil)
	s.ErrorIs(err, ErrEntityNotFound)
}

func (s *ServiceTestSuite) TestCreateEntity_StoreCreateFails() {
	e := testEntity("e1")
	s.store.On("CreateEntity", mock.Anything, *e, json.RawMessage(nil), json.RawMessage(nil)).
		Return(s.testErr)
	_, err := s.svc.CreateEntity(s.ctx, e, nil, nil)
	s.Error(err)
}

func (s *ServiceTestSuite) TestCreateEntity_GetAfterCreateFails() {
	e := testEntity("e2")
	s.store.On("CreateEntity", mock.Anything, *e, json.RawMessage(nil), json.RawMessage(nil)).
		Return(nil)
	s.store.On("GetEntity", mock.Anything, e.ID).Return(Entity{}, s.testErr)
	_, err := s.svc.CreateEntity(s.ctx, e, nil, nil)
	s.Error(err)
}

func (s *ServiceTestSuite) TestCreateEntity_Success() {
	e := testEntity("e3")
	s.store.On("CreateEntity", mock.Anything, *e, json.RawMessage(nil), json.RawMessage(nil)).
		Return(nil)
	s.store.On("GetEntity", mock.Anything, e.ID).Return(*e, nil)
	got, err := s.svc.CreateEntity(s.ctx, e, nil, nil)
	s.NoError(err)
	s.Equal(e.ID, got.ID)
}

// GetEntity

func (s *ServiceTestSuite) TestGetEntity_Success() {
	e := testEntity("e4")
	s.store.On("GetEntity", mock.Anything, e.ID).Return(*e, nil)
	got, err := s.svc.GetEntity(s.ctx, e.ID)
	s.NoError(err)
	s.Equal(e.ID, got.ID)
}

func (s *ServiceTestSuite) TestGetEntity_Error() {
	s.store.On("GetEntity", mock.Anything, "bad").Return(Entity{}, s.testErr)
	_, err := s.svc.GetEntity(s.ctx, "bad")
	s.Error(err)
}

// UpdateEntity

func (s *ServiceTestSuite) TestUpdateEntity_NilEntity() {
	_, err := s.svc.UpdateEntity(s.ctx, "id", nil)
	s.ErrorIs(err, ErrEntityNotFound)
}

func (s *ServiceTestSuite) TestUpdateEntity_StoreFails() {
	e := testEntity("e5")
	s.store.On("UpdateEntity", mock.Anything, e).Return(s.testErr)
	_, err := s.svc.UpdateEntity(s.ctx, e.ID, e)
	s.Error(err)
}

func (s *ServiceTestSuite) TestUpdateEntity_GetAfterUpdateFails() {
	e := testEntity("e6")
	s.store.On("UpdateEntity", mock.Anything, e).Return(nil)
	s.store.On("GetEntity", mock.Anything, e.ID).Return(Entity{}, s.testErr)
	_, err := s.svc.UpdateEntity(s.ctx, e.ID, e)
	s.Error(err)
}

func (s *ServiceTestSuite) TestUpdateEntity_Success() {
	e := testEntity("e7")
	s.store.On("UpdateEntity", mock.Anything, e).Return(nil)
	s.store.On("GetEntity", mock.Anything, e.ID).Return(*e, nil)
	got, err := s.svc.UpdateEntity(s.ctx, e.ID, e)
	s.NoError(err)
	s.Equal(e.ID, got.ID)
}

// DeleteEntity

func (s *ServiceTestSuite) TestDeleteEntity_Delegates() {
	s.store.On("DeleteEntity", mock.Anything, "del1").Return(nil)
	s.NoError(s.svc.DeleteEntity(s.ctx, "del1"))
}

// UpdateSystemCredentials

func (s *ServiceTestSuite) TestUpdateSystemCredentials_Delegates() {
	creds := json.RawMessage(`{"token":"x"}`)
	s.store.On("UpdateSystemCredentials", mock.Anything, "e1", creds).Return(nil)
	s.NoError(s.svc.UpdateSystemCredentials(s.ctx, "e1", creds))
}

// GetEntityWithCredentials

func (s *ServiceTestSuite) TestGetEntityWithCredentials_Success() {
	e := testEntity("ecreds")
	creds := json.RawMessage(`{"password":"h"}`)
	sysCreds := json.RawMessage(`{"tok":"t"}`)
	s.store.On("GetEntityWithCredentials", mock.Anything, e.ID).Return(*e, creds, sysCreds, nil)
	gotE, gotC, gotS, err := s.svc.GetEntityWithCredentials(s.ctx, e.ID)
	s.NoError(err)
	s.Equal(e.ID, gotE.ID)
	s.Equal(string(creds), string(gotC))
	s.Equal(string(sysCreds), string(gotS))
}

func (s *ServiceTestSuite) TestGetEntityWithCredentials_Error() {
	s.store.On("GetEntityWithCredentials", mock.Anything, "bad").
		Return(Entity{}, json.RawMessage(nil), json.RawMessage(nil), s.testErr)
	_, _, _, err := s.svc.GetEntityWithCredentials(s.ctx, "bad")
	s.Error(err)
}

// IdentifyEntity

func (s *ServiceTestSuite) TestIdentifyEntity_Delegates() {
	filters := map[string]interface{}{"email": "x@y.com"}
	id := "found-id"
	s.store.On("IdentifyEntity", mock.Anything, filters).Return(&id, nil)
	got, err := s.svc.IdentifyEntity(s.ctx, filters)
	s.NoError(err)
	s.Equal(&id, got)
}

// GetEntityListCount

func (s *ServiceTestSuite) TestGetEntityListCount_Delegates() {
	s.store.On("GetEntityListCount", mock.Anything, "user", mock.Anything).Return(5, nil)
	count, err := s.svc.GetEntityListCount(s.ctx, EntityCategoryUser, nil)
	s.NoError(err)
	s.Equal(5, count)
}

// GetEntityList

func (s *ServiceTestSuite) TestGetEntityList_Delegates() {
	e := testEntity("le1")
	s.store.On("GetEntityList", mock.Anything, "user", 10, 0, mock.Anything).Return([]Entity{*e}, nil)
	list, err := s.svc.GetEntityList(s.ctx, EntityCategoryUser, 10, 0, nil)
	s.NoError(err)
	s.Len(list, 1)
}

// GetEntityListCountByOUIDs

func (s *ServiceTestSuite) TestGetEntityListCountByOUIDs_Delegates() {
	s.store.On("GetEntityListCountByOUIDs", mock.Anything, "user", []string{"ou1"}, mock.Anything).
		Return(3, nil)
	count, err := s.svc.GetEntityListCountByOUIDs(s.ctx, EntityCategoryUser, []string{"ou1"}, nil)
	s.NoError(err)
	s.Equal(3, count)
}

// GetEntityListByOUIDs

func (s *ServiceTestSuite) TestGetEntityListByOUIDs_Delegates() {
	e := testEntity("ou-e1")
	s.store.On("GetEntityListByOUIDs", mock.Anything, "user", []string{"ou1"}, 10, 0, mock.Anything).
		Return([]Entity{*e}, nil)
	list, err := s.svc.GetEntityListByOUIDs(s.ctx, EntityCategoryUser, []string{"ou1"}, 10, 0, nil)
	s.NoError(err)
	s.Len(list, 1)
}

// ValidateEntityIDs

func (s *ServiceTestSuite) TestValidateEntityIDs_Delegates() {
	s.store.On("ValidateEntityIDs", mock.Anything, []string{"id1", "id2"}).Return([]string{}, nil)
	invalid, err := s.svc.ValidateEntityIDs(s.ctx, []string{"id1", "id2"})
	s.NoError(err)
	s.Empty(invalid)
}

// GetEntitiesByIDs

func (s *ServiceTestSuite) TestGetEntitiesByIDs_Delegates() {
	e := testEntity("bid1")
	s.store.On("GetEntitiesByIDs", mock.Anything, []string{"bid1"}).Return([]Entity{*e}, nil)
	list, err := s.svc.GetEntitiesByIDs(s.ctx, []string{"bid1"})
	s.NoError(err)
	s.Len(list, 1)
}

// ValidateEntityIDsInOUs

func (s *ServiceTestSuite) TestValidateEntityIDsInOUs_Delegates() {
	s.store.On("ValidateEntityIDsInOUs", mock.Anything, []string{"id1"}, []string{"ou1"}).
		Return([]string{}, nil)
	out, err := s.svc.ValidateEntityIDsInOUs(s.ctx, []string{"id1"}, []string{"ou1"})
	s.NoError(err)
	s.Empty(out)
}

// GetGroupCountForEntity

func (s *ServiceTestSuite) TestGetGroupCountForEntity_Delegates() {
	s.store.On("GetGroupCountForEntity", mock.Anything, "e1").Return(2, nil)
	count, err := s.svc.GetGroupCountForEntity(s.ctx, "e1")
	s.NoError(err)
	s.Equal(2, count)
}

// GetEntityGroups

func (s *ServiceTestSuite) TestGetEntityGroups_Delegates() {
	groups := []EntityGroup{{ID: "g1", Name: "Group1", OUID: "ou1"}}
	s.store.On("GetEntityGroups", mock.Anything, "e1", 10, 0).Return(groups, nil)
	got, err := s.svc.GetEntityGroups(s.ctx, "e1", 10, 0)
	s.NoError(err)
	s.Len(got, 1)
}

// IsEntityDeclarative

func (s *ServiceTestSuite) TestIsEntityDeclarative_Delegates() {
	s.store.On("IsEntityDeclarative", mock.Anything, "e1").Return(true, nil)
	ok, err := s.svc.IsEntityDeclarative(s.ctx, "e1")
	s.NoError(err)
	s.True(ok)
}

// LoadDeclarativeResources

func (s *ServiceTestSuite) TestLoadDeclarativeResources_MutableStore_NoOp() {
	cfg := DeclarativeLoaderConfig{
		Directory: "users",
		Category:  EntityCategoryUser,
		Parser: func(data []byte) (*Entity, json.RawMessage, json.RawMessage, error) {
			return nil, nil, nil, nil
		},
	}
	// store is a mock (not file/composite) → fileStore == nil → returns nil immediately
	err := s.svc.LoadDeclarativeResources(cfg)
	s.NoError(err)
}

// UpdateEntityWithCredentials

func (s *ServiceTestSuite) TestUpdateEntityWithCredentials_NilEntity() {
	_, err := s.svc.UpdateEntityWithCredentials(s.ctx, "id", nil, nil)
	s.ErrorIs(err, ErrEntityNotFound)
}

func (s *ServiceTestSuite) TestUpdateEntityWithCredentials_UpdateFails() {
	e := testEntity("uc1")
	s.store.On("UpdateEntity", mock.Anything, e).Return(s.testErr)
	_, err := s.svc.UpdateEntityWithCredentials(s.ctx, e.ID, e, nil)
	s.Error(err)
}

func (s *ServiceTestSuite) TestUpdateEntityWithCredentials_WithCredsUpdateFails() {
	e := testEntity("uc2")
	sysCreds := json.RawMessage(`{"tok":"t"}`)
	s.store.On("UpdateEntity", mock.Anything, e).Return(nil)
	s.store.On("UpdateSystemCredentials", mock.Anything, e.ID, sysCreds).Return(s.testErr)
	_, err := s.svc.UpdateEntityWithCredentials(s.ctx, e.ID, e, sysCreds)
	s.Error(err)
}

func (s *ServiceTestSuite) TestUpdateEntityWithCredentials_GetAfterUpdateFails() {
	e := testEntity("uc3")
	s.store.On("UpdateEntity", mock.Anything, e).Return(nil)
	s.store.On("GetEntity", mock.Anything, e.ID).Return(Entity{}, s.testErr)
	_, err := s.svc.UpdateEntityWithCredentials(s.ctx, e.ID, e, nil)
	s.Error(err)
}

func (s *ServiceTestSuite) TestUpdateEntityWithCredentials_SuccessNoCreds() {
	e := testEntity("uc4")
	s.store.On("UpdateEntity", mock.Anything, e).Return(nil)
	s.store.On("GetEntity", mock.Anything, e.ID).Return(*e, nil)
	got, err := s.svc.UpdateEntityWithCredentials(s.ctx, e.ID, e, nil)
	s.NoError(err)
	s.Equal(e.ID, got.ID)
}

func (s *ServiceTestSuite) TestUpdateEntityWithCredentials_SuccessWithCreds() {
	e := testEntity("uc5")
	sysCreds := json.RawMessage(`{"tok":"t"}`)
	s.store.On("UpdateEntity", mock.Anything, e).Return(nil)
	s.store.On("UpdateSystemCredentials", mock.Anything, e.ID, sysCreds).Return(nil)
	s.store.On("GetEntity", mock.Anything, e.ID).Return(*e, nil)
	got, err := s.svc.UpdateEntityWithCredentials(s.ctx, e.ID, e, sysCreds)
	s.NoError(err)
	s.Equal(e.ID, got.ID)
}
