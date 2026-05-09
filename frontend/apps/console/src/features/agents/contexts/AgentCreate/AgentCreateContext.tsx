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
import type {AgentCreateFlowStep} from '../../models/agent-create-flow';

export interface AgentCreateContextType {
  currentStep: AgentCreateFlowStep;
  setCurrentStep: (step: AgentCreateFlowStep) => void;

  selectedSchema: {id: string; name: string; ouId: string} | null;
  setSelectedSchema: (schema: {id: string; name: string; ouId: string} | null) => void;

  selectedOuId: string | null;
  setSelectedOuId: (ouId: string | null) => void;

  agentName: string;
  setAgentName: (name: string) => void;

  formValues: Record<string, unknown>;
  setFormValues: (values: Record<string, unknown>) => void;

  selectedOwnerId: string | null;
  setSelectedOwnerId: (id: string | null) => void;

  error: string | null;
  setError: (error: string | null) => void;

  reset: () => void;
}

const AgentCreateContext: Context<AgentCreateContextType | undefined> = createContext<
  AgentCreateContextType | undefined
>(undefined);

export default AgentCreateContext;
