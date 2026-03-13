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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import RulesProperties from '../RulesProperties';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

// Mock the SCSS
vi.mock('../RulesProperties.scss', () => ({}));

describe('RulesProperties', () => {
  describe('Rendering', () => {
    it('should render the component', () => {
      render(<RulesProperties />);

      expect(screen.getByText('flows:core.rulesProperties.description')).toBeInTheDocument();
    });

    it('should render Typography with body2 variant', () => {
      const {container} = render(<RulesProperties />);

      const typography = container.querySelector('.MuiTypography-body2');
      expect(typography).toBeInTheDocument();
    });

    it('should render within a Stack component', () => {
      const {container} = render(<RulesProperties />);

      const stack = container.querySelector('.MuiStack-root');
      expect(stack).toBeInTheDocument();
    });
  });
});
