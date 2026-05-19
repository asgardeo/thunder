import React from 'react';
import Link from '@docusaurus/Link';

type Feature = { text: string; available: boolean };

type Station = {
  href: string;
  accentColor: string;
  iconBackground: string;
  stepLabel: string;
  stepTag: string;
  title: string;
  chooseIf: string;
  setupTime: string;
  cta: string;
  icon: React.ReactNode;
  features: Feature[];
  featured?: boolean;
  animClass: string;
};

function DockerLogo() {
  return (
    <img
      src="/img/docker-logo.svg"
      alt="Docker"
      width="22"
      height="22"
      style={{ objectFit: 'contain', display: 'block' }}
    />
  );
}

function KubernetesLogo() {
  return (
    <img
      src="/img/kubernetes-logo.svg"
      alt="Kubernetes"
      width="22"
      height="22"
      style={{ objectFit: 'contain', display: 'block' }}
    />
  );
}

function OpenChoreoLogo() {
  return (
    <img
      src="/img/openchoreo-logo.svg"
      alt="OpenChoreo"
      width="22"
      height="22"
      style={{ objectFit: 'contain', display: 'block' }}
    />
  );
}

function ClockIcon() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10"/>
      <polyline points="12 6 12 12 16 14"/>
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="20 6 9 17 4 12"/>
    </svg>
  );
}

function DashIcon() {
  return (
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
      <line x1="5" y1="12" x2="19" y2="12"/>
    </svg>
  );
}

const stations: Station[] = [
  {
    href: './docker',
    accentColor: '#0ea5e9',
    iconBackground: 'rgba(14,165,233,0.12)',
    stepLabel: '01',
    stepTag: 'Local evaluation',
    title: 'Docker',
    chooseIf: 'You want to try ThunderID locally, test features, or run demos without setting up a cluster.',
    setupTime: '<5 min',
    cta: 'Start locally →',
    icon: <DockerLogo />,
    features: [
      { text: 'Quick setup with pre-built images', available: true },
      { text: 'PostgreSQL integration', available: true },
      { text: 'Custom configuration mounting', available: true },
      { text: 'Not for production use', available: false },
    ],
    animClass: 'dp-row-1',
  },
  {
    href: './kubernetes',
    accentColor: '#326CE5',
    iconBackground: 'rgba(50,108,229,0.12)',
    stepLabel: '02',
    stepTag: 'Self-hosted production',
    title: 'Kubernetes',
    chooseIf: 'You want production control while running ThunderID on infrastructure your team manages.',
    setupTime: '<10 min',
    cta: 'Deploy on Kubernetes →',
    icon: <KubernetesLogo />,
    features: [
      { text: 'Helm chart deployment', available: true },
      { text: 'Multi-replica support', available: true },
      { text: 'Ingress configuration', available: true },
      { text: 'Database flexibility (PostgreSQL/SQLite)', available: true },
      { text: 'Rolling updates and rollbacks', available: true },
    ],
    animClass: 'dp-row-2',
  },
  {
    href: './openchoreo',
    accentColor: '#8b5cf6',
    iconBackground: 'rgba(139,92,246,0.12)',
    stepLabel: '03',
    stepTag: 'Managed platform',
    title: 'OpenChoreo',
    chooseIf: 'You want a platform-managed deployment model with environment separation and promotion workflows.',
    setupTime: '<15 min',
    cta: 'Deploy on OpenChoreo →',
    icon: <OpenChoreoLogo />,
    features: [
      { text: 'Cell-based deployment model', available: true },
      { text: 'Integrated platform services', available: true },
      { text: 'Advanced networking', available: true },
      { text: 'Service mesh integration', available: true },
    ],
    animClass: 'dp-row-3',
  },
]; 

export default function DeploymentCards() {
  return (
    <div className="dp-wrap">

      {/* Page header */}
      <div className="dp-header">
        <div className="dp-header-kicker">Deployment</div>
        <h1 className="dp-header-title">Where are you deploying?</h1>
        <p className="dp-header-sub">
          Pick the environment that matches your setup. Each option has a dedicated installation guide.
        </p>
      </div>

      {/* Route strip */}
      <div className="dp-route">
        <div className="dp-route-stops">
          {stations.map((s) => (
            <div
              key={s.href}
              className="dp-route-stop"
              style={{ ['--dp-accent' as string]: s.accentColor, ['--dp-icon-bg' as string]: s.iconBackground }}
            >
              <div className="dp-route-node">{s.icon}</div>
              <span className="dp-route-stop-label">{s.title}</span>
            </div>
          ))}
        </div>
      </div>

      {/* 3-column cards */}
      <div className="dp-list">
        {stations.map((s) => (
          <Link
            key={s.href}
            to={s.href}
            className={`dp-row-card ${s.animClass}${s.featured ? ' dp-featured' : ''}`}
            style={{ ['--dp-accent' as string]: s.accentColor, ['--dp-icon-bg' as string]: s.iconBackground }}
          >
            <div className="dp-rc-step">Path {s.stepLabel} — {s.stepTag}</div>
            <div className="dp-rc-title">{s.title}</div>

            <div className="dp-setup-badge">
              <span className="dp-setup-badge-icon"><ClockIcon /></span>
              <div>
                <div className="dp-setup-label">Setup time</div>
                <div className="dp-setup-value">{s.setupTime}</div>
              </div>
            </div>

            <ul className="dp-feature-list">
              {s.features.map((f) => (
                <li key={f.text} className={`dp-feature-item ${f.available ? 'dp-feat-check' : 'dp-feat-miss'}`}>
                  <span className="dp-feat-icon">
                    {f.available ? <CheckIcon /> : <DashIcon />}
                  </span>
                  <span>{f.text}</span>
                </li>
              ))}
            </ul>

            <div className="dp-card-choose">
              <div className="dp-card-choose-label">Choose this if…</div>
              <p className="dp-card-choose-text">{s.chooseIf}</p>
            </div>

            <span className="dp-rc-cta">{s.cta}</span>
          </Link>
        ))}
      </div>

      {/* Production recommendation note */}
      <p className="dp-prod-note">
        For production deployments, use <strong>Kubernetes</strong> or <strong>OpenChoreo</strong> with PostgreSQL. Use Docker for local evaluation, demos, and development testing.
      </p>

    </div>
  );
}
