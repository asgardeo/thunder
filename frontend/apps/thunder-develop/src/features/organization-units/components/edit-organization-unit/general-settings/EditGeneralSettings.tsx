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

import type {JSX} from 'react';
import {useState, useCallback, useRef, useEffect} from 'react';
import {Link} from 'react-router';
import {
  Stack,
  Typography,
  Paper,
  CircularProgress,
  TextField,
  InputAdornment,
  Tooltip,
  IconButton,
  FormControl,
  FormLabel,
} from '@wso2/oxygen-ui';
import {Copy, Check} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import type {OrganizationUnit, OUNavigationState} from '../../../types/organization-units';
import useGetOrganizationUnit from '../../../api/useGetOrganizationUnit';

/**
 * Props for the {@link EditGeneralSettings} component.
 */
interface EditGeneralSettingsProps {
  /**
   * The organization unit being displayed
   */
  organizationUnit: OrganizationUnit;
}

/**
 * General settings tab content for the Organization Unit edit page.
 *
 * Displays basic information about the organization unit including:
 * - Handle (unique identifier)
 * - Organization Unit ID
 * - Parent Organization Unit (if any)
 *
 * @param props - Component props
 * @returns General settings tab content
 */
export default function EditGeneralSettings({organizationUnit}: EditGeneralSettingsProps): JSX.Element {
  const {t} = useTranslation();
  const [copiedField, setCopiedField] = useState<string | null>(null);
  const copyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const {data: parentOU, isLoading: isLoadingParent} = useGetOrganizationUnit(
    organizationUnit.parent ?? undefined,
    Boolean(organizationUnit.parent),
  );

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

  const renderParentInfo = (): JSX.Element => {
    if (!organizationUnit.parent) {
      return (
        <TextField
          fullWidth
          id="parent-ou-input"
          value={t('organizationUnits:view.general.noParent')}
          InputProps={{
            readOnly: true,
          }}
        />
      );
    }

    if (isLoadingParent) {
      return <CircularProgress size={16} />;
    }

    if (parentOU) {
      const navigationState: OUNavigationState = {
        fromOU: {
          id: organizationUnit.id,
          name: organizationUnit.name,
        },
      };

      return (
        <Stack direction="row" spacing={1} alignItems="center">
          <Typography
            component={Link}
            to={`/organization-units/${parentOU.id}`}
            state={navigationState}
            variant="body2"
            sx={{
              color: 'primary.main',
              textDecoration: 'none',
              '&:hover': {
                textDecoration: 'underline',
              },
            }}
          >
            {parentOU.name}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            ({parentOU.id})
          </Typography>
        </Stack>
      );
    }

    return (
      <TextField
        fullWidth
        id="parent-ou-input"
        value={organizationUnit.parent}
        InputProps={{
          readOnly: true,
        }}
        sx={{
          '& input': {
            fontFamily: 'monospace',
            fontSize: '0.875rem',
          },
        }}
      />
    );
  };

  return (
    <Paper sx={{p: 3, mb: 3}}>
      <Typography variant="h6" gutterBottom>
        {t('organizationUnits:view.general.title')}
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{mb: 3}}>
        {t('organizationUnits:view.general.subtitle')}
      </Typography>

      <Stack spacing={3}>
        <FormControl fullWidth>
          <FormLabel htmlFor="handle-input">{t('organizationUnits:form.handle')}</FormLabel>
          <TextField
            fullWidth
            id="handle-input"
            value={organizationUnit.handle}
            InputProps={{
              readOnly: true,
              endAdornment: (
                <InputAdornment position="end">
                  <Tooltip title={copiedField === 'handle' ? t('common:actions.copied') : t('common:actions.copy')}>
                    <IconButton
                      aria-label={
                        copiedField === 'handle' ? t('common:actions.copied') : t('common:actions.copy')
                      }
                      onClick={() => {
                        handleCopyToClipboard(organizationUnit.handle, 'handle').catch(() => {});
                      }}
                      edge="end"
                    >
                      {copiedField === 'handle' ? <Check size={16} /> : <Copy size={16} />}
                    </IconButton>
                  </Tooltip>
                </InputAdornment>
              ),
            }}
            sx={{
              '& input': {
                fontFamily: 'monospace',
                fontSize: '0.875rem',
              },
            }}
          />
        </FormControl>

        <FormControl fullWidth>
          <FormLabel htmlFor="ou-id-input">{t('organizationUnits:view.general.id')}</FormLabel>
          <TextField
            fullWidth
            id="ou-id-input"
            value={organizationUnit.id}
            InputProps={{
              readOnly: true,
              endAdornment: (
                <InputAdornment position="end">
                  <Tooltip title={copiedField === 'ou_id' ? t('common:actions.copied') : t('common:actions.copy')}>
                    <IconButton
                      aria-label={
                        copiedField === 'ou_id' ? t('common:actions.copied') : t('common:actions.copy')
                      }
                      onClick={() => {
                        handleCopyToClipboard(organizationUnit.id, 'ou_id').catch(() => {});
                      }}
                      edge="end"
                    >
                      {copiedField === 'ou_id' ? <Check size={16} /> : <Copy size={16} />}
                    </IconButton>
                  </Tooltip>
                </InputAdornment>
              ),
            }}
            sx={{
              '& input': {
                fontFamily: 'monospace',
                fontSize: '0.875rem',
              },
            }}
          />
        </FormControl>

        <FormControl fullWidth>
          <FormLabel htmlFor="parent-ou-input">{t('organizationUnits:view.general.parent')}</FormLabel>
          {renderParentInfo()}
        </FormControl>
      </Stack>
    </Paper>
  );
}
