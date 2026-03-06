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

package consent

import (
	"context"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// ConsentServiceInterface defines the contract for consent management operations.
type ConsentServiceInterface interface {
	// IsEnabled reports whether the consent service is active
	IsEnabled() bool

	// CreateConsentElements creates one or more consent elements
	CreateConsentElements(ctx context.Context, ouID string, elements []ConsentElementInput) (
		[]ConsentElement, *serviceerror.I18nServiceError)

	// ListConsentElements retrieves consent elements, optionally filtered by namespace and name
	ListConsentElements(ctx context.Context, ouID string, ns Namespace, nameFilter string) (
		[]ConsentElement, *serviceerror.I18nServiceError)

	// UpdateConsentElement updates an existing consent element by ID
	UpdateConsentElement(ctx context.Context, ouID, elementID string, element *ConsentElementInput) (
		*ConsentElement, *serviceerror.I18nServiceError)

	// DeleteConsentElement deletes a consent element by ID (idempotent)
	DeleteConsentElement(ctx context.Context, ouID, elementID string) *serviceerror.I18nServiceError

	// ValidateConsentElements validates a list of consent element names and returns the valid ones
	ValidateConsentElements(ctx context.Context, ouID string, names []string) (
		[]string, *serviceerror.I18nServiceError)

	// CreateConsentPurpose creates a consent purpose for a resource
	CreateConsentPurpose(ctx context.Context, ouID string, purpose *ConsentPurposeInput) (
		*ConsentPurpose, *serviceerror.I18nServiceError)

	// ListConsentPurposes retrieves consent purposes for a resource
	ListConsentPurposes(ctx context.Context, ouID, groupID string) (
		[]ConsentPurpose, *serviceerror.I18nServiceError)

	// UpdateConsentPurpose updates an existing consent purpose
	UpdateConsentPurpose(ctx context.Context, ouID, purposeID string,
		purpose *ConsentPurposeInput) (*ConsentPurpose, *serviceerror.I18nServiceError)

	// DeleteConsentPurpose deletes a consent purpose by ID (idempotent)
	DeleteConsentPurpose(ctx context.Context, ouID, purposeID string) *serviceerror.I18nServiceError

	// CreateConsent creates a new consent record
	CreateConsent(ctx context.Context, ouID string, consent *ConsentRequest) (
		*Consent, *serviceerror.I18nServiceError)

	// SearchConsents searches consent records matching the filter
	SearchConsents(ctx context.Context, ouID string, filter *ConsentSearchFilter) (
		[]Consent, *serviceerror.I18nServiceError)

	// ValidateConsent validates a consent by ID and returns the validation result
	ValidateConsent(ctx context.Context, ouID string, consentID string) (
		*ConsentValidationResult, *serviceerror.I18nServiceError)

	// RevokeConsent revokes an active consent record (idempotent)
	RevokeConsent(ctx context.Context, ouID string, consentID string,
		payload *ConsentRevokeRequest) *serviceerror.I18nServiceError
}

// consentClientInterface defines the contract for pluggable consent client implementations.
type consentClientInterface interface {
	// createConsentElements creates one or more consent elements
	createConsentElements(ctx context.Context, ouID string, elements []ConsentElementInput) (
		[]ConsentElement, *serviceerror.I18nServiceError)

	// listConsentElements retrieves consent elements, optionally filtered by name
	listConsentElements(ctx context.Context, ouID string, ns Namespace, nameFilter string) (
		[]ConsentElement, *serviceerror.I18nServiceError)

	// updateConsentElement updates an existing consent element by ID
	updateConsentElement(ctx context.Context, ouID, elementID string, element *ConsentElementInput) (
		*ConsentElement, *serviceerror.I18nServiceError)

	// deleteConsentElement deletes a consent element by ID
	deleteConsentElement(ctx context.Context, ouID, elementID string) *serviceerror.I18nServiceError

	// validateConsentElements validates a list of consent element names and returns the valid ones
	validateConsentElements(ctx context.Context, ouID string, names []string) (
		[]string, *serviceerror.I18nServiceError)

	// createConsentPurpose creates a consent purpose for a resource
	createConsentPurpose(ctx context.Context, ouID string, purpose *ConsentPurposeInput) (
		*ConsentPurpose, *serviceerror.I18nServiceError)

	// listConsentPurposes retrieves consent purposes for a resource
	listConsentPurposes(ctx context.Context, ouID, groupID string) (
		[]ConsentPurpose, *serviceerror.I18nServiceError)

	// updateConsentPurpose updates an existing consent purpose
	updateConsentPurpose(ctx context.Context, ouID, purposeID string, purpose *ConsentPurposeInput) (
		*ConsentPurpose, *serviceerror.I18nServiceError)

	// deleteConsentPurpose deletes a consent purpose by ID
	deleteConsentPurpose(ctx context.Context, ouID, purposeID string) *serviceerror.I18nServiceError

	// createConsent creates a new consent record and returns the created consent with ID
	createConsent(ctx context.Context, ouID string, req *ConsentRequest) (
		*Consent, *serviceerror.I18nServiceError)

	// searchConsents searches consent records matching the filter
	searchConsents(ctx context.Context, ouID string, filter *ConsentSearchFilter) (
		[]Consent, *serviceerror.I18nServiceError)

	// validateConsent validates a consent by ID
	validateConsent(ctx context.Context, ouID, consentID string) (
		*ConsentValidationResult, *serviceerror.I18nServiceError)

	// revokeConsent revokes an active consent record
	revokeConsent(ctx context.Context, ouID, consentID string,
		payload *ConsentRevokeRequest) *serviceerror.I18nServiceError
}
