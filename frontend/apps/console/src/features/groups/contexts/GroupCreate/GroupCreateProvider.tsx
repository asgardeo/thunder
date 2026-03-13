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
import GroupCreateContext, {type GroupCreateContextType} from './GroupCreateContext';
import {GroupCreateFlowStep} from '../../models/group-create-flow';

/**
 * Initial state values for group creation.
 *
 * @internal
 */
const INITIAL_STATE = {
  currentStep: GroupCreateFlowStep.NAME as GroupCreateFlowStep,
  name: '',
  description: '',
  organizationUnitId: '',
  error: null as string | null,
};

/**
 * React context provider component that provides group creation state
 * to all child components in the wizard flow.
 *
 * @public
 */
export default function GroupCreateProvider({children}: PropsWithChildren) {
  const [currentStep, setCurrentStep] = useState<GroupCreateFlowStep>(INITIAL_STATE.currentStep);
  const [name, setName] = useState<string>(INITIAL_STATE.name);
  const [description, setDescription] = useState<string>(INITIAL_STATE.description);
  const [organizationUnitId, setOrganizationUnitId] = useState<string>(INITIAL_STATE.organizationUnitId);
  const [error, setError] = useState<string | null>(INITIAL_STATE.error);

  const reset = useCallback((): void => {
    setCurrentStep(INITIAL_STATE.currentStep);
    setName(INITIAL_STATE.name);
    setDescription(INITIAL_STATE.description);
    setOrganizationUnitId(INITIAL_STATE.organizationUnitId);
    setError(INITIAL_STATE.error);
  }, []);

  const contextValue: GroupCreateContextType = useMemo(
    () => ({
      currentStep,
      setCurrentStep,
      name,
      setName,
      description,
      setDescription,
      organizationUnitId,
      setOrganizationUnitId,
      error,
      setError,
      reset,
    }),
    [currentStep, name, description, organizationUnitId, error, reset],
  );

  return <GroupCreateContext.Provider value={contextValue}>{children}</GroupCreateContext.Provider>;
}
