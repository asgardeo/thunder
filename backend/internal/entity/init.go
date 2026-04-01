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

package entity

import (
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/transaction"
)

// Initialize initializes the entity service.
func Initialize() (EntityServiceInterface, error) {
	storeMode := getEntityStoreMode()

	store, transactioner, err := initializeStore(storeMode)
	if err != nil {
		return nil, err
	}

	svc := newEntityService(store, transactioner)
	return svc, nil
}

// initializeStore creates the appropriate store based on the configured mode.
func initializeStore(storeMode serverconst.StoreMode) (
	entityStoreInterface, transaction.Transactioner, error,
) {
	switch storeMode {
	case serverconst.StoreModeComposite:
		fileStore, _ := newEntityFileBasedStore()
		dbStore, transactioner, err := newEntityDBStore()
		if err != nil {
			return nil, nil, err
		}
		return newEntityCompositeStore(fileStore, dbStore), transactioner, nil

	case serverconst.StoreModeDeclarative:
		fileStore, transactioner := newEntityFileBasedStore()
		return fileStore, transactioner, nil

	default:
		dbStore, transactioner, err := newEntityDBStore()
		if err != nil {
			return nil, nil, err
		}
		return dbStore, transactioner, nil
	}
}
