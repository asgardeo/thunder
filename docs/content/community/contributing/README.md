---
title: Contributing
description: Learn how to contribute to the Thunder project, including reporting bugs, improving documentation, and implementing features.
hide_table_of_contents: false
---

# Contributing to Thunder

Thank you for your interest in contributing to Thunder! We welcome contributions of all kinds—from fixing typos to proposing major features. This guide will help you understand how you can contribute.

---

## 🎯 Ways to Contribute

Thunder welcomes different types of contributions, and you can participate at any level:

| Contribution Type | Documentation |
| ----- | ----- |
| **Bug Reports** | [Report a bug](#-reporting-bugs) |
| **Documentation** | [Improve docs](#-improving-documentation) |
| **Bug Fixes** | [Fix a bug](#-fixing-bugs) |
| **Minor Improvements** | [Make improvements](#-making-improvements) |
| **Share Ideas & Define Problems** | [Problem definition](#-sharing-ideas--defining-problems) |
| **Design Features** | [Design proposals](#-designing-solutions) |
| **Implement Features** | [Development guide](development.md) |

**You don't need to do everything!** You can:

* Share an idea and let someone else design & implement it
* Design a solution and let someone else implement it
* Implement someone else's approved design
* Mix and match based on your interests and availability

---

## 🐛 Reporting Bugs

Found a bug? Help us fix it!

### Before You Report

* Search [existing issues](https://github.com/asgardeo/thunder/issues) to avoid duplicates
* Check if the bug exists in the latest version
* Gather logs, error messages, and reproduction steps

### Create a Bug Report

1. Go to [Issues → New Issue](https://github.com/asgardeo/thunder/issues/new/choose)
2. Select **"Bug Report"** template
3. Fill in all required sections
4. Submit the issue

**Good bug reports include**:

* Clear, specific title: "OAuth2 token refresh fails with 401 after 30 minutes"
* Minimal reproduction steps
* Actual vs. expected behavior
* Logs and error messages (sanitize sensitive data!)
* Environment details

---

## 📚 Improving Documentation

Documentation improvements are always welcome!

### Types of Documentation

* **User guides** - How to use Thunder features
* **API documentation** - Endpoint specifications
* **Admin guides** - Configuration and deployment
* **Code comments** - Inline explanations
* **Examples** - Sample code and tutorials

### How to Contribute

#### Option 1: Fix Documentation Directly (Encouraged!)

We encourage you to fix documentation yourself when possible:

1. Identify what needs improvement (typos, outdated info, missing content)
2. Fork the repo and make changes
3. Submit a PR with clear description of what you changed

#### Option 2: Report Documentation Issues

If you spot documentation problems but can't fix them yourself:

1. Go to [Issues → New Issue](https://github.com/asgardeo/thunder/issues/new/choose)
2. Select **"Documentation Issue"** template or create a general issue
3. Describe what needs improvement (missing info, errors, unclear sections)
4. Submit the issue for others to fix

**Tips**:

* Write for beginners (don't assume knowledge)
* Use examples and code snippets
* Test any code examples you add
* Check spelling and grammar

---

## 🔧 Fast Track Contributions

**Bug fixes, documentation, and minor improvements don't require a design proposal.**

### What Qualifies for Fast Track?

✅ **YES - Use Fast Track**:

* Bug fixes
* Documentation improvements
* Minor API additions (no architectural changes)
* Performance optimizations (no behavior changes)
* Test improvements
* Code cleanup/refactoring (same behavior)

❌ **NO - Use Feature Process**:

* New authentication/authorization flows
* Breaking API changes
* New IAM capabilities
* Database schema changes
* Significant architectural changes

**When in doubt?** Ask in [Discussions](https://github.com/asgardeo/thunder/discussions/new?category=general)

---

### Fast Track Process

1. **Create or find an issue** - Search first to avoid duplicates
2. **Get assigned** - Comment "I'd like to work on this" and wait for maintainer assignment
3. **Fork the Thunder repository** - Create your own copy to work on
4. **Make your changes** - Include tests for code changes
5. **Submit a pull request** - Reference the issue number
6. **Address review feedback** - Work with reviewers to refine your changes
7. **Get merged** - Once approved, your contribution will be merged! 🎉

For detailed development guidelines, see our [development documentation](development.md).

---

## 🔨 Fixing Bugs

Found a bug and want to fix it? Great! Follow the [Fast Track process](#-fast-track-contributions) above.

**Before you start**:

* Check if someone is already working on it
* Understand the root cause
* Add tests to prevent regression

---

## ✨ Making Improvements

Have an idea for a small improvement? Use the [Fast Track process](#-fast-track-contributions).

**Good improvements include**:

* Better error messages
* Input validation enhancements
* Logging improvements
* Code readability refactoring
* Performance optimizations

---

## 💡 Sharing Ideas & Defining Problems

Have an idea or problem to share? Want to propose a new feature?

**This is where all features start!**

You can either:
* **Start a discussion** if your idea needs exploration and community input
* **Create an issue directly** if you have a clear, well-defined feature request

**Choose your path**:

| Your Situation | Action |
|---------------|--------|
| You have a clear problem statement and can provide all details required in the feature template | Create a [Feature Issue](https://github.com/asgardeo/thunder/issues/new?template=feature.yml) |
| You have a basic idea that needs exploration or clarification before creating a formal feature request | Create a [Discussion in Idea Category](https://github.com/asgardeo/thunder/discussions/new?category=ideas) |

**→ Learn more**: [Problem Definition Guide](problem-definition.md)

---

## 📐 Designing Solutions

Assigned to a feature? Time to design the technical solution.

**The design phase involves**:

* Creating a design discussion
* Collaborating on architecture
* Writing a formal design proposal
* Getting security and technical review

**→ Learn more**: [Design Guide](design.md)

---

## 🔨 Implementing Features

Ready to build an approved design?

**Implementation includes**:

* Breaking down the design into tasks
* Writing code with comprehensive tests
* Documenting the feature
* Deploying with feature flags

**→ Learn more**: [Development Guide](development.md)

---

## 👥 How Contributions Work Together

Thunder features often involve contributions from multiple people:

* **Idea contributor** → Shared the problem in a discussion
* **Problem definer** → Created the formal feature request
* **Designer** → Wrote the design proposal
* **Implementer** → Built the feature
* **Documenter** → Wrote user guides
* **Tester** → Provided early feedback

**You can contribute to any part of this journey!** Your contribution is valuable even if you don't complete the entire feature yourself.

Welcome to the Thunder community! ⚡🔐

---

**Questions?** Ask in [Discussions](https://github.com/asgardeo/thunder/discussions/new?category=q-a)
