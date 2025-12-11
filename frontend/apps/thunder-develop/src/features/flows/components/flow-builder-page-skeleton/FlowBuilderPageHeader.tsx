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

import {Box, Breadcrumbs, IconButton, Switch, Tooltip, Typography} from '@wso2/oxygen-ui';
import {ArrowLeftIcon, RotateCcwSquare, RotateCw} from '@wso2/oxygen-ui-icons-react';
import {useEffect, useRef, useState, type ChangeEvent, type MutableRefObject} from 'react';
import {useTranslation} from 'react-i18next';
import useGetFlowConfig from '../../api/useGetFlowConfig';
import useValidationStatus from '../../hooks/useValidationStatus';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import type {FlowTypes} from '../../models/metadata';
import updateFlowConfig from '../../api/updateFlowConfig';
import ValidationStatusLabels from '../validation-panel/ValidationStatusLabels';

/**
 * Props interface of {@link FlowBuilderPageHeader}
 */
export interface FlowBuilderPageHeaderProps {
  flowType: FlowTypes;
  flowTypeDisplayName: string;
  onPublish: () => Promise<boolean>;
  isPublishing?: boolean;
}

/**
 * Header for the flow builder page.
 *
 * @param props - Props injected to the component.
 * @returns FlowBuilderPageHeader component.
 */
function FlowBuilderPageHeader({
  onPublish,
  isPublishing = false,
  flowType,
  flowTypeDisplayName,
}: FlowBuilderPageHeaderProps) {
  const {isValid} = useValidationStatus();
  const {data: flowConfig, mutate: mutateFlowConfig, error: flowConfigError} = useGetFlowConfig(flowType);
  const {isAutoSavingLocalHistory, setIsVersionHistoryPanelOpen} = useFlowBuilderCore();
  const {t} = useTranslation();

  const [isFlowConfigUpdating, setIsFlowConfigUpdating] = useState<boolean>(false);
  const [showSavingSuspense, setShowSavingSuspense] = useState<boolean>(false);

  const suspenseTimeoutRef: MutableRefObject<NodeJS.Timeout | null> = useRef<NodeJS.Timeout | null>(null);

  /**
   * Handle flow config fetch errors using useEffect
   */
  useEffect(() => {
    if (flowConfigError) {
      // TODO: Handle flow config fetch error with proper error notification
    }
  }, [flowConfigError, isPublishing]);

  /**
   * Handle auto-saving suspense logic - keep saving indicator for 2 seconds
   */
  useEffect(() => {
    if (isAutoSavingLocalHistory) {
      setShowSavingSuspense(true);

      if (suspenseTimeoutRef.current) {
        clearTimeout(suspenseTimeoutRef.current);
      }

      suspenseTimeoutRef.current = setTimeout(() => {
        setShowSavingSuspense(false);
        suspenseTimeoutRef.current = null;
      }, 3000);
    }
  }, [isAutoSavingLocalHistory]);

  /**
   * Cleanup timeout on unmount
   */
  useEffect(
    () => () => {
      if (suspenseTimeoutRef.current) {
        clearTimeout(suspenseTimeoutRef.current);
      }
    },
    [],
  );

  /**
   * Handles the back button click event.
   */
  const handleBackButtonClick = (): void => {
    // Navigate back to the flows list page
  };

  /**
   * Dispatches error alerts for flow config API errors.
   */
  const handleFlowConfigError = (): void => {
    // TODO: Dispatch error alert for flow config operation failure
  };

  /**
   * Dispatches success alerts for flow config operations.
   */
  const handleFlowConfigSuccess = (): void => {
    // TODO: Dispatch success alert for flow config operation
  };

  /**
   * Handles the toggle switch change event.
   */
  const handleToggleFlow = (event: ChangeEvent<HTMLInputElement>): void => {
    const isEnabled: boolean = event.target.checked;

    setIsFlowConfigUpdating(true);

    const performToggle = async (): Promise<void> => {
      try {
        let isPublishSuccess = true;

        if (isEnabled) {
          isPublishSuccess = await onPublish();
        }

        if (isPublishSuccess) {
          await updateFlowConfig({
            flowType,
            isEnabled,
          });
          handleFlowConfigSuccess();
          mutateFlowConfig();
        }
      } catch {
        handleFlowConfigError();
      } finally {
        setIsFlowConfigUpdating(false);
      }
    };

    // eslint-disable-next-line @typescript-eslint/no-floating-promises
    performToggle();
  };

  return (
    <Box display="flex" className="page-header" justifyContent="space-between" alignItems="center">
      <Box display="flex" gap={3} alignItems="center">
        <IconButton onClick={handleBackButtonClick}>
          <ArrowLeftIcon />
        </IconButton>
        <Breadcrumbs aria-label="breadcrumb" className="flow-builder-page-header-breadcrumbs">
          <Typography color="inherit" sx={{cursor: 'pointer'}} onClick={() => null}>
            {t('flows:label')}
          </Typography>
          <Typography>{t('flows:core.breadcrumb', {flowType: flowTypeDisplayName})}</Typography>
        </Breadcrumbs>
      </Box>
      <Box display="flex" justifyContent="center" alignItems="center" gap={1.5}>
        {(showSavingSuspense || isFlowConfigUpdating) && (
          <Box sx={{alignItems: 'center', display: 'flex', gap: 1}}>
            <RotateCw />
            <Typography variant="body2" color="text.secondary">
              {t('flows:core.autoSave.savingInProgress')}
            </Typography>
          </Box>
        )}
        <Box>
          <Tooltip title="Version History (Local)">
            <IconButton onClick={() => setIsVersionHistoryPanelOpen(true)}>
              <RotateCcwSquare height={20} width={20} />
            </IconButton>
          </Tooltip>
        </Box>
        {/* TODO:  */}
        <ValidationStatusLabels />
        <Box display="flex" alignItems="center">
          <Typography>
            {flowConfig?.isEnabled ? t('flows:core.labels.disableFlow') : t('flows:core.labels.enableFlow')}
          </Typography>
          <Tooltip
            title={
              flowConfig?.isEnabled
                ? t('flows:core.tooltips.disableFlow', {flowType: flowTypeDisplayName})
                : t('flows:core.tooltips.enableFlow', {flowType: flowTypeDisplayName})
            }
          >
            <Switch
              checked={flowConfig?.isEnabled ?? false}
              onChange={handleToggleFlow}
              disabled={isFlowConfigUpdating || (!flowConfig?.isEnabled && !isValid)}
            />
          </Tooltip>
        </Box>
      </Box>
    </Box>
  );
}

export default FlowBuilderPageHeader;
