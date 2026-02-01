/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import lowerCase from 'lodash-es/lowerCase';
import startCase from 'lodash-es/startCase';
import React, {type ChangeEvent, type ReactElement, type SyntheticEvent, useCallback, useMemo, useState} from 'react';
import {useTranslation} from 'react-i18next';
import {
  Alert,
  Autocomplete,
  type AutocompleteRenderInputParams,
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  CircularProgress,
  Divider,
  IconButton,
  Popover,
  TextField,
  Tooltip,
  Typography,
} from '@wso2/oxygen-ui';
import {PlusIcon, XIcon} from '@wso2/oxygen-ui-icons-react';
import {useGetLanguages, useGetTranslations, useUpdateTranslation} from '@thunder/i18n';
import {useConfig} from '@thunder/shared-contexts';
import {invalidateI18nCache} from '../../../../i18n/invalidate-i18n-cache';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';

/**
 * Props interface for the language text field component.
 */
export interface LanguageTextFieldProps {
  value: string;
  onChange: (event: ChangeEvent<HTMLInputElement>) => void;
  disabled?: boolean;
}

/**
 * Props interface of {@link I18nConfigurationCard}
 */
export interface I18nConfigurationCardPropsInterface {
  open: boolean;
  anchorEl: HTMLElement | null;
  propertyKey: string;
  onClose: () => void;
  i18nKey: string;
  onChange: (i18nKey: string) => void;
}

/**
 * Default namespace for new translations.
 */
const DEFAULT_NAMESPACE = 'flowI18n';

/**
 * Default language for new translations.
 */
const DEFAULT_LANGUAGE = 'en-US';

/**
 * I18n configuration floating card component.
 * Provides a dropdown to select i18n keys and displays the resolved translation value.
 * Also allows creating new translations via the i18n API.
 */
function I18nConfigurationCard({
  open,
  anchorEl,
  propertyKey,
  onClose,
  onChange,
  i18nKey: selectedI18nKey,
}: I18nConfigurationCardPropsInterface): ReactElement {
  const {t} = useTranslation();
  const {getServerUrl} = useConfig();
  const serverUrl = getServerUrl();
  const {i18nText, i18nTextLoading} = useFlowBuilderCore();
  const updateTranslation = useUpdateTranslation({
    serverUrl,
    onMutationSuccess: () => {
      invalidateI18nCache();
    },
  });
  const {data: languagesData} = useGetLanguages({serverUrl});
  const {data: translationsData, isLoading: translationsLoading} = useGetTranslations({
    serverUrl,
    language: DEFAULT_LANGUAGE,
    namespace: DEFAULT_NAMESPACE,
    enabled: open,
  });

  // State for create mode
  const [isCreateMode, setIsCreateMode] = useState<boolean>(false);
  const [newKey, setNewKey] = useState<string>('');
  const [newTranslationValue, setNewTranslationValue] = useState<string>('');
  const [selectedLanguage, setSelectedLanguage] = useState<string>(DEFAULT_LANGUAGE);
  const [error, setError] = useState<string | null>(null);

  /**
   * Get the list of available i18n keys from the flowI18n namespace.
   */
  const availableI18nKeys: string[] = useMemo(() => {
    if (!translationsData?.translations) {
      return [];
    }

    const keys: string[] = [];

    Object.values(translationsData.translations).forEach((namespaceTexts: Record<string, string>) => {
      keys.push(...Object.keys(namespaceTexts).map((key: string) => `${DEFAULT_NAMESPACE}:${key}`));
    });

    return keys;
  }, [translationsData]);

  /**
   * Get the resolved value for the selected i18n key.
   */
  const resolvedValue: string = useMemo(() => {
    if (!selectedI18nKey || !i18nText) {
      return '';
    }

    // Search for the key across all screens
    let foundValue = '';

    Object.values(i18nText).some((screenTexts: Record<string, string>) => {
      if (screenTexts[selectedI18nKey]) {
        foundValue = screenTexts[selectedI18nKey];
        return true;
      }
      return false;
    });

    return foundValue;
  }, [selectedI18nKey, i18nText]);

  /**
   * Available languages from the API or default list.
   */
  const availableLanguages: string[] = useMemo(() => {
    if (languagesData?.languages && languagesData.languages.length > 0) {
      return languagesData.languages;
    }
    return [DEFAULT_LANGUAGE];
  }, [languagesData]);

  /**
   * Reset create mode form.
   */
  const resetCreateForm = useCallback(() => {
    setNewKey('');
    setNewTranslationValue('');
    setSelectedLanguage(DEFAULT_LANGUAGE);
    setError(null);
  }, []);

  /**
   * Handle entering create mode.
   */
  const handleEnterCreateMode = useCallback(() => {
    setIsCreateMode(true);
    resetCreateForm();
  }, [resetCreateForm]);

  /**
   * Handle exiting create mode.
   */
  const handleExitCreateMode = useCallback(() => {
    setIsCreateMode(false);
    resetCreateForm();
  }, [resetCreateForm]);

  /**
   * Handle creating a new translation.
   */
  const handleCreateTranslation = useCallback(() => {
    if (!newKey.trim()) {
      setError(t('common:validation.required', {field: t('flows:core.elements.textPropertyField.i18nCard.i18nKey')}));
      return;
    }

    if (!newTranslationValue.trim()) {
      setError(
        t('common:validation.required', {field: t('flows:core.elements.textPropertyField.i18nCard.languageText')}),
      );
      return;
    }

    // Validate key format (alphanumeric, dots, underscores, hyphens)
    const keyPattern = /^[a-zA-Z0-9._-]+$/;
    if (!keyPattern.test(newKey)) {
      setError(t('flows:core.elements.textPropertyField.i18nCard.invalidKeyFormat'));
      return;
    }

    updateTranslation.mutate(
      {
        language: selectedLanguage,
        namespace: DEFAULT_NAMESPACE,
        key: newKey,
        value: newTranslationValue,
      },
      {
        onSuccess: () => {
          // On success, use the new key with namespace prefix and close create mode
          onChange(`${DEFAULT_NAMESPACE}:${newKey}`);
          handleExitCreateMode();
        },
        onError: (err: Error) => {
          setError(err.message || t('common:errors.unknown'));
        },
      },
    );
  }, [newKey, newTranslationValue, selectedLanguage, updateTranslation, onChange, handleExitCreateMode, t]);

  /**
   * Handle close and reset state.
   */
  const handleClose = useCallback(() => {
    handleExitCreateMode();
    onClose();
  }, [handleExitCreateMode, onClose]);

  /**
   * Renders the loading state content.
   */
  const renderLoadingContent = (): ReactElement => (
    <Box sx={{display: 'flex', justifyContent: 'center', p: 2}}>
      <CircularProgress size={20} />
    </Box>
  );

  /**
   * Renders the create mode content.
   */
  const renderCreateModeContent = (): ReactElement => (
    <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
      {error && (
        <Alert severity="error" onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      <div>
        <Typography variant="subtitle2" gutterBottom>
          {t('flows:core.elements.textPropertyField.i18nCard.language')}
        </Typography>
        <Autocomplete
          options={availableLanguages}
          value={selectedLanguage}
          onChange={(_event: SyntheticEvent, newLang: string | null) => {
            setSelectedLanguage(newLang ?? DEFAULT_LANGUAGE);
          }}
          renderInput={(params: AutocompleteRenderInputParams) => <TextField {...params} size="small" />}
          disableClearable
        />
      </div>

      <div>
        <Typography variant="subtitle2" gutterBottom>
          {t('flows:core.elements.textPropertyField.i18nCard.i18nKey')}
        </Typography>
        <TextField
          fullWidth
          size="small"
          value={newKey}
          onChange={(e: ChangeEvent<HTMLInputElement>) => {
            setNewKey(e.target.value);
            if (error) setError(null);
          }}
          placeholder={t('flows:core.elements.textPropertyField.i18nCard.i18nKeyInputPlaceholder')}
          helperText={t('flows:core.elements.textPropertyField.i18nCard.i18nKeyInputHint', {key: propertyKey})}
        />
      </div>

      <div>
        <Typography variant="subtitle2" gutterBottom>
          {t('flows:core.elements.textPropertyField.i18nCard.languageText')}
        </Typography>
        <TextField
          fullWidth
          size="small"
          multiline
          rows={3}
          value={newTranslationValue}
          onChange={(e: ChangeEvent<HTMLInputElement>) => {
            setNewTranslationValue(e.target.value);
            if (error) setError(null);
          }}
          placeholder={t('flows:core.elements.textPropertyField.i18nCard.languageTextPlaceholder')}
        />
      </div>

      <Box sx={{display: 'flex', gap: 1, justifyContent: 'flex-end'}}>
        <Button variant="text" onClick={handleExitCreateMode}>
          {t('common:cancel')}
        </Button>
        <Button
          variant="contained"
          onClick={handleCreateTranslation}
          disabled={updateTranslation.isPending || !newKey.trim() || !newTranslationValue.trim()}
        >
          {updateTranslation.isPending ? <CircularProgress size={16} /> : t('common:create')}
        </Button>
      </Box>
    </Box>
  );

  /**
   * Renders the select mode content.
   */
  const renderSelectModeContent = (): ReactElement => (
    <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
      <div>
        <Typography variant="subtitle2" gutterBottom>
          {t('flows:core.elements.textPropertyField.i18nCard.i18nKey')}
        </Typography>
        <Autocomplete
          options={availableI18nKeys}
          value={selectedI18nKey === '' ? null : selectedI18nKey}
          onChange={(_event: SyntheticEvent, selectedValue: string | null) => {
            onChange(selectedValue ?? '');
          }}
          renderInput={(params: AutocompleteRenderInputParams) => (
            <TextField
              {...params}
              placeholder={t('flows:core.elements.textPropertyField.i18nCard.selectI18nKey')}
              size="small"
            />
          )}
          renderOption={({key, ...props}: React.HTMLAttributes<HTMLLIElement> & {key: string}, option: string) => (
            <li key={key} {...props}>
              <Tooltip title={option} placement="bottom">
                <span>{option}</span>
              </Tooltip>
            </li>
          )}
        />
      </div>

      {selectedI18nKey && resolvedValue && (
        <Box
          sx={{
            p: 1.5,
            backgroundColor: 'action.hover',
            borderRadius: 1,
            border: '1px solid',
            borderColor: 'divider',
          }}
        >
          <Typography variant="caption" color="text.secondary" sx={{display: 'block', mb: 0.5}}>
            {t('flows:core.elements.textPropertyField.resolvedValue')}
          </Typography>
          <Typography variant="body2" sx={{wordBreak: 'break-word'}}>
            {resolvedValue}
          </Typography>
        </Box>
      )}

      <Divider />

      <Box sx={{display: 'flex', alignItems: 'center', justifyContent: 'center'}}>
        <Tooltip title={t('flows:core.elements.textPropertyField.i18nCard.tooltip.addNewTranslation')}>
          <Button variant="text" startIcon={<PlusIcon size={16} />} onClick={handleEnterCreateMode}>
            {t('flows:core.elements.textPropertyField.i18nCard.createTitle')}
          </Button>
        </Tooltip>
      </Box>
    </Box>
  );

  /**
   * Renders the card content based on the current state.
   */
  const renderCardContent = (): ReactElement => {
    if (i18nTextLoading || translationsLoading) {
      return renderLoadingContent();
    }

    if (isCreateMode) {
      return renderCreateModeContent();
    }

    return renderSelectModeContent();
  };

  return (
    <Popover
      open={open}
      anchorEl={anchorEl}
      onClose={handleClose}
      anchorOrigin={{
        vertical: 'top',
        horizontal: 'right',
      }}
      transformOrigin={{
        vertical: 'top',
        horizontal: 'left',
      }}
    >
      <Card
        sx={{
          width: 400,
        }}
      >
        <CardHeader
          title={
            isCreateMode
              ? t('flows:core.elements.textPropertyField.i18nCard.createTitle')
              : t('flows:core.elements.textPropertyField.i18nCard.title', {
                  field: startCase(lowerCase(propertyKey)),
                })
          }
          action={
            <IconButton aria-label={t('common:close')} onClick={handleClose} size="small">
              <XIcon />
            </IconButton>
          }
        />
        <CardContent>{renderCardContent()}</CardContent>
      </Card>
    </Popover>
  );
}

export default I18nConfigurationCard;
