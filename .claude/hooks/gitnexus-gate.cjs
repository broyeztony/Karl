#!/usr/bin/env node
/**
 * GitNexus gate — turns the "graph first" rules in CLAUDE.md into an enforced
 * precondition instead of a reminder the agent may skip.
 *
 * Two gates, both emitting a real PreToolUse `deny` decision:
 *
 *   1. Orientation — Read/Grep/Glob against source files are denied until the
 *      session has consulted the graph at least once (query/context/impact/
 *      trace/explain, via MCP or the CLI).
 *   2. Pre-edit — Edit/Write against source files are denied until `impact` has
 *      run, matching "MUST run impact before editing" in CLAUDE.md.
 *
 * The companion gitnexus-hook.cjs only ever emits `additionalContext`, which an
 * agent is free to read past. This file is the half that can say no.
 *
 * Satisfying a gate is recorded per session, so orientation is paid once and
 * not re-litigated on every subsequent read.
 *
 * Fails OPEN in every ambiguous case — an unparseable payload, an unindexed
 * repo, a state dir it cannot write. A gate that breaks the session when the
 * graph is unavailable would be worse than no gate, so the only thing it ever
 * blocks is a read or edit it can prove was premature.
 *
 * Bypass: GITNEXUS_GATE=off
 */

'use strict';

const fs = require('fs');
const os = require('os');
const path = require('path');

const STATE_ROOT = path.join(os.tmpdir(), 'gitnexus-gate');

/** Source files. Prose, config and lockfiles are never gated. */
const SOURCE_EXT = new Set([
  '.go', '.ts', '.tsx', '.js', '.jsx', '.mjs', '.cjs', '.py', '.java', '.kt',
  '.kts', '.rb', '.rs', '.c', '.h', '.cc', '.cpp', '.hpp', '.cs', '.php',
  '.swift', '.dart', '.vue', '.scala', '.zig', '.cob', '.cbl',
]);

/**
 * Paths exempt from both gates. Tooling and metadata directories: reading them
 * is how the agent inspects the gate itself, and gating that invites a deadlock.
 */
const EXEMPT_SEGMENTS = ['.gitnexus/', '.claude/', '.git/', 'node_modules/', 'graphify-out/'];

/** MCP tools that count as having consulted the graph. */
const ORIENTING_MCP = new Set([
  'query', 'context', 'impact', 'trace', 'explain', 'pdg_query', 'cypher',
  'route_map', 'tool_map', 'api_impact', 'detect_changes', 'shape_check',
]);

/** CLI subcommands that count, when run through any gitnexus entrypoint. */
const ORIENTING_CLI = /\b(query|context|impact|trace|explain|pdg-query|detect-changes|route-map)\b/;

function readStdin() {
  try {
    return fs.readFileSync(0, 'utf8');
  } catch {
    return '';
  }
}

/** Nearest ancestor holding a .gitnexus index, or null when unindexed. */
function findIndexRoot(startDir) {
  let dir = startDir;
  for (let i = 0; i < 40; i += 1) {
    if (fs.existsSync(path.join(dir, '.gitnexus', 'meta.json'))) return dir;
    const parent = path.dirname(dir);
    if (parent === dir) return null;
    dir = parent;
  }
  return null;
}

function stateFile(sessionId) {
  const safe = String(sessionId || 'unknown').replace(/[^A-Za-z0-9_-]/g, '_');
  return path.join(STATE_ROOT, `${safe}.json`);
}

function loadState(sessionId) {
  try {
    return JSON.parse(fs.readFileSync(stateFile(sessionId), 'utf8'));
  } catch {
    return { oriented: false, impacted: false };
  }
}

function saveState(sessionId, state) {
  try {
    fs.mkdirSync(STATE_ROOT, { recursive: true });
    fs.writeFileSync(stateFile(sessionId), JSON.stringify(state));
  } catch {
    /* fail open: an unwritable state dir must not block the session */
  }
}

function deny(reason) {
  process.stdout.write(JSON.stringify({
    hookSpecificOutput: {
      hookEventName: 'PreToolUse',
      permissionDecision: 'deny',
      permissionDecisionReason: reason,
    },
  }));
  process.exit(0);
}

/** Allowing is the silent default — emit nothing and let normal permissions apply. */
function allow() {
  process.exit(0);
}

function isExempt(filePath, indexRoot) {
  if (!filePath) return true;
  const abs = path.isAbsolute(filePath) ? filePath : path.resolve(indexRoot, filePath);
  // Outside the indexed repo the graph has nothing to say.
  if (!abs.startsWith(indexRoot + path.sep)) return true;
  const rel = abs.slice(indexRoot.length + 1).split(path.sep).join('/');
  return EXEMPT_SEGMENTS.some((seg) => rel.startsWith(seg) || rel.includes(`/${seg}`));
}

function isSource(filePath) {
  return SOURCE_EXT.has(path.extname(filePath || '').toLowerCase());
}

function main() {
  if ((process.env.GITNEXUS_GATE || '').toLowerCase() === 'off') allow();

  let payload;
  try {
    payload = JSON.parse(readStdin());
  } catch {
    allow();
  }
  if (!payload || typeof payload !== 'object') allow();

  const toolName = String(payload.tool_name || '');
  const toolInput = payload.tool_input || {};
  const sessionId = payload.session_id;
  const cwd = payload.cwd || process.cwd();

  const indexRoot = findIndexRoot(cwd);
  if (!indexRoot) allow(); // repo carries no GitNexus index

  const state = loadState(sessionId);

  // --- Record graph consultation ------------------------------------------
  if (toolName.startsWith('mcp__gitnexus__')) {
    const op = toolName.slice('mcp__gitnexus__'.length);
    if (ORIENTING_MCP.has(op)) {
      state.oriented = true;
      if (op === 'impact' || op === 'api_impact') state.impacted = true;
      saveState(sessionId, state);
    }
    allow();
  }

  if (toolName === 'Bash') {
    const cmd = String(toolInput.command || '');
    if (/gitnexus/.test(cmd) && ORIENTING_CLI.test(cmd)) {
      state.oriented = true;
      if (/\bimpact\b/.test(cmd)) state.impacted = true;
      saveState(sessionId, state);
    }
    allow(); // Bash itself is never gated — gitnexus-hook.cjs already advises it
  }

  // --- Enforce ------------------------------------------------------------
  if (toolName === 'Read' || toolName === 'Edit' || toolName === 'Write' || toolName === 'NotebookEdit') {
    const target = toolInput.file_path || toolInput.notebook_path || '';
    if (isExempt(target, indexRoot) || !isSource(target)) allow();

    if ((toolName === 'Edit' || toolName === 'Write' || toolName === 'NotebookEdit') && !state.impacted) {
      deny(
        `GitNexus gate: impact analysis has not run in this session, and CLAUDE.md requires it `
        + `before editing a function, class, or method.\n\n`
        + `Run impact({target: "<symbol>", direction: "upstream"}) — or `
        + `\`node .gitnexus/run.cjs impact "<symbol>" --direction upstream --repo .\` — `
        + `and report callers, processes, and risk before editing ${path.basename(target)}.\n\n`
        + `Treat risk: UNKNOWN as unresolved, not as low: an empty caller set can also mean the `
        + `callers are not resolvable by the index. Confirm with a text search before concluding a `
        + `symbol is unused.\n\n`
        + `Set GITNEXUS_GATE=off to bypass.`,
      );
    }

    if (toolName === 'Read' && !state.oriented) {
      deny(
        `GitNexus gate: the graph has not been consulted in this session, so reading `
        + `${path.basename(target)} directly would skip the index this repo maintains.\n\n`
        + `Orient first — query({search_query: "<concept>"}) for concepts and flows, `
        + `context({name: "<symbol>"}) for a named symbol, impact() for blast radius. `
        + `They return a scoped subgraph instead of whole files.\n\n`
        + `Then read the specific files they point at; this gate is paid once per session.\n\n`
        + `Set GITNEXUS_GATE=off to bypass.`,
      );
    }
    allow();
  }

  if (toolName === 'Grep' || toolName === 'Glob') {
    if (state.oriented) allow();
    const where = toolInput.path || '';
    if (where && isExempt(where, indexRoot)) allow();
    deny(
      `GitNexus gate: text search before graph search. CLAUDE.md is explicit — `
      + `"Never substitute grep for graph analysis."\n\n`
      + `Run query({search_query: "<concept>"}) first; it ranks execution flows rather than `
      + `matching lines. Grep and Glob unlock for the rest of the session once the graph has `
      + `been consulted — they are the right tool for literals, and for the empty or `
      + `UNKNOWN results the graph cannot resolve.\n\n`
      + `Set GITNEXUS_GATE=off to bypass.`,
    );
  }

  allow();
}

main();
