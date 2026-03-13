// Feed a stream into a process sink and inspect RunStatus.

let src = "/tmp/karl-stream-exec.txt"
writeFile(src, "alpha\nbeta\ngamma\n")

let st = read(src, { type: TEXT, }) | exec("wc", "-l")
if !st.ok { fail("wc failed: exit " + str(st.code)) }

log({ lines: st.output.trim(), ok: st.ok, })
