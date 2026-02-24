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

import {
  Alert,
  Autocomplete,
  Box,
  Button,
  Card,
  CircularProgress,
  Divider,
  FormControl,
  FormHelperText,
  FormLabel,
  InputAdornment,
  PageContent,
  PageTitle,
  Snackbar,
  Tab,
  Tabs,
  TextField,
  Typography,
  useColorScheme,
} from '@wso2/oxygen-ui';
import {ArrowLeft, Search} from '@wso2/oxygen-ui-icons-react';
import {useGetTranslations, useUpdateTranslation} from '@thunder/i18n';
import {useConfig} from '@thunder/shared-contexts';
import {useCallback, useEffect, useMemo, useState, type JSX, type SyntheticEvent} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate, useParams} from 'react-router';
import {useLogger} from '@thunder/logger/react';
import TranslationFieldsView from '../components/TranslationFieldsView';
import TranslationJsonEditor from '../components/TranslationJsonEditor';
import TranslationPreviewPanel from '../components/TranslationPreviewPanel';
import getLanguageFlag from '../utils/getLanguageFlag';
import getLanguageDisplayName from '../utils/getLanguageDisplayName';

type ToastState = {open: boolean; message: string; severity: 'success' | 'error'};

export default function TranslationsEditPage(): JSX.Element {
  const {t} = useTranslation('translations');
  const {getServerUrl} = useConfig();
  const serverUrl = getServerUrl();
  const navigate = useNavigate();
  const logger = useLogger('TranslationsEditPage');
  const {language: languageParam} = useParams<{language: string}>();
  const selectedLanguage = languageParam ?? null;

  const {mode, systemMode} = useColorScheme();
  const colorMode: 'light' | 'dark' =
    ((mode === 'system' ? systemMode : mode) ?? 'light') === 'dark' ? 'dark' : 'light';

  const [selectedNamespace, setSelectedNamespace] = useState<string | null>(null);
  const [editView, setEditView] = useState<'fields' | 'json'>('fields');
  const [search, setSearch] = useState('');
  const [localChanges, setLocalChanges] = useState<Record<string, string>>({});
  const [isSaving, setIsSaving] = useState(false);
  const [toast, setToast] = useState<ToastState>({open: false, message: '', severity: 'success'});

  const {data: translationsData, isLoading: translationsLoading} = useGetTranslations({
    serverUrl,
    language: selectedLanguage ?? '',
    enabled: !!selectedLanguage,
  });

  // Fetch the default (en) translations for "Reset to Default"
  const {data: defaultTranslationsData} = useGetTranslations({
    serverUrl,
    language: 'en',
    enabled: !!selectedLanguage && selectedLanguage !== 'en',
  });

  const updateTranslation = useUpdateTranslation({serverUrl});

  const namespaces = useMemo(() => Object.keys(translationsData?.translations ?? {}), [translationsData]);

  // Reset namespace when language changes
  useEffect(() => {
    setSelectedNamespace(null);
    setLocalChanges({});
    setSearch('');
  }, [selectedLanguage]);

  // Initialize namespace once API data arrives
  useEffect(() => {
    if (namespaces.length > 0 && !selectedNamespace) {
      setSelectedNamespace(namespaces[0]);
    }
  }, [namespaces, selectedNamespace]);

  // Reset local changes when namespace switches
  useEffect(() => {
    setLocalChanges({});
    setSearch('');
  }, [selectedNamespace]);

  const serverValues: Record<string, string> = useMemo(
    () => translationsData?.translations?.[selectedNamespace ?? ''] ?? {},
    [translationsData, selectedNamespace],
  );

  const currentValues: Record<string, string> = useMemo(
    () => ({...serverValues, ...localChanges}),
    [serverValues, localChanges],
  );

  const dirtyKeys = useMemo(
    () => Object.keys(localChanges).filter((k) => localChanges[k] !== serverValues[k]),
    [localChanges, serverValues],
  );
  const hasDirtyChanges = dirtyKeys.length > 0;

  const handleFieldChange = useCallback((key: string, value: string) => {
    setLocalChanges((prev) => ({...prev, [key]: value}));
  }, []);

  const handleResetField = useCallback((key: string) => {
    setLocalChanges((prev) => {
      const next = {...prev};
      delete next[key];
      return next;
    });
  }, []);

  const handleJsonChange = useCallback((changes: Record<string, string>) => {
    setLocalChanges(changes);
  }, []);

  const handleSave = async () => {
    if (!selectedLanguage || !selectedNamespace || dirtyKeys.length === 0) return;
    setIsSaving(true);

    const results = await Promise.allSettled(
      dirtyKeys.map((key) =>
        updateTranslation.mutateAsync({
          language: selectedLanguage,
          namespace: selectedNamespace,
          key,
          value: localChanges[key],
        }),
      ),
    );

    const failed = results.filter((r) => r.status === 'rejected').length;
    setIsSaving(false);

    if (failed > 0) {
      setToast({open: true, message: t('editor.jsonSaveError'), severity: 'error'});
    } else {
      setLocalChanges({});
      setToast({open: true, message: t('editor.jsonSaveSuccess'), severity: 'success'});
    }
  };

  const handleDiscard = () => {
    setLocalChanges({});
  };

  const handleResetToDefault = async () => {
    if (!selectedLanguage || !selectedNamespace) return;
    const defaultValues = defaultTranslationsData?.translations?.[selectedNamespace] ?? {};
    const entries = Object.entries(defaultValues);
    if (entries.length === 0) return;

    setIsSaving(true);

    const results = await Promise.allSettled(
      entries.map(([key, value]) =>
        updateTranslation.mutateAsync({
          language: selectedLanguage,
          namespace: selectedNamespace,
          key,
          value,
        }),
      ),
    );

    const failed = results.filter((r) => r.status === 'rejected').length;
    setIsSaving(false);
    setLocalChanges({});

    if (failed > 0) {
      setToast({open: true, message: t('editor.jsonSaveError'), severity: 'error'});
    } else {
      setToast({open: true, message: t('editor.jsonSaveSuccess'), severity: 'success'});
    }
  };

  const handleTabChange = (_: SyntheticEvent, v: 'fields' | 'json') => {
    setEditView(v);
    setSearch('');
  };

  const handleBack = () => {
    (async (): Promise<void> => {
      await navigate('/translations');
    })().catch((_error: unknown) => {
      logger.error('Failed to navigate back to translations list', {error: _error});
    });
  };

  const isLoading = !!selectedLanguage && translationsLoading;

  const namespaceLabel = (selectedNamespace ?? '')
    .replace(/([A-Z])/g, ' $1')
    .replace(/^./, (c) => c.toUpperCase())
    .trim();

  const isEnglish = selectedLanguage === 'en' || selectedLanguage === 'en-US';

  return (
    <PageContent fullWidth sx={{display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0}}>
      <PageTitle>
        <PageTitle.Header>
          <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
            <Button
              size="small"
              variant="text"
              startIcon={<ArrowLeft size={16} />}
              onClick={handleBack}
              sx={{minWidth: 0, px: 1, color: 'text.secondary'}}
            />
            {selectedLanguage ? (
              <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
                <Typography component="span" sx={{fontSize: '1.3rem', lineHeight: 1, userSelect: 'none'}}>
                  {getLanguageFlag(selectedLanguage)}
                </Typography>
                {getLanguageDisplayName(selectedLanguage)}
              </Box>
            ) : (
              t('page.title')
            )}
          </Box>
        </PageTitle.Header>
        <PageTitle.Actions>
          <Box sx={{display: 'flex', gap: 1, alignItems: 'center'}}>
            {hasDirtyChanges && (
              <Typography variant="caption" color="warning.main" sx={{fontWeight: 500}}>
                {t('editor.unsavedCount', {count: dirtyKeys.length})}
              </Typography>
            )}
            <Button size="small" onClick={handleDiscard} disabled={!hasDirtyChanges || isSaving}>
              {t('actions.discardChanges')}
            </Button>
            {!isEnglish && (
              <Button
                size="small"
                onClick={() => void handleResetToDefault()}
                disabled={!selectedNamespace || isSaving}
              >
                {t('actions.resetToDefault')}
              </Button>
            )}
            <Button
              size="small"
              variant="contained"
              onClick={() => void handleSave()}
              disabled={!hasDirtyChanges || isSaving}
              startIcon={isSaving ? <CircularProgress size={14} color="inherit" /> : undefined}
            >
              {t('actions.saveChanges')}
            </Button>
          </Box>
        </PageTitle.Actions>
      </PageTitle>

      <Box sx={{display: 'flex', gap: 2, alignItems: 'center', mb: 2}}>
        <FormControl sx={{maxWidth: 600}}>
          <FormLabel>{t('editor.namespace')}</FormLabel>
          <Autocomplete
            options={namespaces}
            value={selectedNamespace ?? ''}
            onChange={(_, v) => v && setSelectedNamespace(v)}
            disableClearable
            size="small"
            loading={isLoading}
            renderInput={(params) => <TextField {...params} />}
            getOptionLabel={(opt) =>
              opt
                .replace(/([A-Z])/g, ' $1')
                .replace(/^./, (c) => c.toUpperCase())
                .trim()
            }
          />
          <FormHelperText>{t('editor.namespace.helperText')}</FormHelperText>
        </FormControl>
      </Box>

      <Box sx={{flex: 1, overflow: 'hidden', display: 'flex', gap: 2.5, minHeight: 0}}>
        <Card
          variant="outlined"
          sx={{flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column', borderRadius: 2}}
        >
          <Box sx={{px: 2.5, pt: 2, pb: 0, flexShrink: 0}}>
            <Tabs
              value={editView}
              onChange={handleTabChange}
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
              <Divider sx={{mt: 1}} />
              <Box sx={{flex: 1, overflow: 'auto', px: 2.5, py: 2}}>
                <TranslationFieldsView
                  localValues={currentValues}
                  serverValues={serverValues}
                  search={search}
                  onChange={handleFieldChange}
                  onResetField={handleResetField}
                />
              </Box>
            </>
          )}

          {selectedLanguage && !isLoading && editView === 'json' && (
            <Box sx={{flex: 1, overflow: 'hidden', p: 0}}>
              <TranslationJsonEditor values={currentValues} colorMode={colorMode} onChange={handleJsonChange} />
            </Box>
          )}
        </Card>

        <Box sx={{width: 440, flexShrink: 0, display: 'flex', flexDirection: 'column', gap: 1}}>
          <Box
            sx={{
              flex: 1,
              overflow: 'hidden',
              borderRadius: 2,
              border: '1px solid',
              borderColor: 'divider',
              bgcolor: 'background.paper',
            }}
          >
            <TranslationPreviewPanel viewport={{width: '100%', height: '100%'}} />
          </Box>
        </Box>
      </Box>

      <Snackbar
        open={toast.open}
        autoHideDuration={3000}
        onClose={() => setToast((prev) => ({...prev, open: false}))}
        anchorOrigin={{vertical: 'bottom', horizontal: 'center'}}
      >
        <Alert severity={toast.severity} onClose={() => setToast((prev) => ({...prev, open: false}))}>
          {toast.message}
        </Alert>
      </Snackbar>
    </PageContent>
  );
}
