'use client';

/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { useColorScheme } from '@mui/material/styles';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import { DarkModeRounded, LightModeRounded, Monitor } from '@mui/icons-material';
import { useMounted } from '../../hooks/useMounted';

type Mode = 'light' | 'dark' | 'system';

export default function ThemeToggle() {
  const mounted = useMounted();
  const { mode, setMode } = useColorScheme();

  if (!mounted) {
    return (
      <Button variant="outlined">
        <CircularProgress size={24} />
      </Button>
    );
  }

  if (!mode) {
    return null;
  }

  const nextMode = (mode: Mode): Mode => {
    return mode === 'light' ? 'dark' : mode === 'dark' ? 'system' : 'light';
  };

  const currentMode: Mode = mode ?? 'system';

  const ColorModeIcon = () => {
    switch (currentMode) {
      case 'light':
        return <LightModeRounded />;
      case 'dark':
        return <DarkModeRounded />;
      default:
        return <Monitor />;
    }
  };

  return (
    <Button variant="outlined" onClick={() => setMode(nextMode(currentMode))}>
      <ColorModeIcon />
    </Button>
  );
};
