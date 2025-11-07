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

import kebabCase from 'lodash-es/kebabCase';
import {type HTMLAttributes, type ReactElement} from 'react';
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Box,
  Drawer,
  IconButton,
  Stack,
  Tooltip,
  Typography,
} from '@wso2/oxygen-ui';
import {
  BoxesIcon,
  BoxIcon,
  ChevronDownIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  CogIcon,
  LayoutTemplate,
} from '@wso2/oxygen-ui-icons-react';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import ResourcePanelStatic from './ResourcePanelStatic';
import {type Element, ElementTypes} from '../../models/elements';
import type {Resource, Resources} from '../../models/resources';
import type {Step} from '../../models/steps';
import type {Template} from '../../models/templates';
import type {Widget} from '../../models/widget';
import ResourcePanelDraggable from './ResourcePanelDraggable';
import './ResourcePanel.scss';

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
   */
  open?: boolean;
  /**
   * Callback to be triggered when a resource add button is clicked.
   * @param resource - Added resource.
   */
  onAdd: (resource: Resource) => void;
  /**
   * Flag to disable the panel.
   */
  disabled?: boolean;
}

const PANEL_WIDTH = 350;

/**
 * Flow builder resource panel that contains draggable components.
 *
 * @param props - Props injected to the component.
 * @returns The ResourcePanel component.
 */
function ResourcePanel({
  children,
  open,
  resources,
  onAdd,
  disabled = false,
  ...rest
}: ResourcePanelPropsInterface): ReactElement {
  const {setIsResourcePanelOpen} = useFlowBuilderCore();

  const {
    elements: unfilteredElements,
    widgets: unfilteredWidgets,
    steps: unfilteredSteps,
    templates: unfilteredTemplates,
  } = resources;

  const handleTogglePanel = (): void => {
    setIsResourcePanelOpen((prev: boolean) => !prev);
  };

  const elements: Element[] = unfilteredElements?.filter(
    (element: Element) => element.display?.showOnResourcePanel !== false,
  );
  const widgets: Widget[] = unfilteredWidgets?.filter(
    (widget: Widget) => widget.display?.showOnResourcePanel !== false,
  );

  const steps: Step[] = unfilteredSteps?.filter((step: Step) => step.display?.showOnResourcePanel !== false);
  const templates: Template[] = unfilteredTemplates?.filter(
    (template: Template) => template.display?.showOnResourcePanel !== false,
  );

  return (
    <Box width="100%" height="100%" display="flex" position="relative" {...rest}>
      {/* Floating expand button shown when panel is collapsed */}
      {!open && (
        <Tooltip title="Show Resources" placement="right">
          <IconButton
            onClick={handleTogglePanel}
            sx={{
              position: 'absolute',
              top: 16,
              left: 16,
              zIndex: 10,
              borderRadius: 1,
              '&:hover': {
                backgroundColor: 'action.hover',
              },
            }}
            size="small"
          >
            <ChevronRightIcon size={16} />
          </IconButton>
        </Tooltip>
      )}

      {/* Resource Panel using MUI Drawer for animation */}
      <Drawer
        variant="persistent"
        anchor="left"
        open={open}
        sx={{
          width: PANEL_WIDTH,
          height: '100%',
          flexShrink: 0,
          mr: 1,
          transition: (theme) =>
            theme.transitions.create('width', {
              easing: open ? theme.transitions.easing.easeOut : theme.transitions.easing.sharp,
              duration: open ? theme.transitions.duration.enteringScreen : theme.transitions.duration.leavingScreen,
            }),
          ...(!open && {
            width: 0,
            mr: 0,
          }),
          '& .MuiDrawer-paper': {
            width: PANEL_WIDTH,
            position: 'relative',
            border: 'none',
            overflow: 'scroll',
            p: 2,
            gap: 1,
          },
        }}
      >
        {/* Header with title and collapse button */}
        <Box
          display="flex"
          alignItems="center"
          justifyContent="space-between"
          sx={{
            pb: 1,
            borderBottom: '1px solid',
            borderColor: 'divider',
            mb: 1,
            flexShrink: 0,
          }}
        >
          <Typography variant="subtitle1" fontWeight={600}>
            Resources
          </Typography>
          <Tooltip title="Hide Resources" placement="right">
            <IconButton onClick={handleTogglePanel} size="small">
              <ChevronLeftIcon size={16} />
            </IconButton>
          </Tooltip>
        </Box>

        {/* Starter Templates */}
        <Accordion
          square
          disableGutters
          defaultExpanded
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
            aria-controls="panel1-content"
            id="panel1-header"
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
              <LayoutTemplate size={16} />
            </Box>
            <Typography variant="subtitle2" fontWeight={600}>
              Starter Templates
            </Typography>
          </AccordionSummary>
          <AccordionDetails
            sx={{
              pt: 0,
              pb: 2,
              px: 2,
            }}
          >
            <Typography variant="body2" color="text.secondary" gutterBottom sx={{mb: 1.5}}>
              Choose one of these templates to start building registration experience
            </Typography>
            <Stack direction="column" spacing={1}>
              {templates?.map((template: Template, index: number) => (
                <ResourcePanelStatic
                  id={`${template.resourceType}-${template.type}-${index}`}
                  key={template.type}
                  resource={template}
                  onAdd={onAdd}
                  disabled={disabled}
                />
              ))}
            </Stack>
          </AccordionDetails>
        </Accordion>

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
              Widgets
            </Typography>
          </AccordionSummary>
          <AccordionDetails
            sx={{
              pt: 0,
              pb: 2,
              px: 2,
            }}
          >
            <Typography variant="body2" color="text.secondary" gutterBottom sx={{mb: 1.5}}>
              Use these widgets to build up the flow using pre-created flow blocks
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
              Steps
            </Typography>
          </AccordionSummary>
          <AccordionDetails
            sx={{
              pt: 0,
              pb: 2,
              px: 2,
            }}
          >
            <Typography variant="body2" color="text.secondary" gutterBottom sx={{mb: 1.5}}>
              Use these as steps in your flow
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
              Components
            </Typography>
          </AccordionSummary>
          <AccordionDetails
            sx={{
              pt: 0,
              pb: 2,
              px: 2,
            }}
          >
            <Typography variant="body2" color="text.secondary" gutterBottom sx={{mb: 1.5}}>
              Use these components to build up your views
            </Typography>
            <Stack direction="column" spacing={1}>
              {elements?.map((element: Element, index: number) => (
                <ResourcePanelDraggable
                  id={`${element.resourceType}-${element.type}-${index}`}
                  key={
                    element.type === ElementTypes.Input
                      ? `${element.type}_${String(element.variant)}_${index}`
                      : `${element.type}_${index}`
                  }
                  resource={element}
                  onAdd={onAdd}
                  disabled={false}
                />
              ))}
            </Stack>
          </AccordionDetails>
        </Accordion>
      </Drawer>

      {/* Main Content Area (Canvas) */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          height: '100%',
          position: 'relative',
          transition: (theme) =>
            theme.transitions.create(['margin', 'width'], {
              easing: open ? theme.transitions.easing.easeOut : theme.transitions.easing.sharp,
              duration: open ? theme.transitions.duration.enteringScreen : theme.transitions.duration.leavingScreen,
            }),
        }}
      >
        {children}
      </Box>
    </Box>
  );
}

export default ResourcePanel;
