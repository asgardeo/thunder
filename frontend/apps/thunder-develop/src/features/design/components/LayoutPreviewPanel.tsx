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

import {Box, CircularProgress, Stack, Typography} from '@wso2/oxygen-ui';
import {type JSX} from 'react';
import {useGetLayout} from '@thunder/shared-design';

interface LayoutPreviewPanelProps {
  layoutId: string | null;
  selectedScreen?: string | null;
  screenDraft?: Record<string, unknown> | null;
  /** When true, shows horizontal + vertical ruler overlays around the preview */
  showRulers?: boolean;
}

interface ScreenSlot {
  label: string;
  height?: number | string;
  flex?: number;
  color: string;
  details?: string[];
}

interface ScreenPreviewProps {
  screenName: string;
  slots: ScreenSlot[];
  background?: string;
  maxWidth?: number;
}

// ── Ruler components ─────────────────────────────────────────────────────────

const RULER_SIZE = 24;
const MAJOR_INTERVAL = 100;
const MINOR_INTERVAL = 50;
const LABEL_INTERVAL = 100;

function HorizontalRuler({length = 800}: {length?: number}): JSX.Element {
  const ticks: JSX.Element[] = [];
  for (let i = 0; i <= length; i += MINOR_INTERVAL) {
    const isMajor = i % MAJOR_INTERVAL === 0;
    ticks.push(
      <Box
        key={`tick-${i}`}
        sx={{
          position: 'absolute',
          left: i,
          bottom: 0,
          width: '1px',
          height: isMajor ? 11 : 6,
          bgcolor: isMajor ? 'rgba(0,0,0,0.45)' : 'rgba(0,0,0,0.22)',
        }}
      />,
    );
    if (i % LABEL_INTERVAL === 0 && i > 0) {
      ticks.push(
        <Typography
          key={`lbl-${i}`}
          sx={{
            position: 'absolute',
            left: i + 3,
            top: 3,
            fontSize: '0.5rem',
            lineHeight: 1,
            color: 'text.secondary',
            userSelect: 'none',
            fontFamily: 'monospace',
          }}
        >
          {i}
        </Typography>,
      );
    }
  }
  return (
    <Box
      sx={{
        position: 'relative',
        height: RULER_SIZE,
        flex: 1,
        bgcolor: 'grey.50',
        borderBottom: '1px solid',
        borderColor: 'divider',
        overflow: 'hidden',
        flexShrink: 0,
      }}
    >
      {ticks}
    </Box>
  );
}

function VerticalRuler({length = 800}: {length?: number}): JSX.Element {
  const ticks: JSX.Element[] = [];
  for (let i = 0; i <= length; i += MINOR_INTERVAL) {
    const isMajor = i % MAJOR_INTERVAL === 0;
    ticks.push(
      <Box
        key={`tick-${i}`}
        sx={{
          position: 'absolute',
          top: i,
          right: 0,
          height: '1px',
          width: isMajor ? 11 : 6,
          bgcolor: isMajor ? 'rgba(0,0,0,0.45)' : 'rgba(0,0,0,0.22)',
        }}
      />,
    );
    if (i % LABEL_INTERVAL === 0 && i > 0) {
      ticks.push(
        <Typography
          key={`lbl-${i}`}
          sx={{
            position: 'absolute',
            top: i + 3,
            left: 3,
            fontSize: '0.5rem',
            lineHeight: 1,
            color: 'text.secondary',
            userSelect: 'none',
            fontFamily: 'monospace',
            writingMode: 'vertical-rl',
            textOrientation: 'mixed',
            transform: 'rotate(180deg)',
          }}
        >
          {i}
        </Typography>,
      );
    }
  }
  return (
    <Box
      sx={{
        position: 'relative',
        width: RULER_SIZE,
        flexShrink: 0,
        bgcolor: 'grey.50',
        borderRight: '1px solid',
        borderColor: 'divider',
        overflow: 'hidden',
      }}
    >
      {ticks}
    </Box>
  );
}

function RulerLayout({children}: {children: JSX.Element}): JSX.Element {
  return (
    <Box sx={{height: '100%', display: 'flex', flexDirection: 'column'}}>
      {/* Top row: corner + horizontal ruler */}
      <Box sx={{display: 'flex', flexShrink: 0}}>
        <Box
          sx={{
            width: RULER_SIZE,
            height: RULER_SIZE,
            bgcolor: 'grey.100',
            borderRight: '1px solid',
            borderBottom: '1px solid',
            borderColor: 'divider',
            flexShrink: 0,
          }}
        />
        <HorizontalRuler />
      </Box>
      {/* Main row: vertical ruler + content */}
      <Box sx={{flex: 1, display: 'flex', minHeight: 0}}>
        <VerticalRuler />
        <Box sx={{flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column'}}>{children}</Box>
      </Box>
    </Box>
  );
}

// ── Preview components ───────────────────────────────────────────────────────

function ScreenPreview({screenName, slots, background, maxWidth = 360}: ScreenPreviewProps): JSX.Element {
  return (
    <Box sx={{textAlign: 'center'}}>
      <Typography variant="caption" color="text.secondary" sx={{mb: 1, display: 'block'}}>
        {screenName}
      </Typography>
      <Box
        sx={{
          width: '100%',
          maxWidth,
          mx: 'auto',
          border: '1.5px solid',
          borderColor: 'divider',
          borderRadius: 2,
          overflow: 'hidden',
          bgcolor: background ?? 'grey.50',
          minHeight: 200,
          display: 'flex',
          flexDirection: 'column',
          boxShadow: '0 2px 12px rgba(0,0,0,0.08)',
        }}
      >
        {slots.map((slot, idx) => (
          <Box
            key={idx}
            sx={{
              height: typeof slot.height === 'number' ? slot.height : undefined,
              flex: slot.flex ?? undefined,
              bgcolor: slot.color,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              px: 1,
              py: 0.75,
              borderBottom: idx < slots.length - 1 ? '1px dashed rgba(0,0,0,0.1)' : 'none',
              minHeight: typeof slot.height === 'number' ? slot.height : 48,
              position: 'relative',
            }}
          >
            <Typography variant="caption" sx={{fontWeight: 600, fontSize: '0.65rem', opacity: 0.7}}>
              {slot.label}
            </Typography>
            {slot.details?.map((d, i) => (
              <Typography key={i} variant="caption" sx={{fontSize: '0.6rem', opacity: 0.5, lineHeight: 1.3}}>
                {d}
              </Typography>
            ))}
            {/* Dimension marker for height */}
            {typeof slot.height === 'number' && (
              <Box
                sx={{
                  position: 'absolute',
                  right: 6,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 0.25,
                }}
              >
                <Typography sx={{fontSize: '0.5rem', fontFamily: 'monospace', opacity: 0.45, color: 'text.secondary'}}>
                  {slot.height}px
                </Typography>
              </Box>
            )}
          </Box>
        ))}
      </Box>
    </Box>
  );
}

function parseScreenSlots(screenDef: Record<string, unknown>): ScreenSlot[] {
  const slots: ScreenSlot[] = [];
  const slotsData = screenDef.slots as Record<string, unknown> | undefined;

  if (!slotsData) {
    return [{label: 'Content', flex: 1, color: 'rgba(99, 102, 241, 0.1)'}];
  }

  if (slotsData.header) {
    const header = slotsData.header as Record<string, unknown>;
    const h = (header.height as number | undefined) ?? 60;
    const details: string[] = [];
    if (header.showLogo) details.push('Logo');
    if (header.showLanguageSelector) details.push('Lang selector');
    if (header.showBackButton) details.push('Back button');
    slots.push({label: 'Header', height: h, color: 'rgba(99, 102, 241, 0.12)', details});
  }

  if (slotsData.main) {
    const main = slotsData.main as Record<string, unknown>;
    const container = main.container as Record<string, unknown> | undefined;
    const details: string[] = [];
    if (container?.maxWidth) details.push(`max-width: ${String(container.maxWidth)}px`);
    if (container?.padding) details.push(`padding: ${String(container.padding)}px`);
    slots.push({label: 'Main', flex: 1, color: 'rgba(16, 185, 129, 0.1)', details});
  } else {
    slots.push({label: 'Main', flex: 1, color: 'rgba(16, 185, 129, 0.1)'});
  }

  if (slotsData.footer) {
    const footer = slotsData.footer as Record<string, unknown>;
    const h = (footer.height as number | undefined) ?? 48;
    const details: string[] = [];
    if (footer.showLinks) details.push('Links');
    slots.push({label: 'Footer', height: h, color: 'rgba(245, 158, 11, 0.12)', details});
  }

  return slots;
}

function PreviewChrome({label}: {label: string}): JSX.Element {
  return (
    <Box
      sx={{
        px: 3,
        py: 1.5,
        borderBottom: '1px solid',
        borderColor: 'divider',
        display: 'flex',
        alignItems: 'center',
        gap: 1,
        flexShrink: 0,
        bgcolor: 'background.paper',
      }}
    >
      <Box sx={{width: 8, height: 8, borderRadius: '50%', bgcolor: '#fc5c57'}} />
      <Box sx={{width: 8, height: 8, borderRadius: '50%', bgcolor: '#febc2e'}} />
      <Box sx={{width: 8, height: 8, borderRadius: '50%', bgcolor: '#29c840'}} />
      <Box
        sx={{
          flex: 1,
          mx: 2,
          height: 22,
          bgcolor: 'action.hover',
          borderRadius: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Typography variant="caption" color="text.disabled" sx={{fontSize: 10}}>
          {label}
        </Typography>
      </Box>
    </Box>
  );
}

const SLOT_LEGEND = [
  {label: 'Header', color: 'rgba(99, 102, 241, 0.4)'},
  {label: 'Main', color: 'rgba(16, 185, 129, 0.4)'},
  {label: 'Footer', color: 'rgba(245, 158, 11, 0.4)'},
] as const;

function SlotLegend(): JSX.Element {
  return (
    <Box
      sx={{
        px: 3,
        py: 1.5,
        borderTop: '1px solid',
        borderColor: 'divider',
        display: 'flex',
        gap: 1.5,
        flexWrap: 'wrap',
        alignItems: 'center',
        flexShrink: 0,
        bgcolor: 'background.paper',
      }}
    >
      <Typography variant="caption" color="text.secondary">
        Slots:
      </Typography>
      {SLOT_LEGEND.map(({label, color}) => (
        <Stack key={label} direction="row" alignItems="center" spacing={0.5}>
          <Box sx={{width: 10, height: 10, bgcolor: color, borderRadius: 0.5}} />
          <Typography variant="caption" color="text.secondary">
            {label}
          </Typography>
        </Stack>
      ))}
    </Box>
  );
}

// ── Main content renderer ────────────────────────────────────────────────────

function LayoutPreviewContent({
  layoutId,
  selectedScreen,
  screenDraft,
}: LayoutPreviewPanelProps): JSX.Element {
  const {data: layout, isLoading} = useGetLayout(layoutId ?? '');

  if (!layoutId) {
    return (
      <Box sx={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center'}}>
        <Typography variant="body2" color="text.secondary">Select a layout to preview</Typography>
      </Box>
    );
  }

  if (isLoading) {
    return (
      <Box sx={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center'}}>
        <CircularProgress size={32} />
      </Box>
    );
  }

  if (!layout) {
    return (
      <Box sx={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center'}}>
        <Typography variant="body2" color="text.secondary">Failed to load layout</Typography>
      </Box>
    );
  }

  const allScreens = (layout.layout?.screens ?? {}) as Record<string, Record<string, unknown>>;

  // ── Focused single-screen view ───────────────────────────────────────────

  if (selectedScreen) {
    const serverDef = allScreens[selectedScreen] ?? {};
    const liveDef = screenDraft ?? serverDef;
    const baseName = (serverDef.extends as string | undefined);
    const baseDef = baseName ? allScreens[baseName] : undefined;

    const mergedDef = baseDef
      ? {
          ...baseDef,
          ...liveDef,
          slots: {
            ...(baseDef.slots as Record<string, unknown> | undefined),
            ...(liveDef.slots as Record<string, unknown> | undefined),
          },
        }
      : liveDef;

    const slots = parseScreenSlots(mergedDef);
    const bgValue = (mergedDef.background as Record<string, unknown> | undefined)?.value as string | undefined;
    const mainSlot = (mergedDef.slots as Record<string, unknown> | undefined)?.main as Record<string, unknown> | undefined;
    const mainContainer = mainSlot?.container as Record<string, unknown> | undefined;
    const focusMaxWidth = (mainContainer?.maxWidth as number | undefined) ?? 400;

    return (
      <Box sx={{height: '100%', display: 'flex', flexDirection: 'column'}}>
        <PreviewChrome label={`${layout.displayName} — ${selectedScreen}`} />
        <Box
          sx={{
            flex: 1,
            overflowY: 'auto',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            p: 4,
            gap: 1,
            bgcolor: 'grey.50',
          }}
        >
          {baseName && (
            <Typography variant="caption" color="text.secondary" sx={{fontSize: '0.7rem'}}>
              extends <strong>{baseName}</strong>
            </Typography>
          )}
          <Box sx={{width: '100%', maxWidth: Math.max(focusMaxWidth, 280)}}>
            <ScreenPreview
              screenName={selectedScreen}
              slots={slots}
              background={bgValue}
              maxWidth={Math.max(focusMaxWidth, 280)}
            />
          </Box>
        </Box>
        <SlotLegend />
      </Box>
    );
  }

  // ── All-screens overview ─────────────────────────────────────────────────

  const baseScreens = Object.entries(allScreens).filter(([, def]) => !def.extends);
  const derivedScreens = Object.entries(allScreens).filter(([, def]) => !!def.extends);
  const authDef = allScreens['auth'];
  const authBackground = (authDef?.background as Record<string, unknown> | undefined)?.value as string | undefined;

  return (
    <Box sx={{height: '100%', display: 'flex', flexDirection: 'column'}}>
      <PreviewChrome label={`${layout.displayName} — Layout Preview`} />
      <Box sx={{flex: 1, overflowY: 'auto', p: 4, bgcolor: 'grey.50'}}>
        {baseScreens.length > 0 && (
          <Box mb={4}>
            <Typography variant="overline" sx={{color: 'text.secondary', fontSize: '0.65rem', mb: 2, display: 'block'}}>
              Base layout
            </Typography>
            <Stack direction="row" flexWrap="wrap" gap={3} justifyContent="center">
              {baseScreens.map(([name, def]) => (
                <ScreenPreview
                  key={name}
                  screenName={name}
                  slots={parseScreenSlots(def)}
                  background={(def.background as Record<string, unknown> | undefined)?.value as string | undefined}
                />
              ))}
            </Stack>
          </Box>
        )}

        {derivedScreens.length > 0 && (
          <Box>
            <Typography variant="overline" sx={{color: 'text.secondary', fontSize: '0.65rem', mb: 2, display: 'block'}}>
              Screen variants
            </Typography>
            <Stack direction="row" flexWrap="wrap" gap={2} justifyContent="center">
              {derivedScreens.map(([name, def]) => {
                const mergedSlots = parseScreenSlots({
                  ...authDef,
                  slots: {
                    ...(authDef?.slots as Record<string, unknown> | undefined),
                    ...(def.slots as Record<string, unknown> | undefined),
                  },
                });
                const mainContainer = (def.slots as Record<string, unknown> | undefined)?.main as Record<string, unknown> | undefined;
                const containerOverride = mainContainer?.container as Record<string, unknown> | undefined;
                const maxWidth = (containerOverride?.maxWidth as number | undefined) ?? 360;
                return (
                  <ScreenPreview
                    key={name}
                    screenName={name}
                    slots={mergedSlots}
                    background={authBackground}
                    maxWidth={maxWidth}
                  />
                );
              })}
            </Stack>
          </Box>
        )}
      </Box>
      <SlotLegend />
    </Box>
  );
}

// ── Main component ───────────────────────────────────────────────────────────

export default function LayoutPreviewPanel(props: LayoutPreviewPanelProps): JSX.Element {
  const content = <LayoutPreviewContent {...props} />;

  if (props.showRulers) {
    return <RulerLayout>{content}</RulerLayout>;
  }

  return content;
}
