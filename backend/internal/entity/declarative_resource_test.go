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
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
)

type DeclarativeResourceTestSuite struct {
	suite.Suite
}

func TestDeclarativeResourceTestSuite(t *testing.T) {
	suite.Run(t, new(DeclarativeResourceTestSuite))
}

func (s *DeclarativeResourceTestSuite) SetupTest() {
	config.ResetThunderRuntime()
}

func (s *DeclarativeResourceTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

func (s *DeclarativeResourceTestSuite) TestLoadDeclarativeResources_MutableStore_Skipped() {
	mockStore := newEntityStoreInterfaceMock(s.T())
	mockSvc := newEntityServiceMock(s.T())
	cfg := DeclarativeLoaderConfig{Directory: "users", Category: EntityCategoryUser}

	err := loadDeclarativeResources(mockStore, mockSvc, cfg)
	s.NoError(err)
}

func (s *DeclarativeResourceTestSuite) TestLoadDeclarativeResources_FileStore_EmptyDirectory() {
	tmpDir := s.T().TempDir()
	resourceDir := filepath.Join(tmpDir, "repository", "resources", "users")
	s.Require().NoError(os.MkdirAll(resourceDir, 0750))

	s.Require().NoError(config.InitializeThunderRuntime(tmpDir, &config.Config{}))

	fileStore := newEntityFileBasedStore()
	mockSvc := newEntityServiceMock(s.T())

	cfg := DeclarativeLoaderConfig{
		Directory: "users",
		Category:  EntityCategoryUser,
		Parser: func(data []byte) (*Entity, json.RawMessage, json.RawMessage, error) {
			return &Entity{ID: "test-id"}, nil, nil, nil
		},
	}

	err := loadDeclarativeResources(fileStore, mockSvc, cfg)
	s.NoError(err)
}

func (s *DeclarativeResourceTestSuite) TestLoadDeclarativeResources_CompositeStore_ExtractsFileStore() {
	tmpDir := s.T().TempDir()
	resourceDir := filepath.Join(tmpDir, "repository", "resources", "users")
	s.Require().NoError(os.MkdirAll(resourceDir, 0750))
	s.Require().NoError(config.InitializeThunderRuntime(tmpDir, &config.Config{}))

	fileStore := newEntityFileBasedStore()
	dbStoreMock := newEntityStoreInterfaceMock(s.T())
	compositeStore := newEntityCompositeStore(fileStore, dbStoreMock)
	mockSvc := newEntityServiceMock(s.T())

	cfg := DeclarativeLoaderConfig{
		Directory: "users",
		Category:  EntityCategoryUser,
		Parser: func(data []byte) (*Entity, json.RawMessage, json.RawMessage, error) {
			return &Entity{ID: "test-id"}, nil, nil, nil
		},
	}

	err := loadDeclarativeResources(compositeStore, mockSvc, cfg)
	s.NoError(err)
}

func (s *DeclarativeResourceTestSuite) TestLoadDeclarativeResources_CompositeStore_NonFileStore_Skipped() {
	mockFileStore := newEntityStoreInterfaceMock(s.T())
	mockDBStore := newEntityStoreInterfaceMock(s.T())
	compositeStore := newEntityCompositeStore(mockFileStore, mockDBStore)
	mockSvc := newEntityServiceMock(s.T())

	cfg := DeclarativeLoaderConfig{Directory: "users", Category: EntityCategoryUser}
	err := loadDeclarativeResources(compositeStore, mockSvc, cfg)
	s.NoError(err)
}

func (s *DeclarativeResourceTestSuite) TestLoadDeclarativeResources_WithValidator_Called() {
	tmpDir := s.T().TempDir()
	resourceDir := filepath.Join(tmpDir, "repository", "resources", "items")
	s.Require().NoError(os.MkdirAll(resourceDir, 0750))

	entityYAML := []byte(`id: "item-1"
ou_id: "ou-1"
type: "thing"
category: "user"
attributes: {}
`)
	s.Require().NoError(os.WriteFile(filepath.Join(resourceDir, "item1.yaml"), entityYAML, 0600))
	s.Require().NoError(config.InitializeThunderRuntime(tmpDir, &config.Config{}))

	fileStore := newEntityFileBasedStore()
	mockSvc := newEntityServiceMock(s.T())

	validatorCalled := false
	cfg := DeclarativeLoaderConfig{
		Directory: "items",
		Category:  EntityCategoryUser,
		Parser: func(data []byte) (*Entity, json.RawMessage, json.RawMessage, error) {
			attrs, _ := json.Marshal(map[string]interface{}{})
			return &Entity{ID: "item-1", Category: EntityCategoryUser, Type: "thing",
				OUID: "ou-1", Attributes: json.RawMessage(attrs)}, nil, nil, nil
		},
		Validator: func(e *Entity, svc EntityServiceInterface) error {
			validatorCalled = true
			return nil
		},
		IDExtractor: func(e *Entity) string {
			return e.ID
		},
	}

	err := loadDeclarativeResources(fileStore, mockSvc, cfg)
	s.NoError(err)
	s.True(validatorCalled)
}

func (s *DeclarativeResourceTestSuite) TestLoadDeclarativeResources_ParserError() {
	tmpDir := s.T().TempDir()
	resourceDir := filepath.Join(tmpDir, "repository", "resources", "items")
	s.Require().NoError(os.MkdirAll(resourceDir, 0750))
	s.Require().NoError(os.WriteFile(filepath.Join(resourceDir, "bad.yaml"), []byte("id: x"), 0600))
	s.Require().NoError(config.InitializeThunderRuntime(tmpDir, &config.Config{}))

	fileStore := newEntityFileBasedStore()
	mockSvc := newEntityServiceMock(s.T())

	cfg := DeclarativeLoaderConfig{
		Directory: "items",
		Category:  EntityCategoryUser,
		Parser: func(data []byte) (*Entity, json.RawMessage, json.RawMessage, error) {
			return nil, nil, nil, errors.New("parse failed")
		},
	}

	err := loadDeclarativeResources(fileStore, mockSvc, cfg)
	s.Error(err)
}

func (s *DeclarativeResourceTestSuite) TestLoadDeclarativeResources_ValidatorError() {
	tmpDir := s.T().TempDir()
	resourceDir := filepath.Join(tmpDir, "repository", "resources", "items")
	s.Require().NoError(os.MkdirAll(resourceDir, 0750))
	s.Require().NoError(os.WriteFile(filepath.Join(resourceDir, "item.yaml"), []byte("id: x"), 0600))
	s.Require().NoError(config.InitializeThunderRuntime(tmpDir, &config.Config{}))

	fileStore := newEntityFileBasedStore()
	mockSvc := newEntityServiceMock(s.T())

	cfg := DeclarativeLoaderConfig{
		Directory: "items",
		Category:  EntityCategoryUser,
		Parser: func(data []byte) (*Entity, json.RawMessage, json.RawMessage, error) {
			return &Entity{ID: "x"}, nil, nil, nil
		},
		Validator: func(e *Entity, svc EntityServiceInterface) error {
			return errors.New("validation failed")
		},
	}

	err := loadDeclarativeResources(fileStore, mockSvc, cfg)
	s.Error(err)
}

func newEntityServiceMock(t *testing.T) *EntityServiceInterfaceMock {
	m := &EntityServiceInterfaceMock{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}
