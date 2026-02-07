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

import {useState, type JSX} from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  RadioGroup,
  FormControlLabel,
  Radio,
  Typography,
  Box,
} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';

/**
 * Props for ThemeSaveDialog component
 */
interface ThemeSaveDialogProps {
  /** Whether the dialog is open */
  open: boolean;
  /** Callback when dialog is closed */
  onClose: () => void;
  /** Callback when save is confirmed */
  onSave: (name: string, override: boolean) => void;
  /** Name of the base theme being edited (if any) */
  baseThemeName?: string;
  /** Current theme name */
  currentName?: string;
  /** Whether save is in progress */
  isSaving?: boolean;
}

/**
 * Dialog for saving theme with name and override options
 */
export default function ThemeSaveDialog({
  open,
  onClose,
  onSave,
  baseThemeName,
  currentName = '',
  isSaving = false,
}: ThemeSaveDialogProps): JSX.Element {
  const {t} = useTranslation();
  const [themeName, setThemeName] = useState(currentName);
  const [saveMode, setSaveMode] = useState<'new' | 'override'>(baseThemeName ? 'new' : 'new');

  const handleSave = () => {
    const finalName = saveMode === 'override' && baseThemeName ? baseThemeName : themeName;
    onSave(finalName, saveMode === 'override');
  };

  const handleClose = () => {
    if (!isSaving) {
      onClose();
    }
  };

  const isValid = saveMode === 'override' || themeName.trim().length > 0;

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('design:theme.saveTheme', {defaultValue: 'Save Theme'})}</DialogTitle>
      <DialogContent>
        {baseThemeName && (
          <Box sx={{mb: 3}}>
            <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>
              {t('design:theme.baseTheme', {defaultValue: 'Based on: {{name}}', name: baseThemeName})}
            </Typography>
            <RadioGroup value={saveMode} onChange={(e) => setSaveMode(e.target.value as 'new' | 'override')}>
              <FormControlLabel
                value="new"
                control={<Radio />}
                label={t('design:theme.saveAsNew', {defaultValue: 'Save as new theme'})}
              />
              <FormControlLabel
                value="override"
                control={<Radio />}
                label={t('design:theme.overrideExisting', {
                  defaultValue: 'Override "{{name}}"',
                  name: baseThemeName,
                })}
              />
            </RadioGroup>
          </Box>
        )}

        {saveMode === 'new' && (
          <TextField
            autoFocus
            fullWidth
            label={t('design:theme.themeName', {defaultValue: 'Theme Name'})}
            value={themeName}
            onChange={(e) => setThemeName(e.target.value)}
            placeholder={t('design:theme.enterThemeName', {defaultValue: 'Enter theme name'})}
            disabled={isSaving}
            helperText={t('design:theme.themeNameHelper', {
              defaultValue: 'Give your theme a descriptive name',
            })}
          />
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={isSaving}>
          {t('common:actions.cancel', {defaultValue: 'Cancel'})}
        </Button>
        <Button variant="contained" onClick={handleSave} disabled={!isValid || isSaving}>
          {isSaving
            ? t('common:actions.saving', {defaultValue: 'Saving...'})
            : t('common:actions.save', {defaultValue: 'Save'})}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
