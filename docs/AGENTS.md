<!--
Thunder documentation creation instructions

Purpose:
Provide deterministic, enforceable instructions for an AI agent to create Thunder 
-->

# Thunder documentation creation instructions

Follow these instructions when creating new documenttion content for Thunder. Adhere to all guidelines to ensure consistency, clarity, and quality.

## Scope and boundaries

### Audience

- Your primary audience is Thunder users, including system administrators, developers, and IT professionals.
- Assume the audience has a basic understanding of IT concepts but may be unfamiliar with Thunder specifics.
- Avoid jargon and explain concepts clearly.

### What you must do

- On the top of the document, write "This document was created by an AI agent following the Thunder documentation creation instructions." in italics.
- Strictly adhere to the authoring standards outlined below.
- Choose the appropriate navigation location for the new content based on its topic and relevance.
- Create content that is clear, concise, and actionable for the intended audience.
- Ensure all technical details are accurate and up-to-date.
- Use the provided templates and formatting rules consistently.

## Authoring standards

You must follow these standards when creating documentation content for Thunder.

### Voice and tone

- Use active voice and present tense. You can only use passive voice when the actor is unknown or unimportant.
- Use plain language and short sentences.
- Address the reader as “you.”
- Keep a professional, friendly, neutral tone.
- Avoid slang, jokes, sarcasm, and marketing language.

### Terminology and consistency

- Use consistent product names and feature names.
- Define acronyms on first use. If the acronym is widely known (for example, API, URL), you can use it without definition.
- Use the same term for the same concept throughout the document.
- Use standard technical terms where applicable (for example, “server,” “database,” “API,” “endpoint”).
- Avoid ambiguous pronouns like “it” or “this” when the referent is unclear.
- Avoid using "am", "is", "are" as much as possible; prefer strong verbs.
- Prefer concrete nouns and verbs.

### Headings

- You must use sentence case for all headings. This includes h1, h2, h3 and all titles. Pronouns need to be capitalized as per standard grammar rules.
- Use consistent heading levels to reflect document structure.
- Make headings task-focused and descriptive. Do not use generic headings like “Introduction” or “Details.”
- Do not use decorative symbols (for example, “→”, “¶”, “✅”, emojis).

### Lists

- Use numbered lists for procedures and ordered steps.
- Use bulleted lists for non-sequential information.
- Keep list items parallel in grammar and structure.

### Formatting rules

- UI labels, buttons, menu items: use **bold**.
  - Example: Select **Save**.

- Code elements, file names, paths, config keys, commands, URLs: use backticks.
  - Example: Update `deployment.toml`.

- Use descriptive link text. Do not paste raw URLs as link text.
  - Example: `[Microsoft Writing Style Guide](https://learn.microsoft.com/en-us/style-guide/welcome/)`

### Code blocks
- Use fenced code blocks with a language tag when known.
- Keep code blocks focused.
- Do not include secrets, tokens, passwords, or realistic keys.

    Example:

    ```toml
    [server]
    hostname = "localhost"

### Configuration guidance

When documenting configuration:

- Describe what the setting controls.
- State the default value.
- State constraints (type, valid range, allowed values).
- Provide a minimal example.
- Explain when the user should change it.

### Links and references

#### Internal links

- Use descriptive link text.
- Prefer linking to canonical pages (overview or primary reference).
- Avoid linking to unstable or temporary resources.

#### External links

- Use external links sparingly and only when they add clear value.
- Use descriptive link text.
- Prefer authoritative sources (official documentation or standards).

### Images and screenshots

- Do not add, generate, or request new images or screenshots as part of documentation creation.
- Do not reference an image unless the user explicitly confirms it exists and is accessible.
- Do not make images required to complete a task. Provide text alternatives.

## Exampe document template

For a guide, use the following structure. You must follow this structure exactly. You are free to change the title names where appropriate, but the sections must remain the same and in the same order.

### Structure

- Title
- Purpose
- When to use
- Prerequisites
- Steps (numbered procedure)
- Validate (how to confirm it worked)
- Troubleshoot (optional)
- Next steps (optional)

## Creation workflow

### Step 1: Clarify the user goal (internally)

Before writing, identify:

- Primary user goal (what they want to achieve).
- Target audience (role, assumed knowledge).
- Product scope (feature, component, environment).
- Success criteria (what “done” means).

- If the user request is ambiguous, proceed with reasonable assumptions and clearly state them in a short “Assumptions” section near the top.

### Step 2: Outline before drafting

- Create a short outline using the required sections for the chosen document type.
- Keep the outline aligned with the user’s goal.

### Step 3: Write the first draft

- Start with the minimal content needed to complete the user goal.
- Use clear steps and expected outcomes.
- Keep paragraphs short (2–4 lines where possible).

### Step 4: Add examples

Add only examples that help complete the task.

- Keep examples minimal.
- Ensure examples are syntactically correct.
- Explain what the example does.

### Step 5: Validate for completeness

Confirm the draft includes:

- A clear outcome in the overview.
- Prerequisites (if needed).
- Steps that are executable.
- Validation guidance.
- Consistent formatting and terminology.

## Quality checklist (must pass)

Before finalizing output, ensure:

- Headings are sentence case.
- Procedures use numbered lists.
- UI labels are **bold**.
- Code elements and paths are in backticks.
- Links use descriptive text.
- Content is concise, active voice, present tense.
- No unverified claims or placeholders remain.
- No secrets or sensitive data appear in examples.
- After creating content, run Vale locally and resolve all warnings.

## Output requirements

- Output must be Markdown.
- Use a single top-level title (`# ...`).
- Use consistent section ordering and headings.
- If assumptions are made, include an **Assumptions** section near the top.
- End with a **Next steps** section when appropriate.

### Vale verification requirement

Before finalizing documentation output:

- Verify that Vale has been run against the file.
- If Vale results are available, resolve all reported errors and warnings.
- If Vale is not installed or Vale results cannot be verified:
  - Prompt the user to install Vale locally.
  - Provide the official installation instructions.
  - Ask the user to rerun Vale and share the results.
  - Do not assume compliance or guess fixes without Vale feedback.

Vale installation reference:
- https://vale.sh/docs/vale-cli/installation/