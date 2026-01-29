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

import {styled, Typography, Breadcrumbs, breadcrumbsClasses, Box} from '@wso2/oxygen-ui';
import {ChevronRightIcon} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import useNavigation from '@/layouts/contexts/useNavigation';

const StyledBreadcrumbs = styled(Breadcrumbs)(({theme}) => ({
  margin: theme.spacing(1, 0),
  [`& .${breadcrumbsClasses.separator}`]: {
    color: (theme.vars ?? theme).palette.action.disabled,
    margin: 1,
  },
  [`& .${breadcrumbsClasses.ol}`]: {
    alignItems: 'center',
  },
}));

// Map page IDs to translation keys
const PAGE_TRANSLATION_KEY_MAP: Record<string, string> = {
  users: 'navigation:pages.users',
  'user-types': 'navigation:pages.userTypes',
  'organization-units': 'navigation:pages.organizationUnits',
  integrations: 'navigation:pages.integrations',
  applications: 'navigation:pages.applications',
  dashboard: 'navigation:pages.dashboard',
  flows: 'navigation:pages.flows',
};

export default function NavbarBreadcrumbs(): JSX.Element {
  const {t} = useTranslation();
  const {currentPage} = useNavigation();

  // Get the translation key for the current page, fallback to the page ID if not found
  const currentPageTranslationKey = PAGE_TRANSLATION_KEY_MAP[currentPage] ?? currentPage;

  return (
    <StyledBreadcrumbs
      aria-label="breadcrumb"
      separator={
        <Box paddingX={1}>
          <ChevronRightIcon size={11} />
        </Box>
      }
    >
      <Typography variant="body1">{t('navigation:breadcrumb.develop')}</Typography>
      <Typography variant="body1" sx={{color: 'text.primary', fontWeight: 600}}>
        {t(currentPageTranslationKey)}
      </Typography>
    </StyledBreadcrumbs>
  );
}
