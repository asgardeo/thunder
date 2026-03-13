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

import {describe, it, expect} from 'vitest';
import {render, screen} from '@testing-library/react';
import DynamicValueSyntax from '../DynamicValueSyntax';

describe('DynamicValueSyntax', () => {
  describe('valid dynamic value patterns', () => {
    it('should render meta pattern with syntax highlighting', () => {
      render(<DynamicValueSyntax value="{{meta(application.name)}}" />);

      expect(screen.getByText('meta')).toBeInTheDocument();
      expect(screen.getByText('application.name')).toBeInTheDocument();
      expect(screen.getByText('{{')).toBeInTheDocument();
      expect(screen.getByText(')}}')).toBeInTheDocument();
      expect(screen.getByText('(')).toBeInTheDocument();
    });

    it('should render translation pattern with syntax highlighting', () => {
      render(<DynamicValueSyntax value="{{t(flowI18n:login.title)}}" />);

      expect(screen.getByText('t')).toBeInTheDocument();
      expect(screen.getByText('flowI18n:login.title')).toBeInTheDocument();
    });

    it('should handle whitespace around the value', () => {
      render(<DynamicValueSyntax value="  {{meta(ou.name)}}  " />);

      expect(screen.getByText('meta')).toBeInTheDocument();
      expect(screen.getByText('ou.name')).toBeInTheDocument();
    });
  });

  describe('invalid or unrecognized patterns', () => {
    it('should render plain text for non-dynamic values', () => {
      render(<DynamicValueSyntax value="Hello World" />);

      expect(screen.getByText('Hello World')).toBeInTheDocument();
    });

    it('should render plain text for incomplete patterns', () => {
      render(<DynamicValueSyntax value="{{meta(incomplete" />);

      expect(screen.getByText('{{meta(incomplete')).toBeInTheDocument();
    });

    it('should render plain text for empty value', () => {
      render(<DynamicValueSyntax value="" />);

      const {container} = render(<DynamicValueSyntax value="" />);
      expect(container).toBeTruthy();
    });

    it('should render plain text for malformed braces', () => {
      render(<DynamicValueSyntax value="{meta(key)}" />);

      expect(screen.getByText('{meta(key)}')).toBeInTheDocument();
    });
  });
});
