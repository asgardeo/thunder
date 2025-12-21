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
import {useRef, useState, useMemo, memo, type MouseEvent, type ReactElement} from 'react';
import PluginRegistry from '@/features/flows/plugins/PluginRegistry';
import FlowEventTypes from '@/features/flows/models/extension';
import classNames from 'classnames';
import {GripVertical, PencilLineIcon, PlusIcon, Trash2Icon} from '@wso2/oxygen-ui-icons-react';
import useComponentDelete from '@/features/flows/hooks/useComponentDelete';
import useValidationStatus from '@/features/flows/hooks/useValidationStatus';
import Notification, {NotificationType} from '@/features/flows/models/notification';
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
   * @defaultValue undefined
   */
  availableElements?: Resource[];
  /**
   * Callback for adding an element to a form.
   * @param element - The element to add.
   * @param formId - The ID of the form to add to.
   * @defaultValue undefined
   */
  onAddElementToForm?: (element: Resource, formId: string) => void;
}

/**
 * Re-orderable component inside a step node.
 *
 * @param props - Props injected to the component.
 * @returns ReorderableElement component.
 */
function ReorderableElement({
  id,
  index,
  element,
  className,
  availableElements = undefined,
  onAddElementToForm = undefined,
  ...rest
}: ReorderableComponentPropsInterface): ReactElement {
  const handleRef = useRef<HTMLButtonElement>(null);
  const stepId: string | null = useNodeId();
  const {deleteComponent} = useComponentDelete();
  const {ElementFactory, setLastInteractedResource, setLastInteractedStepId, setIsOpenResourcePropertiesPanel} =
    useFlowBuilderCore();
  const {setOpenValidationPanel, setSelectedNotification, addNotification} = useValidationStatus();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const menuOpen = Boolean(anchorEl);

  // Check if this element is a Form
  const isForm = element.type === BlockTypes.Form;

  const depsRef = useRef({
    element,
    onAddElementToForm,
    availableElements,
    stepId,
    setOpenValidationPanel,
    setSelectedNotification,
    addNotification,
    setLastInteractedStepId,
    setLastInteractedResource,
    deleteComponent,
    setIsOpenResourcePropertiesPanel,
    setAnchorEl,
  });

  // Update refs every render (minimal overhead - just assignment)
  depsRef.current = {
    element,
    onAddElementToForm,
    availableElements,
    stepId,
    setOpenValidationPanel,
    setSelectedNotification,
    addNotification,
    setLastInteractedStepId,
    setLastInteractedResource,
    deleteComponent,
    setIsOpenResourcePropertiesPanel,
    setAnchorEl,
  };

  // Store stable references to handler functions
  const handlersRef = useRef<{
    handlePropertyPanelOpen: (event: React.MouseEvent<HTMLElement>) => void;
    handleElementDelete: () => void;
    handleMenuOpen: (event: MouseEvent<HTMLElement>) => void;
    handleMenuClose: () => void;
    handleAddFieldToForm: (fieldElement: Resource) => void;
  } | null>(null);

  // Create handlers only once using lazy initialization - reads ALL deps from ref at call time
  handlersRef.current ??= {
    handlePropertyPanelOpen: (event: React.MouseEvent<HTMLElement>): void => {
      event.stopPropagation();
      const deps = depsRef.current;
      deps.setOpenValidationPanel?.(false);
      deps.setSelectedNotification?.(null);
      if (deps.stepId) {
        deps.setLastInteractedStepId(deps.stepId);
      }
      deps.setLastInteractedResource(deps.element);
    },

    handleElementDelete: (): void => {
      const deps = depsRef.current;
      PluginRegistry.getInstance()
        .executeAsync(FlowEventTypes.ON_NODE_ELEMENT_DELETE, deps.stepId, deps.element)
        .then(() => {
          if (deps.stepId) {
            deps.deleteComponent(deps.stepId, deps.element);
          }
          deps.setIsOpenResourcePropertiesPanel(false);
        })
        .catch((error: Error) => {
          const errorNotification = new Notification(
            `delete-element-error-${deps.element.id}`,
            `Failed to delete element: ${error.message}`,
            NotificationType.ERROR,
          );
          deps.addNotification?.(errorNotification);
        });
    },

    handleMenuOpen: (event: MouseEvent<HTMLElement>): void => {
      event.stopPropagation();
      depsRef.current.setAnchorEl(event.currentTarget);
    },

    handleMenuClose: (): void => {
      depsRef.current.setAnchorEl(null);
    },

    handleAddFieldToForm: (fieldElement: Resource): void => {
      const deps = depsRef.current;
      if (deps.onAddElementToForm) {
        deps.onAddElementToForm(fieldElement, deps.element.id);
      }
      deps.setAnchorEl(null);
    },
  };

  // Extract stable handlers
  const {handlePropertyPanelOpen, handleElementDelete, handleMenuOpen, handleMenuClose, handleAddFieldToForm} =
    handlersRef.current;

  // Filter available elements to only show form-compatible types that are visible on resource panel
  const formCompatibleElements = useMemo(
    () =>
      isForm && depsRef.current.availableElements
        ? depsRef.current.availableElements.filter(
            (el: Resource) =>
              VisualFlowConstants.FLOW_BUILDER_FORM_ALLOWED_RESOURCE_TYPES.includes(el.type) &&
              el.display?.showOnResourcePanel !== false,
          )
        : [],
    [isForm],
  );

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
          {/* eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions */}
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
            formCompatibleElements.map((fieldElement: Resource) => (
              <MenuItem
                key={`${fieldElement.type}-${fieldElement.id}-${typeof fieldElement.variant === 'string' ? fieldElement.variant : ''}`}
                onClick={() => handleAddFieldToForm(fieldElement)}
                sx={{
                  minWidth: 200,
                }}
              >
                {fieldElement.display?.label ?? fieldElement.type}
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

// Only re-render if element.id or element properties actually change
const MemoizedReorderableElement = memo(ReorderableElement, (prevProps, nextProps) => {
  // Re-render if element changed (compare by reference and key props)
  if (prevProps.element !== nextProps.element) {
    return false;
  }
  // Re-render if id or index changed
  if (prevProps.id !== nextProps.id || prevProps.index !== nextProps.index) {
    return false;
  }
  // Re-render if className changed
  if (prevProps.className !== nextProps.className) {
    return false;
  }
  // Don't re-render for availableElements or onAddElementToForm changes
  // (handlers read from refs, so they don't need to trigger re-renders)
  return true;
});

export {MemoizedReorderableElement as ReorderableElement};
export default MemoizedReorderableElement;
