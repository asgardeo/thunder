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

import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, Button, Alert} from '@wso2/oxygen-ui';

/**
 * Type of component being dropped that requires container(s)
 */
export type DropScenario = 'form-on-canvas' | 'input-on-canvas' | 'input-on-view' | 'widget-on-canvas';

/**
 * i18n key mapping for each scenario
 */
const SCENARIO_I18N_KEYS: Record<
  DropScenario,
  {
    title: string;
    description: string;
    alertMessage: string;
    confirmButton: string;
  }
> = {
  'form-on-canvas': {
    title: 'flows:core.dialogs.formRequiresView.formOnCanvas.title',
    description: 'flows:core.dialogs.formRequiresView.formOnCanvas.description',
    alertMessage: 'flows:core.dialogs.formRequiresView.formOnCanvas.alertMessage',
    confirmButton: 'flows:core.dialogs.formRequiresView.formOnCanvas.confirmButton',
  },
  'input-on-canvas': {
    title: 'flows:core.dialogs.formRequiresView.inputOnCanvas.title',
    description: 'flows:core.dialogs.formRequiresView.inputOnCanvas.description',
    alertMessage: 'flows:core.dialogs.formRequiresView.inputOnCanvas.alertMessage',
    confirmButton: 'flows:core.dialogs.formRequiresView.inputOnCanvas.confirmButton',
  },
  'input-on-view': {
    title: 'flows:core.dialogs.formRequiresView.inputOnView.title',
    description: 'flows:core.dialogs.formRequiresView.inputOnView.description',
    alertMessage: 'flows:core.dialogs.formRequiresView.inputOnView.alertMessage',
    confirmButton: 'flows:core.dialogs.formRequiresView.inputOnView.confirmButton',
  },
  'widget-on-canvas': {
    title: 'flows:core.dialogs.formRequiresView.widgetOnCanvas.title',
    description: 'flows:core.dialogs.formRequiresView.widgetOnCanvas.description',
    alertMessage: 'flows:core.dialogs.formRequiresView.widgetOnCanvas.alertMessage',
    confirmButton: 'flows:core.dialogs.formRequiresView.widgetOnCanvas.confirmButton',
  },
};

export interface FormRequiresViewDialogProps {
  /**
   * Whether the dialog is open
   */
  open: boolean;
  /**
   * The scenario that triggered the dialog
   */
  scenario: DropScenario;
  /**
   * Callback when the dialog should be closed
   */
  onClose: () => void;
  /**
   * Callback when the user confirms adding the required containers
   */
  onConfirm: () => void;
}

/**
 * Dialog component informing users about component container requirements.
 * Offers the option to automatically create the required container hierarchy.
 */
export default function FormRequiresViewDialog({
  open,
  scenario,
  onClose,
  onConfirm,
}: FormRequiresViewDialogProps): JSX.Element {
  const {t} = useTranslation();
  const i18nKeys = SCENARIO_I18N_KEYS[scenario];

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t(i18nKeys.title)}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 2}}>{t(i18nKeys.description)}</DialogContentText>
        <Alert severity="info" sx={{mb: 2}}>
          {t(i18nKeys.alertMessage)}
        </Alert>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{t('flows:core.dialogs.formRequiresView.cancelButton')}</Button>
        <Button onClick={onConfirm} color="primary" variant="contained">
          {t(i18nKeys.confirmButton)}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
