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

import type {PropsWithChildren} from 'react';
import {useState, useMemo, useCallback} from 'react';
import UserCreateContext, {type UserCreateContextType} from './UserCreateContext';
import {UserCreateFlowStep} from '../../models/user-create-flow';
import type {SchemaInterface} from '../../types/users';

/**
 * Initial state values for user creation.
 *
 * @internal
 */
const INITIAL_STATE = {
  currentStep: UserCreateFlowStep.USER_TYPE as UserCreateFlowStep,
  selectedSchema: null as SchemaInterface | null,
  formValues: {} as Record<string, unknown>,
  error: null as string | null,
};

/**
 * React context provider component that provides user creation state
 * to all child components in the wizard flow.
 *
 * @public
 */
export default function UserCreateProvider({children}: PropsWithChildren) {
  const [currentStep, setCurrentStep] = useState<UserCreateFlowStep>(INITIAL_STATE.currentStep);
  const [selectedSchema, setSelectedSchema] = useState<SchemaInterface | null>(INITIAL_STATE.selectedSchema);
  const [formValues, setFormValues] = useState<Record<string, unknown>>(INITIAL_STATE.formValues);
  const [error, setError] = useState<string | null>(INITIAL_STATE.error);

  const reset = useCallback((): void => {
    setCurrentStep(INITIAL_STATE.currentStep);
    setSelectedSchema(INITIAL_STATE.selectedSchema);
    setFormValues(INITIAL_STATE.formValues);
    setError(INITIAL_STATE.error);
  }, []);

  const contextValue: UserCreateContextType = useMemo(
    () => ({
      currentStep,
      setCurrentStep,
      selectedSchema,
      setSelectedSchema,
      formValues,
      setFormValues,
      error,
      setError,
      reset,
    }),
    [currentStep, selectedSchema, formValues, error, reset],
  );

  return <UserCreateContext.Provider value={contextValue}>{children}</UserCreateContext.Provider>;
}
