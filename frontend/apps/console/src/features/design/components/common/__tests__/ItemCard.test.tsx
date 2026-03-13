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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@thunder/test-utils';
import userEvent from '@testing-library/user-event';
import ItemCard from '../ItemCard';

vi.mock('react-i18next', async () => {
  const actual = await vi.importActual<typeof import('react-i18next')>('react-i18next');
  return {
    ...actual,
    useTranslation: () => ({t: (key: string) => key}),
  };
});

describe('ItemCard', () => {
  const thumbnail = <div data-testid="thumb">Thumbnail</div>;

  describe('Rendering', () => {
    it('renders the item name', () => {
      render(<ItemCard thumbnail={thumbnail} name="Default Theme" onClick={vi.fn()} />);
      expect(screen.getByText('Default Theme')).toBeInTheDocument();
    });

    it('renders the thumbnail content', () => {
      render(<ItemCard thumbnail={thumbnail} name="My Theme" onClick={vi.fn()} />);
      expect(screen.getByTestId('thumb')).toBeInTheDocument();
    });

    it('renders different names correctly', () => {
      render(<ItemCard thumbnail={thumbnail} name="Ocean Blue" onClick={vi.fn()} />);
      expect(screen.getByText('Ocean Blue')).toBeInTheDocument();
    });
  });

  describe('Interaction', () => {
    it('calls onClick when the card is clicked', async () => {
      const onClick = vi.fn();
      const user = userEvent.setup();
      render(<ItemCard thumbnail={thumbnail} name="Theme A" onClick={onClick} />);

      await user.click(screen.getByText('Theme A'));

      expect(onClick).toHaveBeenCalledOnce();
    });

    it('calls onClick only when the card is clicked', async () => {
      const onClick = vi.fn();
      const user = userEvent.setup();
      render(
        <div>
          <ItemCard thumbnail={thumbnail} name="Theme B" onClick={onClick} />
          <button type="button">Other</button>
        </div>,
      );

      await user.click(screen.getByText('Other'));

      expect(onClick).not.toHaveBeenCalled();
    });
  });
});
