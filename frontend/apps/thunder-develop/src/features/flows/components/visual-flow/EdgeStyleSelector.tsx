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

import {type ReactElement, useCallback} from 'react';
import {Menu, MenuItem, ListItemIcon, ListItemText} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import {EdgeStyleTypes, type EdgeStyleTypes as EdgeStyleTypesType} from '../../models/steps';
import {BezierEdgeIcon, SmoothStepEdgeIcon, StepEdgeIcon} from './EdgeStyleIcons';

/**
 * Props interface of {@link EdgeStyleMenu}
 */
export interface EdgeStyleMenuPropsInterface {
  /**
   * The anchor element for the menu.
   */
  anchorEl: HTMLElement | null;
  /**
   * Callback to close the menu.
   */
  onClose: () => void;
}

/**
 * Edge style menu component that displays edge style options.
 * This is a controlled component that requires anchorEl and onClose props.
 *
 * @param props - Props injected to the component.
 * @returns The EdgeStyleMenu component.
 */
function EdgeStyleMenu({anchorEl, onClose}: EdgeStyleMenuPropsInterface): ReactElement {
  const {t} = useTranslation();
  const {edgeStyle, setEdgeStyle} = useFlowBuilderCore();
  const open = Boolean(anchorEl);

  const handleStyleSelect = useCallback(
    (style: EdgeStyleTypesType) => {
      setEdgeStyle(style);
      onClose();
    },
    [setEdgeStyle, onClose],
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
    <Menu
      id="edge-style-menu"
      anchorEl={anchorEl}
      open={open}
      onClose={onClose}
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
  );
}

export default EdgeStyleMenu;
