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
import type {LayoutListItem} from '@thunder/shared-design';
import type {JSX} from 'react';

export interface LayoutThumbnailProps {
  layout: LayoutListItem;
}

export default function LayoutThumbnail({layout}: LayoutThumbnailProps): JSX.Element {
  const screens = layout.layout?.screens as Record<string, Record<string, unknown>> | undefined;
  const authScreen = screens?.auth;
  const hasHeader = !!(authScreen?.slots as Record<string, unknown> | undefined)?.header;
  const hasFooter = !!(authScreen?.slots as Record<string, unknown> | undefined)?.footer;
  const bg = authScreen?.background
    ? ((authScreen.background as Record<string, unknown>).value as string | undefined)
    : undefined;

  return (
    <Box
      sx={{
        width: '100%',
        height: '100%',
        bgcolor: bg ? undefined : 'grey.100',
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
          bgcolor: 'rgba(255,255,255,0.15)',
          backdropFilter: 'blur(4px)',
          borderRadius: '8px',
          border: '1px solid rgba(255,255,255,0.3)',
          overflow: 'hidden',
          display: 'flex',
          flexDirection: 'column',
          boxShadow: '0 4px 20px rgba(0,0,0,0.2)',
        }}
      >
        {hasHeader && (
          <Box
            sx={{
              height: 18,
              bgcolor: 'rgba(255,255,255,0.9)',
              borderBottom: '1px solid rgba(0,0,0,0.08)',
              display: 'flex',
              alignItems: 'center',
              px: 1.25,
              gap: 0.75,
              flexShrink: 0,
            }}
          >
            <Box sx={{width: 10, height: 6, bgcolor: 'primary.main', borderRadius: '1.5px', opacity: 0.85}} />
            <Box sx={{flex: 1}} />
            <Box sx={{width: 14, height: 5, bgcolor: 'rgba(0,0,0,0.15)', borderRadius: '3px'}} />
          </Box>
        )}
        <Box sx={{flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center'}}>
          <Box
            sx={{
              width: '55%',
              bgcolor: 'rgba(255,255,255,0.95)',
              borderRadius: '5px',
              boxShadow: '0 2px 10px rgba(0,0,0,0.15)',
              p: 1,
              display: 'flex',
              flexDirection: 'column',
              gap: 0.5,
            }}
          >
            <Box sx={{height: 5, width: '62%', bgcolor: 'rgba(0,0,0,0.5)', borderRadius: 0.5, mx: 'auto'}} />
            <Box
              sx={{height: 7, bgcolor: 'rgba(0,0,0,0.06)', borderRadius: '3px', border: '0.5px solid rgba(0,0,0,0.1)'}}
            />
            <Box
              sx={{height: 7, bgcolor: 'rgba(0,0,0,0.06)', borderRadius: '3px', border: '0.5px solid rgba(0,0,0,0.1)'}}
            />
            <Box sx={{height: 10, bgcolor: 'primary.main', borderRadius: '3px', opacity: 0.85}} />
          </Box>
        </Box>
        {hasFooter && (
          <Box
            sx={{
              height: 16,
              bgcolor: 'rgba(255,255,255,0.85)',
              borderTop: '1px solid rgba(0,0,0,0.06)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 1,
              flexShrink: 0,
            }}
          >
            {[1, 2, 3].map((i) => (
              <Box key={i} sx={{width: 14, height: 3.5, bgcolor: 'rgba(0,0,0,0.2)', borderRadius: 0.5}} />
            ))}
          </Box>
        )}
      </Box>
    </Box>
  );
}
