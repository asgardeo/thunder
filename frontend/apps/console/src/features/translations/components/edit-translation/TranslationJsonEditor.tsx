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

/**
 * Props for the {@link TranslationJsonEditor} component.
 *
 * @public
 */
export interface TranslationJsonEditorProps {
  /** Current merged values (server + local edits). */
  values: Record<string, string>;
  /** Keys from the server — used to block adding new keys in non-custom namespaces. */
  serverKeys: string[];
  /** Whether the active namespace is "custom", which allows adding new keys. */
  isCustomNamespace: boolean;
  /** Current color mode used to apply the Monaco editor theme. */
  colorMode: 'light' | 'dark';
  /**
   * Called whenever the editor contains valid JSON that parses to a `Record<string, string>`.
   * The parent uses this to update its local changes state.
   */
  onChange: (changes: Record<string, string>) => void;
}

/**
 * Monaco-based JSON editor for bulk-editing translation key-value pairs.
 *
 * Displays the current translation values as formatted JSON and notifies the
 * parent whenever the editor content is valid JSON that parses to a flat
 * `Record<string, string>`. Invalid JSON is indicated with a warning alert;
 * the {@link TranslationJsonEditorProps.onChange} callback is suppressed until
 * the content is valid again.
 *
 * @param props - The component props
 * @param props.values - Current merged translation values shown in the editor
 * @param props.colorMode - Current color mode used to apply the Monaco editor theme
 * @param props.onChange - Callback invoked with the parsed record when the JSON is valid
 *
 * @returns JSX element rendering the Monaco JSON editor
 *
 * @example
 * ```tsx
 * import TranslationJsonEditor from './TranslationJsonEditor';
 *
 * function Editor() {
 *   const [changes, setChanges] = useState<Record<string, string>>({});
 *   return (
 *     <TranslationJsonEditor
 *       values={{'actions.save': 'Save'}}
 *       colorMode="light"
 *       onChange={setChanges}
 *     />
 *   );
 * }
 * ```
 *
 * @public
 */
export default function TranslationJsonEditor({
  values,
  serverKeys,
  isCustomNamespace,
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
          let stringRecord: Record<string, string> = Object.fromEntries(
            Object.entries(record).filter(([, v]) => typeof v === 'string') as [string, string][],
          );
          // In non-custom namespaces, strip any keys that don't already exist on the server
          if (!isCustomNamespace) {
            const allowed = new Set(serverKeys);
            stringRecord = Object.fromEntries(Object.entries(stringRecord).filter(([k]) => allowed.has(k)));
          }
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
    <Box sx={{display: 'flex', flexDirection: 'column', height: '100%'}}>
      {!isCustomNamespace && (
        <Alert severity="info" sx={{flexShrink: 0, borderRadius: 0, border: 'none'}}>
          {t('editor.readOnlyKeys')}
        </Alert>
      )}
      {jsonError && jsonText.trim().length > 0 && (
        <Alert severity="warning" sx={{flexShrink: 0, borderRadius: 0, border: 'none'}}>
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
