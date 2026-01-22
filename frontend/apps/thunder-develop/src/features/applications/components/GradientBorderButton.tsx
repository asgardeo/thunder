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

import {forwardRef} from 'react';
import {styled, keyframes} from '@mui/material/styles';
import {Button} from '@wso2/oxygen-ui';
import type {ButtonProps} from '@wso2/oxygen-ui';

const rainbow = keyframes`
  0% {
    background-position: 0% 50%;
  }
  100% {
    background-position: 200% 50%;
  }
`;

const StyledGradientButton = styled(Button)(() => ({
  position: 'relative',
  display: 'inline-flex',
  borderRadius: '8px',
  padding: '8px 16px',
  border: '2px solid transparent',
  background: 'transparent',
  color: 'var(--mui-palette-text-primary)',
  fontWeight: 500,
  backgroundClip: 'padding-box',
  isolation: 'isolate',
  '&::before': {
    content: '""',
    position: 'absolute',
    inset: 0,
    borderRadius: '8px',
    padding: '2px',
    background: 'linear-gradient(135deg, #667eea, #764ba2, #f093fb, #4facfe, #00f2fe, #667eea)',
    backgroundSize: '300% 300%',
    WebkitMask: 'linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0)',
    WebkitMaskComposite: 'xor',
    maskComposite: 'exclude',
    animation: `${rainbow} 4s linear infinite`,
    zIndex: -1,
  },
  '&:hover': {
    background: 'var(--mui-palette-action-hover)',
    '&::before': {
      animationPlayState: 'paused',
    },
  },
  '&.Mui-disabled': {
    '&::before': {
      animationPlayState: 'paused',
      opacity: 0.6,
    },
  },
}));

/**
 * A button component with an animated gradient border effect.
 * Based on the MUI documentation "Edit in Chat" button implementation.
 */
const GradientBorderButton = forwardRef<HTMLButtonElement, ButtonProps>((props, ref) => (
  <StyledGradientButton ref={ref} variant="text" disableRipple disableFocusRipple {...props} />
));

GradientBorderButton.displayName = 'GradientBorderButton';

export default GradientBorderButton;
