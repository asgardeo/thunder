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

import {Box} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import type {ColorSchemeColors} from '@thunder/shared-design';
import ColorEditRow from '../common/ColorEditRow';
import ConfigCard from '../common/ConfigCard';

export interface ColorBuilderContentProps {
  colors: ColorSchemeColors;
  onUpdate: (updater: (c: ColorSchemeColors) => void) => void;
}

/**
 * ColorBuilderContent - Theme builder section for editing color scheme colors.
 * Organizes colors into ConfigCard sections: Brand Colors, Backgrounds, Text.
 */
export default function ColorBuilderContent({colors, onUpdate}: ColorBuilderContentProps): JSX.Element {
  return (
    <Box>
      <ConfigCard title="Brand Colors">
        <ColorEditRow
          label="Primary"
          value={colors.primary.main}
          onChange={(v) =>
            onUpdate((c) => {
              c.primary.main = v;
            })
          }
        />
        <ColorEditRow
          label="Primary Foreground"
          value={colors.primary.contrastText}
          onChange={(v) =>
            onUpdate((c) => {
              c.primary.contrastText = v;
            })
          }
        />
        <ColorEditRow
          label="Primary Dark"
          value={colors.primary.dark}
          onChange={(v) =>
            onUpdate((c) => {
              c.primary.dark = v;
            })
          }
        />
        <ColorEditRow
          label="Secondary"
          value={colors.secondary.main}
          onChange={(v) =>
            onUpdate((c) => {
              c.secondary.main = v;
            })
          }
        />
      </ConfigCard>

      {colors.background && (
        <ConfigCard title="Backgrounds" defaultOpen={false}>
          <ColorEditRow
            label="Background"
            value={colors.background.default}
            onChange={(v) =>
              onUpdate((c) => {
                if (c.background) c.background.default = v;
              })
            }
          />
          <ColorEditRow
            label="Surface"
            value={colors.background.paper}
            onChange={(v) =>
              onUpdate((c) => {
                if (c.background) c.background.paper = v;
              })
            }
          />
        </ConfigCard>
      )}

      {colors.text && (
        <ConfigCard title="Text" defaultOpen={false}>
          <ColorEditRow
            label="Text Primary"
            value={colors.text.primary}
            onChange={(v) =>
              onUpdate((c) => {
                if (c.text) c.text.primary = v;
              })
            }
          />
          <ColorEditRow
            label="Text Secondary"
            value={colors.text.secondary}
            onChange={(v) =>
              onUpdate((c) => {
                if (c.text) c.text.secondary = v;
              })
            }
          />
        </ConfigCard>
      )}
    </Box>
  );
}
