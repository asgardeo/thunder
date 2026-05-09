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

import {useState, useCallback, type ReactNode} from 'react';
import AgentCreateContext from './AgentCreateContext';
import {AgentCreateFlowStep} from '../../models/agent-create-flow';

export default function AgentCreateProvider({children}: {children: ReactNode}) {
  const [currentStep, setCurrentStep] = useState<AgentCreateFlowStep>(AgentCreateFlowStep.NAME);
  const [selectedSchema, setSelectedSchema] = useState<{id: string; name: string; ouId: string} | null>(null);
  const [selectedOuId, setSelectedOuId] = useState<string | null>(null);
  const [agentName, setAgentName] = useState('');
  const [formValues, setFormValues] = useState<Record<string, unknown>>({});
  const [selectedOwnerId, setSelectedOwnerId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const reset = useCallback((): void => {
    setCurrentStep(AgentCreateFlowStep.NAME);
    setSelectedSchema(null);
    setSelectedOuId(null);
    setAgentName('');
    setFormValues({});
    setSelectedOwnerId(null);
    setError(null);
  }, []);

  return (
    <AgentCreateContext.Provider
      value={{
        currentStep,
        setCurrentStep,
        selectedSchema,
        setSelectedSchema,
        selectedOuId,
        setSelectedOuId,
        agentName,
        setAgentName,
        formValues,
        setFormValues,
        selectedOwnerId,
        setSelectedOwnerId,
        error,
        setError,
        reset,
      }}
    >
      {children}
    </AgentCreateContext.Provider>
  );
}
