// Forge guard for opencode.
//
// opencode has no PreToolUse hook, so the equivalent of the Claude Code
// guard is a plugin: before opencode writes to a file, ask `forge guard`
// whether a spec is in `implementing`. `forge guard --file <path>` exits 1
// and prints the reason when the edit must be denied.
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

export const ForgeGuard = async ({ $, directory }) => {
  return {
    "tool.execute.before": async (input, output) => {
      const files = EDIT_TOOLS[input.tool];
      if (!files) return;
      for (const file of files(output.args)) {
        if (!file) continue;
        const result = await $`forge guard --file ${file}`
          .cwd(directory)
          .nothrow()
          .quiet();
        if (result.exitCode === 0) continue;
        const reason =
          (result.stderr && result.stderr.toString().trim()) ||
          (result.stdout && result.stdout.toString().trim()) ||
          "Forge: no product code without a spec in implementing.";
        throw new Error(reason);
      }
    },
  };
};
