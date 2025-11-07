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

import {Box} from '@wso2/oxygen-ui';
import type {Edge, Node, NodeTypes, NodeProps} from '@xyflow/react';
import {MarkerType, useEdgesState, useNodesState, useUpdateNodeInternals} from '@xyflow/react';
import {useCallback, useEffect, useLayoutEffect, useMemo, useRef} from 'react';
import '@xyflow/react/dist/style.css';
import cloneDeep from 'lodash-es/cloneDeep';
import isEmpty from 'lodash-es/isEmpty';
import mergeWith from 'lodash-es/mergeWith';
import type {UpdateNodeInternals} from '@xyflow/system';
import FlowBuilder from '@/features/flows/components/FlowBuilder';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import {
  BlockTypes,
  ButtonTypes,
  ButtonVariants,
  ElementCategories,
  ElementTypes,
  InputVariants,
  type Element,
} from '@/features/flows/models/elements';
import {StaticStepTypes, StepTypes, type Step, type StepData} from '@/features/flows/models/steps';
import {type Template, TemplateTypes, type TemplateReplacer} from '@/features/flows/models/templates';
import type {Widget} from '@/features/flows/models/widget';
import generateIdsForResources from '@/features/flows/utils/generateIdsForResources';
import resolveComponentMetadata from '@/features/flows/utils/resolveComponentMetadata';
import resolveStepMetadata from '@/features/flows/utils/resolveStepMetadata';
import updateTemplatePlaceholderReferences from '@/features/flows/utils/updateTemplatePlaceholderReferences';
import {ResourceTypes, type Resource as FlowResource, type Resource} from '@/features/flows/models/resources';
import useFlowBuilderCore from '@/features/flows/hooks/useFlowBuilderCore';
import useGenerateStepElement from '@/features/flows/hooks/useGenerateStepElement';
import StaticStepFactory from './resources/steps/StaticStepFactory';
import StepFactory from './resources/steps/StepFactory';
import useGetLoginFlowBuilderResources from '../api/useGetLoginFlowBuilderResources';

function LoginFlowBuilder() {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const INITIAL_FLOW_START_STEP_ID: string = StaticStepTypes.Start.toLowerCase();
  const INITIAL_FLOW_USER_ONBOARD_STEP_ID: string = StepTypes.End;

  const {data: resources} = useGetLoginFlowBuilderResources();
  const {setFlowCompletionConfigs} = useFlowBuilderCore();
  const {generateStepElement} = useGenerateStepElement();

  const {steps} = resources;

  const updateNodeInternals: UpdateNodeInternals = useUpdateNodeInternals();
  const flowUpdatesInProgress = useRef<boolean>(false);
  const nodesUpdatedRef = useRef<boolean>(false);

  /**
   * Determines if a step is deletable based on its type and executor.
   * @param step - The step to check.
   * @returns true if the step is deletable, false otherwise.
   */
  const isStepDeletable = (step: Node): boolean => {
    let isDeletable = true;

    if (step.type === StepTypes.Execution) {
      isDeletable = false;
    }

    return isDeletable;
  };

  /**
   * Validate edges based on the nodes.
   * @param validatingEdges - Edges to validate.
   * @param _nodes - Nodes for validation (currently unused).
   */
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const validateEdges = (validatingEdges: Edge[], _nodes: Node[]): Edge[] => validatingEdges.slice();

  const getBlankTemplateComponents = (): Element[] => {
    const blankTemplate: Template | undefined = cloneDeep(
      resources?.templates?.find((template: Template) => template.type === TemplateTypes.Blank),
    );

    if (!blankTemplate) {
      return [];
    }

    if (blankTemplate.config && blankTemplate.config.data?.steps?.length > 0) {
      blankTemplate.config.data.steps[0].id = INITIAL_FLOW_START_STEP_ID;
    }

    return resolveComponentMetadata(
      resources,
      generateIdsForResources<Template>(blankTemplate)?.config?.data?.steps[0]?.data?.components,
    );
  };

  /**
   * Handles adding a resource (like Form) from the resource panel via the + icon.
   * This finds or creates a View step and adds the element to it.
   */
  const handleResourceAdd = (resource: FlowResource): void => {
    if (resource.resourceType !== ResourceTypes.Element) {
      return;
    }

    const element = resource as Element;
    const generatedElement: Element = generateStepElement(element);

    // Try to find an existing View step
    const existingViewStep = nodes.find((node) => node.type === StepTypes.View);

    if (existingViewStep) {
      // Add to existing View
      setNodes((prevNodes: Node[]) =>
        prevNodes.map((node) => {
          if (node.id === existingViewStep.id) {
            const nodeData = node.data as StepData | undefined;
            const existingComponents: Element[] = nodeData?.components ?? [];

            // For Forms, replace any existing Form (only one Form per View)
            let updatedComponents: Element[];
            if (generatedElement.type === BlockTypes.Form) {
              updatedComponents = [
                ...existingComponents.filter((comp: Element) => comp.type !== BlockTypes.Form),
                generatedElement,
              ];
            } else {
              updatedComponents = [...existingComponents, generatedElement];
            }

            const mutatedComponents: Element[] | undefined = handleMutateComponents(updatedComponents);

            return {
              ...node,
              data: {
                ...nodeData,
                components: mutatedComponents,
              },
            };
          }

          return node;
        }),
      );

      // Update node internals for rendering
      setTimeout(() => {
        updateNodeInternals(existingViewStep.id);
        updateNodeInternals(generatedElement.id);
      }, 100);
    }
    // If no View exists, the user should add a View step first
    // We could optionally create a View automatically here if needed
  };

  const handleStepLoad = (step: Step): Step => {
    // If the step is of type `VIEW` and has no components, set the default components.
    if (step.type === StepTypes.View) {
      if (isEmpty(step?.data?.components)) {
        return {
          ...step,
          data: {
            ...step.data,
            components: getBlankTemplateComponents(),
          },
        };
      }
    }

    const processedStep: Step = generateIdsForResources<Step>(step);

    if (processedStep?.data?.components) {
      processedStep.data.components = resolveComponentMetadata(resources, processedStep.data.components);
    }

    return resolveStepMetadata(resources, [processedStep])[0];
  };

  const generateEdges = (flowSteps: Step[]): Edge[] => {
    let generatedEdges: Edge[] = [];

    // Get all step IDs for validation
    const stepIds: string[] = flowSteps.map((step: Step) => step.id);

    // Find the user onboard step
    const userOnboardStep: Step | undefined = flowSteps.find((step: Step) => step.type === StepTypes.End);

    // Get the ID of the user onboard step or use the default one
    const userOnboardStepId: string = userOnboardStep?.id ?? INITIAL_FLOW_USER_ONBOARD_STEP_ID;

    // Check if we need to connect start to the first step
    if (flowSteps.length > 0) {
      // eslint-disable-next-line prefer-destructuring
      let firstStep: Step = flowSteps[0];

      // TODO: Handle this better. Templates have a `Start` node, but the black starter doesn't.
      if (firstStep.id === INITIAL_FLOW_START_STEP_ID) {
        // eslint-disable-next-line prefer-destructuring
        firstStep = flowSteps[1];
      }

      if (firstStep) {
        generatedEdges.push({
          animated: false,
          id: `${INITIAL_FLOW_START_STEP_ID}-${firstStep.id}`,
          markerEnd: {
            type: MarkerType.Arrow,
          },
          source: INITIAL_FLOW_START_STEP_ID,
          sourceHandle: `${INITIAL_FLOW_START_STEP_ID}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`,
          target: firstStep.id,
          type: 'base-edge',
        });
      }
    }

    // Flag to track if we've already created an edge to the user onboard step
    let userOnboardEdgeCreated = false;

    const createEdgesForButtons = (step: Step, button: Element) => {
      const buttonEdges: Edge[] = [];

      if (button.action?.next) {
        // If next points to a valid step, create that edge
        if (stepIds.includes(button.action.next)) {
          buttonEdges.push({
            animated: false,
            id: button.id,
            markerEnd: {
              type: MarkerType.Arrow,
            },
            source: step.id,
            sourceHandle: `${button.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`,
            target: button.action.next,
            type: 'base-edge',
          });

          // Check if this is pointing to the user onboard step
          if (button.action.next === userOnboardStepId) {
            userOnboardEdgeCreated = true;
          }
        } else if (button.action.next === StepTypes.End) {
          // If next references a user onboard ID that's not in the steps
          // but follows the naming pattern, connect to our actual user onboard step
          buttonEdges.push({
            animated: false,
            id: button.id,
            markerEnd: {
              type: MarkerType.Arrow,
            },
            source: step.id,
            sourceHandle: `${button.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`,
            target: userOnboardStepId,
            type: 'base-edge',
          });
          userOnboardEdgeCreated = true;
        }
      } else {
        // For PasswordProvisioningExecutor buttons without explicit next,
        // create an edge to the user onboard step
        buttonEdges.push({
          animated: false,
          id: button.id,
          markerEnd: {
            type: MarkerType.Arrow,
          },
          source: step.id,
          sourceHandle: `${button.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`,
          target: userOnboardStepId,
          type: 'base-edge',
        });
        userOnboardEdgeCreated = true;
      }

      return buttonEdges;
    };

    // Create edges based on the action configuration in each step
    flowSteps.forEach((step: Step) => {
      // Skip processing for the user onboard step itself
      if (step.type === StepTypes.End) {
        return;
      }

      // Check if the step has components with actions
      if (step.data?.components) {
        // Look for forms and their buttons
        step.data.components.forEach((component: Element) => {
          if (component.type === BlockTypes.Form) {
            const buttons: Element[] | undefined = component.components?.filter(
              (elem: Element) => elem.type === ElementTypes.Button,
            );

            buttons?.forEach((button: Element) => {
              generatedEdges = [...generatedEdges, ...createEdgesForButtons(step, button)];
            });
          }

          if (component.type === ElementTypes.Button) {
            generatedEdges = [...generatedEdges, ...createEdgesForButtons(step, component)];
          }
        });
      }

      // Check if the step has an action with a next step
      if (step.data?.action?.next) {
        // If next points to a valid step, create that edge
        if (stepIds.includes(step.data.action.next)) {
          generatedEdges.push({
            animated: false,
            id: `${step.id}-to-${step.data.action.next}`,
            markerEnd: {
              type: MarkerType.Arrow,
            },
            source: step.id,
            sourceHandle: `${step.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`,
            target: step.data.action.next,
            type: 'base-edge',
          });

          // Check if this is pointing to the user onboard step
          if (step.data.action.next === userOnboardStepId) {
            userOnboardEdgeCreated = true;
          }
        } else if (step.data.action.next === StepTypes.End) {
          // If next references a user onboard ID that's not in the steps
          // but follows the naming pattern, connect to our actual user onboard step
          generatedEdges.push({
            animated: false,
            id: `${step.id}-to-${userOnboardStepId}`,
            markerEnd: {
              type: MarkerType.Arrow,
            },
            source: step.id,
            sourceHandle: `${step.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`,
            target: userOnboardStepId,
            type: 'base-edge',
          });
          userOnboardEdgeCreated = true;
        }
      }
    });

    // If no edge to user onboard was created and we have view steps,
    // connect the last view step to the user onboard step
    if (!userOnboardEdgeCreated && flowSteps.length > 0) {
      // Find view steps
      const viewSteps: Step[] = flowSteps.filter((step: Step) => step.type === StepTypes.View);

      if (viewSteps.length > 0) {
        // Get the last view step
        const lastViewStep: Step = viewSteps[viewSteps.length - 1];

        // Find a button in this step to use for the connection
        let buttonId: string | null = null;

        if (lastViewStep.data?.components) {
          const formComponent = lastViewStep.data.components.find(
            (component: Element) => component.type === BlockTypes.Form,
          );

          if (formComponent?.components) {
            const button = formComponent.components.find((elem: Element) => elem.type === ElementTypes.Button);

            if (button) {
              buttonId = button.id;
            }
          }
        }

        // If we found a button, use it; otherwise generate a fallback ID
        const edgeId: string = buttonId ?? `${lastViewStep.id}-to-${userOnboardStepId}`;

        generatedEdges.push({
          animated: false,
          id: edgeId,
          markerEnd: {
            type: MarkerType.Arrow,
          },
          source: lastViewStep.id,
          ...(buttonId
            ? {sourceHandle: `${buttonId}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`}
            : {sourceHandle: `${lastViewStep.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`}),
          target: userOnboardStepId,
          type: 'base-edge',
        });
      }
    }

    return generatedEdges;
  };

  const generateSteps = useCallback(
    (stepNodes: Node[]): Node[] => {
      const START_STEP: Node = {
        data: {
          displayOnly: true,
        },
        deletable: false,
        id: INITIAL_FLOW_START_STEP_ID,
        position: {x: -300, y: 330},
        type: StaticStepTypes.Start,
      };

      return generateIdsForResources<Node[]>(
        resolveStepMetadata(resources, [
          START_STEP,
          ...stepNodes.map((step: Node) => ({
            data: step.data?.components
              ? {
                  ...step.data,
                  components: resolveComponentMetadata(resources, step.data.components as Element[]),
                }
              : (step.data ?? {}),
            deletable: step.type === StepTypes.End ? false : isStepDeletable(step),
            id: step.id,
            position: step.position,
            type: step.type,
          })),
        ] as Step[]),
      );
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [resources],
  );

  /**
   * Helper function to update node internals for a given node and its components.
   *
   * @param nodes - Set of nodes to update.
   */
  const updateAllNodeInternals = (updatedNodes: Node[]): void => {
    updatedNodes.forEach((node: Node) => {
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

  const initialNodes: Node[] = useMemo<Node[]>(() => {
    const basicTemplate: Template | undefined = cloneDeep(
      resources?.templates?.find((template: Template) => template.type === TemplateTypes.Basic),
    );

    const templateSteps: Step[] = basicTemplate?.config?.data?.steps ?? [];

    // generateSteps already handles ID generation via generateIdsForResources
    return generateSteps(templateSteps);
  }, [generateSteps, resources?.templates]);

  const generateUnconnectedEdges = (currentEdges: Edge[], currentNodes: Node[]): Edge[] => {
    const nodeIds = new Set<string>(currentNodes.map((node: Node) => node.id));
    const missingEdges: Edge[] = [];

    const processAction = (stepId: string, resourceId: string, action: unknown): void => {
      if (action && typeof action === 'object' && 'next' in action && action.next) {
        const buttonId: string = resourceId;
        const expectedTarget: string = action.next as string;

        // Ensure expected target exists in nodes
        if (!nodeIds.has(expectedTarget)) {
          // Log warning for missing target node
          return;
        }

        const existingEdge: Edge | undefined = currentEdges.find(
          (edge: Edge) => edge.source === stepId && edge.sourceHandle === `${buttonId}_NEXT`,
        );

        // If no edge exists or it's pointing to the wrong node, add a missing edge
        if (!existingEdge || existingEdge.target !== expectedTarget) {
          missingEdges.push({
            animated: false,
            id: `${buttonId}_MISSING_EDGE`,
            markerEnd: {
              type: MarkerType.Arrow,
            },
            source: stepId,
            sourceHandle: `${buttonId}_NEXT`,
            target: expectedTarget,
            type: 'base-edge',
          });
        }
      }
    };

    currentNodes.forEach((node: Node) => {
      if (!node.data) {
        return;
      }

      if (node.data?.components) {
        (node.data.components as Element[]).forEach((component: Element) => {
          processAction(node.id, component.id, component.action);

          // Process `FORM` components.
          if (component?.components) {
            component.components.forEach((nestedComponent: Element) =>
              processAction(node?.id, nestedComponent.id, nestedComponent.action),
            );
          }
        });
      }

      if (node.data?.action) {
        processAction(node.id, node.id, node.data.action);
      }
    });

    return missingEdges;
  };

  const handleWidgetLoad = (
    widget: Widget,
    targetResource: Resource,
    currentNodes: Node[],
    currentEdges: Edge[],
  ): [Node[], Edge[], Resource | null, string | null] => {
    const widgetFlow = widget.config.data as {
      steps?: Step[];
      // eslint-disable-next-line no-underscore-dangle
      __generationMeta__?: {
        replacers?: TemplateReplacer[];
        defaultPropertySelectorId?: string;
      };
    };

    if (!widgetFlow?.steps) {
      return [currentNodes, currentEdges, null, null];
    }

    let newNodes: Node[] = cloneDeep(currentNodes);
    let newEdges: Edge[] = cloneDeep(currentEdges);

    // Custom merge function to handle components specifically
    const customMerge = (objValue: unknown, srcValue: unknown, key: string): Element[] | undefined => {
      // Check if the key is 'components' and both are arrays
      if (key === 'components' && Array.isArray(objValue) && Array.isArray(srcValue)) {
        // Concatenate the arrays - don't use unionWith as it prevents adding multiple
        // similar components (like multiple social login buttons) before IDs are generated
        return [...(objValue as Element[]), ...(srcValue as Element[])];
      }

      return undefined;
    };

    widgetFlow.steps.forEach((step: Step) => {
      // eslint-disable-next-line no-underscore-dangle
      if (
        step.__generationMeta__ &&
        typeof step.__generationMeta__ === 'object' &&
        'strategy' in step.__generationMeta__
      ) {
        // eslint-disable-next-line no-underscore-dangle
        const {strategy} = step.__generationMeta__ as {strategy?: string};

        if (strategy === 'MERGE_WITH_DROP_POINT') {
          newNodes = newNodes.map((node: Node) => {
            if (node.id === targetResource.id) {
              // Use mergeWith with the custom merge function
              return mergeWith(node, step, customMerge);
            }

            return node;
          });
        }
      } else {
        newNodes = [...newNodes, step] as Node[];
      }
    });

    // eslint-disable-next-line no-underscore-dangle
    const replacers = widgetFlow.__generationMeta__?.replacers ?? [];
    // eslint-disable-next-line no-underscore-dangle
    const defaultPropertySelectorId = widgetFlow.__generationMeta__?.defaultPropertySelectorId;
    let defaultPropertySectorStepId: string | null = null;
    let defaultPropertySelector: Resource | null = null;

    // Resolve step & component metadata.
    newNodes = resolveStepMetadata(
      resources,
      generateIdsForResources<Node[]>(
        newNodes.map((step: Node) => ({
          data:
            (step.data?.components
              ? {
                  ...step.data,
                  components: resolveComponentMetadata(resources, step.data.components as Element[]),
                }
              : step.data) ?? {},
          deletable: true,
          id: step.id,
          position: step.position,
          type: step.type,
        })),
      ) as Step[],
    ) as Node[];

    // TODO: Improve this block perf.
    newNodes.forEach((node: Node) => {
      if (node.id === defaultPropertySelectorId) {
        defaultPropertySectorStepId = node.id;
        defaultPropertySelector = node as Resource;

        return;
      }

      if (!isEmpty(node?.data?.components)) {
        (node.data.components as Element[]).forEach((component: Element) => {
          if (component.id === defaultPropertySelectorId) {
            defaultPropertySectorStepId = node.id;
            defaultPropertySelector = component as Resource;

            return;
          }

          if (!isEmpty(component?.components)) {
            if (component.id === defaultPropertySelectorId) {
              defaultPropertySectorStepId = node.id;
              defaultPropertySelector = component as Resource;
            }
          }
        });
      }
    });

    const [updatedNodes, replacedPlaceholders] = updateTemplatePlaceholderReferences(
      generateIdsForResources(newNodes),
      replacers,
    );

    newEdges = [...newEdges, ...generateUnconnectedEdges(newEdges, updatedNodes)];

    // Check if `defaultPropertySelector.id` is in the `replacedPlaceholders`.
    // If so, update them with the replaced value.
    if (defaultPropertySelector && 'id' in defaultPropertySelector) {
      const selectorId = (defaultPropertySelector as Resource & {id: string}).id;

      if (typeof selectorId === 'string') {
        const cleanedId = selectorId.replace(/[{}]/g, '');

        if (replacedPlaceholders.has(cleanedId)) {
          const replacedId = replacedPlaceholders.get(cleanedId);

          if (replacedId) {
            (defaultPropertySelector as {id: string}).id = replacedId;
          }
        }
      }
    }

    // Check if `defaultPropertySectorStepId` is in the `replacedPlaceholders`.
    // If so, update them with the replaced value.
    if (defaultPropertySectorStepId) {
      const stepId: string = defaultPropertySectorStepId;
      const cleanedId = stepId.replace(/[{}]/g, '');

      if (replacedPlaceholders.has(cleanedId)) {
        const replacedId = replacedPlaceholders.get(cleanedId);

        if (replacedId) {
          defaultPropertySectorStepId = replacedId;
        }
      }
    }

    return [updatedNodes, newEdges, defaultPropertySelector, defaultPropertySectorStepId];
  };

  /**
   * Uses a sequence of state updates and RAF calls.
   *
   * @param steps - Steps to update the flow with.
   */
  const updateFlowWithSequence = (updatedSteps: Node[]): void => {
    if (flowUpdatesInProgress.current) {
      // "Flow update already in progress, skipping..
      return;
    }

    flowUpdatesInProgress.current = true;
    nodesUpdatedRef.current = false;

    setNodes(() => {
      nodesUpdatedRef.current = true;

      return updatedSteps;
    });

    const updateSequence = () => {
      if (!nodesUpdatedRef.current) {
        requestAnimationFrame(updateSequence);

        return;
      }

      updateAllNodeInternals(updatedSteps);

      requestAnimationFrame(() => {
        const generatedEdges: Edge[] = generateEdges(updatedSteps as Step[]);
        const validatedEdges: Edge[] = validateEdges(generatedEdges, updatedSteps);

        setEdges(() => validatedEdges);

        flowUpdatesInProgress.current = false;
      });
    };

    requestAnimationFrame(updateSequence);
  };

  /**
   * Effect that updates the flow with the password recovery flow steps.
   */
  useLayoutEffect(() => {
    // if (!isAskPasswordFlowFetchRequestLoading && !isAskPasswordFlowFetchRequestValidating) {
    // if (!isEmpty(askPasswordFlow?.steps)) {
    //   const steps: Node[] = generateSteps(askPasswordFlow.steps);

    //   updateFlowWithSequence(steps);
    // } else {
    // }
    updateFlowWithSequence(initialNodes);
    // }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  /**
   * Handle restore from history event.
   */
  useEffect(() => {
    const handleRestoreFromHistory = (event: CustomEvent) => {
      const {nodes: restoredNodes, edges: restoredEdges} = event.detail as {nodes?: Node[]; edges?: Edge[]};

      if (restoredNodes && restoredEdges) {
        setNodes(restoredNodes);
        setEdges(restoredEdges);
      }
    };

    window.addEventListener('restoreFromHistory', handleRestoreFromHistory as EventListener);

    return () => {
      window.removeEventListener('restoreFromHistory', handleRestoreFromHistory as EventListener);
    };
  }, [setNodes, setEdges]);

  const handleTemplateLoad = (template: Template): [Node[], Edge[], FlowResource?, string?] => {
    if (!template?.config?.data?.steps) {
      return [[], [], {} as FlowResource, ''];
    }

    // eslint-disable-next-line no-underscore-dangle
    const replacers: TemplateReplacer[] | undefined = template?.config?.data?.__generationMeta__?.replacers;

    // Check for End steps and set flow completion configs before processing
    template.config.data.steps.forEach((step: Step) => {
      if (step.type === StepTypes.End) {
        if (step?.config) {
          setFlowCompletionConfigs(step.config);
        }
      }
    });

    const templateSteps: Node[] = replacers
      ? updateTemplatePlaceholderReferences(generateSteps(template.config.data.steps), replacers)[0]
      : generateSteps(template.config.data.steps);

    const templateEdges: Edge[] = validateEdges(generateEdges(templateSteps as Step[]), templateSteps);

    // Handle BASIC_FEDERATED template case.
    if (template.type === TemplateTypes.BasicFederated) {
      const googleExecutionStep: Node | undefined = templateSteps.find(
        (step: Node) => step.type === StepTypes.Execution,
      );

      if (googleExecutionStep) {
        return [templateSteps, templateEdges, googleExecutionStep as unknown as FlowResource, googleExecutionStep.id];
      }
    }

    return [templateSteps, templateEdges, {} as FlowResource, ''];
  };

  const handleMutateComponents = (components: Element[]): Element[] => {
    // Filter out any non-element resources (like widgets) that may have been accidentally added
    let modifiedComponents: Element[] = cloneDeep(components).filter(
      (c: Element) => !c.resourceType || c.resourceType === 'ELEMENT',
    );

    const formCount: number = modifiedComponents.filter((c: Element) => c.type === BlockTypes.Form).length;

    if (formCount > 1) {
      let firstFormFound = false;

      modifiedComponents = modifiedComponents.filter((c: Element) => {
        if (c.type === BlockTypes.Form) {
          if (!firstFormFound) {
            firstFormFound = true;

            return true;
          }

          return false;
        }

        return true;
      });
    }

    // Check inside `forms`, if there is a form with a password field and there's only one submit button,
    // Set the `"action": { "type": "EXECUTOR", "executor": { "name": "PasswordProvisioningExecutor"}, "next": "" }`
    modifiedComponents = modifiedComponents.map((component: Element) => {
      if (component.type === BlockTypes.Form) {
        // Set all the `PRIMARY` buttons inside the form type to `submit`.
        const updatedComponents = component?.components?.map((formComponent: Element) => {
          if (formComponent.type === ElementTypes.Button && formComponent.variant === ButtonVariants.Primary) {
            return {
              ...formComponent,
              config: {
                ...formComponent.config,
                type: ButtonTypes.Submit,
              },
            };
          }

          return formComponent;
        });

        const hasPasswordField: boolean = updatedComponents
          ? updatedComponents.some(
              (formComponent: Element) =>
                formComponent.type === ElementTypes.Input && formComponent.variant === InputVariants.Password,
            )
          : false;

        const hasOtpField: boolean = updatedComponents
          ? updatedComponents.some(
              (formComponent: Element) =>
                formComponent.type === ElementTypes.Input && formComponent.variant === InputVariants.OTP,
            )
          : false;

        const submitButtons: Element[] = updatedComponents
          ? updatedComponents.filter(
              (formComponent: Element) =>
                formComponent.type === ElementTypes.Button &&
                'type' in (formComponent.config ?? {}) &&
                (formComponent.config as {type?: string})?.type === ButtonTypes.Submit,
            )
          : [];

        const finalComponents =
          submitButtons?.length === 1 && updatedComponents
            ? updatedComponents.map((formComponent: Element) => {
                if (hasPasswordField) {
                  if (formComponent.type === ElementTypes.Button) {
                    return {
                      ...formComponent,
                      action: {
                        ...(formComponent?.action ?? {}),
                        executor: {
                          name: 'AskPasswordFlowExecutorConstants.PASSWORD_PROVISIONING_EXECUTOR',
                        },
                        type: 'EXECUTOR',
                      },
                    };
                  }
                } else if (hasOtpField) {
                  if (formComponent.type === ElementTypes.Button) {
                    return {
                      ...formComponent,
                      action: {
                        ...(formComponent?.action ?? {}),
                        executor: {
                          name: 'AskPasswordFlowExecutorConstants.EMAIL_OTP_EXECUTOR',
                        },
                        type: 'EXECUTOR',
                      },
                    };
                  }
                }

                return formComponent;
              })
            : updatedComponents;

        return {
          ...component,
          components: finalComponents,
        };
      }

      return component;
    });

    return modifiedComponents;
  };

  /**
   * Ref to store the callback for adding elements to view.
   * Using a ref ensures nodeTypes doesn't need to be recreated when the callback changes.
   */
  const handleAddElementToViewRef = useRef<(element: Element) => void>(() => {});

  /**
   * Ref to store the callback for adding elements to form.
   * Using a ref ensures nodeTypes doesn't need to be recreated when the callback changes.
   */
  const handleAddElementToFormRef = useRef<(element: Element, formId: string) => void>(() => {});

  /**
   * Callback for adding an element to a view from the context menu.
   * Adds the element to an existing View or creates a new View if needed.
   */
  const handleAddElementToView = useCallback(
    (element: Element): void => {
      // Use generateStepElement to properly apply variants and generate unique IDs
      const generatedElement: Element = generateStepElement(element);
      let viewStepId: string | null = null;

      setNodes((prevNodes: Node[]) => {
        // Find existing View in the current nodes
        const existingViewStep = prevNodes.find((node) => node.type === StepTypes.View);

        if (!existingViewStep) {
          return prevNodes; // No View exists, do nothing
        }

        // Store view ID for later use
        viewStepId = existingViewStep.id;

        // For INPUT elements, add them to Form (create Form if needed)
        if (element.type === ElementTypes.Input) {
          return prevNodes.map((node) => {
            if (node.id === existingViewStep.id) {
              const nodeData = node.data as {components?: Element[]} | undefined;
              const existingComponents: Element[] = cloneDeep(nodeData?.components ?? []);

              // Find existing Form in the View
              const existingForm = existingComponents.find((comp: Element) => comp.type === BlockTypes.Form);

              if (existingForm) {
                // Add input to existing Form
                const updatedForm: Element = {
                  ...existingForm,
                  components: [...(existingForm.components ?? []), generatedElement],
                };

                const componentsWithoutForm = existingComponents.filter(
                  (comp: Element) => comp.type !== BlockTypes.Form,
                );

                return {
                  ...node,
                  data: {
                    ...nodeData,
                    components: handleMutateComponents([...componentsWithoutForm, updatedForm]),
                  },
                };
              }

              // No Form exists - create a new Form with the input
              const newForm: Element = generateIdsForResources<Element>({
                id: '{{ID}}',
                resourceType: ResourceTypes.Element,
                category: ElementCategories.Block,
                type: BlockTypes.Form,
                version: '0.1.0',
                deprecated: false,
                display: {
                  label: 'Form',
                  image: 'assets/images/icons/form.svg',
                },
                config: {},
                components: [generatedElement],
              } as Element);

              return {
                ...node,
                data: {
                  ...nodeData,
                  components: handleMutateComponents([...existingComponents, newForm]),
                },
              };
            }
            return node;
          });
        }

        // For other elements (Buttons, etc.), add to existing View
        return prevNodes.map((node) => {
          if (node.id === existingViewStep.id) {
            const nodeData = node.data as {components?: Element[]} | undefined;
            const existingComponents: Element[] = cloneDeep(nodeData?.components ?? []);

            return {
              ...node,
              data: {
                ...nodeData,
                components: handleMutateComponents([...existingComponents, generatedElement]),
              },
            };
          }
          return node;
        });
      });

      // Schedule node internals update after state has been updated
      if (viewStepId) {
        setTimeout(() => {
          updateNodeInternals(viewStepId!);
        }, 50);
      }
    },
    [setNodes, handleMutateComponents, updateNodeInternals, generateStepElement],
  );

  /**
   * Callback for adding an element to a form from the context menu.
   * Adds the element to the specified form.
   */
  const handleAddElementToForm = useCallback(
    (element: Element, formId: string): void => {
      // Use generateStepElement to properly apply variants and generate unique IDs
      const generatedElement: Element = generateStepElement(element);
      let viewStepId: string | null = null;

      setNodes((prevNodes: Node[]) => {
        // Find the View node that contains the form
        const existingViewStep = prevNodes.find((node) => node.type === StepTypes.View);

        if (!existingViewStep) {
          return prevNodes;
        }

        viewStepId = existingViewStep.id;

        return prevNodes.map((node) => {
          if (node.id === existingViewStep.id) {
            const nodeData = node.data as {components?: Element[]} | undefined;
            const existingComponents: Element[] = cloneDeep(nodeData?.components ?? []);

            // Find the form and add the element to it
            const updatedComponents = existingComponents.map((component: Element) => {
              if (component.id === formId && component.type === BlockTypes.Form) {
                return {
                  ...component,
                  components: [...(component.components ?? []), generatedElement],
                };
              }
              return component;
            });

            return {
              ...node,
              data: {
                ...nodeData,
                components: handleMutateComponents(updatedComponents),
              },
            };
          }
          return node;
        });
      });

      // Schedule node internals update after state has been updated
      if (viewStepId) {
        setTimeout(() => {
          updateNodeInternals(viewStepId!);
        }, 50);
      }
    },
    [setNodes, handleMutateComponents, updateNodeInternals, generateStepElement],
  );

  // Update refs whenever callbacks change
  useEffect(() => {
    handleAddElementToViewRef.current = handleAddElementToView;
    handleAddElementToFormRef.current = handleAddElementToForm;
  }, [handleAddElementToView, handleAddElementToForm]);

  const nodeTypes = useMemo((): NodeTypes => {
    if (!steps) {
      return {};
    }

    const stepsByType: Record<string, Step[]> = steps.reduce((acc: Record<string, Step[]>, step: Step) => {
      if (!acc[step.type]) {
        acc[step.type] = [];
      }
      acc[step.type].push(step);

      return acc;
    }, {});

    const stepNodes: NodeTypes = steps.reduce((acc: NodeTypes, resource: Step) => {
      acc[resource.type] = (props: NodeProps) => (
        // @ts-expect-error NodeProps doesn't include all required properties but they're provided at runtime
        <StepFactory
          resourceId={props.id}
          resources={stepsByType[resource.type]}
          allResources={resources}
          onAddElement={(element: Element) => handleAddElementToViewRef.current?.(element)}
          onAddElementToForm={(element: Element, formId: string) =>
            handleAddElementToFormRef.current?.(element, formId)
          }
          {...props}
        />
      );

      return acc;
    }, {});

    const staticStepNodes: NodeTypes = Object.values(StaticStepTypes).reduce(
      (acc: NodeTypes, type: StaticStepTypes) => {
        acc[type] = (props: NodeProps) => (
          // @ts-expect-error NodeProps doesn't include all required properties but they're provided at runtime
          <StaticStepFactory {...props} type={type} />
        );

        return acc;
      },
      {},
    );

    return {
      ...staticStepNodes,
      ...stepNodes,
    };
  }, [steps, resources]);

  /**
   * Handle save button click - transforms React Flow data to backend format.
   * @param canvasData - The canvas data from React Flow (nodes, edges, viewport).
   */
  const handleSave = useCallback(
    async (canvasData: {nodes: Node[]; edges: Edge[]; viewport: {x: number; y: number; zoom: number}}) => {
      try {
        console.log('canvasData', canvasData);

        const {transformReactFlow, validateFlowGraph} = await import('@/features/flows/utils/reactFlowTransformer');

        const flowGraph = transformReactFlow(canvasData);
        const errors = validateFlowGraph(flowGraph);

        if (errors.length > 0) {
          // eslint-disable-next-line no-console
          console.error('Flow validation errors:', errors);
          // TODO: Show validation errors to user
          return;
        }

        // eslint-disable-next-line no-console
        console.log('✅ Transformed flow:', flowGraph);

        // TODO: Send to backend API
        // await saveLoginFlow(flowGraph);
      } catch (error) {
        // eslint-disable-next-line no-console
        console.error('Error saving flow:', error);
        // TODO: Show error notification to user
      }
    },
    [],
  );

  return (
    <Box
      sx={{
        width: '100%',
        height: '100vh',
      }}
    >
      <FlowBuilder
        resources={resources}
        nodeTypes={nodeTypes}
        mutateComponents={handleMutateComponents}
        onTemplateLoad={handleTemplateLoad}
        onWidgetLoad={handleWidgetLoad}
        onStepLoad={handleStepLoad}
        onResourceAdd={handleResourceAdd}
        onSave={handleSave}
        nodes={nodes}
        edges={edges}
        setNodes={setNodes}
        setEdges={setEdges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
      />
    </Box>
  );
}

export default LoginFlowBuilder;
