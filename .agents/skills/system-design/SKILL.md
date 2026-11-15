---
name: system-design
description: Create a ByteByteGo-inspired system-design topic from a course URL, including a clear TL;DR README, project index update, and tracker entry.
---

# ByteByteGo System Design

Create one self-contained study topic in an existing system-design repository from a ByteByteGo course page.

## When to use

Use this skill when the user asks to add a system-design topic based on a ByteByteGo course URL. Do not use it for generic documentation edits, unrelated code changes, or commits and pushes unless the user explicitly requests those actions.

## Workflow

1. Inspect the target repository before editing:
   - Read the nearest `AGENTS.md` or `agents.md`.
   - Check Git status and preserve unrelated changes.
   - Inspect the project README and one or two neighboring topic READMEs for naming and structure.
   - Confirm the requested topic directory name. Preserve it exactly, even if it contains a typo; use the standard spelling in the human-readable README title when appropriate.
2. Read the linked ByteByteGo course page. Extract only the material needed for an accurate study note: problem scope, assumptions, capacity estimates, APIs, data model, core algorithms, request flows, scaling choices, failure modes, and trade-offs.
3. Write the topic README for a reader with basic system-design knowledge:
   - Open with a one-sentence summary.
   - Follow with a concise TL;DR organized around the system’s decisions and flows.
   - Explain important technical terms when first introduced.
   - Use tables, numbered flows, and code blocks only when they improve clarity.
   - Preserve assumptions and distinguish them from general recommendations.
   - Include a source link to the ByteByteGo course page.
   - Rewrite in original language; do not reproduce long source passages.
4. Update the project index README with the new topic and place it in the learning path when that README has one.
5. Follow the repository’s tracker rule. For a new topic, create `tasks/tracker.md` with today’s date and a one-sentence summary of the interaction. Put task checkboxes only in `tracker.md` if checkboxes are needed.
6. Validate the result:
   - Run `git diff --check`.
   - Confirm the requested directory, README, tracker, index link, title, and source link exist.
   - Follow repository-specific test guidance; skip Go tests when only Markdown or other documentation files changed.

## Boundaries

- Create only the requested topic documentation and the project-index/tracker updates required by repository guidance.
- Do not add implementation code, diagrams, images, dependencies, or auxiliary scripts unless the user explicitly requests them or the repository requires them.
- Do not commit, push, or modify unrelated untracked or modified files without explicit user authorization.
