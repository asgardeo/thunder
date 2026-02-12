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

import React, {ReactNode, Children, isValidElement} from 'react';
import {
  Stepper as MuiStepper,
  Step,
  StepLabel,
  StepContent,
  Box,
  Typography,
  StepIconProps,
  styled,
} from '@wso2/oxygen-ui';

interface StepperProps {
  children: ReactNode;
  stepNode?: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';
  as?: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';
}

interface StepData {
  label: string;
  content: ReactNode[];
}

const ColorlibStepIconRoot = styled('div')<{
  ownerState: {completed?: boolean; active?: boolean};
}>(({theme}) => ({
  backgroundColor: theme.palette.background.paper,
  zIndex: 1,
  color: theme.palette.text.primary,
  width: 30,
  height: 30,
  display: 'flex',
  borderRadius: '50%',
  justifyContent: 'center',
  alignItems: 'center',
}));

function ColorlibStepIcon(props: StepIconProps) {
  const {active, completed, className} = props;

  return (
    <ColorlibStepIconRoot ownerState={{completed, active}} className={className}>
      {props.icon}
    </ColorlibStepIconRoot>
  );
}

export default function Stepper({children, stepNode = 'h2', as = 'h2'}: StepperProps) {
  const steps: StepData[] = [];
  let currentStep: StepData | null = null;

  // Process children to group them into steps
  Children.forEach(children, (child) => {
    if (!isValidElement(child)) {
      if (currentStep) {
        currentStep.content.push(child);
      }
      return;
    }

    // Check if this is a heading that should become a step
    // In MDX, headings can be either string types (h1, h2, etc.) or components
    const isHeading =
      child.type === stepNode ||
      (typeof child.type === 'function' && child.type.name === stepNode) ||
      child.props?.mdxType === stepNode;

    if (isHeading) {
      // Save previous step if it exists
      if (currentStep) {
        steps.push(currentStep);
      }
      // Create new step
      currentStep = {
        label:
          typeof child.props.children === 'string'
            ? child.props.children
            : extractTextFromChildren(child.props.children),
        content: [],
      };
    } else if (currentStep) {
      // Add content to current step
      currentStep.content.push(child);
    }
  });

  // Push the last step
  if (currentStep) {
    steps.push(currentStep);
  }

  return (
    <Box sx={{mt: 4}}>
      <MuiStepper orientation="vertical">
        {steps.map((step) => (
          <Step key={`${step.label}`} active>
            <StepLabel slots={{stepIcon: ColorlibStepIcon}}>
              <Typography variant={as}>{step.label}</Typography>
            </StepLabel>
            <StepContent sx={{pt: 4, pl: 4}}>{step.content}</StepContent>
          </Step>
        ))}
      </MuiStepper>
    </Box>
  );
}

// Helper function to extract text from React children
function extractTextFromChildren(children: ReactNode): string {
  if (typeof children === 'string') {
    return children;
  }
  if (Array.isArray(children)) {
    return children.map(extractTextFromChildren).join('');
  }
  if (isValidElement(children) && children.props.children) {
    return extractTextFromChildren(children.props.children);
  }
  return '';
}
