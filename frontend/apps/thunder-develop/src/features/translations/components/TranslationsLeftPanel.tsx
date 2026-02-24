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

import {Box, Divider, List, ListItemButton, ListItemText, Typography} from '@wso2/oxygen-ui';
import {useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import BuilderPanelHeader from '../../../components/BuilderLayout/BuilderPanelHeader';
import AddLanguageDialog from './AddLanguageDialog';
import LanguageAutocomplete from './LanguageAutocomplete';

function formatNamespace(ns: string): string {
  return ns
    .replace(/([A-Z])/g, ' $1')
    .replace(/^./, (c) => c.toUpperCase())
    .trim();
}

export interface TranslationsLeftPanelProps {
  languages: string[];
  languagesLoading: boolean;
  selectedLanguage: string | null;
  onLanguageChange: (lang: string | null) => void;
  serverUrl: string;
  namespaces: string[];
  selectedNamespace: string;
  onNamespaceChange: (ns: string) => void;
  onPanelToggle: () => void;
  onLanguageAdded: () => void;
}

export default function TranslationsLeftPanel({
  languages,
  languagesLoading,
  selectedLanguage,
  onLanguageChange,
  serverUrl,
  namespaces,
  selectedNamespace,
  onNamespaceChange,
  onPanelToggle,
  onLanguageAdded,
}: TranslationsLeftPanelProps): JSX.Element {
  const {t} = useTranslation('translations');
  const [addDialogOpen, setAddDialogOpen] = useState(false);

  const handleAddLanguage = () => {
    setAddDialogOpen(true);
  };

  const handleLanguageAdded = (code: string) => {
    setAddDialogOpen(false);
    onLanguageAdded();
    onLanguageChange(code);
  };

  return (
    <>
      <BuilderPanelHeader onPanelToggle={onPanelToggle} hidePanelTooltip={t('language.selectPlaceholder')} />

      <Box sx={{px: 0.5, pb: 1}}>
        <LanguageAutocomplete
          options={languages}
          value={selectedLanguage}
          onChange={onLanguageChange}
          onAddNew={handleAddLanguage}
          loading={languagesLoading}
        />
      </Box>

      <Divider />

      <Box sx={{pt: 1.5}}>
        <Typography variant="overline" color="text.secondary" sx={{px: 0.5, display: 'block', mb: 0.5, fontSize: '0.68rem', letterSpacing: '0.08em'}}>
          {t('namespace.label')}
        </Typography>
        <List dense disablePadding>
          {namespaces.map((ns) => (
            <ListItemButton
              key={ns}
              selected={selectedNamespace === ns}
              onClick={() => onNamespaceChange(ns)}
              sx={{borderRadius: 1, px: 1.5, py: 0.75}}
            >
              <ListItemText
                primary={formatNamespace(ns)}
                primaryTypographyProps={{
                  variant: 'body2',
                  sx: {fontWeight: selectedNamespace === ns ? 600 : 400},
                }}
              />
            </ListItemButton>
          ))}
        </List>
      </Box>

      <AddLanguageDialog
        open={addDialogOpen}
        onClose={() => setAddDialogOpen(false)}
        onAdded={handleLanguageAdded}
        serverUrl={serverUrl}
      />
    </>
  );
}
