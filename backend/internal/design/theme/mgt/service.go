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

package thememgt

import (
	"encoding/json"
	"errors"
	"fmt"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "ThemeMgtService"

// ThemeMgtServiceInterface defines the interface for the theme management service.
type ThemeMgtServiceInterface interface {
	GetThemeList(limit, offset int) (*ThemeList, *serviceerror.ServiceError)
	CreateTheme(theme CreateThemeRequest) (*Theme, *serviceerror.ServiceError)
	GetTheme(id string) (*Theme, *serviceerror.ServiceError)
	UpdateTheme(id string, theme UpdateThemeRequest) (*Theme, *serviceerror.ServiceError)
	DeleteTheme(id string) *serviceerror.ServiceError
	IsThemeExist(id string) (bool, *serviceerror.ServiceError)
}

// themeMgtService is the default implementation of the ThemeMgtServiceInterface.
type themeMgtService struct {
	themeMgtStore themeMgtStoreInterface
	logger        *log.Logger
}

// newThemeMgtService creates a new instance of ThemeMgtService with injected dependencies.
func newThemeMgtService(themeMgtStore themeMgtStoreInterface) ThemeMgtServiceInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	return &themeMgtService{
		themeMgtStore: themeMgtStore,
		logger:        logger,
	}
}

// GetThemeList retrieves a list of theme configurations.
func (ts *themeMgtService) GetThemeList(limit, offset int) (*ThemeList, *serviceerror.ServiceError) {
	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := ts.themeMgtStore.GetThemeListCount()
	if err != nil {
		ts.logger.Error("Failed to get theme count", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	themes, err := ts.themeMgtStore.GetThemeList(limit, offset)
	if err != nil {
		ts.logger.Error("Failed to list themes", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	response := &ThemeList{
		TotalResults: totalCount,
		Themes:       themes,
		StartIndex:   offset + 1,
		Count:        len(themes),
		Links:        buildPaginationLinks(limit, offset, totalCount),
	}

	return response, nil
}

// CreateTheme creates a new theme configuration.
func (ts *themeMgtService) CreateTheme(theme CreateThemeRequest) (*Theme, *serviceerror.ServiceError) {
	ts.logger.Debug("Creating theme configuration")

	if theme.DisplayName == "" {
		return nil, &serviceerror.ServiceError{
			Code:             "THM-1005",
			Error:            "Missing display name",
			ErrorDescription: "Display name is required",
		}
	}

	if err := ts.validateThemePreferences(theme.Theme); err != nil {
		return nil, err
	}

	id := utils.GenerateUUID()
	if err := ts.themeMgtStore.CreateTheme(id, theme); err != nil {
		ts.logger.Error("Failed to create theme", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	createdTheme := &Theme{
		ID:          id,
		DisplayName: theme.DisplayName,
		Description: theme.Description,
		Theme:       theme.Theme,
	}

	ts.logger.Debug("Successfully created theme", log.String("id", id))
	return createdTheme, nil
}

// GetTheme retrieves a specific theme configuration by its id.
func (ts *themeMgtService) GetTheme(id string) (*Theme, *serviceerror.ServiceError) {
	ts.logger.Debug("Retrieving theme", log.String("id", id))

	if id == "" {
		return nil, &ErrorInvalidThemeID
	}

	theme, err := ts.themeMgtStore.GetTheme(id)
	if err != nil {
		if errors.Is(err, errThemeNotFound) {
			ts.logger.Debug("Theme not found", log.String("id", id))
			return nil, &ErrorThemeNotFound
		}
		ts.logger.Error("Failed to retrieve theme", log.String("id", id), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	ts.logger.Debug("Successfully retrieved theme", log.String("id", theme.ID))
	return &theme, nil
}

// UpdateTheme updates an existing theme configuration.
func (ts *themeMgtService) UpdateTheme(id string, theme UpdateThemeRequest) (*Theme, *serviceerror.ServiceError) {
	ts.logger.Debug("Updating theme", log.String("id", id))

	if id == "" {
		return nil, &ErrorInvalidThemeID
	}

	if theme.DisplayName == "" {
		return nil, &serviceerror.ServiceError{
			Code:             "THM-1005",
			Error:            "Missing display name",
			ErrorDescription: "Display name is required",
		}
	}

	if err := ts.validateThemePreferences(theme.Theme); err != nil {
		return nil, err
	}

	// Check if theme exists
	exists, err := ts.themeMgtStore.IsThemeExist(id)
	if err != nil {
		ts.logger.Error("Failed to check theme existence", log.String("id", id), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	if !exists {
		return nil, &ErrorThemeNotFound
	}

	if err := ts.themeMgtStore.UpdateTheme(id, theme); err != nil {
		ts.logger.Error("Failed to update theme", log.String("id", id), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	updatedTheme := &Theme{
		ID:          id,
		DisplayName: theme.DisplayName,
		Description: theme.Description,
		Theme:       theme.Theme,
	}

	ts.logger.Debug("Successfully updated theme", log.String("id", id))
	return updatedTheme, nil
}

// DeleteTheme deletes a theme configuration.
func (ts *themeMgtService) DeleteTheme(id string) *serviceerror.ServiceError {
	ts.logger.Debug("Deleting theme", log.String("id", id))

	if id == "" {
		return &ErrorInvalidThemeID
	}

	// Check if theme exists
	exists, err := ts.themeMgtStore.IsThemeExist(id)
	if err != nil {
		ts.logger.Error("Failed to check theme existence", log.String("id", id), log.Error(err))
		return &serviceerror.InternalServerError
	}

	if !exists {
		return &ErrorThemeNotFound
	}

	// Check if theme is used by any applications
	count, err := ts.themeMgtStore.GetApplicationsCountByThemeID(id)
	if err != nil {
		ts.logger.Error("Failed to check applications using theme", log.String("id", id), log.Error(err))
		return &serviceerror.InternalServerError
	}

	if count > 0 {
		return &serviceerror.ServiceError{
			Code:             "THM-1006",
			Error:            "Theme in use",
			ErrorDescription: fmt.Sprintf("Theme is being used by %d application(s)", count),
		}
	}

	if err := ts.themeMgtStore.DeleteTheme(id); err != nil {
		ts.logger.Error("Failed to delete theme", log.String("id", id), log.Error(err))
		return &serviceerror.InternalServerError
	}

	ts.logger.Debug("Successfully deleted theme", log.String("id", id))
	return nil
}

// IsThemeExist checks if a theme exists.
func (ts *themeMgtService) IsThemeExist(id string) (bool, *serviceerror.ServiceError) {
	if id == "" {
		return false, &ErrorInvalidThemeID
	}

	exists, err := ts.themeMgtStore.IsThemeExist(id)
	if err != nil {
		ts.logger.Error("Failed to check theme existence", log.String("id", id), log.Error(err))
		return false, &serviceerror.InternalServerError
	}

	return exists, nil
}

// validateThemePreferences validates the theme JSON.
func (ts *themeMgtService) validateThemePreferences(theme json.RawMessage) *serviceerror.ServiceError {
	if len(theme) == 0 {
		return &ErrorInvalidThemeData
	}

	var result map[string]interface{}
	if err := json.Unmarshal(theme, &result); err != nil {
		ts.logger.Debug("Invalid theme JSON", log.Error(err))
		return &ErrorInvalidThemeData
	}

	return nil
}

// validatePaginationParams validates limit and offset parameters.
func validatePaginationParams(limit, offset int) *serviceerror.ServiceError {
	if limit < 1 || limit > serverconst.MaxPageSize {
		return &serviceerror.ServiceError{
			Code:             "THM-1007",
			Error:            "Invalid limit",
			ErrorDescription: fmt.Sprintf("Limit must be between 1 and %d", serverconst.MaxPageSize),
		}
	}

	if offset < 0 {
		return &serviceerror.ServiceError{
			Code:             "THM-1008",
			Error:            "Invalid offset",
			ErrorDescription: "Offset must be non-negative",
		}
	}

	return nil
}

// buildPaginationLinks builds pagination links for the response.
func buildPaginationLinks(limit, offset, totalCount int) []Link {
	links := make([]Link, 0)

	// Previous link
	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		links = append(links, Link{
			Href: fmt.Sprintf("/design/themes?limit=%d&offset=%d", limit, prevOffset),
			Rel:  "previous",
		})
	}

	// Next link
	if offset+limit < totalCount {
		nextOffset := offset + limit
		links = append(links, Link{
			Href: fmt.Sprintf("/design/themes?limit=%d&offset=%d", limit, nextOffset),
			Rel:  "next",
		})
	}

	return links
}
