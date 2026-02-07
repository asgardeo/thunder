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

import type {JSX} from 'react';
import {IconButton, InputAdornment, TextField} from '@mui/material';
import {Search as SearchIcon, X as CloseIcon} from '@wso2/oxygen-ui-icons-react';

/**
 * Props for the SearchField component.
 */
export interface SearchFieldProps {
  /** Current search value */
  value: string;
  /** Callback when search value changes */
  onChange: (value: string) => void;
  /** Placeholder text */
  placeholder?: string;
  /** Whether to autofocus the input */
  autoFocus?: boolean;
  /** Whether the input should take full width */
  fullWidth?: boolean;
}

/**
 * Search input field with clear button and search icon.
 * The component is controlled and does not implement debouncing internally -
 * wrap the onChange value with useDebounce hook if needed.
 *
 * @param props - Component props
 * @returns SearchField component
 *
 * @example
 * ```tsx
 * const [searchQuery, setSearchQuery] = useState('');
 * const debouncedQuery = useDebounce(searchQuery, 300);
 *
 * // Use debouncedQuery for filtering
 * const filteredItems = items.filter(item =>
 *   item.name.toLowerCase().includes(debouncedQuery.toLowerCase())
 * );
 *
 * return (
 *   <>
 *     <SearchField
 *       value={searchQuery}
 *       onChange={setSearchQuery}
 *       placeholder="Search themes..."
 *       fullWidth
 *     />
 *     {filteredItems.map(item => <Card key={item.id}>...</Card>)}
 *   </>
 * );
 * ```
 */
export default function SearchField({
  value,
  onChange,
  placeholder = 'Search...',
  autoFocus = false,
  fullWidth = false,
}: SearchFieldProps): JSX.Element {
  const handleClear = (): void => {
    onChange('');
  };

  return (
    <TextField
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      autoFocus={autoFocus}
      fullWidth={fullWidth}
      size="small"
      InputProps={{
        startAdornment: (
          <InputAdornment position="start">
            <SearchIcon size={20} />
          </InputAdornment>
        ),
        endAdornment: value ? (
          <InputAdornment position="end">
            <IconButton
              aria-label="Clear search"
              onClick={handleClear}
              edge="end"
              size="small"
              sx={{mr: -0.5}}
            >
              <CloseIcon size={18} />
            </IconButton>
          </InputAdornment>
        ) : null,
      }}
    />
  );
}
