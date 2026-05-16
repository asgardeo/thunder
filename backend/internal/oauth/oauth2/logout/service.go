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

package logout

import (
	"context"
	"errors"

	"github.com/thunder-id/thunderid/internal/inboundclient"
	inboundmodel "github.com/thunder-id/thunderid/internal/inboundclient/model"
	"github.com/thunder-id/thunderid/internal/system/jose/jwt"
)

// logoutServiceInterface is the internal interface used by the handler.
type logoutServiceInterface interface {
	Logout(ctx context.Context, idTokenHint, postLogoutRedirectURI string) (string, error)
}

type logoutService struct {
	jwtService     jwt.JWTServiceInterface
	inboundClients inboundclient.InboundClientServiceInterface
}

func newLogoutService(
	jwtService jwt.JWTServiceInterface,
	inboundClients inboundclient.InboundClientServiceInterface,
) logoutServiceInterface {
	return &logoutService{
		jwtService:     jwtService,
		inboundClients: inboundClients,
	}
}

// Logout validates the id_token_hint, optionally validates the post_logout_redirect_uri
// against the client's registered redirect URIs, and returns the URI to redirect to.
func (s *logoutService) Logout(
	ctx context.Context,
	idTokenHint string,
	postLogoutRedirectURI string,
) (string, error) {
	// Verify the token's signature (ignores expiry — token may already be expired at logout time).
	if svcErr := s.jwtService.VerifyJWTSignature(idTokenHint); svcErr != nil {
		return "", errors.New("invalid id_token_hint: " + svcErr.ErrorDescription.DefaultValue)
	}

	// Decode payload to extract client_id (aud) for redirect URI validation.
	payload, err := jwt.DecodeJWTPayload(idTokenHint)
	if err != nil {
		return "", errors.New("malformed id_token_hint")
	}

	clientID, _ := extractStringClaim(payload, "aud")
	if clientID == "" {
		return "", errors.New("id_token_hint missing aud claim")
	}

	if postLogoutRedirectURI == "" {
		return "", nil
	}

	// Validate post_logout_redirect_uri against the client's registered redirect URIs.
	oauthClient, err := s.inboundClients.GetOAuthClientByClientID(ctx, clientID)
	if err != nil {
		return "", errors.New("client not found")
	}

	if err := inboundmodel.ValidateRedirectURI(oauthClient.RedirectURIs, postLogoutRedirectURI); err != nil {
		return "", errors.New("post_logout_redirect_uri not registered for client")
	}

	return postLogoutRedirectURI, nil
}

// extractStringClaim extracts a string claim from a decoded JWT payload.
// The "aud" claim may be a string or a []interface{}.
func extractStringClaim(payload map[string]interface{}, key string) (string, bool) {
	v, ok := payload[key]
	if !ok {
		return "", false
	}
	switch val := v.(type) {
	case string:
		return val, true
	case []interface{}:
		if len(val) > 0 {
			if s, ok := val[0].(string); ok {
				return s, true
			}
		}
	}
	return "", false
}
