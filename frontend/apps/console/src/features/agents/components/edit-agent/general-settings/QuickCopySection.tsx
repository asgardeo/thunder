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

import {SettingsCard} from '@thunderid/components';
import {Stack, TextField, InputAdornment, Tooltip, IconButton, FormControl, FormLabel} from '@wso2/oxygen-ui';
import {Copy, Check} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import type {Agent, OAuthAgentConfig} from '../../../models/agent';

interface QuickCopySectionProps {
  agent: Agent;
  oauth2Config?: OAuthAgentConfig;
  copiedField: string | null;
  onCopyToClipboard: (text: string, fieldName: string) => Promise<void>;
}

export default function QuickCopySection({
  agent,
  oauth2Config = undefined,
  copiedField,
  onCopyToClipboard,
}: QuickCopySectionProps) {
  const {t} = useTranslation();

  return (
    <SettingsCard
      title={t('agents:edit.general.sections.quickCopy.title', 'Quick Copy')}
      description={t(
        'agents:edit.general.sections.quickCopy.description',
        'Copy agent identifiers for use in your code.',
      )}
    >
      <Stack spacing={3}>
        <FormControl fullWidth>
          <FormLabel htmlFor="agent-id-input">{t('agents:edit.general.labels.agentId', 'Agent ID')}</FormLabel>
          <TextField
            fullWidth
            id="agent-id-input"
            value={agent.id}
            InputProps={{
              readOnly: true,
              endAdornment: (
                <InputAdornment position="end">
                  <Tooltip title={copiedField === 'agent_id' ? t('common:actions.copied') : t('common:actions.copy')}>
                    <IconButton
                      onClick={() => {
                        onCopyToClipboard(agent.id, 'agent_id').catch(() => null);
                      }}
                      edge="end"
                    >
                      {copiedField === 'agent_id' ? <Check size={16} /> : <Copy size={16} />}
                    </IconButton>
                  </Tooltip>
                </InputAdornment>
              ),
            }}
            helperText={t('agents:edit.general.agentId.hint', 'Unique identifier for this agent')}
            sx={{
              '& input': {
                fontFamily: 'monospace',
                fontSize: '0.875rem',
              },
            }}
          />
        </FormControl>

        {oauth2Config?.clientId && (
          <FormControl fullWidth>
            <FormLabel htmlFor="client-id-input">
              {t('applications:edit.general.labels.clientId', 'Client ID')}
            </FormLabel>
            <TextField
              fullWidth
              id="client-id-input"
              value={oauth2Config?.clientId ?? ''}
              InputProps={{
                readOnly: true,
                endAdornment: (
                  <InputAdornment position="end">
                    <Tooltip title={copiedField === 'clientId' ? t('common:actions.copied') : t('common:actions.copy')}>
                      <IconButton
                        onClick={() => {
                          if (oauth2Config?.clientId) {
                            onCopyToClipboard(oauth2Config.clientId, 'clientId').catch(() => null);
                          }
                        }}
                        edge="end"
                      >
                        {copiedField === 'clientId' ? <Check size={16} /> : <Copy size={16} />}
                      </IconButton>
                    </Tooltip>
                  </InputAdornment>
                ),
              }}
              helperText={t(
                'agents:edit.general.clientId.hint',
                'OAuth2 client identifier used by this agent to obtain tokens',
              )}
              sx={{
                '& input': {
                  fontFamily: 'monospace',
                  fontSize: '0.875rem',
                },
              }}
            />
          </FormControl>
        )}

        {agent.owner && (
          <FormControl fullWidth>
            <FormLabel htmlFor="agent-owner-input">{t('agents:edit.general.labels.ownerId', 'Owner ID')}</FormLabel>
            <TextField
              fullWidth
              id="agent-owner-input"
              value={agent.owner}
              InputProps={{
                readOnly: true,
                endAdornment: (
                  <InputAdornment position="end">
                    <Tooltip title={copiedField === 'owner' ? t('common:actions.copied') : t('common:actions.copy')}>
                      <IconButton
                        onClick={() => {
                          if (agent.owner) {
                            onCopyToClipboard(agent.owner, 'owner').catch(() => null);
                          }
                        }}
                        edge="end"
                      >
                        {copiedField === 'owner' ? <Check size={16} /> : <Copy size={16} />}
                      </IconButton>
                    </Tooltip>
                  </InputAdornment>
                ),
              }}
              helperText={t('agents:edit.general.owner.hint', 'Identifier of the user that owns this agent')}
              sx={{'& input': {fontFamily: 'monospace', fontSize: '0.875rem'}}}
            />
          </FormControl>
        )}
      </Stack>
    </SettingsCard>
  );
}
