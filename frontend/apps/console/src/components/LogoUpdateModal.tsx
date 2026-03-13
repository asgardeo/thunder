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

import {useState, useCallback, useEffect} from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Stack,
  Avatar,
  TextField,
  Box,
  Typography,
  Grid,
  IconButton,
  FormControl,
  FormHelperText,
} from '@wso2/oxygen-ui';
import {X, AppWindow, Shuffle} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import generateAppLogoSuggestions from '../features/applications/utils/generateAppLogoSuggestion';

interface LogoUpdateModalProps {
  open: boolean;
  onClose: () => void;
  currentLogoUrl?: string;
  onLogoUpdate: (logoUrl: string) => void;
}

export default function LogoUpdateModal({open, onClose, currentLogoUrl = '', onLogoUpdate}: LogoUpdateModalProps) {
  const {t} = useTranslation();
  const [customUrl, setCustomUrl] = useState(currentLogoUrl);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [selectedLogo, setSelectedLogo] = useState<string>(currentLogoUrl);

  useEffect(() => {
    if (open) {
      // Generate new suggestions when modal opens
      setSuggestions(generateAppLogoSuggestions(12));
      setSelectedLogo(currentLogoUrl);
      setCustomUrl(currentLogoUrl);
    }
  }, [open, currentLogoUrl]);

  const handleRotateLogos = useCallback(() => {
    setSuggestions(generateAppLogoSuggestions(12));
  }, []);

  const handleSave = useCallback(() => {
    const logoToSave = customUrl || selectedLogo;
    if (logoToSave) {
      onLogoUpdate(logoToSave);
    }
  }, [customUrl, selectedLogo, onLogoUpdate]);

  const handleLogoSelect = useCallback((logoUrl: string) => {
    setSelectedLogo(logoUrl);
    setCustomUrl(''); // Clear custom URL when selecting from suggestions
  }, []);

  const handleCustomUrlChange = useCallback((url: string) => {
    setCustomUrl(url);
    if (url) {
      setSelectedLogo(''); // Clear selected logo when entering custom URL
    }
  }, []);

  const displayLogo = customUrl || selectedLogo || currentLogoUrl;

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Typography variant="h5">{t('applications:logoModal.title')}</Typography>
          <IconButton onClick={onClose} size="small">
            <X size={20} />
          </IconButton>
        </Stack>
      </DialogTitle>

      <DialogContent dividers>
        <Stack spacing={3}>
          {/* Preview */}
          <Box>
            <Typography variant="subtitle2" gutterBottom>
              {t('applications:logoModal.preview.title')}
            </Typography>
            <Box
              sx={{
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                py: 3,
                bgcolor: 'action.hover',
                borderRadius: 1,
              }}
            >
              <Avatar
                src={displayLogo}
                slotProps={{
                  img: {
                    onError: (e: React.SyntheticEvent<HTMLImageElement>) => {
                      e.currentTarget.style.display = 'none';
                    },
                  },
                }}
                sx={{
                  width: 120,
                  height: 120,
                }}
              >
                <AppWindow size={48} />
              </Avatar>
            </Box>
          </Box>

          {/* Custom URL Input */}
          <Box>
            <Typography variant="subtitle2" gutterBottom>
              {t('applications:logoModal.customUrl.title')}
            </Typography>
            <FormControl fullWidth>
              <TextField
                fullWidth
                placeholder={t('applications:logoModal.customUrl.placeholder')}
                value={customUrl}
                onChange={(e) => handleCustomUrlChange(e.target.value)}
              />
              <FormHelperText>{t('applications:logoModal.customUrl.hint')}</FormHelperText>
            </FormControl>
          </Box>

          {/* Suggestions */}
          <Box>
            <Stack direction="row" alignItems="center" justifyContent="space-between" mb={2}>
              <Typography variant="h6">{t('applications:logoModal.logos.title')}</Typography>
              <Button
                size="small"
                variant="text"
                startIcon={<Shuffle size={14} />}
                onClick={handleRotateLogos}
                sx={{minWidth: 'auto'}}
              >
                {t('applications:logoModal.logos.shuffle')}
              </Button>
            </Stack>
            <Grid container spacing={2}>
              {suggestions.map((logoUrl) => (
                <Grid size={{xs: 2}} key={logoUrl}>
                  <Box
                    onClick={() => handleLogoSelect(logoUrl)}
                    sx={{
                      cursor: 'pointer',
                      border: '2px solid',
                      borderColor: selectedLogo === logoUrl ? 'primary.main' : 'divider',
                      borderRadius: 1,
                      p: 1,
                      display: 'flex',
                      justifyContent: 'center',
                      alignItems: 'center',
                      bgcolor: 'background.paper',
                      transition: 'all 0.2s',
                      '&:hover': {
                        borderColor: 'primary.light',
                        bgcolor: 'action.hover',
                      },
                    }}
                  >
                    <Avatar
                      src={logoUrl}
                      slotProps={{
                        img: {
                          onError: (e: React.SyntheticEvent<HTMLImageElement>) => {
                            e.currentTarget.style.display = 'none';
                          },
                        },
                      }}
                      sx={{
                        width: 60,
                        height: 60,
                      }}
                    >
                      <AppWindow size={24} />
                    </Avatar>
                  </Box>
                </Grid>
              ))}
            </Grid>
          </Box>
        </Stack>
      </DialogContent>

      <DialogActions>
        <Button onClick={onClose} variant="outlined">
          {t('applications:logoModal.cancel')}
        </Button>
        <Button onClick={handleSave} variant="contained" disabled={!displayLogo}>
          {t('applications:logoModal.update')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
