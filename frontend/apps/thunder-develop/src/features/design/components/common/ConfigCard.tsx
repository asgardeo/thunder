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

import {Accordion, AccordionDetails, AccordionSummary, Typography} from '@wso2/oxygen-ui';
import {ChevronDown} from '@wso2/oxygen-ui-icons-react';
import type {JSX, ReactNode} from 'react';

export interface ConfigCardProps {
  title: string;
  children: ReactNode;
  defaultOpen?: boolean;
}

/**
 * ConfigCard - A collapsible card component for builder mode sections.
 * Used to group related configuration options within theme and layout editors.
 */
export default function ConfigCard({title, children, defaultOpen = true}: ConfigCardProps): JSX.Element {
  return (
    <Accordion
      defaultExpanded={defaultOpen}
      disableGutters
      square
      sx={{
        backgroundColor: 'transparent',
        '&:before': {
          display: 'none',
        },
        overflow: 'hidden',
        flexShrink: 0,
      }}
    >
      <AccordionSummary expandIcon={<ChevronDown size={16} />}>
        <Typography variant="body2" sx={{fontWeight: 600, fontSize: '0.9375rem'}}>
          {title}
        </Typography>
      </AccordionSummary>
      <AccordionDetails>{children}</AccordionDetails>
    </Accordion>
  );
}
