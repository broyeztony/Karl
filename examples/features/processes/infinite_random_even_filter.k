// Producer faster than consumers: compare sequential vs concurrent filtering/writing.
//
// Run:
//   karl run examples/features/processes/infinite_random_even_filter.k

let maxSeen = 30
let workerCount = 4
let filterCostMs = 40
let writeCostMs = 25

let mkProducer = () -> proc({
    command: "sh",
    args: ["-c", "awk 'BEGIN { srand(); while (1) { print int(rand()*101); fflush(); system(\"sleep 0.01\") } }'"],
}, {
    stdout: PIPE,
    stderr: NULL,
    stdoutType: BYTES,
})

let verifyEvenFile = path -> run({
    command: "sh",
    args: [
        "-c",
        "awk 'NF > 0 && ($1 % 2) != 0 { exit 1 } END { if (NR == 0) exit 2 }' \"$1\"",
        "sh",
        path,
    ],
}, {
    maxOutputBytes: 4096,
    overflow: "truncate",
}).ok

let runSequential = outPath -> {
    let p = mkProducer()
    let out = writer(outPath, { type: BYTES, })
    let started = now()

    let stats = for true with carry = "", seen = 0, kept = 0, bytesWritten = 0, chunks = 0 {
        if seen >= maxSeen { break { seen, kept, bytesWritten, chunks, } }

        let [chunk, eof] = p.stdout.read(256)
        if eof { break { seen, kept, bytesWritten, chunks, } }

        chunks++
        let lines = (carry + fromUtf8(chunk)).split("\n")

        let i = 0
        for i < lines.length - 1 && seen < maxSeen with i = 0 {
            let line = lines[i]
            if line != "" {
                let n = parseInt(line)
                seen++
                sleep(filterCostMs)
                if n % 2 == 0 {
                    log("sequential even:", n)
                    sleep(writeCostMs)
                    bytesWritten += out.write(toUtf8(line + "\n"))
                    kept++
                } else {
                    log("sequential odd:", n)
                }
            }
            i++
        } then ()

        carry = lines[lines.length - 1]
    } then ()

    out.close() ? {}
    p.abort()
    let producerStatus = wait p
    let result = {
        mode: "sequential",
        output: outPath,
        durationMs: now() - started,
        stats,
        producerStatus,
        onlyEven: verifyEvenFile(outPath),
    }
    result
}

let runConcurrent = outPath -> {
    let p = mkProducer()
    let numbers = channel()
    let evenLines = channel()
    let started = now()

    // Writer runs independently and can lag behind filters (natural backpressure via channel).
    let writerTask = & (() -> {
        let out = writer(outPath, { type: BYTES, })
        let st = for true with kept = 0, bytesWritten = 0 {
            let [lineBytes, closed] = evenLines.recv()
            if closed { break { kept, bytesWritten, } }
            sleep(writeCostMs)
            bytesWritten += out.write(lineBytes)
            kept++
        } then ()
        out.close() ? {}
        st
    })()

    let worker = id -> & (() -> {
        for true with handled = 0, kept = 0 {
            let [n, closed] = numbers.recv()
            if closed { break { id, handled, kept, } }

            sleep(filterCostMs)
            handled++

            if n % 2 == 0 {
                log("worker", id, "even:", n)
                evenLines.send(toUtf8(str(n) + "\n"))
                kept++
            } else {
                log("worker", id, "odd:", n)
            }
        } then ()
    })()

    let workers = for i < workerCount with i = 0, tasks = [] {
        tasks += [worker(i + 1)]
        i++
    } then tasks

    let readerTask = & (() -> {
        let st = for true with carry = "", seen = 0, chunks = 0 {
            if seen >= maxSeen { break { seen, chunks, } }

            let [chunk, eof] = p.stdout.read(256)
            if eof { break { seen, chunks, } }

            chunks++
            let lines = (carry + fromUtf8(chunk)).split("\n")

            let i = 0
            for i < lines.length - 1 && seen < maxSeen with i = 0 {
                let line = lines[i]
                if line != "" {
                    numbers.send(parseInt(line))
                    seen++
                }
                i++
            } then ()

            carry = lines[lines.length - 1]
        } then ()

        numbers.done()
        st
    })()

    let workersTask = & (() -> {
        let st = for i < workers.length with i = 0, handled = 0, kept = 0 {
            let w = wait workers[i]
            handled += w.handled
            kept += w.kept
            i++
        } then { workers: workers.length, handled, kept, }

        evenLines.done()
        st
    })()

    let readerStats = wait readerTask
    p.abort()
    let producerStatus = wait p
    let workerStats = wait workersTask
    let writerStats = wait writerTask

    let result = {
        mode: "concurrent",
        output: outPath,
        durationMs: now() - started,
        reader: readerStats,
        workers: workerStats,
        writer: writerStats,
        producerStatus,
        onlyEven: verifyEvenFile(outPath),
    }
    result
}

let seq = runSequential("/tmp/karl-random-even-sequential.bin")
let con = runConcurrent("/tmp/karl-random-even-concurrent.bin")

{
    sequential: seq,
    concurrent: con,
    savedMs: seq.durationMs - con.durationMs,
}
