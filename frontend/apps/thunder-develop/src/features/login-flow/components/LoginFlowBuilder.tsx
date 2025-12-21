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

import {Alert, Box, Snackbar} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import {useNavigate, useParams} from 'react-router';
import type {Edge, EdgeTypes, Node, NodeTypes, NodeProps} from '@xyflow/react';
import {MarkerType, useEdgesState, useNodesState, useUpdateNodeInternals} from '@xyflow/react';
import {useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState} from 'react';
import '@xyflow/react/dist/style.css';
import cloneDeep from 'lodash-es/cloneDeep';
import isEmpty from 'lodash-es/isEmpty';
import mergeWith from 'lodash-es/mergeWith';
import type {UpdateNodeInternals} from '@xyflow/system';
import FlowBuilder from '@/features/flows/components/FlowBuilder';
import BaseEdge from '@/features/flows/components/react-flow-overrides/BaseEdge';
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
import useValidationStatus from '@/features/flows/hooks/useValidationStatus';
import useCreateFlow from '@/features/flows/api/useCreateFlow';
import useUpdateFlow from '@/features/flows/api/useUpdateFlow';
import useGetFlowById from '@/features/flows/api/useGetFlowById';
import type {CreateFlowRequest, UpdateFlowRequest} from '@/features/flows/models/responses';
import {createFlowConfiguration, validateFlowGraph} from '@/features/flows/utils/reactFlowTransformer';
import {transformFlowToCanvas} from '@/features/flows/utils/flowToCanvasTransformer';
import StaticStepFactory from './resources/steps/StaticStepFactory';
import StepFactory from './resources/steps/StepFactory';
import useGetLoginFlowBuilderResources from '../api/useGetLoginFlowBuilderResources';
import useEdgeGeneration from '../hooks/useEdgeGeneration';
import LoginFlowConstants from '../constants/LoginFlowConstants';

// Use centralized executor names constant
const {ExecutorNames} = LoginFlowConstants;

/**
 * Process form components to set button types and auto-assign executors.
 * Optimized to use a single pass through the components.
 */
const processFormComponents = (formComponents: Element[] | undefined): Element[] | undefined => {
  if (!formComponents || formComponents.length === 0) {
    return formComponents;
  }

  // Single pass to collect information and transform
  let hasPasswordField = false;
  let hasOtpField = false;
  let submitButtonCount = 0;

  // First pass: collect info and set PRIMARY buttons to submit
  const updatedComponents = formComponents.map((formComponent: Element) => {
    // Check for field types
    if (formComponent.type === ElementTypes.Input) {
      if (formComponent.variant === InputVariants.Password) {
        hasPasswordField = true;
      } else if (formComponent.variant === InputVariants.OTP) {
        hasOtpField = true;
      }
    }

    // Set PRIMARY buttons to submit type
    if (formComponent.type === ElementTypes.Button && formComponent.variant === ButtonVariants.Primary) {
      const updatedButton = {
        ...formComponent,
        config: {
          ...formComponent.config,
          type: ButtonTypes.Submit,
        },
      };
      submitButtonCount += 1;
      return updatedButton;
    }

    // Count existing submit buttons
    if (
      formComponent.type === ElementTypes.Button &&
      (formComponent.config as {type?: string})?.type === ButtonTypes.Submit
    ) {
      submitButtonCount += 1;
    }

    return formComponent;
  });

  // If exactly one submit button and has password/otp field, assign executor
  if (submitButtonCount === 1 && (hasPasswordField || hasOtpField)) {
    const executorName = hasPasswordField ? ExecutorNames.PASSWORD_PROVISIONING : ExecutorNames.EMAIL_OTP;

    return updatedComponents.map((formComponent: Element) => {
      if (formComponent.type === ElementTypes.Button) {
        return {
          ...formComponent,
          action: {
            ...(formComponent?.action ?? {}),
            executor: {name: executorName},
            type: LoginFlowConstants.ActionTypes.EXECUTOR,
          },
        };
      }
      return formComponent;
    });
  }

  return updatedComponents;
};

/**
 * Mutate components to ensure proper form structure and button actions.
 * This is extracted outside the component to prevent unnecessary re-renders and dependency issues.
 *
 * Optimizations:
 * - Single pass for filtering and counting forms
 * - Separated form processing logic for clarity
 * - Uses typed constants for executor names
 *
 * @param components - The components to mutate
 * @returns The mutated components
 */
const mutateComponents = (components: Element[]): Element[] => {
  // Clone and filter in single pass, tracking form count
  let firstFormFound = false;

  const modifiedComponents = cloneDeep(components).filter((component) => {
    // Filter out non-element resources
    if (component.resourceType && component.resourceType !== 'ELEMENT') {
      return false;
    }

    // Keep only the first form
    if (component.type === BlockTypes.Form) {
      if (firstFormFound) {
        return false;
      }
      firstFormFound = true;
    }

    return true;
  });

  // Process forms and their components
  return modifiedComponents.map((component: Element) => {
    if (component.type === BlockTypes.Form) {
      return {
        ...component,
        components: processFormComponents(component.components),
      };
    }
    return component;
  });
};

function LoginFlowBuilder() {
  const {t} = useTranslation();
  const navigate = useNavigate();
  const {flowId} = useParams<{flowId: string}>();
  const [nodes, setNodes, defaultOnNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  // Use centralized constants for step IDs
  const INITIAL_FLOW_START_STEP_ID = LoginFlowConstants.START_STEP_ID;
  const INITIAL_FLOW_USER_ONBOARD_STEP_ID = LoginFlowConstants.END_STEP_ID;

  const {data: resources} = useGetLoginFlowBuilderResources();
  const {setFlowCompletionConfigs, edgeStyle, isVerboseMode} = useFlowBuilderCore();
  const {generateStepElement} = useGenerateStepElement();
  const {isValid: isFlowValid, setOpenValidationPanel} = useValidationStatus();
  const createFlow = useCreateFlow();
  const updateFlow = useUpdateFlow();

  // Fetch the existing flow if flowId is provided (editing an existing flow)
  const {data: existingFlowData, isLoading: isLoadingExistingFlow} = useGetFlowById(flowId);

  // Determine if we're editing an existing flow
  const isEditingExistingFlow = Boolean(flowId && existingFlowData);

  const [errorSnackbar, setErrorSnackbar] = useState<{open: boolean; message: string}>({
    open: false,
    message: '',
  });
  const [successSnackbar, setSuccessSnackbar] = useState<{open: boolean; message: string}>({
    open: false,
    message: '',
  });

  // Flow name state - initialize from existing flow data or use default
  const [flowName, setFlowName] = useState<string>('Login Flow');
  const [flowHandle, setFlowHandle] = useState<string>('login-flow');

  /**
   * Generate a URL-friendly handle from a name.
   * Converts to lowercase, replaces spaces with hyphens, removes special characters.
   */
  const generateHandleFromName = useCallback(
    (name: string): string =>
      name
        .toLowerCase()
        .trim()
        .replace(/\s+/g, '-')
        .replace(/[^a-z0-9-]/g, '')
        .replace(/-+/g, '-')
        .replace(/^-|-$/g, ''),
    [],
  );

  // Sync flowName and flowHandle when existingFlowData is loaded
  useEffect(() => {
    if (existingFlowData?.name) {
      setFlowName(existingFlowData.name);
    }
    const handle = (existingFlowData as {handle?: string} | undefined)?.handle;
    if (handle) {
      setFlowHandle(handle);
    } else if (existingFlowData?.name) {
      setFlowHandle(generateHandleFromName(existingFlowData.name));
    }
  }, [existingFlowData, generateHandleFromName]);

  const handleFlowNameChange = useCallback(
    (newName: string) => {
      setFlowName(newName);
      setFlowHandle(generateHandleFromName(newName));
    },
    [generateHandleFromName],
  );

  const {steps} = resources;

  // Edge generation hook - extracted for better code organization
  const {generateEdges, validateEdges} = useEdgeGeneration({
    startStepId: INITIAL_FLOW_START_STEP_ID,
    endStepId: INITIAL_FLOW_USER_ONBOARD_STEP_ID,
  });

  const updateNodeInternals: UpdateNodeInternals = useUpdateNodeInternals();
  const flowUpdatesInProgress = useRef<boolean>(false);
  const nodesUpdatedRef = useRef<boolean>(false);

  const onNodesChange = defaultOnNodesChange;

  /**
   * Determines if a step is deletable based on its type and executor.
   * @param _step - The step to check.
   * @returns true if the step is deletable, false otherwise.
   */
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const isStepDeletable = (_step: Node): boolean => true;

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

            const mutatedComponents: Element[] | undefined = mutateComponents(updatedComponents);

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

      // Update node internals for rendering - use queueMicrotask for immediate execution after state update
      queueMicrotask(() => {
        updateNodeInternals(existingViewStep.id);
        updateNodeInternals(generatedElement.id);
      });
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
    // eslint-disable-next-line no-underscore-dangle
    const replacers: TemplateReplacer[] | undefined = basicTemplate?.config?.data?.__generationMeta__?.replacers;

    // generateSteps handles {{ID}} placeholders via generateIdsForResources
    // updateTemplatePlaceholderReferences handles named placeholders like {{GOOGLE_EXECUTION_STEP_ID}}
    const generatedSteps = generateSteps(templateSteps);

    return replacers ? updateTemplatePlaceholderReferences(generatedSteps, replacers)[0] : generatedSteps;
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
            type: edgeStyle,
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
      /* eslint-disable no-underscore-dangle */
      if (
        step.__generationMeta__ &&
        typeof step.__generationMeta__ === 'object' &&
        'strategy' in step.__generationMeta__
      ) {
        const {strategy} = step.__generationMeta__ as {strategy?: string};
        /* eslint-enable no-underscore-dangle */

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

        // Apply current edge style to all edges
        const styledEdges = validatedEdges.map((edge) => ({
          ...edge,
          type: edgeStyle,
        }));

        setEdges(() => styledEdges);

        flowUpdatesInProgress.current = false;
      });
    };

    requestAnimationFrame(updateSequence);
  };

  /**
   * Effect that updates the flow with the existing flow data or default template.
   * When opening an existing flow (flowId is provided), transform and load the saved flow data.
   * Otherwise, load the default basic auth flow template.
   */
  useLayoutEffect(() => {
    // Skip if we're still loading the existing flow data
    if (flowId && isLoadingExistingFlow) {
      return;
    }

    // If we have an existing flow, transform and load it
    if (flowId && existingFlowData) {
      const canvasData = transformFlowToCanvas(existingFlowData);

      // Process nodes to resolve metadata without adding a new START node
      // (the transformed data already contains the START node from the API)
      const processedNodes = generateIdsForResources<Node[]>(
        resolveStepMetadata(
          resources,
          canvasData.nodes.map((node: Node) => ({
            data: node.data?.components
              ? {
                  ...node.data,
                  components: resolveComponentMetadata(resources, node.data.components as Element[]),
                }
              : (node.data ?? {}),
            deletable: node.type !== StepTypes.End && node.type !== StaticStepTypes.Start,
            id: node.id,
            position: node.position,
            type: node.type,
          })) as Step[],
        ),
      );

      // Set nodes and edges from the transformed flow data
      setNodes(processedNodes);
      setEdges(
        canvasData.edges.map((edge) => ({
          ...edge,
          type: edgeStyle,
        })),
      );

      // Update all node internals after setting nodes - use queueMicrotask for immediate execution after state update
      queueMicrotask(() => {
        updateAllNodeInternals(processedNodes);
      });
    } else if (!flowId) {
      // No flowId provided - load the default basic auth flow template
      updateFlowWithSequence(initialNodes);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [flowId, existingFlowData, isLoadingExistingFlow]);

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

    const generatedTemplateEdges: Edge[] = validateEdges(generateEdges(templateSteps as Step[]), templateSteps);
    // Apply current edge style to all template edges
    const templateEdges: Edge[] = generatedTemplateEdges.map((edge) => ({
      ...edge,
      type: edgeStyle,
    }));

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

  /**
   * Ref to store the callback for adding elements to view.
   * Using a ref ensures nodeTypes doesn't need to be recreated when the callback changes.
   */
  const handleAddElementToViewRef = useRef<(element: Element, viewId: string) => void>(() => {});

  /**
   * Ref to store the callback for adding elements to form.
   * Using a ref ensures nodeTypes doesn't need to be recreated when the callback changes.
   */
  const handleAddElementToFormRef = useRef<(element: Element, formId: string) => void>(() => {});

  /**
   * Callback for adding an element to a view from the context menu.
   * Adds the element to the specified View.
   */
  const handleAddElementToView = useCallback(
    (element: Element, viewId: string): void => {
      // Use generateStepElement to properly apply variants and generate unique IDs
      const generatedElement: Element = generateStepElement(element);
      let viewStepId: string | null = null;

      setNodes((prevNodes: Node[]) => {
        // Find the View node with the given viewId
        const existingViewStep = prevNodes.find((node) => node.id === viewId && node.type === StepTypes.View);

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
                    components: mutateComponents([...componentsWithoutForm, updatedForm]),
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
                  components: mutateComponents([...existingComponents, newForm]),
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
                components: mutateComponents([...existingComponents, generatedElement]),
              },
            };
          }
          return node;
        });
      });

      // Schedule node internals update after state has been updated - use queueMicrotask for immediate execution
      if (viewStepId) {
        queueMicrotask(() => {
          updateNodeInternals(viewStepId!);
        });
      }
    },
    [setNodes, updateNodeInternals, generateStepElement],
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
        // Find the View node that contains the form with the given formId
        const existingViewStep = prevNodes.find((node) => {
          if (node.type !== StepTypes.View) return false;
          const nodeData = node.data as {components?: Element[]} | undefined;
          const components = nodeData?.components ?? [];
          // Check if this view contains the form with the given formId
          return components.some((component: Element) => component.id === formId && component.type === BlockTypes.Form);
        });

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
                components: mutateComponents(updatedComponents),
              },
            };
          }
          return node;
        });
      });

      // Schedule node internals update after state has been updated - use queueMicrotask for immediate execution
      if (viewStepId) {
        queueMicrotask(() => {
          updateNodeInternals(viewStepId!);
        });
      }
    },
    [setNodes, updateNodeInternals, generateStepElement],
  );

  // Update refs whenever callbacks change
  useEffect(() => {
    handleAddElementToViewRef.current = handleAddElementToView;
    handleAddElementToFormRef.current = handleAddElementToForm;
  }, [handleAddElementToView, handleAddElementToForm]);

  const resourcesRef = useRef(resources);

  // Update refs when data changes (doesn't trigger re-render)
  useEffect(() => {
    resourcesRef.current = resources;
  }, [resources]);

  // Update edge types when edge style changes
  useEffect(() => {
    setEdges((currentEdges) =>
      currentEdges.map((edge) => ({
        ...edge,
        type: edgeStyle,
      })),
    );
  }, [edgeStyle, setEdges]);

  // Filter nodes and edges based on verbose mode
  const filteredNodes = useMemo(() => {
    if (isVerboseMode) {
      return nodes;
    }
    // Hide execution nodes in non-verbose mode
    return nodes.filter((node) => node.type !== StepTypes.Execution);
  }, [nodes, isVerboseMode]);

  const filteredEdges = useMemo(() => {
    if (isVerboseMode) {
      return edges;
    }
    // Hide edges connected to execution nodes in non-verbose mode
    const executionNodeIds = new Set(nodes.filter((node) => node.type === StepTypes.Execution).map((node) => node.id));
    return edges.filter((edge) => !executionNodeIds.has(edge.source) && !executionNodeIds.has(edge.target));
  }, [edges, nodes, isVerboseMode]);

  const stepsByTypeRef = useRef<Record<string, Step[]>>({});
  useEffect(() => {
    if (steps) {
      stepsByTypeRef.current = steps.reduce((acc: Record<string, Step[]>, step: Step) => {
        if (!acc[step.type]) {
          acc[step.type] = [];
        }
        acc[step.type].push(step);
        return acc;
      }, {});
    }
  }, [steps]);

  const nodeTypes = useMemo((): NodeTypes => {
    // Get unique step types from steps (only to determine which types we need)
    const stepTypeSet = new Set(steps?.map((s) => s.type) ?? []);

    const stepNodes: NodeTypes = Array.from(stepTypeSet).reduce((acc: NodeTypes, stepType: string) => {
      // Create a stable component that reads from refs at render time
      acc[stepType] = (props: NodeProps) => (
        // @ts-expect-error NodeProps doesn't include all required properties but they're provided at runtime
        <StepFactory
          resourceId={props.id}
          resources={stepsByTypeRef.current[stepType] ?? []}
          allResources={resourcesRef.current}
          onAddElement={(element: Element) => handleAddElementToViewRef.current?.(element, props.id)}
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
    // IMPORTANT: Only depend on the step types array, not resources
    // The actual data is accessed via refs at render time
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [steps?.map((s) => s.type).join(',')]);

  // Edge types - register BaseEdge which uses SmartStepEdge for intelligent routing around nodes
  const edgeTypes: EdgeTypes = useMemo(
    () => ({
      default: BaseEdge,
      smoothstep: BaseEdge,
      step: BaseEdge,
      [LoginFlowConstants.DEFAULT_EDGE_TYPE]: BaseEdge,
    }),
    [],
  );

  /**
   * Handle save button click - transforms React Flow data to backend format.
   * @param canvasData - The canvas data from React Flow (nodes, edges, viewport).
   */
  const handleSave = useCallback(
    (canvasData: {nodes: Node[]; edges: Edge[]; viewport: {x: number; y: number; zoom: number}}) => {
      // Check if there are validation errors in the validation panel
      if (!isFlowValid) {
        setErrorSnackbar({
          open: true,
          message: t('flows:core.loginFlowBuilder.errors.validationRequired'),
        });
        setOpenValidationPanel?.(true);
        return;
      }

      const flowConfig = createFlowConfiguration(canvasData, flowName, flowHandle, 'AUTHENTICATION');
      const errors = validateFlowGraph({nodes: flowConfig.nodes});

      if (errors.length > 0) {
        setErrorSnackbar({
          open: true,
          message: t('flows:core.loginFlowBuilder.errors.structureValidationFailed', {error: errors[0]}),
        });
        return;
      }

      // Send to backend API - use update if editing existing flow, create if new
      if (isEditingExistingFlow && flowId) {
        // Update existing flow
        updateFlow.mutate(
          {
            flowId,
            flowData: flowConfig as UpdateFlowRequest,
          },
          {
            onSuccess: () => {
              setSuccessSnackbar({
                open: true,
                message: t('flows:core.loginFlowBuilder.success.flowUpdated'),
              });
              // Redirect to flows list after a short delay to show the success message
              setTimeout(() => {
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                navigate('/flows');
              }, 1500);
            },
            onError: () => {
              setErrorSnackbar({
                open: true,
                message: t('flows:core.loginFlowBuilder.errors.saveFailed'),
              });
            },
          },
        );
      } else {
        // Create new flow
        createFlow.mutate(flowConfig as CreateFlowRequest, {
          onSuccess: () => {
            setSuccessSnackbar({
              open: true,
              message: t('flows:core.loginFlowBuilder.success.flowCreated'),
            });
            // Redirect to flows list after a short delay to show the success message
            setTimeout(() => {
              // eslint-disable-next-line @typescript-eslint/no-floating-promises
              navigate('/flows');
            }, 1500);
          },
          onError: () => {
            setErrorSnackbar({
              open: true,
              message: t('flows:core.loginFlowBuilder.errors.saveFailed'),
            });
          },
        });
      }
    },
    [
      createFlow,
      updateFlow,
      flowId,
      isEditingExistingFlow,
      isFlowValid,
      setOpenValidationPanel,
      t,
      flowName,
      flowHandle,
      navigate,
    ],
  );

  /**
   * Handle closing the error snackbar.
   */
  const handleCloseErrorSnackbar = useCallback(() => {
    setErrorSnackbar((prev) => ({...prev, open: false}));
  }, []);

  /**
   * Handle closing the success snackbar.
   */
  const handleCloseSuccessSnackbar = useCallback(() => {
    setSuccessSnackbar((prev) => ({...prev, open: false}));
  }, []);

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
        edgeTypes={edgeTypes}
        mutateComponents={mutateComponents}
        onTemplateLoad={handleTemplateLoad}
        onWidgetLoad={handleWidgetLoad}
        onStepLoad={handleStepLoad}
        onResourceAdd={handleResourceAdd}
        onSave={handleSave}
        nodes={filteredNodes}
        edges={filteredEdges}
        setNodes={setNodes}
        setEdges={setEdges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        flowTitle={flowName}
        flowHandle={flowHandle}
        onFlowTitleChange={handleFlowNameChange}
      />
      <Snackbar
        open={errorSnackbar.open}
        autoHideDuration={6000}
        onClose={handleCloseErrorSnackbar}
        anchorOrigin={{vertical: 'bottom', horizontal: 'center'}}
      >
        <Alert onClose={handleCloseErrorSnackbar} severity="error" sx={{width: '100%'}}>
          {errorSnackbar.message}
        </Alert>
      </Snackbar>
      <Snackbar
        open={successSnackbar.open}
        autoHideDuration={6000}
        onClose={handleCloseSuccessSnackbar}
        anchorOrigin={{vertical: 'bottom', horizontal: 'center'}}
      >
        <Alert onClose={handleCloseSuccessSnackbar} severity="success" sx={{width: '100%'}}>
          {successSnackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}

export default LoginFlowBuilder;
