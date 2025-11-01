# Title

**Authors**:  
_@your-github-handle_

**Reviewers**:  
_@github-handle (assigned reviewers, include security reviewer if applicable)_

**Feature Issue**:  
_#123 (REQUIRED - Link to the approved Feature issue)_

**Design Discussion**:  
_#discussions/456 (Link to the Design Proposals discussion)_

**Created Date**:  
_YYYY-MM-DD_

**Status**:  
_Accepted / Rejected / Implemented_

**Related Proposals**:  
_Links to related or superseded proposals (e.g., Supersedes: TAP-0010)_

---

## Summary

_A high-level overview of the proposal (2-3 sentences). Focus on the what and why. 

Example: "This proposal introduces OAuth 2.0 Device Authorization Grant flow to enable authentication on input-constrained devices like smart TVs and IoT devices."_

---

## Motivation

_Explain the IAM problem this proposal is solving. Consider:_
- _What gap in authentication, authorization, or identity management does this address?_
- _Why is this feature necessary for Thunder users?_
- _What value does it add to administrators, developers, or end users?_
- _Are there compliance, security, or usability drivers?_

---

## Goals

_What this proposal aims to achieve._

**Example:**
- Enable authentication on devices without browser support
- Provide a secure alternative to username/password entry on TVs
- Support OAuth 2.0 Device Flow specification (RFC 8628)

---

## Non-Goals

_What this proposal does **not** aim to solve (to set scope clearly)._

**Example:**
- Support for OAuth 1.0 (legacy protocol)
- Custom device registration flows outside the standard
- Device-specific biometric authentication

---

## User Personas & Use Cases

_Describe who will use this feature and how._

**Personas:**
- End User: Person authenticating from a device
- Administrator: Configures device flow settings
- Developer: Integrates device flow into applications

**Use Cases:**
- Smart TV authentication
- CLI tool authentication
- IoT device provisioning

---

## Security Considerations

_Critical section for IAM proposals._

### Threat Model
_What are the security risks? (e.g., token theft, unauthorized device access, phishing)_

### Mitigation Strategies
_How are threats addressed? (e.g., short-lived codes, rate limiting, user confirmation)_

### Compliance Impact
_Does this affect GDPR, SOC2, ISO 27001, or other compliance requirements?_

---

## Impact

_Describe which areas of Thunder will be affected (eg: Authentication, User Management). Include any considerations around backward compatibility, migration, or operational changes._

---

## Design

_Elaborate on the technical design of the feature. This may include component-level architecture, authentication/authorization flows, state transitions, API interface definitions, database schema changes, configuration options, and error handling. Add sequence diagrams, architecture diagrams, or flowcharts as needed to communicate the design clearly._

---

## Appendix (optional)

_Any extra context, links to discussions, POC code, performance benchmarks, etc._
