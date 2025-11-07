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

/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
import {move} from '@dnd-kit/helpers';
import {DragDropProvider, type DragDropEventHandlers} from '@dnd-kit/react';
import {
  type Connection,
  type Edge,
  MarkerType,
  type Node,
  type OnConnect,
  type OnEdgesChange,
  type OnNodesChange,
  type OnNodesDelete,
  type XYPosition,
  addEdge,
  getConnectedEdges,
  getIncomers,
  getOutgoers,
  useReactFlow,
  useUpdateNodeInternals,
} from '@xyflow/react';
import type {UpdateNodeInternals} from '@xyflow/system';
import cloneDeep from 'lodash-es/cloneDeep';
import {type Dispatch, useCallback, type ReactElement, type SetStateAction, useState, useRef} from 'react';
import {Box} from '@wso2/oxygen-ui';
import classNames from 'classnames';
import VisualFlow, {type VisualFlowPropsInterface} from './VisualFlow';
import {BlockTypes, ElementCategories, ElementTypes, type Element} from '../../models/elements';
import FormRequiresViewDialog, {type DropScenario} from '../form-requires-view-dialog/FormRequiresViewDialog';
import {type Resource, type Resources, ResourceTypes} from '../../models/resources';
import {StepTypes, type Step, type StepData} from '../../models/steps';
import {type Template} from '../../models/templates';
import type {Widget} from '../../models/widget';
import type {DragSourceData, DragTargetData} from '../../models/drag-drop';
import PluginRegistry from '../../plugins/PluginRegistry';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import generateResourceId from '../../utils/generateResourceId';
import autoAssignConnections from '../../utils/autoAssignConnections';
import useGenerateStepElement from '../../hooks/useGenerateStepElement';
import useDeleteExecutionResource from '../../hooks/useDeleteExecutionResource';
import useStaticContentField from '../../hooks/useStaticContentField';
import useConfirmPasswordField from '../../hooks/useConfirmPasswordField';
import VisualFlowConstants from '../../constants/VisualFlowConstants';
import ResourcePanel from '../resource-panel/ResourcePanel';
import HeaderPanel from '../header-panel/HeaderPanel';
import FlowEventTypes from '../../models/extension';
// import ValidationPanel from '../validation-panel/ValidationPanel';
import ResourcePropertyPanel from '../resource-property-panel/ResourcePropertyPanel';
import useComponentDelete from '../../hooks/useComponentDelete';
import applyAutoLayout from '../../utils/applyAutoLayout';
import {resolveCollisions} from '../../utils/resolveCollisions';

/**
 * Props interface of {@link DecoratedVisualFlow}
 */
export interface DecoratedVisualFlowPropsInterface extends VisualFlowPropsInterface {
  /**
   * Flow resources.
   */
  resources: Resources;
  /**
   * Callback to be fired when an edge is resolved.
   * @param connection - Connection object.
   * @returns Edge object.
   */
  onEdgeResolve?: (connection: Connection, nodes: Node[]) => Edge;
  /**
   * Initial nodes and edges to be rendered.
   */
  initialNodes?: Node[];
  /**
   * Initial nodes and edges to be rendered.
   */
  initialEdges?: Edge[];
  /**
   * Current nodes in the flow.
   */
  nodes: Node[];
  /**
   * Current edges in the flow.
   */
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
  /**
   * Flag to control header panel visibility.
   */
  isHeaderPanelOpen?: boolean;
  /**
   * Content to display in the header panel.
   */
  headerPanelContent?: ReactElement | null;
  /**
   * Callback to be triggered when back button is clicked.
   */
  onBack?: () => void;
  /**
   * Callback to be triggered when save button is clicked.
   * Receives the current canvas data (nodes, edges, viewport) from React Flow.
   */
  onSave?: (canvasData: {nodes: Node[]; edges: Edge[]; viewport: {x: number; y: number; zoom: number}}) => void;
  /**
   * Title for the flow builder.
   */
  flowTitle?: string;
}

/**
 * Component to decorate the visual flow editor with the necessary providers.
 *
 * @param props - Props injected to the component.
 * @returns Decorated visual flow component.
 */
function DecoratedVisualFlow({
  resources,
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  initialNodes = [],
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  initialEdges = [],
  setNodes,
  setEdges,
  edges,
  nodes,
  onNodesChange,
  onEdgesChange,
  onEdgeResolve = undefined,
  mutateComponents,
  onTemplateLoad,
  onWidgetLoad,
  onStepLoad,
  onResourceAdd,
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  headerPanelContent = null,
  onSave = undefined,
  ...rest
}: DecoratedVisualFlowPropsInterface): ReactElement {
  // Event handlers for ON_NODE_DELETE event.
  useDeleteExecutionResource();

  // Event handlers for ON_PROPERTY_PANEL_OPEN event.
  useConfirmPasswordField();

  // Event handlers for static content in execution steps.
  useStaticContentField();

  const {screenToFlowPosition, updateNodeData, toObject, fitView} = useReactFlow();
  const {generateStepElement} = useGenerateStepElement();
  const updateNodeInternals: UpdateNodeInternals = useUpdateNodeInternals();
  const {deleteComponent} = useComponentDelete();

  const {isResourcePanelOpen, isResourcePropertiesPanelOpen, onResourceDropOnCanvas, isFlowMetadataLoading, metadata} =
    useFlowBuilderCore();

  // State for the container requirements dialog
  const [isContainerDialogOpen, setIsContainerDialogOpen] = useState<boolean>(false);
  const [dropScenario, setDropScenario] = useState<DropScenario>('form-on-canvas');
  const pendingDropRef = useRef<{
    event: Parameters<DragDropEventHandlers['onDragEnd']>[0];
    sourceData: DragSourceData;
    targetData: DragTargetData;
  } | null>(null);

  const addCanvasNode = (
    event: Parameters<DragDropEventHandlers['onDragEnd']>[0],
    sourceData: DragSourceData,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    _targetData: DragTargetData,
  ): void => {
    const sourceResource: Resource | undefined = cloneDeep(sourceData.dragged);

    if (!sourceResource || !event.nativeEvent) {
      return;
    }

    // Type guard to ensure nativeEvent is a MouseEvent
    const {nativeEvent} = event;
    if (!('clientX' in nativeEvent) || !('clientY' in nativeEvent)) {
      return;
    }

    const {clientX, clientY} = nativeEvent as MouseEvent;

    const position: XYPosition = screenToFlowPosition({
      x: clientX,
      y: clientY,
    });

    let generatedStep: Step = {
      ...sourceResource,
      data: {
        components: [],
        ...(sourceResource?.data ?? {}),
      },
      deletable: true,
      id: generateResourceId(sourceResource.type.toLowerCase()),
      position,
    } as Step;

    // Decorate the step with any additional information
    generatedStep = onStepLoad(generatedStep);

    setNodes((prevNodes: Node[]) => [...prevNodes, generatedStep]);

    onResourceDropOnCanvas(generatedStep, '');
  };

  const addToView = (
    event: Parameters<DragDropEventHandlers['onDragEnd']>[0],
    sourceData: DragSourceData,
    targetData: DragTargetData,
  ): void => {
    const {dragged: sourceResource} = sourceData;
    const {stepId: targetStepId, droppedOn: targetResource} = targetData;

    if (sourceResource?.resourceType === ResourceTypes.Widget && nodes && edges && targetResource) {
      const [newNodes, newEdges, defaultPropertySelector, defaultPropertySectorStepId] = onWidgetLoad(
        sourceResource as Widget,
        targetResource,
        nodes,
        edges,
      );

      // Auto-assign connections for execution steps.
      if (metadata?.executorConnections) {
        autoAssignConnections(newNodes, metadata.executorConnections);
      }

      setNodes(() => newNodes);
      setEdges(() => newEdges);

      onResourceDropOnCanvas(
        defaultPropertySelector ?? sourceResource,
        defaultPropertySectorStepId ?? targetStepId ?? '',
      );

      return;
    }

    if (sourceResource && targetStepId) {
      const generatedElement: Element = generateStepElement(sourceResource);

      updateNodeData(targetStepId, (node: Node) => {
        const nodeData = node?.data as StepData | undefined;
        const updatedComponents: Element[] = move([...(cloneDeep(nodeData?.components) ?? [])], event);

        return {
          components: mutateComponents([...updatedComponents, generatedElement]),
        };
      });

      // Update node internals to fix handle positions after adding element
      setTimeout(() => {
        updateNodeInternals(targetStepId);
      }, 50);

      onResourceDropOnCanvas(generatedElement, targetStepId);
    }
  };

  const addToForm = (
    event: Parameters<DragDropEventHandlers['onDragEnd']>[0],
    sourceData: DragSourceData,
    targetData: DragTargetData,
  ): void => {
    const {dragged: sourceResource} = sourceData;
    const {stepId: targetStepId, droppedOn: targetResource} = targetData;

    if (sourceResource && targetStepId && targetResource) {
      const generatedElement: Element = generateStepElement(sourceResource);

      updateNodeData(targetStepId, (node: Node) => {
        const nodeData = node?.data as StepData | undefined;
        const updatedComponents: Element[] =
          cloneDeep(nodeData?.components)?.map((component: Element) =>
            component.id === targetResource.id
              ? {
                  ...component,
                  components: move([...(component.components ?? [])], event).concat(generatedElement),
                }
              : component,
          ) ?? [];

        return {
          components: mutateComponents(updatedComponents),
        };
      });

      // Update node internals to fix handle positions after adding element
      setTimeout(() => {
        updateNodeInternals(targetStepId);
      }, 50);

      onResourceDropOnCanvas(generatedElement, targetStepId);
    }
  };

  /**
   * Adds a new element to a view at a specific index (when dropping on an existing element).
   */
  const addToViewAtIndex = (sourceData: DragSourceData, targetStepId: string, targetElementId: string): void => {
    const {dragged: sourceResource} = sourceData;

    if (sourceResource && targetStepId) {
      // Check if this is a widget drop - widgets need special handling
      if (sourceResource.resourceType === ResourceTypes.Widget && nodes && edges) {
        const targetNode = nodes.find((n) => n.id === targetStepId);

        if (targetNode) {
          const [newNodes, newEdges, defaultPropertySelector, defaultPropertySectorStepId] = onWidgetLoad(
            sourceResource as Widget,
            targetNode as Resource,
            nodes,
            edges,
          );

          // Now we need to reorder the components to insert at the correct position
          const updatedNodes = newNodes.map((node) => {
            if (node.id === targetStepId) {
              const nodeData = node.data as StepData | undefined;
              const components: Element[] = cloneDeep(nodeData?.components) ?? [];

              // The widget button was appended to the end, find it (it's the last element)
              const widgetButton = components[components.length - 1];

              // Remove it from the end
              components.pop();

              // Find the target index and insert there
              const targetIndex = components.findIndex((c) => c.id === targetElementId);

              if (targetIndex !== -1) {
                // Insert at the target index (before the target element)
                components.splice(targetIndex, 0, widgetButton);
              } else {
                // Fallback: append to end if target not found
                components.push(widgetButton);
              }

              return {
                ...node,
                data: {
                  ...nodeData,
                  components: mutateComponents(components),
                },
              };
            }
            return node;
          });

          // Auto-assign connections for execution steps.
          if (metadata?.executorConnections) {
            autoAssignConnections(updatedNodes, metadata.executorConnections);
          }

          setNodes(() => updatedNodes);
          setEdges(() => newEdges);

          onResourceDropOnCanvas(
            defaultPropertySelector ?? sourceResource,
            defaultPropertySectorStepId ?? targetStepId,
          );
        }
      } else {
        // Regular element drop
        const generatedElement: Element = generateStepElement(sourceResource);

        updateNodeData(targetStepId, (node: Node) => {
          const nodeData = node?.data as StepData | undefined;
          const components: Element[] = cloneDeep(nodeData?.components) ?? [];

          // Find the index of the target element
          const targetIndex = components.findIndex((c) => c.id === targetElementId);

          if (targetIndex !== -1) {
            // Insert at the target index (before the target element)
            components.splice(targetIndex, 0, generatedElement);
          } else {
            // Fallback: append to end if target not found
            components.push(generatedElement);
          }

          return {
            components: mutateComponents(components),
          };
        });

        // Update node internals to fix handle positions after adding element
        setTimeout(() => {
          updateNodeInternals(targetStepId);
        }, 50);

        onResourceDropOnCanvas(generatedElement, targetStepId);
      }
    }
  };

  /**
   * Adds a new element to a form at a specific index (when dropping on an existing element inside a form).
   */
  const addToFormAtIndex = (
    sourceData: DragSourceData,
    targetStepId: string,
    formId: string,
    targetElementId: string,
  ): void => {
    const {dragged: sourceResource} = sourceData;

    if (sourceResource && targetStepId) {
      const generatedElement: Element = generateStepElement(sourceResource);

      updateNodeData(targetStepId, (node: Node) => {
        const nodeData = node?.data as StepData | undefined;
        const components: Element[] =
          cloneDeep(nodeData?.components)?.map((component: Element) => {
            if (component.id === formId && component.components) {
              const formComponents = [...component.components];
              const targetIndex = formComponents.findIndex((c) => c.id === targetElementId);

              if (targetIndex !== -1) {
                // Insert at the target index (before the target element)
                formComponents.splice(targetIndex, 0, generatedElement);
              } else {
                // Fallback: append to end if target not found
                formComponents.push(generatedElement);
              }

              return {
                ...component,
                components: formComponents,
              };
            }

            return component;
          }) ?? [];

        return {
          components: mutateComponents(components),
        };
      });

      // Update node internals to fix handle positions after adding element
      setTimeout(() => {
        updateNodeInternals(targetStepId);
      }, 50);

      onResourceDropOnCanvas(generatedElement, targetStepId);
    }
  };

  const handleDragEnd: DragDropEventHandlers['onDragEnd'] = (event): void => {
    const {source, target} = event.operation;

    if (!source) {
      return;
    }

    const sourceData: DragSourceData = source.data as DragSourceData;
    const targetData: DragTargetData = (target?.data as DragTargetData) ?? {};

    // Check for components that need containers (Form needs View, Input needs Form+View, Widget needs View)
    const isFormDrop = sourceData.dragged?.type === BlockTypes.Form;
    const isInputDrop = sourceData.dragged?.type === ElementTypes.Input;
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

    // For other canceled events or missing target, return early
    if (event.canceled || !target) {
      return;
    }

    if (sourceData.isReordering) {
      if (!sourceData.stepId) {
        return;
      }

      updateNodeData(sourceData.stepId, (node: Node) => {
        const unorderedComponents: Element[] = cloneDeep((node?.data as StepData)?.components ?? []);

        const reorderedNested = unorderedComponents.map((component: Element) => {
          if (component?.components) {
            return {
              ...component,
              components: move(component.components, event),
            };
          }

          return component;
        });

        return {
          components: move(reorderedNested, event),
        };
      });

      // Update node internals to fix handle positions after reordering
      // Use setTimeout with a small delay to ensure React has committed the state changes and DOM has updated
      setTimeout(() => {
        updateNodeInternals(sourceData.stepId!);
      }, 50);
    } else if (typeof target?.id === 'string' && target.id.startsWith(VisualFlowConstants.FLOW_BUILDER_CANVAS_ID)) {
      addCanvasNode(event, sourceData, targetData);
    } else if (typeof target?.id === 'string' && target.id.startsWith(VisualFlowConstants.FLOW_BUILDER_VIEW_ID)) {
      addToView(event, sourceData, targetData);
    } else if (typeof target?.id === 'string' && target.id.startsWith(VisualFlowConstants.FLOW_BUILDER_FORM_ID)) {
      addToForm(event, sourceData, targetData);
    } else if (targetData.isReordering && targetData.stepId && typeof target?.id === 'string') {
      // Dropping on an existing sortable element - insert at that position
      const targetElementId = target.id;

      // Check if the target element is inside a form by searching through all forms
      const nodeData = nodes?.find((n) => n.id === targetData.stepId)?.data as StepData | undefined;
      const parentForm = nodeData?.components?.find(
        (c) => c.type === BlockTypes.Form && c.components?.some((child) => child.id === targetElementId),
      );

      if (parentForm) {
        // Target element is inside a form, insert at that position within the form
        addToFormAtIndex(sourceData, targetData.stepId, parentForm.id, targetElementId);
      } else {
        // Target is a top-level element in the view, add to view at index
        addToViewAtIndex(sourceData, targetData.stepId, targetElementId);
      }
    }
  };

  const handleDragOver: DragDropEventHandlers['onDragOver'] = useCallback(
    (event) => {
      const {source, target} = event.operation;

      if (!source || !target) {
        return;
      }

      // If not a reordering operation, return.
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
        const unorderedComponents: Element[] = cloneDeep(nodeData?.components) ?? [];

        const reorderedNested = unorderedComponents.map((component: Element) => {
          if (component?.components) {
            return {
              ...component,
              components: move(component.components, event),
            };
          }

          return component;
        });

        return {
          components: move(reorderedNested, event),
        };
      });

      // Update node internals to fix handle positions after reordering
      // Use setTimeout to ensure React has committed the state changes and DOM has updated
      setTimeout(() => {
        updateNodeInternals(stepId);
      }, 50);
    },
    [updateNodeData, updateNodeInternals],
  );

  const onConnect: OnConnect = useCallback(
    (connection: Connection) => {
      let edge: Edge | null = onEdgeResolve && nodes ? onEdgeResolve(connection, nodes) : null;

      edge ??= {
        ...connection,
        id: `${connection.source}-${connection.target}`,
        markerEnd: {
          type: MarkerType.Arrow,
        },
        type: 'base-edge',
      };

      setEdges((prevEdges: Edge[]) => addEdge(edge, prevEdges));
    },
    [onEdgeResolve, nodes, setEdges],
  );

  const onNodesDelete: OnNodesDelete<Node> = useCallback(
    (deleted: Node[]) => {
      // Execute plugins for ON_NODE_DELETE event asynchronously (fire and forget)
      // eslint-disable-next-line @typescript-eslint/no-floating-promises
      PluginRegistry.getInstance().executeAsync(FlowEventTypes.ON_NODE_DELETE, deleted);

      if (!nodes || !edges) {
        return;
      }

      const currentEdges: Edge[] = edges;
      const currentNodes: Node[] = nodes;

      setEdges(
        deleted.reduce((acc: Edge[], node: Node) => {
          const incomers: Node[] = getIncomers(node, currentNodes, currentEdges);
          const outgoers: Node[] = getOutgoers(node, currentNodes, currentEdges);
          const connectedEdges: Edge[] = getConnectedEdges([node], currentEdges);

          const remainingEdges: Edge[] = acc.filter((edge: Edge) => !connectedEdges.includes(edge));

          const createdEdges: Edge[] = incomers.flatMap(({id: source}: {id: string}) =>
            outgoers
              .map(({id: target}: {id: string}) => {
                // Find the edge from incomer to the node being deleted
                const edge: Edge | undefined = connectedEdges.find(
                  (e: Edge) => e.source === source && e.target === node.id,
                );

                if (!edge) {
                  return null;
                }

                return {
                  id: `${edge.source}->${target}`,
                  source,
                  sourceHandle: edge?.sourceHandle,
                  target,
                  type: edge?.type,
                } as Edge;
              })
              .filter((edge: Edge | null) => edge !== null),
          );

          return [...remainingEdges, ...createdEdges];
        }, currentEdges),
      );
    },
    [setEdges, edges, nodes],
  );

  /**
   * Handles the deletion of edges.
   *
   * @param deleted - Array of deleted edges.
   */
  const onEdgesDelete: (deleted: Edge[]) => void = useCallback((deleted: Edge[]) => {
    // Execute plugins for ON_EDGE_DELETE event asynchronously (fire and forget)
    // eslint-disable-next-line @typescript-eslint/no-floating-promises
    PluginRegistry.getInstance().executeAsync(FlowEventTypes.ON_EDGE_DELETE, deleted);
  }, []);

  const handleOnAdd = (resource: Resource): void => {
    const clonedResource: Resource = cloneDeep(resource);

    // Handle templates
    if (resource.resourceType === ResourceTypes.Template) {
      const template = clonedResource as Template;

      /**
       * Execute plugins for ON_TEMPLATE_LOAD event.
       */
      PluginRegistry.getInstance().executeSync(FlowEventTypes.ON_TEMPLATE_LOAD, template);

      const [newNodes, newEdges, defaultPropertySelector, defaultPropertySelectorStepId] = onTemplateLoad(template);

      // Auto-assign connections for execution steps.
      if (metadata?.executorConnections) {
        autoAssignConnections(newNodes, metadata.executorConnections);
      }

      // Helper to update node internals after render
      const updateAllNodeInternals = (nodesToUpdate: Node[]): void => {
        nodesToUpdate.forEach((node: Node) => {
          updateNodeInternals(node.id);

          if (node.data?.components) {
            (node.data.components as Element[]).forEach((component: Element) => {
              updateNodeInternals(component.id);

              if (component?.components) {
                component.components.forEach((nestedComponent: Element) => {
                  updateNodeInternals(nestedComponent.id);
                });
              }
            });
          }
        });
      };

      // Apply auto-layout BEFORE setting nodes to avoid visual jump
      // Use estimated dimensions since nodes aren't rendered yet
      applyAutoLayout(newNodes, newEdges, {
        direction: 'RIGHT',
        nodeSpacing: 150,
        rankSpacing: 300,
        offsetX: 50,
        offsetY: 50,
      })
        .then((layoutedNodes) => {
          // Set nodes and edges immediately with layouted positions
          setNodes(layoutedNodes);
          setEdges([...newEdges]);

          // Update node internals after React has rendered using requestAnimationFrame
          // This ensures the DOM is ready before we update internals
          requestAnimationFrame(() => {
            updateAllNodeInternals(layoutedNodes);
            // Fit view after nodes are rendered
            requestAnimationFrame(() => {
              void fitView({padding: 0.2, duration: 300});
            });
          });
        })
        .catch(() => {
          // Layout failed, use original positions
          setNodes(newNodes);
          setEdges([...newEdges]);

          requestAnimationFrame(() => {
            updateAllNodeInternals(newNodes);
          });
        });

      onResourceDropOnCanvas(defaultPropertySelector ?? resource, defaultPropertySelectorStepId ?? '');

      return;
    }

    // Handle widgets - add to existing View or create new View
    if (resource.resourceType === ResourceTypes.Widget) {
      // Try to find an existing View step
      const existingViewStep = nodes.find((node) => node.type === StepTypes.View);

      let targetViewStep: Step;
      let currentNodes: Node[];

      if (existingViewStep) {
        const nodeData = existingViewStep.data as StepData | undefined;
        const existingComponents: Element[] = cloneDeep(nodeData?.components ?? []);

        // Check if the widget being added will create a Form
        // We need to check the widget's data to see if it contains Form components
        const widget = clonedResource as Widget;
        const widgetData = widget.data as {components?: Element[]} | undefined;

        // Check if widget will add a Form (either directly or through onWidgetLoad)
        const widgetWillAddForm = Boolean(
          widgetData?.components?.some((component: Element) => component.type === BlockTypes.Form),
        );

        // Only remove existing Form if the new widget will add a Form
        // This ensures only one Form per View while allowing non-Form widgets to coexist
        let componentsForWidget = existingComponents;
        let oldForm: Element | undefined;

        if (widgetWillAddForm) {
          // Widget will add a Form - remove existing Form to maintain single Form rule
          oldForm = existingComponents.find((component: Element) => component.type === BlockTypes.Form);
          componentsForWidget = existingComponents.filter((component: Element) => component.type !== BlockTypes.Form);
        }

        targetViewStep = {
          ...existingViewStep,
          data: {
            ...nodeData,
            components: componentsForWidget,
          },
        } as Step;

        // Update the nodes list with the modified View
        currentNodes = nodes.map((node) => (node.id === existingViewStep.id ? targetViewStep : node));

        // Store information about edges that should be preserved/recreated
        const edgeTargetsToPreserve = new Map<string, string>();
        let updatedEdges = edges;

        if (oldForm?.components) {
          const oldFormElementIds = new Set<string>();

          // Collect all element IDs from the old Form (including nested elements)
          const collectElementIds = (elements: Element[]): void => {
            elements.forEach((element: Element) => {
              oldFormElementIds.add(element.id);
              if (element.components) {
                collectElementIds(element.components);
              }
            });
          };

          collectElementIds(oldForm.components);

          // Before filtering, save information about outgoing edges from Form elements
          // These typically represent flow connections (like button -> next step)
          // Note: edges from View children have source=viewId and sourceHandle contains the element ID
          edges.forEach((edge: Edge) => {
            // Extract element ID from sourceHandle (format: "elementId_NEXT")
            const {sourceHandle, target} = edge;
            if (sourceHandle) {
              const elementIdFromHandle = sourceHandle.replace(/_NEXT$/, '');

              if (oldFormElementIds.has(elementIdFromHandle)) {
                // This is an edge from an old Form element
                // Save what it was connecting to if the target is outside the Form
                if (!oldFormElementIds.has(target)) {
                  edgeTargetsToPreserve.set(elementIdFromHandle, target);
                }
              }
            }
          });

          // Filter out edges connected to old Form elements
          updatedEdges = edges.filter((edge: Edge) => {
            // Check if the edge's sourceHandle contains an old Form element ID
            const {sourceHandle, target} = edge;
            if (sourceHandle) {
              const elementIdFromHandle = sourceHandle.replace(/_NEXT$/, '');
              if (oldFormElementIds.has(elementIdFromHandle)) {
                return false; // Filter out this edge
              }
            }
            // Also check if target is an old Form element
            return !oldFormElementIds.has(target);
          });
        }

        // Use onWidgetLoad to properly load the widget into the View
        const [loadedNodes, newEdges, defaultPropertySelector, defaultPropertySelectorStepId] = onWidgetLoad(
          clonedResource as Widget,
          targetViewStep,
          currentNodes,
          updatedEdges,
        );

        // CRITICAL: Ensure only one Form exists after onWidgetLoad
        // onWidgetLoad might have added components that include Forms
        // We need to enforce the single Form rule
        const newNodes = loadedNodes.map((node) => {
          if (node.id === existingViewStep.id) {
            const viewNodeData = node.data as StepData | undefined;
            const components = viewNodeData?.components ?? [];

            // Find all Forms in the components
            const forms = components.filter((c: Element) => c.type === BlockTypes.Form);

            if (forms.length > 1) {
              // Multiple Forms detected - keep only the last one (newest)
              const newestForm = forms[forms.length - 1];
              const componentsWithoutForms = components.filter((c: Element) => c.type !== BlockTypes.Form);

              return {
                ...node,
                data: {
                  ...viewNodeData,
                  components: [...componentsWithoutForms, newestForm],
                },
              };
            }
          }
          return node;
        });

        // Auto-assign connections for execution steps.
        if (metadata?.executorConnections) {
          autoAssignConnections(newNodes, metadata.executorConnections);
        }

        setNodes(() => newNodes);
        setEdges(() => newEdges);

        onResourceDropOnCanvas(
          defaultPropertySelector ?? clonedResource,
          defaultPropertySelectorStepId ?? targetViewStep.id ?? '',
        );

        return;
      }

      // Create a new View step at a default position
      const position: XYPosition = screenToFlowPosition({
        x: window.innerWidth / 2,
        y: window.innerHeight / 2,
      });

      targetViewStep = {
        category: ResourceTypes.Step,
        data: {
          components: [],
        },
        deletable: true,
        id: generateResourceId(StepTypes.View.toLowerCase()),
        position,
        resourceType: ResourceTypes.Step,
        type: StepTypes.View,
      } as Step;

      currentNodes = [...nodes, targetViewStep];

      // Use onWidgetLoad to properly load the widget into the View
      const [newNodes, newEdges, defaultPropertySelector, defaultPropertySelectorStepId] = onWidgetLoad(
        clonedResource as Widget,
        targetViewStep,
        currentNodes,
        edges,
      );

      // Auto-assign connections for execution steps.
      if (metadata?.executorConnections) {
        autoAssignConnections(newNodes, metadata.executorConnections);
      }

      setNodes(() => newNodes);
      setEdges(() => newEdges);

      onResourceDropOnCanvas(
        defaultPropertySelector ?? clonedResource,
        defaultPropertySelectorStepId ?? targetViewStep.id ?? '',
      );

      return;
    }

    // Handle steps
    if (resource.resourceType === ResourceTypes.Step) {
      const position: XYPosition = screenToFlowPosition({
        x: window.innerWidth / 2,
        y: window.innerHeight / 2,
      });

      let generatedStep: Step = {
        ...clonedResource,
        data: {
          components: [],
          ...(clonedResource?.data ?? {}),
        },
        deletable: true,
        id: generateResourceId(clonedResource.type.toLowerCase()),
        position,
      } as Step;

      // Decorate the step with any additional information
      generatedStep = onStepLoad(generatedStep);

      setNodes((prevNodes: Node[]) => [...prevNodes, generatedStep]);

      onResourceDropOnCanvas(generatedStep, '');
    }

    // Handle elements (components)
    if (resource.resourceType === ResourceTypes.Element) {
      const element = clonedResource as Element;

      // Special handling for Forms
      if (element.type === BlockTypes.Form) {
        // Try to find an existing View step
        const existingViewStep = nodes.find((node) => node.type === StepTypes.View);

        if (existingViewStep) {
          // Add/replace Form in existing View
          const generatedElement: Element = generateStepElement(element);

          updateNodeData(existingViewStep.id, (node: Node) => {
            const nodeData = node?.data as StepData | undefined;
            const existingComponents: Element[] = cloneDeep(nodeData?.components ?? []);

            // Remove any existing Form
            const componentsWithoutForm = existingComponents.filter(
              (component: Element) => component.type !== BlockTypes.Form,
            );

            // Add the new Form at the beginning
            return {
              components: mutateComponents([generatedElement, ...componentsWithoutForm]),
            };
          });

          // Update node internals to fix handle positions after adding element
          setTimeout(() => {
            updateNodeInternals(existingViewStep.id);
          }, 50);

          onResourceDropOnCanvas(generatedElement, existingViewStep.id);
        } else {
          // Create a new View step with the Form
          const position: XYPosition = screenToFlowPosition({
            x: window.innerWidth / 2,
            y: window.innerHeight / 2,
          });

          const generatedElement: Element = generateStepElement(element);

          let generatedViewStep: Step = {
            category: ResourceTypes.Step,
            data: {
              components: [generatedElement],
            },
            deletable: true,
            id: generateResourceId(StepTypes.View.toLowerCase()),
            position,
            resourceType: ResourceTypes.Step,
            type: StepTypes.View,
          } as Step;

          generatedViewStep = onStepLoad(generatedViewStep);
          setNodes((prevNodes: Node[]) => [...prevNodes, generatedViewStep]);
          onResourceDropOnCanvas(generatedViewStep, '');
        }

        return;
      }

      // For INPUT elements, add them to Form (create Form if needed)
      if (element.type === ElementTypes.Input) {
        const existingViewStep = nodes.find((node) => node.type === StepTypes.View);

        if (existingViewStep) {
          const generatedElement: Element = generateStepElement(element);

          updateNodeData(existingViewStep.id, (node: Node) => {
            const nodeData = node?.data as StepData | undefined;
            const existingComponents: Element[] = cloneDeep(nodeData?.components ?? []);

            // Find existing Form in the View
            const existingForm = existingComponents.find((comp: Element) => comp.type === BlockTypes.Form);

            if (existingForm) {
              // Add input to existing Form
              const updatedForm: Element = {
                ...existingForm,
                components: [...(existingForm.components ?? []), generatedElement],
              };

              const componentsWithoutForm = existingComponents.filter((comp: Element) => comp.type !== BlockTypes.Form);

              return {
                components: mutateComponents([...componentsWithoutForm, updatedForm]),
              };
            }

            // No Form exists - create a new Form with the input
            const newForm: Element = {
              ...generateStepElement({
                resourceType: ResourceTypes.Element,
                category: ElementCategories.Block,
                type: BlockTypes.Form,
                config: {},
              } as Element),
              components: [generatedElement],
            };

            return {
              components: mutateComponents([...existingComponents, newForm]),
            };
          });

          // Update node internals to fix handle positions after adding element
          setTimeout(() => {
            updateNodeInternals(existingViewStep.id);
          }, 50);

          onResourceDropOnCanvas(generatedElement, existingViewStep.id);
        } else {
          // No View exists - create View with Form containing the input
          const position: XYPosition = screenToFlowPosition({
            x: window.innerWidth / 2,
            y: window.innerHeight / 2,
          });

          const generatedInput: Element = generateStepElement(element);

          const newForm: Element = {
            ...generateStepElement({
              resourceType: ResourceTypes.Element,
              category: ElementCategories.Block,
              type: BlockTypes.Form,
              config: {},
            } as Element),
            components: [generatedInput],
          };

          let generatedViewStep: Step = {
            category: ResourceTypes.Step,
            data: {
              components: [newForm],
            },
            deletable: true,
            id: generateResourceId(StepTypes.View.toLowerCase()),
            position,
            resourceType: ResourceTypes.Step,
            type: StepTypes.View,
          } as Step;

          generatedViewStep = onStepLoad(generatedViewStep);
          setNodes((prevNodes: Node[]) => [...prevNodes, generatedViewStep]);
          onResourceDropOnCanvas(generatedViewStep, '');
        }

        return;
      }

      // For other elements (Buttons, etc.), try to add to existing View
      const existingViewStep = nodes.find((node) => node.type === StepTypes.View);

      if (existingViewStep) {
        const generatedElement: Element = generateStepElement(element);

        updateNodeData(existingViewStep.id, (node: Node) => {
          const nodeData = node?.data as StepData | undefined;
          const existingComponents: Element[] = cloneDeep(nodeData?.components ?? []);

          return {
            components: mutateComponents([...existingComponents, generatedElement]),
          };
        });

        // Update node internals to fix handle positions after adding element
        setTimeout(() => {
          updateNodeInternals(existingViewStep.id);
        }, 50);

        onResourceDropOnCanvas(generatedElement, existingViewStep.id);
      }
      // If no View exists and it's not a Form or Input, do nothing (elements can't be added to canvas directly)
    }
  };

  /**
   * Handles closing the container requirements dialog.
   */
  const handleContainerDialogClose = (): void => {
    setIsContainerDialogOpen(false);
    pendingDropRef.current = null;
  };

  /**
   * Handles confirming the container requirements dialog.
   * Creates the necessary container hierarchy based on the drop scenario.
   */
  const handleContainerDialogConfirm = (): void => {
    const pendingData = pendingDropRef.current;

    if (!pendingData) {
      handleContainerDialogClose();
      return;
    }

    const {event, sourceData, targetData} = pendingData;
    const droppedResource: Resource | undefined = cloneDeep(sourceData.dragged);

    if (!droppedResource || !event.nativeEvent) {
      handleContainerDialogClose();
      return;
    }

    // Type guard to ensure nativeEvent is a MouseEvent
    const {nativeEvent} = event;
    if (!('clientX' in nativeEvent) || !('clientY' in nativeEvent)) {
      handleContainerDialogClose();
      return;
    }

    const {clientX, clientY} = nativeEvent as MouseEvent;

    const position: XYPosition = screenToFlowPosition({
      x: clientX,
      y: clientY,
    });

    // Generate the dropped element with a unique ID
    const generatedElement: Element = generateStepElement(droppedResource as Element);

    if (dropScenario === 'form-on-canvas') {
      // Create a View step with the Form inside
      let generatedViewStep: Step = {
        category: ResourceTypes.Step,
        data: {
          components: [generatedElement],
        },
        deletable: true,
        id: generateResourceId(StepTypes.View.toLowerCase()),
        position,
        resourceType: ResourceTypes.Step,
        type: StepTypes.View,
      } as Step;

      generatedViewStep = onStepLoad(generatedViewStep);
      setNodes((prevNodes: Node[]) => [...prevNodes, generatedViewStep]);
      onResourceDropOnCanvas(generatedViewStep, '');
    } else if (dropScenario === 'input-on-canvas') {
      // Create a Form element containing the Input
      const formElement: Element = {
        resourceType: ResourceTypes.Element,
        category: ElementCategories.Block,
        type: BlockTypes.Form,
        id: generateResourceId(BlockTypes.Form.toLowerCase()),
        config: {},
        components: [generatedElement],
      } as Element;

      // Create a View step with the Form (containing the Input) inside
      let generatedViewStep: Step = {
        category: ResourceTypes.Step,
        data: {
          components: [formElement],
        },
        deletable: true,
        id: generateResourceId(StepTypes.View.toLowerCase()),
        position,
        resourceType: ResourceTypes.Step,
        type: StepTypes.View,
      } as Step;

      generatedViewStep = onStepLoad(generatedViewStep);
      setNodes((prevNodes: Node[]) => [...prevNodes, generatedViewStep]);
      onResourceDropOnCanvas(generatedViewStep, '');
    } else if (dropScenario === 'input-on-view') {
      // Create a Form element containing the Input and add it to the View
      const formElement: Element = {
        resourceType: ResourceTypes.Element,
        category: ElementCategories.Block,
        type: BlockTypes.Form,
        id: generateResourceId(BlockTypes.Form.toLowerCase()),
        config: {},
        components: [generatedElement],
      } as Element;

      const targetStepId = targetData.stepId;
      if (targetStepId) {
        updateNodeData(targetStepId, (node: Node) => {
          const existingComponents: Element[] = cloneDeep((node?.data as StepData)?.components ?? []);
          return {
            components: [...existingComponents, formElement],
          };
        });
        onResourceDropOnCanvas(formElement, targetStepId);
      }
    } else if (dropScenario === 'widget-on-canvas') {
      // Create an empty View step - do NOT call onStepLoad here as it would add default components
      // The onWidgetLoad function will handle populating the View with widget-specific content
      const generatedViewStep: Step = {
        category: ResourceTypes.Step,
        data: {
          components: [],
        },
        deletable: true,
        id: generateResourceId(StepTypes.View.toLowerCase()),
        position,
        resourceType: ResourceTypes.Step,
        type: StepTypes.View,
      } as Step;

      // Use onWidgetLoad to properly load the widget into the View
      const [newNodes, newEdges, defaultPropertySelector, defaultPropertySectorStepId] = onWidgetLoad(
        droppedResource as Widget,
        generatedViewStep,
        [...nodes, generatedViewStep],
        edges,
      );

      // Auto-assign connections for execution steps.
      if (metadata?.executorConnections) {
        autoAssignConnections(newNodes, metadata.executorConnections);
      }

      setNodes(() => newNodes);
      setEdges(() => newEdges);

      onResourceDropOnCanvas(
        defaultPropertySelector ?? droppedResource,
        defaultPropertySectorStepId ?? generatedViewStep.id ?? '',
      );
    }

    // Close the dialog and clear pending data
    handleContainerDialogClose();
  };

  const handleSave = (): void => {
    // Get viewport from toObject but use full nodes/edges (not filtered)
    // This ensures we always save the complete graph including execution nodes
    const {viewport} = toObject();
    const canvasData = {
      nodes,
      edges,
      viewport,
    };
    onSave?.(canvasData);
  };

  const handleAutoLayout = useCallback((): void => {
    applyAutoLayout(nodes, edges, {
      direction: 'RIGHT',
      nodeSpacing: 150, // Vertical spacing between nodes in the same layer
      rankSpacing: 300, // Horizontal spacing between layers
      offsetX: 50, // Horizontal offset from the left
      offsetY: 50, // Vertical offset from the top
    })
      .then((layoutedNodes) => {
        setNodes(layoutedNodes);
        // Fit view after layout with a small delay to ensure nodes are rendered
        setTimeout(() => {
          void fitView({padding: 0.2, duration: 300});
        }, 50);
      })
      .catch(() => {
        // Layout failed, keep original positions
      });
  }, [nodes, edges, setNodes, fitView]);

  /**
   * Handles node drag stop event to resolve collisions between nodes.
   * When a node is dragged and released, this function checks for overlapping nodes
   * and pushes them apart to prevent visual collisions.
   */
  const handleNodeDragStop = useCallback((): void => {
    const resolvedNodes = resolveCollisions(nodes, {
      maxIterations: 50,
      overlapThreshold: 0.5,
      margin: 20, // Add margin around nodes to prevent them from being too close
    });

    // Only update if positions actually changed
    const hasChanges = resolvedNodes.some(
      (resolvedNode, index) =>
        resolvedNode.position.x !== nodes[index].position.x || resolvedNode.position.y !== nodes[index].position.y,
    );

    if (hasChanges) {
      setNodes(resolvedNodes);
    }
  }, [nodes, setNodes]);

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
      {/* Header spanning full width at top */}
      <HeaderPanel onSave={handleSave} onAutoLayout={handleAutoLayout} />

      {/* Main content area: ResourcePanel (left) + Canvas (right) */}
      <Box sx={{flexGrow: 1, minHeight: 0}}>
        <DragDropProvider onDragEnd={handleDragEnd} onDragOver={handleDragOver}>
          <ResourcePanel
            resources={resources}
            open={isResourcePanelOpen}
            onAdd={handleOnAdd}
            disabled={isFlowMetadataLoading}
          >
            <ResourcePropertyPanel open={isResourcePropertiesPanelOpen} onComponentDelete={deleteComponent}>
              {/* <VersionHistoryPanel open={isVersionHistoryPanelOpen}> */}
              <VisualFlow
                nodes={nodes}
                onNodesChange={onNodesChange}
                edges={edges}
                onEdgesChange={onEdgesChange}
                onConnect={onConnect}
                onNodesDelete={onNodesDelete}
                onEdgesDelete={onEdgesDelete}
                onNodeDragStop={handleNodeDragStop}
                {...rest}
              />
              {/* </VersionHistoryPanel> */}
              {/* <ValidationPanel /> */}
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
