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

import {type JSX} from 'react';
import {
  Box,
  Typography,
  Button,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  CircularProgress,
  Stack,
  Divider,
  Chip,
  alpha,
} from '@wso2/oxygen-ui';
import {Plus, Check, LayoutGrid} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import type {Layout} from '@/features/layouts/models/layout';

/**
 * Props for LayoutControls component
 */
interface LayoutControlsProps {
  /** Available layouts */
  layouts: Layout[];
  /** Whether layouts are loading */
  isLoading: boolean;
  /** Currently selected layout ID */
  selectedLayoutId: string | null;
  /** Callback when layout is selected */
  onLayoutSelect: (layoutId: string) => void;
  /** Callback when create new layout is clicked */
  onCreateLayout: () => void;
}

/**
 * Layout controls component for selecting and managing layouts
 */
export default function LayoutControls({
  layouts,
  isLoading,
  selectedLayoutId,
  onLayoutSelect,
  onCreateLayout,
}: LayoutControlsProps): JSX.Element {
  const {t} = useTranslation();

  return (
    <Box sx={{height: '100%', display: 'flex', flexDirection: 'column'}}>
      {/* Header */}
      <Box sx={{p: 2, borderBottom: '1px solid', borderColor: 'divider'}}>
        <Stack direction="row" spacing={1} alignItems="center" mb={0.5}>
          <LayoutGrid size={20} />
          <Typography variant="h6">{t('design:layout.layouts', {defaultValue: 'Layouts'})}</Typography>
        </Stack>
        <Typography variant="caption" color="text.secondary">
          {t('design:layout.layoutsDescription', {
            defaultValue: 'Select a layout to preview your theme',
          })}
        </Typography>
      </Box>

      {/* Create New Layout Button */}
      <Box sx={{p: 2, borderBottom: '1px solid', borderColor: 'divider'}}>
        <Button
          variant="contained"
          fullWidth
          onClick={onCreateLayout}
          startIcon={<Plus size={18} />}
          sx={{
            background: (theme) =>
              `linear-gradient(135deg, ${theme.palette.primary.main} 0%, ${theme.palette.primary.dark} 100%)`,
            '&:hover': {
              background: (theme) =>
                `linear-gradient(135deg, ${theme.palette.primary.dark} 0%, ${theme.palette.primary.main} 100%)`,
            },
          }}
        >
          {t('design:layout.createNew', {defaultValue: 'Create New Layout'})}
        </Button>
      </Box>

      {/* Layouts List */}
      <Box sx={{flex: 1, overflow: 'auto'}}>
        {isLoading ? (
          <Box display="flex" justifyContent="center" p={4}>
            <CircularProgress />
          </Box>
        ) : layouts.length === 0 ? (
          <Box p={3} textAlign="center">
            <Typography variant="body2" color="text.secondary">
              {t('design:layout.noLayouts', {defaultValue: 'No layouts available'})}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {t('design:layout.createFirstLayout', {defaultValue: 'Create your first layout to get started'})}
            </Typography>
          </Box>
        ) : (
          <List dense sx={{p: 1}}>
            {layouts.map((layout, index) => (
              <ListItem key={layout.id} disablePadding sx={{mb: 0.5}}>
                <ListItemButton
                  selected={selectedLayoutId === layout.id}
                  onClick={() => onLayoutSelect(layout.id)}
                  sx={{
                    borderRadius: 1,
                    transition: 'all 0.2s',
                    '&.Mui-selected': {
                      bgcolor: (theme) => alpha(theme.palette.primary.main, 0.12),
                      borderLeft: '3px solid',
                      borderLeftColor: 'primary.main',
                      '&:hover': {
                        bgcolor: (theme) => alpha(theme.palette.primary.main, 0.18),
                      },
                    },
                    '&:hover': {
                      bgcolor: 'action.hover',
                    },
                  }}
                >
                  {selectedLayoutId === layout.id && (
                    <ListItemIcon sx={{minWidth: 32}}>
                      <Check size={16} color="primary" />
                    </ListItemIcon>
                  )}
                  <ListItemText
                    primary={
                      <Stack direction="row" spacing={1} alignItems="center">
                        <Typography variant="body2" fontWeight={selectedLayoutId === layout.id ? 600 : 400}>
                          {layout.displayName}
                        </Typography>
                        {selectedLayoutId === layout.id && (
                          <Chip label="Active" size="small" color="primary" sx={{height: 18, fontSize: '0.625rem'}} />
                        )}
                      </Stack>
                    }
                    secondary={
                      <Typography variant="caption" color="text.secondary" sx={{textTransform: 'capitalize'}}>
                        {layout.layout?.screens?.signin?.template || 'custom'}
                      </Typography>
                    }
                    sx={{pl: selectedLayoutId === layout.id ? 0 : 4}}
                  />
                </ListItemButton>
              </ListItem>
            ))}
          </List>
        )}
      </Box>
    </Box>
  );
}
