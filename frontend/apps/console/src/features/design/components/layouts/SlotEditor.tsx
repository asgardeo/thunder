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
import {useTranslation} from 'react-i18next';
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
  const {t} = useTranslation('design');
  const container = slot.container as Record<string, unknown> | undefined;
  const layout = slot.layout as Record<string, unknown> | undefined;
  const position = slot.position as Record<string, unknown> | undefined;

  const num = (v: unknown): number => Number(v) || 0;
  const str = (v: unknown): string => String((v ?? '') as string | number | boolean);
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
            label={t('layouts.forms.slot_editor.fields.height.label', 'Height')}
            value={num(slot.height)}
            min={0}
            max={200}
            onChange={(v) => onUpdate([name, 'height'], v)}
          />
        )}
        {slot.padding !== undefined && (
          <SliderRow
            label={t('layouts.forms.slot_editor.fields.padding.label', 'Padding')}
            value={num(slot.padding)}
            min={0}
            max={64}
            onChange={(v) => onUpdate([name, 'padding'], v)}
          />
        )}
        {slot.showLogo !== undefined && (
          <SwitchRow
            label={t('layouts.forms.slot_editor.fields.show_logo.label', 'Show logo')}
            value={bool(slot.showLogo)}
            onChange={(v) => onUpdate([name, 'showLogo'], v)}
          />
        )}
        {slot.showBackButton !== undefined && (
          <SwitchRow
            label={t('layouts.forms.slot_editor.fields.back_button.label', 'Back button')}
            value={bool(slot.showBackButton)}
            onChange={(v) => onUpdate([name, 'showBackButton'], v)}
          />
        )}
        {slot.showLanguageSelector !== undefined && (
          <SwitchRow
            label={t('layouts.forms.slot_editor.fields.language_selector.label', 'Language selector')}
            value={bool(slot.showLanguageSelector)}
            onChange={(v) => onUpdate([name, 'showLanguageSelector'], v)}
          />
        )}
        {slot.showLinks !== undefined && (
          <SwitchRow
            label={t('layouts.forms.slot_editor.fields.links.label', 'Links')}
            value={bool(slot.showLinks)}
            onChange={(v) => onUpdate([name, 'showLinks'], v)}
          />
        )}

        {position && (
          <>
            <SectionLabel>{t('layouts.forms.slot_editor.position.title', 'Position')}</SectionLabel>
            {position.anchor !== undefined && (
              <SelectRow
                label={t('layouts.forms.slot_editor.fields.anchor.label', 'Anchor')}
                value={str(position.anchor)}
                options={[
                  {value: 'center', label: t('layouts.forms.slot_editor.fields.anchor.options.center.label', 'Center')},
                  {value: 'left', label: t('layouts.forms.slot_editor.fields.anchor.options.left.label', 'Left')},
                  {value: 'right', label: t('layouts.forms.slot_editor.fields.anchor.options.right.label', 'Right')},
                ]}
                onChange={(v) => onUpdate([name, 'position', 'anchor'], v)}
              />
            )}
            {position.verticalAlign !== undefined && (
              <SelectRow
                label={t('layouts.forms.slot_editor.fields.v_align.label', 'V-align')}
                value={str(position.verticalAlign)}
                options={[
                  {value: 'top', label: t('layouts.forms.slot_editor.fields.v_align.options.top.label', 'Top')},
                  {
                    value: 'middle',
                    label: t('layouts.forms.slot_editor.fields.v_align.options.middle.label', 'Middle'),
                  },
                  {
                    value: 'bottom',
                    label: t('layouts.forms.slot_editor.fields.v_align.options.bottom.label', 'Bottom'),
                  },
                ]}
                onChange={(v) => onUpdate([name, 'position', 'verticalAlign'], v)}
              />
            )}
          </>
        )}

        {container && (
          <>
            <SectionLabel>{t('layouts.forms.slot_editor.container.title', 'Container')}</SectionLabel>
            {container.maxWidth !== undefined && (
              <SliderRow
                label={t('layouts.forms.slot_editor.fields.max_width.label', 'Max width')}
                value={num(container.maxWidth)}
                min={200}
                max={900}
                onChange={(v) => onUpdate([name, 'container', 'maxWidth'], v)}
              />
            )}
            {container.padding !== undefined && (
              <SliderRow
                label={t('layouts.forms.slot_editor.fields.padding.label', 'Padding')}
                value={num(container.padding)}
                min={0}
                max={64}
                onChange={(v) => onUpdate([name, 'container', 'padding'], v)}
              />
            )}
            {container.borderRadius !== undefined && (
              <SliderRow
                label={t('layouts.forms.slot_editor.fields.border_radius.label', 'Border radius')}
                value={num(container.borderRadius)}
                min={0}
                max={32}
                onChange={(v) => onUpdate([name, 'container', 'borderRadius'], v)}
              />
            )}
            {container.elevation !== undefined && (
              <SelectRow
                label={t('layouts.forms.slot_editor.fields.elevation.label', 'Elevation')}
                value={str(container.elevation)}
                options={['0', '1', '2', '3', '4'].map((v) => ({value: v, label: v}))}
                onChange={(v) => onUpdate([name, 'container', 'elevation'], Number(v))}
              />
            )}
            {container.background !== undefined && (
              <SelectRow
                label={t('layouts.forms.slot_editor.fields.background.label', 'Background')}
                value={str(container.background)}
                options={[
                  {
                    value: 'paper',
                    label: t('layouts.forms.slot_editor.fields.background.options.paper.label', 'Paper'),
                  },
                  {
                    value: 'default',
                    label: t('layouts.forms.slot_editor.fields.background.options.default.label', 'Default'),
                  },
                  {
                    value: 'transparent',
                    label: t('layouts.forms.slot_editor.fields.background.options.transparent.label', 'Transparent'),
                  },
                ]}
                onChange={(v) => onUpdate([name, 'container', 'background'], v)}
              />
            )}
          </>
        )}

        {layout && (
          <>
            <SectionLabel>{t('layouts.forms.slot_editor.layout.title', 'Layout')}</SectionLabel>
            {layout.type !== undefined && (
              <SelectRow
                label={t('layouts.forms.slot_editor.fields.type.label', 'Type')}
                value={str(layout.type)}
                options={[
                  {value: 'stack', label: t('layouts.forms.slot_editor.fields.type.options.stack.label', 'Stack')},
                  {value: 'grid', label: t('layouts.forms.slot_editor.fields.type.options.grid.label', 'Grid')},
                ]}
                onChange={(v) => onUpdate([name, 'layout', 'type'], v)}
              />
            )}
            {layout.direction !== undefined && (
              <SelectRow
                label={t('layouts.forms.slot_editor.fields.direction.label', 'Direction')}
                value={str(layout.direction)}
                options={[
                  {
                    value: 'column',
                    label: t('layouts.forms.slot_editor.fields.direction.options.column.label', 'Column'),
                  },
                  {value: 'row', label: t('layouts.forms.slot_editor.fields.direction.options.row.label', 'Row')},
                ]}
                onChange={(v) => onUpdate([name, 'layout', 'direction'], v)}
              />
            )}
            {layout.gap !== undefined && (
              <SliderRow
                label={t('layouts.forms.slot_editor.fields.gap.label', 'Gap')}
                value={num(layout.gap)}
                min={0}
                max={64}
                onChange={(v) => onUpdate([name, 'layout', 'gap'], v)}
              />
            )}
            {layout.justify !== undefined && (
              <SelectRow
                label={t('layouts.forms.slot_editor.fields.justify.label', 'Justify')}
                value={str(layout.justify)}
                options={[
                  {
                    value: 'flex-start',
                    label: t('layouts.forms.slot_editor.fields.justify.options.start.label', 'Start'),
                  },
                  {
                    value: 'center',
                    label: t('layouts.forms.slot_editor.fields.justify.options.center.label', 'Center'),
                  },
                  {value: 'flex-end', label: t('layouts.forms.slot_editor.fields.justify.options.end.label', 'End')},
                  {
                    value: 'space-between',
                    label: t('layouts.forms.slot_editor.fields.justify.options.between.label', 'Between'),
                  },
                ]}
                onChange={(v) => onUpdate([name, 'layout', 'justify'], v)}
              />
            )}
            {layout.align !== undefined && (
              <SelectRow
                label={t('layouts.forms.slot_editor.fields.align.label', 'Align')}
                value={str(layout.align)}
                options={[
                  {
                    value: 'flex-start',
                    label: t('layouts.forms.slot_editor.fields.align.options.start.label', 'Start'),
                  },
                  {value: 'center', label: t('layouts.forms.slot_editor.fields.align.options.center.label', 'Center')},
                  {value: 'flex-end', label: t('layouts.forms.slot_editor.fields.align.options.end.label', 'End')},
                  {
                    value: 'stretch',
                    label: t('layouts.forms.slot_editor.fields.align.options.stretch.label', 'Stretch'),
                  },
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
