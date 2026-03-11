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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {render} from '@thunder/test-utils/browser';
import ConfigureExperience from '../ConfigureExperience';
import {ApplicationCreateFlowSignInApproach} from '../../../models/application-create-flow';

describe('ConfigureExperience', () => {
  const mockOnApproachChange = vi.fn();
  const mockOnReadyChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the component with both approach options', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
        />,
      );

      await expect.element(page.getByText('Embedded sign-in/sign-up components in your app')).toBeInTheDocument();
    });

    it('should select INBUILT approach by default', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
        />,
      );

      const inbuiltRadio = page.getByRole('radio').all()[0];
      expect(inbuiltRadio).toBeChecked();
    });

    it('should select EMBEDDED approach when prop is set', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.EMBEDDED}
          onApproachChange={mockOnApproachChange}
        />,
      );

      const embeddedRadio = page.getByRole('radio').all()[1];
      expect(embeddedRadio).toBeChecked();
    });
  });

  describe('User Interactions', () => {
    it('should call onApproachChange when INBUILT is clicked', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.EMBEDDED}
          onApproachChange={mockOnApproachChange}
        />,
      );

      const inbuiltRadio = page.getByRole('radio').all()[0];
      await userEvent.click(inbuiltRadio);

      expect(mockOnApproachChange).toHaveBeenCalledWith(ApplicationCreateFlowSignInApproach.INBUILT);
    });

    it('should call onApproachChange when EMBEDDED is clicked', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
        />,
      );

      const embeddedRadio = page.getByRole('radio').all()[1];
      await userEvent.click(embeddedRadio);

      expect(mockOnApproachChange).toHaveBeenCalledWith(ApplicationCreateFlowSignInApproach.EMBEDDED);
    });
  });

  describe('Ready State', () => {
    it('should call onReadyChange with true on mount', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
          onReadyChange={mockOnReadyChange}
        />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });
  });

  describe('User Types Selection', () => {
    const mockUserTypes = [
      {id: '1', name: 'Internal', ouId: 'INTERNAL', allowSelfRegistration: true},
      {id: '2', name: 'External', ouId: 'EXTERNAL', allowSelfRegistration: false},
    ];
    const mockOnUserTypesChange = vi.fn();

    beforeEach(() => {
      mockOnUserTypesChange.mockClear();
    });

    it('should render user types selection when userTypes prop is provided', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
          userTypes={mockUserTypes}
          onUserTypesChange={mockOnUserTypesChange}
        />,
      );

      expect(
        page.getByText('User Access'),
      ).toBeInTheDocument();
    });

    it('should not render user types selection when userTypes prop is undefined', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
        />,
      );

      expect(
        page.getByText('User Access'),
      ).not.toBeInTheDocument();
    });

    it('should not render user types selection when only one user type exists', async () => {
      const singleUserType = [{id: '1', name: 'Internal', ouId: 'INTERNAL', allowSelfRegistration: true}];

      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
          userTypes={singleUserType}
          onUserTypesChange={mockOnUserTypesChange}
        />,
      );

      expect(
        page.getByText('User Access'),
      ).not.toBeInTheDocument();
    });

    it('should render user type cards and allow selection', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
          userTypes={mockUserTypes}
          selectedUserTypes={['Internal']}
          onUserTypesChange={mockOnUserTypesChange}
        />,
      );

      // Should render user type cards
      await expect.element(page.getByText('Internal')).toBeInTheDocument();
      await expect.element(page.getByText('External')).toBeInTheDocument();

      // Click External card to add it to selection
      const externalCard = page.getByText('External').element().closest('[class*="MuiCard"]');
      expect(externalCard).toBeInTheDocument();
      await userEvent.click(externalCard!);

      expect(mockOnUserTypesChange).toHaveBeenCalledWith(['Internal', 'External']);
    });

    it('should remove user type from selection when clicking selected card', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
          userTypes={mockUserTypes}
          selectedUserTypes={['Internal', 'External']}
          onUserTypesChange={mockOnUserTypesChange}
        />,
      );

      // Click Internal card to remove it from selection
      const internalCard = page.getByText('Internal').element().closest('[class*="MuiCard"]');
      await userEvent.click(internalCard!);

      expect(mockOnUserTypesChange).toHaveBeenCalledWith(['External']);
    });

    it('should auto-select first user type when none selected', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
          userTypes={mockUserTypes}
          selectedUserTypes={[]}
          onUserTypesChange={mockOnUserTypesChange}
        />,
      );

      expect(mockOnUserTypesChange).toHaveBeenCalledWith(['Internal']);
    });

    it('should call onReadyChange with false when no user types selected (with multiple available)', async () => {
      const mockOnReadyChangeLocal = vi.fn();
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
          onReadyChange={mockOnReadyChangeLocal}
          userTypes={mockUserTypes}
          selectedUserTypes={[]}
          onUserTypesChange={mockOnUserTypesChange}
        />,
      );

      // Initially false because no user types selected
      expect(mockOnReadyChangeLocal).toHaveBeenCalledWith(false);
    });

    it('should call onReadyChange with true when user types are selected', async () => {
      const mockOnReadyChangeLocal = vi.fn();
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
          onReadyChange={mockOnReadyChangeLocal}
          userTypes={mockUserTypes}
          selectedUserTypes={['Internal']}
          onUserTypesChange={mockOnUserTypesChange}
        />,
      );

      expect(mockOnReadyChangeLocal).toHaveBeenCalledWith(true);
    });
  });

  describe('User Types Autocomplete (5+ user types)', () => {
    const manyUserTypes = [
      {id: '1', name: 'Type1', ouId: 'TYPE1', allowSelfRegistration: true},
      {id: '2', name: 'Type2', ouId: 'TYPE2', allowSelfRegistration: false},
      {id: '3', name: 'Type3', ouId: 'TYPE3', allowSelfRegistration: true},
      {id: '4', name: 'Type4', ouId: 'TYPE4', allowSelfRegistration: false},
      {id: '5', name: 'Type5', ouId: 'TYPE5', allowSelfRegistration: true},
    ];
    const mockOnUserTypesChange = vi.fn();

    beforeEach(() => {
      mockOnUserTypesChange.mockClear();
    });

    it('should render autocomplete when 5 or more user types exist', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
          userTypes={manyUserTypes}
          selectedUserTypes={['Type1']}
          onUserTypesChange={mockOnUserTypesChange}
        />,
      );

      // Should have an autocomplete input
      await expect.element(page.getByRole('combobox')).toBeInTheDocument();
    });

    it('should allow selecting user types from autocomplete', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
          userTypes={manyUserTypes}
          selectedUserTypes={['Type1']}
          onUserTypesChange={mockOnUserTypesChange}
        />,
      );

      const autocomplete = page.getByRole('combobox');
      await userEvent.click(autocomplete);

      // Should show options
      const type2Option = page.getByText('Type2');
      await userEvent.click(type2Option);

      expect(mockOnUserTypesChange).toHaveBeenCalledWith(['Type1', 'Type2']);
    });

    it('should show error state when no user types selected with autocomplete', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
          userTypes={manyUserTypes}
          selectedUserTypes={[]}
          onUserTypesChange={mockOnUserTypesChange}
        />,
      );

      // The TextField should have error state
      const textField = page.getByRole('combobox');
      expect(textField.element().closest('.MuiAutocomplete-root')?.querySelector('.Mui-error')).toBeInTheDocument();
    });
  });

  describe('Card Click Handlers', () => {
    it('should call onApproachChange when clicking INBUILT card', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.EMBEDDED}
          onApproachChange={mockOnApproachChange}
        />,
      );

      // Find the INBUILT card by its radio button (first radio) and click the card itself
      const inbuiltRadio = page.getByRole('radio').all()[0];
      const inbuiltCard = inbuiltRadio.element().closest('[class*="MuiCard"]');
      await userEvent.click(inbuiltCard!);

      expect(mockOnApproachChange).toHaveBeenCalledWith(ApplicationCreateFlowSignInApproach.INBUILT);
    });

    it('should call onApproachChange when clicking EMBEDDED card', async () => {
      await render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
        />,
      );

      // Find the EMBEDDED card by its title text and click the card itself
      const embeddedTitle = page.getByText('Embedded sign-in/sign-up components in your app');
      const embeddedCard = embeddedTitle.element().closest('[class*="MuiCard"]');
      await userEvent.click(embeddedCard!);

      expect(mockOnApproachChange).toHaveBeenCalledWith(ApplicationCreateFlowSignInApproach.EMBEDDED);
    });
  });
});
