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

import type {Context} from 'react';
import {createContext} from 'react';
import type {UserTypeCreateFlowStep} from '../../models/user-type-create-flow';
import type {SchemaPropertyInput} from '../../types/user-types';

/**
 * User type creation context state interface.
 *
 * Provides centralized state management for the user type creation wizard flow.
 *
 * @public
 */
export interface UserTypeCreateContextType {
  currentStep: UserTypeCreateFlowStep;
  setCurrentStep: (step: UserTypeCreateFlowStep) => void;

  name: string;
  setName: (name: string) => void;

  ouId: string;
  setOuId: (ouId: string) => void;

  allowSelfRegistration: boolean;
  setAllowSelfRegistration: (allow: boolean) => void;

  properties: SchemaPropertyInput[];
  setProperties: (properties: SchemaPropertyInput[]) => void;

  enumInput: Record<string, string>;
  setEnumInput: (enumInput: Record<string, string>) => void;

  error: string | null;
  setError: (error: string | null) => void;

  reset: () => void;
}

const UserTypeCreateContext: Context<UserTypeCreateContextType | undefined> = createContext<
  UserTypeCreateContextType | undefined
>(undefined);

export default UserTypeCreateContext;
