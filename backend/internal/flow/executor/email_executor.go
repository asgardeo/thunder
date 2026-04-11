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
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package executor

import (
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/email"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/template"
)

// emailExecutor sends emails based on the configured email template and runtime context data.
// When email is not configured (emailClient is nil), it completes as a no-op with emailSent=false.
type emailExecutor struct {
	core.ExecutorInterface
	logger          *log.Logger
	emailClient     email.EmailClientInterface
	templateService template.TemplateServiceInterface
}

// newEmailExecutor creates a new instance of the email executor.
// emailClient may be nil if SMTP is not configured; the executor completes as a no-op in that case.
func newEmailExecutor(flowFactory core.FlowFactoryInterface, emailClient email.EmailClientInterface,
	templateService template.TemplateServiceInterface) *emailExecutor {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "EmailExecutor"))
	base := flowFactory.CreateExecutor(
		ExecutorNameEmailExecutor,
		common.ExecutorTypeUtility,
		[]common.Input{},
		[]common.Input{
			{Identifier: userAttributeEmail, Type: common.InputTypeText, Required: true},
		},
	)
	return &emailExecutor{
		ExecutorInterface: base,
		logger:            logger,
		emailClient:       emailClient,
		templateService:   templateService,
	}
}

// Execute sends an email using the data from the runtime context.
func (e *emailExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	switch ctx.ExecutorMode {
	case ExecutorModeSend:
		return e.executeSend(ctx)
	default:
		return nil, fmt.Errorf("invalid executor mode for EmailExecutor: %s", ctx.ExecutorMode)
	}
}

// executeSend resolves the email template, constructs the email, and sends it.
// If the email client is not configured, it completes without sending (no-op).
func (e *emailExecutor) executeSend(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := e.logger.With(log.String(log.LoggerKeyExecutionID, ctx.ExecutionID))
	logger.Debug("Executing email executor in send mode")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// If email client is not configured, complete as a no-op.
	if e.emailClient == nil {
		execResp.AdditionalData[common.DataEmailSent] = dataValueFalse
		logger.Debug("Email client not configured, skipping email send")
		execResp.Status = common.ExecComplete
		return execResp, nil
	}

	if e.templateService == nil {
		return nil, errors.New("template service is not configured")
	}

	// Resolve recipient email from user inputs or runtime data.
	recipient := resolveRecipientEmail(ctx)
	if recipient == "" {
		logger.Debug("Email recipient not found in user inputs or runtime data")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Email recipient is required"
		return execResp, nil
	}

	var scenario template.ScenarioType
	if tmplProp, ok := ctx.NodeProperties[propertyKeyEmailTemplate]; ok {
		tmplStr, ok := tmplProp.(string)
		if !ok {
			return nil, fmt.Errorf("invalid type for %s: expected string, got %T with value %v",
				propertyKeyEmailTemplate, tmplProp, tmplProp)
		}
		if tmplStr == "" {
			scenario = template.ScenarioUserInvite
		} else {
			scenario = template.ScenarioType(tmplStr)
		}
	} else {
		scenario = template.ScenarioUserInvite
	}

	inviteLink := ctx.RuntimeData[common.RuntimeKeyInviteLink]
	if (scenario == template.ScenarioUserInvite || scenario == template.ScenarioSelfRegistration) && inviteLink == "" {
		return nil, errors.New("invite link not found in runtime data")
	}

	templateData := template.TemplateData{
		"inviteLink": inviteLink,
		"appName":    ctx.Application.Name,
	}

	rendered, svcErr := e.templateService.Render(ctx.Context, scenario, template.TemplateTypeEmail, templateData)
	if svcErr != nil {
		return nil, fmt.Errorf("failed to render email template: %s", svcErr.Code)
	}

	emailData := email.EmailData{
		To:      []string{recipient},
		Subject: rendered.Subject,
		Body:    rendered.Body,
		IsHTML:  rendered.IsHTML,
	}

	if err := e.emailClient.Send(emailData); err != nil {
		if isEmailClientError(err) {
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "Failed to send email"
			return execResp, nil
		}
		return nil, fmt.Errorf("email send failed: %w", err)
	}

	logger.Debug("Email sent successfully",
		log.String("recipient", log.MaskString(recipient)))

	execResp.AdditionalData[common.DataEmailSent] = dataValueTrue
	execResp.Status = common.ExecComplete
	return execResp, nil
}

// resolveRecipientEmail retrieves the recipient email from user inputs or runtime data.
func resolveRecipientEmail(ctx *core.NodeContext) string {
	if recipientEmail, ok := ctx.UserInputs[userAttributeEmail]; ok && recipientEmail != "" {
		return recipientEmail
	}
	if recipientEmail, ok := ctx.RuntimeData[userAttributeEmail]; ok && recipientEmail != "" {
		return recipientEmail
	}
	return ""
}

// isEmailClientError returns true if the error is a client-side validation error
// (e.g., invalid recipient, invalid sender) rather than a server-side transport error.
func isEmailClientError(err error) bool {
	return errors.Is(err, email.ErrorInvalidRecipient) ||
		errors.Is(err, email.ErrorInvalidSender) ||
		errors.Is(err, email.ErrorInvalidSubject) ||
		errors.Is(err, email.ErrorInvalidHost) ||
		errors.Is(err, email.ErrorInvalidPort) ||
		errors.Is(err, email.ErrorInvalidCredentials)
}
