# Orchestra Launch Plan

Orchestra has a strong technical foundation, but the launch should sell the outcome first: one cockpit for running multiple coding agents together.

## Positioning

**Primary message**

> A web cockpit for running Claude Code, Gemini CLI, and Aider together.

**Secondary message**

> Coordinate agent terminals, route handoffs through chat, and track implementation work from one browser workspace.

## Target users

- Developers using Claude Code, Gemini CLI, Aider, or similar terminal coding agents.
- Small teams experimenting with multi-agent software development.
- AI tool builders who want a reusable terminal/session orchestration layer.

## 7-day launch checklist

### Day 1: README conversion pass

- [ ] Add a strong hero statement.
- [ ] Add screenshots or a short GIF above the feature list.
- [ ] Make the quick start easier to scan.
- [ ] Move detailed architecture lower on the page.

### Day 2: Demo recording

- [ ] Record a 45-60 second demo using `docs/DEMO_SCRIPT.md`.
- [ ] Export as GIF and MP4.
- [ ] Put the GIF in `docs/assets/orchestra-demo.gif`.
- [ ] Link the MP4 from the README or release post.

### Day 3: One-command local path

- [ ] Add or verify a `docker compose up` path.
- [ ] Document the happy path from clone to browser.
- [ ] Add troubleshooting for missing Claude/Gemini/Aider CLIs.

### Day 4: Social launch copy

- [ ] Write one X thread.
- [ ] Write one Hacker News post.
- [ ] Write one Reddit post for AI coding and open-source communities.
- [ ] Prepare 3 short clips: live terminals, multi-agent chat, Kanban task handoff.

### Day 5: Integrations and examples

- [ ] Add examples for Claude Code, Gemini CLI, and Aider.
- [ ] Add a sample workspace configuration.
- [ ] Add a sample task flow: architect agent -> implementation agent -> review agent.

### Day 6: Community hooks

- [ ] Open issues tagged `good first issue`.
- [ ] Open issues tagged `help wanted` for integrations.
- [ ] Add a contributor guide focused on providers and skills.

### Day 7: Distribution

- [ ] Submit to relevant awesome lists.
- [ ] Post the demo to X, Reddit, V2EX, and Hacker News.
- [ ] Ask early users for screenshots and bug reports.
- [ ] Track README views, stars, issues, and community interest if available.

## Metrics

| Stage | Goal | Signal |
| --- | --- | --- |
| First 48 hours | Clear positioning | People can describe Orchestra in one sentence |
| First week | 100+ new stars | Demo gets shared without explanation |
| First month | 500+ new stars | Users open issues for integrations and setup |
| 90 days | 3k+ stars | Orchestra becomes a reference project for multi-agent coding workflows |

## What not to do first

- Do not lead with internal architecture.
- Do not add more provider abstractions before the demo is compelling.
- Do not polish every screen before shipping screenshots.
- Do not pitch it as a generic collaboration platform. The wedge is AI coding agents.
