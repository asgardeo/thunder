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

import {Alert, Box} from '@wso2/oxygen-ui';
import Editor from '@monaco-editor/react';
import {useEffect, useRef, useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';

export interface TranslationJsonEditorProps {
  /** Current merged values (server + local edits). */
  values: Record<string, string>;
  colorMode: 'light' | 'dark';
  /**
   * Called whenever the editor contains valid JSON that parses to a Record<string, string>.
   * The parent uses this to update its local changes state.
   */
  onChange: (changes: Record<string, string>) => void;
}

export default function TranslationJsonEditor({
  values,
  colorMode,
  onChange,
}: TranslationJsonEditorProps): JSX.Element {
  const {t} = useTranslation('translations');

  const [jsonText, setJsonText] = useState(() => JSON.stringify(values, null, 2));
  const [jsonError, setJsonError] = useState(false);

  // Debounce ref so we don't call onChange on every keystroke
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Keep editor in sync when values change from the outside (e.g. namespace switch)
  const prevValuesRef = useRef(values);
  useEffect(() => {
    if (prevValuesRef.current !== values) {
      prevValuesRef.current = values;
      setJsonText(JSON.stringify(values, null, 2));
      setJsonError(false);
    }
  }, [values]);

  const handleEditorChange = (raw: string | undefined) => {
    const text = raw ?? '';
    setJsonText(text);

    if (debounceRef.current) clearTimeout(debounceRef.current);

    debounceRef.current = setTimeout(() => {
      try {
        const parsed = JSON.parse(text) as unknown;
        if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
          const record = parsed as Record<string, unknown>;
          const stringRecord: Record<string, string> = Object.fromEntries(
            Object.entries(record).filter(([, v]) => typeof v === 'string') as [string, string][],
          );
          setJsonError(false);
          onChange(stringRecord);
        } else {
          setJsonError(true);
        }
      } catch {
        setJsonError(true);
      }
    }, 400);
  };

  return (
    <Box sx={{display: 'flex', flexDirection: 'column', height: '100%', gap: 1}}>
      {jsonError && jsonText.trim().length > 0 && (
        <Alert severity="warning" sx={{flexShrink: 0}}>
          {t('editor.jsonInvalid')}
        </Alert>
      )}

      <Box
        sx={{
          flex: 1,
          overflow: 'hidden',
          borderRadius: 0,
          border: '1px solid',
          borderColor: jsonError ? 'warning.main' : 'divider',
        }}
      >
        <Editor
          height="100%"
          language="json"
          theme={colorMode === 'dark' ? 'vs-dark' : 'vs'}
          value={jsonText}
          onChange={handleEditorChange}
          options={{
            minimap: {enabled: false},
            scrollBeyondLastLine: false,
            automaticLayout: true,
            fontSize: 12,
            tabSize: 2,
            wordWrap: 'on',
            lineNumbers: 'off',
            folding: false,
          }}
        />
      </Box>
    </Box>
  );
}
