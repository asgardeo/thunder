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

import type {ThemeListItem} from '@thunderid/design';
import {Box} from '@wso2/oxygen-ui';
import type {JSX} from 'react';

export interface ThemeThumbnailProps {
  theme: ThemeListItem;
}

/**
 * Derives a deterministic primary color from a name string when no real
 * primaryColor is available from the API.
 */
function primaryFromName(name: string): string {
  let h1 = 5381;
  let h2 = 1779033703;
  for (let i = 0; i < name.length; i += 1) {
    const c = name.charCodeAt(i);
    h1 = (((h1 << 5) + h1) ^ c) >>> 0;
    h2 = Math.imul(h2 ^ c, 0x9e3779b9) >>> 0;
  }
  const hue = h1 % 360;
  const sat = 52 + ((h1 >> 4) % 20);
  const lig = 40 + ((h2 >> 12) % 12);

  return `hsl(${hue}, ${sat}%, ${lig}%)`;
}

export default function ThemeThumbnail({theme}: ThemeThumbnailProps): JSX.Element {
  const isDark = theme.defaultColorScheme === 'dark';
  const primary = String(theme.primaryColor ?? primaryFromName(theme.displayName ?? ''));

  const bg = isDark ? '#1e1e1e' : '#f0ede8';
  const surface = isDark ? '#2a2a2a' : '#ffffff';
  const line = isDark ? 'rgba(255,255,255,0.14)' : 'rgba(0,0,0,0.10)';
  const lineShort = isDark ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.06)';
  const divider = isDark ? 'rgba(255,255,255,0.07)' : 'rgba(0,0,0,0.06)';

  return (
    <Box
      sx={{
        width: '100%',
        height: '100%',
        bgcolor: bg,
        borderRadius: 'inherit',
        overflow: 'hidden',
        p: 1.75,
        display: 'flex',
        flexDirection: 'column',
        gap: 1.25,
        boxSizing: 'border-box',
      }}
    >
      {/* Header: avatar + title lines */}
      <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
        <Box
          sx={{
            width: 18,
            height: 18,
            borderRadius: '50%',
            bgcolor: primary,
            flexShrink: 0,
          }}
        />
        <Box sx={{display: 'flex', flexDirection: 'column', gap: 0.55, flex: 1}}>
          <Box sx={{height: 4, width: '55%', bgcolor: line, borderRadius: 0.5}} />
          <Box sx={{height: 3, width: '35%', bgcolor: lineShort, borderRadius: 0.5}} />
        </Box>
      </Box>

      {/* Divider */}
      <Box sx={{height: '1px', bgcolor: divider, mx: -1.75}} />

      {/* Content lines */}
      <Box sx={{display: 'flex', flexDirection: 'column', gap: 0.65, flex: 1}}>
        <Box sx={{height: 4, width: '80%', bgcolor: line, borderRadius: 0.5}} />
        <Box sx={{height: 4, width: '65%', bgcolor: lineShort, borderRadius: 0.5}} />
        <Box sx={{height: 4, width: '72%', bgcolor: lineShort, borderRadius: 0.5, opacity: 0.7}} />
      </Box>

      {/* Bottom: surface card with overlapping primary circles */}
      <Box
        sx={{
          bgcolor: surface,
          borderRadius: 1.5,
          px: 1.25,
          py: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        {/* Mini text lines inside the card */}
        <Box sx={{display: 'flex', flexDirection: 'column', gap: 0.5}}>
          <Box sx={{height: 3, width: 32, bgcolor: line, borderRadius: 0.5}} />
          <Box sx={{height: 3, width: 22, bgcolor: lineShort, borderRadius: 0.5}} />
        </Box>

        {/* Two overlapping primary-color circles */}
        <Box sx={{display: 'flex', alignItems: 'center', flexShrink: 0}}>
          <Box
            sx={{
              width: 22,
              height: 22,
              borderRadius: '50%',
              bgcolor: primary,
              opacity: 0.65,
              zIndex: 1,
            }}
          />
          <Box
            sx={{
              width: 22,
              height: 22,
              borderRadius: '50%',
              bgcolor: primary,
              ml: '-9px',
              zIndex: 2,
            }}
          />
        </Box>
      </Box>
    </Box>
  );
}
