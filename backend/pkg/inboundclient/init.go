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
	layoutmgt "github.com/asgardeo/thunder/internal/design/layout/mgt"
	thememgt "github.com/asgardeo/thunder/internal/design/theme/mgt"
	"github.com/asgardeo/thunder/internal/entitytype"
	dre "github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
	"github.com/asgardeo/thunder/internal/system/transaction"
	"github.com/asgardeo/thunder/pkg/consent"
	"github.com/asgardeo/thunder/pkg/entityprovider"
	flowmgt "github.com/asgardeo/thunder/pkg/flow/mgt"
)

// Initialize initializes the inbound client service.
func Initialize(certService cert.CertificateServiceInterface,
	entityProvider entityprovider.EntityProviderInterface,
	themeMgt thememgt.ThemeMgtServiceInterface,
	layoutMgt layoutmgt.LayoutMgtServiceInterface,
	flowMgt flowmgt.FlowMgtServiceInterface,
	entityType entitytype.EntityTypeServiceInterface,
	consentService consent.ConsentServiceInterface,
) (InboundClientServiceInterface, error) {
	store, transactioner, err := initializeStore()
	if err != nil {
		return nil, err
	}
	return newInboundClientService(store, transactioner, certService, entityProvider,
		themeMgt, layoutMgt, flowMgt, entityType, consentService), nil
}

// initializeStore always creates a composite store (DB + in-memory file store).
func initializeStore() (inboundClientStoreInterface, transaction.Transactioner, error) {
	fileStore := newFileBasedStore(dre.KeyTypeInboundAuth)
	dbStore, transactioner, err := newStore()
	if err != nil {
		return nil, nil, err
	}
	cached := newCachedBackStore(dbStore)
	return newCompositeStore(fileStore, cached), transactioner, nil
}
