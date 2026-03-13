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

import type {JSX} from 'react';
import {Box, useColorScheme, useTheme} from '@wso2/oxygen-ui';
import type {LayoutListItem} from '@thunder/shared-design';

export interface LayoutThumbnailProps {
  layout: LayoutListItem;
}

export default function LayoutThumbnail({layout}: LayoutThumbnailProps): JSX.Element {
  const theme = useTheme();
  const {mode, systemMode} = useColorScheme();
  const isDark = (mode === 'system' ? systemMode : mode) === 'dark';

  const screens = layout.layout?.screens as Record<string, Record<string, unknown>> | undefined;
  const authScreen = screens?.auth;
  const hasHeader = !!(authScreen?.slots as Record<string, unknown> | undefined)?.header;
  const hasFooter = !!(authScreen?.slots as Record<string, unknown> | undefined)?.footer;
  const bg = authScreen?.background
    ? ((authScreen.background as Record<string, unknown>).value as string | undefined)
    : undefined;

  // When a custom background is set use a glass overlay; otherwise fall back to solid theme surfaces.
  const glassMode = !!bg;
  const outerCard = glassMode
    ? {
        bgcolor: isDark ? 'rgba(0,0,0,0.35)' : 'rgba(255,255,255,0.15)',
        backdropFilter: 'blur(4px)',
        border: isDark ? '1px solid rgba(255,255,255,0.12)' : '1px solid rgba(255,255,255,0.3)',
        boxShadow: '0 4px 20px rgba(0,0,0,0.25)',
      }
    : {
        bgcolor: theme.vars?.palette.background.paper,
        border: `1px solid ${theme.vars?.palette.divider}`,
        boxShadow: isDark ? '0 4px 20px rgba(0,0,0,0.5)' : '0 4px 20px rgba(0,0,0,0.1)',
      };

  let headerFooterBg: string | undefined;
  if (glassMode) {
    headerFooterBg = isDark ? 'rgba(0,0,0,0.5)' : 'rgba(255,255,255,0.88)';
  } else {
    headerFooterBg = theme.vars?.palette.background.paper;
  }

  let headerBorderColor: string | undefined;
  if (glassMode) {
    headerBorderColor = isDark ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)';
  } else {
    headerBorderColor = theme.vars?.palette.divider;
  }

  let navChipBg: string | undefined;
  if (glassMode) {
    navChipBg = isDark ? 'rgba(255,255,255,0.2)' : 'rgba(0,0,0,0.15)';
  } else {
    navChipBg = theme.vars?.palette.action.hover;
  }

  let formCardBg: string | undefined;
  if (glassMode) {
    formCardBg = isDark ? 'rgba(15,15,15,0.85)' : 'rgba(255,255,255,0.95)';
  } else {
    formCardBg = theme.vars?.palette.background.default;
  }

  const titleColor = theme.vars?.palette.text.primary;
  const inputBg = theme.vars?.palette.action.hover;
  const inputBorder = theme.vars?.palette.divider;

  return (
    <Box
      sx={{
        width: '100%',
        height: '100%',
        bgcolor: bg ? undefined : theme.vars?.palette.background.default,
        background: bg ?? undefined,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        p: 1.5,
      }}
    >
      <Box
        sx={{
          width: '82%',
          height: '82%',
          borderRadius: '8px',
          overflow: 'hidden',
          display: 'flex',
          flexDirection: 'column',
          ...outerCard,
        }}
      >
        {hasHeader && (
          <Box
            sx={{
              height: 18,
              bgcolor: headerFooterBg,
              borderBottom: `1px solid ${headerBorderColor}`,
              display: 'flex',
              alignItems: 'center',
              px: 1.25,
              gap: 0.75,
              flexShrink: 0,
            }}
          >
            <Box sx={{width: 10, height: 6, bgcolor: 'primary.main', borderRadius: '1.5px', opacity: 0.85}} />
            <Box sx={{flex: 1}} />
            <Box sx={{width: 14, height: 5, bgcolor: navChipBg, borderRadius: '3px'}} />
          </Box>
        )}
        <Box sx={{flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center'}}>
          <Box
            sx={{
              width: '55%',
              bgcolor: formCardBg,
              borderRadius: '5px',
              boxShadow: isDark ? '0 2px 10px rgba(0,0,0,0.4)' : '0 2px 10px rgba(0,0,0,0.12)',
              p: 1,
              display: 'flex',
              flexDirection: 'column',
              gap: 0.5,
            }}
          >
            <Box sx={{height: 5, width: '62%', bgcolor: titleColor, borderRadius: 0.5, mx: 'auto', opacity: 0.8}} />
            <Box sx={{height: 7, bgcolor: inputBg, borderRadius: '3px', border: `0.5px solid ${inputBorder}`}} />
            <Box sx={{height: 7, bgcolor: inputBg, borderRadius: '3px', border: `0.5px solid ${inputBorder}`}} />
            <Box sx={{height: 10, bgcolor: 'primary.main', borderRadius: '3px', opacity: 0.85}} />
          </Box>
        </Box>
        {hasFooter && (
          <Box
            sx={{
              height: 16,
              bgcolor: headerFooterBg,
              borderTop: `1px solid ${headerBorderColor}`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 1,
              flexShrink: 0,
            }}
          >
            {[1, 2, 3].map((i) => (
              <Box key={i} sx={{width: 14, height: 3.5, bgcolor: navChipBg, borderRadius: 0.5}} />
            ))}
          </Box>
        )}
      </Box>
    </Box>
  );
}
