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

// Package resolve provides functionality for resolving design configurations.
package resolve

import (
	"encoding/json"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/design/common"
	layoutmgt "github.com/asgardeo/thunder/internal/design/layout/mgt"
	thememgt "github.com/asgardeo/thunder/internal/design/theme/mgt"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

const serviceLogger = "DesignResolveService"

// DesignResolveServiceInterface defines the interface for the design resolve service.
type DesignResolveServiceInterface interface {
	ResolveDesign(
		resolveType common.DesignResolveType, id string,
	) (*common.DesignResponse, *serviceerror.ServiceError)
}

// designResolveService is the default implementation of the DesignResolveServiceInterface.
type designResolveService struct {
	themeMgtService    thememgt.ThemeMgtServiceInterface
	layoutMgtService   layoutmgt.LayoutMgtServiceInterface
	applicationService application.ApplicationServiceInterface
	logger             *log.Logger
}

// newDesignResolveService creates a new instance of DesignResolveService with injected dependencies.
func newDesignResolveService(
	themeMgtService thememgt.ThemeMgtServiceInterface,
	layoutMgtService layoutmgt.LayoutMgtServiceInterface,
	applicationService application.ApplicationServiceInterface,
) DesignResolveServiceInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, serviceLogger))
	return &designResolveService{
		themeMgtService:    themeMgtService,
		layoutMgtService:   layoutMgtService,
		applicationService: applicationService,
		logger:             logger,
	}
}

// ResolveDesign resolves a design configuration by type and ID.
// TODO: Add support for OU type and fallback logic.
func (drs *designResolveService) ResolveDesign(
	resolveType common.DesignResolveType, id string,
) (*common.DesignResponse, *serviceerror.ServiceError) {
	if resolveType == "" {
		return nil, &common.ErrorInvalidResolveType
	}

	if id == "" {
		return nil, &common.ErrorMissingResolveID
	}

	// Currently only APP type is supported
	if resolveType != common.DesignResolveTypeAPP {
		return nil, &common.ErrorUnsupportedResolveType
	}

	// Get the application by ID
	if drs.applicationService == nil {
		drs.logger.Error("Application service is not available")
		return nil, &serviceerror.InternalServerError
	}

	app, svcErr := drs.applicationService.GetApplication(id)
	if svcErr != nil {
		// Convert application service errors to design resolve errors
		if svcErr.Code == application.ErrorApplicationNotFound.Code ||
			svcErr.Code == application.ErrorInvalidApplicationID.Code {
			return nil, &common.ErrorApplicationNotFound
		}
		return nil, svcErr
	}

	// Check if the application has theme or layout configured
	if app.ThemeID == "" && app.LayoutID == "" {
		return nil, &common.ErrorApplicationHasNoDesign
	}

	// Prepare merged preferences
	mergedPrefs := make(map[string]interface{})
	var displayName string
	var responseID string

	// Get theme configuration if available
	if app.ThemeID != "" {
		themeConfig, svcErr := drs.themeMgtService.GetTheme(app.ThemeID)
		if svcErr != nil {
			if svcErr.Code == thememgt.ErrorThemeNotFound.Code {
				drs.logger.Error("Data integrity issue: application references non-existent theme",
					log.String("applicationId", id),
					log.String("themeId", app.ThemeID))
				return nil, &serviceerror.InternalServerError
			}
			return nil, svcErr
		}

		// Use theme display name and ID
		displayName = themeConfig.DisplayName
		responseID = themeConfig.ID

		// Parse theme preferences
		var themePrefs map[string]interface{}
		if len(themeConfig.Preferences) > 0 {
			if err := json.Unmarshal(themeConfig.Preferences, &themePrefs); err != nil {
				drs.logger.Error("Failed to parse theme preferences",
					log.String("themeId", app.ThemeID),
					log.Error(err))
				return nil, &serviceerror.InternalServerError
			}
			// Merge theme preferences
			for k, v := range themePrefs {
				mergedPrefs[k] = v
			}
		}
	}

	// Get layout configuration if available
	if app.LayoutID != "" {
		layoutConfig, svcErr := drs.layoutMgtService.GetLayout(app.LayoutID)
		if svcErr != nil {
			if svcErr.Code == layoutmgt.ErrorLayoutNotFound.Code {
				drs.logger.Error("Data integrity issue: application references non-existent layout",
					log.String("applicationId", id),
					log.String("layoutId", app.LayoutID))
				return nil, &serviceerror.InternalServerError
			}
			return nil, svcErr
		}

		// If theme wasn't set, use layout display name and ID
		if displayName == "" {
			displayName = layoutConfig.DisplayName
			responseID = layoutConfig.ID
		} else {
			// Combine display names if both exist
			displayName = displayName + " / " + layoutConfig.DisplayName
			responseID = app.ThemeID + "|" + app.LayoutID
		}

		// Parse layout preferences
		var layoutPrefs map[string]interface{}
		if len(layoutConfig.Preferences) > 0 {
			if err := json.Unmarshal(layoutConfig.Preferences, &layoutPrefs); err != nil {
				drs.logger.Error("Failed to parse layout preferences",
					log.String("layoutId", app.LayoutID),
					log.Error(err))
				return nil, &serviceerror.InternalServerError
			}
			// Merge layout preferences (layout preferences override theme preferences)
			for k, v := range layoutPrefs {
				mergedPrefs[k] = v
			}
		}
	}

	// Convert merged preferences back to JSON
	mergedPrefsJSON, err := json.Marshal(mergedPrefs)
	if err != nil {
		drs.logger.Error("Failed to marshal merged preferences", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	designResponse := &common.DesignResponse{
		ID:          responseID,
		DisplayName: displayName,
		Preferences: mergedPrefsJSON,
	}

	drs.logger.Debug("Successfully resolved design configuration",
		log.String("type", string(resolveType)),
		log.String("id", id),
		log.String("themeId", app.ThemeID),
		log.String("layoutId", app.LayoutID))

	return designResponse, nil
}
