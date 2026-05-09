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

import {BuilderLayout, BuilderPanelHeader} from '@thunderid/components';
import {Accordion, AccordionDetails, AccordionSummary, Box, Stack, Typography} from '@wso2/oxygen-ui';
import {BoxesIcon, BoxIcon, ChevronDownIcon, CogIcon, ZapIcon} from '@wso2/oxygen-ui-icons-react';
import kebabCase from 'lodash-es/kebabCase';
import {memo, useCallback, useMemo, type HTMLAttributes, type ReactElement, type ReactNode} from 'react';
import {useTranslation} from 'react-i18next';
import ResourcePanelDraggable from './ResourcePanelDraggable';
import useUIPanelState from '../../hooks/useUIPanelState';
import type {Element} from '../../models/elements';
import type {Resource, Resources} from '../../models/resources';
import type {Step} from '../../models/steps';
import type {Widget} from '../../models/widget';

/**
 * Props interface of {@link ResourcePanel}
 */
export interface ResourcePanelPropsInterface extends HTMLAttributes<HTMLDivElement> {
  /**
   * Flow resources.
   */
  resources: Resources;
  /**
   * Whether the panel is open.
   * @defaultValue undefined
   */
  open?: boolean;
  /**
   * Callback to be triggered when a resource add button is clicked.
   * @param resource - Added resource.
   */
  onAdd: (resource: Resource) => void;
  /**
   * Flag to disable the panel.
   * @defaultValue false
   */
  disabled?: boolean;
  /**
   * Flow title to display.
   */
  flowTitle?: string;
  /**
   * Flow handle (URL-friendly identifier).
   */
  flowHandle?: string;
  /**
   * Callback to be triggered when flow title changes.
   */
  onFlowTitleChange?: (newTitle: string) => void;
  /**
   * Optional right-hand side panel content rendered via BuilderLayout.
   */
  rightPanel?: ReactNode;
}

/**
 * Flow builder resource panel that contains draggable components.
 *
 * @param props - Props injected to the component.
 * @returns The ResourcePanel component.
 */
function ResourcePanel({
  children,
  open = undefined,
  resources,
  onAdd,
  disabled = false,
  flowTitle = '',
  flowHandle = '',
  onFlowTitleChange = undefined,
  rightPanel = undefined,
  ...rest
}: ResourcePanelPropsInterface): ReactElement {
  const {t} = useTranslation();
  const {setIsResourcePanelOpen} = useUIPanelState();

  const handleTogglePanel = useCallback((): void => {
    setIsResourcePanelOpen((prev: boolean) => !prev);
  }, [setIsResourcePanelOpen]);

  const {
    elements: unfilteredElements,
    widgets: unfilteredWidgets,
    steps: unfilteredSteps,
    executors: unfilteredExecutors,
  } = resources;

  const elements: Element[] = useMemo(
    () => unfilteredElements?.filter((element: Element) => element.display?.showOnResourcePanel !== false),
    [unfilteredElements],
  );
  const widgets: Widget[] = useMemo(
    () => unfilteredWidgets?.filter((widget: Widget) => widget.display?.showOnResourcePanel !== false),
    [unfilteredWidgets],
  );
  const steps: Step[] = useMemo(
    () => unfilteredSteps?.filter((step: Step) => step.display?.showOnResourcePanel !== false),
    [unfilteredSteps],
  );
  const executors: Step[] = useMemo(
    () => unfilteredExecutors?.filter((executor: Step) => executor.display?.showOnResourcePanel !== false),
    [unfilteredExecutors],
  );

  const panelContent = (
    <>
      <BuilderPanelHeader
        title={flowTitle}
        handle={flowHandle}
        onPanelToggle={handleTogglePanel}
        onTitleChange={onFlowTitleChange}
        hidePanelTooltip={t('flows:core.resourcePanel.hideResources')}
        editTitleTooltip={t('flows:core.headerPanel.editTitle')}
        saveTitleTooltip={t('flows:core.headerPanel.saveTitle')}
        cancelEditTooltip={t('flows:core.headerPanel.cancelEdit')}
      />

      {/* Widgets */}
      <Accordion
        square
        disableGutters
        sx={{
          backgroundColor: 'transparent',
          '&:before': {
            display: 'none',
          },
          overflow: 'hidden',
          flexShrink: 0,
        }}
      >
        <AccordionSummary
          expandIcon={<ChevronDownIcon size={14} />}
          id="panel2-header"
          sx={{
            minHeight: 48,
            '&.Mui-expanded': {
              minHeight: 48,
            },
            '& .MuiAccordionSummary-content': {
              margin: '12px 0',
              gap: 1,
            },
          }}
          slotProps={{
            content: {
              sx: {alignItems: 'center'},
            },
          }}
        >
          <Box component="span" display="inline-flex" alignItems="center">
            <CogIcon size={16} />
          </Box>
          <Typography variant="subtitle2" fontWeight={600}>
            {t('flows:core.resourcePanel.widgets.title')}
          </Typography>
        </AccordionSummary>
        <AccordionDetails sx={{pt: 0, pb: 2, px: 2}}>
          <Typography variant="body2" color="text.secondary" gutterBottom sx={{mb: 1.5}}>
            {t('flows:core.resourcePanel.widgets.description')}
          </Typography>
          <Stack direction="column" spacing={1}>
            {widgets?.map((widget: Widget, index: number) => (
              <ResourcePanelDraggable
                id={`${widget.resourceType}-${widget.type}-${index}`}
                key={widget.type}
                resource={widget}
                onAdd={onAdd}
                disabled={disabled}
              />
            ))}
          </Stack>
        </AccordionDetails>
      </Accordion>

      {/* Steps */}
      <Accordion
        square
        disableGutters
        sx={{
          backgroundColor: 'transparent',
          '&:before': {
            display: 'none',
          },
          overflow: 'hidden',
          flexShrink: 0,
        }}
      >
        <AccordionSummary
          expandIcon={<ChevronDownIcon size={14} />}
          aria-controls="panel3-content"
          id="panel3-header"
          sx={{
            minHeight: 48,
            '&.Mui-expanded': {
              minHeight: 48,
            },
            '& .MuiAccordionSummary-content': {
              margin: '12px 0',
              gap: 1,
            },
          }}
          slotProps={{
            content: {
              sx: {alignItems: 'center'},
            },
          }}
        >
          <Box component="span" display="inline-flex" alignItems="center">
            <BoxIcon size={16} />
          </Box>
          <Typography variant="subtitle2" fontWeight={600}>
            {t('flows:core.resourcePanel.steps.title')}
          </Typography>
        </AccordionSummary>
        <AccordionDetails sx={{pt: 0, pb: 2, px: 2}}>
          <Typography variant="body2" color="text.secondary" gutterBottom sx={{mb: 1.5}}>
            {t('flows:core.resourcePanel.steps.description')}
          </Typography>
          <Stack direction="column" spacing={1}>
            {steps?.map((step: Step, index: number) => (
              <ResourcePanelDraggable
                id={`${step.resourceType}-${step.type}-${index}`}
                key={`${step.type}-${kebabCase(step.display.label)}`}
                resource={step}
                onAdd={onAdd}
                disabled={disabled}
              />
            ))}
          </Stack>
        </AccordionDetails>
      </Accordion>

      {/* Components */}
      <Accordion
        square
        disableGutters
        sx={{
          backgroundColor: 'transparent',
          '&:before': {
            display: 'none',
          },
          overflow: 'hidden',
          flexShrink: 0,
        }}
      >
        <AccordionSummary
          expandIcon={<ChevronDownIcon size={14} />}
          aria-controls="panel4-content"
          id="panel4-header"
          sx={{
            minHeight: 48,
            '&.Mui-expanded': {
              minHeight: 48,
            },
            '& .MuiAccordionSummary-content': {
              margin: '12px 0',
              gap: 1,
            },
          }}
          slotProps={{
            content: {
              sx: {alignItems: 'center'},
            },
          }}
        >
          <Box component="span" display="inline-flex" alignItems="center">
            <BoxesIcon size={16} />
          </Box>
          <Typography variant="subtitle2" fontWeight={600}>
            {t('flows:core.resourcePanel.components.title')}
          </Typography>
        </AccordionSummary>
        <AccordionDetails sx={{pt: 0, pb: 2, px: 2}}>
          <Typography variant="body2" color="text.secondary" gutterBottom sx={{mb: 1.5}}>
            {t('flows:core.resourcePanel.components.description')}
          </Typography>
          <Stack direction="column" spacing={1}>
            {elements?.map((element: Element, index: number) => (
              <ResourcePanelDraggable
                id={`${element.resourceType}-${element.type}-${index}`}
                key={`${element.resourceType}-${element.category}-${element.type}`}
                resource={element}
                onAdd={onAdd}
                disabled={disabled}
              />
            ))}
          </Stack>
        </AccordionDetails>
      </Accordion>

      {/* Executors */}
      <Accordion
        square
        disableGutters
        sx={{
          backgroundColor: 'transparent',
          '&:before': {
            display: 'none',
          },
          overflow: 'hidden',
          flexShrink: 0,
        }}
      >
        <AccordionSummary
          expandIcon={<ChevronDownIcon size={14} />}
          aria-controls="panel-executors-content"
          id="panel-executors-header"
          sx={{
            minHeight: 48,
            '&.Mui-expanded': {
              minHeight: 48,
            },
            '& .MuiAccordionSummary-content': {
              margin: '12px 0',
              gap: 1,
            },
          }}
          slotProps={{
            content: {
              sx: {alignItems: 'center'},
            },
          }}
        >
          <Box component="span" display="inline-flex" alignItems="center">
            <ZapIcon size={16} />
          </Box>
          <Typography variant="subtitle2" fontWeight={600}>
            {t('flows:core.resourcePanel.executors.title')}
          </Typography>
        </AccordionSummary>
        <AccordionDetails sx={{pt: 0, pb: 2, px: 2}}>
          <Typography variant="body2" color="text.secondary" gutterBottom sx={{mb: 1.5}}>
            {t('flows:core.resourcePanel.executors.description')}
          </Typography>
          <Stack direction="column" spacing={1}>
            {executors?.map((executor: Step, index: number) => (
              <ResourcePanelDraggable
                id={`${executor.resourceType}-${executor.type}-${index}`}
                key={`${executor.type}-${kebabCase(executor.display.label)}`}
                resource={executor}
                onAdd={onAdd}
                disabled={disabled}
              />
            ))}
          </Stack>
        </AccordionDetails>
      </Accordion>
    </>
  );

  return (
    <BuilderLayout
      open={open}
      onPanelToggle={handleTogglePanel}
      expandTooltip={t('flows:core.resourcePanel.showResources')}
      panelContent={panelContent}
      panelPaperSx={{
        overflow: 'hidden auto',
        display: 'flex',
        flexDirection: 'column',
        borderRight: '1px solid',
        borderColor: 'divider',
      }}
      rightPanel={rightPanel}
      {...rest}
    >
      {children}
    </BuilderLayout>
  );
}

export default memo(ResourcePanel);
