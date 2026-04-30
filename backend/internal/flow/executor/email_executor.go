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

	"github.com/asgardeo/thunder/internal/entityprovider"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/email"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/template"
)

// emailExecutor sends emails based on the configured email template and runtime context data.
type emailExecutor struct {
	core.ExecutorInterface
	logger          *log.Logger
	emailClient     email.EmailClientInterface
	templateService template.TemplateServiceInterface
	entityProvider  entityprovider.EntityProviderInterface
}

// newEmailExecutor creates a new instance of the email executor.
func newEmailExecutor(flowFactory core.FlowFactoryInterface, emailClient email.EmailClientInterface,
	templateService template.TemplateServiceInterface,
	entityProvider entityprovider.EntityProviderInterface) *emailExecutor {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "EmailExecutor"))
	base := flowFactory.CreateExecutor(
		ExecutorNameEmailExecutor,
		common.ExecutorTypeUtility,
		[]common.Input{
			{Identifier: userAttributeEmail, Type: common.InputTypeEmail, Required: true},
		},
		[]common.Input{},
	)
	return &emailExecutor{
		ExecutorInterface: base,
		logger:            logger,
		emailClient:       emailClient,
		templateService:   templateService,
		entityProvider:    entityProvider,
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
func (e *emailExecutor) executeSend(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := e.logger.With(log.String(log.LoggerKeyExecutionID, ctx.ExecutionID))
	logger.Debug("Executing email executor in send mode")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// 1. Check for SkipDelivery FIRST
	if skip, ok := ctx.RuntimeData[common.RuntimeKeySkipDelivery]; ok && skip == dataValueTrue {
		logger.Debug("Delivery marked as skipped, completing without sending email")
		execResp.Status = common.ExecComplete
		return execResp, nil
	}

	if e.emailClient == nil {
		execResp.AdditionalData[common.DataEmailSent] = dataValueFalse
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Email service is not configured"
		logger.Debug("Email client not configured")
		return execResp, nil
	}

	if e.templateService == nil {
		return nil, errors.New("template service is not configured")
	}

	// 2. Resolve recipient email properly
	recipient, err := e.resolveRecipientEmail(ctx, logger)
	if err != nil {
		return nil, err
	}
	if recipient == "" {
		logger.Debug("Email recipient not found")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Email recipient is required"
		return execResp, nil
	}

	// 3. Fail-Fast Template Resolution (No guessing!)
	var scenario template.ScenarioType
	if tmplProp, ok := ctx.NodeProperties[propertyKeyEmailTemplate]; ok {
		tmplStr, ok := tmplProp.(string)
		if !ok {
			return nil, fmt.Errorf("invalid type for %s: expected string, got %T with value %v",
				propertyKeyEmailTemplate, tmplProp, tmplProp)
		}
		if tmplStr == "" {
			return nil, fmt.Errorf("email template property is empty in node configuration")
		}
		scenario = template.ScenarioType(tmplStr)
	} else {
		return nil, fmt.Errorf("missing required property: %s", propertyKeyEmailTemplate)
	}

	// 4. Generic Template Data (Supports Invites AND Magic Links)
	templateData := e.resolveTemplateData(ctx)

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

	// Graceful failure for all client and system email transport errors
	if err := e.emailClient.Send(emailData); err != nil {
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Failed to send email"
		return execResp, nil
	}

	logger.Debug("Email sent successfully", log.MaskedString("recipient", recipient))

	execResp.AdditionalData[common.DataEmailSent] = dataValueTrue
	execResp.Status = common.ExecComplete
	return execResp, nil
}

// resolveRecipientEmail retrieves the recipient email from user inputs, runtime data, or forwarded data.
func (e *emailExecutor) resolveRecipientEmail(ctx *core.NodeContext, logger *log.Logger) (string, error) {
	// Strict mode: Fail immediately if the flow canvas didn't configure the email input.
	input, ok := findInputByType(ctx.NodeInputs, common.InputTypeEmail)
	if !ok {
		return "", errors.New("email input configuration is missing from node inputs")
	}
	emailAttr := input.Identifier

	if recipientEmail, ok := ctx.ForwardedData[emailAttr]; ok {
		if emailStr, isString := recipientEmail.(string); isString && emailStr != "" {
			return emailStr, nil
		}
	}
	if recipientEmail, ok := ctx.RuntimeData[emailAttr]; ok && recipientEmail != "" {
		return recipientEmail, nil
	}
	if recipientEmail, ok := ctx.UserInputs[emailAttr]; ok && recipientEmail != "" {
		return recipientEmail, nil
	}

	if userID, ok := ctx.RuntimeData[userAttributeUserID]; ok && userID != "" {
		if e.entityProvider == nil {
			return "", errors.New("entity provider is not configured for email resolution")
		}
		user, providerErr := e.entityProvider.GetEntity(userID)
		if providerErr != nil {
			if providerErr.Code == entityprovider.ErrorCodeEntityNotFound {
				return "", nil
			}
			return "", fmt.Errorf("failed to fetch user from entity provider: %w", providerErr)
		}

		if recipientEmail, err := GetUserAttribute(user, emailAttr); err == nil {
			return recipientEmail, nil
		}
		logger.Debug("Email attribute not found in user entity", log.String("attribute", emailAttr))
	}

	return "", nil
}

// resolveTemplateData extracts template data from RuntimeData, Context, and ForwardedData.
func (e *emailExecutor) resolveTemplateData(ctx *core.NodeContext) template.TemplateData {
	templateData := template.TemplateData{}

	// 1. PRIMARY: Dump the persistent flow state into the template.
	if ctx.RuntimeData != nil {
		for k, v := range ctx.RuntimeData {
			templateData[k] = fmt.Sprintf("%v", v)
		}
	}

	// 2. SYSTEM: Always provide the app name
	if ctx.Application.Name != "" {
		templateData["appName"] = ctx.Application.Name
	}

	// 3. FALLBACK: Check ForwardedData (with safe type switching)
	if ctx.ForwardedData != nil {
		if forwardedTemplateData, ok := ctx.ForwardedData[common.ForwardedDataKeyTemplateData]; ok {
			switch data := forwardedTemplateData.(type) {
			case map[string]interface{}:
				for k, v := range data {
					templateData[k] = fmt.Sprintf("%v", v)
				}
			case map[string]string:
				for k, v := range data {
					templateData[k] = v
				}
			default:
				e.logger.Debug("Forwarded template data is of unknown type",
					log.String("type", fmt.Sprintf("%T", forwardedTemplateData)))
			}
		}
	}

	return templateData
}
