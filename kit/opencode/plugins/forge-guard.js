// Forge guard for opencode.
//
// opencode has no PreToolUse hook, so the equivalent of the Claude Code
// guard is a plugin: before opencode writes to a file or runs a shell
// command, ask `forge guard`. `forge guard --file <path>` and
// `forge guard --command <cmd>` exit 1 and print the reason when the action
// must be denied.
//
// The rule itself lives in the Go binary, so this file is only an adapter.
// Set `guard: off` in .forge/project.md, or re-run `forge init --no-guard`,
// to turn the denial off.

const EDIT_TOOLS = {
  edit: (args) => [args && args.filePath],
  write: (args) => [args && args.filePath],
  apply_patch: (args) => patchPaths(args && args.patchText),
};

// apply_patch hides its paths inside the patch markers, one per operation.
function patchPaths(text) {
  if (typeof text !== "string") return [];
  const out = [];
  for (const line of text.split("\n")) {
    const m =
      line.match(/^\*\*\* (?:Add|Update|Delete) File:\s*(.+)$/) ||
      line.match(/^\*\*\* Move to:\s*(.+)$/);
    if (m) out.push(m[1].trim());
  }
  return out;
}

function throwIfDenied(result) {
  if (result.exitCode === 0) return;
  const reason =
    (result.stderr && result.stderr.toString().trim()) ||
    (result.stdout && result.stdout.toString().trim()) ||
    "Forge: blocked by the guard.";
  throw new Error(reason);
}

export const ForgeGuard = async ({ $, directory }) => {
  return {
    "tool.execute.before": async (input, output) => {
      if (input.tool === "bash") {
        const command = output.args && output.args.command;
        if (!command) return;
        throwIfDenied(
          await $`forge guard --command ${command}`.cwd(directory).nothrow().quiet()
        );
        return;
      }
      const files = EDIT_TOOLS[input.tool];
      if (!files) return;
      for (const file of files(output.args)) {
        if (!file) continue;
        throwIfDenied(
          await $`forge guard --file ${file}`.cwd(directory).nothrow().quiet()
        );
      }
    },
  };
};
