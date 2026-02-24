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
  Box,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  FormHelperText,
  LinearProgress,
  Switch,
  TextField,
  Typography,
} from '@wso2/oxygen-ui';
import {useAsgardeo} from '@asgardeo/react';
import {useQueryClient} from '@tanstack/react-query';
import {I18nQueryKeys, enUS} from '@thunder/i18n';
import {useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';

export interface AddLanguageDialogProps {
  open: boolean;
  onClose: () => void;
  onAdded: (code: string) => void;
  serverUrl: string;
}

export default function AddLanguageDialog({open, onClose, onAdded, serverUrl}: AddLanguageDialogProps): JSX.Element {
  const {t} = useTranslation('translations');
  const {http} = useAsgardeo() as unknown as {
    http: {
      request: (config: {
        url: string;
        method: string;
        headers?: Record<string, string>;
        data?: string;
        attachToken?: boolean;
        withCredentials?: boolean;
      }) => Promise<{data: unknown}>;
    };
  };
  const queryClient = useQueryClient();

  const [languageCode, setLanguageCode] = useState('');
  const [populateFromEnglish, setPopulateFromEnglish] = useState(true);
  const [isAdding, setIsAdding] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const handleClose = () => {
    if (isAdding) return;
    setLanguageCode('');
    setPopulateFromEnglish(true);
    setError(null);
    setProgress(0);
    onClose();
  };

  const handleAdd = async () => {
    const code = languageCode.trim();
    if (!code) return;

    setIsAdding(true);
    setError(null);
    setProgress(0);

    // Collect all string keys across all namespaces from base English translations
    const entries: Array<{namespace: string; key: string; value: string}> = [];
    for (const [ns, nsValues] of Object.entries(enUS)) {
      for (const [key, value] of Object.entries(nsValues as Record<string, unknown>)) {
        if (typeof value === 'string') {
          entries.push({namespace: ns, key, value: populateFromEnglish ? value : ''});
        }
      }
    }

    let completed = 0;
    const total = entries.length;

    await Promise.allSettled(
      entries.map(async ({namespace, key, value}) => {
        try {
          await (
            http as {
              request: (config: unknown) => Promise<unknown>;
            }
          ).request({
            url: `${serverUrl}/i18n/languages/${code}/translations/ns/${namespace}/keys/${key}`,
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            data: JSON.stringify({value}),
          });
        } finally {
          completed++;
          setProgress(Math.round((completed / total) * 100));
        }
      }),
    );

    try {
      await queryClient.invalidateQueries({queryKey: [I18nQueryKeys.TRANSLATIONS]});
      await queryClient.invalidateQueries({queryKey: [I18nQueryKeys.LANGUAGES]});
      setIsAdding(false);
      setLanguageCode('');
      setPopulateFromEnglish(true);
      setProgress(0);
      onAdded(code);
    } catch {
      setError(t('language.add.error'));
      setIsAdding(false);
    }
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('language.add.dialogTitle')}</DialogTitle>

      <DialogContent>
        <Box sx={{display: 'flex', flexDirection: 'column', gap: 2, pt: 0.5}}>
          {error && <Alert severity="error">{error}</Alert>}

          <Box>
            <TextField
              label={t('language.add.codeLabel')}
              placeholder={t('language.add.codePlaceholder')}
              value={languageCode}
              onChange={(e) => setLanguageCode(e.target.value)}
              disabled={isAdding}
              fullWidth
              size="small"
            />
            <FormHelperText>{t('language.add.codeHelper')}</FormHelperText>
          </Box>

          <Box>
            <FormControlLabel
              control={
                <Switch
                  checked={populateFromEnglish}
                  onChange={(e) => setPopulateFromEnglish(e.target.checked)}
                  disabled={isAdding}
                />
              }
              label={t('language.add.populateLabel')}
            />
            <FormHelperText sx={{ml: 0}}>
              {populateFromEnglish ? t('language.add.populateHelper') : t('language.add.emptyHelper')}
            </FormHelperText>
          </Box>

          {isAdding && (
            <Box sx={{display: 'flex', flexDirection: 'column', gap: 1}}>
              <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
                <CircularProgress size={16} />
                <Typography variant="body2" color="text.secondary">
                  {t('language.add.adding')} ({progress}%)
                </Typography>
              </Box>
              <LinearProgress variant="determinate" value={progress} />
            </Box>
          )}
        </Box>
      </DialogContent>

      <DialogActions>
        <Button onClick={handleClose} disabled={isAdding}>
          {t('actions.cancel', {ns: 'common'})}
        </Button>
        <Button
          variant="contained"
          onClick={() => void handleAdd()}
          disabled={isAdding || !languageCode.trim()}
        >
          {t('language.add.dialogTitle')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
