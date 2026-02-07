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
import {Box, Paper, Stack} from '@wso2/oxygen-ui';
import type {LayoutConfig} from '../models/layout';

/**
 * Props for LayoutPreview component
 */
interface LayoutPreviewProps {
  layout: LayoutConfig;
  width?: number;
  height?: number;
}

/**
 * Preview component for layout builder
 */
export default function LayoutPreview({layout, width = 1440, height = 900}: LayoutPreviewProps): JSX.Element {
  const signinScreen = layout?.screens?.signin;
  const template = signinScreen?.template || 'centered';
  const container = signinScreen?.container || {};
  const background = signinScreen?.background || '#f5f5f5';

  const containerMaxWidth = container.maxWidth || 400;
  const containerPadding = container.padding || 24;
  const containerBorderRadius = container.borderRadius || '8px';
  const containerElevation = container.elevation || 2;

  return (
    <Box
      sx={{
        width: '100%',
        height: '100%',
        bgcolor: background,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        overflow: 'hidden',
      }}
    >
      {template === 'centered' && (
        <Paper
          elevation={containerElevation}
          sx={{
            maxWidth: containerMaxWidth,
            width: '100%',
            p: `${containerPadding}px`,
            borderRadius: containerBorderRadius,
            mx: 4,
          }}
        >
          <LayoutContent />
        </Paper>
      )}

      {template === 'split' && (
        <Box sx={{display: 'flex', width: '100%', height: '100%'}}>
          <Box
            sx={{
              flex: 1,
              bgcolor: 'primary.main',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Box sx={{width: 200, height: 200, bgcolor: 'rgba(255,255,255,0.2)', borderRadius: 2}} />
          </Box>
          <Box
            sx={{
              flex: 1,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              p: 4,
            }}
          >
            <Paper
              elevation={containerElevation}
              sx={{
                maxWidth: containerMaxWidth,
                width: '100%',
                p: `${containerPadding}px`,
                borderRadius: containerBorderRadius,
              }}
            >
              <LayoutContent />
            </Paper>
          </Box>
        </Box>
      )}

      {template === 'fullscreen' && (
        <Box
          sx={{
            width: '100%',
            height: '100%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            p: 4,
          }}
        >
          <Box sx={{maxWidth: containerMaxWidth, width: '100%'}}>
            <LayoutContent />
          </Box>
        </Box>
      )}
    </Box>
  );
}

/**
 * Placeholder content for layout preview
 */
function LayoutContent(): JSX.Element {
  return (
    <Stack spacing={2}>
      {/* Title */}
      <Box sx={{width: '60%', height: 32, bgcolor: 'grey.300', borderRadius: 1}} />
      {/* Subtitle */}
      <Box sx={{width: '40%', height: 16, bgcolor: 'grey.200', borderRadius: 1}} />
      {/* Input field 1 */}
      <Box sx={{width: '100%', height: 40, bgcolor: 'grey.100', borderRadius: 1}} />
      {/* Input field 2 */}
      <Box sx={{width: '100%', height: 40, bgcolor: 'grey.100', borderRadius: 1}} />
      {/* Button */}
      <Box sx={{width: '100%', height: 40, bgcolor: 'primary.main', borderRadius: 1, opacity: 0.7}} />
      {/* Footer text */}
      <Box sx={{width: '50%', height: 12, bgcolor: 'grey.200', borderRadius: 1, mx: 'auto', mt: 2}} />
    </Stack>
  );
}
