// Runtime context primitives: argv(), programPath(), environ(), env(name).
// Run with:
//   FOO=bar karl run examples/features/runtime_args_env.k -- hello world

let output = {
    programPath: programPath(),
    argv: argv(),
    argc: argv().length,
    foo: env("FOO"),
    envCount: environ().length,
    environ: environ()
}
output
