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

import {Box, Typography} from '@wso2/oxygen-ui';
import {Layers} from '@wso2/oxygen-ui-icons-react';
import {type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import type {Element as FlowElement} from '@/features/flows/models/elements';

export interface DynamicInputPlaceholderAdapterPropsInterface {
  resource: FlowElement;
}

function DynamicInputPlaceholderAdapter({resource}: DynamicInputPlaceholderAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();

  const placeholder =
    (resource as FlowElement & {placeholder?: string}).placeholder ??
    t('flows:core.placeholders.dynamicInputPlaceholder.title');
  const hint =
    (resource as FlowElement & {hint?: string}).hint ?? t('flows:core.placeholders.dynamicInputPlaceholder.hint');

  return (
    <Box
      className="adapter dynamic-input-placeholder-adapter"
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 0.75,
        minHeight: 72,
        px: 2,
        py: 1.5,
        borderRadius: 1,
        border: '1px dashed',
        borderColor: 'primary.light',
        backgroundColor: 'action.hover',
        textAlign: 'center',
      }}
    >
      <Layers size={20} color="primary" />
      <Typography variant="h5" color="primary">
        {placeholder}
      </Typography>
      <Typography variant="subtitle2" color="text.secondary" sx={{fontSize: '0.7rem'}}>
        {hint}
      </Typography>
    </Box>
  );
}

export default DynamicInputPlaceholderAdapter;
