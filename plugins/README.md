# Karl Editor Plugins

This folder contains Karl editor plugins:

- `plugins/karl-vscode/` for VS Code and Cursor
- `plugins/karl-sublime/` for Sublime Text

## VS Code / Cursor

Install from repo root:

```bash
make vscode-reinstall
```

Or manually package/install:

```bash
cd plugins/karl-vscode
npm install
npm run package
code --install-extension karl-lang-0.1.0.vsix --force
cursor --install-extension karl-lang-0.1.0.vsix --force
```

## Sublime Text (macOS)

1. Open `Preferences -> Browse Packages...`
2. Create folder `Karl` if missing.
3. Copy these files from `plugins/karl-sublime/` into that folder:
   - `Karl.sublime-syntax`
   - `Karl.sublime-color-scheme`

Target path:

`~/Library/Application Support/Sublime Text/Packages/Karl/`

Then open a `.k` file and select `View -> Syntax -> Karl`.
