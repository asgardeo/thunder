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
	"net/http"
	"net/url"
	"strconv"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

const handlerLoggerComponentName = "ThemeMgtHandler"

// themeMgtHandler is the handler for theme management operations.
type themeMgtHandler struct {
	themeMgtService ThemeMgtServiceInterface
	logger          *log.Logger
}

// newThemeMgtHandler creates a new instance of themeMgtHandler
func newThemeMgtHandler(themeMgtService ThemeMgtServiceInterface) *themeMgtHandler {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))
	return &themeMgtHandler{
		themeMgtService: themeMgtService,
		logger:          logger,
	}
}

// HandleThemeListRequest handles the list theme configurations request.
func (th *themeMgtHandler) HandleThemeListRequest(w http.ResponseWriter, r *http.Request) {
	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	themeList, svcErr := th.themeMgtService.GetThemeList(limit, offset)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	themes := make([]ThemeListItem, 0, len(themeList.Themes))
	for _, theme := range themeList.Themes {
		themes = append(themes, ThemeListItem{
			ID:          theme.ID,
			DisplayName: theme.DisplayName,
		})
	}

	themeListResponse := &ThemeListResponse{
		TotalResults: themeList.TotalResults,
		StartIndex:   themeList.StartIndex,
		Count:        themeList.Count,
		Themes:       themes,
		Links:        toHTTPLinks(themeList.Links),
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, themeListResponse)

	th.logger.Debug("Successfully listed theme configurations with pagination",
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", themeListResponse.TotalResults),
		log.Int("count", themeListResponse.Count))
}

// HandleThemePostRequest handles the create theme configuration request.
func (th *themeMgtHandler) HandleThemePostRequest(w http.ResponseWriter, r *http.Request) {
	createRequest, err := sysutils.DecodeJSONBody[CreateThemeRequest](r)
	if err != nil {
		handleError(w, &ErrorInvalidThemeData)
		return
	}

	createdTheme, svcErr := th.themeMgtService.CreateTheme(*createRequest)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	themeResponse := Theme{
		ID:          createdTheme.ID,
		DisplayName: createdTheme.DisplayName,
		Preferences: createdTheme.Preferences,
	}

	sysutils.WriteSuccessResponse(w, http.StatusCreated, themeResponse)

	th.logger.Debug("Successfully created theme configuration", log.String("id", createdTheme.ID))
}

// HandleThemeGetRequest handles the get theme configuration request.
func (th *themeMgtHandler) HandleThemeGetRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	theme, svcErr := th.themeMgtService.GetTheme(id)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	themeResponse := Theme{
		ID:          theme.ID,
		DisplayName: theme.DisplayName,
		Preferences: theme.Preferences,
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, themeResponse)

	th.logger.Debug("Successfully retrieved theme configuration", log.String("id", id))
}

// HandleThemePutRequest handles the update theme configuration request.
func (th *themeMgtHandler) HandleThemePutRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	updateRequest, err := sysutils.DecodeJSONBody[UpdateThemeRequest](r)
	if err != nil {
		handleError(w, &ErrorInvalidThemeData)
		return
	}

	updatedTheme, svcErr := th.themeMgtService.UpdateTheme(id, *updateRequest)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	themeResponse := Theme{
		ID:          updatedTheme.ID,
		DisplayName: updatedTheme.DisplayName,
		Preferences: updatedTheme.Preferences,
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, themeResponse)

	th.logger.Debug("Successfully updated theme configuration", log.String("id", id))
}

// HandleThemeDeleteRequest handles the delete theme configuration request.
func (th *themeMgtHandler) HandleThemeDeleteRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	svcErr := th.themeMgtService.DeleteTheme(id)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusNoContent, nil)
	th.logger.Debug("Successfully deleted theme configuration", log.String("id", id))
}

// parsePaginationParams parses limit and offset query parameters from the request.
func parsePaginationParams(query url.Values) (int, int, *serviceerror.ServiceError) {
	limit := 0
	offset := 0

	if limitStr := query.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err != nil {
			return 0, 0, &serviceerror.ServiceError{
				Type:             serviceerror.ClientErrorType,
				Code:             "THM-1009",
				Error:            "Invalid limit",
				ErrorDescription: "Limit must be a valid integer",
			}
		} else {
			limit = parsedLimit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err != nil {
			return 0, 0, &serviceerror.ServiceError{
				Type:             serviceerror.ClientErrorType,
				Code:             "THM-1010",
				Error:            "Invalid offset",
				ErrorDescription: "Offset must be a valid integer",
			}
		} else {
			offset = parsedOffset
		}
	}

	if limit == 0 {
		limit = serverconst.DefaultPageSize
	}

	return limit, offset, nil
}

// toHTTPLinks converts service layer Links to HTTP LinkResponses.
func toHTTPLinks(links []Link) []LinkResponse {
	httpLinks := make([]LinkResponse, len(links))
	for i, link := range links {
		httpLinks[i] = LinkResponse(link)
	}
	return httpLinks
}

// handleError handles service errors and returns appropriate HTTP responses.
func handleError(w http.ResponseWriter, svcErr *serviceerror.ServiceError) {
	statusCode := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		switch svcErr.Code {
		case ErrorThemeNotFound.Code:
			statusCode = http.StatusNotFound
		case ErrorThemeAlreadyExists.Code:
			statusCode = http.StatusConflict
		case ErrorInvalidThemeID.Code, ErrorInvalidThemeData.Code:
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusBadRequest
		}
	}

	errResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}

	sysutils.WriteErrorResponse(w, statusCode, errResp)
}
