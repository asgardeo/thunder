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

package authnprovider

// AuthnMetadata contains metadata for authentication.
type AuthnMetadata struct {
	AppMetadata map[string]interface{} `json:"appMetadata,omitempty"`
}

// AuthnResult represents the result of an authentication attempt.
type AuthnResult struct {
	UserID              string               `json:"userId"`
	UserType            string               `json:"userType"`
	OrganizationUnitID  string               `json:"ouId"`
	Token               string               `json:"token"`
	AvailableAttributes *AvailableAttributes `json:"availableAttributes"`
}

// GetAttributesMetadata contains metadata for fetching attributes.
type GetAttributesMetadata struct {
	AppMetadata map[string]interface{} `json:"appMetadata,omitempty"`
	Locale      string                 `json:"locale"`
}

// GetAttributesResult represents the result of fetching attributes.
type GetAttributesResult struct {
	UserID             string              `json:"userId"`
	UserType           string              `json:"userType"`
	OrganizationUnitID string              `json:"ouId"`
	AttributesResponse *AttributesResponse `json:"attributeResponse,omitempty"`
}

// AvailableAttributes contains the available attributes and verifications for a user.
type AvailableAttributes struct {
	Attributes    map[string]*AttributeMetadataResponse `json:"attributes,omitempty"`
	Verifications map[string]*VerificationResponse      `json:"verifications,omitempty"`
}

// AttributeMetadataResponse contains metadata for an attribute.
type AttributeMetadataResponse struct {
	AssuranceMetadataResponse *AssuranceMetadataResponse `json:"assuranceMetadataResponse,omitempty"`
}

// AssuranceMetadataResponse contains assurance metadata for an attribute.
type AssuranceMetadataResponse struct {
	IsVerified bool `json:"isVerified"`
	// this should be the key of the corresponding verification response in the verifications map
	VerificationID string `json:"verificationId,omitempty"`
}

// VerificationResponse contains verification details for an attribute.
type VerificationResponse struct {
	TrustFramework      string `json:"trustFramework,omitempty"`
	Time                string `json:"time,omitempty"`
	VerificationProcess string `json:"verificationProcess,omitempty"`
}

// RequestedAttributes contains the requested attributes and verifications.
type RequestedAttributes struct {
	Attributes    map[string]*AttributeMetadataRequest `json:"attributes,omitempty"`
	Verifications map[string]*VerificationRequest      `json:"verifications,omitempty"`
}

// AttributeMetadataRequest contains metadata request details for an attribute.
type AttributeMetadataRequest struct {
	GenericMetadataRequest   *GenericMetadataRequest   `json:"genericMetadataRequest,omitempty"`
	AssuranceMetadataRequest *AssuranceMetadataRequest `json:"assuranceMetadataRequest,omitempty"`
}

// GenericMetadataRequest contains generic metadata request details.
type GenericMetadataRequest struct {
	Essential bool     `json:"essential,omitempty"`
	Value     string   `json:"value,omitempty"`
	Values    []string `json:"values,omitempty"`
}

// GenericTimeMetadataRequest extends GenericMetadataRequest with time-related metadata.
type GenericTimeMetadataRequest struct {
	GenericMetadataRequest
	MaxAge *int `json:"maxAge,omitempty"`
}

// AssuranceMetadataRequest contains assurance metadata request details.
type AssuranceMetadataRequest struct {
	ShouldVerify bool `json:"shouldVerify,omitempty"`
	// this should be the key of the corresponding verification request in the verifications map
	VerificationID string `json:"verificationId,omitempty"`
}

// VerificationRequest contains verification request details.
type VerificationRequest struct {
	TrustFramework      *GenericMetadataRequest     `json:"trustFramework,omitempty"`
	VerificationProcess *GenericMetadataRequest     `json:"verificationProcess,omitempty"`
	Time                *GenericTimeMetadataRequest `json:"time,omitempty"`
}

// AttributesResponse contains the response with attributes and verifications.
type AttributesResponse struct {
	Attributes    map[string]*AttributeResponse    `json:"attributes,omitempty"`
	Verifications map[string]*VerificationResponse `json:"verifications,omitempty"`
}

// AttributeResponse contains the response for an attribute with its value and assurance metadata.
type AttributeResponse struct {
	Value                     interface{}                `json:"value,omitempty"`
	AssuranceMetadataResponse *AssuranceMetadataResponse `json:"assuranceMetadataResponse,omitempty"`
}
