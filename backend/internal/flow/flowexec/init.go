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

package flowexec

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/flow/executor"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	dbprovider "github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/system/observability"
)

// Initialize creates and configures the flow execution service components.
// The observabilitySvc parameter is optional (can be nil) - if nil, observability events won't be published.
func Initialize(
	mux *http.ServeMux,
	flowMgtService flowmgt.FlowMgtServiceInterface,
	applicationService application.ApplicationServiceInterface,
	executorRegistry executor.ExecutorRegistryInterface,
	observabilitySvc observability.ObservabilityServiceInterface,
) FlowExecServiceInterface {
	dbProvider := dbprovider.GetDBProvider()
	flowStore := newFlowStore(dbProvider)
	flowEngine := newFlowEngine(executorRegistry, observabilitySvc)
	flowExecService := newFlowExecService(flowMgtService, flowStore, flowEngine, applicationService, observabilitySvc)

	handler := newFlowExecutionHandler(flowExecService)
	registerRoutes(mux, handler)

	return flowExecService
}

func registerRoutes(mux *http.ServeMux, handler *flowExecutionHandler) {
	opts := middleware.CORSOptions{
		AllowedMethods:   "POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /flow/execute",
		middleware.CorrelationIDMiddleware(http.HandlerFunc(handler.HandleFlowExecutionRequest)).ServeHTTP, opts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /flow/execute",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts))
}
