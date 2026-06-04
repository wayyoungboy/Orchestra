# Orchestra Demo Script

Use this script to record the first public demo. Keep the video short, visual, and focused on the outcome: multiple coding agents working from one browser cockpit.

## Demo title

**I ran Claude Code, Gemini CLI, and Aider together from one web cockpit**

## Recording length

Target: 45-60 seconds.

## Setup before recording

- Start the backend and frontend.
- Open Orchestra in the browser.
- Prepare one workspace that points to a small codebase.
- Prepare three members or sessions:
  - `Architect` using Claude Code
  - `Builder` using Gemini CLI or Aider
  - `Reviewer` using another agent terminal
- Create one simple task, such as adding a health check endpoint, writing tests, or refactoring a small component.

## Shot list

### 1. Hook: show the cockpit

**Visual:** Browser view with workspace, chat, terminal sessions, and task board.

**Voiceover / caption:**

> Most coding agents run alone in separate terminals. Orchestra puts them in one web cockpit.

### 2. Start multiple agents

**Visual:** Open or switch between Claude Code, Gemini CLI, and Aider sessions.

**Voiceover / caption:**

> Start Claude Code, Gemini CLI, and Aider side by side from the same workspace.

### 3. Assign a task

**Visual:** Create or move a task in the Kanban board. Mention the right assistant in chat.

**Voiceover / caption:**

> Assign work through chat and track it as a task instead of copying terminal output between windows.

### 4. Show terminal output streaming

**Visual:** Show live terminal output in the browser.

**Voiceover / caption:**

> Terminals stream in real time, with every agent attached to the same workspace context.

### 5. Show handoff

**Visual:** Builder posts a result, reviewer gets mentioned, task moves forward.

**Voiceover / caption:**

> One agent can implement, another can review, and the team can follow the workflow from one place.

### 6. Close with the tagline

**Visual:** README or landing screen.

**Voiceover / caption:**

> Orchestra is a web cockpit for running coding agents together.

## GIF version

Create a shorter 12-18 second GIF:

1. Show the dashboard.
2. Switch between two agent terminals.
3. Send one chat message with an @mention.
4. Move one task on the board.

Put the GIF at:

```text
docs/assets/orchestra-demo.gif
```

Then add this to the README hero area:

```md
![Orchestra demo](docs/assets/orchestra-demo.gif)
```

## Social captions

### X / Twitter

> I built Orchestra: a web cockpit for running Claude Code, Gemini CLI, and Aider together.
>
> Multi-agent terminals, chat routing, and task tracking in one workspace.

### Hacker News

> Show HN: Orchestra - a web cockpit for running multiple coding agents together

### Reddit

> I wanted a way to coordinate Claude Code, Gemini CLI, and Aider without juggling separate terminals, so I built Orchestra: a browser cockpit with agent sessions, chat routing, and task tracking.
