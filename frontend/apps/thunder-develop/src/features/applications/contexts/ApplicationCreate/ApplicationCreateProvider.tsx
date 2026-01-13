/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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
import {useState, useMemo, useCallback, useEffect} from 'react';
import useGetApplications from '../../api/useGetApplications';
import {AuthenticatorTypes} from '../../../integrations/models/authenticators';
import ApplicationCreateContext, {type ApplicationCreateContextType} from './ApplicationCreateContext';
import {ApplicationCreateFlowSignInApproach, ApplicationCreateFlowStep} from '../../models/application-create-flow';
import type {
  TechnologyApplicationTemplate,
  PlatformApplicationTemplate,
  ApplicationTemplate,
} from '../../models/application-templates';
import type {BasicFlowDefinition} from '../../../flows/models/responses';
import generateAppPrimaryColorSuggestions from '../../utils/generateAppPrimaryColorSuggestions';

/**
 * Props for the {@link ApplicationCreateProvider} component.
 *
 * @public
 */
export type ApplicationCreateProviderProps = PropsWithChildren;

/**
 * Initial state values for application creation
 *
 * @internal
 */
const INITIAL_STATE: {
  currentStep: ApplicationCreateFlowStep;
  appName: string;
  selectedColor: string;
  appLogo: string | null;
  integrations: Record<string, boolean>;
  selectedAuthFlow: BasicFlowDefinition | null;
  signInApproach: ApplicationCreateFlowSignInApproach;
  selectedTechnology: TechnologyApplicationTemplate | null;
  selectedPlatform: PlatformApplicationTemplate | null;
  hostingUrl: string;
  callbackUrlFromConfig: string;
  hasCompletedOnboarding: boolean;
  error: string | null;
} = {
  currentStep: ApplicationCreateFlowStep.NAME,
  appName: '',
  selectedColor: generateAppPrimaryColorSuggestions()[0],
  appLogo: null,
  integrations: {
    [AuthenticatorTypes.BASIC_AUTH]: true,
  },
  selectedAuthFlow: null,
  signInApproach: ApplicationCreateFlowSignInApproach.INBUILT as ApplicationCreateFlowSignInApproach,
  selectedTechnology: null,
  selectedPlatform: null,
  hostingUrl: '',
  callbackUrlFromConfig: '',
  hasCompletedOnboarding: false,
  error: null,
};

/**
 * React context provider component that provides application creation state
 * to all child components.
 *
 * This component manages all the state needed across the multi-step onboarding process
 * for creating a new application. It provides state variables, setter functions, and
 * utility methods like toggleIntegration and reset.
 *
 * The provider creates utility methods for common operations such as toggling integrations
 * and resetting all state to initial values.
 *
 * @param props - The component props
 * @param props.children - React children to be wrapped with the application create context
 *
 * @returns JSX element that provides application creation context to children
 *
 * @example
 * ```tsx
 * import ApplicationCreateProvider from './ApplicationCreateProvider';
 * import ApplicationCreatePage from './ApplicationCreatePage';
 *
 * function App() {
 *   return (
 *     <ApplicationCreateProvider>
 *       <ApplicationCreatePage />
 *     </ApplicationCreateProvider>
 *   );
 * }
 * ```
 *
 * @public
 */
export default function ApplicationCreateProvider({children}: ApplicationCreateProviderProps) {
  const {data: applicationsData} = useGetApplications({limit: 1, offset: 0});
  const [currentStep, setCurrentStep] = useState<ApplicationCreateFlowStep>(INITIAL_STATE.currentStep);
  const [appName, setAppName] = useState<string>(INITIAL_STATE.appName);
  const [selectedColor, setSelectedColor] = useState<string>(INITIAL_STATE.selectedColor);
  const [appLogo, setAppLogo] = useState<string | null>(INITIAL_STATE.appLogo);
  const [integrations, setIntegrations] = useState<Record<string, boolean>>(INITIAL_STATE.integrations);
  const [selectedAuthFlow, setSelectedAuthFlow] = useState<BasicFlowDefinition | null>(INITIAL_STATE.selectedAuthFlow);
  const [signInApproach, setSignInApproach] = useState<ApplicationCreateFlowSignInApproach>(
    INITIAL_STATE.signInApproach,
  );
  const [selectedTechnology, setSelectedTechnology] = useState<TechnologyApplicationTemplate | null>(
    INITIAL_STATE.selectedTechnology,
  );
  const [selectedPlatform, setSelectedPlatform] = useState<PlatformApplicationTemplate | null>(
    INITIAL_STATE.selectedPlatform,
  );
  const [selectedTemplateConfig, setSelectedTemplateConfig] = useState<ApplicationTemplate | null>(null);
  const [hostingUrl, setHostingUrl] = useState<string>(INITIAL_STATE.hostingUrl);
  const [callbackUrlFromConfig, setCallbackUrlFromConfig] = useState<string>(INITIAL_STATE.callbackUrlFromConfig);
  const [hasCompletedOnboarding, setHasCompletedOnboarding] = useState<boolean>(INITIAL_STATE.hasCompletedOnboarding);
  const [error, setError] = useState<string | null>(INITIAL_STATE.error);

  // Check if user has completed onboarding by checking if they have any applications
  useEffect(() => {
    if (applicationsData?.applications && applicationsData.totalResults > 1) {
      setHasCompletedOnboarding(true);
    } else if (applicationsData?.applications && applicationsData.totalResults === 0) {
      setHasCompletedOnboarding(false);
    }
  }, [applicationsData?.applications, applicationsData?.totalResults]);

  const toggleIntegration = useCallback((integrationId: string): void => {
    setIntegrations((prev) => ({
      ...prev,
      [integrationId]: !prev[integrationId],
    }));
  }, []);

  const reset = useCallback((): void => {
    setCurrentStep(INITIAL_STATE.currentStep);
    setAppName(INITIAL_STATE.appName);
    setSelectedColor(INITIAL_STATE.selectedColor);
    setAppLogo(INITIAL_STATE.appLogo);
    setIntegrations(INITIAL_STATE.integrations);
    setSelectedAuthFlow(INITIAL_STATE.selectedAuthFlow);
    setSignInApproach(INITIAL_STATE.signInApproach);
    setSelectedTechnology(INITIAL_STATE.selectedTechnology);
    setSelectedPlatform(INITIAL_STATE.selectedPlatform);
    setHostingUrl(INITIAL_STATE.hostingUrl);
    setCallbackUrlFromConfig(INITIAL_STATE.callbackUrlFromConfig);
    setHasCompletedOnboarding(INITIAL_STATE.hasCompletedOnboarding);
    setError(INITIAL_STATE.error);
  }, []);

  const contextValue: ApplicationCreateContextType = useMemo(
    () => ({
      currentStep,
      setCurrentStep,
      appName,
      setAppName,
      selectedColor,
      setSelectedColor,
      appLogo,
      setAppLogo,
      integrations,
      setIntegrations,
      toggleIntegration,
      selectedAuthFlow,
      setSelectedAuthFlow,
      signInApproach,
      setSignInApproach,
      selectedTechnology,
      setSelectedTechnology,
      selectedPlatform,
      setSelectedPlatform,
      selectedTemplateConfig,
      setSelectedTemplateConfig,
      hostingUrl,
      setHostingUrl,
      callbackUrlFromConfig,
      setCallbackUrlFromConfig,
      hasCompletedOnboarding,
      setHasCompletedOnboarding,
      error,
      setError,
      reset,
    }),
    [
      currentStep,
      appName,
      selectedColor,
      appLogo,
      integrations,
      toggleIntegration,
      selectedAuthFlow,
      signInApproach,
      selectedTechnology,
      selectedPlatform,
      selectedTemplateConfig,
      hostingUrl,
      callbackUrlFromConfig,
      hasCompletedOnboarding,
      error,
      reset,
    ],
  );

  return <ApplicationCreateContext.Provider value={contextValue}>{children}</ApplicationCreateContext.Provider>;
}
