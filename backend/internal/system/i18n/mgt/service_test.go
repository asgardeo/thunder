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
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	immutableresource "github.com/asgardeo/thunder/internal/system/immutable_resource"
)

type I18nMgtServiceTestSuite struct {
	suite.Suite
	mockStore *i18nStoreInterfaceMock
	service   I18nServiceInterface
}

func TestI18nMgtServiceTestSuite(t *testing.T) {
	suite.Run(t, new(I18nMgtServiceTestSuite))
}

func (suite *I18nMgtServiceTestSuite) SetupTest() {
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		ImmutableResources: config.ImmutableResources{
			Enabled: false,
		},
	}
	_ = config.InitializeThunderRuntime("/tmp/test", testConfig)
	suite.mockStore = newI18nStoreInterfaceMock(suite.T())
	suite.service = newI18nService(suite.mockStore)
}

func (suite *I18nMgtServiceTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

// ListLanguages Tests
func (suite *I18nMgtServiceTestSuite) TestListLanguages_Success() {
	expectedLangs := []string{"en-US", "fr-FR"}
	suite.mockStore.On("GetDistinctLanguages").Return(expectedLangs, nil)

	result, err := suite.service.ListLanguages()

	suite.Nil(err)
	suite.NotNil(result)
	suite.Contains(result, "en-US")
	suite.Contains(result, "fr-FR")
}

func (suite *I18nMgtServiceTestSuite) TestListLanguages_StoreError() {
	suite.mockStore.On("GetDistinctLanguages").Return(nil, errors.New("db error"))

	result, err := suite.service.ListLanguages()

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestListLanguages_AddsSystemLanguage() {
	// If store returns empty or doesn't have system language, it should be added
	suite.mockStore.On("GetDistinctLanguages").Return([]string{"fr-FR"}, nil)

	result, err := suite.service.ListLanguages()

	suite.Nil(err)
	suite.Contains(result, SystemLanguage)
	suite.Contains(result, "fr-FR")
}

// ResolveTranslationsForKey Tests
func (suite *I18nMgtServiceTestSuite) TestResolveTranslationsForKey_Success_FromStore() {
	translation := Translation{
		Key:       "welcome",
		Namespace: "common",
		Language:  "en-US",
		Value:     "Welcome override",
	}
	translationsMap := map[string]Translation{"en-US": translation}

	suite.mockStore.On("GetTranslationsByKey", "welcome", "common").Return(translationsMap, nil)

	result, err := suite.service.ResolveTranslationsForKey("en-US", "common", "welcome")

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("Welcome override", result.Value)
}

func (suite *I18nMgtServiceTestSuite) TestResolveTranslationsForKey_ValidationErrors() {
	testCases := []struct {
		name      string
		lang      string
		namespace string
		key       string
		errCode   string
	}{
		{"MissingLanguage", "", "ns", "key", ErrorMissingLanguage.Code},
		{"InvalidLanguage", "invalid", "ns", "key", ErrorInvalidLanguage.Code},
		{"InvalidNamespace", "en-US", "invalid!", "key", ErrorInvalidNamespace.Code},
		{"InvalidKey", "en-US", "ns", "invalid key!", ErrorInvalidKey.Code},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result, err := suite.service.ResolveTranslationsForKey(tc.lang, tc.namespace, tc.key)
			suite.Nil(result)
			suite.NotNil(err)
			suite.Equal(tc.errCode, err.Code)
		})
	}
}

func (suite *I18nMgtServiceTestSuite) TestResolveTranslationsForKey_NotFound() {
	suite.mockStore.On("GetTranslationsByKey", "unknown", "common").Return((map[string]Translation)(nil), nil)

	result, err := suite.service.ResolveTranslationsForKey("en-US", "common", "unknown")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorTranslationNotFound.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestResolveTranslationsForKey_StoreError() {
	suite.mockStore.On("GetTranslationsByKey", "welcome", "common").Return(nil, errors.New("db error"))

	result, err := suite.service.ResolveTranslationsForKey("en-US", "common", "welcome")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

// SetTranslationOverrideForKey Tests
func (suite *I18nMgtServiceTestSuite) TestSetTranslationOverrideForKey_Success() {
	suite.mockStore.On("UpsertTranslation", mock.AnythingOfType("mgt.Translation")).Return(nil)

	result, err := suite.service.SetTranslationOverrideForKey("en-US", "common", "welcome", "Hello")

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("Hello", result.Value)
}

func (suite *I18nMgtServiceTestSuite) TestSetTranslationOverrideForKey_ValidationErrors() {
	// Simple check for one validation case as others share logic
	result, err := suite.service.SetTranslationOverrideForKey("", "ns", "key", "val")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingLanguage.Code, err.Code)

	// Invalid Lang
	result, err = suite.service.SetTranslationOverrideForKey("invalid", "ns", "key", "val")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidLanguage.Code, err.Code)

	// Invalid Namespace
	result, err = suite.service.SetTranslationOverrideForKey("en-US", "invalid!", "key", "val")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidNamespace.Code, err.Code)

	// Invalid Key
	result, err = suite.service.SetTranslationOverrideForKey("en-US", "common", "invalid key!", "val")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidKey.Code, err.Code)

	result, err = suite.service.SetTranslationOverrideForKey("en-US", "ns", "key", "")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingValue.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestSetTranslationOverrideForKey_StoreError() {
	suite.mockStore.On("UpsertTranslation", mock.AnythingOfType("mgt.Translation")).Return(errors.New("db error"))

	result, err := suite.service.SetTranslationOverrideForKey("en-US", "common", "welcome", "Hello")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestSetTranslationOverrideForKey_Immutable() {
	// Enable immutable mode
	config.GetThunderRuntime().Config.ImmutableResources.Enabled = true
	defer func() {
		config.GetThunderRuntime().Config.ImmutableResources.Enabled = false
	}()

	result, err := suite.service.SetTranslationOverrideForKey("en-US", "common", "welcome", "Hello")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(immutableresource.I18nErrorImmutableResourceUpdateOperation.Code, err.Code)
}

// ClearTranslationOverrideForKey Tests
func (suite *I18nMgtServiceTestSuite) TestClearTranslationOverrideForKey_Success() {
	suite.mockStore.On("DeleteTranslation", "en-US", "welcome", "common").Return(nil)

	err := suite.service.ClearTranslationOverrideForKey("en-US", "common", "welcome")

	suite.Nil(err)
}

func (suite *I18nMgtServiceTestSuite) TestClearTranslationOverrideForKey_ValidationErrors() {
	err := suite.service.ClearTranslationOverrideForKey("", "ns", "key")
	suite.NotNil(err)
	suite.Equal(ErrorMissingLanguage.Code, err.Code)

	err = suite.service.ClearTranslationOverrideForKey("invalid", "ns", "key")
	suite.NotNil(err)
	suite.Equal(ErrorInvalidLanguage.Code, err.Code)

	err = suite.service.ClearTranslationOverrideForKey("en-US", "invalid!", "key")
	suite.NotNil(err)
	suite.Equal(ErrorInvalidNamespace.Code, err.Code)

	err = suite.service.ClearTranslationOverrideForKey("en-US", "ns", "invalid key!")
	suite.NotNil(err)
	suite.Equal(ErrorInvalidKey.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestClearTranslationOverrideForKey_StoreError() {
	suite.mockStore.On("DeleteTranslation", "en-US", "welcome", "common").Return(errors.New("db error"))

	err := suite.service.ClearTranslationOverrideForKey("en-US", "common", "welcome")

	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestClearTranslationOverrideForKey_Immutable() {
	// Enable immutable mode
	config.GetThunderRuntime().Config.ImmutableResources.Enabled = true
	defer func() {
		config.GetThunderRuntime().Config.ImmutableResources.Enabled = false
	}()

	err := suite.service.ClearTranslationOverrideForKey("en-US", "common", "welcome")

	suite.NotNil(err)
	suite.Equal(immutableresource.I18nErrorImmutableResourceDeleteOperation.Code, err.Code)
}

// ResolveTranslations Tests
func (suite *I18nMgtServiceTestSuite) TestResolveTranslations_CustomNamespace_Success() {
	translation := Translation{
		Key:       "btn_ok",
		Namespace: "console",
		Language:  "en-US",
		Value:     "OK",
	}

	// Let's correct the mock data structure
	mockDataCorrect := map[string]map[string]Translation{
		"btn_ok": {
			"en-US": translation,
		},
	}

	suite.mockStore.On("GetTranslationsByNamespace", "console").Return(mockDataCorrect, nil)

	result, err := suite.service.ResolveTranslations("en-US", "console")

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(1, result.TotalResults)
	suite.Equal("OK", result.Translations["console"]["btn_ok"])
}

func (suite *I18nMgtServiceTestSuite) TestResolveTranslations_InvalidNamespace() {
	result, err := suite.service.ResolveTranslations("en-US", "invalid!")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidNamespace.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestResolveTranslations_StoreError() {
	suite.mockStore.On("GetTranslationsByNamespace", "console").Return(nil, errors.New("db error"))

	result, err := suite.service.ResolveTranslations("en-US", "console")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestResolveTranslations_DefaultsFallback() {
	// Mock: no custom translations in DB
	suite.mockStore.On("GetTranslationsByNamespace", "system").
		Return(make(map[string]map[string]Translation), nil)

	result, err := suite.service.ResolveTranslations("fr-FR", "system")

	suite.Nil(err)
	suite.NotNil(result)
	// Just verify we got a result structure back even if translations loop was empty or default only
	suite.Equal("fr-FR", result.Language)
}

func (suite *I18nMgtServiceTestSuite) TestResolveTranslations_MergeLogic() {
	// DB has an override for "welcome"
	translationDB := Translation{
		Key:       "welcome",
		Namespace: "system",
		Language:  "en-US",
		Value:     "Welcome Override",
	}
	dbTranslations := map[string]map[string]Translation{
		"welcome": {"en-US": translationDB},
	}

	suite.mockStore.On("GetTranslationsByNamespace", "system").Return(dbTranslations, nil)

	result, err := suite.service.ResolveTranslations("en-US", "system")

	suite.Nil(err)
	suite.NotNil(result)
	// Should contain override
	suite.Equal("Welcome Override", result.Translations["system"]["welcome"])
}

func (suite *I18nMgtServiceTestSuite) TestResolveTranslations_AllNamespaces() {
	// Test without namespace filter
	// Mock: returns mixed namespaces
	dbTranslations := map[string]map[string]Translation{
		"k1": {"en-US": {Key: "k1", Namespace: "ns1", Language: "en-US", Value: "v1"}},
		"k2": {"en-US": {Key: "k2", Namespace: "ns2", Language: "en-US", Value: "v2"}},
	}

	suite.mockStore.On("GetTranslations").Return(dbTranslations, nil)

	result, err := suite.service.ResolveTranslations("en-US", "")

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("v1", result.Translations["ns1"]["k1"])
	suite.Equal("v2", result.Translations["ns2"]["k2"])
}

func (suite *I18nMgtServiceTestSuite) TestResolveTranslations_AllNamespaces_StoreError() {
	suite.mockStore.On("GetTranslations").Return(nil, errors.New("db error"))

	result, err := suite.service.ResolveTranslations("en-US", "")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestCompareLangs() {
	// Directly test the unexported compareLangs function

	// Case 1: Both exist (en-US=0, en=1) -> 0 - 1 = -1
	suite.Equal(-1, compareLangs("en-US", "en"))

	// Case 2: Both exist reverse (en=1, en-US=0) -> 1 - 0 = 1
	suite.Equal(1, compareLangs("en", "en-US"))

	// Case 3: A exists, B does not
	suite.Equal(-1, compareLangs("en-US", "fr"))

	// Case 4: A does not exist, B does
	suite.Equal(1, compareLangs("fr", "en-US"))

	// Case 5: Neither exists
	suite.Equal(0, compareLangs("fr", "de"))

	// Case 6: Same language
	suite.Equal(0, compareLangs("en-US", "en-US"))
}

// SetTranslationOverrides Tests
func (suite *I18nMgtServiceTestSuite) TestSetTranslationOverrides_Success() {
	translations := map[string]map[string]string{
		"console": {
			"k1": "v1",
		},
	}

	suite.mockStore.On("UpsertTranslationsByLanguage", "en-US", mock.AnythingOfType("[]mgt.Translation")).Return(nil)

	result, err := suite.service.SetTranslationOverrides("en-US", translations)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(1, result.TotalResults)
}

func (suite *I18nMgtServiceTestSuite) TestSetTranslationOverrides_Empty() {
	translations := map[string]map[string]string{}
	result, err := suite.service.SetTranslationOverrides("en-US", translations)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorEmptyTranslations.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestSetTranslationOverrides_ValidationErrors() {
	// Invalid Namespace
	translations1 := map[string]map[string]string{
		"invalid!": {"k": "v"},
	}
	result, err := suite.service.SetTranslationOverrides("en-US", translations1)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidNamespace.Code, err.Code)

	// Invalid Key
	translations2 := map[string]map[string]string{
		"console": {"invalid key!": "v"},
	}
	result, err = suite.service.SetTranslationOverrides("en-US", translations2)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidKey.Code, err.Code)

	// Empty Value
	translations3 := map[string]map[string]string{
		"console": {"key": ""},
	}
	result, err = suite.service.SetTranslationOverrides("en-US", translations3)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingValue.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestSetTranslationOverrides_StoreError() {
	translations := map[string]map[string]string{
		"console": {"k": "v"},
	}
	suite.mockStore.On("UpsertTranslationsByLanguage", "en-US", mock.AnythingOfType("[]mgt.Translation")).
		Return(errors.New("db error"))

	result, err := suite.service.SetTranslationOverrides("en-US", translations)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestSetTranslationOverrides_Immutable() {
	// Enable immutable mode
	config.GetThunderRuntime().Config.ImmutableResources.Enabled = true
	defer func() {
		config.GetThunderRuntime().Config.ImmutableResources.Enabled = false
	}()

	translations := map[string]map[string]string{
		"console": {"k": "v"},
	}

	result, err := suite.service.SetTranslationOverrides("en-US", translations)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(immutableresource.I18nErrorImmutableResourceUpdateOperation.Code, err.Code)
}

// ClearTranslationOverrides Tests
func (suite *I18nMgtServiceTestSuite) TestClearTranslationOverrides_Success() {
	suite.mockStore.On("DeleteTranslationsByLanguage", "en-US").Return(nil)

	err := suite.service.ClearTranslationOverrides("en-US")

	suite.Nil(err)
}

func (suite *I18nMgtServiceTestSuite) TestClearTranslationOverrides_StoreError() {
	suite.mockStore.On("DeleteTranslationsByLanguage", "en-US").Return(errors.New("db error"))

	err := suite.service.ClearTranslationOverrides("en-US")

	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestClearTranslationOverrides_ValidationErrors() {
	err := suite.service.ClearTranslationOverrides("")
	suite.NotNil(err)
	suite.Equal(ErrorMissingLanguage.Code, err.Code)

	err = suite.service.ClearTranslationOverrides("invalid")
	suite.NotNil(err)
	suite.Equal(ErrorInvalidLanguage.Code, err.Code)
}

func (suite *I18nMgtServiceTestSuite) TestClearTranslationOverrides_Immutable() {
	// Enable immutable mode
	config.GetThunderRuntime().Config.ImmutableResources.Enabled = true
	defer func() {
		config.GetThunderRuntime().Config.ImmutableResources.Enabled = false
	}()

	err := suite.service.ClearTranslationOverrides("en-US")

	suite.NotNil(err)
	suite.Equal(immutableresource.I18nErrorImmutableResourceDeleteOperation.Code, err.Code)
}
