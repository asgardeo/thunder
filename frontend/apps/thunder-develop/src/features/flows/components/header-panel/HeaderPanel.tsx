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

import {type HTMLAttributes, type ReactElement, useContext, useState} from 'react';
import {useNavigate} from 'react-router';
import {
  Box,
  Button,
  Card,
  type CardProps,
  ColorSchemeToggle,
  IconButton,
  Stack,
  Typography,
  Tooltip,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
} from '@wso2/oxygen-ui';
import {ArrowLeftIcon, HistoryIcon, SaveIcon, LayoutTemplate, Eye, EyeOff, Spline, Check, ShieldCheck, ShieldOff} from '@wso2/oxygen-ui-icons-react';
import FlowBuilderCoreContext from '../../context/FlowBuilderCoreContext';
import ValidationStatusLabels from '../validation-panel/ValidationStatusLabels';
import {EdgeStyleTypes} from '../../models/steps';

/**
 * Props interface of {@link HeaderPanel}
 */
export interface HeaderPanelPropsInterface extends Omit<CardProps, 'open' | 'title'>, HTMLAttributes<HTMLDivElement> {
  /**
   * Title to display in the header.
   */
  title?: string;
  /**
   * Callback to be triggered when back button is clicked.
   */
  onBack?: () => void;
  /**
   * Callback to be triggered when save button is clicked.
   */
  onSave?: () => void;
  /**
   * Callback to be triggered when auto-layout button is clicked.
   */
  onAutoLayout?: () => void;
}

/**
 * Flow builder header panel that appears at the top.
 *
 * @param props - Props injected to the component.
 * @returns The HeaderPanel component.
 */
function HeaderPanel({
  title = 'Login Flow',
  onBack = undefined,
  onSave = undefined,
  onAutoLayout = undefined,
  ...rest
}: HeaderPanelPropsInterface): ReactElement {
  const navigate = useNavigate();
  const {isVerboseMode, setIsVerboseMode, edgeStyle, setEdgeStyle, isCollisionAvoidanceEnabled, setIsCollisionAvoidanceEnabled} = useContext(FlowBuilderCoreContext);
  const [edgeStyleMenuAnchor, setEdgeStyleMenuAnchor] = useState<null | HTMLElement>(null);
  const isEdgeStyleMenuOpen = Boolean(edgeStyleMenuAnchor);

  const handleEdgeStyleMenuOpen = (event: React.MouseEvent<HTMLButtonElement>): void => {
    setEdgeStyleMenuAnchor(event.currentTarget);
  };

  const handleEdgeStyleMenuClose = (): void => {
    setEdgeStyleMenuAnchor(null);
  };

  const handleEdgeStyleChange = (style: typeof edgeStyle): void => {
    setEdgeStyle(style);
    handleEdgeStyleMenuClose();
  };

  const getEdgeStyleLabel = (): string => {
    switch (edgeStyle) {
      case EdgeStyleTypes.Bezier:
        return 'Bézier';
      case EdgeStyleTypes.SmoothStep:
        return 'Smooth Step';
      case EdgeStyleTypes.Step:
        return 'Step';
      default:
        return 'Edge Style';
    }
  };

  const handleBackClick = () => {
    if (onBack) {
      onBack();
      return;
    }

    // Go back to the flows list
    // eslint-disable-next-line @typescript-eslint/no-floating-promises
    navigate('/flows');
  };

  const handleSaveClick = () => {
    // Note: getNodes/getEdges returns visually displayed nodes (may be filtered)
    // The actual save handler uses the full unfiltered data
    onSave?.();
  };

  return (
    <Box sx={{flexShrink: 0, pb: 1}} {...rest}>
      <Card
        elevation={0}
        sx={{
          borderRadius: 1,
          height: 50,
          display: 'flex',
          alignItems: 'center',
          px: 2,
        }}
      >
        <Stack direction="row" alignItems="center" spacing={2} width="100%">
          {/* Left section - Back button and title */}
          <Stack direction="row" alignItems="center">
            <Button onClick={handleBackClick} startIcon={<ArrowLeftIcon size={20} />}>
              Go back to Flows
            </Button>
          </Stack>

          <Typography variant="h4">{title}</Typography>

          {/* Right section - Action buttons */}
          <Stack direction="row" alignItems="center" spacing={1} ml="auto">
            <ValidationStatusLabels />
            <Tooltip title="History">
              <IconButton size="small">
                <HistoryIcon size={20} />
              </IconButton>
            </Tooltip>
            {onAutoLayout && (
              <Tooltip title="Auto Layout">
                <IconButton size="small" onClick={onAutoLayout}>
                  <LayoutTemplate size={20} />
                </IconButton>
              </Tooltip>
            )}
            <Tooltip title={isVerboseMode ? 'Hide Executors' : 'Show Executors'}>
              <IconButton size="small" onClick={() => setIsVerboseMode(!isVerboseMode)}>
                {isVerboseMode ? <Eye size={20} /> : <EyeOff size={20} />}
              </IconButton>
            </Tooltip>
            <Tooltip title={`Edge Style: ${getEdgeStyleLabel()}`}>
              <IconButton size="small" onClick={handleEdgeStyleMenuOpen}>
                <Spline size={20} />
              </IconButton>
            </Tooltip>
            <Tooltip title={isCollisionAvoidanceEnabled ? 'Disable Collision Avoidance (Better Performance)' : 'Enable Collision Avoidance'}>
              <IconButton size="small" onClick={() => setIsCollisionAvoidanceEnabled(!isCollisionAvoidanceEnabled)}>
                {isCollisionAvoidanceEnabled ? <ShieldCheck size={20} /> : <ShieldOff size={20} />}
              </IconButton>
            </Tooltip>
            <Menu
              anchorEl={edgeStyleMenuAnchor}
              open={isEdgeStyleMenuOpen}
              onClose={handleEdgeStyleMenuClose}
              anchorOrigin={{
                vertical: 'bottom',
                horizontal: 'right',
              }}
              transformOrigin={{
                vertical: 'top',
                horizontal: 'right',
              }}
            >
              <MenuItem onClick={() => handleEdgeStyleChange(EdgeStyleTypes.Bezier)}>
                <ListItemIcon>
                  {edgeStyle === EdgeStyleTypes.Bezier && <Check size={16} />}
                </ListItemIcon>
                <ListItemText>Bézier (Curved)</ListItemText>
              </MenuItem>
              <MenuItem onClick={() => handleEdgeStyleChange(EdgeStyleTypes.SmoothStep)}>
                <ListItemIcon>
                  {edgeStyle === EdgeStyleTypes.SmoothStep && <Check size={16} />}
                </ListItemIcon>
                <ListItemText>Smooth Step</ListItemText>
              </MenuItem>
              <MenuItem onClick={() => handleEdgeStyleChange(EdgeStyleTypes.Step)}>
                <ListItemIcon>
                  {edgeStyle === EdgeStyleTypes.Step && <Check size={16} />}
                </ListItemIcon>
                <ListItemText>Step (Angular)</ListItemText>
              </MenuItem>
            </Menu>
            <ColorSchemeToggle data-testid="theme-toggle" />
            <Button variant="contained" startIcon={<SaveIcon size={20} />} onClick={handleSaveClick}>
              Save
            </Button>
          </Stack>
        </Stack>
      </Card>
    </Box>
  );
}

export default HeaderPanel;
