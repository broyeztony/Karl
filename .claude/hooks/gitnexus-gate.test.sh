#!/bin/bash
# Exercises every finding raised against the gate. Run from the repo root.
REPO="$1"
H="$REPO/.claude/hooks/gitnexus-gate.cjs"
EXT=".go"                      # built at runtime so this file's own text carries no source path
SRC="parser/parser$EXT"
SRC2="interpreter/interpreter$EXT"
pass=0; fail=0

pre() { # session, tool, input-json  -> prints deny reason or nothing
  printf '{"hook_event_name":"PreToolUse","session_id":"%s","tool_name":"%s","cwd":"%s","tool_input":%s}' \
    "$1" "$2" "$REPO" "$3" | node "$H"
}
post() {
  printf '{"hook_event_name":"PostToolUse","session_id":"%s","tool_name":"%s","cwd":"%s","tool_input":%s,"tool_response":%s}' \
    "$1" "$2" "$REPO" "$3" "$4" | node "$H"
}
expect() { # label, expected(DENY|ALLOW), actual-output
  local got="ALLOW"; [ -n "$3" ] && got="DENY"
  if [ "$got" = "$2" ]; then pass=$((pass+1)); printf '  ok    %-58s %s\n' "$1" "$got"
  else fail=$((fail+1)); printf '  FAIL  %-58s got %s want %s\n' "$1" "$got" "$2"; fi
}

echo "### Finding 1 — Bash bypass of the read gate"
S="s1-$$"
expect "cat SRC"              DENY "$(pre $S Bash "{\"command\":\"cat $SRC\"}")"
expect "sed -n range SRC"     DENY "$(pre $S Bash "{\"command\":\"sed -n '1,400p' $SRC2\"}")"
expect "head SRC"             DENY "$(pre $S Bash "{\"command\":\"head -50 $SRC\"}")"
expect "tail SRC"             DENY "$(pre $S Bash "{\"command\":\"tail -n 20 $SRC\"}")"
expect "grep (search)"        DENY "$(pre $S Bash '{"command":"grep -rn mutex ."}')"
expect "rg (search)"          DENY "$(pre $S Bash '{"command":"rg goroutine"}')"

echo "### Finding 1 — Bash bypass of the write gate"
expect "sed -i SRC"           DENY "$(pre $S Bash "{\"command\":\"sed -i 's/a/b/' $SRC\"}")"
expect "heredoc onto SRC"     DENY "$(pre $S Bash "{\"command\":\"cat > $SRC <<EOF\"}")"
expect "tee SRC"              DENY "$(pre $S Bash "{\"command\":\"tee $SRC\"}")"
expect "git apply"            DENY  "$(pre $S Bash '{"command":"git apply /tmp/p.diff"}')"

echo "### Finding 1 — must NOT over-block ordinary commands"
expect "go build"             ALLOW "$(pre $S Bash '{"command":"go build ./..."}')"
expect "go test"              ALLOW "$(pre $S Bash '{"command":"go test ./..."}')"
expect "git status"           ALLOW "$(pre $S Bash '{"command":"git status"}')"
expect "ls"                   ALLOW "$(pre $S Bash '{"command":"ls -la"}')"
expect "cat a markdown file"  ALLOW "$(pre $S Bash '{"command":"cat README.md"}')"
expect "gitnexus CLI itself"  ALLOW "$(pre $S Bash '{"command":"node .gitnexus/run.cjs query \"x\" --repo ."}')"

echo "### Finding 2 — unlock records intent, not execution"
S2="s2-$$"
post $S2 Bash '{"command":"grep -rn impact .gitnexus/"}' '{"stdout":"x"}' >/dev/null
expect "grep ... impact does NOT orient"   DENY "$(pre $S2 Bash "{\"command\":\"cat $SRC\"}")"
S3="s3-$$"
post $S3 Bash '{"command":"echo \"gitnexus impact\""}' '{"stdout":"x"}' >/dev/null
expect "echo gitnexus impact does NOT orient" DENY "$(pre $S3 Bash "{\"command\":\"cat $SRC\"}")"
S4="s4-$$"
post $S4 mcp__gitnexus__impact '{}' '{"isError":true}' >/dev/null
expect "errored impact does NOT unlock"    DENY "$(pre $S4 Edit "{\"file_path\":\"$REPO/$SRC\"}")"
S5="s5-$$"
post $S5 mcp__gitnexus__impact '{}' 'null' >/dev/null
expect "null response does NOT unlock"     DENY "$(pre $S5 Edit "{\"file_path\":\"$REPO/$SRC\"}")"

echo "### Real unlock still works"
S6="s6-$$"
expect "Read denied before orienting"      DENY  "$(pre $S6 Read "{\"file_path\":\"$REPO/$SRC\"}")"
post $S6 mcp__gitnexus__query '{}' '{"processes":[]}' >/dev/null
expect "Read allowed after real query"     ALLOW "$(pre $S6 Read "{\"file_path\":\"$REPO/$SRC\"}")"
expect "Bash cat allowed after query"      ALLOW "$(pre $S6 Bash "{\"command\":\"cat $SRC\"}")"
expect "grep allowed after query"          ALLOW "$(pre $S6 Bash '{"command":"grep -rn mutex ."}')"
expect "Edit still denied (no impact)"     DENY  "$(pre $S6 Edit "{\"file_path\":\"$REPO/$SRC\"}")"
post $S6 Bash '{"command":"node .gitnexus/run.cjs impact Foo --repo ."}' '{"stdout":"ok"}' >/dev/null
expect "Edit allowed after CLI impact"     ALLOW "$(pre $S6 Edit "{\"file_path\":\"$REPO/$SRC\"}")"
expect "sed -i allowed after impact"       ALLOW "$(pre $S6 Bash "{\"command\":\"sed -i 's/a/b/' $SRC\"}")"
expect "git apply allowed after impact"    ALLOW "$(pre $S6 Bash '{"command":"git apply /tmp/p.diff"}')"

echo "### Finding 5 — creating a new source file"
S7="s7-$$"
expect "Write to NEW path allowed"         ALLOW "$(pre $S7 Write "{\"file_path\":\"$REPO/evaluator/brand_new$EXT\"}")"
expect "Write to EXISTING path denied"     DENY  "$(pre $S7 Write "{\"file_path\":\"$REPO/$SRC\"}")"

echo "### Finding 3 — concurrent unlocks do not clobber each other"
S9="s9-$$"
post $S9 mcp__gitnexus__impact '{}' '{"ok":1}' >/dev/null &
post $S9 mcp__gitnexus__query  '{}' '{"ok":1}' >/dev/null &
wait
expect "parallel impact+query: Edit allowed"  ALLOW "$(pre $S9 Edit "{\"file_path\":\"$REPO/$SRC\"}")"
expect "parallel impact+query: Read allowed"  ALLOW "$(pre $S9 Read "{\"file_path\":\"$REPO/$SRC\"}")"

echo "### Fail-open paths"
S8="s8-$$"
expect "malformed payload"                 ALLOW "$(echo 'not json' | node "$H")"
expect "empty payload"                     ALLOW "$(printf '' | node "$H")"
expect "unindexed repo"                    ALLOW "$(pre $S8 Read '{"file_path":"/tmp/x.go"}' )"
expect "file outside repo"                 ALLOW "$(pre $S8 Read '{"file_path":"/etc/hosts"}')"
expect "markdown file"                     ALLOW "$(pre $S8 Read "{\"file_path\":\"$REPO/README.md\"}")"
expect ".claude exempt"                    ALLOW "$(pre $S8 Read "{\"file_path\":\"$REPO/.claude/hooks/gitnexus-gate.cjs\"}")"
expect "GITNEXUS_GATE=off"                 ALLOW "$(printf '{"hook_event_name":"PreToolUse","session_id":"x","tool_name":"Read","cwd":"%s","tool_input":{"file_path":"%s/%s"}}' "$REPO" "$REPO" "$SRC" | GITNEXUS_GATE=off node "$H")"

echo
echo "passed=$pass failed=$fail"
[ "$fail" -eq 0 ]
