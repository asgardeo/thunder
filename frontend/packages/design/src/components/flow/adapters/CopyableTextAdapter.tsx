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

import {Box, Button, Typography} from '@wso2/oxygen-ui';
import {Check, Copy} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useCallback, useEffect, useRef, useState} from 'react';
import {useTranslation} from 'react-i18next';
import type {FlowComponent} from '../../../models/flow';

/**
 * Props for the CopyableTextAdapter component.
 */
interface CopyableTextAdapterProps {
  additionalData?: Record<string, unknown>;
  component: FlowComponent;
  resolve: (template: string | undefined) => string | undefined;
}

/**
 * Adapter component to render a copyable text field within a flow. It displays a label (if provided) and a value
 * with a copy-to-clipboard button. The value is sourced from the `additionalData` using the `source` key defined
 * in the component configuration. When the copy button is clicked, it attempts to copy the value to the clipboard
 * and provides feedback to the user.
 *
 * @param {CopyableTextAdapterProps} props - The properties for the adapter, including the flow component
 * configuration, the resolve function for template strings, and any additional data needed to source the value.
 * @returns {JSX.Element} The rendered copyable text field with label and copy button.
 */
export default function CopyableTextAdapter({
  component,
  resolve,
  additionalData = undefined,
}: CopyableTextAdapterProps): JSX.Element {
  const {t} = useTranslation();
  const [copied, setCopied] = useState(false);
  const copyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(
    () => () => {
      if (copyTimeoutRef.current !== null) clearTimeout(copyTimeoutRef.current);
    },
    [],
  );

  const sourceKey = (component as FlowComponent & {source?: string}).source;
  const rawValue = sourceKey && additionalData ? additionalData[sourceKey] : undefined;
  const value =
    typeof rawValue === 'string' || typeof rawValue === 'number' || typeof rawValue === 'boolean'
      ? String(rawValue)
      : '';
  const label = component.label ? t(resolve(component.label) ?? component.label) : undefined;

  // Handle the copy action, attempting to use the Clipboard API and falling back to a textarea method if necessary.
  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(value);
    } catch {
      const textArea = document.createElement('textarea');
      textArea.value = value;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
    }
    setCopied(true);
    if (copyTimeoutRef.current !== null) clearTimeout(copyTimeoutRef.current);
    copyTimeoutRef.current = setTimeout(() => setCopied(false), 3000);
  }, [value]);

  return (
    <Box sx={{display: 'flex', flexDirection: 'column', gap: 0.5, width: '100%'}}>
      {label && (
        <Typography variant="body2" color="text.secondary" sx={{fontWeight: 500}}>
          {label}
        </Typography>
      )}
      <Box
        sx={{
          alignItems: 'center',
          backgroundColor: 'background.default',
          border: '1px solid',
          borderColor: 'divider',
          borderRadius: 1,
          display: 'flex',
          gap: 1,
          p: 1.5,
        }}
      >
        <Typography
          variant="body2"
          sx={{
            flex: 1,
            fontFamily: 'monospace',
            fontSize: '0.85rem',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          {value}
        </Typography>
        <Button
          variant={copied ? 'text' : 'outlined'}
          size="small"
          color={copied ? 'success' : 'primary'}
          startIcon={copied ? <Check size={16} /> : <Copy size={16} />}
          onClick={() => {
            void handleCopy();
          }}
          aria-label={copied ? t('common:actions.copied', 'Copied!') : t('common:actions.copy', 'Copy')}
        >
          {copied ? t('common:actions.copied', 'Copied!') : t('common:actions.copy', 'Copy')}
        </Button>
      </Box>
    </Box>
  );
}
