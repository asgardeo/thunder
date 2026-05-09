/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import {useContext} from 'react';
import FlowConfigContext, {type FlowConfigContextProps} from '../context/FlowConfigContext';

/**
 * Hook that provides access to flow configuration (metadata, verbose mode, edge style, node/edge types, factories).
 *
 * Use this hook when a component needs flow-level settings that rarely change after initialization.
 * For other context domains, see useUIPanelState, useInteractionState, useI18nConfig.
 */
const useFlowConfig = (): FlowConfigContextProps => {
  const context = useContext(FlowConfigContext);

  if (context === undefined) {
    throw new Error('useFlowConfig must be used within a FlowBuilderCoreProvider');
  }

  return context;
};

export default useFlowConfig;
