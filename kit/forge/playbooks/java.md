# Playbook — Java

Load only if `stack.md` lists `java`.

## Signals

`pom.xml` or `build.gradle`, `src/main/java`, JUnit.

## Habits

- Public API and code comments follow `memory/constitution.md` (English
  if it does not say otherwise).
- Do not introduce Spring where the project does not use it; do not
  remove Spring where it is the runtime.
- Persistence: keep the project's mechanism (JPA, JDBC, MyBatis). Do
  not add an ORM a SPEC did not ask for.
- Tests: the command in `stack.md` (e.g. `./gradlew test`).
- Domain errors: typed errors the project already uses, not ad-hoc
  strings at the HTTP edge if an envelope exists.

## Do not

- Migrate to Kotlin/Go/Node without a SPEC.
- Switch `ddl-auto` to `create`/`update` in a project with migrations.
