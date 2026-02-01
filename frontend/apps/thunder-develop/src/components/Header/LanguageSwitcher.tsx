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

import {Menu, MenuItem, ListItemText, Tooltip} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useState} from 'react';
import {Languages} from '@wso2/oxygen-ui-icons-react';
import {useLanguage, type SupportedLanguage} from '@thunder/i18n';
import MenuButton from '../Sidebar/MenuButton';

export default function LanguageSwitcher(): JSX.Element {
  const {availableLanguages, currentLanguage, setLanguage} = useLanguage();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLanguageChange = (language: SupportedLanguage) => {
    setLanguage(language).catch(() => {
      // TODO: Handle language change error
    });
    handleClose();
  };

  return (
    <>
      <Tooltip title="Change language">
        <MenuButton
          onClick={handleClick}
          aria-controls={open ? 'language-menu' : undefined}
          aria-haspopup="true"
          aria-expanded={open ? 'true' : undefined}
          aria-label="Change language"
        >
          <Languages strokeWidth={1} />
        </MenuButton>
      </Tooltip>
      <Menu
        anchorEl={anchorEl}
        id="language-menu"
        open={open}
        onClose={handleClose}
        transformOrigin={{horizontal: 'right', vertical: 'top'}}
        anchorOrigin={{horizontal: 'right', vertical: 'bottom'}}
        slotProps={{
          paper: {
            elevation: 0,
            sx: {
              overflow: 'visible',
              filter: 'drop-shadow(0px 2px 8px rgba(0,0,0,0.32))',
              mt: 1.5,
              minWidth: 180,
            },
          },
        }}
      >
        {availableLanguages.map((lang) => (
          <MenuItem
            key={lang.code}
            onClick={() => handleLanguageChange(lang.code)}
            selected={currentLanguage === lang.code}
          >
            <ListItemText primary={lang.nativeName} secondary={lang.name !== lang.nativeName ? lang.name : undefined} />
          </MenuItem>
        ))}
      </Menu>
    </>
  );
}
