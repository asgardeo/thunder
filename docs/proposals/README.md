# Thunder Proposals

This directory contains design proposals for Thunder. Proposals are used to document significant changes to the platform before implementation.

## When to Write a Proposal

A proposal is required for:

- New authentication flows or protocols
- New identity provider integrations
- Major architectural changes to the IAM core
- Breaking changes to APIs or authentication behavior
- New authorization models or policy engines
- Changes to token management or session handling
- Security-related enhancements
- New features that impact multiple components
- Changes affecting the developer/administrator experience
- Database schema changes affecting identity or access data
- New compliance or audit features

## Proposal Process

### 1. Create a GitHub Issue

First, create a GitHub issue describing the problem or enhancement you want to address. This issue will be assigned a number that becomes the proposal identifier.

### 2. Open a Discussion

After creating the issue, open a [GitHub Discussion under the "Proposals" category](https://github.com/asgardeo/thunder/discussions/categories/proposals) to:

- Present your initial design ideas
- Gather feedback from the community and security experts
- Discuss security implications and threat models
- Iterate on the design based on input
- Build consensus before formal documentation
- Validate compliance requirements if applicable

### 3. Write the Proposal

Once the discussion reaches consensus:

1. Copy the `xxxx-proposal-template.md` file
2. Rename it to `<issue-number>-<descriptive-title>.md` where:
   - `<issue-number>` is zero-padded to 4 digits (e.g., 0042)
   - `<descriptive-title>` uses kebab-case (e.g., oauth2-device-flow)
   - Example: `0042-oauth2-device-flow.md`

3. If your proposal includes diagrams or images:
   - Create a directory `<issue-number>-assets/` under the `assets` directory (e.g., `assets/0042-assets/`)
   - Place all sequence diagrams, architecture diagrams, flow charts, and other assets in this directory
   - Reference them from your proposal using relative paths (e.g., `![Auth Flow](assets/0042-assets/auth-flow.png)`)
   - **Recommended diagrams for IAM proposals:**
     - Authentication/authorization flow diagrams
     - System architecture diagrams
     - Token flow diagrams
     - User journey diagrams

4. Fill out all sections of the template
5. Reference both the GitHub issue and discussion if applicable

### 4. Submit via Pull Request

1. Create a PR with your proposal file
2. Link the PR to the original GitHub issue and discussion
3. Request reviews from relevant maintainers

### 5. Proposal Status Lifecycle

Once merged into the repository, proposals have these statuses:

- **Accepted**: Proposal approved, implementation can begin
- **Rejected**: Proposal was not accepted (kept with rejection reasoning documented)
- **Implemented**: Feature is complete and merged

Note: Both approved and rejected proposals remain in the repository for historical reference and to document decision rationale.


## Directory Structure

```
docs/proposals/
├── README.md                              # This file
├── xxxx-proposal-template.md              # Proposal template
├── 0042-oauth2-device-flow.md            # Example: OAuth2 Device Flow
├── assets/
│   ├── 0042-assets/
│   │   ├── device-flow-sequence.png
│   │   └── user-journey.png
```

## Improving the Proposal Process

Have suggestions to improve the proposal process itself? Please open a GitHub Discussion to share your ideas and gather feedback from the community.

---

## Quick Reference: TEP Lifecycle

```
1. GitHub Issue (#123)
   ↓
2. GitHub Discussion (Ideas & Design)
   ↓
3. Write Proposal (0123-feature-name.md)
   ↓
4. Submit PR (Review & Iterate)
   ↓
5. Merge (Status: Accepted)
   ↓
6. Implementation (Create impl issues)
   ↓
7. Update Status (Status: Implemented)
```