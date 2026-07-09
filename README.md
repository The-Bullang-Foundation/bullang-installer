# Bullang Installer

One-click installer for the [Bullang](https://github.com/The-Bullang-Foundation/Bullang) ecosystem.

Download the installer for your platform from the [latest release](https://github.com/The-Bullang-Foundation/bullang-installer/releases/latest) and double-click to run it.

---

## Download

### For Windows
→ Download **`bullang-installer-windows-x86_64.exe`**
Right-click → Run as administrator if prompted.

### For macOS — Apple Silicon (M1, M2, M3, M4)
→ Download **`bullang-installer-macos-apple-silicon`**
Then open Terminal and run:
```bash
chmod +x bullang-installer-macos-apple-silicon
./bullang-installer-macos-apple-silicon
```

### For macOS — Intel
→ Download **`bullang-installer-macos-intel`**
Then open Terminal and run:
```bash
chmod +x bullang-installer-macos-intel
./bullang-installer-macos-intel
```

### For Linux — x86_64 (Ubuntu, Debian, Arch, most desktop distros)
→ Download **`bullang-installer-linux-amd64`**
Then open Terminal and run:
```bash
chmod +x bullang-installer-linux-amd64
./bullang-installer-linux-amd64
```

### For Linux — ARM64 (Raspberry Pi, ARM servers)
→ Download **`bullang-installer-linux-arm64`**
Then open Terminal and run:
```bash
chmod +x bullang-installer-linux-arm64
./bullang-installer-linux-arm64
```

---

## What gets installed

| Component | Description |
|---|---|
| Go | Programming language runtime |
| Rust + Cargo | Required by Bullscript and Bullarchy |
| Bullscript | Interactive Bullang REPL and tester |
| Bullarchy | Bullang project toolchain (GUI + CLI) |
| Python 3 | Target language support |
| Java (JDK) | Target language support |
| GCC / G++ | C and C++ target language support |

---

## Supported platforms

| OS | Package manager used |
|---|---|
| Ubuntu / Debian | apt |
| Arch / Manjaro | pacman |
| macOS | Homebrew (installed automatically if missing) |
| Windows | winget |

---

## After installation

Restart your terminal, then:

```bash
bullarchy        # launch Bullarchy (GUI opens automatically)
bullarchy --cli  # launch Bullarchy in terminal mode
bullscript       # launch the interactive Bullang REPL
```
