# Thunder Design Proposals

This directory contains design proposals for Thunder. Proposals are used to document significant changes to the platform before implementation.

## Proposal Process

### Entry Points

You can start the proposal process in two ways:

**Path A: Start with Ideas Discussion** (for unclear or complex problems)
- Best when the problem needs refinement
- Community helps clarify requirements
- Evolves into a Feature issue

**Path B: Start with Feature Issue** (for clear, well-defined problems)
- Skip Ideas discussion if you already know:
  - What problem you're solving
  - Why it should be solved now
  - Who you're solving it for

### Step 1: Ideas Discussion (Optional)

If your idea needs refinement or community validation:

1. **Create an Ideas Discussion** in the "Ideas" category
2. **Describe the problem** you want to solve
3. **Engage with the community** to refine the idea

**Possible Outcomes:**
- ❌ **Invalid** → Close discussion, document why
- 🐛 **Bug** → Create bug issue (standard bug workflow)
- 🔧 **Improvement** → Create improvement issue
- ⭐ **Feature** → Move to Step 2

### Step 2: Create Feature Issue

Create a GitHub issue using the Feature template.

**Required Information:**
- **What is the problem we are trying to solve?**
- **Why should it be solved now?** (urgency, business value)
- **Who are we solving the problem for?** (end users, admins, developers)

**Issue receives number** (e.g., #123) which becomes the proposal identifier.

**Note:** If you started with an Ideas discussion, link it in the feature issue and close the discussion pointing to the new issue.

### Step 3: Assignment & Claiming

Before starting design work, someone must claim the feature:

1. **Express interest** by commenting on the feature issue

2. **Maintainer evaluates** and either:
   - ✅ **Assigns** the issue to you
   - ❓ **Asks for more information**
   - 👥 **Assigns to core team** if more appropriate

**Why this step?**
- Prevents duplicate design work
- Ensures contributor has capacity and skills
- Allows maintainers to provide guidance upfront
- Creates clear ownership

### Step 4: Design Discussion

Once assigned, create a discussion in the "Design Proposals" category.

**Requirements:**
- **MUST link** to the Feature Issue (#123)
- **MUST be assigned** to someone before creating discussion

**Discussion should cover:**
- High-level architecture and design
- Authentication/authorization flows
- Security considerations and threat model
- Performance implications
- Backward compatibility
- Alternative approaches
- Open questions

**Community collaborates to:**
- Refine the technical approach
- Identify potential issues
- Reach consensus on design direction

**Possible Outcomes:**
- ❌ **No Consensus** → Close feature issue, document reasoning
- ⏸️ **Needs More Work** → Continue discussion
- ✅ **Consensus Reached** → Move to Step 5

### Step 5: Write Design Proposal

Once consensus is reached in the discussion:

1. **Copy the template:** `xxxx-design-proposal-template.md`
2. **Rename to:** `<issue-number>-<descriptive-title>.md`
   - Issue number: Zero-padded to 4 digits (e.g., 0123)
   - Title: kebab-case (e.g., oauth2-device-flow)
   - Example: `0123-oauth2-device-flow.md`

3. **Create assets directory** (if needed):
   - Directory: `<issue-number>/` under the `assets` directory (e.g., `assets/0042/`)
   - Include: Architecture diagrams, sequence diagrams, flow charts
   - Reference: Use relative paths (e.g., `![Flow](assets/0042/auth-flow.png)`)

4. **Fill out all sections** of the template:

5. **Set Status:** `Accepted` (default status when submitting)

6. **Reference both:**
   - Feature Issue (#123)
   - Architecture Discussion (#456)

### Step 6: Submit Pull Request

1. **Create PR** with your design proposal document
3. **Link to:**
   - Feature Issue #123
   - Design Proposal Discussion
4. **Request reviews from:**
   - Relevant maintainers
   - Security team (for security-sensitive features)
   - Technical leads

### Step 7: Proposal Review

Reviewers evaluate the proposal for:

- ✅ **Technical Design:** Is the architecture sound and scalable?
- ✅ **Security Model:** Are threats identified and properly mitigated?
- ✅ **Feasibility:** Can this actually be implemented?
- ✅ **Performance:** Will this scale with expected load?
- ✅ **Compatibility:** Impact on existing features and users?
- ✅ **Standards Compliance:** Does it follow OAuth 2.0, OIDC, SAML standards?

**Critical:** Detailed security and technical scrutiny happens here. Issues not visible during high-level discussion may be discovered.

**Possible Outcomes:**

#### ✅ Approved
- Design is sound and ready for implementation
- **Action:** Merge PR with `Status: Accepted`
- **Feature Issue:**  Keep open for implementation tracking
- **Next:** Teams can begin implementation

#### 🔄 Needs Revision
- Good direction but needs changes
- **Action:** Author updates proposal based on feedback
- **Iterate** until concerns addressed
- Return to review

#### ❌ Rejected
- Technical infeasibility discovered
- Unfixable security vulnerabilities
- Performance concerns that can't be mitigated
- Breaking changes not acceptable

**When Rejected:**
1. **Update proposal** in the PR:
   - Change status to `Rejected`
   - Add detailed rejection reasoning:
     - Why it was rejected
     - What technical/security issues were found
     - What alternatives should be considered
     - What was learned from this proposal
2. **Merge the PR** (even though rejected)
   - Provides historical record
   - Helps future contributors understand why this won't work
   - Documents decision-making process
3. **Close Feature Issue** with explanation
4. **Update Architecture Discussion** with outcome

**Why merge rejected proposals?**
- Documents "what not to do" and why
- Prevents duplicate failed attempts
- Transparent decision-making
- Educational for community


## Directory Structure

```
docs/proposals/
├── README.md                                # This file
├── xxxx-design-proposal-template.md         # Proposal template
├── 0042-oauth2-device-flow.md               # Example: OAuth2 Device Flow
├── assets/
│   ├── 0042-assets/
│   │   ├── device-flow-sequence.png
│   │   └── user-journey.png
```

## Improving the Proposal Process

Have suggestions to improve the proposal process itself? Please open a GitHub Discussion to share your ideas and gather feedback from the community.

---

## Quick Reference

## Simplified Overview

**Path A: Complex/Unclear Problem**
```
Ideas Discussion → Feature Issue → Assignment → Design Proposal Discussion → Proposal PR → Approved
     💭               🎯              👤                     💬                    📄             ✅
```

**Path B: Clear Problem**
```
Feature Issue → Assignment → Design Proposal Discussion → Proposal PR → Approved
     🎯            👤                    💬                    📄             ✅
```

### When to Create What:

| Situation | Create | Category/Template |
|-----------|--------|-------------------|
| Unclear problem needing refinement | Discussion | Ideas |
| Clear problem ready to be worked on | Issue | Feature Issue Template |
| Ready for design  | Discussion | Design Proposals (after assignment) |
| Consensus reached on design | File + PR | Copy design proposal template |

---

## Example Journey

**Scenario:**  Adding passwordless authentication with passkeys

1. **Ideas Discussion** (optional): "How can we eliminate passwords for Thunder users?"
   - Community discusses: WebAuthn vs other approaches, browser support, UX concerns
   - Consensus: Implement FIDO2/WebAuthn for passkey support
2. **Feature Issue #42** created: "Add Passkey (WebAuthn/FIDO2) Authentication"
   - Problem: Users suffer from password fatigue, weak passwords, and phishing attacks. Passwords are the weakest link in authentication security.
   - Why now: Rising adoption of passkeys, improved browser support, security benefits.
   - Who for: End users wanting secure, easy login; admins wanting to reduce password-related support costs.
3. **Contributor claims** Issue #42, gets assigned by maintainer
4. **Architecture Discussion** created, linked to #42
   - Community discusses design options
   - Consensus: Use WebAuthn standard, support platform and roaming authenticators, fallback to existing MFA methods.
5. **Architecture Proposal** written: `0042-oauth2-device-flow.md`
   - Complete API specs, database schema, security analysis
   - Status: Accepted (default)
6. **PR submitted** for review
   - Security team reviews threat model
   - Maintainers review design and architecture
   - Approved and merged
7. **Implementation** begins
8. **Completion**
   - Proposal updated to `Status: Implemented`
   - Feature Issue #42 closed
   - Shipped in Thunder v1.5.0
