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

// Package managers provides functionality for managing and registering system services.
package main

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/system/log"
	appbuilder "github.com/asgardeo/thunder/pkg/app"
)

// thunderApp holds the running application instance used for graceful shutdown.
var thunderApp *appbuilder.ThunderApp

// registerServices wires all ThunderID services onto mux using the pkg/app builder.
func registerServices(mux *http.ServeMux) func(http.Handler) http.Handler {
	app, err := appbuilder.New().Build(mux)
	if err != nil {
		log.GetLogger().Fatal("Failed to initialize services", log.Error(err))
	}
	thunderApp = app
	return app.SecurityMiddleware()
}

// unregisterServices shuts down background services during graceful shutdown.
func unregisterServices() {
	if thunderApp != nil {
		thunderApp.Shutdown()
	}
}
