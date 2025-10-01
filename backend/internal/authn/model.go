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

package authn

import "github.com/asgardeo/thunder/internal/idp"

// IDPAuthInitDTO represents the data transfer object for initiating IDP authentication.
type IDPAuthInitDTO struct {
	RedirectURL  string
	SessionToken string
}

// IDPAuthFinishDTO represents the data transfer object for completing IDP authentication.
type IDPAuthFinishDTO struct {
	ID               string
	Type             string
	OrganizationUnit string
}

// AuthSessionData represents the data stored in the authentication session token.
type AuthSessionData struct {
	IDPID   string      `json:"idp_id"`
	IDPType idp.IDPType `json:"idp_type"`
}

// IDPAuthInitRequest is the request to initiate IDP authentication.
type IDPAuthInitRequest struct {
	IDPID string `json:"idp_id"`
}

// IDPAuthInitResponse is the response for IDP authentication initiation.
type IDPAuthInitResponse struct {
	RedirectURL  string `json:"redirect_url,omitempty"`
	SessionToken string `json:"session_token"`
}

// IDPAuthFinishRequest is the request to complete IDP authentication.
type IDPAuthFinishRequest struct {
	SessionToken string `json:"session_token"`
	Code         string `json:"code"`
}

// IDPAuthFinishResponse is the response for completed IDP authentication.
type IDPAuthFinishResponse struct {
	ID               string `json:"id"`
	Type             string `json:"type,omitempty"`
	OrganizationUnit string `json:"organization_unit,omitempty"`
}
