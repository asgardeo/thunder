---
title: Problem Definition Guide
description: Learn how to share ideas, explore problems, and create feature requests for Thunder through discussions and formal issues.
hide_table_of_contents: false
---

# Problem Definition Guide - Sharing Ideas and Defining Features

This guide explains how to share ideas, explore problems, and create feature requests for Thunder.

---

## 🎯 When to Use This Guide

Use this guide when:

* ✅ You have an idea or problem to share
* ✅ You want to propose a new feature or capability
* ✅ You're exploring whether something is worth pursuing
* ✅ You want community input on a problem

**Don't use this guide for**:

* ❌ Bug reports → Use [bug report](README.md#-reporting-bugs)
* ❌ Minor improvements → Use [fast track](README.md#-making-improvements)

---

## 📝 Two Ways to Share

Choose the approach that fits your situation:

### Option 1: Discussion (Idea Not Fully Formed) 💭

**Use when**:
* Problem needs refinement or exploration
* You want community input before formal request
* The idea is complex and needs discussion
* You're not sure if it's worth pursuing

**Process**: Create a discussion → Collaborate → Create issue (if validated)

**→ [Jump to Discussion Guide](#-creating-a-discussion)**

---

### Option 2: Issue (Clear Feature Request) 🎯

**Use when**:
* You have a clear, well-defined problem
* You can answer: What? Why? Who?
* The feature is ready for formal consideration
* You've already validated the need

**Process**: Create an issue directly

**→ [Jump to Feature Issue Guide](#-creating-a-feature-issue)**

---

### Not Sure Which to Use?

**Start with a discussion if**:
* You're exploring an idea
* The problem isn't clearly defined yet
* You want feedback before formal submission

**Create an issue directly if**:
* The problem is specific and observable
* You have clear answers to What/Why/Who
* The need is validated (you or others experience it)

**When in doubt**: Start with a discussion. You can always create an issue later!

---

## 💡 Creating a Discussion

Use discussions to explore problems and refine ideas with the community.

---

### Step 1: Identify the Problem

Before sharing, crystallize your thinking:

**Ask yourself**:
* What problem am I observing?
* Who is experiencing this problem?
* How often does this happen?
* What's the impact when it happens?
* Have I searched for existing discussions/issues?

**Good problems are**:
* Specific (not "Thunder needs better UX")
* Observable (not theoretical)
* Impactful (affects real users)
* Current (not solved elsewhere)

---

### Step 2: Create the Discussion

1. Go to [GitHub Discussions](https://github.com/asgardeo/thunder/discussions/new/choose)
2. Select **"Ideas"** category
3. Fill in your idea

### Step 3: Community Collaboration

Once posted, the community will help refine your idea.

**Expect questions like**:
* "Have you considered...?"
* "How would this work for...?"
* "What about existing users who...?"
* "Is this similar to...?"

**Your role**:
* ✅ Respond to questions promptly
* ✅ Clarify the problem (not the solution)
* ✅ Share more context if needed
* ✅ Be open to alternative perspectives
* ✅ Update the original post as thinking evolves

**Community's role**:
* 🤝 Ask clarifying questions
* 💡 Share similar experiences
* 🔍 Identify edge cases
* 📚 Point to related issues/discussions
* 🎯 Help scope the problem

**Timeline**: Most discussions reach clarity within 1-2 weeks.

---

### Step 4: Discussion Outcomes

After community discussion, the idea will evolve:

#### ✅ Valid Feature Idea

**What it means**: Clear problem, valuable to solve, worth pursuing.

**Next steps**:
1. Create a [Feature Request issue](#-creating-a-feature-issue)
2. Link back to this discussion
3. Close the discussion with:
   ```
   Thanks everyone! Created feature request: #123
   Closing this discussion as we're moving to formal feature definition.
   ```

**You can stop here!** Someone (could be you or others) will take it forward.

---

#### 🐛 It's Actually a Bug

**Next steps**:
1. Create a [Bug Report](https://github.com/asgardeo/thunder/issues/new?template=bug.yml)
2. Close the discussion

**Outcome**: Bug fix process (faster than feature development).

---

#### 🔧 It's a Minor Improvement

**Next steps**:
1. Create an [Improvement Issue](https://github.com/asgardeo/thunder/issues/new?template=improvement.yml)
2. Close the discussion

**Outcome**: Faster path to implementation.

---

#### ❌ Not Viable / Out of Scope

**What it means**: Good idea, but not appropriate for Thunder.

**Reasons**:
* Out of Thunder's scope
* Already solved differently
* Technical infeasibility
* Conflicts with Thunder's vision

**Outcome**: Maintainer explains reasoning, discussion closed and documented.

**This is okay!** Not all ideas should be implemented.

---

## 🎯 Creating a Feature Issue

Use feature issues to formally define problems that need solving.

---

### Before You Create

Ensure you can answer these **three critical questions**:

#### Question 1: What problem are we solving?

Be specific about the problem, not the solution.

**Bad** ❌: "Thunder needs SAML support"  
**Good** ✅: "Enterprise customers cannot integrate Thunder with their existing SAML-based SSO systems, forcing them to maintain duplicate user identities."

**Template**:
```
When [user type] wants to [goal], they [current limitation/problem], 
which causes [negative impact].
```

---

#### Question 2: Why should we solve this now?

Provide justification for prioritization.

**Consider**:
* **User impact**: How many users? How severely affected?
* **Business value**: Revenue? Customer requests? Competitive gap?
* **Technical debt**: Does delaying make it harder later?
* **Compliance**: Regulatory or security requirements?

**Example**:
```
Why now:
- 12 enterprise customers requested this in Q3 (3 deals blocked)
- Competitors have this as standard (Okta, Auth0)
- Largest prospect requires this by end of year
- Estimated $500K annual revenue impact
```

---

#### Question 3: Who are we solving this for?

Identify user personas and their goals.

**Personas to consider**:
* **End Users**: People using the application
* **Administrators**: Configure and manage Thunder
* **Developers**: Integrate applications with Thunder
* **Security Teams**: Ensure compliance and security

**Example**:
```
Primary: Enterprise IT administrators
- Goal: Onboard 1000+ users quickly with minimal effort
- Pain: Manual account creation takes days
- Success metric: 1000 users onboarded in <1 hour

Secondary: End users
- Goal: Use existing corporate credentials
- Pain: Need to remember separate Thunder password
- Success metric: Single sign-on with LDAP credentials
```

---

### Create the Issue

1. Navigate to [Issues → New Issue](https://github.com/asgardeo/thunder/issues/new/choose)
2. Select **"Feature Request"** template
3. Fill in all sections

---

## ✨ What Happens Next?

After you create your issue:

### Triage
Maintainers will review and:
* Evaluate strategic fit and priority
* Assess scope and effort
* Add appropriate labels
* Decide if approved for work

### Assignment
If approved:
* Available for community or maintainers to work on
* You (or others) can express interest
* Maintainers will assign to someone
* Work begins on design

### Questions?
* Comment on your issue
* Join [community discussions](https://github.com/asgardeo/thunder/discussions)

**Note**: Not all issues will be accepted. High-priority issues move faster.

---

### Want to Work On This?

After your issue is approved, anyone can work on it!

**To express interest**:
1. Comment: "I'd like to work on this"
2. Share relevant experience (optional but helpful)
3. Estimate your timeline
4. Wait for maintainer to assign you

**Once assigned**: Move to the [Design Guide](design.md) to start designing the solution.

**You can stop here!** Let someone else design and implement if you prefer.

---

## 💡 Tips for Success

### Do's ✅

**Focus on the problem, not the solution**:
* ❌ "Add LDAP integration using ldapjs library"
* ✅ "Users cannot sync from existing LDAP directories"

**Be specific and quantify impact**:
* ❌ "Login is slow"
* ✅ "Token validation takes 450ms (target: &lt;200ms), affecting 5000 req/s"

**Provide context and data**:
* ❌ "Customers want this"
* ✅ "12 customers requested in Q3, 3 deals blocked"

**Consider all affected users**:
* Think beyond just end users
* Include admins, developers, security teams

---

### Don'ts ❌

**Don't prescribe the solution**:
* Let the community collaborate on approach
* Focus on what/why, not how

**Don't skip the "why now?"**:
* Justification helps with prioritization
* Explain urgency and impact

**Don't forget to search first**:
* Check existing issues and discussions
* Reference related work

**Don't ghost the discussion**:
* Respond to questions
* Update as thinking evolves

---

## ❓ FAQ

**Q: Should I create a discussion or an issue?**  
A: Discussion if exploring, issue if problem is clear.

**Q: How long does triage take?**  
A: Usually within 1 week. Complex features may take longer.

**Q: What if my issue is declined?**  
A: Ask why in the issue. Sometimes timing is wrong, sometimes there's a better approach.

**Q: Can I work on my own issue?**  
A: Absolutely! Express interest after it's approved.

**Q: What if no one claims my approved issue?**  
A: It stays in backlog. We'll reassess quarterly. You can help recruit contributors!

**Q: Can I create an issue for someone else's discussion?**  
A: Yes, if the discussion reached consensus and the original author hasn't done it.

**Q: How do I know if my problem qualifies as a "feature"?**  
A: If it requires architectural changes or new capabilities, it's a feature. When in doubt, ask!

---

## 📞 Next Steps

### After Creating a Discussion
* Engage with community feedback
* Refine the problem
* Create an issue when ready (or let someone else)

### After Creating an Issue
* Wait for triage
* Respond to maintainer questions
* Express interest in working on it (optional)
* Move to [Design Guide](design.md) if assigned

---

**Thank you for helping shape Thunder's roadmap!** Every problem definition makes Thunder better. 🙏⚡

**Questions?** Ask in [Discussions](https://github.com/asgardeo/thunder/discussions/new?category=q-a)
