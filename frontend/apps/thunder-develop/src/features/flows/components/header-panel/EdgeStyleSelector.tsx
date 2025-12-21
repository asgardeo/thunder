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

import {type ReactElement, useState, useCallback} from 'react';
import {IconButton, Menu, MenuItem, ListItemIcon, ListItemText, Tooltip} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import {EdgeStyleTypes, type EdgeStyleTypes as EdgeStyleTypesType} from '../../models/steps';

/**
 * Icon component for Bezier edge style
 */
function BezierEdgeIcon(): ReactElement {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M4 12 C 8 4, 16 20, 20 12" strokeLinecap="round" />
    </svg>
  );
}

/**
 * Icon component for Smooth Step edge style
 */
function SmoothStepEdgeIcon(): ReactElement {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M4 6 H 10 Q 12 6, 12 8 V 16 Q 12 18, 14 18 H 20" strokeLinecap="round" />
    </svg>
  );
}

/**
 * Icon component for Step edge style
 */
function StepEdgeIcon(): ReactElement {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M4 6 H 12 V 18 H 20" strokeLinecap="round" strokeLinejoin="miter" />
    </svg>
  );
}

/**
 * Returns the appropriate icon component for the given edge style
 */
function getEdgeStyleIcon(style: EdgeStyleTypesType): ReactElement {
  switch (style) {
    case EdgeStyleTypes.Bezier:
      return <BezierEdgeIcon />;
    case EdgeStyleTypes.SmoothStep:
      return <SmoothStepEdgeIcon />;
    case EdgeStyleTypes.Step:
      return <StepEdgeIcon />;
    default:
      return <SmoothStepEdgeIcon />;
  }
}

/**
 * Edge style selector component for the header panel.
 * Allows users to switch between different edge rendering styles.
 *
 * @returns The EdgeStyleSelector component.
 */
function EdgeStyleSelector(): ReactElement {
  const {t} = useTranslation();
  const {edgeStyle, setEdgeStyle} = useFlowBuilderCore();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const handleClick = useCallback((event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  }, []);

  const handleClose = useCallback(() => {
    setAnchorEl(null);
  }, []);

  const handleStyleSelect = useCallback(
    (style: EdgeStyleTypesType) => {
      setEdgeStyle(style);
      handleClose();
    },
    [setEdgeStyle, handleClose],
  );

  const edgeStyleOptions: {value: EdgeStyleTypesType; label: string; icon: ReactElement}[] = [
    {
      value: EdgeStyleTypes.Bezier,
      label: t('flows:core.headerPanel.edgeStyles.bezier'),
      icon: <BezierEdgeIcon />,
    },
    {
      value: EdgeStyleTypes.SmoothStep,
      label: t('flows:core.headerPanel.edgeStyles.smoothStep'),
      icon: <SmoothStepEdgeIcon />,
    },
    {
      value: EdgeStyleTypes.Step,
      label: t('flows:core.headerPanel.edgeStyles.step'),
      icon: <StepEdgeIcon />,
    },
  ];

  return (
    <>
      <Tooltip title={t('flows:core.headerPanel.edgeStyleTooltip')}>
        <IconButton
          size="small"
          onClick={handleClick}
          aria-controls={open ? 'edge-style-menu' : undefined}
          aria-haspopup="true"
          aria-expanded={open ? 'true' : undefined}
        >
          {getEdgeStyleIcon(edgeStyle)}
        </IconButton>
      </Tooltip>
      <Menu
        id="edge-style-menu"
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'right',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}
      >
        {edgeStyleOptions.map((option) => (
          <MenuItem
            key={option.value}
            onClick={() => handleStyleSelect(option.value)}
            selected={edgeStyle === option.value}
          >
            <ListItemIcon>{option.icon}</ListItemIcon>
            <ListItemText>{option.label}</ListItemText>
          </MenuItem>
        ))}
      </Menu>
    </>
  );
}

export default EdgeStyleSelector;
