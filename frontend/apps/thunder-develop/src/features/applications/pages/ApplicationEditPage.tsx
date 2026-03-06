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

import {useState, useCallback, useMemo, type SyntheticEvent} from 'react';
import {Link, useNavigate, useParams} from 'react-router';
import {
  Box,
  Stack,
  Typography,
  Button,
  Paper,
  CircularProgress,
  Alert,
  Avatar,
  IconButton,
  TextField,
  Chip,
  Tabs,
  Tab,
  PageContent,
  PageTitle,
} from '@wso2/oxygen-ui';
import {ArrowLeft, AppWindow, Edit} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger';
import useGetApplication from '../api/useGetApplication';
import useUpdateApplication from '../api/useUpdateApplication';
import type {Application} from '../models/application';
import type {OAuth2Config} from '../models/oauth';
import LogoUpdateModal from '../../../components/LogoUpdateModal';
import IntegrationGuides from '../components/edit-application/integration-guides/IntegrationGuides';
import EditGeneralSettings from '../components/edit-application/general-settings/EditGeneralSettings';
import EditFlowsSettings from '../components/edit-application/flows-settings/EditFlowsSettings';
import EditCustomizationSettings from '../components/edit-application/customization-settings/EditCustomizationSettings';
import EditTokenSettings from '../components/edit-application/token-settings/EditTokenSettings';
import EditAdvancedSettings from '../components/edit-application/advanced-settings/EditAdvancedSettings';
import getTemplateMetadata from '../utils/getTemplateMetadata';
import getIntegrationGuidesForTemplate from '../utils/getIntegrationGuidesForTemplate';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel({children = null, value, index, ...other}: TabPanelProps) {
  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`edit-tabpanel-${index}`}
      aria-labelledby={`edit-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{py: 3}}>{children}</Box>}
    </div>
  );
}

export default function ApplicationEditPage() {
  const logger = useLogger('ApplicationEditPage');
  const {t} = useTranslation();
  const navigate = useNavigate();
  const {applicationId} = useParams<{applicationId: string}>();

  const {data: application, isLoading, error, isError, refetch} = useGetApplication(applicationId ?? '');
  const updateApplication = useUpdateApplication();

  const [activeTab, setActiveTab] = useState(0);
  const [isLogoModalOpen, setIsLogoModalOpen] = useState(false);
  const [editedApp, setEditedApp] = useState<Partial<Application>>({});
  const [copiedField, setCopiedField] = useState<string | null>(null);
  const [isEditingName, setIsEditingName] = useState(false);
  const [isEditingDescription, setIsEditingDescription] = useState(false);
  const [tempName, setTempName] = useState('');
  const [tempDescription, setTempDescription] = useState('');

  const handleBack = async () => {
    await navigate('/applications');
  };

  const handleTabChange = (_event: SyntheticEvent, newValue: number) => {
    setActiveTab(newValue);
  };

  const handleCopyToClipboard = useCallback(
    async (text: string, fieldName: string) => {
      try {
        await navigator.clipboard.writeText(text);
        setCopiedField(fieldName);
        setTimeout(() => setCopiedField(null), 2000);
      } catch {
        logger.error('Failed to copy to clipboard');
      }
    },
    [logger],
  );

  const hasIntegrationGuides = useMemo(
    () => application && getIntegrationGuidesForTemplate(application.template) !== null,
    [application],
  );

  const handleFieldChange = useCallback((field: keyof Application, value: unknown) => {
    setEditedApp((prev) => ({...prev, [field]: value}));
  }, []);

  const handleSave = useCallback(async () => {
    if (!application || !applicationId) return;

    const updatedData = {
      ...application,
      ...editedApp,
    };

    try {
      await updateApplication.mutateAsync({
        applicationId,
        data: updatedData,
      });
      setEditedApp({});
      await refetch();
    } catch {
      logger.error('Failed to update application');
    }
  }, [application, applicationId, editedApp, updateApplication, refetch, logger]);

  const hasChanges = useMemo(() => Object.keys(editedApp).length > 0, [editedApp]);

  if (isLoading) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px'}}>
        <CircularProgress />
      </Box>
    );
  }

  if (isError || error) {
    return (
      <PageContent>
        <Alert severity="error" sx={{mb: 2}}>
          {error?.message ?? t('applications:edit.page.error')}
        </Alert>
        <Button
          onClick={() => {
            handleBack().catch(() => {});
          }}
          startIcon={<ArrowLeft size={16} />}
        >
          {t('applications:edit.page.back')}
        </Button>
      </PageContent>
    );
  }

  if (!application) {
    return (
      <PageContent>
        <Alert severity="warning" sx={{mb: 2}}>
          {t('applications:edit.page.notFound')}
        </Alert>
        <Button
          onClick={() => {
            handleBack().catch(() => {});
          }}
          startIcon={<ArrowLeft size={16} />}
        >
          {t('applications:edit.page.back')}
        </Button>
      </PageContent>
    );
  }

  const oauth2Config: OAuth2Config | undefined = application.inbound_auth_config?.find(
    (config) => config.type === 'oauth2',
  )?.config;

  return (
    <PageContent>
      {/* Header */}
      <PageTitle>
        <PageTitle.BackButton component={<Link to="/applications" />}>
          {t('applications:edit.page.back')}
        </PageTitle.BackButton>
        <PageTitle.Avatar sx={{position: 'relative', overflow: 'visible'}}>
          <Avatar
            src={editedApp.logo_url ?? application.logo_url}
            slotProps={{
              img: {
                onError: (e: SyntheticEvent<HTMLImageElement>) => {
                  e.currentTarget.style.display = 'none';
                },
              },
            }}
            sx={{
              cursor: 'pointer',
              '&:hover': {
                opacity: 0.8,
              },
            }}
            onClick={() => setIsLogoModalOpen(true)}
          >
            <AppWindow />
          </Avatar>
          <IconButton
            size="small"
            aria-label={t('applications:edit.page.logoUpdate.label')}
            sx={{
              position: 'absolute',
              bottom: -4,
              right: -4,
              bgcolor: 'background.paper',
              boxShadow: 1,
              '&:hover': {bgcolor: 'action.hover'},
            }}
            onClick={() => setIsLogoModalOpen(true)}
          >
            <Edit size={14} />
          </IconButton>
        </PageTitle.Avatar>
        <PageTitle.Header>
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
                <Typography variant="h3">{editedApp.name ?? application.name}</Typography>
                <IconButton
                  size="small"
                  onClick={() => {
                    setTempName(editedApp.name ?? application.name);
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
        </PageTitle.Header>
        <PageTitle.SubHeader>
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
                  if (trimmedDescription || trimmedDescription !== (application.description ?? '')) {
                    handleFieldChange('description', trimmedDescription);
                  }
                  setIsEditingDescription(false);
                }}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && e.ctrlKey) {
                    const trimmedDescription = tempDescription.trim();
                    if (trimmedDescription || trimmedDescription !== (application.description ?? '')) {
                      handleFieldChange('description', trimmedDescription);
                    }
                    setIsEditingDescription(false);
                  } else if (e.key === 'Escape') {
                    setIsEditingDescription(false);
                  }
                }}
                size="small"
                placeholder={t('applications:edit.page.description.placeholder')}
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
                  {editedApp.description ?? application.description ?? t('applications:edit.page.description.empty')}
                </Typography>
                <IconButton
                  size="small"
                  onClick={() => {
                    setTempDescription(editedApp.description ?? application.description ?? '');
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
          {(editedApp.template ?? application.template) &&
            (() => {
              const templateMetadata = getTemplateMetadata(editedApp.template ?? application.template);
              return templateMetadata ? (
                <Box sx={{mt: 1}}>
                  <Chip
                    icon={
                      <Box sx={{display: 'flex', alignItems: 'center', '& > *': {width: 16, height: 16}}}>
                        {templateMetadata.icon}
                      </Box>
                    }
                    label={templateMetadata.displayName}
                    size="small"
                    sx={{px: 0.5}}
                  />
                </Box>
              ) : null;
            })()}
        </PageTitle.SubHeader>
      </PageTitle>

      {/* Tabs */}
      <Tabs value={activeTab} onChange={handleTabChange} aria-label="application settings tabs">
        {hasIntegrationGuides && (
          <Tab
            label={t('applications:edit.page.tabs.overview')}
            id="edit-tab-0"
            aria-controls="edit-tabpanel-0"
            sx={{textTransform: 'none'}}
          />
        )}
        <Tab
          label={t('applications:edit.page.tabs.general')}
          id={`edit-tab-${hasIntegrationGuides ? 1 : 0}`}
          aria-controls={`edit-tabpanel-${hasIntegrationGuides ? 1 : 0}`}
          sx={{textTransform: 'none'}}
        />
        <Tab
          label={t('applications:edit.page.tabs.flows')}
          id={`edit-tab-${hasIntegrationGuides ? 2 : 1}`}
          aria-controls={`edit-tabpanel-${hasIntegrationGuides ? 2 : 1}`}
          sx={{textTransform: 'none'}}
        />
        <Tab
          label={t('applications:edit.page.tabs.customization')}
          id={`edit-tab-${hasIntegrationGuides ? 3 : 2}`}
          aria-controls={`edit-tabpanel-${hasIntegrationGuides ? 3 : 2}`}
          sx={{textTransform: 'none'}}
        />
        <Tab
          label={t('applications:edit.page.tabs.token')}
          id={`edit-tab-${hasIntegrationGuides ? 4 : 3}`}
          aria-controls={`edit-tabpanel-${hasIntegrationGuides ? 4 : 3}`}
          sx={{textTransform: 'none'}}
        />
        <Tab
          label={t('applications:edit.page.tabs.advanced')}
          id={`edit-tab-${hasIntegrationGuides ? 5 : 4}`}
          aria-controls={`edit-tabpanel-${hasIntegrationGuides ? 5 : 4}`}
          sx={{textTransform: 'none'}}
        />
      </Tabs>

      {/* Tab Panels */}
      <>
        {/* Overview Tab */}
        {hasIntegrationGuides && (
          <TabPanel value={activeTab} index={0}>
            <IntegrationGuides application={application} oauth2Config={oauth2Config} />
          </TabPanel>
        )}

        {/* General Tab */}
        <TabPanel value={activeTab} index={hasIntegrationGuides ? 1 : 0}>
          <EditGeneralSettings
            application={application}
            editedApp={editedApp}
            onFieldChange={handleFieldChange}
            oauth2Config={oauth2Config}
            copiedField={copiedField}
            onCopyToClipboard={handleCopyToClipboard}
          />
        </TabPanel>

        {/* Flows Tab */}
        <TabPanel value={activeTab} index={hasIntegrationGuides ? 2 : 1}>
          <EditFlowsSettings application={application} editedApp={editedApp} onFieldChange={handleFieldChange} />
        </TabPanel>

        {/* Customization Tab */}
        <TabPanel value={activeTab} index={hasIntegrationGuides ? 3 : 2}>
          <EditCustomizationSettings
            application={application}
            editedApp={editedApp}
            onFieldChange={handleFieldChange}
          />
        </TabPanel>

        {/* Token Tab */}
        <TabPanel value={activeTab} index={hasIntegrationGuides ? 4 : 3}>
          <EditTokenSettings application={application} oauth2Config={oauth2Config} onFieldChange={handleFieldChange} />
        </TabPanel>

        {/* Advanced Settings Tab */}
        <TabPanel value={activeTab} index={hasIntegrationGuides ? 5 : 4}>
          <EditAdvancedSettings
            application={application}
            editedApp={editedApp}
            oauth2Config={oauth2Config}
            onFieldChange={handleFieldChange}
          />
        </TabPanel>
      </>

      {/* Logo Update Modal */}
      <LogoUpdateModal
        open={isLogoModalOpen}
        onClose={() => setIsLogoModalOpen(false)}
        currentLogoUrl={editedApp.logo_url ?? application.logo_url}
        onLogoUpdate={(newLogoUrl: string) => {
          setEditedApp({...editedApp, logo_url: newLogoUrl});
          setIsLogoModalOpen(false);
        }}
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
              {t('applications:edit.page.unsavedChanges')}
            </Typography>
            <Button variant="outlined" color="error" onClick={() => setEditedApp({})}>
              {t('applications:edit.page.reset')}
            </Button>
            <Button
              variant="contained"
              onClick={() => {
                handleSave().catch(() => {});
              }}
              disabled={updateApplication.isPending}
            >
              {updateApplication.isPending ? t('applications:edit.page.saving') : t('applications:edit.page.save')}
            </Button>
          </Stack>
        </Paper>
      )}
    </PageContent>
  );
}
