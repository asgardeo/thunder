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

import type {ApiAgentType} from '@thunderid/configure-agent-types';
import {renderSchemaField} from '@thunderid/configure-users';
import {useResolveDisplayName} from '@thunderid/hooks';
import {Box, Stack, Typography} from '@wso2/oxygen-ui';
import {useEffect, type JSX} from 'react';
import {useForm} from 'react-hook-form';
import {useTranslation} from 'react-i18next';

export interface ConfigureAgentDetailsProps {
  schema: ApiAgentType;
  defaultValues: Record<string, unknown>;
  onFormValuesChange: (values: Record<string, unknown>) => void;
  onReadyChange?: (isReady: boolean) => void;
}

type AgentDetailsFormData = Record<string, unknown>;

export default function ConfigureAgentDetails({
  schema,
  defaultValues,
  onFormValuesChange,
  onReadyChange = undefined,
}: ConfigureAgentDetailsProps): JSX.Element {
  const {t} = useTranslation();
  const {resolveDisplayName} = useResolveDisplayName({handlers: {t}});

  const {
    control,
    watch,
    formState: {errors, isValid},
  } = useForm<AgentDetailsFormData>({
    defaultValues,
    mode: 'onChange',
  });

  useEffect(() => {
    // eslint-disable-next-line react-hooks/incompatible-library
    const subscription = watch((values) => {
      onFormValuesChange(values as Record<string, unknown>);
    });

    return () => subscription.unsubscribe();
  }, [watch, onFormValuesChange]);

  useEffect((): void => {
    onReadyChange?.(isValid);
  }, [isValid, onReadyChange]);

  return (
    <Stack direction="column" spacing={4} data-testid="configure-agent-details">
      <Typography variant="h1" gutterBottom>
        {t('agents:createWizard.agentDetails.title', 'Agent attributes')}
      </Typography>
      <Typography variant="body1" color="text.secondary">
        {t(
          'agents:createWizard.agentDetails.subtitle',
          'Provide values for the attributes defined by the agent schema.',
        )}
      </Typography>

      <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
        {schema.schema &&
          Object.entries(schema.schema).map(([fieldName, fieldDef]) =>
            renderSchemaField(fieldName, fieldDef, control, errors, resolveDisplayName),
          )}
      </Box>
    </Stack>
  );
}
