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

import Link from '@docusaurus/Link';
import React from 'react';

interface JourneyStep {
  label: string;
  href: string;
  docIds: string[];
}

export const GETTING_STARTED_STEPS: JourneyStep[] = [
  {
    label: 'Run ThunderID',
    href: '/docs/next/guides/getting-started/get-thunderid',
    docIds: ['guides/getting-started/get-thunderid'],
  },
  {
    label: 'Register an app',
    href: '/docs/next/guides/quick-start/register-an-application',
    docIds: ['guides/quick-start/register-an-application'],
  },
  {
    label: 'Build a flow',
    href: '/docs/next/guides/guides/flows/build-a-flow',
    docIds: ['guides/guides/flows/build-a-flow'],
  },
  {
    label: 'Connect your app',
    href: '/docs/next/guides/quick-start/connect-your-application',
    docIds: [
      'guides/quick-start/connect-your-application/index',
      'guides/quick-start/connect-your-application/react',
      'guides/quick-start/connect-your-application/nextjs',
    ],
  },
];

export function getGettingStartedStepIndex(docId?: string): number | null {
  if (!docId) {
    return null;
  }

  const stepIndex = GETTING_STARTED_STEPS.findIndex((step) => step.docIds.includes(docId));

  return stepIndex >= 0 ? stepIndex + 1 : null;
}

interface GettingStartedJourneyProps {
  current: number; // 1-based index
}

export default function GettingStartedJourney({current}: GettingStartedJourneyProps) {
  return (
    <div className="gsj">
      {GETTING_STARTED_STEPS.map((step, i) => {
        const index = i + 1;
        const isDone = index < current;
        const isActive = index === current;
        const isNext = index === current + 1;

        const circleClass = [
          'gsj__circle',
          isDone ? 'gsj__circle--done' : '',
          isActive ? 'gsj__circle--active' : '',
        ]
          .filter(Boolean)
          .join(' ');

        const labelClass = [
          'gsj__label',
          isActive ? 'gsj__label--active' : '',
          isDone ? 'gsj__label--done' : '',
        ]
          .filter(Boolean)
          .join(' ');

        const content = (
          <div className="gsj__step">
            <div className={circleClass}>
              {isDone ? (
                <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                  <path d="M2 6l3 3 5-5" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              ) : (
                <span>{index}</span>
              )}
            </div>
            <span className={labelClass}>{step.label}</span>
          </div>
        );

        return (
          <React.Fragment key={step.href}>
            {isNext ? (
              <Link to={step.href} className="gsj__step-link" aria-label={`Next: ${step.label}`}>
                {content}
              </Link>
            ) : isActive ? (
              <div className="gsj__step-link gsj__step-link--active" aria-current="step">
                {content}
              </div>
            ) : (
              <Link to={step.href} className="gsj__step-link gsj__step-link--muted">
                {content}
              </Link>
            )}
            {i < GETTING_STARTED_STEPS.length - 1 && (
              <div className={`gsj__connector ${isDone ? 'gsj__connector--done' : ''}`} aria-hidden />
            )}
          </React.Fragment>
        );
      })}
    </div>
  );
}
