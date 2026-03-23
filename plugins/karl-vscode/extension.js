const vscode = require("vscode");

class KarlDebugAdapterDescriptorFactory {
    createDebugAdapterDescriptor() {
        const command = process.env.KARL_BIN && process.env.KARL_BIN.trim() !== ""
            ? process.env.KARL_BIN.trim()
            : "karl";
        const args = ["trace", "dap"];
        return new vscode.DebugAdapterExecutable(command, args);
    }
}

function activate(context) {
    const factory = new KarlDebugAdapterDescriptorFactory();
    context.subscriptions.push(
        vscode.debug.registerDebugAdapterDescriptorFactory("karl", factory),
    );
}

function deactivate() {}

module.exports = {
    activate,
    deactivate,
};

