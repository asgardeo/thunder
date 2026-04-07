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

import {
  Box,
  Card,
  CardActionArea,
  CardContent,
  Chip,
  InputAdornment,
  Stack,
  TextField,
  Typography,
} from '@wso2/oxygen-ui';
import {Globe, KeyRound, LayoutTemplate, Lock, Search, Smartphone} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useEffect, useMemo, useState} from 'react';
import {useTranslation} from 'react-i18next';
import useGetFlowsMeta from '../../api/useGetFlowsMeta';
import type {FlowType} from '../../models/flows';
import type {FlowTemplate} from '../../models/templates';

interface SelectFlowTemplateProps {
  flowType: FlowType;
  selectedTemplate: FlowTemplate | null;
  onTemplateChange: (template: FlowTemplate) => void;
}

const CATEGORY_ORDER = ['STARTER', 'PASSWORD', 'SOCIAL_LOGIN', 'MFA', 'PASSWORDLESS'];

const CATEGORY_LABELS: Record<string, string> = {
  STARTER: 'Starter',
  PASSWORD: 'Password',
  SOCIAL_LOGIN: 'Social Login',
  MFA: 'MFA',
  PASSWORDLESS: 'Passwordless',
};

const CATEGORY_ICONS: Record<string, JSX.Element> = {
  STARTER: <LayoutTemplate size={20} />,
  PASSWORD: <Lock size={20} />,
  SOCIAL_LOGIN: <Globe size={20} />,
  MFA: <Smartphone size={20} />,
  PASSWORDLESS: <KeyRound size={20} />,
};

export default function SelectFlowTemplate({
  flowType,
  selectedTemplate,
  onTemplateChange,
}: SelectFlowTemplateProps): JSX.Element {
  const {t} = useTranslation();
  const {data} = useGetFlowsMeta({flowType});
  const templates = data.templates;

  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    if (!selectedTemplate && templates.length > 0) {
      onTemplateChange(templates[0]);
    }
  }, [templates, selectedTemplate, onTemplateChange]);

  const categories = useMemo(() => {
    const present = new Set(templates.map((template) => template.category));
    return CATEGORY_ORDER.filter((cat) => present.has(cat));
  }, [templates]);

  const filteredTemplates = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    return templates.filter((template) => {
      if (selectedCategory && template.category !== selectedCategory) return false;
      if (query) {
        const inLabel = template.display.label.toLowerCase().includes(query);
        const inDescription = template.display.description?.toLowerCase().includes(query) ?? false;
        return inLabel || inDescription;
      }
      return true;
    });
  }, [templates, selectedCategory, searchQuery]);

  return (
    <Stack direction="column" spacing={4} data-testid="select-flow-template">
      <Typography variant="h1" gutterBottom>
        {t('flows:create.template.title', 'Choose a starting template')}
      </Typography>

      {/* Search + category filters */}
      <Stack direction="row" spacing={2} alignItems="center" sx={{mt: 2, mb: 3, flexWrap: 'wrap', gap: 1}}>
        <TextField
          size="small"
          placeholder={t('flows:create.template.search', 'Search templates...')}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          sx={{width: 260, flexShrink: 0}}
          slotProps={{
            input: {
              startAdornment: (
                <InputAdornment position="start">
                  <Search size={16} />
                </InputAdornment>
              ),
            },
          }}
        />
        <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1}}>
          <Chip
            label={t('common:all', 'All')}
            onClick={() => setSelectedCategory(null)}
            color={selectedCategory === null ? 'primary' : 'default'}
            variant={selectedCategory === null ? 'filled' : 'outlined'}
            size="small"
          />
          {categories.map((cat) => (
            <Chip
              key={cat}
              label={CATEGORY_LABELS[cat] ?? cat}
              onClick={() => setSelectedCategory(cat)}
              color={selectedCategory === cat ? 'primary' : 'default'}
              variant={selectedCategory === cat ? 'filled' : 'outlined'}
              size="small"
            />
          ))}
        </Box>
      </Stack>

      {filteredTemplates.length === 0 && (
        <Typography variant="body2" color="text.secondary" sx={{mt: 2}}>
          {t('flows:create.template.noResults', 'No templates match your search.')}
        </Typography>
      )}

      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))',
          gap: 2,
        }}
      >
        {filteredTemplates.map((template) => {
          const isBlank = template.type === 'BLANK';
          const isSelected =
            selectedTemplate?.type === template.type && selectedTemplate?.flowType === template.flowType;
          return (
            <Card
              key={`${template.flowType}-${template.type}`}
              variant="outlined"
              sx={{
                display: 'flex',
                flexDirection: 'column',
                ...(isBlank && {
                  background: (theme) => `linear-gradient(135deg, ${theme.palette.primary.main}18 0%, transparent 60%)`,
                }),
              }}
            >
              <CardActionArea
                onClick={() => onTemplateChange(template)}
                sx={{
                  flex: 1,
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'stretch',
                  border: 1,
                  borderColor: isSelected ? 'primary.main' : 'divider',
                  transition: 'all 0.2s ease-in-out',
                  '&:hover': {
                    borderColor: 'primary.main',
                    bgcolor: isSelected ? 'action.selected' : 'action.hover',
                  },
                }}
              >
                <CardContent
                  sx={{
                    py: 2,
                    px: 2,
                    flex: 1,
                    display: 'flex',
                    flexDirection: 'column',
                    '&:last-child': {pb: 2},
                  }}
                >
                  {/* Icon */}
                  <Box sx={{mb: 1.5, color: isSelected ? 'primary.main' : 'text.secondary'}}>
                    {CATEGORY_ICONS[template.category] ?? <LayoutTemplate size={20} />}
                  </Box>

                  {/* Title + description — grows to fill space */}
                  <Box sx={{flex: 1}}>
                    <Typography variant="subtitle1" sx={{fontWeight: isBlank ? 600 : 500}}>
                      {template.display.label}
                    </Typography>
                    {template.display.description && (
                      <Typography variant="body2" color="text.secondary" sx={{mt: 0.5}}>
                        {template.display.description}
                      </Typography>
                    )}
                  </Box>

                  {/* Category chip — pinned to bottom */}
                  {!selectedCategory && (
                    <Box sx={{mt: 2}}>
                      <Chip
                        label={CATEGORY_LABELS[template.category] ?? template.category}
                        size="small"
                        variant="outlined"
                        sx={{
                          fontSize: '0.65rem',
                          height: 20,
                          color: 'text.secondary',
                          borderColor: 'divider',
                        }}
                      />
                    </Box>
                  )}
                </CardContent>
              </CardActionArea>
            </Card>
          );
        })}
      </Box>
    </Stack>
  );
}
