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

import {CollisionPriority} from '@dnd-kit/abstract';
import {move} from '@dnd-kit/helpers';
import {DragDropProvider, type DragDropEventHandlers} from '@dnd-kit/react';
import {
  type Connection,
  type Edge,
  type Node,
  type OnEdgesChange,
  type OnNodesChange,
  useReactFlow,
  useUpdateNodeInternals,
} from '@xyflow/react';
import type {UpdateNodeInternals} from '@xyflow/system';
import {
  type Dispatch,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactElement,
  type SetStateAction,
} from 'react';
import {Box} from '@wso2/oxygen-ui';
import classNames from 'classnames';
import useIdentityProviders from '@/features/integrations/api/useIdentityProviders';
import useNotificationSenders from '@/features/notification-senders/api/useNotificationSenders';
import VisualFlow, {type VisualFlowPropsInterface} from './VisualFlow';
import Droppable from '../dnd/Droppable';
import VisualFlowConstants from '../../constants/VisualFlowConstants';
import generateResourceId from '../../utils/generateResourceId';
import {BlockTypes, type Element} from '../../models/elements';
import type {DragSourceData, DragTargetData, DragEventWithNative} from '../../models/drag-drop';
import {ResourceTypes, type Resource, type Resources} from '../../models/resources';
import FormRequiresViewDialog from './FormRequiresViewDialog';
import {type Step, type StepData} from '../../models/steps';
import {type Template} from '../../models/templates';
import type {Widget} from '../../models/widget';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import useComponentDelete from '../../hooks/useComponentDelete';
import useResourceAdd from '../../hooks/useResourceAdd';
import useGenerateStepElement from '../../hooks/useGenerateStepElement';
import useDeleteExecutionResource from '../../hooks/useDeleteExecutionResource';
import useStaticContentField from '../../hooks/useStaticContentField';
import useConfirmPasswordField from '../../hooks/useConfirmPasswordField';
import useVisualFlowHandlers from '../../hooks/useVisualFlowHandlers';
import useContainerDialogConfirm from '../../hooks/useContainerDialogConfirm';
import useDragDropHandlers from '../../hooks/useDragDropHandlers';
import applyAutoLayout from '../../utils/applyAutoLayout';
import {resolveCollisions} from '../../utils/resolveCollisions';
import ResourcePanel from '../resource-panel/ResourcePanel';
import ResourcePropertyPanel from '../resource-property-panel/ResourcePropertyPanel';
import ValidationPanel from '../validation-panel/ValidationPanel';
import computeExecutorConnections from '../../utils/computeExecutorConnections';
import type {MetadataInterface} from '../../models/metadata';

/**
 * Props interface of {@link DecoratedVisualFlow}
 */
export interface DecoratedVisualFlowPropsInterface extends Omit<VisualFlowPropsInterface, 'edgeTypes' | 'onSave'> {
  resources: Resources;
  edgeTypes?: VisualFlowPropsInterface['edgeTypes'];
  onEdgeResolve?: (connection: Connection, nodes: Node[]) => Edge;
  initialNodes?: Node[];
  initialEdges?: Edge[];
  nodes: Node[];
  edges: Edge[];
  mutateComponents: (components: Element[]) => Element[];
  onTemplateLoad: (template: Template) => [Node[], Edge[], Resource?, string?];
  onWidgetLoad: (
    widget: Widget,
    targetResource: Resource,
    currentNodes: Node[],
    edges: Edge[],
  ) => [Node[], Edge[], Resource | null, string | null];
  onStepLoad: (step: Step) => Step;
  onResourceAdd: (resource: Resource) => void;
  setNodes: Dispatch<SetStateAction<Node[]>>;
  setEdges: Dispatch<SetStateAction<Edge[]>>;
  onNodesChange: OnNodesChange<Node>;
  onEdgesChange: OnEdgesChange<Edge>;
  flowTitle: string;
  flowHandle: string;
  onFlowTitleChange: (newTitle: string) => void;
  onSave?: (canvasData: {nodes: Node[]; edges: Edge[]; viewport: {x: number; y: number; zoom: number}}) => void;
  /**
   * When true, triggers auto-layout on initial render if nodes lack proper layout data.
   * This is useful when loading flows that don't have saved canvas positions.
   */
  triggerAutoLayoutOnLoad?: boolean;
}

/**
 * Decorated visual flow component with drag-and-drop support.
 *
 * @param props - Props injected to the component.
 * @returns The DecoratedVisualFlow component.
 */
function DecoratedVisualFlow({
  resources,
  nodes,
  edges,
  setNodes,
  setEdges,
  onNodesChange,
  onEdgesChange,
  onEdgeResolve = undefined,
  edgeTypes = {},
  mutateComponents,
  onTemplateLoad,
  onWidgetLoad,
  onStepLoad,
  onSave = undefined,
  flowTitle,
  flowHandle,
  onFlowTitleChange,
  triggerAutoLayoutOnLoad = false,
  ...rest
}: DecoratedVisualFlowPropsInterface): ReactElement {
  useDeleteExecutionResource();
  useConfirmPasswordField();
  useStaticContentField();

  const {toObject, getNodes, getEdges, updateNodeData, fitView} = useReactFlow();
  const updateNodeInternals: UpdateNodeInternals = useUpdateNodeInternals();
  const {deleteComponent} = useComponentDelete();
  const {isResourcePanelOpen, isResourcePropertiesPanelOpen, isFlowMetadataLoading, metadata, onResourceDropOnCanvas} =
    useFlowBuilderCore();
  const {generateStepElement} = useGenerateStepElement();

  // Fetch identity providers and notification senders to compute executor connections
  const {data: identityProviders} = useIdentityProviders();
  const {data: notificationSenders} = useNotificationSenders();
  const computedMetadata: MetadataInterface | undefined = useMemo(() => {
    const executorConnections = computeExecutorConnections({identityProviders, notificationSenders});

    if (executorConnections.length === 0 && !metadata) {
      return undefined;
    }

    return {
      ...metadata,
      executorConnections: executorConnections.length > 0 ? executorConnections : (metadata?.executorConnections ?? []),
    } as MetadataInterface;
  }, [identityProviders, notificationSenders, metadata]);

  const [isContainerDialogOpen, setIsContainerDialogOpen] = useState<boolean>(false);
  const [dropScenario, setDropScenario] = useState<
    'form-on-canvas' | 'input-on-canvas' | 'input-on-view' | 'widget-on-canvas'
  >('form-on-canvas');

  const pendingDropRef = useRef<{
    event: DragEventWithNative;
    sourceData: DragSourceData;
    targetData: DragTargetData;
  } | null>(null);

  const handleContainerDialogClose = useCallback((): void => {
    setIsContainerDialogOpen(false);
    pendingDropRef.current = null;
  }, []);

  const handleContainerDialogConfirm = useContainerDialogConfirm({
    dropScenario,
    handleContainerDialogClose,
    generateStepElement,
    onStepLoad,
    setNodes,
    setEdges,
    onResourceDropOnCanvas,
    onWidgetLoad,
    metadata: computedMetadata,
    pendingDropRef,
  });

  const handleOnAdd = useResourceAdd({
    onTemplateLoad,
    onWidgetLoad,
    onStepLoad,
    setNodes,
    setEdges,
    generateStepElement,
    metadata: computedMetadata,
    onResourceDropOnCanvas,
  });

  const {handleConnect, handleNodesDelete, handleEdgesDelete} = useVisualFlowHandlers({
    onEdgeResolve,
    setEdges,
  });

  const {addCanvasNode, addToView, addToForm, addToViewAtIndex, addToFormAtIndex} = useDragDropHandlers({
    onStepLoad,
    setNodes,
    setEdges,
    onResourceDropOnCanvas,
    generateStepElement,
    mutateComponents,
    onWidgetLoad,
    metadata: computedMetadata,
  });

  // Memoized handleSave
  const handleSave = useCallback((): void => {
    const {viewport} = toObject();
    const canvasData = {
      nodes: getNodes(),
      edges: getEdges(),
      viewport,
    };
    onSave?.(canvasData);
  }, [toObject, getNodes, getEdges, onSave]);

  const handleAutoLayout = useCallback((): void => {
    const currentNodes = getNodes();
    const currentEdges = getEdges();
    applyAutoLayout(currentNodes, currentEdges, {
      direction: 'RIGHT',
      nodeSpacing: 150,
      rankSpacing: 300,
      offsetX: 50,
      offsetY: 50,
    })
      .then((layoutedNodes) => {
        setNodes(layoutedNodes);
        requestAnimationFrame(() => {
          fitView({padding: 0.2, duration: 300}).catch(() => {
            // Ignore fitView errors - layout is still applied
          });
        });
      })
      .catch(() => {
        // Layout failed, keep original positions
      });
  }, [getNodes, getEdges, setNodes, fitView]);

  // Track whether auto-layout has been triggered to prevent multiple triggers
  const autoLayoutTriggeredRef = useRef<boolean>(false);

  // Listen for auto-layout trigger events from parent components
  useEffect(() => {
    const handleTriggerAutoLayout = () => {
      handleAutoLayout();
    };

    window.addEventListener('triggerAutoLayout', handleTriggerAutoLayout);

    return () => {
      window.removeEventListener('triggerAutoLayout', handleTriggerAutoLayout);
    };
  }, [handleAutoLayout]);

  // Effect to trigger auto-layout on initial load when nodes lack proper layout data
  useEffect(() => {
    if (!triggerAutoLayoutOnLoad || autoLayoutTriggeredRef.current) {
      return;
    }

    const currentNodes = getNodes();

    // Skip if no nodes or only one node (nothing to layout)
    if (currentNodes.length <= 1) {
      return;
    }

    // Check if nodes need auto-layout by detecting if multiple nodes are at the same position
    // (which happens when layout data is missing and all default to {x: 0, y: 0})
    const nodesAtOrigin = currentNodes.filter((node) => node.position.x === 0 && node.position.y === 0);

    // If more than one node is at the origin, we need auto-layout
    const needsAutoLayout = nodesAtOrigin.length > 1;

    if (needsAutoLayout) {
      autoLayoutTriggeredRef.current = true;
      // Delay slightly to ensure nodes are fully rendered with their measured dimensions
      requestAnimationFrame(() => {
        handleAutoLayout();
      });
    }
  }, [triggerAutoLayoutOnLoad, getNodes, handleAutoLayout]);

  const handleNodeDragStop = useCallback((): void => {
    const currentNodes = getNodes();
    const resolvedNodes = resolveCollisions(currentNodes, {
      maxIterations: 50,
      overlapThreshold: 0.5,
      margin: 50,
    });

    // Check if any nodes were moved by collision resolution
    const hasChanges = resolvedNodes.some(
      (resolvedNode: Node, index: number) =>
        resolvedNode.position.x !== currentNodes[index].position.x ||
        resolvedNode.position.y !== currentNodes[index].position.y,
    );

    if (hasChanges) {
      setNodes(resolvedNodes);
    }
  }, [getNodes, setNodes]);

  const handleDragEnd: DragDropEventHandlers['onDragEnd'] = useCallback(
    (event): void => {
      const {source, target} = event.operation;

      if (!source) {
        return;
      }

      const sourceData: DragSourceData = source.data as DragSourceData;
      const targetData = (target?.data ?? {}) as DragTargetData;

      // Check for components that need containers
      const isFormDrop = sourceData.dragged?.type === BlockTypes.Form;
      const isInputDrop = sourceData.dragged?.category === 'FIELD';
      const isWidgetDrop = sourceData.dragged?.resourceType === ResourceTypes.Widget;
      const isCanvasTarget =
        typeof target?.id === 'string' && target.id.startsWith(VisualFlowConstants.FLOW_BUILDER_CANVAS_ID);
      const isViewTarget =
        typeof target?.id === 'string' && target.id.startsWith(VisualFlowConstants.FLOW_BUILDER_VIEW_ID);

      // Form dropped on canvas -> needs View
      if (isFormDrop && isCanvasTarget) {
        pendingDropRef.current = {event, sourceData, targetData};
        setDropScenario('form-on-canvas');
        setIsContainerDialogOpen(true);
        return;
      }

      // Input dropped on canvas -> needs View + Form
      if (isInputDrop && isCanvasTarget) {
        pendingDropRef.current = {event, sourceData, targetData};
        setDropScenario('input-on-canvas');
        setIsContainerDialogOpen(true);
        return;
      }

      // Input dropped on View -> needs Form
      if (isInputDrop && isViewTarget) {
        pendingDropRef.current = {event, sourceData, targetData};
        setDropScenario('input-on-view');
        setIsContainerDialogOpen(true);
        return;
      }

      // Widget dropped on canvas -> needs View
      if (isWidgetDrop && isCanvasTarget) {
        pendingDropRef.current = {event, sourceData, targetData};
        setDropScenario('widget-on-canvas');
        setIsContainerDialogOpen(true);
        return;
      }

      // Check if this is a step being added to canvas (not reordering)
      const isStepDrop = sourceData.dragged?.resourceType === ResourceTypes.Step;
      if (isStepDrop && isCanvasTarget && !sourceData.isReordering) {
        // Dispatch custom event to notify about element addition (for auto-layout hint)
        window.dispatchEvent(new CustomEvent('flowElementAdded', {detail: {type: 'step'}}));
      }

      // For canceled events or missing target, return early
      if (event.canceled || !target) {
        return;
      }

      // Handle reordering
      if (sourceData.isReordering) {
        if (!sourceData.stepId) {
          return;
        }

        updateNodeData(sourceData.stepId, (node: Node) => {
          const unorderedComponents: Element[] = (node?.data as StepData)?.components ?? [];

          const reorderedNested = unorderedComponents.map((component: Element) => {
            if (component?.components) {
              return {
                ...component,
                components: move([...component.components], event),
              };
            }

            return component;
          });

          return {
            components: move(reorderedNested, event),
          };
        });

        requestAnimationFrame(() => {
          updateNodeInternals(sourceData.stepId!);
        });

        return;
      }

      // Handle dropping on canvas
      if (typeof target?.id === 'string' && target.id.startsWith(VisualFlowConstants.FLOW_BUILDER_CANVAS_ID)) {
        addCanvasNode(event, sourceData, targetData);
        return;
      }

      // Handle dropping on View
      if (typeof target?.id === 'string' && target.id.startsWith(VisualFlowConstants.FLOW_BUILDER_VIEW_ID)) {
        addToView(event, sourceData, targetData);
        return;
      }

      // Handle dropping on Form
      if (typeof target?.id === 'string' && target.id.startsWith(VisualFlowConstants.FLOW_BUILDER_FORM_ID)) {
        addToForm(event, sourceData, targetData);
        return;
      }

      // Handle dropping on an existing element (at specific position)
      if (targetData.isReordering && targetData.stepId && typeof target?.id === 'string') {
        // Dropping on an existing sortable element - insert at that position
        const targetElementId = target.id;

        // Check if the target element is inside a form by searching through all forms
        const targetNode = getNodes().find((n) => n.id === targetData.stepId);
        const nodeData = targetNode?.data as StepData | undefined;
        const parentForm = nodeData?.components?.find(
          (c: Element) =>
            c.type === BlockTypes.Form && c.components?.some((child: Element) => child.id === targetElementId),
        );

        if (parentForm) {
          // Phase 1.6: Target element is inside a form, insert at that position within the form
          addToFormAtIndex(sourceData, targetData.stepId, parentForm.id, targetElementId);
        } else {
          // Phase 1.5: Target is a top-level element in the view, add to view at index
          addToViewAtIndex(sourceData, targetData.stepId, targetElementId);
        }
      }
    },
    [
      updateNodeData,
      updateNodeInternals,
      addCanvasNode,
      addToView,
      addToForm,
      addToViewAtIndex,
      addToFormAtIndex,
      getNodes,
    ],
  );

  const handleDragOver: DragDropEventHandlers['onDragOver'] = useCallback(
    (event) => {
      const {source, target} = event.operation;

      if (!source || !target) {
        return;
      }

      if (!source.data.isReordering) {
        return;
      }

      const {data: sourceData} = source;
      const stepId = (sourceData as DragSourceData)?.stepId;

      if (!stepId) {
        return;
      }

      updateNodeData(stepId, (node: Node) => {
        const nodeData = node?.data as StepData | undefined;
        const unorderedComponents: Element[] = nodeData?.components ?? [];

        const reorderedNested = unorderedComponents.map((component: Element) => {
          if (component?.components) {
            return {
              ...component,
              components: move([...component.components], event),
            };
          }

          return component;
        });

        return {
          components: move(reorderedNested, event),
        };
      });

      requestAnimationFrame(() => {
        updateNodeInternals(stepId);
      });
    },
    [updateNodeData, updateNodeInternals],
  );

  return (
    <Box
      className={classNames('decorated-visual-flow', 'react-flow-container')}
      sx={(theme) => ({
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        backgroundColor: 'var(--xy-background-color-default)',
        p: 1,
        ...theme.applyStyles('dark', {
          backgroundColor: 'var(--xy-background-color-default)',
        }),
      })}
    >
      <Box sx={{flexGrow: 1, minHeight: 0}}>
        <DragDropProvider onDragEnd={handleDragEnd} onDragOver={handleDragOver}>
          <ResourcePanel
            resources={resources}
            open={isResourcePanelOpen}
            onAdd={handleOnAdd}
            disabled={isFlowMetadataLoading}
            flowTitle={flowTitle}
            flowHandle={flowHandle}
            onFlowTitleChange={onFlowTitleChange}
          >
            <ResourcePropertyPanel open={isResourcePropertiesPanelOpen} onComponentDelete={deleteComponent}>
              <Droppable
                id={generateResourceId(VisualFlowConstants.FLOW_BUILDER_CANVAS_ID)}
                type={VisualFlowConstants.FLOW_BUILDER_DROPPABLE_CANVAS_ID}
                accept={[...VisualFlowConstants.FLOW_BUILDER_CANVAS_ALLOWED_RESOURCE_TYPES]}
                collisionPriority={CollisionPriority.Low}
              >
                <VisualFlow
                  nodes={nodes}
                  onNodesChange={onNodesChange}
                  edges={edges}
                  edgeTypes={edgeTypes}
                  onEdgesChange={onEdgesChange}
                  onConnect={handleConnect}
                  onNodesDelete={handleNodesDelete}
                  onEdgesDelete={handleEdgesDelete}
                  onNodeDragStop={handleNodeDragStop}
                  handleAutoLayout={handleAutoLayout}
                  onSave={handleSave}
                  {...rest}
                />
              </Droppable>
              <ValidationPanel />
            </ResourcePropertyPanel>
          </ResourcePanel>
        </DragDropProvider>
      </Box>

      <FormRequiresViewDialog
        open={isContainerDialogOpen}
        scenario={dropScenario}
        onClose={handleContainerDialogClose}
        onConfirm={handleContainerDialogConfirm}
      />
    </Box>
  );
}

export default DecoratedVisualFlow;
