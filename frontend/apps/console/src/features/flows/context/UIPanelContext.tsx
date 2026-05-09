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

import type {Context, Dispatch, ReactNode, SetStateAction} from 'react';
import {createContext} from 'react';

/**
 * Props interface of {@link UIPanelContext}
 */
export interface UIPanelContextProps {
  /**
   * Indicates whether the element panel is open.
   */
  isResourcePanelOpen: boolean;
  /**
   * Indicates whether the element properties panel is open.
   */
  isResourcePropertiesPanelOpen: boolean;
  /**
   * Indicates whether the version history panel is open.
   */
  isVersionHistoryPanelOpen: boolean;
  /**
   * The heading for the element properties panel.
   */
  resourcePropertiesPanelHeading: ReactNode;
  /**
   * Function to set the state of the element panel.
   */
  setIsResourcePanelOpen: Dispatch<SetStateAction<boolean>>;
  /**
   * Function to set the state of the element properties panel.
   */
  setIsOpenResourcePropertiesPanel: (isOpen: boolean) => void;
  /**
   * Function to set the state of the version history panel.
   */
  setIsVersionHistoryPanelOpen: Dispatch<SetStateAction<boolean>>;
  /**
   * Sets the heading for the element properties panel.
   */
  setResourcePropertiesPanelHeading: Dispatch<SetStateAction<ReactNode>>;
  /**
   * Registers a callback to close the validation panel.
   * This is used for mutual exclusion between the resource properties panel and validation panel.
   */
  registerCloseValidationPanel: (callback: () => void) => void;
}

const UIPanelContext: Context<UIPanelContextProps | undefined> = createContext<UIPanelContextProps | undefined>(
  undefined,
);

UIPanelContext.displayName = 'UIPanelContext';

export default UIPanelContext;
