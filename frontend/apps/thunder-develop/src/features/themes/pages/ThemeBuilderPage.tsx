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

import {useEffect, useRef, useState, type JSX} from 'react';
import {useNavigate, useParams} from 'react-router';
import {Box, Button, Chip, CircularProgress, Grid, Paper, Stack, Tooltip, Typography} from '@wso2/oxygen-ui';
import {ArrowLeft, Save} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger';
import useDebounce from '@/features/design-system/hooks/useDebounce';
import useNotification from '@/features/design-system/hooks/useNotification';
import useUnsavedChanges from '@/features/design-system/hooks/useUnsavedChanges';
import useKeyboardShortcuts from '@/features/design-system/hooks/useKeyboardShortcuts';
import UnsavedChangesDialog from '@/features/design-system/components/UnsavedChangesDialog';
import {defaultThemeFormData, type ThemeFormData} from '../schemas/themeSchema';
import useGetTheme from '../api/useGetTheme';
import useCreateTheme from '../api/useCreateTheme';
import useUpdateTheme from '../api/useUpdateTheme';
import ThemeBuilderForm from '../components/ThemeBuilderForm';
import ThemePreview from '../components/ThemePreview';

/**
 * Page component for theme builder with unsaved changes detection,
 * keyboard shortcuts, and real-time preview.
 */
export default function ThemeBuilderPage(): JSX.Element {
  const {t} = useTranslation();
  const navigate = useNavigate();
  const logger = useLogger('ThemeBuilderPage');
  const {showSuccess, showError} = useNotification();
  const {id: themeId} = useParams<{id: string}>();
  const isEditMode = Boolean(themeId);

  const {data: existingTheme, isLoading: isLoadingTheme} = useGetTheme(themeId ?? '');
  const {mutate: createTheme, isPending: isCreating} = useCreateTheme();
  const {mutate: updateTheme, isPending: isUpdating} = useUpdateTheme();

  const [formData, setFormData] = useState<ThemeFormData>(defaultThemeFormData);
  const [isDirty, setIsDirty] = useState(false);
  const formRef = useRef<{submit: () => void}>();

  // Unsaved changes protection
  const {showDialog, showUnsavedDialog, confirmNavigation, cancelNavigation} = useUnsavedChanges(isDirty);

  // Debounce form data for preview (reduce re-renders)
  const debouncedFormData = useDebounce(formData, 300);

  // Load existing theme data
  useEffect(() => {
    if (existingTheme) {
      setFormData({
        displayName: existingTheme.displayName,
        theme: existingTheme.theme,
      });
    }
  }, [existingTheme]);

  // Keyboard shortcuts
  useKeyboardShortcuts([
    {
      key: 's',
      meta: true,
      callback: (e) => {
        e.preventDefault();
        if (isDirty && !isSaving) {
          handleSave(formData);
        }
      },
      description: 'Save theme',
    },
  ]);

  const handleFormSubmit = (data: ThemeFormData) => {
    handleSave(data);
  };

  const handleSave = (data: ThemeFormData) => {
    if (isEditMode && themeId) {
      updateTheme(
        {themeId, themeData: data},
        {
          onSuccess: () => {
            showSuccess(
              t('themes:notifications.updateSuccess', {
                name: data.displayName,
                defaultValue: `Theme "${data.displayName}" updated successfully`,
              })
            );
            setIsDirty(false);
            logger.info('Theme updated successfully', {themeId});
            (async () => {
              await navigate('/themes');
            })().catch((navError: unknown) => {
              logger.error('Navigation failed', {error: navError});
            });
          },
          onError: (error) => {
            showError(t('themes:notifications.updateError', {defaultValue: 'Failed to update theme'}));
            logger.error('Failed to update theme', {error});
          },
        }
      );
    } else {
      createTheme(data, {
        onSuccess: () => {
          showSuccess(
            t('themes:notifications.createSuccess', {
              name: data.displayName,
              defaultValue: `Theme "${data.displayName}" created successfully`,
            })
          );
          logger.info('Theme created successfully');
          (async () => {
            await navigate('/themes');
          })().catch((navError: unknown) => {
            logger.error('Navigation failed', {error: navError});
          });
        },
        onError: (error) => {
          showError(t('themes:notifications.createError', {defaultValue: 'Failed to create theme'}));
          logger.error('Failed to create theme', {error});
        },
      });
    }
  };

  const handleSaveAndNavigate = () => {
    handleSave(formData);
    // Navigation will happen in handleSave's onSuccess
  };

  const handleBack = () => {
    if (isDirty) {
      // Show unsaved changes dialog
      showUnsavedDialog();
    } else {
      // Navigate directly if no changes
      (async () => {
        await navigate('/themes');
      })().catch((navError: unknown) => {
        logger.error('Navigation failed', {error: navError});
      });
    }
  };

  if (isLoadingTheme && isEditMode) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  const isSaving = isCreating || isUpdating;

  return (
    <>
      <Box sx={{height: '100%', display: 'flex', flexDirection: 'column'}}>
        {/* Header */}
        <Paper elevation={1} sx={{p: 2, mb: 0, borderRadius: 0}}>
          <Stack direction="row" justifyContent="space-between" alignItems="center">
            <Stack direction="row" spacing={2} alignItems="center">
              <Button variant="text" startIcon={<ArrowLeft size={18} />} onClick={handleBack}>
                {t('common:actions.back', {defaultValue: 'Back'})}
              </Button>
              <Typography variant="h5">
                {isEditMode
                  ? t('themes:builder.titleEdit', {defaultValue: 'Edit Theme'})
                  : t('themes:builder.titleCreate', {defaultValue: 'Create Theme'})}
              </Typography>
              {isDirty && (
                <Chip
                  label={t('common:unsavedChanges.label', {defaultValue: 'Unsaved changes'})}
                  size="small"
                  color="warning"
                />
              )}
            </Stack>
            <Tooltip title={isDirty ? 'Cmd+S' : ''}>
              <Button
                variant="contained"
                startIcon={<Save size={18} />}
                onClick={() => handleSave(formData)}
                disabled={isSaving || !isDirty}
              >
                {t('common:actions.save', {defaultValue: 'Save'})}
              </Button>
            </Tooltip>
          </Stack>
        </Paper>

        {/* Form and Preview */}
        <Box sx={{flex: 1, overflow: 'hidden', p: 3}}>
          <Grid container spacing={3} sx={{height: '100%'}}>
            <Grid item xs={12} md={6} sx={{height: '100%', overflow: 'auto'}}>
              <ThemeBuilderForm
                initialValues={formData}
                onSubmit={handleFormSubmit}
                onValuesChange={setFormData}
                onDirtyChange={setIsDirty}
                isSubmitting={isSaving}
              />
            </Grid>
            <Grid item xs={12} md={6} sx={{height: '100%', overflow: 'auto'}}>
              <Paper elevation={2} sx={{p: 3, height: '100%', position: 'sticky', top: 0}}>
                <Typography variant="h6" gutterBottom>
                  {t('themes:builder.preview', {defaultValue: 'Preview'})}
                </Typography>
                <ThemePreview themeConfig={debouncedFormData.theme} />
              </Paper>
            </Grid>
          </Grid>
        </Box>
      </Box>

      {/* Unsaved changes dialog */}
      <UnsavedChangesDialog
        open={showDialog}
        onSave={handleSaveAndNavigate}
        onDiscard={() => {
          confirmNavigation();
          (async () => {
            await navigate('/themes');
          })().catch((navError: unknown) => {
            logger.error('Navigation failed', {error: navError});
          });
        }}
        onCancel={cancelNavigation}
        message={t('common:unsavedChanges.message', {
          defaultValue: 'You have unsaved changes. Do you want to save them?',
        })}
      />
    </>
  );
}
