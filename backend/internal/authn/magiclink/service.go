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

// Package magiclink implements the magic link authentication service.
package magiclink

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/email"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/template"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userprovider"
)

// MagicLinkAuthnServiceInterface defines the interface for magic link authentication operations.
type MagicLinkAuthnServiceInterface interface {
	SendMagicLink(
		ctx context.Context,
		recipient string,
		expirySeconds int64,
		flowID string,
		magicLinkURL *string,
	) *serviceerror.I18nServiceError
	VerifyMagicLink(ctx context.Context, token string) (*userprovider.User, *serviceerror.I18nServiceError)
}

// magicLinkAuthnService is the default implementation of MagicLinkAuthnServiceInterface.
type magicLinkAuthnService struct {
	jwtService      jwt.JWTServiceInterface
	emailClient     email.EmailClientInterface
	userProvider    userprovider.UserProviderInterface
	templateService template.TemplateServiceInterface
	logger          *log.Logger
}

func newMagicLinkAuthnService(
	jwtSvc jwt.JWTServiceInterface,
	emailClient email.EmailClientInterface,
	userProvider userprovider.UserProviderInterface,
	templateService template.TemplateServiceInterface,
) MagicLinkAuthnServiceInterface {
	service := &magicLinkAuthnService{
		jwtService:      jwtSvc,
		emailClient:     emailClient,
		userProvider:    userProvider,
		templateService: templateService,
		logger:          log.GetLogger().With(log.String(log.LoggerKeyComponentName, "MagicLinkAuthnService")),
	}
	common.RegisterAuthenticator(service.getMetadata())

	return service
}

func (s *magicLinkAuthnService) SendMagicLink(ctx context.Context,
	recipient string, expirySeconds int64, flowID string, magicLinkURL *string) *serviceerror.I18nServiceError {
	s.logger.Debug("Sending magic link", log.String("recipient", log.MaskString(recipient)))

	if strings.ContainsAny(recipient, "\r\n") {
		return &ErrorInvalidRecipient
	}

	recipient = strings.TrimSpace(recipient)
	if recipient == "" || !email.IsValidEmail(recipient) {
		return &ErrorInvalidRecipient
	}

	// Privacy-preserving: return success without sending for non-existent users
	// to prevent user enumeration attacks.
	userID, upErr := s.userProvider.IdentifyUser(map[string]interface{}{userAttributeEmail: recipient})
	if upErr != nil {
		if upErr.Code == userprovider.ErrorCodeUserNotFound {
			s.logger.Debug("User not found for recipient, returning success without sending email",
				log.String("recipient", log.MaskString(recipient)))
			return nil
		}
		return &serviceerror.InternalServerErrorWithI18n
	}

	if userID == nil || *userID == "" {
		s.logger.Debug("No user found for recipient, returning success without sending email",
			log.String("recipient", log.MaskString(recipient)))
		return nil
	}

	issuer := config.GetThunderRuntime().Config.JWT.Issuer
	token, _, jwtErr := s.jwtService.GenerateJWT(
		*userID,
		tokenAudience,
		issuer,
		expirySeconds,
		nil,
		jwt.TokenTypeJWT,
	)
	if jwtErr != nil {
		return &ErrorTokenGenerationFailed
	}

	verifyURL := s.buildMagicLinkURL(flowID, token, magicLinkURL)

	expiryMinutes := utils.SecondsToMinutes(expirySeconds)
	templateData := template.TemplateData{
		"magicLink":     verifyURL,
		"expiryMinutes": expiryMinutes,
	}

	rendered, svcErr := s.templateService.Render(ctx, template.ScenarioMagicLink, templateData)
	if svcErr != nil {
		return &serviceerror.InternalServerErrorWithI18n
	}

	if s.emailClient == nil {
		return &serviceerror.InternalServerErrorWithI18n
	}

	sendErr := s.emailClient.Send(email.EmailData{
		To:      []string{recipient},
		Subject: rendered.Subject,
		Body:    rendered.Body,
		IsHTML:  rendered.IsHTML,
	})
	if sendErr != nil {
		return &serviceerror.InternalServerErrorWithI18n
	}

	s.logger.Debug("Magic link sent successfully",
		log.String("recipient", log.MaskString(recipient)),
		log.String("userID", *userID),
		log.String("expiryMinutes", expiryMinutes))

	return nil
}

// VerifyMagicLink verifies the validity of a magic link token and retrieves the associated user information.
// Returns a user object on success or a localized service error if the token is invalid, expired, or malformed.
func (s *magicLinkAuthnService) VerifyMagicLink(_ context.Context,
	token string) (*userprovider.User, *serviceerror.I18nServiceError) {
	s.logger.Debug("Verifying magic link token")

	token = strings.TrimSpace(token)
	if token == "" {
		return nil, &ErrorInvalidToken
	}

	issuer := config.GetThunderRuntime().Config.JWT.Issuer
	verifyErr := s.jwtService.VerifyJWT(token, tokenAudience, issuer)
	if verifyErr != nil {
		if verifyErr.Code == jwt.ErrorTokenExpired.Code {
			return nil, &ErrorExpiredToken
		}
		s.logger.Debug("Invalid magic link token", log.String("errorCode", verifyErr.Code))
		return nil, &ErrorInvalidToken
	}

	payload, decodeErr := jwt.DecodeJWTPayload(token)
	if decodeErr != nil {
		s.logger.Debug("Failed to decode magic link token payload", log.Error(decodeErr))
		return nil, &ErrorInvalidToken
	}

	userID := utils.ConvertInterfaceValueToString(payload["sub"])
	if userID == "" {
		s.logger.Debug("User ID claim not found or invalid")
		return nil, &ErrorMalformedTokenClaims
	}

	user, upErr := s.userProvider.GetUser(userID)
	if upErr != nil {
		return nil, s.handleUserProviderError(upErr)
	}

	s.logger.Debug("Magic link verification successful", log.String("userId", user.UserID))
	return user, nil
}

// buildMagicLinkURL constructs a magic link URL by appending query parameters to a base URL or default configuration.
func (s *magicLinkAuthnService) buildMagicLinkURL(flowID string, token string, magicLinkURL *string) string {
	if magicLinkURL != nil && strings.TrimSpace(*magicLinkURL) != "" {
		u, err := url.Parse(strings.TrimSpace(*magicLinkURL))
		if err == nil {
			q := u.Query()
			q.Set("flowId", flowID)
			q.Set("code", token)
			u.RawQuery = q.Encode()

			return u.String()
		}
		s.logger.Debug("Failed to parse magic link URL; using default GateClient configuration")
	}

	gateConfig := config.GetThunderRuntime().Config.GateClient
	loginPath := gateConfig.LoginPath
	if strings.TrimSpace(loginPath) == "" {
		loginPath = "/signin"
	}

	u := &url.URL{
		Scheme: gateConfig.Scheme,
		Host:   fmt.Sprintf("%s:%d", gateConfig.Hostname, gateConfig.Port),
		Path:   loginPath,
	}

	q := u.Query()
	q.Set("flowId", flowID)
	q.Set("code", token)
	u.RawQuery = q.Encode()

	return u.String()
}

func (s *magicLinkAuthnService) handleUserProviderError(
	upErr *userprovider.UserProviderError,
) *serviceerror.I18nServiceError {
	if upErr.Code == userprovider.ErrorCodeUserNotFound {
		return &common.ErrorUserNotFoundWithI18n
	}
	if upErr.Code == userprovider.ErrorCodeSystemError {
		return &serviceerror.InternalServerErrorWithI18n
	}
	s.logger.Debug("User provider returned an error while resolving user",
		log.String("description", upErr.Description))
	return &ErrorClientErrorWhileResolvingUser
}

func (s *magicLinkAuthnService) getMetadata() common.AuthenticatorMeta {
	return common.AuthenticatorMeta{
		Name:    common.AuthenticatorMagicLink,
		Factors: []common.AuthenticationFactor{common.FactorPossession},
	}
}
