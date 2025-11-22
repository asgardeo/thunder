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

import {Box, Typography, Stack, TextField, Chip, FormControl, FormLabel, useTheme} from '@wso2/oxygen-ui';
import {Lightbulb} from '@wso2/oxygen-ui-icons-react';
import type {ChangeEvent, JSX} from 'react';
import {useMemo, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import generateAppNameSuggestions from '../../utils/generateAppNameSuggestions';

/**
 * Props for the ConfigureName component.
 */
export interface ConfigureNameProps {
  /**
   * The current application name
   */
  appName: string;

  /**
   * Callback function when the application name changes
   */
  onAppNameChange: (name: string) => void;

  /**
   * Callback function to broadcast whether this step is ready to proceed
   */
  onReadyChange?: (isReady: boolean) => void;
}

export default function ConfigureName({
  appName,
  onAppNameChange,
  onReadyChange = undefined,
}: ConfigureNameProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();

  const appNameSuggestions: string[] = useMemo((): string[] => generateAppNameSuggestions(), []);

  // Broadcast readiness whenever appName changes
  useEffect(() => {
    const isReady = appName.trim().length > 0;
    if (onReadyChange) {
      onReadyChange(isReady);
    }
  }, [appName, onReadyChange]);

  const handleNameSuggestionClick = (suggestion: string): void => {
    onAppNameChange(suggestion);
  };

  return (
    <Stack direction="column" spacing={4}>
      <Typography variant="h1" gutterBottom>
        {t('applications:onboarding.configure.name.title')}
      </Typography>

      <FormControl fullWidth required>
        <FormLabel htmlFor="app-name-input">{t('applications:onboarding.configure.name.fieldLabel')}</FormLabel>
        <TextField
          fullWidth
          id="app-name-input"
          value={appName}
          onChange={(e: ChangeEvent<HTMLInputElement>): void => onAppNameChange(e.target.value)}
          placeholder={t('applications:onboarding.configure.name.placeholder')}
        />
      </FormControl>

      {/* Name suggestions */}
      <Stack direction="column" spacing={2}>
        <Stack direction="row" alignItems="center" spacing={1}>
          <Lightbulb size={20} color={theme.vars?.palette.warning.main} />
          <Typography variant="body2" color="text.secondary">
            {t('applications:onboarding.configure.name.suggestions.label')}
          </Typography>
        </Stack>
        <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1}}>
          {appNameSuggestions.map(
            (suggestion: string): JSX.Element => (
              <Chip
                key={suggestion}
                label={suggestion}
                onClick={(): void => handleNameSuggestionClick(suggestion)}
                variant="outlined"
                clickable
                sx={{
                  '&:hover': {
                    bgcolor: 'primary.main',
                    color: 'primary.contrastText',
                    borderColor: 'primary.main',
                  },
                }}
              />
            ),
          )}
        </Box>
      </Stack>
    </Stack>
  );
}
