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
import {Copy, EllipsisVertical, Eye, LayoutGrid, Search as SearchIcon, Trash2} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger';
import useMultiSelect from '@/features/design-system/hooks/useMultiSelect';
import useNotification from '@/features/design-system/hooks/useNotification';
import CardSkeleton from '@/features/design-system/components/CardSkeleton';
import EmptyState from '@/features/design-system/components/EmptyState';
import BulkActionBar from '@/features/design-system/components/BulkActionBar';
import ConfirmDialog from '@/features/design-system/components/ConfirmDialog';
import useGetLayouts from '../api/useGetLayouts';
import useDuplicateLayout from '../api/useDuplicateLayout';
import useBulkDeleteLayouts from '../api/useBulkDeleteLayouts';
import LayoutPreviewPlaceholder from './LayoutPreviewPlaceholder';
import LayoutDeleteDialog from './LayoutDeleteDialog';

/**
 * Props for the LayoutsList component.
 */
interface LayoutsListProps {
  /** Search query to filter layouts */
  searchQuery?: string;
}

/**
 * Component to display layouts list as cards with multi-select and search functionality.
 */
export default function LayoutsList({searchQuery = ''}: LayoutsListProps): JSX.Element {
  const {t} = useTranslation();
  const navigate = useNavigate();
  const logger = useLogger('LayoutsList');
  const {showSuccess, showError} = useNotification();
  const {data, isLoading, error} = useGetLayouts();
  const {mutate: duplicateLayout} = useDuplicateLayout();
  const {mutate: bulkDeleteLayouts} = useBulkDeleteLayouts();

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedLayoutId, setSelectedLayoutId] = useState<string | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [bulkDeleteDialogOpen, setBulkDeleteDialogOpen] = useState(false);

  // Multi-select state
  const {selectedIds, isSelected, toggleSelection, selectAll, clearSelection, selectedCount} = useMultiSelect();

  // Filter layouts based on search query
  const filteredLayouts = useMemo(() => {
    if (!data?.layouts) return [];
    if (!searchQuery.trim()) return data.layouts;

    const query = searchQuery.toLowerCase();
    return data.layouts.filter((layout) => layout.displayName.toLowerCase().includes(query));
  }, [data?.layouts, searchQuery]);

  const handleMenuOpen = (event: MouseEvent<HTMLElement>, layoutId: string) => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
    setSelectedLayoutId(layoutId);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleViewClick = () => {
    if (selectedLayoutId) {
      (async () => {
        await navigate(`/layouts/builder/${selectedLayoutId}`);
      })().catch((error: unknown) => {
        logger.error('Navigation to layout builder failed', {error});
      });
    }
    handleMenuClose();
  };

  const handleDuplicateClick = () => {
    if (selectedLayoutId) {
      duplicateLayout(selectedLayoutId, {
        onSuccess: () => {
          showSuccess(t('layouts:notifications.duplicateSuccess', {defaultValue: 'Layout duplicated successfully'}));
          logger.info('Layout duplicated successfully', {layoutId: selectedLayoutId});
        },
        onError: (error) => {
          showError(t('layouts:notifications.duplicateError', {defaultValue: 'Failed to duplicate layout'}));
          logger.error('Failed to duplicate layout', {error});
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
    setSelectedLayoutId(null);
  };

  const handleBulkDelete = () => {
    setBulkDeleteDialogOpen(true);
  };

  const handleBulkDeleteConfirm = () => {
    bulkDeleteLayouts(Array.from(selectedIds), {
      onSuccess: () => {
        showSuccess(
          t('layouts:notifications.bulkDeleteSuccess', {
            count: selectedCount,
            defaultValue: `${selectedCount} layouts deleted successfully`,
          })
        );
        clearSelection();
        setBulkDeleteDialogOpen(false);
        logger.info('Bulk delete successful', {count: selectedCount});
      },
      onError: (error) => {
        showError(t('layouts:notifications.deleteError', {defaultValue: 'Failed to delete layouts'}));
        logger.error('Bulk delete failed', {error});
        setBulkDeleteDialogOpen(false);
      },
    });
  };

  const handleCardClick = (layoutId: string) => {
    if (selectedCount > 0) {
      // If in selection mode, toggle selection
      toggleSelection(layoutId);
    } else {
      // Otherwise navigate to builder
      (async () => {
        await navigate(`/layouts/builder/${layoutId}`);
      })().catch((error: unknown) => {
        logger.error('Navigation to layout builder failed', {error});
      });
    }
  };

  const handleCheckboxClick = (event: MouseEvent, layoutId: string) => {
    event.stopPropagation();
    toggleSelection(layoutId);
  };

  if (isLoading) {
    return <CardSkeleton count={8} height={280} />;
  }

  if (error) {
    return (
      <Alert severity="error">
        {t('layouts:listing.error', {defaultValue: 'Failed to load layouts. Please try again.'})}
      </Alert>
    );
  }

  // Empty state - no layouts at all
  if (!data?.layouts || data.layouts.length === 0) {
    return (
      <EmptyState
        icon={<LayoutGrid size={48} />}
        title={t('layouts:listing.empty.title', {defaultValue: 'No layouts yet'})}
        description={t('layouts:listing.empty.description', {
          defaultValue: 'Create your first layout to customize your authentication screens',
        })}
        action={{
          label: t('layouts:listing.create', {defaultValue: 'Create Layout'}),
          icon: <LayoutGrid size={20} />,
          onClick: () => {
            (async () => {
              await navigate('/layouts/builder');
            })().catch((error: unknown) => {
              logger.error('Navigation to layout builder failed', {error});
            });
          },
        }}
      />
    );
  }

  // Empty state - no search results
  if (filteredLayouts.length === 0 && searchQuery) {
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
        {filteredLayouts.map((layout) => (
          <Grid item xs={12} sm={6} md={4} lg={3} key={layout.id}>
            <Card
              sx={{
                cursor: 'pointer',
                transition: 'all 0.2s ease-in-out',
                position: 'relative',
                border: isSelected(layout.id) ? 2 : 1,
                borderColor: isSelected(layout.id) ? 'primary.main' : 'divider',
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
              onClick={() => handleCardClick(layout.id)}
            >
              {/* Selection checkbox */}
              <Checkbox
                checked={isSelected(layout.id)}
                onChange={(e) => handleCheckboxClick(e as unknown as MouseEvent, layout.id)}
                className="card-checkbox"
                sx={{
                  position: 'absolute',
                  top: 8,
                  left: 8,
                  zIndex: 1,
                  bgcolor: 'background.paper',
                  borderRadius: 1,
                  opacity: selectedCount > 0 || isSelected(layout.id) ? 1 : 0,
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
                  <LayoutPreviewPlaceholder layout={layout} screenType="signin" />
                </Box>

                {/* Title and metadata */}
                <Stack spacing={1}>
                  <Typography variant="h6" noWrap>
                    {layout.displayName}
                  </Typography>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Chip
                      label={layout.layout?.screens?.signin?.template || 'centered'}
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
                    onClick={(e) => handleMenuOpen(e, layout.id)}
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
      {selectedLayoutId && (
        <LayoutDeleteDialog open={deleteDialogOpen} layoutId={selectedLayoutId} onClose={handleDeleteDialogClose} />
      )}

      {/* Bulk delete dialog */}
      <ConfirmDialog
        open={bulkDeleteDialogOpen}
        onClose={() => setBulkDeleteDialogOpen(false)}
        onConfirm={handleBulkDeleteConfirm}
        title={t('layouts:dialog.bulkDelete.title', {defaultValue: 'Delete Layouts'})}
        message={t('layouts:dialog.bulkDelete.message', {
          count: selectedCount,
          defaultValue: `Are you sure you want to delete ${selectedCount} layouts? This action cannot be undone.`,
        })}
        confirmLabel={t('common:actions.delete', {defaultValue: 'Delete'})}
        severity="error"
      />
    </>
  );
}
