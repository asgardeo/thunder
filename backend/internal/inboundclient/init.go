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

package inboundclient

import (
	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/consent"
	layoutmgt "github.com/asgardeo/thunder/internal/design/layout/mgt"
	thememgt "github.com/asgardeo/thunder/internal/design/theme/mgt"
	"github.com/asgardeo/thunder/internal/entityprovider"
	"github.com/asgardeo/thunder/internal/entitytype"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	inboundmodel "github.com/asgardeo/thunder/internal/inboundclient/model"
	"github.com/asgardeo/thunder/internal/system/cache"
	dre "github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
	"github.com/asgardeo/thunder/internal/system/transaction"
)

// Initialize initializes the inbound client service.
func Initialize(
	cacheManager cache.CacheManagerInterface,
	certService cert.CertificateServiceInterface,
	entityProvider entityprovider.EntityProviderInterface,
	themeMgt thememgt.ThemeMgtServiceInterface,
	layoutMgt layoutmgt.LayoutMgtServiceInterface,
	flowMgt flowmgt.FlowMgtServiceInterface,
	entityType entitytype.EntityTypeServiceInterface,
	consentService consent.ConsentServiceInterface,
) (InboundClientServiceInterface, error) {
	store, transactioner, err := initializeStore(cacheManager)
	if err != nil {
		return nil, err
	}
	return newInboundClientService(store, transactioner, certService, entityProvider,
		themeMgt, layoutMgt, flowMgt, entityType, consentService), nil
}

// initializeStore always creates a composite store (DB + in-memory file store).
func initializeStore(cacheManager cache.CacheManagerInterface) (
	inboundClientStoreInterface, transaction.Transactioner, error) {
	fileStore := newFileBasedStore(dre.KeyTypeInboundAuth)
	dbStore, transactioner, err := newStore()
	if err != nil {
		return nil, nil, err
	}
	inboundClientCache := cache.GetCache[*inboundmodel.InboundClient](cacheManager, inboundClientCacheName)
	oauthProfileCache := cache.GetCache[*inboundmodel.OAuthProfile](cacheManager, oauthProfileCacheName)
	cached := newCachedBackStore(dbStore, inboundClientCache, oauthProfileCache)
	return newCompositeStore(fileStore, cached), transactioner, nil
}
