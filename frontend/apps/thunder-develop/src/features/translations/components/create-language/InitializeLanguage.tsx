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

import {Box, Card, CardActionArea, CircularProgress, LinearProgress, Stack, Typography} from '@wso2/oxygen-ui';
import {type JSX} from 'react';
import {useTranslation} from 'react-i18next';

export interface InitializeLanguageProps {
  populateFromEnglish: boolean;
  onPopulateChange: (value: boolean) => void;
  isCreating: boolean;
  progress: number;
}

export default function InitializeLanguage({
  populateFromEnglish,
  onPopulateChange,
  isCreating,
  progress,
}: InitializeLanguageProps): JSX.Element {
  const {t} = useTranslation('translations');

  return (
    <Stack spacing={4}>
      <Box>
        <Typography variant="h1" gutterBottom>
          {t('language.create.initialize.title')}
        </Typography>
        <Typography variant="body1" color="text.secondary">
          {t('language.create.initialize.subtitle')}
        </Typography>
      </Box>

      <Stack spacing={2}>
        <Card
          variant="outlined"
          onClick={() => !isCreating && onPopulateChange(true)}
          sx={{
            borderColor: populateFromEnglish ? 'primary.main' : 'divider',
            borderWidth: populateFromEnglish ? 2 : 1,
            cursor: isCreating ? 'default' : 'pointer',
            opacity: isCreating ? 0.6 : 1,
          }}
        >
          <CardActionArea disabled={isCreating} sx={{p: 2.5}}>
            <Stack spacing={0.5}>
              <Typography variant="body1" fontWeight={600}>
                {t('language.create.initialize.copyFromEnglish.label')}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {t('language.create.initialize.copyFromEnglish.description')}
              </Typography>
            </Stack>
          </CardActionArea>
        </Card>

        <Card
          variant="outlined"
          onClick={() => !isCreating && onPopulateChange(false)}
          sx={{
            borderColor: !populateFromEnglish ? 'primary.main' : 'divider',
            borderWidth: !populateFromEnglish ? 2 : 1,
            cursor: isCreating ? 'default' : 'pointer',
            opacity: isCreating ? 0.6 : 1,
          }}
        >
          <CardActionArea disabled={isCreating} sx={{p: 2.5}}>
            <Stack spacing={0.5}>
              <Typography variant="body1" fontWeight={600}>
                {t('language.create.initialize.startEmpty.label')}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {t('language.create.initialize.startEmpty.description')}
              </Typography>
            </Stack>
          </CardActionArea>
        </Card>
      </Stack>

      {isCreating && (
        <Box sx={{display: 'flex', flexDirection: 'column', gap: 1}}>
          <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
            <CircularProgress size={16} />
            <Typography variant="body2" color="text.secondary">
              {t('language.add.adding')} ({progress}%)
            </Typography>
          </Box>
          <LinearProgress variant="determinate" value={progress} />
        </Box>
      )}
    </Stack>
  );
}
