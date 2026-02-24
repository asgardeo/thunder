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

import {Box, Button, Stack, TextField} from '@wso2/oxygen-ui';
import {Plus} from '@wso2/oxygen-ui-icons-react';
import {useState, type JSX} from 'react';

export interface AddScreenRowProps {
  baseScreens: string[];
  onAdd: (name: string, extendsBase: string) => void;
}

export default function AddScreenRow({baseScreens, onAdd}: AddScreenRowProps): JSX.Element {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState('');

  const handleAdd = (): void => {
    const trimmed = name.trim();
    if (!trimmed || baseScreens.length === 0) return;
    onAdd(trimmed, baseScreens[0]);
    setName('');
    setOpen(false);
  };

  if (!open) {
    return (
      <Button
        size="small"
        variant="text"
        startIcon={<Plus size={13} />}
        onClick={() => setOpen(true)}
        sx={{
          textTransform: 'none',
          fontSize: '0.8rem',
          color: 'text.secondary',
          justifyContent: 'flex-start',
          px: 1.5,
          py: 0.75,
          borderRadius: 1.5,
          width: '100%',
          '&:hover': {color: 'primary.main'},
        }}
      >
        Add screen
      </Button>
    );
  }

  return (
    <Box sx={{px: 0.5, py: 0.5, display: 'flex', flexDirection: 'column', gap: 0.75}}>
      <TextField
        autoFocus
        value={name}
        onChange={(e) => setName(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') handleAdd();
          if (e.key === 'Escape') {
            setOpen(false);
            setName('');
          }
        }}
        placeholder="Screen name…"
        size="small"
        fullWidth
        sx={{
          '& .MuiInputBase-root': {
            fontSize: '0.8125rem',
          },
        }}
      />
      <Stack direction="row" spacing={0.5}>
        <Button
          size="small"
          variant="contained"
          disableElevation
          onClick={handleAdd}
          sx={{textTransform: 'none', fontSize: '0.75rem', flex: 1, py: 0.4, borderRadius: 1}}
        >
          Add
        </Button>
        <Button
          size="small"
          variant="text"
          onClick={() => {
            setOpen(false);
            setName('');
          }}
          sx={{textTransform: 'none', fontSize: '0.75rem', py: 0.4, borderRadius: 1}}
        >
          Cancel
        </Button>
      </Stack>
    </Box>
  );
}
