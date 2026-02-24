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

import {Box, Chip, FormControl, FormLabel, Stack, TextField, Typography, useTheme} from '@wso2/oxygen-ui';
import {Lightbulb} from '@wso2/oxygen-ui-icons-react';
import {type ChangeEvent, type JSX, useEffect, useMemo} from 'react';
import {generateRandomHumanReadableIdentifiers} from '@thunder/shared-utils';

export interface ConfigureThemeNameProps {
  themeName: string;
  onThemeNameChange: (name: string) => void;
  onReadyChange?: (isReady: boolean) => void;
}

export default function ConfigureThemeName({
  themeName,
  onThemeNameChange,
  onReadyChange,
}: ConfigureThemeNameProps): JSX.Element {
  const theme = useTheme();
  const suggestions = useMemo(() => generateRandomHumanReadableIdentifiers(), []);

  useEffect(() => {
    onReadyChange?.(themeName.trim().length > 0);
  }, [themeName, onReadyChange]);

  return (
    <Stack direction="column" spacing={4}>
      <Typography variant="h1">Let's give a name to your theme</Typography>

      <FormControl fullWidth required>
        <FormLabel htmlFor="theme-name-input">Theme name</FormLabel>
        <TextField
          fullWidth
          id="theme-name-input"
          value={themeName}
          onChange={(e: ChangeEvent<HTMLInputElement>) => onThemeNameChange(e.target.value)}
          placeholder="e.g. Solarized Light"
          autoFocus
        />
      </FormControl>

      <Stack direction="column" spacing={2}>
        <Stack direction="row" alignItems="center" spacing={1}>
          <Lightbulb size={20} color={theme.vars?.palette.warning.main} />
          <Typography variant="body2" color="text.secondary">
            Need inspiration? Pick a suggestion:
          </Typography>
        </Stack>
        <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1}}>
          {suggestions.map((s) => (
            <Chip
              key={s}
              label={s}
              onClick={() => onThemeNameChange(s)}
              variant="outlined"
              clickable
              sx={{
                '&:hover': {bgcolor: 'primary.main', color: 'primary.contrastText', borderColor: 'primary.main'},
              }}
            />
          ))}
        </Box>
      </Stack>
    </Stack>
  );
}
