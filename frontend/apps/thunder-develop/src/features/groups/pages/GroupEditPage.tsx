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

import {useState, useCallback, useMemo, useRef, useEffect} from 'react';
import type {ReactNode, SyntheticEvent, JSX} from 'react';
import {useNavigate, useParams} from 'react-router';
import {
  Avatar,
  Box,
  Stack,
  Typography,
  Button,
  TextField,
  Paper,
  Alert,
  IconButton,
  CircularProgress,
  Tabs,
  Tab,
  Snackbar,
  Tooltip,
} from '@wso2/oxygen-ui';
import {ArrowLeft, Copy, Check, Edit, Users} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger/react';
import useGetGroup from '../api/useGetGroup';
import useUpdateGroup from '../api/useUpdateGroup';
import type {Group} from '../models/group';
import GroupDeleteDialog from '../components/GroupDeleteDialog';
import EditGeneralSettings from '../components/edit-group/general-settings/EditGeneralSettings';
import EditMembersSettings from '../components/edit-group/members-settings/EditMembersSettings';

interface TabPanelProps {
  children?: ReactNode;
  index: number;
  value: number;
}

function TabPanel({children = null, value, index, ...other}: TabPanelProps): JSX.Element {
  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`group-tabpanel-${index}`}
      aria-labelledby={`group-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{py: 3}}>{children}</Box>}
    </div>
  );
}

export default function GroupEditPage(): JSX.Element {
  const {groupId} = useParams<{groupId: string}>();
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('GroupEditPage');

  const {data: group, isLoading, error: fetchError, refetch} = useGetGroup(groupId ?? '');
  const updateGroup = useUpdateGroup();

  const [activeTab, setActiveTab] = useState(0);
  const [editedGroup, setEditedGroup] = useState<Partial<Group>>({});
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);
  const [snackbar, setSnackbar] = useState<{open: boolean; message: string}>({open: false, message: ''});
  const [isEditingName, setIsEditingName] = useState(false);
  const [isEditingDescription, setIsEditingDescription] = useState(false);
  const [tempName, setTempName] = useState('');
  const [tempDescription, setTempDescription] = useState('');
  const [copiedField, setCopiedField] = useState<string | null>(null);
  const copyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(
    () => () => {
      if (copyTimeoutRef.current) {
        clearTimeout(copyTimeoutRef.current);
      }
    },
    [],
  );

  const handleCopyToClipboard = useCallback(async (text: string, fieldName: string): Promise<void> => {
    await navigator.clipboard.writeText(text);
    setCopiedField(fieldName);
    if (copyTimeoutRef.current) {
      clearTimeout(copyTimeoutRef.current);
    }
    copyTimeoutRef.current = setTimeout(() => {
      setCopiedField(null);
    }, 2000);
  }, []);

  const listUrl = '/groups';

  const handleBack = async (): Promise<void> => {
    await navigate(listUrl);
  };

  const handleTabChange = (_event: SyntheticEvent, newValue: number): void => {
    setActiveTab(newValue);
  };

  const handleFieldChange = useCallback((field: keyof Group, value: unknown): void => {
    setEditedGroup((prev) => ({...prev, [field]: value}));
  }, []);

  const handleSave = useCallback(async (): Promise<void> => {
    if (!group || !groupId) return;

    const updatedData = {
      name: editedGroup.name ?? group.name,
      description: 'description' in editedGroup ? editedGroup.description : group.description,
      organizationUnitId: group.organizationUnitId,
    };

    try {
      await updateGroup.mutateAsync({
        groupId,
        data: updatedData,
      });
      setEditedGroup({});
      await refetch();
    } catch (err: unknown) {
      logger.error('Failed to update group', {error: err});
      setSnackbar({
        open: true,
        message: err instanceof Error ? err.message : t('groups:edit.page.error'),
      });
    }
  }, [group, groupId, editedGroup, updateGroup, refetch, logger, t]);

  const hasChanges = useMemo(() => Object.keys(editedGroup).length > 0, [editedGroup]);

  // Resolve the effective description accounting for user edits (including clearing).
  // 'description' in editedGroup means the user has touched the field; otherwise fall back to server value.
  const effectiveDescription =
    'description' in editedGroup ? (editedGroup.description ?? '') : (group?.description ?? '');

  const handleDeleteSuccess = (): void => {
    (async (): Promise<void> => {
      await navigate(listUrl);
    })().catch((_error: unknown) => {
      logger.error('Failed to navigate after deleting group', {error: _error});
    });
  };

  if (isLoading) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px'}}>
        <CircularProgress />
      </Box>
    );
  }

  if (fetchError) {
    return (
      <Box sx={{maxWidth: 1200, mx: 'auto', px: 2, pt: 6}}>
        <Alert severity="error" sx={{mb: 2}}>
          {fetchError.message ?? t('groups:edit.page.error')}
        </Alert>
        <Button
          onClick={() => {
            handleBack().catch((error: unknown) => {
              logger.error('Failed to navigate back', {error});
            });
          }}
          startIcon={<ArrowLeft size={16} />}
        >
          {t('groups:edit.page.back')}
        </Button>
      </Box>
    );
  }

  if (!group) {
    return (
      <Box sx={{maxWidth: 1200, mx: 'auto', px: 2, pt: 6}}>
        <Alert severity="warning" sx={{mb: 2}}>
          {t('groups:edit.page.notFound')}
        </Alert>
        <Button
          onClick={() => {
            handleBack().catch((error: unknown) => {
              logger.error('Failed to navigate back', {error});
            });
          }}
          startIcon={<ArrowLeft size={16} />}
        >
          {t('groups:edit.page.back')}
        </Button>
      </Box>
    );
  }

  return (
    <Box>
      {/* Header */}
      <Stack direction="row" alignItems="center" justifyContent="space-between" mb={3}>
        <Button
          onClick={() => {
            handleBack().catch((error: unknown) => {
              logger.error('Failed to navigate back', {error});
            });
          }}
          variant="text"
          startIcon={<ArrowLeft size={16} />}
        >
          {t('groups:edit.page.back')}
        </Button>
      </Stack>

      {/* Group Header */}
      <Box sx={{p: 3, mb: 3}}>
        <Stack direction="row" spacing={3} alignItems="flex-start">
          <Avatar
            sx={{
              width: 80,
              height: 80,
            }}
          >
            <Users size={32} />
          </Avatar>
          <Box flex={1}>
            <Stack direction="row" alignItems="center" spacing={1} mb={0.5}>
              {isEditingName ? (
                <TextField
                  autoFocus
                  value={tempName}
                  onChange={(e) => setTempName(e.target.value)}
                  onBlur={() => {
                    if (tempName.trim()) {
                      handleFieldChange('name', tempName.trim());
                    }
                    setIsEditingName(false);
                  }}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      if (tempName.trim()) {
                        handleFieldChange('name', tempName.trim());
                      }
                      setIsEditingName(false);
                    } else if (e.key === 'Escape') {
                      setTempName(editedGroup.name ?? group.name);
                      setIsEditingName(false);
                    }
                  }}
                  size="small"
                />
              ) : (
                <>
                  <Typography variant="h3">{editedGroup.name ?? group.name}</Typography>
                  <Tooltip title={t('groups:edit.page.header.editName')} placement="right">
                    <IconButton
                      size="small"
                      onClick={() => {
                        setTempName(editedGroup.name ?? group.name);
                        setIsEditingName(true);
                      }}
                      sx={{
                        opacity: 0.6,
                        '&:hover': {opacity: 1},
                      }}
                    >
                      <Edit size={16} />
                    </IconButton>
                  </Tooltip>
                </>
              )}
            </Stack>
            <Stack direction="row" alignItems="flex-start" spacing={1}>
              {isEditingDescription ? (
                <TextField
                  autoFocus
                  fullWidth
                  multiline
                  rows={2}
                  value={tempDescription}
                  onChange={(e) => setTempDescription(e.target.value)}
                  onBlur={() => {
                    const trimmedDescription = tempDescription.trim();
                    if (trimmedDescription !== effectiveDescription) {
                      handleFieldChange('description', trimmedDescription || undefined);
                    }
                    setIsEditingDescription(false);
                  }}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && e.ctrlKey) {
                      const trimmedDescription = tempDescription.trim();
                      if (trimmedDescription !== effectiveDescription) {
                        handleFieldChange('description', trimmedDescription || undefined);
                      }
                      setIsEditingDescription(false);
                    } else if (e.key === 'Escape') {
                      setTempDescription(effectiveDescription);
                      setIsEditingDescription(false);
                    }
                  }}
                  size="small"
                  placeholder={t('groups:edit.page.description.placeholder')}
                  sx={{
                    maxWidth: '600px',
                    '& .MuiInputBase-root': {
                      fontSize: '0.875rem',
                    },
                  }}
                />
              ) : (
                <>
                  <Typography variant="body2" color="text.secondary">
                    {effectiveDescription || t('groups:edit.page.description.empty')}
                  </Typography>
                  <Tooltip title={t('groups:edit.page.header.editDescription')} placement="right">
                    <IconButton
                      size="small"
                      onClick={() => {
                        setTempDescription(effectiveDescription);
                        setIsEditingDescription(true);
                      }}
                      sx={{
                        opacity: 0.6,
                        '&:hover': {opacity: 1},
                        mt: -0.5,
                      }}
                    >
                      <Edit size={14} />
                    </IconButton>
                  </Tooltip>
                </>
              )}
            </Stack>

            {/* Group ID */}
            <Tooltip
              title={
                copiedField === 'groupId'
                  ? t('common:actions.copied')
                  : t('groups:edit.general.sections.quickCopy.copyGroupId')
              }
              placement="right"
            >
              <Stack
                direction="row"
                alignItems="center"
                spacing={0.5}
                role="button"
                tabIndex={0}
                onClick={() => {
                  handleCopyToClipboard(group.id, 'groupId').catch(() => {});
                }}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    handleCopyToClipboard(group.id, 'groupId').catch(() => {});
                  }
                }}
                sx={{
                  cursor: 'pointer',
                  width: 'fit-content',
                  mt: 0.5,
                  '&:hover .copy-icon': {opacity: 1},
                  '&:focus-visible .copy-icon': {opacity: 1},
                }}
              >
                <Typography
                  variant="caption"
                  sx={{fontFamily: 'monospace', color: 'text.disabled', fontSize: '0.75rem'}}
                >
                  {group.id}
                </Typography>
                {copiedField === 'groupId' ? (
                  <Check size={12} color="var(--mui-palette-success-main)" />
                ) : (
                  <Copy size={12} className="copy-icon" style={{opacity: 0.4}} />
                )}
              </Stack>
            </Tooltip>
          </Box>
        </Stack>
      </Box>

      {/* Tabs */}
      <Tabs value={activeTab} onChange={handleTabChange} aria-label="group settings tabs">
        <Tab
          label={t('groups:edit.page.tabs.general')}
          id="group-tab-0"
          aria-controls="group-tabpanel-0"
          sx={{textTransform: 'none'}}
        />
        <Tab
          label={t('groups:edit.page.tabs.members')}
          id="group-tab-1"
          aria-controls="group-tabpanel-1"
          sx={{textTransform: 'none'}}
        />
      </Tabs>

      {/* Tab Panels */}
      <>
        <TabPanel value={activeTab} index={0}>
          <EditGeneralSettings group={group} onDeleteClick={() => setDeleteDialogOpen(true)} />
        </TabPanel>

        <TabPanel value={activeTab} index={1}>
          <EditMembersSettings group={group} />
        </TabPanel>
      </>

      {/* Delete Dialog */}
      <GroupDeleteDialog
        open={deleteDialogOpen}
        groupId={groupId ?? null}
        onClose={() => setDeleteDialogOpen(false)}
        onSuccess={handleDeleteSuccess}
      />

      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={(_event, reason) => {
          if (reason === 'clickaway') return;
          setSnackbar((prev) => ({...prev, open: false}));
        }}
        anchorOrigin={{vertical: 'bottom', horizontal: 'right'}}
      >
        <Alert onClose={() => setSnackbar((prev) => ({...prev, open: false}))} severity="error" sx={{width: '100%'}}>
          {snackbar.message}
        </Alert>
      </Snackbar>

      {/* Floating Action Bar */}
      {hasChanges && (
        <Paper
          sx={{
            position: 'fixed',
            bottom: 0,
            left: 0,
            right: 0,
            p: 2,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 2,
            borderRadius: '12px 12px 0 0',
            boxShadow: '0 -4px 20px rgba(0, 0, 0, 0.1)',
            zIndex: 1000,
            bgcolor: 'background.paper',
          }}
        >
          <Stack direction="row" spacing={2} alignItems="center">
            <Typography variant="body2" sx={{display: 'flex', alignItems: 'center', gap: 1}}>
              <Box
                component="span"
                sx={{
                  width: 20,
                  height: 20,
                  borderRadius: '50%',
                  border: '2px solid',
                  borderColor: 'warning.main',
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: '12px',
                  fontWeight: 'bold',
                }}
              >
                !
              </Box>
              {t('groups:edit.page.unsavedChanges')}
            </Typography>
            <Button variant="outlined" color="error" onClick={() => setEditedGroup({})}>
              {t('groups:edit.page.reset')}
            </Button>
            <Button
              variant="contained"
              onClick={() => {
                handleSave().catch(() => {});
              }}
              disabled={updateGroup.isPending}
            >
              {updateGroup.isPending ? t('groups:edit.page.saving') : t('groups:edit.page.save')}
            </Button>
          </Stack>
        </Paper>
      )}
    </Box>
  );
}
