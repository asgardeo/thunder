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

import {useMemo, useState, type JSX, type MouseEvent} from 'react';
import {useNavigate} from 'react-router';
import {
  Alert,
  Box,
  Card,
  CardContent,
  Checkbox,
  Chip,
  Divider,
  Grid,
  IconButton,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  Stack,
  Typography,
} from '@wso2/oxygen-ui';
import {Copy, EllipsisVertical, Eye, Palette, Search as SearchIcon, Trash2} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger';
import useMultiSelect from '@/features/design-system/hooks/useMultiSelect';
import useNotification from '@/features/design-system/hooks/useNotification';
import CardSkeleton from '@/features/design-system/components/CardSkeleton';
import EmptyState from '@/features/design-system/components/EmptyState';
import BulkActionBar from '@/features/design-system/components/BulkActionBar';
import ConfirmDialog from '@/features/design-system/components/ConfirmDialog';
import useGetThemes from '../api/useGetThemes';
import useDuplicateTheme from '../api/useDuplicateTheme';
import useBulkDeleteThemes from '../api/useBulkDeleteThemes';
import ThemePreviewPlaceholder from './ThemePreviewPlaceholder';
import ThemeDeleteDialog from './ThemeDeleteDialog';

/**
 * Props for the ThemesList component.
 */
interface ThemesListProps {
  /** Search query to filter themes */
  searchQuery?: string;
}

/**
 * Component to display themes list as cards with multi-select and search functionality.
 */
export default function ThemesList({searchQuery = ''}: ThemesListProps): JSX.Element {
  const {t} = useTranslation();
  const navigate = useNavigate();
  const logger = useLogger('ThemesList');
  const {showSuccess, showError} = useNotification();
  const {data, isLoading, error} = useGetThemes();
  const {mutate: duplicateTheme} = useDuplicateTheme();
  const {mutate: bulkDeleteThemes} = useBulkDeleteThemes();

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedThemeId, setSelectedThemeId] = useState<string | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [bulkDeleteDialogOpen, setBulkDeleteDialogOpen] = useState(false);

  // Multi-select state
  const {selectedIds, isSelected, toggleSelection, selectAll, clearSelection, selectedCount} = useMultiSelect();

  // Filter themes based on search query
  const filteredThemes = useMemo(() => {
    if (!data?.themes) return [];
    if (!searchQuery.trim()) return data.themes;

    const query = searchQuery.toLowerCase();
    return data.themes.filter((theme) => theme.displayName.toLowerCase().includes(query));
  }, [data?.themes, searchQuery]);

  const handleMenuOpen = (event: MouseEvent<HTMLElement>, themeId: string) => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
    setSelectedThemeId(themeId);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleViewClick = () => {
    if (selectedThemeId) {
      (async () => {
        await navigate(`/themes/builder/${selectedThemeId}`);
      })().catch((error: unknown) => {
        logger.error('Navigation to theme builder failed', {error});
      });
    }
    handleMenuClose();
  };

  const handleDuplicateClick = () => {
    if (selectedThemeId) {
      duplicateTheme(selectedThemeId, {
        onSuccess: () => {
          showSuccess(t('themes:notifications.duplicateSuccess', {defaultValue: 'Theme duplicated successfully'}));
          logger.info('Theme duplicated successfully', {themeId: selectedThemeId});
        },
        onError: (error) => {
          showError(t('themes:notifications.duplicateError', {defaultValue: 'Failed to duplicate theme'}));
          logger.error('Failed to duplicate theme', {error});
        },
      });
    }
    handleMenuClose();
  };

  const handleDeleteClick = () => {
    setDeleteDialogOpen(true);
    handleMenuClose();
  };

  const handleDeleteDialogClose = () => {
    setDeleteDialogOpen(false);
    setSelectedThemeId(null);
  };

  const handleBulkDelete = () => {
    setBulkDeleteDialogOpen(true);
  };

  const handleBulkDeleteConfirm = () => {
    bulkDeleteThemes(Array.from(selectedIds), {
      onSuccess: () => {
        showSuccess(
          t('themes:notifications.bulkDeleteSuccess', {
            count: selectedCount,
            defaultValue: `${selectedCount} themes deleted successfully`,
          }),
        );
        clearSelection();
        setBulkDeleteDialogOpen(false);
        logger.info('Bulk delete successful', {count: selectedCount});
      },
      onError: (error) => {
        showError(t('themes:notifications.deleteError', {defaultValue: 'Failed to delete themes'}));
        logger.error('Bulk delete failed', {error});
        setBulkDeleteDialogOpen(false);
      },
    });
  };

  const handleCardClick = (themeId: string) => {
    if (selectedCount > 0) {
      // If in selection mode, toggle selection
      toggleSelection(themeId);
    } else {
      // Otherwise navigate to builder
      (async () => {
        await navigate(`/themes/builder/${themeId}`);
      })().catch((error: unknown) => {
        logger.error('Navigation to theme builder failed', {error});
      });
    }
  };

  const handleCheckboxClick = (event: MouseEvent, themeId: string) => {
    event.stopPropagation();
    toggleSelection(themeId);
  };

  if (isLoading) {
    return <CardSkeleton count={8} height={280} />;
  }

  if (error) {
    return (
      <Alert severity="error">
        {t('themes:listing.error', {defaultValue: 'Failed to load themes. Please try again.'})}
      </Alert>
    );
  }

  // Empty state - no themes at all
  if (!data?.themes || data.themes.length === 0) {
    return (
      <EmptyState
        icon={<Palette size={48} />}
        title={t('themes:listing.empty.title', {defaultValue: 'No themes yet'})}
        description={t('themes:listing.empty.description', {
          defaultValue: 'Create your first theme to customize your authentication experience',
        })}
        action={{
          label: t('themes:listing.create', {defaultValue: 'Create Theme'}),
          icon: <Palette size={20} />,
          onClick: () => {
            (async () => {
              await navigate('/themes/builder');
            })().catch((error: unknown) => {
              logger.error('Navigation to theme builder failed', {error});
            });
          },
        }}
      />
    );
  }

  // Empty state - no search results
  if (filteredThemes.length === 0 && searchQuery) {
    return (
      <EmptyState
        icon={<SearchIcon size={48} />}
        title={t('common:search.noResults', {defaultValue: 'No results found'})}
        description={t('common:search.tryDifferent', {defaultValue: 'Try a different search term'})}
      />
    );
  }

  return (
    <>
      <Grid container spacing={3}>
        {filteredThemes.map((theme) => (
          <Grid item xs={12} sm={6} md={4} lg={3} key={theme.id}>
            <Card
              sx={{
                cursor: 'pointer',
                transition: 'all 0.2s ease-in-out',
                position: 'relative',
                border: isSelected(theme.id) ? 2 : 1,
                borderColor: isSelected(theme.id) ? 'primary.main' : 'divider',
                '&:hover': {
                  boxShadow: 8,
                  transform: 'translateY(-4px)',
                  '& .card-actions': {
                    opacity: 1,
                  },
                  '& .card-checkbox': {
                    opacity: 1,
                  },
                },
              }}
              onClick={() => handleCardClick(theme.id)}
            >
              {/* Selection checkbox */}
              <Checkbox
                checked={isSelected(theme.id)}
                onChange={(e) => handleCheckboxClick(e as unknown as MouseEvent, theme.id)}
                className="card-checkbox"
                sx={{
                  position: 'absolute',
                  top: 8,
                  left: 8,
                  zIndex: 1,
                  bgcolor: 'background.paper',
                  borderRadius: 1,
                  opacity: selectedCount > 0 || isSelected(theme.id) ? 1 : 0,
                  transition: 'opacity 0.2s',
                  '&:hover': {
                    bgcolor: 'background.paper',
                  },
                }}
                onClick={(e) => e.stopPropagation()}
              />

              <CardContent>
                {/* Preview */}
                <Box
                  mb={2}
                  sx={{
                    borderRadius: 2,
                    overflow: 'hidden',
                    border: '1px solid',
                    borderColor: 'divider',
                  }}
                >
                  <ThemePreviewPlaceholder theme={theme} colorScheme={theme?.theme?.defaultColorScheme || 'light'} />
                </Box>

                {/* Title and metadata */}
                <Stack spacing={1}>
                  <Typography variant="h6" noWrap>
                    {theme.displayName}
                  </Typography>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Chip
                      label={theme.theme?.defaultColorScheme || 'dark'}
                      size="small"
                      sx={{textTransform: 'capitalize'}}
                    />
                  </Stack>
                </Stack>

                {/* Actions menu */}
                <Box
                  className="card-actions"
                  sx={{
                    position: 'absolute',
                    top: 8,
                    right: 8,
                    opacity: 0,
                    transition: 'opacity 0.2s',
                  }}
                >
                  <IconButton
                    size="small"
                    onClick={(e) => handleMenuOpen(e, theme.id)}
                    sx={{
                      bgcolor: 'background.paper',
                      '&:hover': {
                        bgcolor: 'action.hover',
                      },
                    }}
                    aria-label={t('common:actions.moreOptions', {defaultValue: 'More options'})}
                  >
                    <EllipsisVertical size={18} />
                  </IconButton>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* Context menu */}
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
        anchorOrigin={{vertical: 'bottom', horizontal: 'right'}}
        transformOrigin={{vertical: 'top', horizontal: 'right'}}
      >
        <MenuItem onClick={handleViewClick}>
          <ListItemIcon>
            <Eye size={18} />
          </ListItemIcon>
          <ListItemText>{t('common:actions.view', {defaultValue: 'View'})}</ListItemText>
        </MenuItem>
        <MenuItem onClick={handleDuplicateClick}>
          <ListItemIcon>
            <Copy size={18} />
          </ListItemIcon>
          <ListItemText>{t('common:actions.duplicate', {defaultValue: 'Duplicate'})}</ListItemText>
        </MenuItem>
        <Divider />
        <MenuItem onClick={handleDeleteClick}>
          <ListItemIcon>
            <Trash2 size={18} />
          </ListItemIcon>
          <ListItemText>{t('common:actions.delete', {defaultValue: 'Delete'})}</ListItemText>
        </MenuItem>
      </Menu>

      {/* Bulk action bar */}
      {selectedCount > 0 && (
        <BulkActionBar
          selectedCount={selectedCount}
          onClearSelection={clearSelection}
          actions={[
            {
              label: t('common:actions.delete', {defaultValue: 'Delete'}),
              icon: <Trash2 size={20} />,
              onClick: handleBulkDelete,
              color: 'error',
            },
          ]}
        />
      )}

      {/* Single delete dialog */}
      {selectedThemeId && (
        <ThemeDeleteDialog open={deleteDialogOpen} themeId={selectedThemeId} onClose={handleDeleteDialogClose} />
      )}

      {/* Bulk delete dialog */}
      <ConfirmDialog
        open={bulkDeleteDialogOpen}
        onClose={() => setBulkDeleteDialogOpen(false)}
        onConfirm={handleBulkDeleteConfirm}
        title={t('themes:dialog.bulkDelete.title', {defaultValue: 'Delete Themes'})}
        message={t('themes:dialog.bulkDelete.message', {
          count: selectedCount,
          defaultValue: `Are you sure you want to delete ${selectedCount} themes? This action cannot be undone.`,
        })}
        confirmLabel={t('common:actions.delete', {defaultValue: 'Delete'})}
        severity="error"
      />
    </>
  );
}
