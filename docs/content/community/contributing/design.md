---
title: Design Guide
description: Learn how to create design proposals for approved Thunder features, including collaboration, consensus building, and formal review processes.
hide_table_of_contents: false
---

# Design Guide - Creating Design Proposals

This guide explains how to create design discussions for approved Thunder features. The design discussion serves as your proposal for how to solve the problem.

---

## 🎯 When to Use This Guide

Create a design discussion when:

* ✅ You've been **assigned** to a feature request  
* ✅ The feature requires **architectural changes** or **new capabilities**  
* ✅ Community consensus exists on **what** to build  
* ✅ You need to specify **how** it will be built

**Don't create a design discussion for**:

* ❌ Bug reports → Use [bug report](README.md#-reporting-bugs)
* ❌ Minor improvements → Use [fast track](README.md#-making-improvements)
* ❌ Unclear ideas or feature requests → Use [problem definition guide](problem-definition.md)
* ❌ Features you're not assigned to → Express interest on the feature issue first

---

## Step 1: Create a Design Discussion

Start a design discussion to explore **how** to solve the problem and gather community feedback on your approach.

### Prerequisites

* \[ \] Feature issue approved and assigned to you  
* \[ \] You understand the problem (read the feature issue thoroughly)  
* \[ \] You've researched existing solutions and standards

### Creating the Discussion

1. Go to [GitHub Discussions](https://github.com/asgardeo/thunder/discussions/new/choose)  
2. Select **"Design"** category  
3. Fill in all sections

---

## Step 2: Community Collaboration

Once posted, collaborate with the community to refine the design.

### Who Participates?

**Design Author (you)**:

* Present the high-level approach  
* Respond to questions and feedback  
* Update the discussion as design evolves  
* Drive toward consensus

**Community Members**:

* Ask clarifying questions  
* Identify edge cases and potential issues  
* Share relevant experience  
* Suggest alternative approaches  
* Validate security considerations

**Maintainers**:

* Provide architectural guidance  
* Flag integration concerns  
* Ensure alignment with Thunder's vision  
* Guide toward feasible solutions

---

### Building Consensus

**Good signs of consensus**:

* ✅ No major objections from maintainers  
* ✅ Security concerns identified and addressed  
* ✅ Integration points clarified  
* ✅ Positive feedback from community  
* ✅ Open questions resolved

**Red flags**:

* ❌ Maintainers expressing concerns about approach  
* ❌ Multiple people suggesting different alternatives  
* ❌ Unresolved security issues  
* ❌ Technical feasibility questioned  
* ❌ Conflicts with existing architecture

**If consensus isn't reached**:

* Keep discussing (may take 3-4 weeks)  
* Consider a proof-of-concept implementation  
* Bring to community call for live discussion  
* In some cases, the feature may be declined at this stage

---

## Step 3: Design Review

Maintainers will formally review your design proposal.


### Review Outcomes

#### ✅ Approved

**What this means**: Design is sound, ready for implementation.

**Next steps**:

1. You can begin implementation (or someone else can)

**Move to**: [Implementation Guide](#after-your-design-is-approved)

---

#### 🔄 Needs Revision

**What this means**: Good direction, but changes needed.

**Common revision requests**:

* Clarify component interactions  
* Add missing security considerations  
* Provide more implementation details  
* Address backward compatibility concerns  

**Next steps**:

1. Address feedback in the design   
2. Iterate until approved

**Timeline**: Usually 1-2 revision cycles

---

#### ❌ Rejected

**What this means**: Design has fundamental issues that can't be resolved.

**Common rejection reasons**:

* Technical infeasibility discovered  
* Unfixable security vulnerabilities  
* Performance concerns that can't be mitigated  
* Breaking changes not acceptable  
* Conflicts with architectural principles

**When rejected**:

1. **Close feature issue** with explanation

2. **Update design discussion** with outcome

**This is okay!** Not all designs work out. Better to discover issues during design than during implementation.

---

## 🔄 After Your Design is Approved

### Implementation Phase

**You can**:

* Implement it yourself (move to [Development Guide](development.md))  
* Let someone else implement it  
* Collaborate with multiple implementers

**Track progress**:

* Feature issue stays open until implemented  
* Implementation PRs reference the design discussion  
* You're credited as the design author

---

### Design Evolution

**Minor changes during implementation**:

* Expected and normal  
* Document in implementation PRs  
* No need to update design discussion

**Major changes during implementation**:

* Bring back to design discussion  
* Update design discussion if significant  
* Get maintainer approval for major pivots

---

**Remember**: Good designs take time. A thorough design saves weeks of rework during implementation. Take the time to get it right! 🎯⚡
