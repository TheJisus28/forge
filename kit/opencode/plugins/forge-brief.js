// Forge brief for opencode.
//
// opencode has no SessionStart hook, so this plugin does the job of the
// Claude Code hook: it appends `forge brief` to the system prompt, so every
// request carries the current Forge state, at session start and after every
// compaction. It also feeds the brief into the compaction prompt, so the
// summary keeps that state.
//
// The text comes from the Go binary; this file is only an adapter. A
// repository without Forge, or a machine without `forge`, injects nothing.
// Delete the file to turn it off.

const HEADING = "Forge brief:";

export const ForgeBrief = async ({ $, directory }) => {
  async function brief() {
    const result = await $`forge brief`.cwd(directory).nothrow().quiet();
    if (result.exitCode !== 0) return "";
    return result.stdout ? result.stdout.toString().trim() : "";
  }

  return {
    "experimental.chat.system.transform": async (input, output) => {
      const text = await brief();
      if (!text) return;
      // Replace our block so state changes and compactions refresh it
      // instead of piling up, and leave the static instructions first.
      for (let i = output.system.length - 1; i >= 0; i--) {
        if (output.system[i].startsWith(HEADING)) output.system.splice(i, 1);
      }
      output.system.push(HEADING + "\n" + text);
    },
    "experimental.session.compacting": async (input, output) => {
      const text = await brief();
      if (text) output.context.push(text);
    },
  };
};
