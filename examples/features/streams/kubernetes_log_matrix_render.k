// Utilities to render a pod/level matrix in a top-like terminal view.

let spaces = n -> for i < n with i = 0, out = "" { out = out + " "; i++ } then out
let fill = (ch, n) -> for i < n with i = 0, out = "" { out = out + ch; i++ } then out
let pad = (text, width) -> if text.length >= width { text } else { text + spaces(width - text.length) }
let key = (pod, level) -> pod + "|" + level
let at = (counts, pod, level) -> get(counts, key(pod, level)) ? 0

let create = (ns, selector, names) -> {
    let levelW = 7
    let countW = 6
    let border =
        "+"
        + fill("-", levelW + 2)
        + "+"
        + fill("-", countW + 2)
        + "+"
        + fill("-", countW + 2)
        + "+"
        + fill("-", countW + 2)
        + "+"

    let row = (a, b, c, d) ->
        "| "
        + pad(a, levelW)
        + " | "
        + pad(b, countW)
        + " | "
        + pad(c, countW)
        + " | "
        + pad(d, countW)
        + " |"

    let render = counts -> {
        log("\u001b[2J\u001b[H")
        log("ns=" + ns + " selector=" + selector)
        log("pod1=" + names[0])
        log("pod2=" + names[1])
        log("pod3=" + names[2])
        log("")
        log(border)
        log(row("LEVEL", "POD1", "POD2", "POD3"))
        log(border)
        log(row("DEBUG", str(at(counts, "pod1", "DEBUG")), str(at(counts, "pod2", "DEBUG")), str(at(counts, "pod3", "DEBUG"))))
        log(row("INFO", str(at(counts, "pod1", "INFO")), str(at(counts, "pod2", "INFO")), str(at(counts, "pod3", "INFO"))))
        log(row("WARN", str(at(counts, "pod1", "WARN")), str(at(counts, "pod2", "WARN")), str(at(counts, "pod3", "WARN"))))
        log(row("ERROR", str(at(counts, "pod1", "ERROR")), str(at(counts, "pod2", "ERROR")), str(at(counts, "pod3", "ERROR"))))
        log(row("OTHER", str(at(counts, "pod1", "OTHER")), str(at(counts, "pod2", "OTHER")), str(at(counts, "pod3", "OTHER"))))
        log(border)
    }

    { render, }
}
