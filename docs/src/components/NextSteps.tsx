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
import React, {ReactNode} from 'react';

/* ─── NextStepsCard ───────────────────────────────────────────────────────── */

interface NextStepsCardProps {
  title: string;
  description?: string;
  href: string;
  step?: number;
  cta?: string;
}

export function NextStepsCard({title, description, href, step, cta}: NextStepsCardProps) {
  const isExternal = href.startsWith('http');
  const ctaLabel = cta ?? `Go to ${title}`;

  return (
    <Link
      to={href}
      {...(isExternal ? {target: '_blank', rel: 'noopener noreferrer'} : {})}
      style={{textDecoration: 'none', display: 'block'}}
    >
      <div className="next-steps-card">
        <span className="next-steps-card__title">
          {step != null ? `Step ${step} → ${title}` : title}
        </span>
        {description && <span className="next-steps-card__desc">{description}</span>}
        <span className="next-steps-card__cta">{ctaLabel} →</span>
      </div>
    </Link>
  );
}

/* ─── NextStepsGroup ──────────────────────────────────────────────────────── */

interface NextStepsGroupProps {
  label: string;
  children: ReactNode;
}

export function NextStepsGroup({label, children}: NextStepsGroupProps) {
  return (
    <div className="next-steps-group">
      <span className="next-steps-group__label">{label}</span>
      <div className="next-steps-group__cards">{children}</div>
    </div>
  );
}

/* ─── NextSteps (wrapper) ─────────────────────────────────────────────────── */

interface NextStepsProps {
  children: ReactNode;
}

export function NextSteps({children}: NextStepsProps) {
  return <div className="next-steps">{children}</div>;
}
