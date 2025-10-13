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

package notification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/cmodels"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// messageNotificationSenderHandler handles HTTP requests for message notification sender management
type messageNotificationSenderHandler struct {
	mgtService NotificationSenderMgtSvcInterface
	otpService OTPServiceInterface
}

// newMessageNotificationSenderHandler creates a new instance of MessageNotificationSenderHandler
func newMessageNotificationSenderHandler(
	mgtService NotificationSenderMgtSvcInterface,
	otpService OTPServiceInterface) *messageNotificationSenderHandler {
	return &messageNotificationSenderHandler{
		mgtService: mgtService,
		otpService: otpService,
	}
}

// HandleSenderListRequest handles the request to list all message notification senders
func (h *messageNotificationSenderHandler) HandleSenderListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationHandler"))

	senders, svcErr := h.mgtService.ListSenders()
	if svcErr != nil {
		h.handleError(w, logger, svcErr, "")
		return
	}

	senderResponses := make([]common.NotificationSenderResponse, 0, len(senders))
	for _, sender := range senders {
		senderResponse, err := getSenderResponseFromDTO(&sender)
		if err != nil {
			logger.Error("Failed to convert sender to response", log.String("sender", sender.Name), log.Error(err))
			h.handleError(w, logger, &ErrorInternalServerError, "Failed to convert sender to response: "+err.Error())
			return
		}
		senderResponses = append(senderResponses, senderResponse)
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(senderResponses); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleSenderCreateRequest handles the request to create a new message notification sender
func (h *messageNotificationSenderHandler) HandleSenderCreateRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationHandler"))

	sender, err := sysutils.DecodeJSONBody[common.NotificationSenderRequest](r)
	if err != nil {
		h.handleError(w, logger, &ErrorInvalidRequestFormat, "Failed to parse request body: "+err.Error())
		return
	}

	senderDTO, err := getDTOFromSenderRequest(sender)
	if err != nil {
		logger.Error("Failed to process sender request", log.Error(err))
		h.handleError(w, logger, &ErrorInternalServerError, "Failed to process sender request: "+err.Error())
		return
	}

	createdSender, svcErr := h.mgtService.CreateSender(*senderDTO)
	if svcErr != nil {
		if svcErr.Code == ErrorDuplicateSenderName.Code {
			w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
			w.WriteHeader(http.StatusConflict)

			errResp := apierror.ErrorResponse{
				Code:        svcErr.Code,
				Message:     svcErr.Error,
				Description: svcErr.ErrorDescription,
			}

			if err := json.NewEncoder(w).Encode(errResp); err != nil {
				logger.Error("Error encoding error response", log.Error(err))
				http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
			}
			return
		}

		h.handleError(w, logger, svcErr, "")
		return
	}

	senderResponse, err := getSenderResponseFromDTO(createdSender)
	if err != nil {
		logger.Error("Failed to convert sender to response", log.String("sender", createdSender.Name), log.Error(err))
		h.handleError(w, logger, &ErrorInternalServerError, "Failed to convert sender to response: "+err.Error())
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(senderResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleSenderGetRequest handles the request to get a message notification sender by ID
func (h *messageNotificationSenderHandler) HandleSenderGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationHandler"))

	id := r.PathValue("id")
	if !h.validateSenderID(w, id) {
		return
	}

	sender, svcErr := h.mgtService.GetSender(id)
	if svcErr != nil {
		h.handleError(w, logger, svcErr, "")
		return
	}
	if sender == nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusNotFound)
		errResp := apierror.ErrorResponse{
			Code:        ErrorSenderNotFound.Code,
			Message:     ErrorSenderNotFound.Error,
			Description: ErrorSenderNotFound.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
			return
		}
		return
	}

	senderResponse, err := getSenderResponseFromDTO(sender)
	if err != nil {
		logger.Error("Failed to convert sender to response", log.String("sender", sender.Name), log.Error(err))
		h.handleError(w, logger, &ErrorInternalServerError, "Failed to convert sender to response: "+err.Error())
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(senderResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleSenderUpdateRequest handles the request to update a message notification sender
func (h *messageNotificationSenderHandler) HandleSenderUpdateRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationHandler"))

	id := r.PathValue("id")
	if !h.validateSenderID(w, id) {
		return
	}

	sender, err := sysutils.DecodeJSONBody[common.NotificationSenderRequest](r)
	if err != nil {
		h.handleError(w, logger, &ErrorInvalidRequestFormat, "Failed to parse request body: "+err.Error())
		return
	}

	senderDTO, err := getDTOFromSenderRequest(sender)
	if err != nil {
		logger.Error("Failed to process sender request", log.Error(err))
		h.handleError(w, logger, &ErrorInternalServerError, "Failed to process sender request: "+err.Error())
		return
	}

	updatedSender, svcErr := h.mgtService.UpdateSender(id, *senderDTO)
	if svcErr != nil {
		h.handleError(w, logger, svcErr, "")
		return
	}

	senderResponse, err := getSenderResponseFromDTO(updatedSender)
	if err != nil {
		logger.Error("Failed to convert sender to response", log.String("sender", updatedSender.Name), log.Error(err))
		h.handleError(w, logger, &ErrorInternalServerError, "Failed to convert sender to response: "+err.Error())
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(senderResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleSenderDeleteRequest handles the request to delete a message notification sender
func (h *messageNotificationSenderHandler) HandleSenderDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationHandler"))

	id := r.PathValue("id")
	if !h.validateSenderID(w, id) {
		return
	}

	svcErr := h.mgtService.DeleteSender(id)
	if svcErr != nil {
		h.handleError(w, logger, svcErr, "")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleOTPSendRequest handles the request to send an OTP.
func (h *messageNotificationSenderHandler) HandleOTPSendRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationHandler"))

	request, err := sysutils.DecodeJSONBody[common.SendOTPRequest](r)
	if err != nil {
		h.handleError(w, logger, &ErrorInvalidRequestFormat, "Failed to parse request body: "+err.Error())
		return
	}

	otpDTO := common.SendOTPDTO(*request)
	resultDTO, svcErr := h.otpService.SendOTP(otpDTO)
	if svcErr != nil {
		h.handleError(w, logger, svcErr, "")
		return
	}

	otpResponse := common.SendOTPResponse{
		Status:       "SUCCESS",
		SessionToken: resultDTO.SessionToken,
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(otpResponse); err != nil {
		logger.Error("Failed to encode response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleOTPVerifyRequest handles the request to verify an OTP.
func (h *messageNotificationSenderHandler) HandleOTPVerifyRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationHandler"))

	request, err := sysutils.DecodeJSONBody[common.VerifyOTPRequest](r)
	if err != nil {
		h.handleError(w, logger, &ErrorInvalidRequestFormat, "Failed to parse request body: "+err.Error())
		return
	}

	verifyDTO := common.VerifyOTPDTO(*request)
	resultDTO, svcErr := h.otpService.VerifyOTP(verifyDTO)
	if svcErr != nil {
		h.handleError(w, logger, svcErr, "")
		return
	}

	response := common.VerifyOTPResponse{
		Status: string(resultDTO.Status),
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleError handles service errors and returns appropriate HTTP responses.
func (h *messageNotificationSenderHandler) handleError(w http.ResponseWriter, logger *log.Logger,
	svcErr *serviceerror.ServiceError, customErrDesc string) {
	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)

	apiErrDesc := svcErr.ErrorDescription
	if customErrDesc != "" {
		apiErrDesc = customErrDesc
	}
	errResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: apiErrDesc,
	}

	statusCode := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		switch svcErr.Code {
		case ErrorSenderNotFound.Code:
			statusCode = http.StatusNotFound
		case ErrorDuplicateSenderName.Code:
			statusCode = http.StatusConflict
		default:
			statusCode = http.StatusBadRequest
		}
	}
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		logger.Error("Error encoding error response", log.Error(err))
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// validateSenderID validates the sender ID and returns true if valid
func (h *messageNotificationSenderHandler) validateSenderID(w http.ResponseWriter, id string) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "NotificationHandler"))

	if strings.TrimSpace(id) == "" {
		h.handleError(w, logger, &ErrorInvalidSenderID, "Sender ID is required")
		return false
	}
	return true
}

// getDTOFromSenderRequest sanitizes the sender request and converts it to a NotificationSenderDTO.
func getDTOFromSenderRequest(sender *common.NotificationSenderRequest) (*common.NotificationSenderDTO, error) {
	name := sysutils.SanitizeString(sender.Name)
	description := sysutils.SanitizeString(sender.Description)
	providerStr := sysutils.SanitizeString(sender.Provider)

	// Sanitize properties
	properties := make([]cmodels.Property, 0, len(sender.Properties))
	for _, propDTO := range sender.Properties {
		sanitizedDTO := cmodels.PropertyDTO{
			Name:     sysutils.SanitizeString(propDTO.Name),
			Value:    sysutils.SanitizeString(propDTO.Value),
			IsSecret: propDTO.IsSecret,
		}
		property, err := sanitizedDTO.ToProperty()
		if err != nil {
			return nil, fmt.Errorf("failed to create property %s: %w", propDTO.Name, err)
		}
		properties = append(properties, *property)
	}

	senderDTO := common.NotificationSenderDTO{
		Name:        name,
		Description: description,
		Type:        common.NotificationSenderTypeMessage,
		Provider:    common.MessageProviderType(providerStr),
		Properties:  properties,
	}
	return &senderDTO, nil
}

// getSenderResponseFromDTO converts a NotificationSenderDTO to a response object, masking secret properties.
func getSenderResponseFromDTO(sender *common.NotificationSenderDTO) (common.NotificationSenderResponse, error) {
	returnSender := common.NotificationSenderResponse{
		ID:          sender.ID,
		Name:        sender.Name,
		Description: sender.Description,
		Provider:    sender.Provider,
	}

	// Mask secret properties in the response.
	senderProperties := make([]cmodels.PropertyDTO, 0, len(sender.Properties))
	for _, property := range sender.Properties {
		if property.IsSecret() {
			maskedProperty := &cmodels.PropertyDTO{
				Name:     property.GetName(),
				Value:    "******",
				IsSecret: property.IsSecret(),
			}
			senderProperties = append(senderProperties, *maskedProperty)
		} else {
			propertyDTO, err := property.ToPropertyDTO()
			if err != nil {
				return common.NotificationSenderResponse{}, fmt.Errorf("failed to convert property %s: %w", property.GetName(), err)
			}
			senderProperties = append(senderProperties, *propertyDTO)
		}
	}
	returnSender.Properties = senderProperties

	return returnSender, nil
}
