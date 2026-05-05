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

package oauth

// This file declares the minimum method-sets (ports) that the oauth package requires from its
// infrastructure dependencies. Each port is a strict subset of the corresponding internal interface.
// When the underlying data-model types are promoted to pkg/, these ports become the decoupling
// boundary that allows external consumers to provide their own implementations without importing
// internal packages.
//
// Current state: the return types still reference internal/ types (same module — permitted).
// Roadmap: move the shared model types (OrganizationUnit, ResourceServer, etc.) to pkg/ so that
// external consumers can implement these ports without depending on internal/.

import (
	"context"
	"crypto"
	"crypto/x509"

	"github.com/asgardeo/thunder/internal/attributecache"
	"github.com/asgardeo/thunder/internal/authz"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	i18nmgt "github.com/asgardeo/thunder/internal/system/i18n/mgt"
	"github.com/asgardeo/thunder/internal/system/observability/event"
)

// JWTServicePort is the minimum JWT capability required by the oauth package.
type JWTServicePort interface {
	GenerateJWT(sub, iss string, validityPeriod int64, claims map[string]interface{}, typ, alg string) (
		string, *serviceerror.ServiceError)
	VerifyJWT(jwtToken string, expectedAud, expectedIss string) *serviceerror.ServiceError
	VerifyJWTWithPublicKey(jwtToken string, jwtPublicKey crypto.PublicKey, expectedAud,
		expectedIss string) *serviceerror.ServiceError
	VerifyJWTWithJWKS(jwtToken, jwksURL, expectedAud, expectedIss string) *serviceerror.ServiceError
	VerifyJWTSignature(jwtToken string) *serviceerror.ServiceError
}

// JWEServicePort is the minimum JWE capability required by the oauth package.
type JWEServicePort interface {
	Encrypt(payload []byte, recipientPublicKey crypto.PublicKey,
		keyAlg, contentAlg string) (string, error)
}

// PKIServicePort is the minimum PKI capability required by the oauth package.
type PKIServicePort interface {
	GetCertThumbprint(id string) string
	GetAllX509Certificates() (map[string]*x509.Certificate, *serviceerror.ServiceError)
	GetSupportedSigningAlgorithms() []string
}

// ObservabilityServicePort is the minimum observability capability required by the oauth package.
type ObservabilityServicePort interface {
	IsEnabled() bool
	PublishEvent(evt *event.Event)
}

// OUServicePort is the minimum OU capability required by the oauth package.
type OUServicePort interface {
	GetOrganizationUnit(ctx context.Context, id string) (ou.OrganizationUnit, *serviceerror.ServiceError)
	GetOrganizationUnitList(ctx context.Context, limit, offset int) (
		*ou.OrganizationUnitListResponse, *serviceerror.ServiceError)
}

// AttributeCacheServicePort is the minimum attribute-cache capability required by the oauth package.
type AttributeCacheServicePort interface {
	GetAttributeCache(ctx context.Context, id string) (*attributecache.AttributeCache, *serviceerror.ServiceError)
	ExtendAttributeCacheTTL(ctx context.Context, id string, ttlSeconds int) *serviceerror.ServiceError
}

// AuthZServicePort is the minimum authorization capability required by the oauth package.
type AuthZServicePort interface {
	GetAuthorizedPermissions(
		ctx context.Context, request authz.GetAuthorizedPermissionsRequest,
	) (*authz.GetAuthorizedPermissionsResponse, *serviceerror.ServiceError)
}

// ResourceServicePort is the minimum resource-server capability required by the oauth package.
type ResourceServicePort interface {
	GetResourceServerByIdentifier(
		ctx context.Context, identifier string,
	) (*resource.ResourceServer, *serviceerror.ServiceError)
	ValidatePermissions(
		ctx context.Context, resourceServerID string, permissions []string,
	) ([]string, *serviceerror.ServiceError)
	FindResourceServersByPermissions(
		ctx context.Context, permissions []string,
	) ([]resource.ResourceServer, *serviceerror.ServiceError)
}

// I18nServicePort is the minimum i18n capability required by the oauth package.
type I18nServicePort interface {
	SetTranslationOverridesForNamespace(ctx context.Context, namespace string,
		entries map[string]map[string]string) *serviceerror.ServiceError
	DeleteTranslationsByKey(ctx context.Context, namespace string, key string) *serviceerror.ServiceError
}

// compile-time checks: verify that the internal implementations satisfy the port interfaces.
var (
	_ OUServicePort             = (ou.OrganizationUnitServiceInterface)(nil)
	_ AttributeCacheServicePort = (attributecache.AttributeCacheServiceInterface)(nil)
	_ AuthZServicePort          = (authz.AuthorizationServiceInterface)(nil)
	_ ResourceServicePort       = (resource.ResourceServiceInterface)(nil)
	_ I18nServicePort           = (i18nmgt.I18nServiceInterface)(nil)
)
