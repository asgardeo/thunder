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

import React from 'react';

// ── Diagram geometry ─────────────────────────────────────────────────────────
// Total height: 270px
//   3 role cards × 76px + 2 gaps × 15px + 6px top + 6px bottom = 270px
//   4 cap items × 60px + 3 gaps × 10px = 270px  (exact match, no padding)
// Role y-centers: 44, 135, 226
// Cap y-centers:  30, 100, 170, 240

const H = 270;
const ROLE_Y = [44, 135, 226] as const;
const CAP_Y  = [30, 100, 170, 240] as const;

const roles = [
  {
    id: 'subject',
    label: 'as Subject',
    connects: 'Tools & APIs',
    desc: 'Initiates, decides, acts',
  },
  {
    id: 'client',
    label: 'as Client',
    connects: 'Users & Systems',
    desc: 'Delegates, acts on behalf of',
  },
  {
    id: 'resource',
    label: 'as Resource',
    connects: 'Other Agents',
    desc: 'Receives calls, enforces boundaries',
  },
];

const caps = [
  {
    id: 'admin',
    title: 'Administration',
    href: '#administration',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/>
        <circle cx="12" cy="12" r="3"/>
      </svg>
    ),
  },
  {
    id: 'authn',
    title: 'Authentication',
    href: '#authentication',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <path d="m21 2-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0 3 3L22 7l-3-3m-3.5 3.5L19 4"/>
      </svg>
    ),
  },
  {
    id: 'authz',
    title: 'Authorization',
    href: '#authorization',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
        <path d="m9 12 2 2 4-4"/>
      </svg>
    ),
  },
  {
    id: 'audit',
    title: 'Audit',
    href: '#audit',
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
        <rect x="8" y="2" width="8" height="4" rx="1" ry="1"/>
        <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/>
        <path d="m9 14 2 2 4-4"/>
      </svg>
    ),
  },
];

export function AgentIdentityModel() {
  return (
    <figure className="aim" aria-label="Agent identity model">
      <div className="aim__scroll">
        <div className="aim__layout" style={{ height: H }}>

          {/* ── Column 1: Agent node ── */}
          <div className="aim__agent-col">
            <div className="aim__agent-node">
              <svg className="aim__agent-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                <rect x="7" y="7" width="10" height="8" rx="2"/>
                <path d="M9 11h.01M15 11h.01"/>
                <path d="M12 7V4"/>
                <path d="M9 15v3M15 15v3"/>
              </svg>
              <span className="aim__agent-name">AI Agent</span>
            </div>
          </div>

          {/* ── Column 2: Left connector SVG ── */}
          <svg className="aim__conn" viewBox={`0 0 60 ${H}`} width="60" height={H} aria-hidden="true" xmlns="http://www.w3.org/2000/svg">
            {/* Horizontal stem from agent center */}
            <line x1="0" y1={H / 2} x2="25" y2={H / 2} className="aim__line" />
            {/* Vertical bar spanning all three role y-centers */}
            <line x1="25" y1={ROLE_Y[0]} x2="25" y2={ROLE_Y[2]} className="aim__line" />
            {/* Three horizontal branches to role cards */}
            {ROLE_Y.map((y) => (
              <line key={y} x1="25" y1={y} x2="60" y2={y} className="aim__line" />
            ))}
            {/* Junction dots */}
            {ROLE_Y.map((y) => (
              <circle key={y} cx="25" cy={y} r="2.5" className="aim__dot" />
            ))}
          </svg>

          {/* ── Column 3: Role cards ── */}
          <div className="aim__roles-col">
            {roles.map((r) => (
              <div key={r.id} className={`aim__role aim__role--${r.id}`}>
                <span className="aim__role-badge">{r.label}</span>
                <span className="aim__role-connects">{r.connects}</span>
                <span className="aim__role-desc">{r.desc}</span>
              </div>
            ))}
          </div>

          {/* ── Column 4: Right connector / bracket SVG ── */}
          <svg className="aim__conn" viewBox={`0 0 60 ${H}`} width="60" height={H} aria-hidden="true" xmlns="http://www.w3.org/2000/svg">
            {/* Horizontal lines from each role to the bracket spine */}
            {ROLE_Y.map((y) => (
              <line key={y} x1="0" y1={y} x2="30" y2={y} className="aim__line" />
            ))}
            {/* Bracket spine spanning all capability y-centers */}
            <line x1="30" y1={CAP_Y[0]} x2="30" y2={CAP_Y[3]} className="aim__line" />
            {/* Four horizontal branches to capability cards */}
            {CAP_Y.map((y) => (
              <line key={y} x1="30" y1={y} x2="60" y2={y} className="aim__line" />
            ))}
            {/* Junction dots */}
            {CAP_Y.map((y) => (
              <circle key={y} cx="30" cy={y} r="2.5" className="aim__dot" />
            ))}
          </svg>

          {/* ── Column 5: Capability cards ── */}
          <div className="aim__caps-col">
            {caps.map((c) => (
              <a key={c.id} href={c.href} className={`aim__cap aim__cap--${c.id}`}>
                <span className="aim__cap-icon">{c.icon}</span>
                <span className="aim__cap-title">{c.title}</span>
                <span className="aim__cap-arrow" aria-hidden="true">→</span>
              </a>
            ))}
          </div>

        </div>
      </div>
    </figure>
  );
}
