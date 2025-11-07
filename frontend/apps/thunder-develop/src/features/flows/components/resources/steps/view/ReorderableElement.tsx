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

import {Box, Menu, MenuItem, type BoxProps} from '@wso2/oxygen-ui';
import {useRef, useState, type MouseEvent, type ReactElement} from 'react';
import PluginRegistry from '@/features/flows/plugins/PluginRegistry';
import FlowEventTypes from '@/features/flows/models/extension';
import classNames from 'classnames';
import {GripVertical, PencilLineIcon, PlusIcon, Trash2Icon} from '@wso2/oxygen-ui-icons-react';
import useComponentDelete from '@/features/flows/hooks/useComponentDelete';
import useValidationStatus from '@/features/flows/hooks/useValidationStatus';
import {useNodeId} from '@xyflow/react';
import useFlowBuilderCore from '@/features/flows/hooks/useFlowBuilderCore';
import type {Resource} from '@/features/flows/models/resources';
import {BlockTypes} from '@/features/flows/models/elements';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import Handle from '../../../dnd/handle';
import Sortable from '../../../dnd/sortable';
import type {SortableProps} from '../../../dnd/sortable';
import ValidationErrorBoundary from '../../../validation-panel/ValidationErrorBoundary';

/**
 * Props interface of {@link ReorderableElement}
 */
export interface ReorderableComponentPropsInterface
  extends Omit<SortableProps, 'element'>,
    Omit<BoxProps, 'children' | 'id'> {
  /**
   * The element to be rendered.
   */
  element: Resource;
  /**
   * List of available elements that can be added.
   */
  availableElements?: Resource[];
  /**
   * Callback for adding an element to a form.
   * @param element - The element to add.
   * @param formId - The ID of the form to add to.
   */
  onAddElementToForm?: (element: Resource, formId: string) => void;
}

/**
 * Re-orderable component inside a step node.
 *
 * @param props - Props injected to the component.
 * @returns ReorderableElement component.
 */
export function ReorderableElement({
  id,
  index,
  element,
  className,
  availableElements,
  onAddElementToForm,
  ...rest
}: ReorderableComponentPropsInterface): ReactElement {
  const handleRef = useRef<HTMLButtonElement>(null);
  const stepId: string | null = useNodeId();
  const {deleteComponent} = useComponentDelete();
  const {ElementFactory, setLastInteractedResource, setLastInteractedStepId, setIsOpenResourcePropertiesPanel} =
    useFlowBuilderCore();
  const {setOpenValidationPanel, setSelectedNotification} = useValidationStatus();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const menuOpen = Boolean(anchorEl);

  // Check if this element is a Form
  const isForm = element.type === BlockTypes.Form;

  /**
   * Handles the opening of the property panel for the resource.
   *
   * @param event - React MouseEvent triggered on element interaction.
   */
  const handlePropertyPanelOpen = (event: React.MouseEvent<HTMLElement>): void => {
    event.stopPropagation();
    setOpenValidationPanel?.(false);
    setSelectedNotification?.(null);
    if (stepId) {
      setLastInteractedStepId(stepId);
    }
    setLastInteractedResource(element);
  };

  /**
   * Handles the deletion of the element.
   */
  const handleElementDelete = (): void => {
    /**
     * Execute plugins for ON_NODE_ELEMENT_DELETE event and handle deletion.
     */
    PluginRegistry.getInstance()
      .executeAsync(FlowEventTypes.ON_NODE_ELEMENT_DELETE, stepId, element)
      .then(() => {
        if (stepId) {
          deleteComponent(stepId, element);
        }
        setIsOpenResourcePropertiesPanel(false);
      })
      .catch((error: Error) => {
        // TODO: Handle error with proper error notification
        throw error;
      });
  };

  /**
   * Handles opening the add field menu for Forms.
   */
  const handleMenuOpen = (event: MouseEvent<HTMLElement>): void => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
  };

  /**
   * Handles closing the add field menu.
   */
  const handleMenuClose = (): void => {
    setAnchorEl(null);
  };

  /**
   * Handles adding an element to the form.
   */
  const handleAddFieldToForm = (fieldElement: Resource): void => {
    if (onAddElementToForm) {
      onAddElementToForm(fieldElement, element.id);
    }
    handleMenuClose();
  };

  // Filter available elements to only show form-compatible types
  const formCompatibleElements =
    isForm && availableElements
      ? availableElements.filter((el: Resource) =>
          VisualFlowConstants.FLOW_BUILDER_FORM_ALLOWED_RESOURCE_TYPES.includes(el.type),
        )
      : [];

  return (
    <Sortable
      id={id}
      index={index}
      handleRef={handleRef}
      data={{isReordering: true, resource: element, stepId}}
      {...rest}
    >
      <ValidationErrorBoundary resource={element} key={element.id}>
        <Box
          display="flex"
          alignItems="center"
          className={classNames('reorderable-component', className)}
          onDoubleClick={handlePropertyPanelOpen}
        >
          <Box className="flow-builder-dnd-actions">
            <Handle label="Drag" cursor="grab" ref={handleRef}>
              <GripVertical size={16} color="white" />
            </Handle>
            <Handle label="Edit" onClick={handlePropertyPanelOpen}>
              <PencilLineIcon size={16} color="white" />
            </Handle>
            {isForm && formCompatibleElements.length > 0 && (
              <Handle label="Add Field" onClick={handleMenuOpen}>
                <PlusIcon size={16} color="white" />
              </Handle>
            )}
            <Handle label="Delete" onClick={handleElementDelete}>
              <Trash2Icon size={16} color="white" />
            </Handle>
          </Box>
          <div className="flow-builder-step-content-form-field-content" onClick={handlePropertyPanelOpen}>
            <ElementFactory
              stepId={stepId ?? ''}
              resource={element}
              elementIndex={index}
              availableElements={availableElements}
              onAddElementToForm={onAddElementToForm}
            />
          </div>
        </Box>
      </ValidationErrorBoundary>

      {/* Menu for adding fields to Form */}
      {isForm && (
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
          {formCompatibleElements && formCompatibleElements.length > 0 ? (
            formCompatibleElements.map((fieldElement: Resource, index: number) => (
              <MenuItem
                key={`${fieldElement.type}-${fieldElement.variant || ''}-${index}`}
                onClick={() => handleAddFieldToForm(fieldElement)}
                sx={{
                  minWidth: 200,
                }}
              >
                {fieldElement.display?.label || fieldElement.type}
              </MenuItem>
            ))
          ) : (
            <MenuItem disabled sx={{minWidth: 200}}>
              No fields available
            </MenuItem>
          )}
        </Menu>
      )}
    </Sortable>
  );
}

export default ReorderableElement;
