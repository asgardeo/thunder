/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import {Box} from '@wso2/oxygen-ui';
import type {ThemeListItem} from '@thunder/shared-design';
import type {JSX} from 'react';

export interface ThemeThumbnailProps {
  theme: ThemeListItem;
}

export default function ThemeThumbnail({theme}: ThemeThumbnailProps): JSX.Element {
  const light = theme.theme?.colorSchemes?.light?.colors;
  const primary = light?.primary?.main ?? '#1976d2';
  const bg = light?.background?.default ?? '#f5f5f5';
  const paper = light?.background?.paper ?? '#ffffff';
  const text = light?.text?.primary ?? '#000000';
  const radius = (() => {
    const r = theme.theme?.shape?.borderRadius;
    if (typeof r === 'number') return r;
    if (typeof r === 'string') return parseInt(r, 10) || 8;
    return 8;
  })();
  const rBtn = Math.max(radius * 0.45, 2);
  const rCard = Math.max(radius * 0.55, 3);

  return (
    <Box sx={{width: '100%', height: '100%', bgcolor: bg, display: 'flex', flexDirection: 'column'}}>
      {/* Browser chrome */}
      <Box
        sx={{
          height: 20,
          bgcolor: 'rgba(0,0,0,0.07)',
          display: 'flex',
          alignItems: 'center',
          px: 1,
          gap: 0.5,
          flexShrink: 0,
        }}
      >
        <Box sx={{width: 5, height: 5, borderRadius: '50%', bgcolor: '#ff5f57'}} />
        <Box sx={{width: 5, height: 5, borderRadius: '50%', bgcolor: '#febc2e'}} />
        <Box sx={{width: 5, height: 5, borderRadius: '50%', bgcolor: '#28c840'}} />
        <Box sx={{flex: 1, height: 8, bgcolor: 'rgba(0,0,0,0.05)', borderRadius: '4px', mx: 1}} />
      </Box>
      {/* Page content */}
      <Box sx={{flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', p: 1.5}}>
        <Box
          sx={{
            width: '68%',
            bgcolor: paper,
            borderRadius: `${rCard}px`,
            boxShadow: '0 2px 12px rgba(0,0,0,0.12)',
            px: 1.5,
            py: 1.25,
            display: 'flex',
            flexDirection: 'column',
            gap: 0.75,
          }}
        >
          <Box sx={{width: 16, height: 16, borderRadius: '3px', bgcolor: primary}} />
          <Box sx={{height: 5, width: '55%', bgcolor: text, borderRadius: 0.5, opacity: 0.8}} />
          <Box sx={{height: 3, width: '80%', bgcolor: text, borderRadius: 0.5, opacity: 0.15}} />
          <Box sx={{height: 9, border: '1px solid rgba(0,0,0,0.12)', borderRadius: `${rBtn}px`, bgcolor: bg}} />
          <Box sx={{height: 9, border: '1px solid rgba(0,0,0,0.12)', borderRadius: `${rBtn}px`, bgcolor: bg}} />
          <Box sx={{height: 9, bgcolor: primary, borderRadius: `${rBtn}px`, mt: 0.25}} />
          <Box sx={{height: 3, width: '48%', bgcolor: primary, borderRadius: 0.5, opacity: 0.55, mx: 'auto'}} />
        </Box>
      </Box>
    </Box>
  );
}
