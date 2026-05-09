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

import {generateRandomHumanReadableIdentifiers} from '@thunderid/utils';
import {Box, Chip, FormControl, FormLabel, Stack, TextField, Typography, useTheme} from '@wso2/oxygen-ui';
import {Lightbulb} from '@wso2/oxygen-ui-icons-react';
import {useEffect, useMemo, type ChangeEvent, type JSX} from 'react';
import {useTranslation} from 'react-i18next';

export interface ConfigureNameProps {
  agentName: string;
  onAgentNameChange: (name: string) => void;
  onReadyChange?: (isReady: boolean) => void;
}

export default function ConfigureName({
  agentName,
  onAgentNameChange,
  onReadyChange = undefined,
}: ConfigureNameProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();

  const nameSuggestions: string[] = useMemo(() => generateRandomHumanReadableIdentifiers(), []);

  useEffect((): void => {
    onReadyChange?.(agentName.trim().length > 0);
  }, [agentName, onReadyChange]);

  return (
    <Stack direction="column" spacing={4} data-testid="configure-agent-name">
      <Typography variant="h1" gutterBottom>
        {t('agents:createWizard.name.title', "What's this agent called?")}
      </Typography>

      <FormControl fullWidth required>
        <FormLabel htmlFor="agent-name-input">{t('agents:createWizard.name.fieldLabel', 'Agent name')}</FormLabel>
        <TextField
          fullWidth
          id="agent-name-input"
          value={agentName}
          onChange={(e: ChangeEvent<HTMLInputElement>): void => onAgentNameChange(e.target.value)}
          placeholder={t('agents:createWizard.name.placeholder', 'e.g. Billing Service')}
          inputProps={{'data-testid': 'agent-name-input'}}
        />
      </FormControl>

      <Stack direction="column" spacing={2}>
        <Stack direction="row" alignItems="center" spacing={1}>
          <Lightbulb size={20} color={theme.vars?.palette.warning.main} />
          <Typography variant="body2" color="text.secondary">
            {t('agents:createWizard.name.suggestions.label', 'Need inspiration? Pick one:')}
          </Typography>
        </Stack>
        <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1}}>
          {nameSuggestions.map(
            (suggestion: string): JSX.Element => (
              <Chip
                key={suggestion}
                label={suggestion}
                onClick={(): void => onAgentNameChange(suggestion)}
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
