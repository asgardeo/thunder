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

package ou

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ---------------------------------------------------------------------------
// Test suite
// ---------------------------------------------------------------------------

type HierarchyResolverTestSuite struct {
	suite.Suite
}

func TestHierarchyResolverTestSuite(t *testing.T) {
	suite.Run(t, new(HierarchyResolverTestSuite))
}

// ---------------------------------------------------------------------------
// newOUHierarchyAdapter
// ---------------------------------------------------------------------------

func (suite *HierarchyResolverTestSuite) TestNewOUHierarchyAdapter_ReturnsNonNil() {
	mockStore := newOrganizationUnitStoreInterfaceMock(suite.T())
	resolver := newOUHierarchyAdapter(mockStore)
	assert.NotNil(suite.T(), resolver)
}

// ---------------------------------------------------------------------------
// IsAncestorOrSelf
// ---------------------------------------------------------------------------

func (suite *HierarchyResolverTestSuite) TestIsAncestorOrSelf() {
	parentID := testCoverageParentOUID
	genericErr := errors.New("database error")

	tests := []struct {
		name           string
		ancestorOUID   string
		descendantOUID string
		setupMock      func(m *organizationUnitStoreInterfaceMock)
		wantResult     bool
		wantErr        bool
	}{
		{
			name:           "EmptyAncestorOUID_ReturnsFalse",
			ancestorOUID:   "",
			descendantOUID: "child-ou",
			setupMock:      func(m *organizationUnitStoreInterfaceMock) {},
			wantResult:     false,
		},
		{
			name:           "EmptyDescendantOUID_ReturnsFalse",
			ancestorOUID:   testCoverageParentOUID,
			descendantOUID: "",
			setupMock:      func(m *organizationUnitStoreInterfaceMock) {},
			wantResult:     false,
		},
		{
			name:           "BothEmpty_ReturnsFalse",
			ancestorOUID:   "",
			descendantOUID: "",
			setupMock:      func(m *organizationUnitStoreInterfaceMock) {},
			wantResult:     false,
		},
		{
			name:           "SameOUID_ReturnsTrue",
			ancestorOUID:   "ou1",
			descendantOUID: "ou1",
			setupMock:      func(m *organizationUnitStoreInterfaceMock) {},
			wantResult:     true,
		},
		{
			name:           "DirectParent_ReturnsTrue",
			ancestorOUID:   testCoverageParentOUID,
			descendantOUID: "child-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				m.On("GetOrganizationUnit", "child-ou").
					Return(OrganizationUnit{ID: "child-ou", Parent: &parentID}, nil)
			},
			wantResult: true,
		},
		{
			name:           "Grandparent_ReturnsTrue",
			ancestorOUID:   "root-ou",
			descendantOUID: "grandchild-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				parentRef := testCoverageParentOUID
				rootRef := "root-ou"
				m.On("GetOrganizationUnit", "grandchild-ou").
					Return(OrganizationUnit{ID: "grandchild-ou", Parent: &parentRef}, nil)
				m.On("GetOrganizationUnit", testCoverageParentOUID).
					Return(OrganizationUnit{ID: testCoverageParentOUID, Parent: &rootRef}, nil)
			},
			wantResult: true,
		},
		{
			name:           "UnrelatedOU_ReturnsFalse",
			ancestorOUID:   "unrelated-ou",
			descendantOUID: "child-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				parentRef := testCoverageParentOUID
				m.On("GetOrganizationUnit", "child-ou").
					Return(OrganizationUnit{ID: "child-ou", Parent: &parentRef}, nil)
				// parent-ou is root (no parent).
				m.On("GetOrganizationUnit", testCoverageParentOUID).
					Return(OrganizationUnit{ID: testCoverageParentOUID, Parent: nil}, nil)
			},
			wantResult: false,
		},
		{
			name:           "RootOU_NoParent_ReturnsFalse",
			ancestorOUID:   "some-ou",
			descendantOUID: "root-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				m.On("GetOrganizationUnit", "root-ou").
					Return(OrganizationUnit{ID: "root-ou", Parent: nil}, nil)
			},
			wantResult: false,
		},
		{
			name:           "BrokenChain_OUNotFound_ReturnsFalseNoError",
			ancestorOUID:   testCoverageParentOUID,
			descendantOUID: "orphan-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				m.On("GetOrganizationUnit", "orphan-ou").
					Return(OrganizationUnit{}, ErrOrganizationUnitNotFound)
			},
			wantResult: false,
		},
		{
			name:           "StoreError_ReturnsFalseWithError",
			ancestorOUID:   testCoverageParentOUID,
			descendantOUID: "child-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				m.On("GetOrganizationUnit", "child-ou").
					Return(OrganizationUnit{}, genericErr)
			},
			wantResult: false,
			wantErr:    true,
		},
		{
			name:           "CyclicChain_ReturnsFalseNoError",
			ancestorOUID:   "root-ou",
			descendantOUID: "child-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				parentRef := testCoverageParentOUID
				childRef := "child-ou"
				m.On("GetOrganizationUnit", "child-ou").
					Return(OrganizationUnit{ID: "child-ou", Parent: &parentRef}, nil).Times(1)
				m.On("GetOrganizationUnit", testCoverageParentOUID).
					Return(OrganizationUnit{ID: testCoverageParentOUID, Parent: &childRef}, nil).Times(1)
			},
			wantResult: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			mockStore := newOrganizationUnitStoreInterfaceMock(suite.T())
			tt.setupMock(mockStore)
			resolver := newOUHierarchyAdapter(mockStore)

			result, svcErr := resolver.IsAncestorOrSelf(context.Background(), tt.ancestorOUID, tt.descendantOUID)
			assert.Equal(suite.T(), tt.wantResult, result)
			if tt.wantErr {
				assert.NotNil(suite.T(), svcErr)
			} else {
				assert.Nil(suite.T(), svcErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetAncestorOUIDs
// ---------------------------------------------------------------------------

func (suite *HierarchyResolverTestSuite) TestGetAncestorOUIDs() {
	genericErr := errors.New("database error")

	tests := []struct {
		name      string
		ouID      string
		setupMock func(m *organizationUnitStoreInterfaceMock)
		wantIDs   []string
		wantErr   bool
	}{
		{
			name:      "EmptyOUID_ReturnsEmptySlice",
			ouID:      "",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {},
			wantIDs:   []string{},
		},
		{
			name: "RootOU_NoParent_ReturnsSelf",
			ouID: "root-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				m.On("GetOrganizationUnit", "root-ou").
					Return(OrganizationUnit{ID: "root-ou", Parent: nil}, nil)
			},
			wantIDs: []string{"root-ou"},
		},
		{
			name: "ChildOU_ReturnsSelfAndParent",
			ouID: "child-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				parentRef := testCoverageParentOUID
				m.On("GetOrganizationUnit", "child-ou").
					Return(OrganizationUnit{ID: "child-ou", Parent: &parentRef}, nil)
				m.On("GetOrganizationUnit", testCoverageParentOUID).
					Return(OrganizationUnit{ID: testCoverageParentOUID, Parent: nil}, nil)
			},
			wantIDs: []string{"child-ou", testCoverageParentOUID},
		},
		{
			name: "ThreeLevelHierarchy_ReturnsSelfParentGrandparent",
			ouID: "grandchild-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				parentRef := testCoverageParentOUID
				rootRef := "root-ou"
				m.On("GetOrganizationUnit", "grandchild-ou").
					Return(OrganizationUnit{ID: "grandchild-ou", Parent: &parentRef}, nil)
				m.On("GetOrganizationUnit", testCoverageParentOUID).
					Return(OrganizationUnit{ID: testCoverageParentOUID, Parent: &rootRef}, nil)
				m.On("GetOrganizationUnit", "root-ou").
					Return(OrganizationUnit{ID: "root-ou", Parent: nil}, nil)
			},
			wantIDs: []string{"grandchild-ou", testCoverageParentOUID, "root-ou"},
		},
		{
			name: "BrokenChain_OUNotFound_ReturnsNilAndError",
			ouID: "orphan-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				m.On("GetOrganizationUnit", "orphan-ou").
					Return(OrganizationUnit{}, ErrOrganizationUnitNotFound)
			},
			wantErr: true,
		},
		{
			name: "BrokenChainMidWalk_OUNotFound_ReturnsNilAndError",
			ouID: "child-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				missingRef := "missing-ou"
				m.On("GetOrganizationUnit", "child-ou").
					Return(OrganizationUnit{ID: "child-ou", Parent: &missingRef}, nil)
				m.On("GetOrganizationUnit", "missing-ou").
					Return(OrganizationUnit{}, ErrOrganizationUnitNotFound)
			},
			wantErr: true,
		},
		{
			name: "StoreError_ReturnsNilAndError",
			ouID: "child-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				m.On("GetOrganizationUnit", "child-ou").
					Return(OrganizationUnit{}, genericErr)
			},
			wantErr: true,
		},
		{
			name: "StoreErrorMidWalk_ReturnsNilAndError",
			ouID: "child-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				parentRef := testCoverageParentOUID
				m.On("GetOrganizationUnit", "child-ou").
					Return(OrganizationUnit{ID: "child-ou", Parent: &parentRef}, nil)
				m.On("GetOrganizationUnit", testCoverageParentOUID).
					Return(OrganizationUnit{}, genericErr)
			},
			wantErr: true,
		},
		{
			name: "CyclicChain_ReturnsNilAndError",
			ouID: "child-ou",
			setupMock: func(m *organizationUnitStoreInterfaceMock) {
				parentRef := testCoverageParentOUID
				childRef := "child-ou"
				m.On("GetOrganizationUnit", "child-ou").
					Return(OrganizationUnit{ID: "child-ou", Parent: &parentRef}, nil).Times(1)
				m.On("GetOrganizationUnit", testCoverageParentOUID).
					Return(OrganizationUnit{ID: testCoverageParentOUID, Parent: &childRef}, nil).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			mockStore := newOrganizationUnitStoreInterfaceMock(suite.T())
			tt.setupMock(mockStore)
			resolver := newOUHierarchyAdapter(mockStore)

			ids, svcErr := resolver.GetAncestorOUIDs(context.Background(), tt.ouID)
			if tt.wantErr {
				assert.NotNil(suite.T(), svcErr)
				assert.Nil(suite.T(), ids)
			} else {
				assert.Nil(suite.T(), svcErr)
				assert.Equal(suite.T(), tt.wantIDs, ids)
			}
		})
	}
}
