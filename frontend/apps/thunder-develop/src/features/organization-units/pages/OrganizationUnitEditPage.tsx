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

import {useState, useCallback, useMemo, type JSX} from 'react';
import {useNavigate, useParams, useLocation} from 'react-router';
import {
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
} from '@wso2/oxygen-ui';
import {ArrowLeft, Edit, Building} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger/react';
import useGetOrganizationUnit from '../api/useGetOrganizationUnit';
import useUpdateOrganizationUnit from '../api/useUpdateOrganizationUnit';
import type {OrganizationUnit, OUNavigationState} from '../types/organization-units';
import OrganizationUnitDeleteDialog from '../components/OrganizationUnitDeleteDialog';
import EditGeneralSettings from '../components/edit-organization-unit/general-settings/EditGeneralSettings';
import EditChildOUs from '../components/edit-organization-unit/child-ous/EditChildOUs';
import EditUsers from '../components/edit-organization-unit/users/EditUsers';
import EditGroups from '../components/edit-organization-unit/groups/EditGroups';
import EditAdvancedSettings from '../components/edit-organization-unit/advanced-settings/EditAdvancedSettings';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel({children = null, value, index, ...other}: TabPanelProps): JSX.Element {
  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`ou-tabpanel-${index}`}
      aria-labelledby={`ou-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{py: 3}}>{children}</Box>}
    </div>
  );
}

export default function OrganizationUnitEditPage(): JSX.Element {
  const {id} = useParams<{id: string}>();
  const navigate = useNavigate();
  const location = useLocation();
  const {t} = useTranslation();
  const logger = useLogger('OrganizationUnitEditPage');

  // Check if we came from another OU (via parent or child OU link)
  const navigationState = location.state as OUNavigationState | null;
  const fromOU = navigationState?.fromOU;

  const {data: organizationUnit, isLoading, error: fetchError, refetch} = useGetOrganizationUnit(id);
  const updateOrganizationUnit = useUpdateOrganizationUnit();

  const [activeTab, setActiveTab] = useState(0);
  const [editedOU, setEditedOU] = useState<Partial<OrganizationUnit>>({});
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);
  const [isEditingName, setIsEditingName] = useState(false);
  const [isEditingDescription, setIsEditingDescription] = useState(false);
  const [tempName, setTempName] = useState('');
  const [tempDescription, setTempDescription] = useState('');

  const handleBack = async (): Promise<void> => {
    if (fromOU) {
      await navigate(`/organization-units/${fromOU.id}`);
    } else {
      await navigate('/organization-units');
    }
  };

  const backButtonText = fromOU
    ? t('organizationUnits:view.backToOU', {name: fromOU.name})
    : t('organizationUnits:view.back');

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number): void => {
    setActiveTab(newValue);
  };

  const handleFieldChange = useCallback((field: keyof OrganizationUnit, value: unknown): void => {
    setEditedOU((prev) => ({...prev, [field]: value}));
  }, []);

  const handleSave = useCallback(async (): Promise<void> => {
    if (!organizationUnit || !id) return;

    const updatedData = {
      handle: editedOU.handle ?? organizationUnit.handle,
      name: editedOU.name ?? organizationUnit.name,
      description: editedOU.description ?? organizationUnit.description,
    };

    try {
      await updateOrganizationUnit.mutateAsync({
        id,
        data: updatedData,
      });
      setEditedOU({});
      await refetch();
    } catch {
      logger.error('Failed to update organization unit');
    }
  }, [organizationUnit, id, editedOU, updateOrganizationUnit, refetch, logger]);

  const hasChanges = useMemo(() => Object.keys(editedOU).length > 0, [editedOU]);

  const handleDeleteSuccess = (): void => {
    (async (): Promise<void> => {
      await navigate('/organization-units');
    })().catch((_error: unknown) => {
      logger.error('Failed to navigate after deleting organization unit', {error: _error});
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
          {fetchError.message ?? t('organizationUnits:view.error')}
        </Alert>
        <Button
          onClick={() => {
            handleBack().catch(() => {});
          }}
          startIcon={<ArrowLeft size={16} />}
        >
          {t('organizationUnits:view.back')}
        </Button>
      </Box>
    );
  }

  if (!organizationUnit) {
    return (
      <Box sx={{maxWidth: 1200, mx: 'auto', px: 2, pt: 6}}>
        <Alert severity="warning" sx={{mb: 2}}>
          {t('organizationUnits:view.notFound')}
        </Alert>
        <Button
          onClick={() => {
            handleBack().catch(() => {});
          }}
          startIcon={<ArrowLeft size={16} />}
        >
          {t('organizationUnits:view.back')}
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
            handleBack().catch(() => {});
          }}
          variant="text"
          startIcon={<ArrowLeft size={16} />}
        >
          {backButtonText}
        </Button>
      </Stack>

      {/* Organization Unit Header */}
      <Box sx={{p: 3, mb: 3}}>
        <Stack direction="row" spacing={3} alignItems="center">
          <Box
            sx={{
              width: 80,
              height: 80,
              borderRadius: 2,
              bgcolor: 'action.hover',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Building size={32} />
          </Box>
          <Box flex={1}>
            <Stack direction="row" alignItems="center" spacing={1} mb={1}>
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
                      setIsEditingName(false);
                    }
                  }}
                  size="small"
                />
              ) : (
                <>
                  <Typography variant="h3">{editedOU.name ?? organizationUnit.name}</Typography>
                  <IconButton
                    size="small"
                    onClick={() => {
                      setTempName(editedOU.name ?? organizationUnit.name);
                      setIsEditingName(true);
                    }}
                    sx={{
                      opacity: 0.6,
                      '&:hover': {opacity: 1},
                    }}
                  >
                    <Edit size={16} />
                  </IconButton>
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
                    if (trimmedDescription !== (organizationUnit.description ?? '')) {
                      handleFieldChange('description', trimmedDescription || null);
                    }
                    setIsEditingDescription(false);
                  }}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && e.ctrlKey) {
                      const trimmedDescription = tempDescription.trim();
                      if (trimmedDescription !== (organizationUnit.description ?? '')) {
                        handleFieldChange('description', trimmedDescription || null);
                      }
                      setIsEditingDescription(false);
                    } else if (e.key === 'Escape') {
                      setIsEditingDescription(false);
                    }
                  }}
                  size="small"
                  placeholder={t('organizationUnits:view.description.placeholder')}
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
                    {editedOU.description ?? organizationUnit.description ?? t('organizationUnits:view.description.empty')}
                  </Typography>
                  <IconButton
                    size="small"
                    onClick={() => {
                      setTempDescription(editedOU.description ?? organizationUnit.description ?? '');
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
                </>
              )}
            </Stack>
          </Box>
        </Stack>
      </Box>

      {/* Tabs */}
      <Tabs value={activeTab} onChange={handleTabChange} aria-label="organization unit settings tabs">
        <Tab
          label={t('organizationUnits:view.tabs.general')}
          id="ou-tab-0"
          aria-controls="ou-tabpanel-0"
          sx={{textTransform: 'none'}}
        />
        <Tab
          label={t('organizationUnits:view.tabs.childOUs')}
          id="ou-tab-1"
          aria-controls="ou-tabpanel-1"
          sx={{textTransform: 'none'}}
        />
        <Tab
          label={t('organizationUnits:view.tabs.users')}
          id="ou-tab-2"
          aria-controls="ou-tabpanel-2"
          sx={{textTransform: 'none'}}
        />
        <Tab
          label={t('organizationUnits:view.tabs.groups')}
          id="ou-tab-3"
          aria-controls="ou-tabpanel-3"
          sx={{textTransform: 'none'}}
        />
        <Tab
          label={t('organizationUnits:view.tabs.advanced')}
          id="ou-tab-4"
          aria-controls="ou-tabpanel-4"
          sx={{textTransform: 'none'}}
        />
      </Tabs>

      {/* Tab Panels */}
      <>
        {/* General Settings Tab */}
        <TabPanel value={activeTab} index={0}>
          <EditGeneralSettings organizationUnit={organizationUnit} />
        </TabPanel>

        {/* Child OUs Tab */}
        <TabPanel value={activeTab} index={1}>
          <EditChildOUs organizationUnitId={id!} organizationUnitName={organizationUnit.name} />
        </TabPanel>

        {/* Users Tab */}
        <TabPanel value={activeTab} index={2}>
          <EditUsers organizationUnitId={id!} />
        </TabPanel>

        {/* Groups Tab */}
        <TabPanel value={activeTab} index={3}>
          <EditGroups organizationUnitId={id!} />
        </TabPanel>

        {/* Advanced Settings Tab */}
        <TabPanel value={activeTab} index={4}>
          <EditAdvancedSettings onDeleteClick={() => setDeleteDialogOpen(true)} />
        </TabPanel>
      </>

      {/* Delete Dialog */}
      <OrganizationUnitDeleteDialog
        open={deleteDialogOpen}
        organizationUnitId={id ?? null}
        onClose={() => setDeleteDialogOpen(false)}
        onSuccess={handleDeleteSuccess}
      />

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
              {t('organizationUnits:view.unsavedChanges')}
            </Typography>
            <Button variant="outlined" color="error" onClick={() => setEditedOU({})}>
              {t('organizationUnits:view.reset')}
            </Button>
            <Button
              variant="contained"
              onClick={() => {
                handleSave().catch(() => {});
              }}
              disabled={updateOrganizationUnit.isPending}
            >
              {updateOrganizationUnit.isPending ? t('organizationUnits:view.saving') : t('organizationUnits:view.save')}
            </Button>
          </Stack>
        </Paper>
      )}
    </Box>
  );
}
