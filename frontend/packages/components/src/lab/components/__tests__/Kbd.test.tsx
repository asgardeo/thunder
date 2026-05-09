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

import {render, screen} from '@testing-library/react';
import {describe, it, expect} from 'vitest';
import Kbd from '../Kbd';

describe('Kbd', () => {
  describe('Rendering', () => {
    it('should render the children text', () => {
      render(<Kbd>Enter</Kbd>);

      expect(screen.getByText('Enter')).toBeInTheDocument();
    });

    it('should use a kbd HTML element', () => {
      render(<Kbd>Tab</Kbd>);

      const element = screen.getByText('Tab');

      expect(element.tagName.toLowerCase()).toBe('kbd');
    });

    it('should render multiple characters', () => {
      render(<Kbd>Ctrl+K</Kbd>);

      expect(screen.getByText('Ctrl+K')).toBeInTheDocument();
    });

    it('should render nested elements as children', () => {
      render(
        <Kbd>
          <strong>⌘</strong>
        </Kbd>,
      );

      expect(screen.getByText('⌘')).toBeInTheDocument();
    });

    it('should apply display inline-block style', () => {
      render(<Kbd>Space</Kbd>);

      const element = screen.getByText('Space');

      expect(element).toHaveStyle({display: 'inline-block'});
    });

    it('should apply white-space nowrap style', () => {
      render(<Kbd>Shift+Enter</Kbd>);

      const element = screen.getByText('Shift+Enter');

      expect(element).toHaveStyle({whiteSpace: 'nowrap'});
    });

    it('should render multiple Kbd instances independently', () => {
      render(
        <>
          <Kbd>Ctrl</Kbd>
          <Kbd>C</Kbd>
        </>,
      );

      expect(screen.getByText('Ctrl').tagName.toLowerCase()).toBe('kbd');
      expect(screen.getByText('C').tagName.toLowerCase()).toBe('kbd');
    });
  });
});
