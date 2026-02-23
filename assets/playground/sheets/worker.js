importScripts("../wasm_exec.js");

const go = new Go();
const decoder = new TextDecoder("utf-8");
const workerURL = new URL(self.location.href);
const cacheBust = workerURL.searchParams.get("v");
const wasmURL = cacheBust
    ? `karl-sheets.wasm?v=${encodeURIComponent(cacheBust)}`
    : "karl-sheets.wasm";

let runtimeReady = false;

function emitOutput(str) {
    self.postMessage({ type: "output", data: str });
    if (str.includes("Karl Sheets WASM Runtime initialized")) {
        runtimeReady = true;
        self.postMessage({ type: "ready" });
    }
}

self.fs = {
    constants: { O_WRONLY: -1, O_RDWR: -1, O_CREAT: -1, O_TRUNC: -1, O_APPEND: -1, O_EXCL: -1 },
    writeSync(fd, buf) {
        emitOutput(decoder.decode(buf));
        return buf.length;
    },
    write(fd, buf, offset, length, position, callback) {
        if (offset !== undefined && length !== undefined) {
            buf = buf.subarray(offset, offset + length);
        }
        emitOutput(decoder.decode(buf));
        callback(null, buf.length);
    },
    open(path, flags, mode, callback) {
        callback(new Error("not implemented"));
    }
};

function dispatchCommand(cmd) {
    if (!runtimeReady || typeof self.runKarlSheets !== "function") {
        self.postMessage({ type: "error", data: "Runtime loading..." });
        return;
    }

    let raw;
    try {
        raw = self.runKarlSheets(JSON.stringify(cmd));
    } catch (err) {
        self.postMessage({ type: "error", data: String(err) });
        return;
    }

    let payload;
    try {
        payload = JSON.parse(raw || "{}");
    } catch (err) {
        self.postMessage({ type: "error", data: "invalid runtime payload: " + String(err) });
        return;
    }

    if (payload.error) {
        self.postMessage({ type: "error", data: payload.error });
        return;
    }

    if (payload.reset) {
        self.postMessage({ type: "sheet_message", data: { type: "reset" } });
    }
    if (Array.isArray(payload.messages)) {
        for (const msg of payload.messages) {
            self.postMessage({ type: "sheet_message", data: msg });
        }
    }
}

self.onmessage = (e) => {
    const msg = e.data || {};
    if (msg.type === "init") {
        if (runtimeReady) {
            dispatchCommand({ type: "init" });
            return;
        }
        const started = Date.now();
        const poll = setInterval(() => {
            if (runtimeReady) {
                clearInterval(poll);
                dispatchCommand({ type: "init" });
                return;
            }
            if (Date.now() - started > 15000) {
                clearInterval(poll);
                self.postMessage({ type: "error", data: "WASM runtime init timeout" });
            }
        }, 25);
        return;
    }
    if (msg.type === "cmd") {
        dispatchCommand(msg.data || {});
    }
};

(async () => {
    try {
        let result;
        if (WebAssembly.instantiateStreaming) {
            result = await WebAssembly.instantiateStreaming(fetch(wasmURL, { cache: "no-store" }), go.importObject);
        } else {
            const resp = await fetch(wasmURL, { cache: "no-store" });
            const buf = await resp.arrayBuffer();
            result = await WebAssembly.instantiate(buf, go.importObject);
        }
        go.run(result.instance);
    } catch (e) {
        self.postMessage({ type: "error", data: String(e) });
    }
})();
