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

import {useState} from 'react';
import {Box, TextField, Chip, IconButton} from '@wso2/oxygen-ui';
import {Plus} from '@wso2/oxygen-ui-icons-react';

/**
 * Array input component for adding multiple values as chips
 */
function ArrayFieldInput({
  value,
  onChange,
  fieldLabel,
}: {
  value: string[];
  onChange: (value: string[]) => void;
  fieldLabel: string;
}) {
  const [inputValue, setInputValue] = useState('');
  const currentValue = Array.isArray(value) ? value : [];

  const handleAdd = () => {
    if (inputValue.trim()) {
      onChange([...currentValue, inputValue.trim()]);
      setInputValue('');
    }
  };

  const handleDelete = (indexToDelete: number) => {
    onChange(currentValue.filter((_, index) => index !== indexToDelete));
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleAdd();
    }
  };

  return (
    <Box>
      <Box sx={{display: 'flex', gap: 1, mb: 1}}>
        <TextField
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={`Add ${fieldLabel.toLowerCase()}`}
          fullWidth
          size="small"
          variant="outlined"
        />
        <IconButton size="small" onClick={handleAdd} disabled={!inputValue.trim()}>
          <Plus size={16} />
        </IconButton>
      </Box>
      <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1}}>
        {currentValue.length > 0 &&
          currentValue.map((item, itemIndex) => (
            <Chip
              key={`chip-${item}`}
              label={String(item)}
              onDelete={() => handleDelete(itemIndex)}
              variant="outlined"
              size="medium"
            />
          ))}
      </Box>
    </Box>
  );
}

export default ArrayFieldInput;
