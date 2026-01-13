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
import PlaceholderComponent from '../PlaceholderComponent';

// Mock the SCSS file
vi.mock('../PlaceholderComponent.scss', () => ({}));

describe('PlaceholderComponent', () => {
  describe('Regular Text Rendering', () => {
    it('should render plain text value in a span', () => {
      render(<PlaceholderComponent value="Hello World" />);

      expect(screen.getByText('Hello World')).toBeInTheDocument();
    });

    it('should render empty string', () => {
      const {container} = render(<PlaceholderComponent value="" />);

      const span = container.querySelector('span');
      expect(span).toBeInTheDocument();
      expect(span?.textContent).toBe('');
    });

    it('should render text with special characters', () => {
      render(<PlaceholderComponent value="Hello & <World>" />);

      expect(screen.getByText('Hello & <World>')).toBeInTheDocument();
    });
  });

  describe('i18n Pattern Detection', () => {
    it('should detect i18n pattern {{key}}', () => {
      const {container} = render(<PlaceholderComponent value="{{i18n.key}}" />);

      expect(container.querySelector('.flow-builder-display-field-i18n-placeholder')).toBeInTheDocument();
    });

    it('should detect i18n pattern with simple key', () => {
      const {container} = render(<PlaceholderComponent value="{{label}}" />);

      expect(container.querySelector('.flow-builder-display-field-i18n-placeholder')).toBeInTheDocument();
    });

    it('should display the i18n key in placeholder', () => {
      render(<PlaceholderComponent value="{{my.translation.key}}" />);

      expect(screen.getByText('{{my.translation.key}}')).toBeInTheDocument();
    });

    it('should detect i18n pattern with whitespace around', () => {
      const {container} = render(<PlaceholderComponent value="  {{key}}  " />);

      expect(container.querySelector('.flow-builder-display-field-i18n-placeholder')).toBeInTheDocument();
    });

    it('should not detect i18n pattern for partial match', () => {
      const {container} = render(<PlaceholderComponent value="text {{key}} more text" />);

      expect(container.querySelector('.flow-builder-display-field-i18n-placeholder')).not.toBeInTheDocument();
    });

    it('should not detect i18n pattern with single braces', () => {
      const {container} = render(<PlaceholderComponent value="{key}" />);

      expect(container.querySelector('.flow-builder-display-field-i18n-placeholder')).not.toBeInTheDocument();
    });

    it('should not detect i18n pattern with empty braces', () => {
      const {container} = render(<PlaceholderComponent value="{{}}" />);

      expect(container.querySelector('.flow-builder-display-field-i18n-placeholder')).not.toBeInTheDocument();
    });
  });

  describe('Children Rendering', () => {
    it('should render children when provided and value is not i18n pattern', () => {
      render(
        <PlaceholderComponent value="regular text">
          <div data-testid="child">Child Content</div>
        </PlaceholderComponent>,
      );

      expect(screen.getByTestId('child')).toBeInTheDocument();
      expect(screen.getByText('Child Content')).toBeInTheDocument();
    });

    it('should not render children when value is i18n pattern', () => {
      render(
        <PlaceholderComponent value="{{i18n.key}}">
          <div data-testid="child">Child Content</div>
        </PlaceholderComponent>,
      );

      expect(screen.queryByTestId('child')).not.toBeInTheDocument();
      expect(screen.getByText('{{i18n.key}}')).toBeInTheDocument();
    });

    it('should prefer children over value when not i18n pattern', () => {
      render(
        <PlaceholderComponent value="value text">
          <span>Children text</span>
        </PlaceholderComponent>,
      );

      expect(screen.getByText('Children text')).toBeInTheDocument();
      expect(screen.queryByText('value text')).not.toBeInTheDocument();
    });

    it('should handle null children', () => {
      render(<PlaceholderComponent value="text">{null}</PlaceholderComponent>);

      expect(screen.getByText('text')).toBeInTheDocument();
    });
  });

  describe('CSS Classes', () => {
    it('should apply i18n placeholder class for i18n pattern', () => {
      const {container} = render(<PlaceholderComponent value="{{key}}" />);

      expect(container.querySelector('.flow-builder-display-field-i18n-placeholder')).toBeInTheDocument();
      expect(container.querySelector('.flow-builder-display-field-i18n-placeholder-key')).toBeInTheDocument();
    });

    it('should not apply i18n classes for regular text', () => {
      const {container} = render(<PlaceholderComponent value="regular text" />);

      expect(container.querySelector('.flow-builder-display-field-i18n-placeholder')).not.toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle undefined-like empty value', () => {
      const {container} = render(<PlaceholderComponent value="" />);

      expect(container.querySelector('.flow-builder-display-field-i18n-placeholder')).not.toBeInTheDocument();
    });

    it('should handle numeric-like strings', () => {
      render(<PlaceholderComponent value="12345" />);

      expect(screen.getByText('12345')).toBeInTheDocument();
    });

    it('should handle newlines in text', () => {
      render(<PlaceholderComponent value={'Line1\nLine2'} />);

      // The component renders the value as-is, so we check for the exact text content
      expect(screen.getByText((content) => content.includes('Line1') && content.includes('Line2'))).toBeInTheDocument();
    });
  });
});
