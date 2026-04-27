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

import {Box, Card, CircularProgress, Divider, InputAdornment, Tab, Tabs, TextField, Typography} from '@wso2/oxygen-ui';
import {Search} from '@wso2/oxygen-ui-icons-react';
import {type JSX, type SyntheticEvent} from 'react';
import {useTranslation} from 'react-i18next';
import TranslationFieldsView from './TranslationFieldsView';
import TranslationJsonEditor from './TranslationJsonEditor';

/**
 * Props for the {@link TranslationEditorCard} component.
 *
 * @public
 */
export interface TranslationEditorCardProps {
  /** The currently selected language code, or null if none. */
  selectedLanguage: string | null;
  /** Whether the translation data is loading. */
  isLoading: boolean;
  /** The active editor tab. */
  editView: 'fields' | 'json';
  /** Current search query for the fields view. */
  search: string;
  /** Merged current values (server + local changes) for the selected namespace. */
  currentValues: Record<string, string>;
  /** Server-saved values for the selected namespace. */
  serverValues: Record<string, string>;
  /** Whether the active namespace is "custom", which allows adding new keys. */
  isCustomNamespace: boolean;
  /** Color mode passed to the JSON editor. */
  colorMode: 'light' | 'dark';
  /** Called when the user switches between the Fields and JSON tabs. */
  onTabChange: (_: SyntheticEvent, value: 'fields' | 'json') => void;
  /** Called when the search query changes. */
  onSearchChange: (search: string) => void;
  /** Called when a field value is edited. */
  onFieldChange: (key: string, value: string) => void;
  /** Called when a field is reset to its server value. */
  onResetField: (key: string) => void;
  /** Called when the JSON editor emits a full set of changes. */
  onJsonChange: (changes: Record<string, string>) => void;
}

/**
 * Tabbed card editor for translation key-value pairs. Renders a Fields view
 * (searchable list of text inputs) and a Raw JSON view. Shows a loading
 * spinner while data is being fetched.
 *
 * @param props - The component props
 *
 * @returns JSX element rendering the editor card
 *
 * @public
 */
export default function TranslationEditorCard({
  selectedLanguage,
  isLoading,
  editView,
  search,
  currentValues,
  serverValues,
  isCustomNamespace,
  colorMode,
  onTabChange,
  onSearchChange,
  onFieldChange,
  onResetField,
  onJsonChange,
}: TranslationEditorCardProps): JSX.Element {
  const {t} = useTranslation('translations');

  return (
    <Box sx={{flex: 1, overflow: 'hidden', display: 'flex', gap: 2.5, minHeight: 0}}>
      <Card
        variant="outlined"
        sx={{flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column', borderRadius: 2}}
      >
        <Box sx={{px: 2.5, pt: 2, pb: 0, flexShrink: 0}}>
          <Tabs
            value={editView}
            onChange={onTabChange}
            sx={{'& .MuiTab-root': {minHeight: 38, py: 0.5, fontSize: '0.8125rem', textTransform: 'none'}}}
          >
            <Tab label={t('editor.textFields')} value="fields" />
            <Tab label={t('editor.rawJson')} value="json" />
          </Tabs>
        </Box>

        <Divider />

        {isLoading && (
          <Box sx={{flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 1.5}}>
            <CircularProgress size={20} />
            <Typography variant="body2" color="text.secondary">
              {t('editor.loading')}
            </Typography>
          </Box>
        )}

        {selectedLanguage && !isLoading && editView === 'fields' && (
          <>
            <Box sx={{px: 2.5, pt: 1.5, pb: 0.5, flexShrink: 0}}>
              <TextField
                size="small"
                fullWidth
                placeholder={t('editor.searchPlaceholder')}
                value={search}
                onChange={(e) => onSearchChange(e.target.value)}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <Search size={14} />
                    </InputAdornment>
                  ),
                }}
              />
            </Box>
            <Divider sx={{mt: 1}} />
            <Box sx={{flex: 1, overflow: 'auto', px: 2.5, py: 2}}>
              <TranslationFieldsView
                localValues={currentValues}
                serverValues={serverValues}
                search={search}
                isCustomNamespace={isCustomNamespace}
                onChange={onFieldChange}
                onResetField={onResetField}
              />
            </Box>
          </>
        )}

        {selectedLanguage && !isLoading && editView === 'json' && (
          <Box sx={{flex: 1, overflow: 'hidden', p: 0}}>
            <TranslationJsonEditor
              values={currentValues}
              serverKeys={Object.keys(serverValues)}
              isCustomNamespace={isCustomNamespace}
              colorMode={colorMode}
              onChange={onJsonChange}
            />
          </Box>
        )}
      </Card>
    </Box>
  );
}
