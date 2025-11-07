/**
 * Copyright (c) 2023-2025, WSO2 LLC. (https://www.wso2.com).
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

import {useCallback, useEffect, useMemo, type ReactElement} from 'react';
import cloneDeep from 'lodash-es/cloneDeep';
import {useTranslation} from 'react-i18next';
import useFlowBuilderCore from '@/features/flows/hooks/useFlowBuilderCore';
import useValidationStatus from '@/features/flows/hooks/useValidationStatus';
import {ExecutionStepViewTypes, ExecutionTypes, type Step, type StepAction} from '@/features/flows/models/steps';
import type {Element} from '@/features/flows/models/elements';
import Notification, {NotificationType} from '@/features/flows/models/notification';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import ValidationErrorBoundary from '../../../validation-panel/ValidationErrorBoundary';
import ExecutionMinimal from './ExecutionMinimal';
import View from '../view/View';
import type {CommonStepFactoryPropsInterface} from '../CommonStepFactory';

/**
 * Props interface of {@link Execution}
 */
export type ExecutionPropsInterface = CommonStepFactoryPropsInterface;

/**
 * Execution Node component.
 *
 * @param props - Props injected to the component.
 * @returns Execution node component.
 */
function Execution({id, data, resources}: ExecutionPropsInterface): ReactElement | null {
  const {setLastInteractedResource, setLastInteractedStepId} = useFlowBuilderCore();
  const {addNotification, removeNotification, setOpenValidationPanel, setSelectedNotification} = useValidationStatus();
  const {t} = useTranslation();

  const components: Element[] = (data?.components as Element[]) || [];

  /**
   * Find the default execution resource as a fallback.
   */
  const findDefaultResource = useCallback((): Step | undefined => {
    return resources?.find((r: Step) => r.display.label === ExecutionStepViewTypes.Default);
  }, [resources]);

  /*
   * Resolve resource for the execution step.
   */
  const resolveResource = useCallback(
    (executionType: ExecutionTypes): Step | undefined => {
      switch (executionType) {
        case ExecutionTypes.PasskeyEnrollment: {
          let resource: Step | undefined = resources?.find(
            (r: Step) => r.display.label === ExecutionStepViewTypes.PasskeyView,
          );

          // Fallback to default if specific view not found
          if (!resource) {
            resource = findDefaultResource();
          }

          if (resource) {
            const resourceCopy: Step = cloneDeep(resource);

            (resourceCopy.display as {displayname?: string}).displayname = t('flows:core.executions.names.passkeyEnrollment');

            return resourceCopy;
          }

          return resource;
        }
        case ExecutionTypes.MagicLinkExecutor: {
          let resource: Step | undefined = resources?.find(
            (r: Step) => r.display.label === ExecutionStepViewTypes.MagicLinkView,
          );

          // Fallback to default if specific view not found
          if (!resource) {
            resource = findDefaultResource();
          }

          if (resource) {
            const resourceCopy: Step = cloneDeep(resource);

            (resourceCopy.display as {displayname?: string}).displayname = t('flows:core.executions.names.magicLink');

            return resourceCopy;
          }

          return resource;
        }
        case ExecutionTypes.ConfirmationCode: {
          const resource: Step | undefined = resources?.find(
            (r: Step) => r.display.label === ExecutionStepViewTypes.Default,
          );

          if (resource) {
            const resourceCopy: Step = cloneDeep(resource);

            (resourceCopy.display as {displayname?: string}).displayname = t('flows:core.executions.names.confirmationCode');

            return resourceCopy;
          }

          return resource;
        }
        default: {
          return findDefaultResource();
        }
      }
    },
    [resources, t, findDefaultResource],
  );

  /**
   * Full resource object with data included.
   */
  const action = data?.action as StepAction | undefined;
  const fullResource: Step | undefined = useMemo(() => {
    const executorName = action?.executor?.name as ExecutionTypes;
    const resource: Step | undefined = resolveResource(executorName);

    if (!resource || !data) {
      return undefined;
    }

    return cloneDeep({
      ...resource,
      id,
      data,
    }) as Step;
  }, [data, id, resolveResource, action]);

  useEffect(() => {
    // Executors that need to show the landing info notification.
    const executorsWithLandingInfo: ExecutionTypes[] = [ExecutionTypes.ConfirmationCode];
    const executorName = (data?.action as StepAction | undefined)?.executor?.name as ExecutionTypes;

    if (fullResource && executorsWithLandingInfo.includes(executorName)) {
      const infoNotification: Notification = new Notification(
        `${id}_info_landing`,
        t('flows:core.executions.landing.message', {
          executor: (fullResource.display as {displayname?: string})?.displayname ?? fullResource.display.label,
        }),
        NotificationType.INFO,
      );

      addNotification?.(infoNotification);

      return () => {
        // Remove the notification on unmount.
        removeNotification?.(infoNotification.getId());
      };
    }

    return undefined;
  }, [fullResource, data?.action, id, t, addNotification, removeNotification]);

  /**
   * Resolves the execution name based on the type.
   *
   * @param executionType - The type of the execution.
   * @returns Resolved execution name.
   */
  const resolveExecutionName = (executionType: ExecutionTypes): string => {
    switch (executionType) {
      case ExecutionTypes.GoogleFederation:
        return t('flows:core.executions.names.google');
      case ExecutionTypes.AppleFederation:
        return t('flows:core.executions.names.apple');
      case ExecutionTypes.GithubFederation:
        return t('flows:core.executions.names.github');
      case ExecutionTypes.FacebookFederation:
        return t('flows:core.executions.names.facebook');
      case ExecutionTypes.MicrosoftFederation:
        return t('flows:core.executions.names.microsoft');
      case ExecutionTypes.PasskeyEnrollment:
        return t('flows:core.executions.names.passkeyEnrollment');
      case ExecutionTypes.ConfirmationCode:
        return t('flows:core.executions.names.confirmationCode');
      case ExecutionTypes.MagicLinkExecutor:
        return t('flows:core.executions.names.magicLink');
      case ExecutionTypes.SendEmailOTP:
        return t('flows:core.executions.names.sendEmailOTP');
      case ExecutionTypes.VerifyEmailOTP:
        return t('flows:core.executions.names.verifyEmailOTP');
      case ExecutionTypes.SendSMS:
        return t('flows:core.executions.names.sendSMS');
      case ExecutionTypes.VerifySMSOTP:
        return t('flows:core.executions.names.verifySMSOTP');
      default:
        return t('flows:core.executions.names.default');
    }
  };

  if (!fullResource) {
    return null;
  }

  const executorName = ((data?.action as StepAction | undefined)?.executor?.name ?? '') as ExecutionTypes;

  return (
    <ValidationErrorBoundary disableErrorBoundaryOnHover={false} resource={fullResource}>
      {components && components.length > 0 ? (
        <View
          heading={resolveExecutionName(executorName)}
          data={data}
          enableSourceHandle
          droppableAllowedTypes={VisualFlowConstants.FLOW_BUILDER_STATIC_CONTENT_ALLOWED_RESOURCE_TYPES}
          onActionPanelDoubleClick={() => {
            setOpenValidationPanel?.(false);
            setSelectedNotification?.(null);
            setLastInteractedStepId(id);
            setLastInteractedResource(fullResource);
          }}
          resources={resources}
        />
      ) : (
        <ExecutionMinimal resource={fullResource} />
      )}
    </ValidationErrorBoundary>
  );
}

export default Execution;
