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

// Package mgt provides internationalization functionality.
package mgt

import (
	"slices"

	goi18n "golang.org/x/text/language"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	sysi18n "github.com/asgardeo/thunder/internal/system/i18n/core"
	immutableresource "github.com/asgardeo/thunder/internal/system/immutable_resource"
	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "I18nMgtService"

// I18nServiceInterface defines the interface for the i18n service.
type I18nServiceInterface interface {
	ListLanguages() ([]string, *serviceerror.I18nServiceError)
	ResolveTranslations(language string, namespace string) (
		*LanguageTranslationsResponse, *serviceerror.I18nServiceError)
	SetTranslationOverrides(language string, translations map[string]map[string]string) (
		*LanguageTranslationsResponse, *serviceerror.I18nServiceError)
	ClearTranslationOverrides(language string) *serviceerror.I18nServiceError
	ResolveTranslationsForKey(language string, namespace string, key string) (
		*TranslationResponse, *serviceerror.I18nServiceError)
	SetTranslationOverrideForKey(language string, namespace string, key string, value string) (
		*TranslationResponse, *serviceerror.I18nServiceError)
	ClearTranslationOverrideForKey(language string, namespace string, key string) *serviceerror.I18nServiceError
}

// i18nService is the default implementation of I18nServiceInterface.
type i18nService struct {
	store  i18nStoreInterface
	logger *log.Logger
}

// newI18nService creates a new instance of i18nService with injected dependencies.
func newI18nService(store i18nStoreInterface) I18nServiceInterface {
	return &i18nService{
		store:  store,
		logger: log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
	}
}

// ListLanguages retrieves all locale codes that have translations in the system.
// The default locale is always included in the response, even if it has no translations in the DB.
func (s *i18nService) ListLanguages() ([]string, *serviceerror.I18nServiceError) {
	localeCodes, err := s.store.GetDistinctLanguages()
	if err != nil {
		s.logger.Error("Failed to get locales from store", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	// Ensure default language is always in the list
	hasDefaultLanguage := false
	for _, code := range localeCodes {
		if code == SystemLanguage {
			hasDefaultLanguage = true
			break
		}
	}

	if !hasDefaultLanguage {
		localeCodes = append(localeCodes, SystemLanguage)
	}

	return localeCodes, nil
}

// ResolveTranslationsForKey resolves a single translation by language, namespace, and key.
// It merges custom overrides with default values.
func (s *i18nService) ResolveTranslationsForKey(
	language string, namespace string, key string) (*TranslationResponse, *serviceerror.I18nServiceError) {
	if err := validate(language, namespace, key); err != nil {
		return nil, err
	}

	// Try to get from database first (custom override)
	trans, err := s.store.GetTranslationsByKey(key, namespace)
	if err != nil {
		s.logger.Error("Failed to get translation from store", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	requestedLang := goi18n.Make(language)

	bestTranslation := selectBestTranslation(trans, requestedLang)

	if bestTranslation.Value != "" {
		return &TranslationResponse{
			Language:  bestTranslation.Language,
			Namespace: bestTranslation.Namespace,
			Key:       bestTranslation.Key,
			Value:     bestTranslation.Value,
		}, nil
	}

	// If not in DB, check system translations for the key
	if namespace == SystemNamespace {
		defaultValue, exists := sysi18n.GetDefault(key)
		if exists {
			return &TranslationResponse{
				Language:  language,
				Namespace: SystemNamespace,
				Key:       key,
				Value:     defaultValue,
			}, nil
		}
	}

	// Not found (neither custom nor default)
	return nil, &ErrorTranslationNotFound
}

// SetTranslationOverrideForKey creates or updates a custom override for a single translation.
func (s *i18nService) SetTranslationOverrideForKey(
	language string, namespace string, key string, value string) (
	*TranslationResponse, *serviceerror.I18nServiceError) {
	if err := immutableresource.CheckImmutableUpdateI18n(); err != nil {
		return nil, err
	}
	if err := validate(language, namespace, key); err != nil {
		return nil, err
	}
	if value == "" {
		return nil, &ErrorMissingValue
	}

	trans := Translation{
		Key:       key,
		Language:  language,
		Namespace: namespace,
		Value:     value,
	}

	// Use upsert to create or update
	if err := s.store.UpsertTranslation(trans); err != nil {
		s.logger.Error("Failed to set translation override", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return &TranslationResponse{
		Language:  language,
		Namespace: namespace,
		Key:       key,
		Value:     value,
	}, nil
}

// ClearTranslationOverrideForKey removes the custom override for a single translation.
func (s *i18nService) ClearTranslationOverrideForKey(
	language string, namespace string, key string) *serviceerror.I18nServiceError {
	if err := immutableresource.CheckImmutableDeleteI18n(); err != nil {
		return err
	}
	if err := validate(language, namespace, key); err != nil {
		return err
	}

	if err := s.store.DeleteTranslation(language, key, namespace); err != nil {
		s.logger.Error("Failed to clear translation override", log.Error(err))
		return &ErrorInternalServerError
	}

	return nil
}

// ResolveTranslations resolves all translations for a language, organized by namespace.
// Merges custom overrides with default values.
func (s *i18nService) ResolveTranslations(
	language string, namespace string) (*LanguageTranslationsResponse, *serviceerror.I18nServiceError) {
	if language == "" {
		language = SystemLanguage
	}
	if !ValidateLanguage(language) {
		return nil, &ErrorInvalidLanguage
	}

	// If namespace is provided, validate it and filter by it
	if namespace != "" && !ValidateNamespace(namespace) {
		return nil, &ErrorInvalidNamespace
	}

	requestedLang := goi18n.Make(language)

	// Build the merged translations map
	result := make(map[string]map[string]string)
	totalResults := 0

	var dbTranslations map[string]map[string]Translation
	var err error

	if namespace == "" {
		// Get all namespaces

		// Start with defaults for system namespace
		result, totalResults = getSystemTranslations()

		dbTranslations, err = s.store.GetTranslations()
		if err != nil {
			s.logger.Error("Failed to get translations from store", log.Error(err))
			return nil, &ErrorInternalServerError
		}
	} else {
		// Get translations for specific namespace

		if namespace == SystemNamespace {
			// Start with defaults for system namespace
			result, totalResults = getSystemTranslations()
		}

		// Get DB translations for this language and namespace
		dbTranslations, err = s.store.GetTranslationsByNamespace(namespace)
		if err != nil {
			s.logger.Error("Failed to get translations from store", log.Error(err))
			return nil, &ErrorInternalServerError
		}
	}

	// Overlay DB translations
	for _, translations := range dbTranslations {
		translation := selectBestTranslation(translations, requestedLang)

		if result[translation.Namespace] == nil {
			result[translation.Namespace] = make(map[string]string)
		}
		// Check if this key already exists (from defaults)
		if _, exists := result[translation.Namespace][translation.Key]; !exists {
			totalResults++
		}
		result[translation.Namespace][translation.Key] = translation.Value
	}

	return &LanguageTranslationsResponse{
		Language:     language,
		TotalResults: totalResults,
		Translations: result,
	}, nil
}

// SetTranslationOverrides replaces all custom overrides for a language with provided values.
func (s *i18nService) SetTranslationOverrides(
	language string, translations map[string]map[string]string) (
	*LanguageTranslationsResponse, *serviceerror.I18nServiceError) {
	if err := immutableresource.CheckImmutableUpdateI18n(); err != nil {
		return nil, err
	}
	if language == "" {
		return nil, &ErrorMissingLanguage
	}
	if !ValidateLanguage(language) {
		return nil, &ErrorInvalidLanguage
	}
	if len(translations) == 0 {
		return nil, &ErrorEmptyTranslations
	}

	// Validate all entries first
	for ns, keys := range translations {
		if !ValidateNamespace(ns) {
			return nil, &ErrorInvalidNamespace
		}
		for key, value := range keys {
			if !ValidateKey(key) {
				return nil, &ErrorInvalidKey
			}
			if value == "" {
				return nil, &ErrorMissingValue
			}
		}
	}

	flattenedTranslations := []Translation{}
	for ns, keys := range translations {
		for key, value := range keys {
			flattenedTranslations = append(flattenedTranslations, Translation{
				Key:       key,
				Language:  language,
				Namespace: ns,
				Value:     value,
			})
		}
	}

	if err := s.store.UpsertTranslationsByLanguage(language, flattenedTranslations); err != nil {
		s.logger.Error("Failed to upsert translations", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	// TODO: return actual stored translations from DB
	return &LanguageTranslationsResponse{
		Language:     language,
		TotalResults: len(flattenedTranslations),
		Translations: translations,
	}, nil
}

// ClearTranslationOverrides removes all custom overrides for a language.
func (s *i18nService) ClearTranslationOverrides(language string) *serviceerror.I18nServiceError {
	if err := immutableresource.CheckImmutableDeleteI18n(); err != nil {
		return err
	}
	if language == "" {
		return &ErrorMissingLanguage
	}
	if !ValidateLanguage(language) {
		return &ErrorInvalidLanguage
	}

	if err := s.clearAllOverrides(language); err != nil {
		s.logger.Error("Failed to clear overrides", log.Error(err))
		return &ErrorInternalServerError
	}

	return nil
}

func (s *i18nService) clearAllOverrides(language string) error {
	err := s.store.DeleteTranslationsByLanguage(language)
	if err != nil {
		return err
	}
	return nil
}

func validate(language string, namespace string, key string) *serviceerror.I18nServiceError {
	if language == "" {
		return &ErrorMissingLanguage
	}
	if !ValidateLanguage(language) {
		return &ErrorInvalidLanguage
	}
	if !ValidateNamespace(namespace) {
		return &ErrorInvalidNamespace
	}
	if !ValidateKey(key) {
		return &ErrorInvalidKey
	}
	return nil
}

func getSystemTranslations() (map[string]map[string]string, int) {
	result := make(map[string]map[string]string)
	allDefaults := sysi18n.GetAllDefaults()
	if len(allDefaults) > 0 {
		result[SystemNamespace] = make(map[string]string)
		for key, value := range allDefaults {
			result[SystemNamespace][key] = value
		}
	}
	return result, len(allDefaults)
}

func selectBestTranslation(availableTranslations map[string]Translation, requestedLang goi18n.Tag) Translation {
	if len(availableTranslations) == 0 {
		return Translation{}
	}

	availableLangTags := make([]string, 0, len(availableTranslations))
	for langTag := range availableTranslations {
		availableLangTags = append(availableLangTags, langTag)
	}
	slices.SortFunc(availableLangTags, compareLangs)

	availableLangs := make([]goi18n.Tag, 0, len(availableLangTags))

	for _, langTag := range availableLangTags {
		availableLangs = append(availableLangs, goi18n.BCP47.Make(langTag))
	}

	matcher := goi18n.NewMatcher(availableLangs)
	_, index, _ := matcher.Match(requestedLang)

	return availableTranslations[availableLangTags[index]]
}

func compareLangs(langA, langB string) int {
	langAPref, aExists := LanguagePreferenceOrder[langA]
	langBPref, bExists := LanguagePreferenceOrder[langB]

	if aExists && bExists {
		return langAPref - langBPref
	}
	if aExists {
		return -1
	}
	if bExists {
		return 1
	}
	return 0
}
