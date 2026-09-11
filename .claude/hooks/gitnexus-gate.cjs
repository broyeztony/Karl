#!/usr/bin/env node
/**
 * GitNexus gate — turns the "graph first" rules in CLAUDE.md into an enforced
 * precondition instead of a reminder the agent may skip.
 *
 * Two gates, both emitting a real PreToolUse `deny` decision:
 *
 *   1. Orientation — reading or searching source is denied until the session
 *      has actually consulted the graph (query/context/impact/trace, via MCP or
 *      the CLI).
 *   2. Pre-edit — modifying existing source is denied until `impact` has
 *      returned, per "MUST run impact before editing".
 *
 * Both gates cover Bash as well as the dedicated tools. That matters more than
 * it looks: `Read` on a file and `sed -n 1,400p` on the same file deliver the
 * same bytes to the agent, so gating only the former leaves the deny message
 * recommending a graph call that costs more than the bypass sitting next to it.
 * The Bash arm recognises the common read, search and in-place-write verbs
 * against source paths; it is a real boundary for the obvious forms, not a
 * sandbox, and a determined caller can still write Go through a program the
 * pattern list does not model.
 *
 * It classifies command text, so it cannot tell a command that writes source
 * from one that merely quotes such a command — a heredoc, a test fixture, or a
 * shell one-liner in a commit message trips it too. That direction is the safe
 * one to be wrong in (a deny you can re-issue, not a silent bypass), and
 * GITNEXUS_GATE=off is the escape when the false positive is genuine.
 *
 * Unlock is recorded in PostToolUse, never PreToolUse. PreToolUse fires before
 * the tool runs, so recording there would credit a call that was denied at the
 * permission prompt, errored, or was never more than the word "impact" inside
 * some other command. PostToolUse carries tool_response, so the gate can
 * require that the call actually returned something.
 *
 * State is one empty marker file per flag. Two markers cannot clobber each
 * other, there is no document to parse, and a hook process racing another sees
 * either "file present" or "file absent" — never a half-written JSON blob whose
 * parse failure would silently re-lock a gate the session already paid for.
 *
 * Fails OPEN in every ambiguous case — an unparseable payload, an unindexed
 * repo, a command it cannot classify. A gate that breaks the session when the
 * graph is unavailable would be worse than no gate, so the only thing it ever
 * blocks is an access it can show was premature.
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
const EXEMPT_SEGMENTS = ['.gitnexus/', '.claude/', '.git/', 'node_modules/'];

/** MCP tools that count as having consulted the graph. */
const ORIENTING_MCP = new Set([
  'query', 'context', 'impact', 'trace', 'explain', 'pdg_query', 'cypher',
  'route_map', 'tool_map', 'api_impact', 'detect_changes', 'shape_check',
]);

/** CLI subcommands that count, in argv position after a gitnexus entrypoint. */
const ORIENTING_SUBCOMMAND = new Set([
  'query', 'context', 'impact', 'trace', 'explain', 'pdg-query',
  'detect-changes', 'route-map', 'api-impact',
]);

const IMPACT_SUBCOMMAND = new Set(['impact', 'api-impact']);

/** Runners that can front a gitnexus entrypoint. */
const RUNNERS = new Set(['node', 'npx', 'bunx', 'bun', 'pnpm', 'dlx', 'yarn']);

// --- Bash command classification ------------------------------------------
// Deliberately narrow. Each pattern names a verb that puts file contents in
// front of the agent, or edits a file in place. Anything unrecognised is
// allowed — over-blocking Bash would wedge sessions for no security gain,
// since the gate is about sequencing work, not confining a hostile process.

/** Dumps a file's contents to stdout. */
const READ_VERB = /(^|[\s;|&(])(cat|bat|head|tail|less|more|nl|od|xxd|strings|view)(\s|$)/;
/** `sed -n …p` is the idiomatic ranged read. `sed -i` is handled as a write. */
const SED_READ = /(^|[\s;|&(])sed(\s+--?\w+)*\s+-[a-zA-Z]*n/;
/** Text search — the thing CLAUDE.md says must not substitute for the graph. */
const SEARCH_VERB = /(^|[\s;|&(])(grep|egrep|fgrep|rg|ag|ack)(\s|$)/;
/** In-place edits and patch application. */
const WRITE_VERB = /(^|[\s;|&(])(tee|patch|dd|truncate)(\s|$)/;
const SED_WRITE = /(^|[\s;|&(])sed(\s+--?\w+)*\s+-[a-zA-Z]*i/;
const GIT_WRITE = /(^|[\s;|&(])git\s+(apply|restore|checkout\s+--)(\s|$)/;
/** Shell redirection that creates or appends to a file. */
const REDIRECT_WRITE = /(^|[^0-9>])>>?\s*[^\s|&;]+/;

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

function markerPath(sessionId, flag) {
  const safe = String(sessionId || 'unknown').replace(/[^A-Za-z0-9_-]/g, '_');
  return path.join(STATE_ROOT, `${safe}.${flag}`);
}

function hasFlag(sessionId, flag) {
  try {
    return fs.existsSync(markerPath(sessionId, flag));
  } catch {
    return false;
  }
}

/**
 * Set a flag by creating an empty marker. Concurrent hook processes each create
 * their own file, so there is no read-modify-write window to lose an update in
 * and nothing a torn read could mis-parse.
 */
function setFlag(sessionId, flag) {
  try {
    fs.mkdirSync(STATE_ROOT, { recursive: true });
    fs.writeFileSync(markerPath(sessionId, flag), '');
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
  if (!abs.startsWith(indexRoot + path.sep)) return true; // outside the indexed repo
  const rel = abs.slice(indexRoot.length + 1).split(path.sep).join('/');
  return EXEMPT_SEGMENTS.some((seg) => rel.startsWith(seg) || rel.includes(`/${seg}`));
}

function isSource(filePath) {
  return SOURCE_EXT.has(path.extname(filePath || '').toLowerCase());
}

/** Source-file paths named anywhere in a shell command, minus exempt ones. */
function sourcePathsIn(command, indexRoot) {
  const tokens = String(command).match(/[\w./~@+-]+/g) || [];
  return tokens.filter((t) => isSource(t) && !isExempt(t, indexRoot));
}

/** Split a command line into pipeline/sequence segments. */
function segments(command) {
  return String(command).split(/\|\||&&|[;|\n]/).map((s) => s.trim()).filter(Boolean);
}

/**
 * Does this segment invoke a gitnexus entrypoint, and if so with which
 * subcommand? Matching argv position rather than "the string appears somewhere"
 * is what stops `grep -rn impact .gitnexus/` and `echo "gitnexus impact"` from
 * counting as graph consultation.
 */
function gitnexusSubcommand(segment) {
  const tokens = (String(segment).match(/\S+/g) || [])
    .filter((t) => !/^[A-Za-z_][A-Za-z0-9_]*=/.test(t)) // drop env assignments
    .map((t) => t.replace(/^["']|["']$/g, ''));
  if (!tokens.length) return null;

  let i = 0;
  while (i < tokens.length && RUNNERS.has(path.basename(tokens[i]))) i += 1;

  // The entrypoint must be the command head (after any runner), not an argument.
  let head = tokens[i];
  while (head && head.startsWith('-')) { i += 1; head = tokens[i]; }
  if (!head) return null;
  const base = path.basename(head);
  const isEntry = base === 'gitnexus' || base === 'run.cjs' || base.startsWith('gitnexus@');
  if (!isEntry) return null;

  for (let j = i + 1; j < tokens.length; j += 1) {
    if (tokens[j].startsWith('-')) continue;
    return tokens[j];
  }
  return null;
}

/** Did the tool actually return something? PostToolUse only. */
function looksSuccessful(toolResponse) {
  if (toolResponse === undefined || toolResponse === null) return false;
  if (typeof toolResponse === 'object') {
    if (toolResponse.isError === true) return false;
    if (toolResponse.interrupted === true) return false;
    if (typeof toolResponse.is_error === 'boolean' && toolResponse.is_error) return false;
  }
  return true;
}

// --- PostToolUse: record what actually happened ---------------------------

function recordUnlock(payload) {
  const toolName = String(payload.tool_name || '');
  const sessionId = payload.session_id;
  if (!looksSuccessful(payload.tool_response)) allow();

  if (toolName.startsWith('mcp__gitnexus__')) {
    const op = toolName.slice('mcp__gitnexus__'.length);
    if (ORIENTING_MCP.has(op)) {
      setFlag(sessionId, 'oriented');
      if (op === 'impact' || op === 'api_impact') setFlag(sessionId, 'impacted');
    }
    allow();
  }

  if (toolName === 'Bash') {
    for (const seg of segments((payload.tool_input || {}).command || '')) {
      const sub = gitnexusSubcommand(seg);
      if (sub && ORIENTING_SUBCOMMAND.has(sub)) {
        setFlag(sessionId, 'oriented');
        if (IMPACT_SUBCOMMAND.has(sub)) setFlag(sessionId, 'impacted');
      }
    }
  }
  allow();
}

// --- PreToolUse: enforce ---------------------------------------------------

const ORIENT_HINT = 'Orient first — query({search_query: "<concept>"}) for concepts and flows, '
  + 'context({name: "<symbol>"}) for a named symbol, impact() for blast radius. They return a '
  + 'scoped subgraph instead of whole files, and this gate is paid once per session.';

const IMPACT_HINT = 'Run impact({target: "<symbol>", direction: "upstream"}) — or '
  + '`node .gitnexus/run.cjs impact "<symbol>" --direction upstream --repo .` — and report '
  + 'callers, processes and risk first.\n\nTreat risk: UNKNOWN as unresolved, not as low: an '
  + 'empty caller set can also mean the callers are not resolvable by the index. Confirm with a '
  + 'text search before concluding a symbol is unused.';

function enforce(payload, indexRoot) {
  const toolName = String(payload.tool_name || '');
  const toolInput = payload.tool_input || {};
  const sessionId = payload.session_id;
  const oriented = hasFlag(sessionId, 'oriented');
  const impacted = hasFlag(sessionId, 'impacted');

  if (toolName === 'Read' || toolName === 'Edit' || toolName === 'Write') {
    const target = toolInput.file_path || '';
    if (isExempt(target, indexRoot) || !isSource(target)) allow();
    const abs = path.isAbsolute(target) ? target : path.resolve(indexRoot, target);

    if (toolName === 'Edit' || toolName === 'Write') {
      // Creating a new file has no symbol to analyse and no node in the index,
      // so demanding impact first would only teach that the gate is a formality.
      if (!fs.existsSync(abs)) allow();
      if (!impacted) {
        deny(`GitNexus gate: impact analysis has not run in this session, and CLAUDE.md `
          + `requires it before editing a function, class, or method.\n\n${IMPACT_HINT}\n\n`
          + `Target: ${path.basename(target)}. Set GITNEXUS_GATE=off to bypass.`);
      }
      allow();
    }

    if (!oriented) {
      deny(`GitNexus gate: the graph has not been consulted in this session, so reading `
        + `${path.basename(target)} directly would skip the index this repo maintains.\n\n`
        + `${ORIENT_HINT}\n\nSet GITNEXUS_GATE=off to bypass.`);
    }
    allow();
  }

  if (toolName === 'Grep' || toolName === 'Glob') {
    if (oriented) allow();
    const where = toolInput.path || '';
    if (where && isExempt(where, indexRoot)) allow();
    deny(`GitNexus gate: text search before graph search. CLAUDE.md is explicit — `
      + `"Never substitute grep for graph analysis."\n\nRun query({search_query: "<concept>"}) `
      + `first; it ranks execution flows rather than matching lines. Grep and Glob unlock for `
      + `the rest of the session once the graph has been consulted — they are the right tool `
      + `for literals, and for the empty or UNKNOWN results the graph cannot resolve.\n\n`
      + `Set GITNEXUS_GATE=off to bypass.`);
  }

  if (toolName === 'Bash') {
    const command = String(toolInput.command || '');
    for (const seg of segments(command)) {
      // Never gate the graph tooling itself, or the gate would deadlock.
      if (gitnexusSubcommand(seg)) continue;

      const paths = sourcePathsIn(seg, indexRoot);
      const writes = SED_WRITE.test(seg) || WRITE_VERB.test(seg) || REDIRECT_WRITE.test(seg);
      const reads = READ_VERB.test(seg) || SED_READ.test(seg);

      // `git apply` and `patch` carry their targets inside the patch file, so
      // no source path appears in the command to match on. Gate them on the
      // verb alone: applying a patch is how you edit many files at once, and
      // that is precisely what the pre-edit rule exists for.
      if (GIT_WRITE.test(seg) && !impacted) {
        deny(`GitNexus gate: applying a patch edits whatever it touches, and impact analysis `
          + `has not run in this session. The targets are inside the patch rather than the `
          + `command, so this is gated on the verb.\n\n${IMPACT_HINT}\n\n`
          + `If the patch touches no source, set GITNEXUS_GATE=off to bypass.`);
      }

      if (paths.length && writes && !impacted) {
        deny(`GitNexus gate: this command writes to source (${paths[0]}), and impact analysis `
          + `has not run in this session. Editing through a shell is the same edit — the rule `
          + `in CLAUDE.md does not depend on which tool makes it.\n\n${IMPACT_HINT}\n\n`
          + `Set GITNEXUS_GATE=off to bypass.`);
      }
      if (paths.length && reads && !oriented) {
        deny(`GitNexus gate: this command reads source (${paths[0]}) and the graph has not been `
          + `consulted in this session. \`cat\`/\`sed -n\` on a file delivers the same bytes as `
          + `Read, so it is gated the same way.\n\n${ORIENT_HINT}\n\n`
          + `Set GITNEXUS_GATE=off to bypass.`);
      }
      if (SEARCH_VERB.test(seg) && !oriented) {
        deny(`GitNexus gate: text search before graph search, via the shell. CLAUDE.md is `
          + `explicit — "Never substitute grep for graph analysis."\n\n${ORIENT_HINT}\n\n`
          + `Shell search unlocks with everything else once the graph has been consulted. `
          + `Set GITNEXUS_GATE=off to bypass.`);
      }
    }
    allow();
  }

  allow();
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

  const indexRoot = findIndexRoot(payload.cwd || process.cwd());
  if (!indexRoot) allow(); // repo carries no GitNexus index

  if (payload.hook_event_name === 'PostToolUse') recordUnlock(payload);
  enforce(payload, indexRoot);
}

main();
