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

package mgt

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	immutableresource "github.com/asgardeo/thunder/internal/system/immutable_resource"
	"github.com/asgardeo/thunder/internal/system/immutable_resource/entity"
)

type ImmutableResourceTestSuite struct {
	suite.Suite
	mockStore *i18nStoreInterfaceMock
	exporter  *TranslationExporter
}

func TestImmutableResourceTestSuite(t *testing.T) {
	suite.Run(t, new(ImmutableResourceTestSuite))
}

func (s *ImmutableResourceTestSuite) SetupTest() {
	s.mockStore = newI18nStoreInterfaceMock(s.T())
	s.exporter = newTranslationExporter(s.mockStore)
}

func (s *ImmutableResourceTestSuite) TestGetResourceType() {
	resourceType := s.exporter.GetResourceType()
	assert.Equal(s.T(), "translation", resourceType)
}

func (s *ImmutableResourceTestSuite) TestGetParameterizerType() {
	paramType := s.exporter.GetParameterizerType()
	assert.Equal(s.T(), "Translation", paramType)
}

func (s *ImmutableResourceTestSuite) TestGetResourceByID() {
	translations := map[string]map[string]Translation{
		"welcome": {
			"en-US": {
				Key:       "welcome",
				Language:  "en-US",
				Namespace: "common",
				Value:     "Welcome",
			},
		},
		"goodbye": {
			"en-US": {
				Key:       "goodbye",
				Language:  "en-US",
				Namespace: "common",
				Value:     "Goodbye",
			},
		},
	}

	s.mockStore.On("GetTranslations").Return(translations, nil)

	resource, name, err := s.exporter.GetResourceByID("en-US")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "en-US", name)
	assert.NotNil(s.T(), resource)

	trans, ok := resource.(*LanguageTranslations)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), "en-US", trans.Language)
	assert.Equal(s.T(), "Welcome", trans.Translations["common"]["welcome"])
	assert.Equal(s.T(), "Goodbye", trans.Translations["common"]["goodbye"])
}

func (s *ImmutableResourceTestSuite) TestGetResourceByID_NotFound() {
	s.mockStore.On("GetTranslations").Return(map[string]map[string]Translation{}, nil)

	_, _, err := s.exporter.GetResourceByID("fr-FR")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "TRANSLATION_NOT_FOUND", err.Code)
}

func (s *ImmutableResourceTestSuite) TestGetResourceByID_StoreError() {
	s.mockStore.On("GetTranslations").Return(nil, errors.New("db error"))

	_, _, err := s.exporter.GetResourceByID("en-US")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "I18N_FETCH_ERROR", err.Code)
}

func (s *ImmutableResourceTestSuite) TestValidateResource() {
	trans := &LanguageTranslations{
		Language: "en-US",
		Translations: map[string]map[string]string{
			"common": {"ok": "OK"},
		},
	}

	name, err := s.exporter.ValidateResource(trans, "en-US", nil)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "en-US", name)
}

func (s *ImmutableResourceTestSuite) TestValidateResourceInvalidType() {
	invalidResource := "not a translation"

	name, err := s.exporter.ValidateResource(invalidResource, "en-US", nil)
	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), "INVALID_TYPE", err.Code)
}

func (s *ImmutableResourceTestSuite) TestValidateResourceMissingLanguage() {
	trans := &LanguageTranslations{
		Translations: map[string]map[string]string{
			"common": {"ok": "OK"},
		},
	}

	name, err := s.exporter.ValidateResource(trans, "en-US", nil)
	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), "INVALID_TRANSLATION", err.Code)
}

func (s *ImmutableResourceTestSuite) TestValidateResourceMissingTranslations() {
	trans := &LanguageTranslations{
		Language: "en-US",
	}

	name, err := s.exporter.ValidateResource(trans, "en-US", nil)
	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), "INVALID_TRANSLATION", err.Code)
}

func (s *ImmutableResourceTestSuite) TestGetAllResourceIDs() {
	languages := []string{"en-US", "fr-FR"}
	s.mockStore.On("GetDistinctLanguages").Return(languages, nil)

	ids, err := s.exporter.GetAllResourceIDs()
	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 2)
	assert.Contains(s.T(), ids, "en-US")
	assert.Contains(s.T(), ids, "fr-FR")
}

func (s *ImmutableResourceTestSuite) TestGetAllResourceIDs_StoreError() {
	s.mockStore.On("GetDistinctLanguages").Return(nil, errors.New("db error"))

	ids, err := s.exporter.GetAllResourceIDs()
	assert.NotNil(s.T(), err)
	assert.Nil(s.T(), ids)
	assert.Equal(s.T(), "I18N_EXPORT_ERROR", err.Code)
}

func (s *ImmutableResourceTestSuite) TestParseToLangTranslation() {
	yamlData := []byte(`
language: en-US
translations:
  common:
    welcome: Welcome
    goodbye: Goodbye
`)

	trans, err := parseToLangTranslation(yamlData)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), trans)
	assert.Equal(s.T(), "en-US", trans.Language)
	assert.Equal(s.T(), "Welcome", trans.Translations["common"]["welcome"])
}

func (s *ImmutableResourceTestSuite) TestValidateTranslationWrapper() {
	store := newFileBasedStore().(*fileBasedStore)
	trans := &LanguageTranslations{
		Language: "en-US",
		Translations: map[string]map[string]string{
			"common": {"ok": "OK"},
		},
	}

	err := validateTranslationWrapper(trans, store)
	assert.NoError(s.T(), err)
}

func (s *ImmutableResourceTestSuite) TestValidateTranslationWrapperInvalidType() {
	store := newFileBasedStore().(*fileBasedStore)
	err := validateTranslationWrapper("invalid", store)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid type")
}

func (s *ImmutableResourceTestSuite) TestValidateTranslationWrapperMissingLang() {
	store := newFileBasedStore().(*fileBasedStore)
	trans := &LanguageTranslations{
		Translations: map[string]map[string]string{
			"common": {"ok": "OK"},
		},
	}
	err := validateTranslationWrapper(trans, store)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "language is required")
}

func (s *ImmutableResourceTestSuite) TestValidateTranslationWrapperMissingTrans() {
	store := newFileBasedStore().(*fileBasedStore)
	trans := &LanguageTranslations{
		Language: "en-US",
	}
	err := validateTranslationWrapper(trans, store)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "translations is required")
}

func (s *ImmutableResourceTestSuite) TestValidateTranslationWrapperDuplicateID() {
	store := newFileBasedStore().(*fileBasedStore)
	trans := &LanguageTranslations{
		Language: "en-US",
		Translations: map[string]map[string]string{
			"common": {"ok": "OK"},
		},
	}

	err := store.Create("en-US", trans)
	assert.NoError(s.T(), err)

	// validate duplicate
	err = validateTranslationWrapper(trans, store)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "duplicate translation ID")
}

func (s *ImmutableResourceTestSuite) TestGetResourceRules() {
	rules := s.exporter.GetResourceRules()
	assert.NotNil(s.T(), rules)
	assert.Empty(s.T(), rules.Variables)
	assert.Empty(s.T(), rules.ArrayVariables)
}

func (s *ImmutableResourceTestSuite) TestLoadImmutableResources_InvalidStoreType() {
	// Pass a mock store that is NOT *fileBasedStore
	// s.mockStore is *i18nStoreInterfaceMock which satisfies i18nStoreInterface
	// but is not *fileBasedStore.
	err := loadImmutableResources(s.mockStore)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "fileStore must be a file-based store implementation")
}

func (s *ImmutableResourceTestSuite) TestLoadImmutableResources_Success() {
	// Setup temp dir
	tempDir, err := os.MkdirTemp("", "thunder_test_resources_success")
	assert.NoError(s.T(), err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Setup Runtime with temp dir
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		ImmutableResources: config.ImmutableResources{
			Enabled: true,
		},
	}
	_ = config.InitializeThunderRuntime(tempDir, testConfig)
	defer config.ResetThunderRuntime()

	// Create translations directory
	translationsDir := filepath.Join(tempDir, "repository", "resources", "translations")
	err = os.MkdirAll(translationsDir, 0750)
	assert.NoError(s.T(), err)

	// Create valid YAML file
	yamlContent := []byte(`
language: es-ES
translations:
  common:
    hello: Hola
`)
	validFile := filepath.Join(translationsDir, "es.yaml")
	err = os.WriteFile(validFile, yamlContent, 0600)
	assert.NoError(s.T(), err)

	// Create actual fileStore (using generic store underneath)
	// We use NewGenericFileBasedStoreForTest to avoid using the singleton entity store
	// and ensuring test isolation.
	genericStore := immutableresource.NewGenericFileBasedStoreForTest(entity.KeyTypeTranslation)
	store := &fileBasedStore{
		GenericFileBasedStore: genericStore,
	}

	err = loadImmutableResources(store)
	assert.NoError(s.T(), err)

	exists, err := store.IsTranslationExists("es-ES")
	assert.NoError(s.T(), err)
	assert.True(s.T(), exists)
}

func (s *ImmutableResourceTestSuite) TestParseToTranslationWrapper() {
	yamlData := []byte(`
language: de-DE
translations:
  common:
    hello: Hallo
`)
	data, err := parseToTranslationWrapper(yamlData)
	assert.NoError(s.T(), err)
	trans, ok := data.(*LanguageTranslations)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), "de-DE", trans.Language)
}

func (s *ImmutableResourceTestSuite) TestParseToLangTranslation_Error() {
	yamlData := []byte(`invalid_yaml: [`)
	_, err := parseToLangTranslation(yamlData)
	assert.Error(s.T(), err)
}
