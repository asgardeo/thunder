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

import {Box, Button, Card, Grid, PageContent, PageTitle, Skeleton, Typography} from '@wso2/oxygen-ui';
import {ArrowUpRight, LayoutTemplate, Palette, Plus} from '@wso2/oxygen-ui-icons-react';
import {useState, type JSX} from 'react';
import {useNavigate} from 'react-router';
import {useGetThemes} from '@thunder/shared-design';
import SectionHeader from '../components/common/SectionHeader';
import ItemCard from '../components/common/ItemCard';
import ThemeThumbnail from '../components/themes/ThemeThumbnail';
import LayoutPresetThumbnail, {type LayoutPresetVariant} from '../components/layouts/LayoutPresetThumbnail';
import DesignUIConstants from '../constants/design-ui-constants';

const LAYOUT_PRESETS: {id: LayoutPresetVariant; name: string}[] = [
  {id: 'centered', name: 'Centered'},
  {id: 'split', name: 'Split Screen'},
  {id: 'fullscreen', name: 'Full Screen'},
  {id: 'popup', name: 'Popup'},
];

export default function DesignPage(): JSX.Element {
  const navigate = useNavigate();
  const {data: themesData, isLoading: themesLoading} = useGetThemes();

  const [showAllThemes, setShowAllThemes] = useState(false);

  const allThemes = themesData?.themes ?? [];
  const visibleThemes = showAllThemes ? allThemes : allThemes.slice(0, DesignUIConstants.INITIAL_LIMIT);

  const skeletonCount = 4;

  return (
    <PageContent>
      <PageTitle>
        <PageTitle.Header>Design</PageTitle.Header>
        <PageTitle.SubHeader>
          Create, customize, and manage visual themes & layouts for your applications.
        </PageTitle.SubHeader>
      </PageTitle>

      <Box>
        {/* ── Themes section ─────────────────────────────────────────────── */}
        <SectionHeader
          title="Themes"
          count={allThemes.length}
          icon={<Palette size={18} />}
          action={
            <Button
              variant="contained"
              size="small"
              startIcon={<Plus size={16} />}
              onClick={() => void navigate('/design/themes/create')}
            >
              Add Theme
            </Button>
          }
        />

        <Grid container spacing={2} sx={{mb: 5}}>
          {themesLoading
            ? Array.from({length: skeletonCount}).map((_, i) => (
                <Grid key={i} size={{xs: 6, sm: 4, md: 3, lg: 2}}>
                  <Skeleton variant="rounded" sx={{aspectRatio: '4/3', height: 'auto', borderRadius: 2}} />
                </Grid>
              ))
            : [
                ...visibleThemes.map((theme) => (
                  <Grid key={theme.id} size={{xs: 6, sm: 4, md: 3, lg: 2}}>
                    <ItemCard
                      thumbnail={<ThemeThumbnail theme={theme} />}
                      name={theme.displayName}
                      onClick={() => navigate(`/design/themes/${theme.id}`)}
                    />
                  </Grid>
                )),
                ...(!showAllThemes && allThemes.length > DesignUIConstants.INITIAL_LIMIT
                  ? [
                      <Grid key="show-more" size={{xs: 6, sm: 4, md: 3, lg: 2}}>
                        <Box
                          onClick={() => setShowAllThemes(true)}
                          sx={{
                            cursor: 'pointer',
                            borderRadius: 1,
                            border: '1.5px dashed',
                            borderColor: 'divider',
                            aspectRatio: '4/3',
                            display: 'flex',
                            flexDirection: 'column',
                            alignItems: 'center',
                            justifyContent: 'center',
                            gap: 0.75,
                            color: 'text.secondary',
                            transition: 'all 0.18s ease',
                            width: '100%',
                            height: '100%',
                            '&:hover': {borderColor: 'primary.main', color: 'primary.main', bgcolor: 'primary.50'},
                          }}
                        >
                          <ArrowUpRight size={20} />
                          <Typography variant="caption" sx={{fontSize: '0.75rem', fontWeight: 500}}>
                            Show {allThemes.length - DesignUIConstants.INITIAL_LIMIT} more
                          </Typography>
                        </Box>
                      </Grid>,
                    ]
                  : []),
              ]}
        </Grid>

        {!themesLoading && allThemes.length === 0 && (
          <Box sx={{mb: 5, py: 6, textAlign: 'center', color: 'text.secondary'}}>
            <Palette size={32} style={{opacity: 0.3, marginBottom: 8}} />
            <Typography variant="body2">No themes yet</Typography>
          </Box>
        )}
      </Box>

      <Box>
        {/* ── Layouts section ────────────────────────────────────────────── */}
        <SectionHeader
          title="Layouts"
          count={LAYOUT_PRESETS.length}
          icon={<LayoutTemplate size={18} />}
          comingSoon
        />

        <Grid container spacing={2}>
          {LAYOUT_PRESETS.map((preset) => (
            <Grid key={preset.id} size={{xs: 6, sm: 4, md: 3, lg: 2}}>
              <Card sx={{cursor: 'default', opacity: 0.72, pointerEvents: 'none'}}>
                <Box sx={{aspectRatio: '4/3', overflow: 'hidden', position: 'relative'}}>
                  <Box sx={{width: '100%', height: '100%', filter: 'grayscale(1)'}}>
                    <LayoutPresetThumbnail variant={preset.id} />
                  </Box>
                  {/* Coming Soon badge */}
                  <Box
                    sx={{
                      position: 'absolute',
                      top: 8,
                      right: 8,
                      bgcolor: 'warning.main',
                      color: 'warning.contrastText',
                      px: 1,
                      py: 0.4,
                      borderRadius: 1,
                      fontSize: '0.68rem',
                      fontWeight: 700,
                      letterSpacing: '0.03em',
                    }}
                  >
                    Coming Soon
                  </Box>
                </Box>
                <Box sx={{px: 1.5, py: 1, borderTop: '1px solid', borderColor: 'divider'}}>
                  <Typography
                    variant="body2"
                    sx={{fontWeight: 500, fontSize: '0.8125rem', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap'}}
                  >
                    {preset.name}
                  </Typography>
                </Box>
              </Card>
            </Grid>
          ))}
        </Grid>
      </Box>
    </PageContent>
  );
}
