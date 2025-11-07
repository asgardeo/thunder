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
import {Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, Button, Alert} from '@wso2/oxygen-ui';

/**
 * Type of component being dropped that requires container(s)
 */
export type DropScenario = 'form-on-canvas' | 'input-on-canvas' | 'input-on-view' | 'widget-on-canvas';

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
 * Configuration for each drop scenario
 */
const scenarioConfig: Record<
  DropScenario,
  {
    title: string;
    description: string;
    alertMessage: string;
    confirmButtonText: string;
  }
> = {
  'form-on-canvas': {
    title: 'Form Requires a View',
    description: 'Form components cannot be placed directly on the canvas. They must be inside a View component.',
    alertMessage: 'Would you like to create a View and add the Form inside it?',
    confirmButtonText: 'Add View with Form',
  },
  'input-on-canvas': {
    title: 'Input Requires a Form and View',
    description:
      'Input components cannot be placed directly on the canvas. They must be inside a Form, which must be inside a View.',
    alertMessage: 'Would you like to create a View with a Form and add the Input inside it?',
    confirmButtonText: 'Add View, Form and Input',
  },
  'input-on-view': {
    title: 'Input Requires a Form',
    description: 'Input components cannot be placed directly inside a View. They must be inside a Form component.',
    alertMessage: 'Would you like to create a Form and add the Input inside it?',
    confirmButtonText: 'Add Form with Input',
  },
  'widget-on-canvas': {
    title: 'Widget Requires a View',
    description: 'Widgets cannot be placed directly on the canvas. They must be inside a View component.',
    alertMessage: 'Would you like to create a View and add the Widget inside it?',
    confirmButtonText: 'Add View with Widget',
  },
};

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
  const config = scenarioConfig[scenario];

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{config.title}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 2}}>{config.description}</DialogContentText>
        <Alert severity="info" sx={{mb: 2}}>
          {config.alertMessage}
        </Alert>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={onConfirm} color="primary" variant="contained">
          {config.confirmButtonText}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
