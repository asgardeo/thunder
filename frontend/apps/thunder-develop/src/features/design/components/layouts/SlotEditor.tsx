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

import {Card, CardContent, Typography} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import SectionLabel from '../common/SectionLabel';
import SelectRow from '../common/SelectRow';
import SliderRow from '../common/SliderRow';
import SwitchRow from '../common/SwitchRow';

export interface SlotEditorProps {
  name: string;
  slot: Record<string, unknown>;
  onUpdate: (path: string[], value: unknown) => void;
}

/**
 * SlotEditor - Edits properties of a single layout slot.
 * Handles container, position, and layout properties for slot configuration.
 */
export default function SlotEditor({name, slot, onUpdate}: SlotEditorProps): JSX.Element {
  const container = slot.container as Record<string, unknown> | undefined;
  const layout = slot.layout as Record<string, unknown> | undefined;
  const position = slot.position as Record<string, unknown> | undefined;

  const num = (v: unknown): number => Number(v) || 0;
  const str = (v: unknown): string => String(v ?? '');
  const bool = (v: unknown): boolean => Boolean(v);

  return (
    <Card sx={{mb: 1.5, borderLeft: '2px solid', borderColor: 'primary.light'}}>
      <CardContent sx={{pt: 1, '&:last-child': {pb: 1}}}>
        <Typography
          variant="caption"
          sx={{
            fontWeight: 700,
            fontSize: '0.7rem',
            textTransform: 'uppercase',
            color: 'primary.main',
            letterSpacing: '0.05em',
            display: 'block',
            mb: 0.75,
          }}
        >
          {name}
        </Typography>

        {slot.height !== undefined && (
          <SliderRow
            label="Height"
            value={num(slot.height)}
            min={0}
            max={200}
            onChange={(v) => onUpdate([name, 'height'], v)}
          />
        )}
        {slot.padding !== undefined && (
          <SliderRow
            label="Padding"
            value={num(slot.padding)}
            min={0}
            max={64}
            onChange={(v) => onUpdate([name, 'padding'], v)}
          />
        )}
        {slot.showLogo !== undefined && (
          <SwitchRow label="Show logo" value={bool(slot.showLogo)} onChange={(v) => onUpdate([name, 'showLogo'], v)} />
        )}
        {slot.showBackButton !== undefined && (
          <SwitchRow
            label="Back button"
            value={bool(slot.showBackButton)}
            onChange={(v) => onUpdate([name, 'showBackButton'], v)}
          />
        )}
        {slot.showLanguageSelector !== undefined && (
          <SwitchRow
            label="Language selector"
            value={bool(slot.showLanguageSelector)}
            onChange={(v) => onUpdate([name, 'showLanguageSelector'], v)}
          />
        )}
        {slot.showLinks !== undefined && (
          <SwitchRow label="Links" value={bool(slot.showLinks)} onChange={(v) => onUpdate([name, 'showLinks'], v)} />
        )}

        {position && (
          <>
            <SectionLabel>Position</SectionLabel>
            {position.anchor !== undefined && (
              <SelectRow
                label="Anchor"
                value={str(position.anchor)}
                options={[
                  {value: 'center', label: 'Center'},
                  {value: 'left', label: 'Left'},
                  {value: 'right', label: 'Right'},
                ]}
                onChange={(v) => onUpdate([name, 'position', 'anchor'], v)}
              />
            )}
            {position.verticalAlign !== undefined && (
              <SelectRow
                label="V-align"
                value={str(position.verticalAlign)}
                options={[
                  {value: 'top', label: 'Top'},
                  {value: 'middle', label: 'Middle'},
                  {value: 'bottom', label: 'Bottom'},
                ]}
                onChange={(v) => onUpdate([name, 'position', 'verticalAlign'], v)}
              />
            )}
          </>
        )}

        {container && (
          <>
            <SectionLabel>Container</SectionLabel>
            {container.maxWidth !== undefined && (
              <SliderRow
                label="Max width"
                value={num(container.maxWidth)}
                min={200}
                max={900}
                onChange={(v) => onUpdate([name, 'container', 'maxWidth'], v)}
              />
            )}
            {container.padding !== undefined && (
              <SliderRow
                label="Padding"
                value={num(container.padding)}
                min={0}
                max={64}
                onChange={(v) => onUpdate([name, 'container', 'padding'], v)}
              />
            )}
            {container.borderRadius !== undefined && (
              <SliderRow
                label="Border radius"
                value={num(container.borderRadius)}
                min={0}
                max={32}
                onChange={(v) => onUpdate([name, 'container', 'borderRadius'], v)}
              />
            )}
            {container.elevation !== undefined && (
              <SelectRow
                label="Elevation"
                value={str(container.elevation)}
                options={['0', '1', '2', '3', '4'].map((v) => ({value: v, label: v}))}
                onChange={(v) => onUpdate([name, 'container', 'elevation'], Number(v))}
              />
            )}
            {container.background !== undefined && (
              <SelectRow
                label="Background"
                value={str(container.background)}
                options={[
                  {value: 'paper', label: 'Paper'},
                  {value: 'default', label: 'Default'},
                  {value: 'transparent', label: 'Transparent'},
                ]}
                onChange={(v) => onUpdate([name, 'container', 'background'], v)}
              />
            )}
          </>
        )}

        {layout && (
          <>
            <SectionLabel>Layout</SectionLabel>
            {layout.type !== undefined && (
              <SelectRow
                label="Type"
                value={str(layout.type)}
                options={[
                  {value: 'stack', label: 'Stack'},
                  {value: 'grid', label: 'Grid'},
                ]}
                onChange={(v) => onUpdate([name, 'layout', 'type'], v)}
              />
            )}
            {layout.direction !== undefined && (
              <SelectRow
                label="Direction"
                value={str(layout.direction)}
                options={[
                  {value: 'column', label: 'Column'},
                  {value: 'row', label: 'Row'},
                ]}
                onChange={(v) => onUpdate([name, 'layout', 'direction'], v)}
              />
            )}
            {layout.gap !== undefined && (
              <SliderRow
                label="Gap"
                value={num(layout.gap)}
                min={0}
                max={64}
                onChange={(v) => onUpdate([name, 'layout', 'gap'], v)}
              />
            )}
            {layout.justify !== undefined && (
              <SelectRow
                label="Justify"
                value={str(layout.justify)}
                options={[
                  {value: 'flex-start', label: 'Start'},
                  {value: 'center', label: 'Center'},
                  {value: 'flex-end', label: 'End'},
                  {value: 'space-between', label: 'Between'},
                ]}
                onChange={(v) => onUpdate([name, 'layout', 'justify'], v)}
              />
            )}
            {layout.align !== undefined && (
              <SelectRow
                label="Align"
                value={str(layout.align)}
                options={[
                  {value: 'flex-start', label: 'Start'},
                  {value: 'center', label: 'Center'},
                  {value: 'flex-end', label: 'End'},
                  {value: 'stretch', label: 'Stretch'},
                ]}
                onChange={(v) => onUpdate([name, 'layout', 'align'], v)}
              />
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}
