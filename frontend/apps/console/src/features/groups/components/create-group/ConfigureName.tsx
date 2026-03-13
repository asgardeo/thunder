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

import {Box, Typography, Stack, TextField, Chip, FormControl, FormLabel, useTheme} from '@wso2/oxygen-ui';
import {Lightbulb} from '@wso2/oxygen-ui-icons-react';
import type {ChangeEvent, JSX} from 'react';
import {useMemo, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {generateRandomHumanReadableIdentifiers} from '@thunder/utils';

/**
 * Props for the {@link ConfigureName} component.
 *
 * @public
 */
export interface ConfigureNameProps {
  name: string;
  onNameChange: (name: string) => void;
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * Step 1 of the group creation wizard: configure the group name.
 *
 * @public
 */
export default function ConfigureName({
  name,
  onNameChange,
  onReadyChange = undefined,
}: ConfigureNameProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();

  const nameSuggestions: string[] = useMemo((): string[] => generateRandomHumanReadableIdentifiers(), []);

  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(name.trim().length > 0);
    }
  }, [name, onReadyChange]);

  const handleNameSuggestionClick = (suggestion: string): void => {
    onNameChange(suggestion);
  };

  return (
    <Stack direction="column" spacing={4} data-testid="configure-name">
      <Typography variant="h1" gutterBottom>
        {t('groups:createWizard.name.title')}
      </Typography>

      <FormControl fullWidth required>
        <FormLabel htmlFor="group-name-input">{t('groups:create.form.name.label')}</FormLabel>
        <TextField
          fullWidth
          id="group-name-input"
          value={name}
          onChange={(e: ChangeEvent<HTMLInputElement>): void => onNameChange(e.target.value)}
          placeholder={t('groups:create.form.name.placeholder')}
          inputProps={{
            'data-testid': 'group-name-input',
          }}
        />
      </FormControl>

      {/* Name suggestions */}
      <Stack direction="column" spacing={2}>
        <Stack direction="row" alignItems="center" spacing={1}>
          <Lightbulb size={20} color={theme.vars?.palette.warning.main} />
          <Typography variant="body2" color="text.secondary">
            {t('groups:createWizard.name.suggestions.label')}
          </Typography>
        </Stack>
        <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1}}>
          {nameSuggestions.map(
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
