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
 * Unless requi by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import {Box, FormGroup, IconButton, Menu, MenuItem, Paper, Tooltip, Typography} from '@wso2/oxygen-ui';
import {CogIcon, PlusIcon, TrashIcon} from '@wso2/oxygen-ui-icons-react';
import {Handle, Position, useNodeId, useNodesData, useReactFlow, type Node} from '@xyflow/react';
import classNames from 'classnames';
import {useEffect, useState, type HTMLAttributes, type MouseEvent, type ReactElement} from 'react';
import generateResourceId from '@/features/flows/utils/generateResourceId';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import PluginRegistry from '@/features/flows/plugins/PluginRegistry';
import FlowEventTypes from '@/features/flows/models/extension';
import {CollisionPriority} from '@dnd-kit/abstract';
import {ElementTypes, type Element} from '@/features/flows/models/elements';
import type {StepData} from '@/features/flows/models/steps';
import ReorderableViewElement from './ReorderableElement';
import Droppable from '../../../dnd/droppable';
import './View.scss';

/**
 * Props interface of {@link View}
 */
export interface ViewPropsInterface extends Omit<HTMLAttributes<HTMLDivElement>, 'resource'> {
  /**
   * Resources for the view (required by parent components but not used here).
   * @internal
   */
  // eslint-disable-next-line react/no-unused-prop-types
  resources?: unknown;
  /**
   * Data for the view component.
   */
  data?: StepData;
  /**
   * Name of the view.
   */
  heading?: string;
  /**
   * Droppable allowed resource list.
   */
  droppableAllowedTypes?: string[];
  /**
   * Droppable restricted resource list that should not be accepted.
   */
  droppableRestrictedTypes?: string[];
  /**
   * Flag to enable source handle.
   */
  enableSourceHandle?: boolean;
  /**
   * Event handler for double click on the action panel.
   *
   * @param event - The mouse event.
   */
  onActionPanelDoubleClick?: (event: MouseEvent<HTMLDivElement>) => void;
  /**
   * Is the view deletable.
   */
  deletable?: boolean;
  /**
   * Does the view has configurations.
   */
  configurable?: boolean;
  /**
   * Callback for configure action.
   */
  onConfigure?: () => void;
  /**
   * Callback for adding an element to the view.
   * @param element - The element to add.
   */
  onAddElement?: (element: Element) => void;
  /**
   * List of available elements that can be added to the view.
   */
  availableElements?: Element[];
  /**
   * Callback for adding an element to a form.
   * @param element - The element to add.
   * @param formId - The ID of the form to add to.
   */
  onAddElementToForm?: (element: Element, formId: string) => void;
}

/**
 * Node for representing an empty view as a step in the flow builder.
 *
 * @param props - Props injected to the component.
 * @returns Step Node component.
 */
function View({
  heading = 'View',
  droppableAllowedTypes = undefined,
  droppableRestrictedTypes = undefined,
  enableSourceHandle = false,
  data = undefined,
  onActionPanelDoubleClick = undefined,
  className,
  deletable = true,
  configurable = false,
  onConfigure = undefined,
  resources = undefined,
  onAddElement = undefined,
  availableElements = [],
  onAddElementToForm = undefined,
}: ViewPropsInterface): ReactElement {
  // Suppress unused variable warning - resources is required by interface but not used
  // @ts-expect-error - intentionally unused
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const unusedResources = resources;
  const stepId: string | null = useNodeId();
  const node: Pick<Node, 'data'> | null = useNodesData(stepId ?? '');
  const {deleteElements, updateNodeData} = useReactFlow();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const menuOpen = Boolean(anchorEl);

  const handleMenuOpen = (event: MouseEvent<HTMLElement>): void => {
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = (): void => {
    setAnchorEl(null);
  };

  const handleAddResource = (element: Element): void => {
    if (onAddElement) {
      onAddElement(element);
    }
    handleMenuClose();
  };

  //   useOTPValidation(node as unknown as Node);
  //   useRecoveryFactorValidation(node as unknown as Node);

  useEffect(() => {
    if (!data?.components || data.components.length <= 0 || !stepId) {
      return;
    }

    updateNodeData(stepId, () => ({
      components: data?.components,
    }));
  }, [data?.components, stepId, updateNodeData]);

  return (
    // <ValidationErrorBoundary disableErrorBoundaryOnHover={false} resource={node}>
    <Box
      className={classNames('flow-builder-step', className)}
      sx={{
        overflow: 'hidden',
      }}
    >
      <Box
        display="flex"
        justifyContent="space-between"
        alignItems="center"
        className="flow-builder-step-action-panel"
        onDoubleClick={onActionPanelDoubleClick}
        sx={{
          backgroundColor: 'secondary.main',
          px: 2,
          py: 1.25,
          height: 44,
        }}
      >
        <Typography
          variant="body2"
          className="flow-builder-step-id"
          sx={{
            color: 'text.primary',
            fontWeight: 500,
          }}
        >
          {heading ?? 'View'}
        </Typography>
        <Box display="flex" gap={0.5}>
          <Tooltip
            title={
              // TODO: Add i18n
              'Add Component'
            }
          >
            <IconButton
              size="small"
              onClick={handleMenuOpen}
              className="flow-builder-step-action"
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
              <PlusIcon size={18} />
            </IconButton>
          </Tooltip>
          {configurable && (
            <Tooltip
              title={
                // TODO: Add i18n
                'Configure'
              }
            >
              <IconButton
                size="small"
                onClick={() => {
                  onConfigure?.();
                }}
                className="flow-builder-step-action"
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
          )}
          {deletable && (
            <Tooltip
              title={
                // TODO: Add i18n
                'Remove'
              }
            >
              <IconButton
                size="small"
                onClick={() => {
                  if (stepId) {
                    // eslint-disable-next-line @typescript-eslint/no-floating-promises
                    deleteElements({nodes: [{id: stepId}]});
                  }
                }}
                className="flow-builder-step-action"
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
                <TrashIcon size={18} />
              </IconButton>
            </Tooltip>
          )}
        </Box>
      </Box>

      {/* Context Menu for adding components */}
      <Menu
        anchorEl={anchorEl}
        open={menuOpen}
        onClose={handleMenuClose}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'right',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}
      >
        {availableElements && availableElements.length > 0 ? (
          availableElements.map((element: Element) => (
            <MenuItem
              key={element.id || element.type}
              onClick={() => handleAddResource(element)}
              sx={{
                minWidth: 200,
              }}
            >
              {element.display?.label || element.type}
            </MenuItem>
          ))
        ) : (
          <MenuItem disabled sx={{minWidth: 200}}>
            No components available
          </MenuItem>
        )}
      </Menu>

      <Handle type="target" position={Position.Left} />
      <Box className="flow-builder-step-content">
        <Paper
          className="flow-builder-step-content-box"
          elevation={0}
          sx={{
            backgroundColor: 'background.paper',
            borderRadius: 0,
            border: 'none',
            width: 350,
            minHeight: 100,
          }}
        >
          <Box className="flow-builder-step-content-form">
            <FormGroup>
              <Droppable
                id={generateResourceId(`${VisualFlowConstants.FLOW_BUILDER_VIEW_ID}_${stepId}`)}
                data={{droppedOn: node, stepId}}
                type={VisualFlowConstants.FLOW_BUILDER_DROPPABLE_VIEW_ID}
                accept={
                  droppableAllowedTypes
                    ? [
                        VisualFlowConstants.FLOW_BUILDER_DRAGGABLE_ID,
                        ...droppableAllowedTypes.filter((type: string) => !droppableRestrictedTypes?.includes(type)),
                      ]
                    : [
                        VisualFlowConstants.FLOW_BUILDER_DRAGGABLE_ID,
                        ...VisualFlowConstants.FLOW_BUILDER_VIEW_ALLOWED_RESOURCE_TYPES.filter(
                          (type: string) => !droppableRestrictedTypes?.includes(type),
                        ),
                      ]
                }
                collisionPriority={CollisionPriority.High}
              >
                {((node?.data?.components ?? data?.components) as Element[])?.map(
                  (component: Element, index: number) =>
                    PluginRegistry.getInstance().executeSync(FlowEventTypes.ON_NODE_ELEMENT_FILTER, component) && (
                      <ReorderableViewElement
                        key={component.id}
                        id={component.id}
                        index={index}
                        element={component}
                        className={classNames('flow-builder-step-content-form-field')}
                        type={VisualFlowConstants.FLOW_BUILDER_DRAGGABLE_ID}
                        accept={[
                          VisualFlowConstants.FLOW_BUILDER_DRAGGABLE_ID,
                          // Exclude INPUT from view-level sortables - inputs should only go inside forms
                          ...VisualFlowConstants.FLOW_BUILDER_VIEW_ALLOWED_RESOURCE_TYPES.filter(
                            (type) => type !== ElementTypes.Input,
                          ),
                        ]}
                        group={stepId ?? ''}
                        availableElements={availableElements}
                        onAddElementToForm={onAddElementToForm}
                      />
                    ),
                )}
              </Droppable>
            </FormGroup>
          </Box>
        </Paper>
      </Box>
      {enableSourceHandle && (
        <Handle
          type="source"
          position={Position.Right}
          id={`${stepId}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`}
        />
      )}
    </Box>
    // </ValidationErrorBoundary>
  );
}

export default View;
