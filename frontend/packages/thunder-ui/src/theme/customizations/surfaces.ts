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
import {gray} from '../themePrimitives';

const surfacesCustomizations: Components<Theme> = {
  MuiAccordion: {
    defaultProps: {
      elevation: 0,
      disableGutters: true,
    },
    styleOverrides: {
      root: ({theme}) => ({
        padding: 4,
        overflow: 'clip',
        backgroundColor: (theme.vars ?? theme).palette.background.default,
        border: '1px solid',
        borderColor: (theme.vars ?? theme).palette.divider,
        ':before': {
          backgroundColor: 'transparent',
        },
        '&:not(:last-of-type)': {
          borderBottom: 'none',
        },
        '&:first-of-type': {
          borderTopLeftRadius: (theme.vars ?? theme).shape.borderRadius,
          borderTopRightRadius: (theme.vars ?? theme).shape.borderRadius,
        },
        '&:last-of-type': {
          borderBottomLeftRadius: (theme.vars ?? theme).shape.borderRadius,
          borderBottomRightRadius: (theme.vars ?? theme).shape.borderRadius,
        },
      }),
    },
  },
  MuiAccordionSummary: {
    styleOverrides: {
      root: ({theme}) => ({
        border: 'none',
        borderRadius: 8,
        '&:hover': {backgroundColor: gray[50]},
        '&:focus-visible': {backgroundColor: 'transparent'},
        ...theme.applyStyles('dark', {
          '&:hover': {backgroundColor: gray[800]},
        }),
      }),
    },
  },
  MuiAccordionDetails: {
    styleOverrides: {
      root: {mb: 20, border: 'none'},
    },
  },
  MuiPaper: {
    defaultProps: {
      elevation: 0,
    },
  },
  MuiCard: {
    styleOverrides: {
      root: ({theme}) => ({
        padding: 16,
        gap: 16,
        transition: 'all 100ms ease',
        backgroundColor: gray[50],
        borderRadius: (theme.vars ?? theme).shape.borderRadius,
        border: `1px solid ${(theme.vars ?? theme).palette.divider}`,
        boxShadow: 'none',
        ...theme.applyStyles('dark', {
          backgroundColor: gray[800],
        }),
        variants: [
          {
            props: {
              variant: 'outlined',
            },
            style: {
              border: `1px solid ${(theme.vars ?? theme).palette.divider}`,
              boxShadow: 'none',
              background: 'hsl(0, 0%, 100%)',
              ...theme.applyStyles('dark', {
                background: alpha(gray[900], 0.4),
              }),
            },
          },
        ],
      }),
    },
  },
  MuiCardContent: {
    styleOverrides: {
      root: {
        padding: 0,
        '&:last-child': {paddingBottom: 0},
      },
    },
  },
  MuiCardHeader: {
    styleOverrides: {
      root: {
        padding: 0,
      },
    },
  },
  MuiCardActions: {
    styleOverrides: {
      root: {
        padding: 0,
      },
    },
  },
};

export default surfacesCustomizations;
