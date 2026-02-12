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

import {Alert, Box, Snackbar, Stack} from '@wso2/oxygen-ui';
import {useParams} from 'react-router';
import type {Edge, Node} from '@xyflow/react';
import {useEdgesState, useNodesState, useUpdateNodeInternals} from '@xyflow/react';
import {useCallback, useEffect, useMemo} from 'react';
import {useTranslation} from 'react-i18next';
import '@xyflow/react/dist/style.css';
import FlowBuilder from '@/features/flows/components/FlowBuilder';
import {StepTypes} from '@/features/flows/models/steps';
import useFlowBuilderCore from '@/features/flows/hooks/useFlowBuilderCore';
import useValidationStatus from '@/features/flows/hooks/useValidationStatus';
import useGetFlowById from '@/features/flows/api/useGetFlowById';
import GradientBorderButton from '@/features/applications/components/GradientBorderButton';
import useGetLoginFlowBuilderResources from '../api/useGetLoginFlowBuilderResources';
import useEdgeGeneration from '../hooks/useEdgeGeneration';
import useFlowNaming from '../hooks/useFlowNaming';
import useSnackbarNotifications from '../hooks/useSnackbarNotifications';
import useFlowInitialization from '../hooks/useFlowInitialization';
import useNodeTypes from '../hooks/useNodeTypes';
import useFlowSave from '../hooks/useFlowSave';
import useElementAddition from '../hooks/useElementAddition';
import useTemplateAndWidgetLoading from '../hooks/useTemplateAndWidgetLoading';
import {mutateComponents} from '../utils/componentMutations';
import LoginFlowConstants from '../constants/LoginFlowConstants';

function LoginFlowBuilder() {
  const {flowId} = useParams<{flowId: string}>();
  const [nodes, setNodes, defaultOnNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const {t} = useTranslation();

  const {data: resources} = useGetLoginFlowBuilderResources();
  const {edgeStyle, isVerboseMode} = useFlowBuilderCore();
  const {isValid: isFlowValid, setOpenValidationPanel} = useValidationStatus();
  const updateNodeInternals = useUpdateNodeInternals();

  // Fetch the existing flow if flowId is provided (editing an existing flow)
  const {data: existingFlowData, isLoading: isLoadingExistingFlow} = useGetFlowById(flowId);

  // Determine if we're editing an existing flow
  const isEditingExistingFlow = Boolean(flowId && existingFlowData);

  // Flow naming hook
  const {flowName, flowHandle, needsAutoLayout, setNeedsAutoLayout, handleFlowNameChange} = useFlowNaming({
    existingFlowData: existingFlowData as {name?: string; handle?: string} | undefined,
  });

  // Snackbar notifications hook
  const {
    errorSnackbar,
    successSnackbar,
    infoSnackbar,
    showError,
    showSuccess,
    showInfo,
    handleCloseErrorSnackbar,
    handleCloseSuccessSnackbar,
    handleCloseInfoSnackbar,
  } = useSnackbarNotifications();

  // Callback to trigger auto-layout from the snackbar via custom event
  const handleAutoLayoutClick = useCallback(() => {
    window.dispatchEvent(new CustomEvent('triggerAutoLayout'));
    handleCloseInfoSnackbar();
  }, [handleCloseInfoSnackbar]);

  // Edge generation hook
  const {generateEdges, validateEdges} = useEdgeGeneration({
    startStepId: LoginFlowConstants.START_STEP_ID,
    endStepId: LoginFlowConstants.END_STEP_ID,
  });

  // Flow initialization hook
  const {generateSteps, getBlankTemplateComponents} = useFlowInitialization({
    resources,
    flowId,
    existingFlowData,
    isLoadingExistingFlow,
    setNodes,
    setEdges,
    updateNodeInternals,
    generateEdges,
    validateEdges,
    edgeStyle,
    onNeedsAutoLayout: setNeedsAutoLayout,
  });

  // Element addition hook
  const {handleAddElementToView, handleAddElementToForm} = useElementAddition({
    setNodes,
    updateNodeInternals,
  });

  // Node types hook
  const {nodeTypes, edgeTypes} = useNodeTypes({
    steps: resources.steps,
    resources,
    onAddElementToView: handleAddElementToView,
    onAddElementToForm: handleAddElementToForm,
  });

  // Template and widget loading hook
  const {handleStepLoad, handleTemplateLoad, handleWidgetLoad, handleResourceAdd} = useTemplateAndWidgetLoading({
    resources,
    generateSteps,
    generateEdges,
    validateEdges,
    getBlankTemplateComponents,
    setNodes,
    updateNodeInternals,
  });

  // Flow save hook
  const {handleSave} = useFlowSave({
    flowId,
    isEditingExistingFlow,
    isFlowValid,
    flowName,
    flowHandle,
    showError,
    showSuccess,
    setOpenValidationPanel,
  });

  const onNodesChange = defaultOnNodesChange;

  // Handle restore from history event
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

  // Listen for element added events to show auto-layout hint
  useEffect(() => {
    const handleElementAdded = (event: CustomEvent<{type: string}>) => {
      const {type} = event.detail;
      // Only show hint for steps, widgets, and templates (not individual components)
      if (type === 'step' || type === 'widget' || type === 'template') {
        showInfo(t('flows:core.canvas.hints.autoLayout'));
      }
    };

    window.addEventListener('flowElementAdded', handleElementAdded as EventListener);

    return () => {
      window.removeEventListener('flowElementAdded', handleElementAdded as EventListener);
    };
  }, [showInfo, t]);

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
        triggerAutoLayoutOnLoad={needsAutoLayout}
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
      <Snackbar
        open={infoSnackbar.open}
        autoHideDuration={8000}
        onClose={handleCloseInfoSnackbar}
        anchorOrigin={{vertical: 'top', horizontal: 'center'}}
      >
        <Alert
          onClose={handleCloseInfoSnackbar}
          severity="info"
          sx={{
            width: '100%',
            alignItems: 'center',
          }}
        >
          <Stack direction="row" spacing={2} alignItems="center">
            <span>{infoSnackbar.message}</span>
            <GradientBorderButton size="small" onClick={handleAutoLayoutClick}>
              {t('flows:core.canvas.buttons.autoLayout')}
            </GradientBorderButton>
          </Stack>
        </Alert>
      </Snackbar>
    </Box>
  );
}

export default LoginFlowBuilder;
