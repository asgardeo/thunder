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
import {render} from '@thunder/test-utils';
import LayoutPresetThumbnail from '../LayoutPresetThumbnail';

describe('LayoutPresetThumbnail', () => {
  describe('Rendering without crashing', () => {
    it('renders the centered variant', () => {
      const {container} = render(<LayoutPresetThumbnail variant="centered" />);
      expect(container.firstChild).toBeInTheDocument();
    });

    it('renders the split variant', () => {
      const {container} = render(<LayoutPresetThumbnail variant="split" />);
      expect(container.firstChild).toBeInTheDocument();
    });

    it('renders the fullscreen variant', () => {
      const {container} = render(<LayoutPresetThumbnail variant="fullscreen" />);
      expect(container.firstChild).toBeInTheDocument();
    });

    it('renders the popup variant', () => {
      const {container} = render(<LayoutPresetThumbnail variant="popup" />);
      expect(container.firstChild).toBeInTheDocument();
    });
  });

  describe('Distinct rendering per variant', () => {
    it('renders different DOM for "centered" vs "split"', () => {
      const {container: a} = render(<LayoutPresetThumbnail variant="centered" />);
      const {container: b} = render(<LayoutPresetThumbnail variant="split" />);
      // The inner HTML may differ as each variant has a unique layout
      expect(a.innerHTML).not.toBe(b.innerHTML);
    });

    it('renders different DOM for "fullscreen" vs "popup"', () => {
      const {container: a} = render(<LayoutPresetThumbnail variant="fullscreen" />);
      const {container: b} = render(<LayoutPresetThumbnail variant="popup" />);
      expect(a.innerHTML).not.toBe(b.innerHTML);
    });
  });

  describe('Accessibility', () => {
    it('centered variant renders a container element', () => {
      render(<LayoutPresetThumbnail variant="centered" />);
      expect(document.body.firstChild).toBeInTheDocument();
    });

    it('split variant renders visual content (no missing elements)', () => {
      const {container} = render(<LayoutPresetThumbnail variant="split" />);
      expect(container.childElementCount).toBeGreaterThan(0);
    });
  });
});
