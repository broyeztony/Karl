# Graph Report - Karl  (2026-09-09)

## Corpus Check
- 226 files · ~207,496 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2528 nodes · 6165 edges · 157 communities (118 shown, 21 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 836 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `49f37533`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- testing.T
- Token
- value_stream_algebra.go
- server
- Server
- Parser
- builtins_process.go
- channelSendBlocking
- Tree
- evalInput
- DebugController
- Value
- recoverableError
- wasm_exec.js
- builtins_stream_pipeline.go
- Environment
- server_test.go
- RegisterBuiltins
- NTree
- streamIterator
- convert.go
- ValueType
- taskAwaitWithCancel
- StreamReader
- runtimeState
- New
- package.json
- process_api_test.go
- Karl Language
- Concurrency Expressions (&, spawn, !&, race)
- main
- Process
- io.Writer
- repository
- Recoverable Errors (`?`) Runtime Semantics
- Identifier
- builtins_math.go
- main.go
- evalWithConfiguredEvaluator
- .playground/main.k Demo Program
- Node
- stringArg
- Kernel
- builtins_runtime_system.go
- karl loom CLI Command
- Karl Tooling Naming System
- Evaluator
- main_test.go
- Karl Language
- .evalInfixExpression
- streamPartitionRouter
- Karl Language Specification
- builtins_sql.go
- value.go
- FormatRuntimeError
- builtins_time.go
- parseProgram
- Channel
- engine.k (core workflow engine)
- builtinSignalWatch
- SQLTx
- io.Reader
- Lexer
- Expression
- .knb Notebook JSON Format
- Karl Language Support (VS Code Extension)
- newModuleState
- teeStreamIterator
- bindPattern
- Ordered Index Tree tree(kind?)
- builtinUUIDParse
- connect (WASM worker bootstrap)
- Karl Notebook System
- Karl Process API Examples
- SliceExpression
- builtinHTTPServe
- .Pretty
- Builtin
- karl-health-demo Namespace Fixture
- sheets/worker.js
- mapKeyForValue
- inspectObjectPairs
- Karl REPL (karl loom)
- Expression-Based Language Design
- String
- builtinFromJSON
- traceCommand
- Karl API-First Stdlib Skill
- Karl Logo (playground asset)
- playground/worker.js
- <process> Value Type
- runtimeState
- build-and-test Job
- Karl Project Logo
- Karl Standard Library Reference
- Signal
- buildRange
- Karl Jupyter Kernel
- Karl VS Code Extension Icon
- evalWithPolicy
- Karl CodeMirror Syntax Mode
- Karl Stream Examples
- formatLogValue
- sortRows
- extension.js
- test_debugger_cli_e2e.sh
- Deterministic Example Corpus (examples_diff.txt)
- Concurrency in Karl (Tasks, Failures, Recovery)
- Browser Tab Icon Role for Web Playground
- ImportExpression
- Karl Kernel Favicon (32x32 K mark)
- Karl Kernel Logo (64x64 PNG)
- newTTYLineWriter
- run_all_tests.sh
- builtinFromBase58
- Server
- diff_examples_against_ref.sh
- Import Factory and Live Module Object
- Karl Landing Page
- install_skill_sync_hooks.sh
- run_examples_runtime.sh
- Karl Reactive Spreadsheet (WASM worker)
- .resolveImportPath
- karl-remote.sh
- karl-repl.sh
- install.sh
- sync_skills_to_codex.sh
- Query Expression Execution
- test_examples.sh
- test_repl.sh
- One log() per API Call Demo Rule
- Codex Skill Sync Requirement
- applyTheme
- initGrid
- karl
- Set Collection
- Pattern
- vigil_builtins_test.go
- graphify

## God Nodes (most connected - your core abstractions)
1. `Value` - 416 edges
2. `evalInput()` - 99 edges
3. `mustEval()` - 85 edges
4. `assertString()` - 79 edges
5. `Parser` - 77 edges
6. `Expression` - 74 edges
7. `Environment` - 70 edges
8. `recoverableError()` - 65 edges
9. `Signal` - 61 edges
10. `assertInteger()` - 59 edges

## Surprising Connections (you probably didn't know these)
- `Fan-out / Fan-in Pattern` --semantically_similar_to--> `Live Matrix Loop (one-statement fan-in)`  [INFERRED] [semantically similar]
  examples/contrib/workflow/README.md → assets/playground/docs/idiomatic/index.html
- `Deterministic Example Corpus (examples_diff.txt)` --semantically_similar_to--> `Deterministic Semantics and Tie-Break Rules`  [INFERRED] [semantically similar]
  tests/corpus/examples_diff.txt → skills/karl-api-first-stdlib/SKILL.md
- `loom (runtime)` --conceptually_related_to--> `karl loom (REPL)`  [AMBIGUOUS]
  SPECS/tooling_naming.md → README.md
- `Release v0.5.0 (notebook + Jupyter kernel)` --references--> `Karl Tooling Naming System`  [INFERRED]
  CHANGELOG.md → SPECS/tooling_naming.md
- `Notebook State Persistence Across Cells` --semantically_similar_to--> `REPL Persistent Environment`  [INFERRED] [semantically similar]
  notebook/README.md → repl/README.md

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Logo lockup: framed K monogram stacked over the KARL wordmark** — assets_playground_karl_logo, assets_playground_karl_square_frame, assets_playground_karl_monogram_k, assets_playground_karl_wordmark [EXTRACTED 1.00]
- **map/reduce pipeline: nums -> squares -> sum printed via log** — assets_playground_nums, assets_playground_map, assets_playground_squares, assets_playground_reduce, assets_playground_add, assets_playground_log_builtin, assets_playground_console_output [EXTRACTED 1.00]
- **Concurrent Task Join Demo** — assets_repl_zero_arg_closure, assets_repl_sleep_builtin, assets_repl_wait_ampersand_concurrency, assets_repl_array_literal_output [EXTRACTED 1.00]
- **Range to Grade Data Pipeline Demo** — assets_repl_range_literal, assets_repl_filter_map_reduce_pipeline, assets_repl_lambda_arrow_syntax, assets_repl_match_expression, assets_repl_array_literal_output [EXTRACTED 1.00]
- **Debugger Engine Components Across Milestones 1-3** — specs_debugger_execution_hooks, specs_debugger_frame_tracking, specs_debugger_breakpoint_step_engine, specs_debugger_cli_command, specs_debugger_dap_bridge, specs_debugger_frame_expression_eval [EXTRACTED 1.00]
- **Karl Documentation Site Surface** — assets_playground_docs_index_karl_documentation_hub, assets_playground_docs_specs_index_karl_language_specification, assets_playground_docs_std_index_standard_library_reference, assets_playground_docs_idiomatic_index_idiomatic_karl_for_devops, assets_playground_bench_index_karl_bench_playground, assets_playground_index_karl_landing_page [EXTRACTED 1.00]
- **Karl's Four Primary Abstractions** — agents_stream_abstraction, agents_task_abstraction, agents_channel_abstraction, agents_sink_abstraction, agents_design_philosophy [EXTRACTED 1.00]
- **karl-health-demo Kubernetes Demo Fixture** — examples_features_streams_k8s_00_namespace_karl_health_demo, examples_features_streams_k8s_10_web_ok_deployment_web_ok, examples_features_streams_k8s_11_crashloop_app_deployment_crashloop_app, examples_features_streams_k8s_12_pending_app_deployment_pending_app, examples_features_streams_k8s_20_chain_node_deployment_chain_node, examples_features_streams_readme_karl_stream_examples [EXTRACTED 1.00]
- **Karl Stdlib API Design Gates** — skills_karl_api_first_stdlib_skill_api_over_userland_algorithms, skills_karl_api_first_stdlib_skill_complexity_contract, skills_karl_api_first_stdlib_skill_deterministic_semantics, skills_karl_api_first_stdlib_skill_ai_ergonomics, skills_karl_api_first_stdlib_references_api_review_checklist_release_readiness [EXTRACTED 1.00]
- **Stream Pipeline Execution Flow (source -> stage -> sink under backpressure)** — specs_stream_sources, specs_stream_stages, specs_stream_sinks, specs_stream_execution_model, specs_stream_backpressure, specs_stream_pipeline_operator [EXTRACTED 1.00]
- **Workflow Engine Resilience Stack** — examples_contrib_workflow_readme_engine, examples_contrib_workflow_readme_retry_policy, examples_contrib_workflow_readme_parallel_executor, examples_contrib_workflow_readme_storage, examples_contrib_workflow_readme_circuit_breaker, examples_contrib_workflow_readme_persisted_dag_state [EXTRACTED 1.00]
- **Async fetch, recoverable decode, guarded classification pipeline** — assets_vscode_async_task_wait, assets_vscode_recoverable_block, assets_vscode_decodejson, assets_vscode_match_with_guard [EXTRACTED 1.00]
- **Logo composition: framed monogram over wordmark** — assets_karl_logo, assets_karl_square_frame, assets_karl_k_monogram, assets_karl_wordmark [EXTRACTED 1.00]
- **Module factory import wiring main.k to helpers.k bindings** — assets_vscode_import_expression, assets_vscode_makeutil, assets_vscode_playground_helpers_k, assets_vscode_compose, assets_vscode_scale [EXTRACTED 1.00]
- **Edit-run-observe loop: editor pane, Run/Auto Run controls, runtime, output console** — assets_playground_editor_pane, assets_playground_run_button, assets_playground_auto_run_toggle, assets_playground_wasm_runtime, assets_playground_output_console, assets_playground_ready_status [INFERRED 0.85]
- **REPL Session Shell UX** — assets_repl_ascii_art_banner, assets_repl_cli_version_string, assets_repl_meta_commands, assets_repl_prompt_line, assets_repl_multiline_continuation, assets_repl_quit_goodbye [INFERRED 0.85]
- **Karl Interactive Execution Surfaces (REPL, notebook, Jupyter kernel, playground)** — repl_readme_karl_repl, notebook_readme_karl_notebook_system, kernel_readme_karl_jupyter_kernel, playground_readme_karl_playground, repl_remote_karl_remote_repl [INFERRED 0.85]
- **Persistent-Environment Execution Pattern Shared Across Surfaces** — repl_readme_persistent_environment, notebook_readme_state_persistence, kernel_readme_state_persistence, repl_remote_isolated_sessions [INFERRED 0.85]
- **Karl kernel brand mark: bold K monogram in a monochrome square badge, sized as a 64x64 icon asset** — kernel_logo_64x64_logo_image, kernel_logo_64x64_monogram_k, kernel_logo_64x64_monochrome_square_badge, kernel_logo_64x64_favicon_asset [INFERRED 0.85]
- **Karl logo lockup: boxed K monogram over letterspaced KARL wordmark in flat monochrome** — plugins_karl_vscode_images_karl_icon_icon, plugins_karl_vscode_images_karl_icon_k_monogram_mark, plugins_karl_vscode_images_karl_icon_karl_wordmark, plugins_karl_vscode_images_karl_icon_monochrome_minimal_design_choice, plugins_karl_vscode_images_karl_icon_karl_brand_identity [INFERRED 0.85]
- **Everything-is-an-expression control flow surface** — assets_vscode_for_with_then, assets_vscode_if_else_expression, assets_vscode_match_with_guard, assets_vscode_program_return_array [INFERRED 0.85]
- **Low-resolution favicon legibility strategy: single-letter monogram + heavy stroke + square frame at 32x32** — assets_playground_logo_32x32_favicon, assets_playground_logo_32x32_monogram_k, assets_playground_logo_32x32_minimal_monochrome_design, assets_playground_logo_32x32_browser_tab_icon_role [INFERRED 0.85]
- **Small-size branding asset pattern: a single bold K glyph in a bordered monochrome square, legible at 32x32 for tab/kernel icon use** — kernel_logo_32x32_favicon, kernel_logo_32x32_k_monogram, kernel_logo_32x32_monochrome_square_badge, kernel_logo_32x32_browser_tab_icon_asset [INFERRED 0.85]

## Communities (157 total, 21 thin omitted)

### Community 0 - "testing.T"
Cohesion: 0.07
Nodes (92): testing.T, NewEvaluatorWithSourceFilenameAndRoot(), assertEquivalent(), assertFloat(), assertInteger(), captureStdout(), mustEval(), TestEvalArithmetic() (+84 more)

### Community 1 - "Token"
Cohesion: 0.03
Nodes (29): ArrayLiteral, AssignExpression, AwaitExpression, BreakExpression, CallExpression, IndexExpression, InfixExpression, MatchExpression (+21 more)

### Community 2 - "value_stream_algebra.go"
Cohesion: 0.06
Nodes (41): streamChunkIterator, streamDistinctIterator, streamDropIterator, streamFilterIterator, streamFlatMapIterator, streamFromUTF8Iterator, streamMapIterator, StreamSinkValue (+33 more)

### Community 3 - "server"
Cohesion: 0.08
Nodes (27): evaluateArgs, event, launchArgs, request, response, scopesArgs, server, setBreakpointsArgs (+19 more)

### Community 4 - "Server"
Cohesion: 0.06
Nodes (37): github.com/gorilla/websocket.Conn, net/http.Request, net/http.ResponseWriter, sync.RWMutex, syscall/js.Value, sheetCommand, sheetCommandResult, CellID (+29 more)

### Community 5 - "Parser"
Cohesion: 0.13
Nodes (7): infixParseFn, Parser, isAssignable(), statementIsNil(), prefixParseFn, TokenType, LookupIdent()

### Community 6 - "builtins_process.go"
Cohesion: 0.18
Nodes (23): io.ReadCloser, builtinProc(), builtinRun(), collectProcessStream(), executeRunSpec(), Evaluator, processSpec, processStageSpec (+15 more)

### Community 7 - "channelSendBlocking"
Cohesion: 0.17
Nodes (16): builtinDone(), builtinRecv(), builtinSend(), channelSendBlocking(), Evaluator, isTopLevelRuntimeDeadlocked(), builtinExit(), builtinFail() (+8 more)

### Community 8 - "Tree"
Cohesion: 0.12
Nodes (25): avlDelete(), avlInsert(), balanceFactor(), compareTreeKey(), findTreeNode(), Tree, inOrderTree(), minTreeNode() (+17 more)

### Community 9 - "evalInput"
Cohesion: 0.06
Nodes (63): assertString(), evalInput(), TestChannelDeadlockRecvReturnsRuntimeError(), TestChannelDeadlockSendReturnsRuntimeError(), TestEvalArrayIndexStillRequiresInteger(), TestEvalDecodeJSONOverflow(), TestEvalDivisionByZeroCompoundAssign(), TestEvalDivisionByZeroFloat() (+55 more)

### Community 10 - "DebugController"
Cohesion: 0.08
Nodes (11): sync.Cond, DebugController, DebugStopReason, IsDebugTerminated(), normalizeTaskID(), DebugBreakpoint, DebugEvent, DebugFrame (+3 more)

### Community 11 - "Value"
Cohesion: 0.14
Nodes (33): isCallable(), builtinChunk(), builtinCount(), builtinDistinct(), builtinDrop(), builtinFilter(), builtinFind(), builtinFlatMap() (+25 more)

### Community 12 - "recoverableError"
Cohesion: 0.08
Nodes (29): builtinAppendFile(), builtinDeleteFile(), builtinReadFile(), builtinWriteFile(), Evaluator, builtinExists(), builtinListDir(), Evaluator (+21 more)

### Community 13 - "wasm_exec.js"
Cohesion: 0.06
Nodes (5): constructor(), _makeFuncWrapper(), _resume(), write(), writeSync()

### Community 14 - "builtins_stream_pipeline.go"
Cohesion: 0.14
Nodes (22): builtinCollectSink(), builtinDebounceStage(), builtinExecSink(), builtinFromChannelSource(), builtinJoinSource(), builtinLinesStage(), builtinMergeSource(), builtinSpillStage() (+14 more)

### Community 15 - "Environment"
Cohesion: 0.12
Nodes (13): Environment, NewEnclosedEnvironment(), NewEnvironment(), Evaluator, errorValue(), Evaluator, Evaluator, Evaluator (+5 more)

### Community 16 - "server_test.go"
Cohesion: 0.25
Nodes (29): testClient, bodyArrayOfMaps(), bodyMap(), canonicalPath(), currentTopFrameLine(), currentTopFramePath(), findVarByName(), intFromAny() (+21 more)

### Community 17 - "RegisterBuiltins"
Cohesion: 0.07
Nodes (25): bindReceiver(), builtinSHA256(), Evaluator, registerCryptoBuiltins(), builtinBytesJoin(), builtinFromUtf8(), builtinToUtf8(), Evaluator (+17 more)

### Community 18 - "NTree"
Cohesion: 0.11
Nodes (15): builtinLen(), builtinNTree(), builtinTree(), Evaluator, registerCollectionBuiltins(), NTree, indexOfID(), insertIDAt() (+7 more)

### Community 19 - "streamIterator"
Cohesion: 0.11
Nodes (26): debounceEvent, linesIterator, streamCursorState, streamDebounceIterator, streamIterator, StreamPlanValue, streamReaderIterator, streamSinkPlanRunFunc (+18 more)

### Community 20 - "convert.go"
Cohesion: 0.24
Nodes (12): convertCommand(), notebookCommand(), notebookUsage(), ConvertCommand(), formatFromFilename(), jupyterSourceToString(), loadIPYNB(), saveIPYNB() (+4 more)

### Community 21 - "ValueType"
Cohesion: 0.08
Nodes (9): Boolean, Bytes, Char, Float, Integer, Null, treeDistanceValue(), Unit (+1 more)

### Community 22 - "taskAwaitWithCancel"
Cohesion: 0.11
Nodes (15): builtinThen(), Evaluator, registerAsyncBuiltins(), Evaluator, Evaluator, Task, runtimeState, panicToError() (+7 more)

### Community 23 - "StreamReader"
Cohesion: 0.09
Nodes (5): io.Closer, Process, StreamReader, StreamWriter, streamReadEnded()

### Community 24 - "runtimeState"
Cohesion: 0.11
Nodes (5): cloneStrings(), runtimeState, Task, makeEnvMap(), trimLineEnding()

### Community 25 - "New"
Cohesion: 0.06
Nodes (58): Program, parseProgram(), testing.B, NewBaseEnvironment(), NewDebugController(), EvalDebugExpression(), NewEvaluatorWithSourceAndFilename(), parseImportProgram() (+50 more)

### Community 26 - "package.json"
Cohesion: 0.07
Nodes (27): activationEvents, categories, contributes, breakpoints, debuggers, grammars, languages, description (+19 more)

### Community 27 - "process_api_test.go"
Cohesion: 0.18
Nodes (24): assertBoolean(), assertSamePath(), mustExecutable(), TestDecodeUtf8InvalidBytesRecoverable(), TestProcBytesReadLoopThenWait(), TestProcessAPIHelperProcess(), TestProcessPipeOperatorRuntimeError(), TestProcModeConstantsWork() (+16 more)

### Community 28 - "Karl Language"
Cohesion: 0.11
Nodes (26): add function, Arrow Lambda Syntax (a, b) -> expr, Auto Run Checkbox, channel() primitive, Commented-out Concurrency Example, Console Output (Squares / Sum of squares / hello 12), Left Code Editor Pane, Karl is a functional-first language (+18 more)

### Community 29 - "Concurrency Expressions (&, spawn, !&, race)"
Cohesion: 0.10
Nodes (25): Channel<T> Abstraction, Implementation Rules for Agents, Sink<T> Abstraction, Source-of-Truth Ordering (SPECS > tests > implementation), Stream<T> Abstraction, Stream Model Guardrails, Task<T> Abstraction, Release v0.8.4 (spawn/race aliases) (+17 more)

### Community 30 - "main"
Cohesion: 0.16
Nodes (16): runtime/debug.BuildInfo, buildInfoSetting(), cliVersion(), kernelCommand(), loomCommand(), loomUsage(), main(), printVersion() (+8 more)

### Community 31 - "Process"
Cohesion: 0.13
Nodes (9): os/exec.Cmd, time.Time, Process, processAwaitWithCancel(), processExitState(), processStatusValue(), processWaitLoop(), processWaitResult (+1 more)

### Community 32 - "io.Writer"
Cohesion: 0.24
Nodes (18): bufio.Scanner, io.Writer, clearScreen(), findExamplesFile(), handleCommand(), hasUnclosedDelimiters(), isCtrlL(), isFatalREPLError() (+10 more)

### Community 33 - "repository"
Cohesion: 0.08
Nodes (23): patterns, patterns, patterns, patterns, patterns, name, patterns, patterns (+15 more)

### Community 34 - "Recoverable Errors (`?`) Runtime Semantics"
Cohesion: 0.11
Nodes (23): Deadlock Probe Correctness Rule, Karl Error Handling Model, Release v0.3.4 (task-failure policies), vigil_in_karl Migration, Collection Error Semantics, Map Collection, Cooperative Cancellation, Recoverable Errors (`?`) Runtime Semantics (+15 more)

### Community 35 - "Identifier"
Cohesion: 0.09
Nodes (16): BlockExpression, Identifier, IfExpression, MemberExpression, QueryExpression, Statement, expressionsToJSON(), FormatJSON() (+8 more)

### Community 36 - "builtins_math.go"
Cohesion: 0.18
Nodes (20): builtinAbs(), builtinCeil(), builtinClamp(), builtinCos(), builtinFloor(), builtinMax(), builtinMin(), builtinPow() (+12 more)

### Community 37 - "main.go"
Cohesion: 0.26
Nodes (18): debugSessionState, debugWatch, debugCommand(), debugUsage(), displayName(), handleDebugCommand(), parseCommand(), parseParseArgs() (+10 more)

### Community 38 - "evalWithConfiguredEvaluator"
Cohesion: 0.18
Nodes (15): failingReader, evalWithConfiguredEvaluator(), stringsFromArray(), TestRuntimeIOArgvDefaultsToEmpty(), TestRuntimeIOArgvInRunContext(), TestRuntimeIOEnvironReturnsSnapshot(), TestRuntimeIOEnvPreservesEmptyString(), TestRuntimeIOEnvReturnsNullWhenMissing() (+7 more)

### Community 39 - ".playground/main.k Demo Program"
Cohesion: 0.16
Nodes (21): Async Task (& http) and wait, compose Higher-Order Helper, decodeJson Builtin, double Helper, for/with/then Accumulating Loop Expression, if/else as Expression, import Expression Returning Module Factory, inc Helper (+13 more)

### Community 40 - "Node"
Cohesion: 0.17
Nodes (8): Node, Format(), printer, bytes.Buffer, Evaluator, tokenFromNode(), Evaluator, annotateErrorToken()

### Community 41 - "stringArg"
Cohesion: 0.12
Nodes (25): net/http.Response, parseStreamType(), rejectUnknownObjectKeys(), builtinHTTP(), Evaluator, extractHeaders(), httpResponseObject(), parseHTTPHandlerResponse() (+17 more)

### Community 42 - "Kernel"
Cohesion: 0.06
Nodes (29): context.Context, database/sql.DB, database/sql/driver.Conn, database/sql/driver.NamedValue, database/sql/driver.Result, database/sql/driver.Rows, database/sql/driver.Stmt, database/sql/driver.Tx (+21 more)

### Community 43 - "builtins_runtime_system.go"
Cohesion: 0.17
Nodes (18): registerRuntimeCoreBuiltins(), registerRuntimeBuiltins(), builtinArgv(), builtinEnv(), builtinEnviron(), builtinProgramPath(), builtinReadLine(), Evaluator (+10 more)

### Community 44 - "karl loom CLI Command"
Cohesion: 0.14
Nodes (20): Heterogeneous Array Literal Echo Output, ASCII Art Loom Banner, Karl CLI Version v0.4.3-0.20260215224035-e6f5ea078ab5+dirty, repl/EXAMPLES.md Reference, filter/map/reduce Collection Pipeline, karl loom CLI Command, Arrow Lambda Syntax (x -> expr), let Binding Declaration (+12 more)

### Community 45 - "Karl Tooling Naming System"
Cohesion: 0.12
Nodes (20): Release v0.8.3 (first-class debugger), Collection Naming Rules, Breakpoint and Step Engine, DAP Bridge (Milestone 3), Evaluator Execution Hooks (beforeNode/afterNode), Karl Debugger (CLI First), Event Loop Runtime Architecture, Interpreter Known Limitations (+12 more)

### Community 46 - "Evaluator"
Cohesion: 0.17
Nodes (8): Debugger, Evaluator, runtimeState, Task, NewEvaluator(), NewEvaluatorWithSource(), FrameAwareDebugger, newRuntimeState()

### Community 47 - "main_test.go"
Cohesion: 0.15
Nodes (19): evalDebugExpression(), parseBreakpointSpec(), parseDebugArgs(), parseRunArgs(), TestEvalDebugExpressionRejectsStatement(), TestEvalDebugExpressionUsesCurrentEnv(), TestParseBreakpointSpecFileAndLine(), TestParseBreakpointSpecInvalid() (+11 more)

### Community 48 - "Karl Language"
Cohesion: 0.12
Nodes (19): Release Workflow (tag v*), GitHub Pages Static Deploy Workflow, Karl Product Context (between Bash and Go), Repo Map, Release v0.4.0 (REPL + modular interpreter), Release v0.6.0 (Sheets + WASM playground), bench (Karl Playground), Editor Plugins (VS Code + Sublime) (+11 more)

### Community 49 - ".evalInfixExpression"
Cohesion: 0.23
Nodes (9): StrictEqual(), Evaluator, evalArrayInfix(), evalFloatInfix(), evalIntegerInfix(), evalNumericInfix(), evalStringInfix(), Array (+1 more)

### Community 50 - "streamPartitionRouter"
Cohesion: 0.21
Nodes (8): streamPartitionBranch, streamPartitionBranchIterator, streamPartitionMode, streamPartitionRouter, Evaluator, newStreamPartitionRouter(), newStreamPartitionSink(), Object

### Community 51 - "Karl Language Specification"
Cohesion: 0.14
Nodes (18): Karl Bench Playground, Karl Anti-Patterns, Idiomatic Karl for DevOps Tooling, run() vs proc() Guidance, Stream-First Transforms, Karl Documentation Hub, Process + Stream Runtime Direction, Vigil Reimplementation in Karl (+10 more)

### Community 52 - "builtins_sql.go"
Cohesion: 0.23
Nodes (22): applySQLOpenOptions(), builtinSQLBegin(), builtinSQLClose(), builtinSQLCommit(), builtinSQLExec(), builtinSQLOpen(), builtinSQLQuery(), builtinSQLQueryOne() (+14 more)

### Community 53 - "value.go"
Cohesion: 0.22
Nodes (3): Array, Set, Tree

### Community 54 - "FormatRuntimeError"
Cohesion: 0.14
Nodes (9): FormatRuntimeError(), formatRuntimeError(), Evaluator, Task, ExitError, RecoverableError, RuntimeError, exitProcess() (+1 more)

### Community 55 - "builtins_time.go"
Cohesion: 0.43
Nodes (6): builtinTimeAdd(), builtinTimeDiff(), builtinTimeFormatRFC3339(), builtinTimeParseRFC3339(), Evaluator, registerTimeBuiltins()

### Community 56 - "parseProgram"
Cohesion: 0.13
Nodes (19): TestExpressionKinds(), TestMatchGuardNestedMatch(), TestMatchGuardParsesWithoutLambda(), TestPatternKinds(), TestPatternLetObjectDestructure(), TestPatternTrailingComma(), TestDebuggerHooksReceiveNodeCallbacks(), TestDebuggerTracksFunctionFrames() (+11 more)

### Community 57 - "Channel"
Cohesion: 0.15
Nodes (4): Channel, channelStreamIterator, taskResult, Task

### Community 58 - "engine.k (core workflow engine)"
Cohesion: 0.13
Nodes (16): concurrent_pipeline.k (8-worker multi-stage pipeline), Parallel Executor Deadlock Fix (buffered channels), Circuit Breaker Pattern, DAG Workflow Mode, engine.k (core workflow engine), Exponential Back-off Retry, Backward-Compatible Opt-in Integration Philosophy, Karl Workflow Engine (+8 more)

### Community 59 - "builtinSignalWatch"
Cohesion: 0.22
Nodes (7): os.Signal, builtinSignalWatch(), channelTrySend(), Evaluator, signalFromName(), platformSignalFromName(), platformSignalFromName()

### Community 60 - "SQLTx"
Cohesion: 0.15
Nodes (6): context.CancelFunc, database/sql.Tx, sync.Once, registerHTTPBuiltins(), cleanupStreamIterator, SQLTx

### Community 61 - "io.Reader"
Cohesion: 0.28
Nodes (7): io.Reader, time.Duration, newStreamInput(), newTTYInput(), redrawLine(), ttyByteEvent, ttyInput

### Community 62 - "Lexer"
Cohesion: 0.40
Nodes (5): Lexer, isDigit(), isHexDigit(), isLetter(), newToken()

### Community 63 - "Expression"
Cohesion: 0.11
Nodes (5): Binding, Expression, ExpressionStatement, ForExpression, SpawnExpression

### Community 64 - ".knb Notebook JSON Format"
Cohesion: 0.15
Nodes (14): 01-quickstart.knb (Beginner), 02-functions-closures.knb (Intermediate), 03-collections.knb (Intermediate), 04-advanced.knb (Advanced), Notebook Cell (code | markdown), .knb Notebook JSON Format, Notebook --output Results JSON, Notebook Runner (+6 more)

### Community 65 - "Karl Language Support (VS Code Extension)"
Cohesion: 0.20
Nodes (14): Karl (Bench Light) Color Scheme, Karl Sublime Text Syntax, karl-lang 0.1.0 Initial Release, Three Extension Installation Methods, vsce package -> karl-lang-0.1.0.vsix, Async Operator Highlighting (& spawn, !& race), Karl Built-in Function Catalog (highlighted), karl trace dap Debug Adapter Bridge (+6 more)

### Community 66 - "newModuleState"
Cohesion: 0.53
Nodes (4): Evaluator, newModuleState(), moduleDefinition, moduleState

### Community 67 - "teeStreamIterator"
Cohesion: 0.26
Nodes (4): teeSideIterator, teeSideMessage, teeStreamIterator, Evaluator

### Community 68 - "bindPattern"
Cohesion: 0.22
Nodes (4): bindPattern(), Evaluator, Function, functionDebugName()

### Community 69 - "Ordered Index Tree tree(kind?)"
Cohesion: 0.18
Nodes (12): Release v0.3.1 (truthy/falsy semantics), Array Collection, Collection Design Goals, Hierarchical Node Tree ntree(...), Ordered Index Tree tree(kind?), Built-in Function Surface, For-Expression Evaluation Algorithm, Member Call Desugaring (no implicit receiver) (+4 more)

### Community 70 - "builtinUUIDParse"
Cohesion: 0.57
Nodes (6): builtinUUIDNew(), builtinUUIDParse(), builtinUUIDValid(), formatUUID(), Evaluator, parseUUID()

### Community 71 - "connect (WASM worker bootstrap)"
Cohesion: 0.18
Nodes (11): initWorker, WASM Worker Transport (bench), connect (WASM worker bootstrap), handleMessage, sendCmd (worker postMessage), Sheet Message Protocol, updateCell, connect (WebSocket bootstrap) (+3 more)

### Community 72 - "Karl Notebook System"
Cohesion: 0.18
Nodes (11): Release v0.5.0 (notebook + Jupyter kernel), Notebook + Jupyter Integration, Debugger Expression Evaluation in Frame, Debugger Frame Tracking, Environment and Lexical Scope, Notebook Cell (code/markdown), .knb JSON Notebook Format, Notebook Minimal Design Principles (+3 more)

### Community 73 - "Karl Process API Examples"
Cohesion: 0.24
Nodes (11): Integer-Only Mathematics (no floats), monte_carlo_pi.k (5-worker parallel Pi estimation), Nico's Karl Examples, parallel_health_checker.k (concurrent HTTP health checks), Producer-Consumer Pattern (rendezvous channels), BYTES Mode Streaming, Karl Process API Examples, kubernetes_health_report.k (+3 more)

### Community 74 - "SliceExpression"
Cohesion: 0.16
Nodes (6): SliceExpression, Evaluator, objectIndexKey(), Evaluator, normalizeIndex(), assignFunctionName()

### Community 75 - "builtinHTTPServe"
Cohesion: 0.40
Nodes (9): net/http.ServeMux, buildHTTPRequestValue(), builtinHTTPServe(), builtinHTTPServerStop(), Evaluator, parseHTTPRoutes(), registerHTTPRoutes(), selectRouteByMethod() (+1 more)

### Community 76 - ".Pretty"
Cohesion: 0.22
Nodes (6): PrettyPrinter, colorize(), Array, Map, Object, Set

### Community 77 - "Builtin"
Cohesion: 0.29
Nodes (3): Builtin, BuiltinFunction, Partial

### Community 78 - "karl-health-demo Namespace Fixture"
Cohesion: 0.31
Nodes (10): Bench Example Snippet Catalog, loadExample, setEditorValue, Namespace karl-health-demo, Deployment web-ok (healthy nginx), Deployment crashloop-app (intentional CrashLoopBackOff), Deployment pending-app (unschedulable via oversized requests), Deployment chain-node (3-replica log generator) (+2 more)

### Community 79 - "sheets/worker.js"
Cohesion: 0.24
Nodes (7): cacheBust, decoder, emitOutput(), go, workerURL, write(), writeSync()

### Community 80 - "mapKeyForValue"
Cohesion: 0.15
Nodes (17): builtinMap(), builtinMapGet(), builtinMapKeys(), builtinMapSet(), Evaluator, builtinMapDelete(), builtinMapHas(), builtinMapValues() (+9 more)

### Community 81 - "inspectObjectPairs"
Cohesion: 0.24
Nodes (5): ModuleObject, Object, inspectObjectKey(), inspectObjectPairs(), isIdentifierKey()

### Community 82 - "Karl REPL (karl loom)"
Cohesion: 0.22
Nodes (10): Client-Side WASM Interpreter, Karl Playground, syscall/js Go-to-Browser Bridge, WASM Rebuild Step (GOOS=js GOARCH=wasm), REPL Colon Commands (:help :quit :examples :env :clear), Karl REPL (karl loom), Automatic Multi-line Input Detection, test_repl.sh REPL Test Suite (+2 more)

### Community 83 - "Expression-Based Language Design"
Cohesion: 0.22
Nodes (9): Karl Design Philosophy, Equality Semantics (== identity, eqv structural), Partial Application with `_`, Pattern Matching Semantics, Runtime Value Model, `->` Single Meaning Principle, Destructuring and Structural Patterns, Expression-Based Language Design (+1 more)

### Community 85 - "builtinFromJSON"
Cohesion: 0.18
Nodes (7): builtinFromJSON(), builtinToJSON(), decodeJSONValue(), encodeJSONValue(), Evaluator, registerJSONBuiltins(), streamFromJSONIterator

### Community 86 - "traceCommand"
Cohesion: 0.67
Nodes (3): TestTraceDAPCommandRejectsArgs(), traceCommand(), traceDAPCommand()

### Community 87 - "Karl API-First Stdlib Skill"
Cohesion: 0.31
Nodes (9): OpenAI Agent Manifest for API-First Stdlib Skill, API Review Checklist, AI-Friendly API Ergonomics, Prefer API-Level Solutions Over User-Land Algorithms, Explicit Complexity Contracts, Karl API-First Stdlib Skill, Hierarchical Node-Tree API (parent, children, siblings, ancestors, descendants), Ordered Tree Search API (closest, floor, ceil, predecessor, successor, range) (+1 more)

### Community 88 - "Karl Logo (playground asset)"
Cohesion: 0.39
Nodes (8): Karl Language Brand Identity, Karl Logo (playground asset), Monochrome Black-on-White Palette, Bold 'K' Monogram Mark, Playground Web UI Branding Asset, Small-Size / Favicon Legibility Rationale, Square Outline Frame Containing the Monogram, Letterspaced 'KARL' Wordmark

### Community 89 - "playground/worker.js"
Cohesion: 0.25
Nodes (4): cacheBust, decoder, go, workerURL

### Community 90 - "<process> Value Type"
Cohesion: 0.32
Nodes (8): Release v0.7.0 (runtime I/O primitives), System Primitives (argv/env/environ/programPath/readLine), proc(spec, opts?) Built-in, <process> Value Type, Process Spec Object, Stream/Process Integration Points, Low-level <stream-reader> / <stream-writer>, Completed Foundations

### Community 92 - "build-and-test Job"
Cohesion: 0.29
Nodes (7): build-and-test Job, Debugger CLI E2E Step, Example Corpus Diff Against Base Branch, Karl CI Workflow, lint Job (golangci-lint), Workflow Engine Tests Workflow, karl debug CLI Command

### Community 93 - "Karl Project Logo"
Cohesion: 0.43
Nodes (7): Karl Brand Identity, Bold K Monogram, Karl Project Logo, Monochrome Minimalist Design Language, Karl (Project / Language Name), Square Outline Frame, KARL Letter-Spaced Wordmark

### Community 94 - "Karl Standard Library Reference"
Cohesion: 0.29
Nodes (7): Collection Families (Array, Map, Set, Tree), Concurrency Helper Built-ins, Method Sugar, Karl Standard Library Reference, Collections Examples, ntree Hierarchical Navigation, Ordered Tree (AVL / Treap)

### Community 95 - "Signal"
Cohesion: 0.13
Nodes (10): Evaluator, Evaluator, Evaluator, Evaluator, Evaluator, Evaluator, Evaluator, isTruthy() (+2 more)

### Community 96 - "buildRange"
Cohesion: 0.43
Nodes (5): buildCharRange(), buildFloatRange(), buildIntRange(), buildRange(), Evaluator

### Community 97 - "Karl Jupyter Kernel"
Cohesion: 0.29
Nodes (7): kernel/install.sh Kernelspec Installer, Karl Jupyter Kernel, Kernel Cross-Cell State Persistence, ZeroMQ Kernel Transport (go-zeromq/zmq4), Karl Notebook System, Notebook --step and --repl Interactive Modes, Per-Client Isolated Sessions

### Community 98 - "Karl VS Code Extension Icon"
Cohesion: 0.48
Nodes (7): Karl VS Code Extension Icon, Boxed K Monogram Mark, Karl Brand Identity, karl-vscode Plugin, KARL Letterspaced Wordmark, VS Code Marketplace Icon Asset Role, Monochrome Minimal Design Choice

### Community 99 - "evalWithPolicy"
Cohesion: 0.52
Nodes (6): evalWithPolicy(), TestEvalDeferCanceledDetachedTaskNotReportedUnhandled(), TestEvalDeferModeDoesNotInterruptMainFlow(), TestEvalDeferModeObservedRecoveredTaskNotReportedUnhandled(), TestEvalFailFastCanceledBlockedRecvTaskCanBeRecoveredOnWait(), TestEvalFailFastCanceledDetachedTaskNotReportedUnhandled()

### Community 100 - "Karl CodeMirror Syntax Mode"
Cohesion: 0.33
Nodes (6): Karl CodeMirror Syntax Mode, buildDecodePlan, decodeTo, highlight, randomGlyph, renderDecoded

### Community 101 - "Karl Stream Examples"
Cohesion: 0.33
Nodes (6): Live Matrix Loop (one-statement fan-in), Stream API Reference, Fan-out / Fan-in Pattern, Karl Stream Examples, kubernetes_logs_error_channel.k, Stream Builtins (merge/join/debounce/tee/spill/split/top)

### Community 102 - "formatLogValue"
Cohesion: 0.54
Nodes (7): builtinLog(), builtinLogt(), builtinStr(), formatLogValue(), Evaluator, writeLogLine(), streamValueToBytes()

### Community 103 - "sortRows"
Cohesion: 0.83
Nodes (3): compareForSort(), sortRows(), queryRow

### Community 104 - "extension.js"
Cohesion: 0.40
Nodes (3): activate(), KarlDebugAdapterDescriptorFactory, vscode

### Community 105 - "test_debugger_cli_e2e.sh"
Cohesion: 0.73
Nodes (5): assert_grep(), assert_not_grep(), fail(), run_trace(), test_debugger_cli_e2e.sh script

### Community 106 - "Deterministic Example Corpus (examples_diff.txt)"
Cohesion: 0.40
Nodes (6): Release Readiness Gate, Deterministic Semantics and Tie-Break Rules, Deterministic Example Corpus (examples_diff.txt), examples_test.go (parses all example programs), Karl Unit Test Suite (tests/), object_disambiguation_test.go (block vs object literal)

### Community 107 - "Concurrency in Karl (Tasks, Failures, Recovery)"
Cohesion: 0.40
Nodes (5): Graceful Shutdown Pattern, Concurrency Model (Tasks, Channels, Cancellation), Concurrency in Karl (Tasks, Failures, Recovery), Cooperative Cancellation, Race Operator (!&)

### Community 108 - "Browser Tab Icon Role for Web Playground"
Cohesion: 0.60
Nodes (5): Browser Tab Icon Role for Web Playground, Karl Playground Favicon (32x32), Karl Language Brand Identity, Minimal Monochrome Framed-Square Design, Bold 'K' Monogram Mark

### Community 110 - "Karl Kernel Favicon (32x32 K mark)"
Cohesion: 0.60
Nodes (5): Browser Tab / Notebook Kernel Icon Asset (32x32 raster), Karl Kernel Favicon (32x32 K mark), K Monogram Brand Mark, Karl Project Identity, Monochrome Square Badge Design (black bold glyph, thin border, white ground)

### Community 111 - "Karl Kernel Logo (64x64 PNG)"
Cohesion: 0.60
Nodes (5): 64x64 Favicon / Icon Asset Sizing, Karl Kernel Project Brand Identity, Karl Kernel Logo (64x64 PNG), Monochrome Square Badge Design, Letter K Monogram Mark

### Community 112 - "newTTYLineWriter"
Cohesion: 0.22
Nodes (8): net.Conn, os.File, newTTYLineWriter(), Client(), enableClientRawMode(), handleConnection(), Server(), ttyLineWriter

### Community 113 - "run_all_tests.sh"
Cohesion: 0.83
Nodes (3): run_test(), run_all_tests.sh script, timeout()

### Community 114 - "builtinFromBase58"
Cohesion: 0.48
Nodes (6): builtinFromBase58(), builtinToBase58(), decodeBase58(), encodeBase58(), Evaluator, reverseBytes()

### Community 115 - "Server"
Cohesion: 0.67
Nodes (3): playgroundCommand(), Server, NewServer()

### Community 117 - "Import Factory and Live Module Object"
Cohesion: 0.67
Nodes (3): Single-Instance Module Import Shorthand, Import Factory and Live Module Object, Import Expressions and Module Factory

### Community 118 - "Karl Landing Page"
Cohesion: 0.67
Nodes (3): Decoding Glyph Animation, Karl Landing Page, Karl Logo Typography Recommendation

### Community 154 - "Pattern"
Cohesion: 0.10
Nodes (9): ArrayPattern, CallPattern, LetStatement, Pattern, RangePattern, TuplePattern, LambdaExpression, matchRangePattern() (+1 more)

### Community 162 - "vigil_builtins_test.go"
Cohesion: 0.22
Nodes (11): database/sql/driver.Value, fakeSQLRows, ensureFakeSQLDriverRegistered(), fakeDriverResetDSN(), reserveLocalAddr(), TestVigilBuiltinsHTTPServeAndStop(), TestVigilBuiltinsSHAUUIDTime(), TestVigilBuiltinsSignalWatchType() (+3 more)

### Community 163 - "graphify"
Cohesion: 0.50
Nodes (3): graphify, Keeping the graph fresh, One-time setup

## Ambiguous Edges - Review These
- `Karl Project Identity` → `Karl Kernel Favicon (32x32 K mark)`  [AMBIGUOUS]
  kernel/logo-32x32.png · relation: references
- `spawn concurrency primitive` → `Karl Language`  [AMBIGUOUS]
  assets/playground.png · relation: implements
- `Commented-out Concurrency Example` → `In-browser Karl Runtime (WASM)`  [AMBIGUOUS]
  assets/playground.png · relation: conceptually_related_to
- `loom (runtime)` → `karl loom (REPL)`  [AMBIGUOUS]
  SPECS/tooling_naming.md · relation: conceptually_related_to

## Knowledge Gaps
- **157 isolated node(s):** `go`, `decoder`, `workerURL`, `cacheBust`, `go` (+152 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 515 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **21 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `Karl Project Identity` and `Karl Kernel Favicon (32x32 K mark)`?**
  _Edge tagged AMBIGUOUS (relation: references) - confidence is low._
- **What is the exact relationship between `spawn concurrency primitive` and `Karl Language`?**
  _Edge tagged AMBIGUOUS (relation: implements) - confidence is low._
- **What is the exact relationship between `Commented-out Concurrency Example` and `In-browser Karl Runtime (WASM)`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `loom (runtime)` and `karl loom (REPL)`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **Why does `Value` connect `Value` to `testing.T`, `value_stream_algebra.go`, `server`, `builtins_process.go`, `channelSendBlocking`, `Tree`, `evalInput`, `DebugController`, `recoverableError`, `builtins_stream_pipeline.go`, `Environment`, `RegisterBuiltins`, `NTree`, `streamIterator`, `ValueType`, `taskAwaitWithCancel`, `New`, `Pattern`, `process_api_test.go`, `Process`, `builtins_math.go`, `main.go`, `evalWithConfiguredEvaluator`, `Node`, `stringArg`, `Kernel`, `builtins_runtime_system.go`, `main_test.go`, `.evalInfixExpression`, `streamPartitionRouter`, `builtins_sql.go`, `value.go`, `builtins_time.go`, `Channel`, `builtinSignalWatch`, `SQLTx`, `teeStreamIterator`, `bindPattern`, `builtinUUIDParse`, `SliceExpression`, `builtinHTTPServe`, `.Pretty`, `Builtin`, `mapKeyForValue`, `inspectObjectPairs`, `builtinFromJSON`, `Signal`, `buildRange`, `evalWithPolicy`, `formatLogValue`, `sortRows`, `builtinFromBase58`?**
  _High betweenness centrality (0.411) - this node is a cross-community bridge._
- **Why does `Environment` connect `Environment` to `buildRange`, `io.Writer`, `server`, `Server`, `bindPattern`, `Node`, `DebugController`, `Value`, `SliceExpression`, `Kernel`, `main_test.go`, `.evalInfixExpression`, `inspectObjectPairs`, `taskAwaitWithCancel`, `New`, `Signal`?**
  _High betweenness centrality (0.034) - this node is a cross-community bridge._
- **Why does `Expression` connect `Expression` to `Token`, `Identifier`, `bindPattern`, `Parser`, `Node`, `SliceExpression`, `taskAwaitWithCancel`, `parseProgram`, `Pattern`?**
  _High betweenness centrality (0.034) - this node is a cross-community bridge._