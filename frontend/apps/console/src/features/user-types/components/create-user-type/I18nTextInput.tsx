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

import {type ChangeEvent, type ReactElement, type SyntheticEvent, useCallback, useMemo, useRef, useState} from 'react';
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
  FormControl,
  FormLabel,
  IconButton,
  InputAdornment,
  Popover,
  TextField,
  Tooltip,
  Typography,
} from '@wso2/oxygen-ui';
import {PlusIcon, SquareFunction, XIcon} from '@wso2/oxygen-ui-icons-react';
import {useGetLanguages, useGetTranslations, useUpdateTranslation, NamespaceConstants} from '@thunder/i18n';
import {isI18nTemplatePattern, I18N_KEY_PATTERN} from '@thunder/utils';
import {useTemplateLiteralResolver} from '@thunder/shared-hooks';
import {invalidateI18nCache} from '../../../../i18n/invalidate-i18n-cache';

const DEFAULT_LANGUAGE = 'en-US';

/**
 * Sanitizes a string for use as a translation key.
 * Replaces spaces with underscores, lowercases, and strips invalid characters.
 */
function sanitizeTranslationKey(key: string): string {
  return key
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '_')
    .replace(/[^a-zA-Z0-9._-]/g, '');
}

/**
 * Props interface of {@link I18nTextInput}
 */
export interface I18nTextInputProps {
  label: string;
  value: string;
  onChange: (newValue: string) => void;
  placeholder?: string;
  defaultNewKey?: string;
}

/**
 * Props for the i18n content component.
 */
interface I18nContentProps {
  i18nKey: string;
  isActive: boolean;
  isCreateMode: boolean;
  onChange: (key: string) => void;
  onCreateModeChange: (isCreateMode: boolean) => void;
  defaultNewKey?: string;
}

/**
 * Content component for the i18n popover with select and create modes.
 */
function I18nContent({i18nKey, isActive, isCreateMode, onChange, onCreateModeChange, defaultNewKey = undefined}: I18nContentProps): ReactElement {
  const {t, i18n} = useTranslation();
  const {data: languagesData} = useGetLanguages();
  const {data: translationsData, isLoading} = useGetTranslations({
    language: DEFAULT_LANGUAGE,
    namespace: NamespaceConstants.CUSTOM_NAMESPACE,
    enabled: isActive,
  });
  const updateTranslation = useUpdateTranslation({
    onMutationSuccess: () => {
      invalidateI18nCache();
    },
  });

  const sanitizedDefaultKey = defaultNewKey ? sanitizeTranslationKey(defaultNewKey) : '';
  const [newKey, setNewKey] = useState(sanitizedDefaultKey);
  const [newValue, setNewValue] = useState('');
  const [selectedLanguage, setSelectedLanguage] = useState(DEFAULT_LANGUAGE);
  const [error, setError] = useState<string | null>(null);

  const availableKeys = useMemo(() => {
    if (!translationsData?.translations) return [];

    const keys: string[] = [];
    Object.entries(translationsData.translations).forEach(
      ([namespace, translations]: [string, Record<string, string>]) => {
        keys.push(...Object.keys(translations).map((key: string) => `${namespace}:${key}`));
      },
    );
    return keys;
  }, [translationsData]);

  const resolvedValue = useMemo(() => {
    if (!i18nKey || !translationsData?.translations) return '';

    // i18nKey may be namespaced (e.g. "custom:myKey") — split and look up in the correct namespace
    const colonIdx = i18nKey.indexOf(':');
    if (colonIdx !== -1) {
      const ns = i18nKey.slice(0, colonIdx);
      const bareKey = i18nKey.slice(colonIdx + 1);
      return translationsData.translations[ns]?.[bareKey] ?? '';
    }

    // Bare key — search across all namespaces
    let found = '';
    Object.values(translationsData.translations).some((translations: Record<string, string>) => {
      if (translations[i18nKey]) {
        found = translations[i18nKey];
        return true;
      }
      return false;
    });
    return found;
  }, [i18nKey, translationsData]);

  const availableLanguages = useMemo(() => {
    if (languagesData?.languages && languagesData.languages.length > 0) {
      return languagesData.languages;
    }
    return [DEFAULT_LANGUAGE];
  }, [languagesData]);

  const resetCreateForm = useCallback(() => {
    setNewKey(sanitizedDefaultKey);
    setNewValue('');
    setSelectedLanguage(DEFAULT_LANGUAGE);
    setError(null);
  }, [sanitizedDefaultKey]);

  const handleCreate = useCallback(() => {
    if (!newKey.trim()) {
      setError(t('userTypes:displayNameI18n.keyRequired', 'Translation key is required'));
      return;
    }
    if (!newValue.trim()) {
      setError(t('userTypes:displayNameI18n.valueRequired', 'Translation value is required'));
      return;
    }
    if (!/^[a-zA-Z0-9._-]+$/.test(newKey)) {
      setError(t('userTypes:displayNameI18n.invalidKeyFormat', 'Key may only contain letters, numbers, dots, hyphens, and underscores'));
      return;
    }

    updateTranslation.mutate(
      {
        language: selectedLanguage,
        namespace: NamespaceConstants.CUSTOM_NAMESPACE,
        key: newKey,
        value: newValue,
      },
      {
        onSuccess: () => {
          // Synchronously add the new translation to i18next so t() can resolve it
          // immediately when the parent re-renders (before the async I18nProvider refresh)
          i18n.addResourceBundle(
            selectedLanguage,
            NamespaceConstants.CUSTOM_NAMESPACE,
            {[newKey]: newValue},
            true,
            true,
          );
          onChange(`${NamespaceConstants.CUSTOM_NAMESPACE}:${newKey}`);
          onCreateModeChange(false);
          resetCreateForm();
        },
        onError: (err: Error) => {
          setError(err.message ?? t('common:errors.unknown'));
        },
      },
    );
  }, [newKey, newValue, selectedLanguage, updateTranslation, onChange, onCreateModeChange, resetCreateForm, t, i18n]);

  if (isLoading) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', p: 2}}>
        <CircularProgress size={20} />
      </Box>
    );
  }

  if (isCreateMode) {
    return (
      <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
        {error && (
          <Alert severity="error" onClose={() => setError(null)}>
            {error}
          </Alert>
        )}
        <div>
          <Typography variant="subtitle2" gutterBottom>
            {t('userTypes:displayNameI18n.language', 'Language')}
          </Typography>
          <Autocomplete
            options={availableLanguages}
            value={selectedLanguage}
            onChange={(_e: SyntheticEvent, newLang: string | null) =>
              setSelectedLanguage(newLang ?? DEFAULT_LANGUAGE)
            }
            renderInput={(params: AutocompleteRenderInputParams) => <TextField {...params} size="small" />}
            disableClearable
          />
        </div>
        <div>
          <Typography variant="subtitle2" gutterBottom>
            {t('userTypes:displayNameI18n.i18nKey', 'Translation Key')}
          </Typography>
          <TextField
            fullWidth
            size="small"
            value={newKey}
            onChange={(e: ChangeEvent<HTMLInputElement>) => {
              setNewKey(e.target.value);
              if (error) setError(null);
            }}
            placeholder="e.g., user.firstName"
          />
        </div>
        <div>
          <Typography variant="subtitle2" gutterBottom>
            {t('userTypes:displayNameI18n.translationValue', 'Translation Value')}
          </Typography>
          <TextField
            fullWidth
            size="small"
            multiline
            rows={2}
            value={newValue}
            onChange={(e: ChangeEvent<HTMLInputElement>) => {
              setNewValue(e.target.value);
              if (error) setError(null);
            }}
            placeholder="e.g., First Name"
          />
        </div>
        <Box sx={{display: 'flex', gap: 1, justifyContent: 'flex-end'}}>
          <Button
            variant="text"
            onClick={() => {
              onCreateModeChange(false);
              resetCreateForm();
            }}
          >
            {t('common:cancel')}
          </Button>
          <Button
            variant="contained"
            onClick={handleCreate}
            disabled={updateTranslation.isPending || !newKey.trim() || !newValue.trim()}
          >
            {updateTranslation.isPending ? <CircularProgress size={16} /> : t('common:create')}
          </Button>
        </Box>
      </Box>
    );
  }

  return (
    <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
      <div>
        <Typography variant="subtitle2" gutterBottom>
          {t('userTypes:displayNameI18n.i18nKey', 'Translation Key')}
        </Typography>
        <Autocomplete
          options={availableKeys}
          value={i18nKey === '' ? null : i18nKey}
          onChange={(_e: SyntheticEvent, selected: string | null) => onChange(selected ?? '')}
          renderInput={(params: AutocompleteRenderInputParams) => (
            <TextField
              {...params}
              placeholder={t('userTypes:displayNameI18n.selectKey', 'Select a translation key')}
              size="small"
            />
          )}
          renderOption={({key, ...props}: React.HTMLAttributes<HTMLLIElement> & {key: string}, option: string) => (
            <li key={key} {...props}>
              <span>{option}</span>
            </li>
          )}
        />
      </div>

      {i18nKey && resolvedValue && (
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
            {t('userTypes:displayNameI18n.resolvedValue', 'Resolved value')}
          </Typography>
          <Typography variant="body2" sx={{wordBreak: 'break-word'}}>
            {resolvedValue}
          </Typography>
        </Box>
      )}

      <Divider />

      <Box sx={{display: 'flex', alignItems: 'center', justifyContent: 'center'}}>
        <Tooltip title={t('userTypes:displayNameI18n.createTooltip', 'Create a new translation key')}>
          <Button variant="text" startIcon={<PlusIcon size={16} />} onClick={() => onCreateModeChange(true)}>
            {t('userTypes:displayNameI18n.createTitle', 'Create New Translation')}
          </Button>
        </Tooltip>
      </Box>
    </Box>
  );
}

/**
 * Props for the i18n popover component.
 */
interface I18nPopoverProps {
  open: boolean;
  anchorEl: HTMLElement | null;
  onClose: () => void;
  value: string;
  onChange: (newValue: string) => void;
  defaultNewKey?: string;
}

/**
 * Popover with i18n key selection and creation UI.
 */
function I18nPopover({open, anchorEl, onClose, value, onChange, defaultNewKey = undefined}: I18nPopoverProps): ReactElement {
  const {t} = useTranslation();
  const [isCreateMode, setIsCreateMode] = useState(false);

  const handleClose = useCallback(() => {
    setIsCreateMode(false);
    onClose();
  }, [onClose]);

  const i18nKey = useMemo(() => I18N_KEY_PATTERN.exec(value.trim())?.[1] ?? '', [value]);

  const handleChange = useCallback(
    (key: string) => {
      onChange(key ? `{{t(${key})}}` : '');
    },
    [onChange],
  );

  return (
    <Popover
      open={open}
      anchorEl={anchorEl}
      onClose={handleClose}
      anchorOrigin={{vertical: 'top', horizontal: 'right'}}
      transformOrigin={{vertical: 'top', horizontal: 'left'}}
    >
      <Card sx={{width: 400}}>
        <CardHeader
          title={
            isCreateMode
              ? t('userTypes:displayNameI18n.createTitle', 'Create New Translation')
              : t('userTypes:displayNameI18n.title', 'Translation')
          }
          action={
            <IconButton aria-label={t('common:close')} onClick={handleClose} size="small">
              <XIcon />
            </IconButton>
          }
        />
        <CardContent>
          <I18nContent
            i18nKey={i18nKey}
            isActive={open}
            isCreateMode={isCreateMode}
            onChange={handleChange}
            onCreateModeChange={setIsCreateMode}
            defaultNewKey={defaultNewKey}
          />
        </CardContent>
      </Card>
    </Popover>
  );
}

/**
 * A text input field with an i18n button that opens a popover for selecting
 * or creating translation keys. Similar to the flow builder's TextPropertyField
 * but standalone (no flow builder context dependency).
 */
export default function I18nTextInput({label, value, onChange, placeholder = undefined, defaultNewKey = undefined}: I18nTextInputProps): ReactElement {
  const {t} = useTranslation();
  const {resolve} = useTemplateLiteralResolver();
  const iconButtonRef = useRef<HTMLButtonElement>(null);
  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  const isDynamic = isI18nTemplatePattern(value);
  const resolvedValue = isDynamic ? resolve(value, {t}) ?? '' : '';

  return (
    <FormControl fullWidth>
      <FormLabel>{label}</FormLabel>
      <TextField
        fullWidth
        value={value}
        onChange={(e: ChangeEvent<HTMLInputElement>) => onChange(e.target.value)}
        placeholder={placeholder}
        size="small"
        sx={
          isDynamic
            ? {
                '& .MuiOutlinedInput-root': {
                  backgroundColor: 'rgba(var(--mui-palette-primary-mainChannel) / 0.1)',
                  '& fieldset': {borderColor: 'primary.main'},
                  '&:hover fieldset': {borderColor: 'primary.dark'},
                  '&.Mui-focused fieldset': {borderColor: 'primary.main'},
                },
              }
            : undefined
        }
        InputProps={{
          endAdornment: (
            <InputAdornment position="end">
              <Tooltip title={t('userTypes:displayNameI18n.tooltip', 'Configure translation')}>
                <IconButton
                  ref={iconButtonRef}
                  onClick={() => setIsPopoverOpen(!isPopoverOpen)}
                  size="small"
                  edge="end"
                  color={isDynamic ? 'primary' : 'default'}
                >
                  <SquareFunction size={16} />
                </IconButton>
              </Tooltip>
            </InputAdornment>
          ),
        }}
      />
      {isDynamic && resolvedValue && (
        <Box
          sx={{
            mt: 1,
            p: 1.5,
            backgroundColor: 'action.hover',
            borderRadius: 1,
            border: '1px solid',
            borderColor: 'divider',
          }}
        >
          <Typography variant="caption" color="text.secondary" sx={{display: 'block', mb: 0.5}}>
            {t('userTypes:displayNameI18n.resolvedValue', 'Resolved value')}
          </Typography>
          <Typography variant="body2" sx={{wordBreak: 'break-word'}}>
            {resolvedValue}
          </Typography>
        </Box>
      )}
      <I18nPopover
        open={isPopoverOpen}
        anchorEl={iconButtonRef.current}
        onClose={() => setIsPopoverOpen(false)}
        value={value}
        onChange={onChange}
        defaultNewKey={defaultNewKey}
      />
    </FormControl>
  );
}
