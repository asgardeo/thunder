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

import {describe, it, expect} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {BezierEdgeIcon, SmoothStepEdgeIcon, StepEdgeIcon} from '../EdgeStyleIcons';

describe('EdgeStyleIcons', () => {
  describe('BezierEdgeIcon', () => {
    it('should render an SVG element', async () => {
      const {container} = await render(<BezierEdgeIcon />);

      const svg = container.querySelector('svg');
      expect(svg).toBeInTheDocument();
    });

    it('should have correct dimensions', async () => {
      const {container} = await render(<BezierEdgeIcon />);

      const svg = container.querySelector('svg');
      expect(svg).toHaveAttribute('width', '20');
      expect(svg).toHaveAttribute('height', '20');
    });

    it('should have correct viewBox', async () => {
      const {container} = await render(<BezierEdgeIcon />);

      const svg = container.querySelector('svg');
      expect(svg).toHaveAttribute('viewBox', '0 0 24 24');
    });

    it('should have correct stroke properties', async () => {
      const {container} = await render(<BezierEdgeIcon />);

      const svg = container.querySelector('svg');
      expect(svg).toHaveAttribute('stroke', 'currentColor');
      expect(svg).toHaveAttribute('stroke-width', '2');
      expect(svg).toHaveAttribute('fill', 'none');
    });

    it('should contain a path element with bezier curve', async () => {
      const {container} = await render(<BezierEdgeIcon />);

      const path = container.querySelector('path');
      expect(path).toBeInTheDocument();
      expect(path).toHaveAttribute('d', 'M4 12 C 8 4, 16 20, 20 12');
      expect(path).toHaveAttribute('stroke-linecap', 'round');
    });
  });

  describe('SmoothStepEdgeIcon', () => {
    it('should render an SVG element', async () => {
      const {container} = await render(<SmoothStepEdgeIcon />);

      const svg = container.querySelector('svg');
      expect(svg).toBeInTheDocument();
    });

    it('should have correct dimensions', async () => {
      const {container} = await render(<SmoothStepEdgeIcon />);

      const svg = container.querySelector('svg');
      expect(svg).toHaveAttribute('width', '20');
      expect(svg).toHaveAttribute('height', '20');
    });

    it('should have correct viewBox', async () => {
      const {container} = await render(<SmoothStepEdgeIcon />);

      const svg = container.querySelector('svg');
      expect(svg).toHaveAttribute('viewBox', '0 0 24 24');
    });

    it('should have correct stroke properties', async () => {
      const {container} = await render(<SmoothStepEdgeIcon />);

      const svg = container.querySelector('svg');
      expect(svg).toHaveAttribute('stroke', 'currentColor');
      expect(svg).toHaveAttribute('stroke-width', '2');
      expect(svg).toHaveAttribute('fill', 'none');
    });

    it('should contain a path element with smooth step curve', async () => {
      const {container} = await render(<SmoothStepEdgeIcon />);

      const path = container.querySelector('path');
      expect(path).toBeInTheDocument();
      expect(path).toHaveAttribute('d', 'M4 6 H 10 Q 12 6, 12 8 V 16 Q 12 18, 14 18 H 20');
      expect(path).toHaveAttribute('stroke-linecap', 'round');
    });
  });

  describe('StepEdgeIcon', () => {
    it('should render an SVG element', async () => {
      const {container} = await render(<StepEdgeIcon />);

      const svg = container.querySelector('svg');
      expect(svg).toBeInTheDocument();
    });

    it('should have correct dimensions', async () => {
      const {container} = await render(<StepEdgeIcon />);

      const svg = container.querySelector('svg');
      expect(svg).toHaveAttribute('width', '20');
      expect(svg).toHaveAttribute('height', '20');
    });

    it('should have correct viewBox', async () => {
      const {container} = await render(<StepEdgeIcon />);

      const svg = container.querySelector('svg');
      expect(svg).toHaveAttribute('viewBox', '0 0 24 24');
    });

    it('should have correct stroke properties', async () => {
      const {container} = await render(<StepEdgeIcon />);

      const svg = container.querySelector('svg');
      expect(svg).toHaveAttribute('stroke', 'currentColor');
      expect(svg).toHaveAttribute('stroke-width', '2');
      expect(svg).toHaveAttribute('fill', 'none');
    });

    it('should contain a path element with step curve', async () => {
      const {container} = await render(<StepEdgeIcon />);

      const path = container.querySelector('path');
      expect(path).toBeInTheDocument();
      expect(path).toHaveAttribute('d', 'M4 6 H 12 V 18 H 20');
      expect(path).toHaveAttribute('stroke-linecap', 'round');
      expect(path).toHaveAttribute('stroke-linejoin', 'miter');
    });
  });

  describe('Icon Consistency', () => {
    it('all icons should have the same dimensions', async () => {
      const {container: bezierContainer} = await render(<BezierEdgeIcon />);
      const {container: smoothStepContainer} = await render(<SmoothStepEdgeIcon />);
      const {container: stepContainer} = await render(<StepEdgeIcon />);

      const bezierSvg = bezierContainer.querySelector('svg');
      const smoothStepSvg = smoothStepContainer.querySelector('svg');
      const stepSvg = stepContainer.querySelector('svg');

      expect(bezierSvg?.getAttribute('width')).toBe(smoothStepSvg?.getAttribute('width'));
      expect(smoothStepSvg?.getAttribute('width')).toBe(stepSvg?.getAttribute('width'));
      expect(bezierSvg?.getAttribute('height')).toBe(smoothStepSvg?.getAttribute('height'));
      expect(smoothStepSvg?.getAttribute('height')).toBe(stepSvg?.getAttribute('height'));
    });

    it('all icons should have the same stroke width', async () => {
      const {container: bezierContainer} = await render(<BezierEdgeIcon />);
      const {container: smoothStepContainer} = await render(<SmoothStepEdgeIcon />);
      const {container: stepContainer} = await render(<StepEdgeIcon />);

      const bezierSvg = bezierContainer.querySelector('svg');
      const smoothStepSvg = smoothStepContainer.querySelector('svg');
      const stepSvg = stepContainer.querySelector('svg');

      expect(bezierSvg?.getAttribute('stroke-width')).toBe(smoothStepSvg?.getAttribute('stroke-width'));
      expect(smoothStepSvg?.getAttribute('stroke-width')).toBe(stepSvg?.getAttribute('stroke-width'));
    });

    it('all icons should use currentColor for stroke', async () => {
      const {container: bezierContainer} = await render(<BezierEdgeIcon />);
      const {container: smoothStepContainer} = await render(<SmoothStepEdgeIcon />);
      const {container: stepContainer} = await render(<StepEdgeIcon />);

      const bezierSvg = bezierContainer.querySelector('svg');
      const smoothStepSvg = smoothStepContainer.querySelector('svg');
      const stepSvg = stepContainer.querySelector('svg');

      expect(bezierSvg).toHaveAttribute('stroke', 'currentColor');
      expect(smoothStepSvg).toHaveAttribute('stroke', 'currentColor');
      expect(stepSvg).toHaveAttribute('stroke', 'currentColor');
    });
  });
});
