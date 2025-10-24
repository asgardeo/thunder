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

package dcr

import (
	"encoding/json"
	"net/http"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// dcrHandler defines the handler for DCR API requests.
type dcrHandler struct {
	dcrService DCRServiceInterface
}

// newDCRHandler creates a new instance of dcrHandler.
func newDCRHandler(dcrService DCRServiceInterface) *dcrHandler {
	return &dcrHandler{
		dcrService: dcrService,
	}
}

// HandleDCRRegistration handles the DCR client registration request.
func (dh *dcrHandler) HandleDCRRegistration(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "DCRHandler"))

	dcrRequest, err := sysutils.DecodeJSONBody[DCRRegistrationRequest](r)
	if err != nil {
		sysutils.WriteJSONError(w, ErrorInvalidRequestFormat.Code,
			ErrorInvalidRequestFormat.ErrorDescription, http.StatusBadRequest, nil)
		return
	}

	dcrResponse, svcErr := dh.dcrService.RegisterClient(dcrRequest)
	if svcErr != nil {
		dh.writeServiceErrorResponse(w, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)

	if encodeErr := json.NewEncoder(w).Encode(dcrResponse); encodeErr != nil {
		logger.Error("Error encoding DCR response", log.Error(encodeErr))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// writeServiceErrorResponse writes a service error response.
func (dh *dcrHandler) writeServiceErrorResponse(w http.ResponseWriter, svcErr *serviceerror.ServiceError) {
	var statusCode int

	switch svcErr.Type {
	case serviceerror.ClientErrorType:
		statusCode = http.StatusBadRequest
	case serviceerror.ServerErrorType:
		statusCode = http.StatusInternalServerError
	default:
		statusCode = http.StatusBadRequest
	}

	sysutils.WriteJSONError(w, svcErr.Code, svcErr.ErrorDescription, statusCode, nil)
}
