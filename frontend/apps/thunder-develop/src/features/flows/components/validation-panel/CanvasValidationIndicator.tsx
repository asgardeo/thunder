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

import {type ReactElement, useMemo} from 'react';
import {Panel} from '@xyflow/react';
import {useTranslation} from 'react-i18next';
import {Box, Typography, Tooltip, ButtonBase} from '@wso2/oxygen-ui';
import classNames from 'classnames';
import {CircleXIcon, TriangleAlertIcon, InfoIcon} from '@wso2/oxygen-ui-icons-react';
import useValidationStatus from '../../hooks/useValidationStatus';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import Notification, {NotificationType} from '../../models/notification';
import './CanvasValidationIndicator.scss';

/**
 * Floating validation indicator component that appears on the canvas.
 * Shows error/warning/info counts and opens the validation panel when clicked.
 *
 * @returns The CanvasValidationIndicator component.
 */
function CanvasValidationIndicator(): ReactElement | null {
  const {notifications, setCurrentActiveTab, openValidationPanel, setOpenValidationPanel} = useValidationStatus();
  const {setIsOpenResourcePropertiesPanel, isResourcePropertiesPanelOpen} = useFlowBuilderCore();
  const {t} = useTranslation();

  // Calculate all notification counts in a single pass for better performance
  const {errorCount, warningCount, infoCount, totalCount} = useMemo(() => {
    let errors = 0;
    let warnings = 0;
    let info = 0;

    notifications?.forEach((notification: Notification) => {
      const type = notification.getType();

      if (type === NotificationType.ERROR) {
        errors += 1;
      } else if (type === NotificationType.WARNING) {
        warnings += 1;
      } else if (type === NotificationType.INFO) {
        info += 1;
      }
    });

    return {
      errorCount: errors,
      warningCount: warnings,
      infoCount: info,
      totalCount: errors + warnings + info,
    };
  }, [notifications]);

  // Don't render if there are no notifications
  if (totalCount === 0) {
    return null;
  }

  const handleClick = (tabIndex: number): void => {
    if (openValidationPanel) {
      setOpenValidationPanel?.(false);
      return;
    }

    setCurrentActiveTab?.(tabIndex);
    setIsOpenResourcePropertiesPanel(false);
    setOpenValidationPanel?.(true);
  };

  // Determine primary color based on highest priority notification
  const getPrimaryStatus = (): 'error' | 'warning' | 'info' => {
    if (errorCount > 0) return 'error';
    if (warningCount > 0) return 'warning';
    return 'info';
  };

  const primaryStatus = getPrimaryStatus();

  // Determine if the indicator should be shifted left due to an open panel
  const isPanelOpen = (openValidationPanel ?? false) || (isResourcePropertiesPanelOpen ?? false);

  return (
    <Panel
      position="top-right"
      className={classNames('canvas-validation-indicator', {
        'canvas-validation-indicator--panel-open': isPanelOpen,
      })}
    >
      <Box
        className={`canvas-validation-indicator__button canvas-validation-indicator__button--${primaryStatus}`}
        role="group"
        aria-label={t('flows:core.notificationPanel.trigger.label')}
      >
        {errorCount > 0 && (
          <Tooltip title={t('flows:core.notificationPanel.tabs.errors')} placement="bottom">
            <ButtonBase
              className="canvas-validation-indicator__item canvas-validation-indicator__item--error"
              onClick={() => handleClick(0)}
              aria-label={`${errorCount} ${t('flows:core.notificationPanel.tabs.errors')}`}
            >
              <CircleXIcon className="canvas-validation-indicator__icon" />
              <Typography className="canvas-validation-indicator__count">{errorCount}</Typography>
            </ButtonBase>
          </Tooltip>
        )}
        {warningCount > 0 && (
          <Tooltip title={t('flows:core.notificationPanel.tabs.warnings')} placement="bottom">
            <ButtonBase
              className="canvas-validation-indicator__item canvas-validation-indicator__item--warning"
              onClick={() => handleClick(1)}
              aria-label={`${warningCount} ${t('flows:core.notificationPanel.tabs.warnings')}`}
            >
              <TriangleAlertIcon className="canvas-validation-indicator__icon" />
              <Typography className="canvas-validation-indicator__count">{warningCount}</Typography>
            </ButtonBase>
          </Tooltip>
        )}
        {infoCount > 0 && (
          <Tooltip title={t('flows:core.notificationPanel.tabs.info')} placement="bottom">
            <ButtonBase
              className="canvas-validation-indicator__item canvas-validation-indicator__item--info"
              onClick={() => handleClick(2)}
              aria-label={`${infoCount} ${t('flows:core.notificationPanel.tabs.info')}`}
            >
              <InfoIcon className="canvas-validation-indicator__icon" />
              <Typography className="canvas-validation-indicator__count">{infoCount}</Typography>
            </ButtonBase>
          </Tooltip>
        )}
      </Box>
    </Panel>
  );
}

export default CanvasValidationIndicator;
