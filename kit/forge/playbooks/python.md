# Playbook — Python

Load only if `stack.md` lists `python`.

## Signals

`pyproject.toml` or `requirements.txt`, `uv.lock` / `poetry.lock`,
`manage.py` (Django).

## Habits

- Dependency tool: the one with a lockfile in the repo (uv, poetry,
  pip). Do not migrate between them inside an unrelated SPEC.
- Type hints on new public functions if the repo already uses them.
- Tests: the command in `stack.md` (`pytest`, `uv run pytest`, …).
  Keep the existing test layout (`tests/`, `*_test.py`).
- Migrations: the project's tool (alembic, Django migrations). Never
  edit an applied migration; add a new one.
- Settings and secrets: the project's mechanism (env vars, settings
  module). Do not hardcode credentials.

## Do not

- Add FastAPI/Django/Flask if the runtime is already another one.
- Commit `.venv/`, `__pycache__/`, or a lockfile for a tool the repo
  does not use.
- Swap `requirements.txt` for `pyproject.toml` in passing.
