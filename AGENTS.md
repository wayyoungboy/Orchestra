# Orchestra Agent Notes

Use these notes when working in this repository with Codex or another coding agent.

## Project Shape

Orchestra is a browser-based multi-agent workspace. The backend is Go/Gin/SQLite/tmux, and the frontend is Vue 3/TypeScript/Pinia/Tailwind.

## Common Commands

```bash
cd backend && make test
cd frontend && pnpm install
cd frontend && pnpm build
```

After backend code changes, restart the Go process. After frontend dependency, env, proxy, or state changes, restart Vite and hard-refresh the browser if needed.

## Agent Integrations

- Claude Code is the primary supported agent terminal and can use stream-json mode.
- Codex is supported as a local CLI provider through `codex` sessions and `~/.codex/skills` symlinks.
- Keep `backend/configs/config.yaml` `security.allowed_commands` in sync when adding new terminal agent commands.

## Editing Rules

- Keep backend changes gofmt'd.
- Keep frontend changes typed and avoid broad UI refactors unless the task is explicitly visual.
- Do not commit local database files, terminal logs, or workspace artifacts.
