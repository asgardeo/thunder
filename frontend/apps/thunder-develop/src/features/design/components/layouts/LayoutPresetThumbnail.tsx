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
import type {JSX} from 'react';

export type LayoutPresetVariant = 'centered' | 'split' | 'fullscreen' | 'popup';

export interface LayoutPresetThumbnailProps {
  variant: LayoutPresetVariant;
}

const BORDER = '1.5px solid rgba(0,0,0,0.18)';
const FILL = 'rgba(0,0,0,0.07)';
const FILL_DARK = 'rgba(0,0,0,0.13)';
const BG = '#f5f5f5';
const RADIUS = '3px';

/**
 * Centered — plain background with a single card block centered on the page.
 *
 *  ┌────────────────────┐
 *  │                    │
 *  │     ┌────────┐     │
 *  │     │        │     │
 *  │     │        │     │
 *  │     └────────┘     │
 *  │                    │
 *  └────────────────────┘
 */
function CenteredThumbnail(): JSX.Element {
  return (
    <Box sx={{width: '100%', height: '100%', bgcolor: BG, display: 'flex', alignItems: 'center', justifyContent: 'center', p: 2}}>
      <Box
        sx={{
          width: '48%',
          height: '60%',
          border: BORDER,
          borderRadius: '4px',
          bgcolor: '#fff',
          boxShadow: '0 2px 6px rgba(0,0,0,0.1)',
        }}
      />
    </Box>
  );
}

/**
 * Split Screen — left panel + right panel, equal halves.
 *
 *  ┌──────────┬──────────┐
 *  │          │          │
 *  │  ██████  │  ┌────┐  │
 *  │  ██████  │  │    │  │
 *  │  ██████  │  └────┘  │
 *  │          │          │
 *  └──────────┴──────────┘
 */
function SplitThumbnail(): JSX.Element {
  return (
    <Box sx={{width: '100%', height: '100%', bgcolor: BG, display: 'flex', border: BORDER, borderRadius: RADIUS, overflow: 'hidden'}}>
      {/* Left panel — decorative fill */}
      <Box
        sx={{
          flex: '0 0 48%',
          bgcolor: FILL_DARK,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 1,
          borderRight: BORDER,
          p: 1.5,
        }}
      >
        {[70, 55, 40].map((w, i) => (
          <Box key={i} sx={{height: 5, width: `${w}%`, bgcolor: 'rgba(0,0,0,0.2)', borderRadius: 0.5}} />
        ))}
      </Box>

      {/* Right panel — form card */}
      <Box sx={{flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', p: 1.5, bgcolor: '#fff'}}>
        <Box sx={{width: '70%', height: '65%', border: BORDER, borderRadius: '4px', bgcolor: BG}} />
      </Box>
    </Box>
  );
}

/**
 * Full Screen — top navigation bar + full-width content below.
 *
 *  ┌────────────────────┐
 *  │████████████████████│  ← nav bar
 *  ├────────────────────┤
 *  │                    │
 *  │  ┌──────────────┐  │
 *  │  │              │  │
 *  │  └──────────────┘  │
 *  └────────────────────┘
 */
function FullScreenThumbnail(): JSX.Element {
  return (
    <Box
      sx={{
        width: '100%',
        height: '100%',
        bgcolor: BG,
        display: 'flex',
        flexDirection: 'column',
        border: BORDER,
        borderRadius: RADIUS,
        overflow: 'hidden',
      }}
    >
      {/* Nav bar */}
      <Box
        sx={{
          height: '18%',
          bgcolor: FILL_DARK,
          borderBottom: BORDER,
          flexShrink: 0,
          display: 'flex',
          alignItems: 'center',
          px: 1.5,
          gap: 0.75,
        }}
      >
        <Box sx={{width: 18, height: 5, bgcolor: 'rgba(0,0,0,0.22)', borderRadius: 0.5}} />
        <Box sx={{flex: 1}} />
        <Box sx={{width: 10, height: 5, bgcolor: 'rgba(0,0,0,0.12)', borderRadius: 0.5}} />
      </Box>

      {/* Content */}
      <Box sx={{flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', p: 1.5}}>
        <Box sx={{width: '80%', height: '70%', border: BORDER, borderRadius: '4px', bgcolor: '#fff'}} />
      </Box>
    </Box>
  );
}

/**
 * Floating Card — card elevated above a textured/patterned background.
 *
 *  ┌────────────────────┐
 *  │ ░░░░░░░░░░░░░░░░░░ │  ← patterned bg
 *  │ ░░  ┌──────┐  ░░░ │
 *  │ ░░  │      │  ░░░ │
 *  │ ░░  └──────┘  ░░░ │
 *  │ ░░░░░░░░░░░░░░░░░░ │
 *  └────────────────────┘
 */
function FloatingThumbnail(): JSX.Element {
  return (
    <Box
      sx={{
        width: '100%',
        height: '100%',
        bgcolor: FILL,
        backgroundImage: `repeating-linear-gradient(
          45deg,
          rgba(0,0,0,0.04) 0px,
          rgba(0,0,0,0.04) 1px,
          transparent 1px,
          transparent 7px
        )`,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        p: 2,
        border: BORDER,
        borderRadius: RADIUS,
        overflow: 'hidden',
      }}
    >
      <Box
        sx={{
          width: '48%',
          height: '60%',
          border: BORDER,
          borderRadius: '5px',
          bgcolor: '#fff',
          boxShadow: '0 4px 16px rgba(0,0,0,0.15)',
        }}
      />
    </Box>
  );
}

export default function LayoutPresetThumbnail({variant}: LayoutPresetThumbnailProps): JSX.Element {
  switch (variant) {
    case 'centered':
      return <CenteredThumbnail />;
    case 'split':
      return <SplitThumbnail />;
    case 'fullscreen':
      return <FullScreenThumbnail />;
    case 'popup':
      return <FloatingThumbnail />;
  }
}
