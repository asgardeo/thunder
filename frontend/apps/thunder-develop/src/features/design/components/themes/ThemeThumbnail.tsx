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

/**
 * Derives multiple independent values from a name string using two separate
 * hash functions (djb2 + LCG) so that similar names still produce visually
 * distinct results.
 */
function hashName(name: string): {h1: number; h2: number} {
  let h1 = 5381; // djb2 seed
  let h2 = 1779033703; // FNV-like seed
  for (let i = 0; i < name.length; i++) {
    const c = name.charCodeAt(i);
    h1 = (((h1 << 5) + h1) ^ c) >>> 0;
    h2 = Math.imul(h2 ^ c, 0x9e3779b9) >>> 0;
  }
  return {h1, h2};
}

/** Build a flat palette with varied hue, saturation and lightness. */
function paletteFromName(name: string) {
  const {h1, h2} = hashName(name);

  const hue = h1 % 360;
  // Saturation varies between 52% and 72%
  const sat = 52 + ((h1 >> 4) % 20);
  // Lightness for primary varies between 40% and 52%
  const lig = 40 + ((h2 >> 12) % 12);

  return {
    primary: `hsl(${hue}, ${sat}%, ${lig}%)`,
    primaryLight: `hsl(${hue}, ${sat}%, ${lig + 42}%)`,
    bg: `hsl(${hue}, ${Math.round(sat * 0.18)}%, 97%)`,
    paper: '#ffffff',
    text: `hsl(${hue}, 15%, 18%)`,
    textFaint: `hsl(${hue}, 10%, 55%)`,
    inputBorder: `hsl(${hue}, ${Math.round(sat * 0.18)}%, 87%)`,
    divider: `hsl(${hue}, ${Math.round(sat * 0.12)}%, 90%)`,
  };
}

/** Returns up to 2 uppercase initials from a display name. */
function initials(name: string): string {
  const words = name.trim().split(/\s+/).filter(Boolean);
  if (words.length === 0) return '?';
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase();
  return (words[0][0] + words[1][0]).toUpperCase();
}

export default function ThemeThumbnail({theme}: ThemeThumbnailProps): JSX.Element {
  const p = paletteFromName(theme.displayName ?? '');
  const abbr = initials(theme.displayName ?? '');

  const radius = (() => {
    const r = theme.theme?.shape?.borderRadius;
    if (typeof r === 'number') return r;
    if (typeof r === 'string') return parseInt(r, 10) || 8;
    return 8;
  })();
  const rCard = Math.min(Math.max(radius * 0.6, 4), 10);
  const rBtn = Math.min(Math.max(radius * 0.5, 3), 8);
  const rInput = Math.min(Math.max(radius * 0.4, 2), 6);

  return (
    <Box
      sx={{
        width: '100%',
        height: '100%',
        bgcolor: p.bg,
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      }}
    >
      <Box sx={{flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center'}}>
        <Box
          sx={{
            bgcolor: p.paper,
            borderRadius: `${rCard}px`,
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            width: '60%',
            px: 1.5,
            py: 1.5,
            gap: 0.6,
          }}
        >
          {/* Logo avatar */}
          <Box
            sx={{
              width: 20,
              height: 20,
              borderRadius: '50%',
              bgcolor: p.primary,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              mb: 0.25,
              flexShrink: 0,
            }}
          >
            <Box
              sx={{
                color: '#fff',
                fontSize: '7px',
                fontWeight: 700,
                lineHeight: 1,
                fontFamily: 'sans-serif',
                letterSpacing: '-0.3px',
              }}
            >
              {abbr}
            </Box>
          </Box>

          {/* Input 1 — with inner label line */}
          <Box
            sx={{
              width: '100%',
              height: 9,
              border: `1px solid ${p.inputBorder}`,
              borderRadius: `${rInput}px`,
              bgcolor: p.bg,
              px: 0.75,
              display: 'flex',
              alignItems: 'center',
            }}
          >
            <Box sx={{height: 3, width: '45%', bgcolor: p.textFaint, borderRadius: 0.5, opacity: 0.6}} />
          </Box>

          {/* Input 2 */}
          <Box
            sx={{
              width: '100%',
              height: 9,
              border: `1px solid ${p.inputBorder}`,
              borderRadius: `${rInput}px`,
              bgcolor: p.bg,
              px: 0.75,
              display: 'flex',
              alignItems: 'center',
              gap: 0.5,
            }}
          >
            <Box sx={{height: 3, width: '40%', bgcolor: p.textFaint, borderRadius: 0.5, opacity: 0.6}} />
            {/* Lock icon dots */}
            <Box sx={{display: 'flex', gap: 0.3, ml: 'auto'}}>
              {[0, 1, 2, 3].map((i) => (
                <Box key={i} sx={{width: 2, height: 2, borderRadius: '50%', bgcolor: p.textFaint, opacity: 0.5}} />
              ))}
            </Box>
          </Box>

          {/* Primary button — flat, no gradient */}
          <Box
            sx={{
              width: '100%',
              height: 9,
              bgcolor: p.primary,
              borderRadius: `${rBtn}px`,
              mt: 0.25,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Box sx={{height: 3, width: '35%', bgcolor: '#fff', borderRadius: 0.5, opacity: 0.9}} />
          </Box>

          {/* Divider */}
          <Box sx={{width: '100%', display: 'flex', alignItems: 'center', gap: 0.5, my: 0.1}}>
            <Box sx={{flex: 1, height: '1px', bgcolor: p.divider}} />
            <Box sx={{height: 3, width: 8, bgcolor: p.textFaint, borderRadius: 0.5, opacity: 0.4}} />
            <Box sx={{flex: 1, height: '1px', bgcolor: p.divider}} />
          </Box>

          {/* Social login placeholders */}
          <Box sx={{display: 'flex', gap: 0.75, justifyContent: 'center'}}>
            {[0, 1, 2].map((i) => (
              <Box
                key={i}
                sx={{
                  width: 14,
                  height: 9,
                  border: `1px solid ${p.inputBorder}`,
                  borderRadius: `${rInput}px`,
                  bgcolor: p.paper,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <Box sx={{width: 5, height: 5, borderRadius: '50%', bgcolor: p.primaryLight}} />
              </Box>
            ))}
          </Box>

          {/* Link */}
          <Box
            sx={{height: 3, width: '52%', bgcolor: p.primary, borderRadius: 0.5, opacity: 0.55, mx: 'auto', mt: 0.1}}
          />
        </Box>
      </Box>
    </Box>
  );
}
