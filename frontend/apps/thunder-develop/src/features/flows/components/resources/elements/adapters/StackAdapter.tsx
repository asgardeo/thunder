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

import {useMemo, type ReactElement} from 'react';
import classNames from 'classnames';
import {Box, Typography, type SxProps, type Theme} from '@wso2/oxygen-ui';
import {CollisionPriority} from '@dnd-kit/abstract';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import {type Element as FlowElement} from '@/features/flows/models/elements';
import FlowEventTypes from '@/features/flows/models/extension';
import PluginRegistry from '@/features/flows/plugins/PluginRegistry';
import generateResourceId from '@/features/flows/utils/generateResourceId';
import ReorderableFlowElement from '../../steps/view/ReorderableElement';
import Droppable from '../../../dnd/Droppable';

/**
 * Stack element type with layout configuration at top level.
 * When `columns` >= 2, the stack uses CSS Grid instead of flexbox.
 */
export type StackElement = FlowElement & {
  direction?: 'row' | 'column';
  gap?: number;
  align?: string;
  justify?: string;
  /** Number of equal slots. 1 = flex mode, ≥2 = grid mode with that many slots. */
  items?: number;
};

/**
 * Props interface of {@link StackAdapter}
 */
export interface StackAdapterPropsInterface {
  /**
   * The step id the resource resides on.
   */
  stepId: string;
  /**
   * The stack element properties.
   */
  resource: FlowElement;
  /**
   * List of available elements that can be added.
   */
  availableElements?: FlowElement[];
  /**
   * Callback for adding an element to a form.
   * @param element - The element to add.
   * @param formId - The ID of the form to add to.
   */
  onAddElementToForm?: (element: FlowElement, formId: string) => void;
}

/**
 * sx applied to every grid cell — browser-inspector-style tinted slot.
 */
const SLOT_SX: SxProps<Theme> = {
  borderRadius: 1,
  border: '0px dashed',
  borderColor: 'divider',
  backgroundColor: 'rgba(var(--mui-palette-primary-mainChannel) / 0.04)',
  overflow: 'visible',
  width: '100%',
  // Floating action toolbar: reposition above the cell so it never obscures
  // compact content (buttons, short labels, etc.).
  '& .reorderable-component': {
    position: 'relative',
    p: '4px 6px',
    borderRadius: 1,
    border: '2px dashed transparent',
    '& .flow-builder-dnd-actions': {
      visibility: 'hidden',
      position: 'absolute',
      top: '-30px',
      right: 0,
      height: '26px',
      width: 'auto',
      minWidth: '72px',
      display: 'flex',
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '2px',
      background: 'var(--flow-builder-dragging-form-field-action-panel-background-color)',
      borderRadius: '4px 4px 4px 0',
      zIndex: 20,
      pointerEvents: 'none',
      '& svg': {pointerEvents: 'auto'},
    },
    '&:hover, &:focus-within': {
      borderColor: 'var(--flow-builder-dragging-form-field-hover-border-color)',
      backgroundColor: 'var(--flow-builder-dragging-form-field-hover-background-color)',
      '& .flow-builder-dnd-actions': {visibility: 'visible'},
    },
  },
};

/**
 * sx for empty placeholder slots.
 */
const EMPTY_SLOT_SX: SxProps<Theme> = {
  ...SLOT_SX,
  backgroundColor: 'action.hover',
  minHeight: '44px',
  padding: '8px',
  transition: 'background-color 150ms ease',
  '&:hover': {backgroundColor: 'action.selected'},
};

/**
 * Adapter for the Stack layout container.
 * Renders children in a configurable flex or grid layout with a droppable
 * zone for adding elements via drag-and-drop — same drop behaviour as Form.
 *
 * @param props - Props injected to the component.
 * @returns The StackAdapter component.
 */
function StackAdapter({
  stepId,
  resource,
  availableElements = [],
  onAddElementToForm = undefined,
}: StackAdapterPropsInterface): ReactElement {
  const stackElement = resource as StackElement;
  const items = stackElement?.items ?? 1;
  const useGrid = items >= 2;

  const filteredComponents = useMemo(() => {
    if (!resource?.components) return [];
    return resource.components.filter((component: FlowElement) =>
      PluginRegistry.getInstance().executeSync(FlowEventTypes.ON_NODE_ELEMENT_FILTER, component),
    );
  }, [resource?.components]);

  // In grid mode: always fill defined slots — show placeholders for every unoccupied slot.
  // In flex mode: show a single placeholder only when there are no children.
  const emptySlotCount = useGrid
    ? Math.max(0, items - filteredComponents.length)
    : filteredComponents.length === 0
      ? 1
      : 0;

  const layoutSx: SxProps<Theme> = useGrid
    ? {
        display: 'grid',
        gridTemplateColumns: `repeat(${items}, 1fr)`,
        gap: stackElement?.gap ?? 1,
        alignItems: stackElement?.align ?? 'start',
        px: 2,
      }
    : {
        display: 'flex',
        flexDirection: stackElement?.direction ?? 'row',
        flexWrap: 'wrap',
        gap: stackElement?.gap ?? 2,
        alignItems: stackElement?.align ?? 'center',
        justifyContent: stackElement?.justify ?? 'center',
        px: 2,
      };

  return (
    <Droppable
      id={generateResourceId(`${VisualFlowConstants.FLOW_BUILDER_STACK_ID}_${resource.id}`)}
      data={{droppedOn: resource, stepId}}
      collisionPriority={CollisionPriority.High}
      type={VisualFlowConstants.FLOW_BUILDER_DROPPABLE_STACK_ID}
      accept={[
        VisualFlowConstants.FLOW_BUILDER_DRAGGABLE_ID,
        ...VisualFlowConstants.FLOW_BUILDER_STACK_ALLOWED_RESOURCE_TYPES,
      ]}
      sx={layoutSx}
      bottomZoneMinHeight={0}
    >
      {filteredComponents.map((component: FlowElement, index: number) => (
        <ReorderableFlowElement
          key={component.id}
          id={component.id}
          index={index}
          element={component}
          className={classNames('flow-builder-step-content-form-field')}
          group={resource.id}
          type={VisualFlowConstants.FLOW_BUILDER_DRAGGABLE_ID}
          accept={[
            VisualFlowConstants.FLOW_BUILDER_DRAGGABLE_ID,
            ...VisualFlowConstants.FLOW_BUILDER_STACK_ALLOWED_RESOURCE_TYPES,
          ]}
          availableElements={availableElements}
          onAddElementToForm={onAddElementToForm}
          sx={SLOT_SX}
          dropIndicatorStyles={{
            width: '100%',
          }}
          slotProps={{
            ContentContainer: {
              sx: {
                alignItems: 'center',
              },
            },
          }}
        />
      ))}
      {Array.from({length: emptySlotCount}, (_, i) => (
        <Box key={`stack-empty-${i}`} sx={EMPTY_SLOT_SX} display="flex" alignItems="center" justifyContent="center">
          <Typography variant="caption" color="text.disabled">
            Drop here
          </Typography>
        </Box>
      ))}
    </Droppable>
  );
}

export default StackAdapter;
