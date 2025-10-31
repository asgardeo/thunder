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

import {styled} from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import Breadcrumbs, {breadcrumbsClasses} from '@mui/material/Breadcrumbs';
import {ChevronRightIcon} from 'lucide-react';
import Box from '@mui/material/Box';
import type {JSX} from 'react';
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

export default function NavbarBreadcrumbs(): JSX.Element {
  const {currentPage} = useNavigation();

  return (
    <StyledBreadcrumbs
      aria-label="breadcrumb"
      separator={
        <Box paddingX={1}>
          <ChevronRightIcon size={11} />
        </Box>
      }
    >
      <Typography variant="body1">Develop</Typography>
      <Typography variant="body1" sx={{color: 'text.primary', fontWeight: 600}}>
        {currentPage}
      </Typography>
    </StyledBreadcrumbs>
  );
}
