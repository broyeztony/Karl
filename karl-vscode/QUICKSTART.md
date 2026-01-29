# Karl VS Code Extension - Quick Start

## 🚀 Installation & Testing

### Method 1: Install Locally (Easiest)

1. **Open VS Code**
2. **Copy the extension folder:**
   ```bash
   # Mac/Linux
   cp -r karl-vscode ~/.vscode/extensions/
   
   # Or manually copy to:
   # Mac: ~/.vscode/extensions/
   # Windows: %USERPROFILE%\.vscode\extensions\
   # Linux: ~/.vscode/extensions/
   ```

3. **Reload VS Code:**
   - Press `Cmd+Shift+P` (Mac) or `Ctrl+Shift+P` (Windows/Linux)
   - Type "Reload Window"
   - Press Enter

4. **Test it:**
   - Open any `.k` file from `examples/`
   - Enjoy beautiful syntax highlighting! ✨

### Method 2: Package as VSIX (For Distribution)

1. **Install dependencies:**
   ```bash
   cd karl-vscode
   npm install
   ```

2. **Package the extension:**
   ```bash
   npm install -g vsce
   vsce package
   ```
   
   This creates `karl-lang-0.1.0.vsix`

3. **Install the VSIX:**
   - In VS Code: `Cmd+Shift+P` → "Install from VSIX"
   - Select `karl-lang-0.1.0.vsix`

### Method 3: Development Mode

1. **Open extension in VS Code:**
   ```bash
   cd karl-vscode
   code .
   ```

2. **Press F5** to launch Extension Development Host

3. **Open a `.k` file** in the new window to see highlighting

## 🎨 What You Get

### Syntax Highlighting for:
- ✅ Keywords (`let`, `if`, `for`, `match`, `wait`)
- ✅ Async operators (`&` spawn, `|` race)
- ✅ Arrow functions (`->`)
- ✅ Built-in functions (`http`, `map`, `rendezvous`)
- ✅ Strings and numbers
- ✅ Comments (`//`)

### Editor Features:
- ✅ Auto-closing brackets/quotes
- ✅ Comment toggling (`Cmd+/`)
- ✅ Bracket matching
- ✅ Smart indentation

## 🧪 Test Files

Try opening these example files to see the highlighting:
```bash
code examples/nico/concurrent_pipeline.k
code examples/nico/monte_carlo_pi.k
code examples/nico/parallel_health_checker.k
```

## 🎯 Customization

The extension supports all VS Code color themes. Try:
- **Dark:** One Dark Pro, Dracula, Nord
- **Light:** GitHub Light, Atom One Light

## 📝 File Association

The extension automatically recognizes `.k` files as Karl language.

## 🐛 Troubleshooting

**Syntax highlighting not showing?**
1. Check file extension is `.k`
2. Reload VS Code window
3. Check extension is installed: `code --list-extensions | grep karl`

**Want to modify colors?**
Edit your VS Code settings:
```json
"editor.tokenColorCustomizations": {
  "textMateRules": [
    {
      "scope": "keyword.operator.async.karl",
      "settings": { "foreground": "#FF6B6B" }
    }
  ]
}
```

## 🚀 Next Steps

1. Install the extension
2. Open a Karl file
3. Enjoy beautiful, highlighted code!
4. Share screenshots on social media! 📸

---

**Happy Karl coding!** ✨🎨
