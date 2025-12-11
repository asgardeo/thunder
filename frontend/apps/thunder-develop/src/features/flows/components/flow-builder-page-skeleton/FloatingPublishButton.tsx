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

import {Button} from '@wso2/oxygen-ui';
import './FloatingPublishButton.scss';
import classNames from 'classnames';
import {useTranslation} from 'react-i18next';
import {useEffect, type HTMLAttributes} from 'react';
import useGetFlowConfig from '../../api/useGetFlowConfig';
import useValidationStatus from '../../hooks/useValidationStatus';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import type {FlowTypes} from '../../models/metadata';

/**
 * Props interface of {@link FloatingPublishButton}
 */
export interface FloatingPublishButtonProps extends HTMLAttributes<HTMLDivElement> {
  flowType: FlowTypes;
  flowTypeDisplayName: string;
  isPublishing: boolean;
  onPublish: () => Promise<boolean>;
}

/**
 * Publish button for the for the flow builder page.
 *
 * @param props - Props injected to the component.
 * @returns FloatingPublishButton component.
 */
function FloatingPublishButton({
  className,
  flowType,
  flowTypeDisplayName,
  isPublishing,
  onPublish,
}: FloatingPublishButtonProps) {
  const {t} = useTranslation();
  const {isResourcePropertiesPanelOpen} = useFlowBuilderCore();
  const {openValidationPanel, isValid} = useValidationStatus();
  const {data: flowConfig, error: flowConfigError} = useGetFlowConfig(flowType);

  useEffect(() => {
    if (flowConfigError) {
      // TODO: Handle flow config fetch error with proper error notification
    }
  }, [flowConfigError, flowTypeDisplayName]);

  return (
    <Button
      className={classNames(
        'floating-publish-button',
        {
          transition: isResourcePropertiesPanelOpen ?? openValidationPanel,
        },
        className,
      )}
      variant="contained"
      loading={isPublishing}
      onClick={() => {
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        onPublish();
      }}
      disabled={!isValid}
    >
      {flowConfig?.isEnabled ? t('common:publish') : t('common:saveDraft')}
    </Button>
  );
}

export default FloatingPublishButton;
