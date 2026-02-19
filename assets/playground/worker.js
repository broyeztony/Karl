
// Web wrapper for Go WASM in Worker
importScripts("wasm_exec.js");

const go = new Go();
const decoder = new TextDecoder("utf-8");

self.runKarl = null;
self.__karl_done = () => {
    self.postMessage({ type: 'done' });
};

// Override fs for Worker
self.fs = {
    constants: { O_WRONLY: -1, O_RDWR: -1, O_CREAT: -1, O_TRUNC: -1, O_APPEND: -1, O_EXCL: -1 },
    writeSync(fd, buf) {
        const str = decoder.decode(buf);
        self.postMessage({ type: 'output', data: str });
        if (str.includes("Karl WASM Runtime initialized")) {
            self.postMessage({ type: 'ready' });
        }
        return buf.length;
    },
    write(fd, buf, offset, length, position, callback) {
        if (offset !== undefined && length !== undefined) {
            buf = buf.subarray(offset, offset + length);
        }
        const str = decoder.decode(buf);
        self.postMessage({ type: 'output', data: str });
        if (str.includes("Karl WASM Runtime initialized")) {
            self.postMessage({ type: 'ready' });
        }
        callback(null, buf.length);
    },
    open(path, flags, mode, callback) {
        callback(new Error("not implemented"));
    }
};

(async () => {
    try {
        let result;
        if (WebAssembly.instantiateStreaming) {
            result = await WebAssembly.instantiateStreaming(fetch("karl.wasm"), go.importObject);
        } else {
            const resp = await fetch("karl.wasm");
            const buf = await resp.arrayBuffer();
            result = await WebAssembly.instantiate(buf, go.importObject);
        }
        // Run Go (blocks until exit)
        go.run(result.instance);
    } catch (e) {
        console.error(e);
        self.postMessage({ type: 'error', data: e.toString() });
    }
})();

self.onmessage = (e) => {
    const msg = e.data;
    if (msg.type === 'run') {
        if (self.runKarl) {
            try {
                const res = self.runKarl(msg.source);
                if (typeof res === 'string') self.postMessage({ type: 'output', data: "\nResult: " + res });
            } catch (err) {
                self.postMessage({ type: 'output', data: "\nPanic: " + err });
                self.postMessage({ type: 'done' });
            }
        } else {
            self.postMessage({ type: 'error', data: "Runtime loading..." });
        }
    } else if (msg.type === 'init') {
        // Init handled automatically on load, but we can check if already ready
        if (self.runKarl) {
            self.postMessage({ type: 'ready' });
        }
    }
};
