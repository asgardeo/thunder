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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import FormRequiresViewDialog, {type DropScenario} from '../FormRequiresViewDialog';

describe('FormRequiresViewDialog', () => {
  const mockOnClose = vi.fn();
  const mockOnConfirm = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Dialog Visibility', () => {
    it('should render dialog when open is true', async () => {
      await render(
        <FormRequiresViewDialog open scenario="form-on-canvas" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      await expect.element(page.getByRole('dialog')).toBeInTheDocument();
    });

    it('should not render dialog content when open is false', async () => {
      await render(
        <FormRequiresViewDialog
          open={false}
          scenario="form-on-canvas"
          onClose={mockOnClose}
          onConfirm={mockOnConfirm}
        />,
      );

      await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();
    });
  });

  describe('Form on Canvas Scenario', () => {
    it('should display correct title for form-on-canvas', async () => {
      await render(
        <FormRequiresViewDialog open scenario="form-on-canvas" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      await expect.element(page.getByText('Form requires a View')).toBeInTheDocument();
    });

    it('should display correct description for form-on-canvas', async () => {
      await render(
        <FormRequiresViewDialog open scenario="form-on-canvas" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      await expect.element(page.getByText('A Form component must be inside a View step.')).toBeInTheDocument();
    });

    it('should display correct alert message for form-on-canvas', async () => {
      await render(
        <FormRequiresViewDialog open scenario="form-on-canvas" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      await expect.element(page.getByText('A View step will be created to contain this Form.')).toBeInTheDocument();
    });

    it('should display correct confirm button text for form-on-canvas', async () => {
      await render(
        <FormRequiresViewDialog open scenario="form-on-canvas" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      await expect.element(page.getByRole('button', {name: 'Create View'})).toBeInTheDocument();
    });
  });

  describe('Input on Canvas Scenario', () => {
    it('should display correct title for input-on-canvas', async () => {
      await render(
        <FormRequiresViewDialog
          open
          scenario="input-on-canvas"
          onClose={mockOnClose}
          onConfirm={mockOnConfirm}
        />,
      );

      await expect.element(page.getByText('Input requires a View and Form')).toBeInTheDocument();
    });

    it('should display correct description for input-on-canvas', async () => {
      await render(
        <FormRequiresViewDialog
          open
          scenario="input-on-canvas"
          onClose={mockOnClose}
          onConfirm={mockOnConfirm}
        />,
      );

      expect(
        page.getByText('An Input component must be inside a Form, which is inside a View step.'),
      ).toBeInTheDocument();
    });

    it('should display correct confirm button text for input-on-canvas', async () => {
      await render(
        <FormRequiresViewDialog
          open
          scenario="input-on-canvas"
          onClose={mockOnClose}
          onConfirm={mockOnConfirm}
        />,
      );

      await expect.element(page.getByRole('button', {name: 'Create View and Form'})).toBeInTheDocument();
    });
  });

  describe('Input on View Scenario', () => {
    it('should display correct title for input-on-view', async () => {
      await render(
        <FormRequiresViewDialog open scenario="input-on-view" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      await expect.element(page.getByText('Input requires a Form')).toBeInTheDocument();
    });

    it('should display correct description for input-on-view', async () => {
      await render(
        <FormRequiresViewDialog open scenario="input-on-view" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      await expect.element(page.getByText('An Input component must be inside a Form.')).toBeInTheDocument();
    });

    it('should display correct confirm button text for input-on-view', async () => {
      await render(
        <FormRequiresViewDialog open scenario="input-on-view" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      await expect.element(page.getByRole('button', {name: 'Create Form'})).toBeInTheDocument();
    });
  });

  describe('Widget on Canvas Scenario', () => {
    it('should display correct title for widget-on-canvas', async () => {
      await render(
        <FormRequiresViewDialog
          open
          scenario="widget-on-canvas"
          onClose={mockOnClose}
          onConfirm={mockOnConfirm}
        />,
      );

      await expect.element(page.getByText('Widget requires a View')).toBeInTheDocument();
    });

    it('should display correct description for widget-on-canvas', async () => {
      await render(
        <FormRequiresViewDialog
          open
          scenario="widget-on-canvas"
          onClose={mockOnClose}
          onConfirm={mockOnConfirm}
        />,
      );

      await expect.element(page.getByText('A Widget component must be inside a View step.')).toBeInTheDocument();
    });

    it('should display correct confirm button text for widget-on-canvas', async () => {
      await render(
        <FormRequiresViewDialog
          open
          scenario="widget-on-canvas"
          onClose={mockOnClose}
          onConfirm={mockOnConfirm}
        />,
      );

      await expect.element(page.getByRole('button', {name: 'Create View'})).toBeInTheDocument();
    });
  });

  describe('Button Actions', () => {
    it('should call onClose when Cancel button is clicked', async () => {
      await render(
        <FormRequiresViewDialog open scenario="form-on-canvas" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      const cancelButton = page.getByRole('button', {name: 'Cancel'});
      await userEvent.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should call onConfirm when confirm button is clicked', async () => {
      await render(
        <FormRequiresViewDialog open scenario="form-on-canvas" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      const confirmButton = page.getByRole('button', {name: 'Create View'});
      await userEvent.click(confirmButton);

      expect(mockOnConfirm).toHaveBeenCalledTimes(1);
    });

    it('should display Cancel button for all scenarios', async () => {
      const scenarios: DropScenario[] = ['form-on-canvas', 'input-on-canvas', 'input-on-view', 'widget-on-canvas'];

      // eslint-disable-next-line no-restricted-syntax, no-await-in-loop
      for (const scenario of scenarios) {
        // eslint-disable-next-line no-await-in-loop
        const {unmount} = await render(
          <FormRequiresViewDialog open scenario={scenario} onClose={mockOnClose} onConfirm={mockOnConfirm} />,
        );

        // eslint-disable-next-line no-await-in-loop
        await expect.element(page.getByRole('button', {name: 'Cancel'})).toBeInTheDocument();
        // eslint-disable-next-line no-await-in-loop
        await unmount();
      }
    });
  });

  describe('Alert Component', () => {
    it('should render an info alert for all scenarios', async () => {
      await render(
        <FormRequiresViewDialog open scenario="form-on-canvas" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      const alert = page.getByRole('alert');
      expect(alert).toBeInTheDocument();
    });

    it('should display alert with info severity', async () => {
      await render(
        <FormRequiresViewDialog open scenario="input-on-canvas" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      const alert = page.getByRole('alert');
      expect(alert).toHaveClass('MuiAlert-standardInfo');
    });
  });

  describe('Dialog Structure', () => {
    it('should render dialog title', async () => {
      await render(
        <FormRequiresViewDialog open scenario="form-on-canvas" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      await expect.element(page.getByRole('heading')).toBeInTheDocument();
    });

    it('should render two buttons (Cancel and Confirm)', async () => {
      await render(
        <FormRequiresViewDialog open scenario="form-on-canvas" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      const buttons = page.getByRole('button');
      expect(buttons).toHaveLength(2);
    });

    it('should render confirm button with contained variant', async () => {
      await render(
        <FormRequiresViewDialog open scenario="form-on-canvas" onClose={mockOnClose} onConfirm={mockOnConfirm} />,
      );

      const confirmButton = page.getByRole('button', {name: 'Create View'});
      expect(confirmButton).toHaveClass('MuiButton-contained');
    });
  });

  describe('Scenario Coverage', () => {
    it('should handle all four drop scenarios correctly', async () => {
      const scenarios: DropScenario[] = ['form-on-canvas', 'input-on-canvas', 'input-on-view', 'widget-on-canvas'];

      // eslint-disable-next-line no-restricted-syntax, no-await-in-loop
      for (const scenario of scenarios) {
        // eslint-disable-next-line no-await-in-loop
        const {unmount} = await render(
          <FormRequiresViewDialog open scenario={scenario} onClose={mockOnClose} onConfirm={mockOnConfirm} />,
        );

        // Each scenario should render without errors
        // eslint-disable-next-line no-await-in-loop
        await expect.element(page.getByRole('dialog')).toBeInTheDocument();
        // eslint-disable-next-line no-await-in-loop
        await expect.element(page.getByRole('alert')).toBeInTheDocument();
        expect(page.getByRole('button')).toHaveLength(2);

        // eslint-disable-next-line no-await-in-loop
        await unmount();
      }
    });
  });
});
