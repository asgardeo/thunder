# User Registration

Thunder provides flexible self-registration capabilities that allow users to create accounts in your applications. Whether you're building a consumer application, B2B portal, or mobile app, Thunder's registration flows can be customized to match your onboarding requirements.

## 🎯 Overview

Self-registration (also known as self-service sign-up) enables users to create their own accounts without administrator intervention. Thunder currently support only the server orchestrated registration flows.

## 🚀 Registration Approaches

### 1. Server Orchestrated Registration Flows

REST API-based registration where Thunder orchestrates the entire process step-by-step.

**Use this when:**
- Building native mobile applications (iOS, Android)
- Developing single-page applications (SPAs) with custom UI
- Implementing multi-step registration with custom validation

**Supported registration methods:**
- **Username and Password** - Traditional account creation
- **SMS OTP** - Phone number verification during registration
- **Social Sign-Up** - Google, GitHub account-based registration
- **Progressive Profiling** - Collect user data across multiple steps

[Learn more about Server Orchestrated Registration →](./server-orchestrated-flow/registration.md)
[Learn more about customizing registration flows →](./server-orchestrated-flow/customize-registration-flow.md)
