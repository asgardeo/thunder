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

import {alpha} from '@mui/material/styles';
import type {Theme, Components} from '@mui/material/styles';
import {gray, orange, red, green, blue} from '../themePrimitives';

const feedbackCustomizations: Components<Theme> = {
  MuiAlert: {
    styleOverrides: {
      root: ({theme, ownerState}) => ({
        borderRadius: 10,
        color: (theme.vars ?? theme).palette.text.primary,
        ...(ownerState.severity === 'error' && {
          backgroundColor: red[100],
          border: `1px solid ${alpha(red[300], 0.5)}`,
          '& .MuiAlert-icon': {
            color: red[500],
          },
          ...theme.applyStyles('dark', {
            backgroundColor: alpha(red[900], 0.5),
            border: `1px solid ${alpha(red[800], 0.5)}`,
          }),
        }),
        ...(ownerState.severity === 'warning' && {
          backgroundColor: orange[100],
          border: `1px solid ${alpha(orange[300], 0.5)}`,
          '& .MuiAlert-icon': {
            color: orange[500],
          },
          ...theme.applyStyles('dark', {
            backgroundColor: alpha(orange[900], 0.5),
            border: `1px solid ${alpha(orange[800], 0.5)}`,
          }),
        }),
        ...(ownerState.severity === 'info' && {
          backgroundColor: blue[100],
          border: `1px solid ${alpha(blue[300], 0.5)}`,
          '& .MuiAlert-icon': {
            color: blue[500],
          },
          ...theme.applyStyles('dark', {
            backgroundColor: alpha(blue[900], 0.5),
            border: `1px solid ${alpha(blue[800], 0.5)}`,
          }),
        }),
        ...(ownerState.severity === 'success' && {
          backgroundColor: green[100],
          border: `1px solid ${alpha(green[300], 0.5)}`,
          '& .MuiAlert-icon': {
            color: green[500],
          },
          ...theme.applyStyles('dark', {
            backgroundColor: alpha(green[900], 0.5),
            border: `1px solid ${alpha(green[800], 0.5)}`,
          }),
        }),
      }),
    },
  },
  MuiDialog: {
    styleOverrides: {
      root: ({theme}) => ({
        '& .MuiDialog-paper': {
          borderRadius: '10px',
          border: '1px solid',
          borderColor: (theme.vars ?? theme).palette.divider,
        },
      }),
    },
  },
  MuiLinearProgress: {
    styleOverrides: {
      root: ({theme}) => ({
        height: 8,
        borderRadius: 8,
        backgroundColor: gray[200],
        ...theme.applyStyles('dark', {
          backgroundColor: gray[800],
        }),
      }),
    },
  },
};

export default feedbackCustomizations;
