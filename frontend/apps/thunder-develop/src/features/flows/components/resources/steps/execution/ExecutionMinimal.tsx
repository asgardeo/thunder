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

import type {Step, StepData} from '@/features/flows/models/steps';
import type {ReactElement} from 'react';
import useFlowBuilderCore from '@/features/flows/hooks/useFlowBuilderCore';
import {Handle, Position, useNodeId} from '@xyflow/react';
import {useTranslation} from 'react-i18next';
import {Box, Card, IconButton, Tooltip, Typography} from '@wso2/oxygen-ui';
import {CogIcon} from '@wso2/oxygen-ui-icons-react';
import classNames from 'classnames';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import ExecutionFactory from './execution-factory/ExecutionFactory';
import './ExecutionMinimal.scss';

/**
 * Props interface of {@link ExecutionMinimal}
 */
export interface ExecutionMinimalPropsInterface {
  /**
   * Resource object of the execution step.
   */
  resource: Step;
}

/**
 * Execution (Minimal) Node component.
 *
 * @param props - Props injected to the component.
 * @returns Execution (Minimal) node component.
 */
function ExecutionMinimal({resource}: ExecutionMinimalPropsInterface): ReactElement {
  const {setLastInteractedResource, setLastInteractedStepId, setIsOpenResourcePropertiesPanel} = useFlowBuilderCore();
  const stepId: string | null = useNodeId();

  const {t} = useTranslation();

  // Get the display label from resource.display.label, falling back to executor name
  const displayLabel = resource.display?.label ?? resource.data?.action?.executor?.name ?? 'Executor';

  // Check if the node has action data with onSuccess/onFailure fields defined (even if empty)
  // This indicates the node supports branching and should show both handles
  const stepData = resource.data as StepData | undefined;
  const hasBranchingSupport = stepData?.action && 'onFailure' in stepData.action;

  const handleConfigClick = (): void => {
    if (stepId !== null) {
      setLastInteractedStepId(stepId);
    }
    setLastInteractedResource({
      ...resource,
      config: {
        ...(resource?.config || {}),
        ...(typeof resource.data?.config === 'object' && resource.data?.config !== null ? resource.data.config : {}),
      },
    });
    setIsOpenResourcePropertiesPanel(true);
  };

  return (
    <Box className={classNames('execution-minimal-step', {'has-branching': hasBranchingSupport})}>
      <Box
        display="flex"
        justifyContent="space-between"
        alignItems="center"
        className="execution-minimal-step-action-panel"
        sx={{
          backgroundColor: 'secondary.main',
          px: 2,
          py: 1.25,
          height: 44,
        }}
      >
        <Typography
          variant="body2"
          className="execution-minimal-step-title"
          sx={{
            color: 'text.primary',
            fontWeight: 500,
          }}
        >
          {displayLabel}
        </Typography>
        <Tooltip title={t('flows:core.executions.tooltip.configurationHint')}>
          <IconButton
            size="small"
            onClick={handleConfigClick}
            className="execution-minimal-step-action"
            sx={(theme) => ({
              color: 'text.secondary',
              '&:hover': {
                ...theme.applyStyles('dark', {
                  backgroundColor: 'rgba(0, 0, 0, 0.2)',
                  color: 'common.white',
                }),
                ...theme.applyStyles('light', {
                  backgroundColor: 'rgba(0, 0, 0, 0.1)',
                  color: 'text.primary',
                }),
              },
            })}
          >
            <CogIcon size={18} />
          </IconButton>
        </Tooltip>
      </Box>
      <Handle type="target" position={Position.Left} />
      <Card
        className="execution-minimal-step-content"
        onClick={() => {
          setLastInteractedStepId(resource.id);
          setLastInteractedResource(resource);
        }}
      >
        <ExecutionFactory resource={resource} />
      </Card>
      {/* Success handle - always shown on the right */}
      {hasBranchingSupport ? (
        <Tooltip title={t('flows:core.executions.handles.success')} placement="right">
          <Box className="handle-wrapper success-wrapper">
            <Handle
              type="source"
              position={Position.Right}
              id={`${resource.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`}
              className="execution-handle-success"
            />
          </Box>
        </Tooltip>
      ) : (
        <Handle
          type="source"
          position={Position.Right}
          id={`${resource.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`}
        />
      )}
      {/* Failure handle - shown at the bottom when the action supports branching (has onFailure property) */}
      {hasBranchingSupport && (
        <Tooltip title={t('flows:core.executions.handles.failure')} placement="bottom">
          <Box className="handle-wrapper failure-wrapper">
            <Handle type="source" position={Position.Bottom} id="failure" className="execution-handle-failure" />
          </Box>
        </Tooltip>
      )}
    </Box>
  );
}

export default ExecutionMinimal;
