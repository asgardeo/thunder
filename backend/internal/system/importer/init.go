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
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package importer

import (
	"net/http"

	layoutmgt "github.com/asgardeo/thunder/internal/design/layout/mgt"
	thememgt "github.com/asgardeo/thunder/internal/design/theme/mgt"
	"github.com/asgardeo/thunder/internal/entitytype"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/resource"
	"github.com/asgardeo/thunder/internal/role"
	i18nmgt "github.com/asgardeo/thunder/internal/system/i18n/mgt"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/pkg/application"
	flowmgt "github.com/asgardeo/thunder/pkg/flow/mgt"
)

// Initialize wires the importer service and registers its HTTP routes.
func Initialize(
	mux *http.ServeMux,
	applicationService application.ApplicationServiceInterface,
	idpService idp.IDPServiceInterface,
	flowService flowmgt.FlowMgtServiceInterface,
	ouService ou.OrganizationUnitServiceInterface,
	entityTypeService entitytype.EntityTypeServiceInterface,
	roleService role.RoleServiceInterface,
	resourceService resource.ResourceServiceInterface,
	themeService thememgt.ThemeMgtServiceInterface,
	layoutService layoutmgt.LayoutMgtServiceInterface,
	userService user.UserServiceInterface,
	translationService i18nmgt.I18nServiceInterface,
) ImportServiceInterface {
	importService := newImportService(
		applicationService,
		idpService,
		flowService,
		ouService,
		entityTypeService,
		roleService,
		resourceService,
		themeService,
		layoutService,
		userService,
		translationService,
	)
	importHandler := newImportHandler(importService)

	registerRoutes(mux, importHandler)

	return importService
}

func registerRoutes(mux *http.ServeMux, importHandler *importHandler) {
	opts := middleware.CORSOptions{
		AllowedMethods:   []string{"POST"},
		AllowedHeaders:   middleware.DefaultAllowedHeaders,
		AllowCredentials: true,
		MaxAge:           600,
	}

	mux.HandleFunc(middleware.WithCORS("POST /import",
		importHandler.HandleImportRequest, opts))

	mux.HandleFunc(middleware.WithCORS("OPTIONS /import",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts))

	mux.HandleFunc(middleware.WithCORS("POST /import/delete",
		importHandler.HandleDeleteImportRequest, opts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /import/delete",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts))
}
