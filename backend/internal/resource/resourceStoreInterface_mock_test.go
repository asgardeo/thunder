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

package resource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
)

// resourceStoreInterfaceMock is a testify mock for resourceStoreInterface.
type resourceStoreInterfaceMock struct {
	mock.Mock
}

func newResourceStoreInterfaceMock(t *testing.T) *resourceStoreInterfaceMock {
	m := &resourceStoreInterfaceMock{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

// Resource Server operations

func (m *resourceStoreInterfaceMock) CreateResourceServer(
	ctx context.Context, id string, rs ResourceServer,
) error {
	args := m.Called(ctx, id, rs)
	return args.Error(0)
}

func (m *resourceStoreInterfaceMock) GetResourceServer(
	ctx context.Context, id string,
) (ResourceServer, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(ResourceServer), args.Error(1)
}

func (m *resourceStoreInterfaceMock) GetResourceServerList(
	ctx context.Context, limit, offset int,
) ([]ResourceServer, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ResourceServer), args.Error(1)
}

func (m *resourceStoreInterfaceMock) GetResourceServerListCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *resourceStoreInterfaceMock) UpdateResourceServer(
	ctx context.Context, id string, rs ResourceServer,
) error {
	args := m.Called(ctx, id, rs)
	return args.Error(0)
}

func (m *resourceStoreInterfaceMock) DeleteResourceServer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *resourceStoreInterfaceMock) CheckResourceServerNameExists(
	ctx context.Context, name string,
) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *resourceStoreInterfaceMock) CheckResourceServerIdentifierExists(
	ctx context.Context, identifier string,
) (bool, error) {
	args := m.Called(ctx, identifier)
	return args.Bool(0), args.Error(1)
}

func (m *resourceStoreInterfaceMock) CheckResourceServerHasDependencies(
	ctx context.Context, resServerID string,
) (bool, error) {
	args := m.Called(ctx, resServerID)
	return args.Bool(0), args.Error(1)
}

// Resource operations

func (m *resourceStoreInterfaceMock) CreateResource(
	ctx context.Context, uuid string, resServerID string, parentID *string, res Resource,
) error {
	args := m.Called(ctx, uuid, resServerID, parentID, res)
	return args.Error(0)
}

func (m *resourceStoreInterfaceMock) GetResource(
	ctx context.Context, id string, resServerID string,
) (Resource, error) {
	args := m.Called(ctx, id, resServerID)
	return args.Get(0).(Resource), args.Error(1)
}

func (m *resourceStoreInterfaceMock) GetResourceList(
	ctx context.Context, resServerID string, limit, offset int,
) ([]Resource, error) {
	args := m.Called(ctx, resServerID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Resource), args.Error(1)
}

func (m *resourceStoreInterfaceMock) GetResourceListByParent(
	ctx context.Context, resServerID string, parentID *string, limit, offset int,
) ([]Resource, error) {
	args := m.Called(ctx, resServerID, parentID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Resource), args.Error(1)
}

func (m *resourceStoreInterfaceMock) GetResourceListCount(
	ctx context.Context, resServerID string,
) (int, error) {
	args := m.Called(ctx, resServerID)
	return args.Int(0), args.Error(1)
}

func (m *resourceStoreInterfaceMock) GetResourceListCountByParent(
	ctx context.Context, resServerID string, parentID *string,
) (int, error) {
	args := m.Called(ctx, resServerID, parentID)
	return args.Int(0), args.Error(1)
}

func (m *resourceStoreInterfaceMock) UpdateResource(
	ctx context.Context, id string, resServerID string, res Resource,
) error {
	args := m.Called(ctx, id, resServerID, res)
	return args.Error(0)
}

func (m *resourceStoreInterfaceMock) DeleteResource(
	ctx context.Context, id string, resServerID string,
) error {
	args := m.Called(ctx, id, resServerID)
	return args.Error(0)
}

func (m *resourceStoreInterfaceMock) CheckResourceHandleExists(
	ctx context.Context, resServerID string, handle string, parentID *string,
) (bool, error) {
	args := m.Called(ctx, resServerID, handle, parentID)
	return args.Bool(0), args.Error(1)
}

func (m *resourceStoreInterfaceMock) CheckResourceHasDependencies(
	ctx context.Context, resID string,
) (bool, error) {
	args := m.Called(ctx, resID)
	return args.Bool(0), args.Error(1)
}

func (m *resourceStoreInterfaceMock) CheckCircularDependency(
	ctx context.Context, resourceID, newParentID string,
) (bool, error) {
	args := m.Called(ctx, resourceID, newParentID)
	return args.Bool(0), args.Error(1)
}

// Action operations

func (m *resourceStoreInterfaceMock) CreateAction(
	ctx context.Context, uuid string, resServerID string, resID *string, action Action,
) error {
	args := m.Called(ctx, uuid, resServerID, resID, action)
	return args.Error(0)
}

func (m *resourceStoreInterfaceMock) GetAction(
	ctx context.Context, id string, resServerID string, resID *string,
) (Action, error) {
	args := m.Called(ctx, id, resServerID, resID)
	return args.Get(0).(Action), args.Error(1)
}

func (m *resourceStoreInterfaceMock) GetActionList(
	ctx context.Context, resServerID string, resID *string, limit, offset int,
) ([]Action, error) {
	args := m.Called(ctx, resServerID, resID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Action), args.Error(1)
}

func (m *resourceStoreInterfaceMock) GetActionListCount(
	ctx context.Context, resServerID string, resID *string,
) (int, error) {
	args := m.Called(ctx, resServerID, resID)
	return args.Int(0), args.Error(1)
}

func (m *resourceStoreInterfaceMock) UpdateAction(
	ctx context.Context, id string, resServerID string, resID *string, action Action,
) error {
	args := m.Called(ctx, id, resServerID, resID, action)
	return args.Error(0)
}

func (m *resourceStoreInterfaceMock) DeleteAction(
	ctx context.Context, id string, resServerID string, resID *string,
) error {
	args := m.Called(ctx, id, resServerID, resID)
	return args.Error(0)
}

func (m *resourceStoreInterfaceMock) IsActionExist(
	ctx context.Context, id string, resServerID string, resID *string,
) (bool, error) {
	args := m.Called(ctx, id, resServerID, resID)
	return args.Bool(0), args.Error(1)
}

func (m *resourceStoreInterfaceMock) CheckActionHandleExists(
	ctx context.Context, resServerID string, resID *string, handle string,
) (bool, error) {
	args := m.Called(ctx, resServerID, resID, handle)
	return args.Bool(0), args.Error(1)
}

func (m *resourceStoreInterfaceMock) ValidatePermissions(
	ctx context.Context, resServerID string, permissions []string,
) ([]string, error) {
	args := m.Called(ctx, resServerID, permissions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
