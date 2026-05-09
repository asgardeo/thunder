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

import {getCnPrefix} from '@thunderid/utils';
import {Box, Typography} from '@wso2/oxygen-ui';
import {useEffect, useRef, useState, type JSX, type ReactNode} from 'react';

export interface ElementInspectorProps {
  enabled: boolean;
  children: ReactNode;
  onSelectSelector?: (selector: string) => void;
}

interface HighlightInfo {
  rect: DOMRect;
  classes: string[];
  tag: string;
}

const HIGHLIGHT_COLOR = 'rgba(59, 130, 246, 0.15)';
const BORDER_COLOR = 'rgba(59, 130, 246, 0.7)';

/**
 * Returns product name prefixed CSS classes from the element.
 * Since the preview now renders the same adapters as the Gate app,
 * product name prefixed classes are present directly in the DOM.
 */
function getProductNamePrefixedClasses(el: HTMLElement): string[] {
  const prefix = getCnPrefix();
  return Array.from(el.classList).filter((c) => c.startsWith(prefix));
}

/**
 * Picks the best product name prefixed class to use as a CSS selector.
 * Prefers classes with `--` (BEM modifier pattern).
 */
function pickBestClass(classes: string[]): string | undefined {
  const prefix = getCnPrefix();
  return classes.find((c) => c.startsWith(prefix) && c.includes('--')) ?? classes[0];
}

/**
 * Wraps preview content and provides element inspection on hover.
 * When enabled, hovering over elements highlights them and shows their CSS classes.
 * Clicking an element copies the best CSS selector to clipboard.
 */
export default function ElementInspector({
  enabled,
  children,
  onSelectSelector = undefined,
}: ElementInspectorProps): JSX.Element {
  const containerRef = useRef<HTMLDivElement>(null);
  const [highlight, setHighlight] = useState<HighlightInfo | null>(null);
  const [copiedText, setCopiedText] = useState<string | null>(null);
  const onSelectSelectorRef = useRef(onSelectSelector);

  useEffect(() => {
    onSelectSelectorRef.current = onSelectSelector;
  });

  useEffect(() => {
    const container = containerRef.current;
    if (!container || !enabled) return undefined;

    const handleMouseOver = (e: MouseEvent): void => {
      let target = e.target as HTMLElement;
      if (!container.contains(target)) return;
      if (target === container) {
        setHighlight(null);
        return;
      }

      // Walk up to the nearest element with product name prefixed classes so structural
      // wrapper divs (Box, ThemeProvider, etc.) are skipped by the inspector.
      while (target && target !== container && getProductNamePrefixedClasses(target).length === 0) {
        if (!target.parentElement) break;
        target = target.parentElement;
      }
      if (!target || target === container) {
        setHighlight(null);
        return;
      }

      const classes = getProductNamePrefixedClasses(target);
      const rect = target.getBoundingClientRect();
      const containerRect = container.getBoundingClientRect();

      setHighlight({
        rect: new DOMRect(rect.x - containerRect.x, rect.y - containerRect.y, rect.width, rect.height),
        classes,
        tag: target.tagName.toLowerCase(),
      });
    };

    const handleMouseOut = (e: MouseEvent): void => {
      const related = e.relatedTarget as HTMLElement | null;
      if (!related || !container.contains(related)) {
        setHighlight(null);
      }
    };

    const handleClick = (e: MouseEvent): void => {
      e.preventDefault();
      e.stopPropagation();

      // Walk up to nearest element with product name prefixed class (same logic as hover)
      let target = e.target as HTMLElement;
      while (target && target !== container && getProductNamePrefixedClasses(target).length === 0) {
        if (!target.parentElement) break;
        target = target.parentElement;
      }
      if (!target || target === container) return;

      const classes = getProductNamePrefixedClasses(target);
      const bestClass = pickBestClass(classes);

      if (bestClass) {
        const selector = `.${bestClass}`;
        onSelectSelectorRef.current?.(selector);
        navigator.clipboard.writeText(selector).then(
          () => {
            setCopiedText(selector);
            setTimeout(() => setCopiedText(null), 1500);
          },
          () => {
            // Clipboard API may fail — show what was selected anyway
            setCopiedText(selector);
            setTimeout(() => setCopiedText(null), 1500);
          },
        );
      }
    };

    container.addEventListener('mouseover', handleMouseOver);
    container.addEventListener('mouseout', handleMouseOut);
    container.addEventListener('click', handleClick, true);

    return () => {
      container.removeEventListener('mouseover', handleMouseOver);
      container.removeEventListener('mouseout', handleMouseOut);
      container.removeEventListener('click', handleClick, true);
      setHighlight(null);
      setCopiedText(null);
    };
  }, [enabled]);

  return (
    <Box
      ref={containerRef}
      sx={{
        position: 'relative',
        height: '100%',
        cursor: enabled ? 'crosshair' : 'default',
      }}
    >
      {children}

      {/* Highlight overlay */}
      {enabled && highlight && (
        <>
          <Box
            sx={{
              position: 'absolute',
              top: highlight.rect.y,
              left: highlight.rect.x,
              width: highlight.rect.width,
              height: highlight.rect.height,
              bgcolor: HIGHLIGHT_COLOR,
              border: `1.5px solid ${BORDER_COLOR}`,
              borderRadius: '2px',
              pointerEvents: 'none',
              zIndex: 1000,
            }}
          />

          {/* Class tooltip */}
          <Box
            sx={{
              position: 'absolute',
              top: Math.max(0, highlight.rect.y - 28),
              left: highlight.rect.x,
              bgcolor: '#1e293b',
              color: '#e2e8f0',
              px: 1,
              py: 0.25,
              borderRadius: '4px',
              pointerEvents: 'none',
              zIndex: 1001,
              maxWidth: 500,
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
            }}
          >
            <Typography component="span" sx={{fontSize: '0.68rem', fontFamily: 'monospace', fontWeight: 500}}>
              <span style={{color: '#94a3b8'}}>{`<${highlight.tag}>`}</span>
              {highlight.classes.length > 0 ? (
                <span style={{color: '#7dd3fc', marginLeft: 6}}>.{highlight.classes.join(' .')}</span>
              ) : (
                <span style={{color: '#94a3b8', marginLeft: 6, fontStyle: 'italic'}}>no classes</span>
              )}
            </Typography>
          </Box>
        </>
      )}

      {/* Copied feedback */}
      {copiedText && (
        <Box
          sx={{
            position: 'absolute',
            bottom: 12,
            left: '50%',
            transform: 'translateX(-50%)',
            bgcolor: '#1e293b',
            color: '#34d399',
            px: 1.5,
            py: 0.5,
            borderRadius: '6px',
            zIndex: 1002,
            pointerEvents: 'none',
          }}
        >
          <Typography sx={{fontSize: '0.75rem', fontFamily: 'monospace', fontWeight: 600}}>
            Copied: {copiedText}
          </Typography>
        </Box>
      )}
    </Box>
  );
}
