# Playbook — Node / TypeScript

Load only if `stack.md` lists `node` or `typescript`.

## Signals

`package.json`, `pnpm-lock.yaml` / `package-lock.json` / `yarn.lock`.

## Habits

- Package manager: the lockfile already in the repo. Do not switch
  npm↔pnpm↔yarn.
- Strict TS if the repo is already TS; do not add TypeScript to plain
  JS without a SPEC.
- Tests: the script in `stack.md` (`pnpm test`, `npm test`, …).
- Env vars: the project's mechanism. Do not read secrets from
  committed files.

## Do not

- Add Next/Express/Fastify if the runtime is already something else.
- Commit `node_modules`.
- Change the bundler in passing (vite, webpack, esbuild).
