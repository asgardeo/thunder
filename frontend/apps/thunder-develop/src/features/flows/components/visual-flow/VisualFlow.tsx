/**
 * Copyright (c) 2023-2025, WSO2 LLC. (https://www.wso2.com).
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

import {
  Background,
  ControlButton,
  Controls,
  type EdgeTypes,
  type NodeTypes,
  ReactFlow,
  type ReactFlowProps,
} from '@xyflow/react';
import {type ReactElement, useCallback} from 'react';
import {Button, Card, Tooltip, useColorScheme} from '@wso2/oxygen-ui';
import '@xyflow/react/dist/style.css';
import './VisualFlow.scss';
import {LayoutGrid, Save} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import CanvasValidationIndicator from '../validation-panel/CanvasValidationIndicator';
import EdgeStyleMenu from './EdgeStyleSelector';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import useEdgeStyleSelector from '../../hooks/useEdgeStyleSelector';
import getEdgeStyleIcon from '../../utils/getEdgeStyleIcon';

/**
 * Props interface of {@link VisualFlow}
 */
export interface VisualFlowPropsInterface extends ReactFlowProps {
  /**
   * Edge types to be rendered.
   */
  edgeTypes?: EdgeTypes;
  /**
   * Node types to be rendered.
   */
  nodeTypes?: NodeTypes;
  /**
   * Callback to be triggered when auto-layout button is clicked.
   */
  handleAutoLayout?: () => void;
  /**
   * Callback to be triggered when save button is clicked.
   */
  onSave?: () => void;
}

/**
 * Wrapper component for React Flow used in the Visual Editor.
 *
 * @param props - Props injected to the component.
 * @returns Visual editor flow component.
 */
function VisualFlow({
  nodeTypes = {},
  edgeTypes = {},
  handleAutoLayout = undefined,
  onSave = undefined,
  nodes,
  onNodesChange,
  edges,
  onEdgesChange,
  onConnect,
  onNodesDelete,
  onEdgesDelete,
  onNodeDragStop,
}: VisualFlowPropsInterface): ReactElement {
  const {mode, systemMode} = useColorScheme();
  const {t} = useTranslation();
  const {edgeStyle} = useFlowBuilderCore();
  const {anchorEl, handleClick, handleClose} = useEdgeStyleSelector();

  // Determine the effective color mode for React Flow
  const colorMode = mode === 'system' ? systemMode : mode;

  const handleSaveClick = useCallback(() => {
    onSave?.();
  }, [onSave]);

  return (
    <ReactFlow
      fitView
      nodes={nodes}
      edges={edges}
      nodeTypes={nodeTypes}
      edgeTypes={edgeTypes}
      onNodesChange={onNodesChange}
      onEdgesChange={onEdgesChange}
      onConnect={onConnect}
      onNodesDelete={onNodesDelete}
      onEdgesDelete={onEdgesDelete}
      onNodeDragStop={onNodeDragStop}
      proOptions={{hideAttribution: true}}
      colorMode={colorMode}
      minZoom={0.2}
      maxZoom={4}
    >
      {/* Center controls - auto-layout and edge style */}
      <Controls position="top-center" orientation="horizontal">
        {handleAutoLayout && (
          <Tooltip title={t('flows:core.headerPanel.autoLayout')}>
            <ControlButton
              onClick={handleAutoLayout}
              className="custom-control-button"
              aria-label={t('flows:core.headerPanel.autoLayout')}
            >
              <LayoutGrid size={20} />
            </ControlButton>
          </Tooltip>
        )}
        <Tooltip title={t('flows:core.headerPanel.edgeStyleTooltip')}>
          <ControlButton
            onClick={handleClick}
            className="custom-control-button"
            aria-label={t('flows:core.headerPanel.edgeStyleTooltip')}
          >
            {getEdgeStyleIcon(edgeStyle)}
          </ControlButton>
        </Tooltip>
      </Controls>

      {/* Save button - top right overlay */}
      <Card
        elevation={3}
        sx={{
          position: 'absolute',
          top: 5,
          right: 10,
          zIndex: 5,
          borderRadius: 1,
        }}
      >
        <Tooltip title={t('flows:core.headerPanel.save')}>
          <Button variant="contained" startIcon={<Save size={18} />} onClick={handleSaveClick}>
            {t('flows:core.headerPanel.save')}
          </Button>
        </Tooltip>
      </Card>

      <Background gap={20} />
      <CanvasValidationIndicator />
      <EdgeStyleMenu anchorEl={anchorEl} onClose={handleClose} />
    </ReactFlow>
  );
}

export default VisualFlow;
