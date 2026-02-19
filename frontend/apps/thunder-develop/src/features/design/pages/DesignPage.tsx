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

import {Box, Grid, Skeleton, Stack, Typography} from '@wso2/oxygen-ui';
import {LayoutTemplate, Palette} from '@wso2/oxygen-ui-icons-react';
import {useState, type JSX} from 'react';
import {useNavigate} from 'react-router';
import {useGetThemes, useGetLayouts} from '@thunder/shared-design';
import {useTranslation} from 'react-i18next';
import SectionHeader from '../components/common/SectionHeader';
import ItemCard from '../components/common/ItemCard';
import AddCard from '../components/common/AddCard';
import ThemeThumbnail from '../components/themes/ThemeThumbnail';
import LayoutThumbnail from '../components/layouts/LayoutThumbnail';
import DesignUIConstants from '../constants/design-ui-constants';

export default function DesignPage(): JSX.Element {
  const {t} = useTranslation();
  const navigate = useNavigate();
  const {data: themesData, isLoading: themesLoading} = useGetThemes();
  const {data: layoutsData, isLoading: layoutsLoading} = useGetLayouts();

  const [showAllThemes, setShowAllThemes] = useState(false);
  const [showAllLayouts, setShowAllLayouts] = useState(false);

  const allThemes = themesData?.themes ?? [];
  const allLayouts = layoutsData?.layouts ?? [];

  const visibleThemes = showAllThemes ? allThemes : allThemes.slice(0, DesignUIConstants.INITIAL_LIMIT);
  const visibleLayouts = showAllLayouts ? allLayouts : allLayouts.slice(0, DesignUIConstants.INITIAL_LIMIT);

  const skeletonCount = 4;

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={4} flexWrap="wrap" gap={2}>
        <Box>
          <Typography variant="h1" gutterBottom>
            {t('applications:listing.title')}
          </Typography>
          <Typography variant="subtitle1" color="text.secondary">
            {t('applications:listing.subtitle')}
          </Typography>
        </Box>
      </Stack>

      <Box>
        {/* ── Themes section ─────────────────────────────────────────────── */}
        <SectionHeader
          title="Themes"
          count={allThemes.length}
          icon={<Palette size={18} />}
          hasMore={!showAllThemes && allThemes.length > DesignUIConstants.INITIAL_LIMIT}
          onShowMore={() => setShowAllThemes(true)}
        />

        <Grid container spacing={2} sx={{mb: 5}}>
          {themesLoading
            ? Array.from({length: skeletonCount}).map((_, i) => (
                <Grid key={i} size={{xs: 6, sm: 4, md: 3, lg: 2}}>
                  <Skeleton variant="rounded" sx={{aspectRatio: '4/3', height: 'auto', borderRadius: 2}} />
                </Grid>
              ))
            : visibleThemes.map((theme) => (
                <Grid key={theme.id} size={{xs: 6, sm: 4, md: 3, lg: 2}}>
                  <ItemCard
                    thumbnail={<ThemeThumbnail theme={theme} />}
                    name={theme.displayName}
                    onClick={() => navigate(`/design/themes/${theme.id}`)}
                  />
                </Grid>
              ))}
          {!themesLoading && (
            <Grid size={{xs: 6, sm: 4, md: 3, lg: 2}}>
              <AddCard
                label="New theme"
                onClick={() => {
                  /* TODO: create theme */
                }}
              />
            </Grid>
          )}
        </Grid>

        {/* ── Layouts section ────────────────────────────────────────────── */}
        <SectionHeader
          title="Layouts"
          count={allLayouts.length}
          icon={<LayoutTemplate size={18} />}
          hasMore={!showAllLayouts && allLayouts.length > DesignUIConstants.INITIAL_LIMIT}
          onShowMore={() => setShowAllLayouts(true)}
        />

        <Grid container spacing={2}>
          {layoutsLoading
            ? Array.from({length: skeletonCount}).map((_, i) => (
                <Grid key={i} size={{xs: 6, sm: 4, md: 3, lg: 2}}>
                  <Skeleton variant="rounded" sx={{aspectRatio: '4/3', height: 'auto', borderRadius: 2}} />
                </Grid>
              ))
            : visibleLayouts.map((layout) => (
                <Grid key={layout.id} size={{xs: 6, sm: 4, md: 3, lg: 2}}>
                  <ItemCard
                    thumbnail={<LayoutThumbnail layout={layout} />}
                    name={layout.displayName}
                    onClick={() => navigate(`/design/layouts/${layout.id}`)}
                  />
                </Grid>
              ))}
          {!layoutsLoading && (
            <Grid size={{xs: 6, sm: 4, md: 3, lg: 2}}>
              <AddCard
                label="New layout"
                onClick={() => {
                  /* TODO: create layout */
                }}
              />
            </Grid>
          )}
        </Grid>

        {/* Empty states */}
        {!themesLoading && allThemes.length === 0 && (
          <Box sx={{mb: 5, py: 6, textAlign: 'center', color: 'text.secondary'}}>
            <Palette size={32} style={{opacity: 0.3, marginBottom: 8}} />
            <Typography variant="body2">No themes yet</Typography>
          </Box>
        )}
        {!layoutsLoading && allLayouts.length === 0 && (
          <Box sx={{py: 6, textAlign: 'center', color: 'text.secondary'}}>
            <LayoutTemplate size={32} style={{opacity: 0.3, marginBottom: 8}} />
            <Typography variant="body2">No layouts yet</Typography>
          </Box>
        )}
      </Box>
    </Box>
  );
}
