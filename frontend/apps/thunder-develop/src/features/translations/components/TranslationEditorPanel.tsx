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

import {Box, CircularProgress, InputAdornment, Tab, Tabs, TextField, Typography} from '@wso2/oxygen-ui';
import {Search} from '@wso2/oxygen-ui-icons-react';
import {useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import TranslationFieldsView from './TranslationFieldsView';
import TranslationJsonEditor from './TranslationJsonEditor';

export interface TranslationEditorPanelProps {
  language: string | null;
  namespace: string;
  translations: Record<string, string>;
  serverUrl: string;
  editView: 'fields' | 'json';
  onEditViewChange: (view: 'fields' | 'json') => void;
  isLoading: boolean;
  colorMode: 'light' | 'dark';
}

export default function TranslationEditorPanel({
  language,
  namespace,
  translations,
  serverUrl,
  editView,
  onEditViewChange,
  isLoading,
  colorMode,
}: TranslationEditorPanelProps): JSX.Element {
  const {t} = useTranslation('translations');
  const [search, setSearch] = useState('');

  if (!language) {
    return (
      <Box sx={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', p: 3}}>
        <Typography variant="body2" color="text.secondary" align="center">
          {t('editor.noLanguageSelected')}
        </Typography>
      </Box>
    );
  }

  if (isLoading) {
    return (
      <Box sx={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 1.5}}>
        <CircularProgress size={20} />
        <Typography variant="body2" color="text.secondary">
          {t('editor.loading')}
        </Typography>
      </Box>
    );
  }

  return (
    <Box sx={{display: 'flex', flexDirection: 'column', height: '100%'}}>
      {/* Tab switcher */}
      <Box sx={{flexShrink: 0, borderBottom: '1px solid', borderColor: 'divider', px: 2}}>
        <Tabs
          value={editView}
          onChange={(_, v: 'fields' | 'json') => {
            onEditViewChange(v);
            setSearch('');
          }}
          sx={{'& .MuiTab-root': {minHeight: 40, py: 0.5, fontSize: '0.8125rem'}}}
        >
          <Tab label={t('editor.fieldsTab')} value="fields" />
          <Tab label={t('editor.jsonTab')} value="json" />
        </Tabs>
      </Box>

      {/* Fields view: search box + scrollable fields */}
      {editView === 'fields' && (
        <>
          <Box sx={{flexShrink: 0, px: 2, pt: 1.5}}>
            <TextField
              size="small"
              fullWidth
              placeholder={t('editor.searchPlaceholder')}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <Search size={14} />
                  </InputAdornment>
                ),
              }}
            />
          </Box>
          <Box sx={{flex: 1, overflow: 'auto', px: 2, py: 1.5}}>
            <TranslationFieldsView
              language={language}
              namespace={namespace}
              translations={translations}
              serverUrl={serverUrl}
              search={search}
            />
          </Box>
        </>
      )}

      {/* JSON editor */}
      {editView === 'json' && (
        <Box sx={{flex: 1, overflow: 'hidden', p: 1.5}}>
          <TranslationJsonEditor
            language={language}
            namespace={namespace}
            translations={translations}
            serverUrl={serverUrl}
            colorMode={colorMode}
          />
        </Box>
      )}
    </Box>
  );
}
