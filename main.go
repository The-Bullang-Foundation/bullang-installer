package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

//go:embed Icon.png
var iconBytes []byte

// ── OS detection ──────────────────────────────────────────────────────────────

type OS int

const (
	OSLinuxApt  OS = iota
	OSLinuxArch
	OSMac
	OSWindows
	OSUnknown
)

func detectOS() OS {
	switch runtime.GOOS {
	case "windows":
		return OSWindows
	case "darwin":
		return OSMac
	case "linux":
		data, err := os.ReadFile("/etc/os-release")
		if err != nil {
			return OSLinuxApt
		}
		content := strings.ToLower(string(data))
		if strings.Contains(content, "arch") || strings.Contains(content, "manjaro") {
			return OSLinuxArch
		}
		return OSLinuxApt
	default:
		return OSUnknown
	}
}

// ── Installation steps ────────────────────────────────────────────────────────

type Step struct {
	label string
	run   func(log func(string)) error
}

func buildSteps(currentOS OS) []Step {
	return []Step{
		{
			label: "Installing Go...",
			run: func(log func(string)) error {
				if _, err := exec.LookPath("go"); err == nil {
					log("Go is already installed, skipping.")
					return nil
				}
				switch currentOS {
				case OSLinuxApt:
					_ = priv(log, "apt-get", "update", "-y")
					return priv(log, "apt-get", "install", "-y", "golang")
				case OSLinuxArch:
					return priv(log, "pacman", "-Sy", "--noconfirm", "go")
				case OSMac:
					if err := ensureBrew(log); err != nil {
						return err
					}
					return runCmd(log, "brew", "install", "go")
				case OSWindows:
					return runCmd(log, "winget", "install", "-e", "--id", "GoLang.Go", "--silent")
				default:
					return fmt.Errorf("unsupported OS")
				}
			},
		},
		{
			label: "Installing Rust and Cargo via rustup...",
			run: func(log func(string)) error {
				if _, err := exec.LookPath("cargo"); err == nil {
					log("Cargo is already installed, skipping.")
					return nil
				}
				switch currentOS {
				case OSWindows:
					return runCmd(log, "powershell", "-Command",
						"Invoke-WebRequest -Uri https://win.rustup.rs/x86_64 -OutFile $env:TEMP\\rustup-init.exe; "+
							"Start-Process -Wait -FilePath $env:TEMP\\rustup-init.exe -ArgumentList '-y'")
				default:
					return runCmd(log, "sh", "-c",
						"curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y")
				}
			},
		},
		{
			label: "Updating PATH for Cargo...",
			run: func(log func(string)) error {
				home, _ := os.UserHomeDir()
				cargoPath := home + "/.cargo/bin"
				_ = os.Setenv("PATH", cargoPath+string(os.PathListSeparator)+os.Getenv("PATH"))
				log("Cargo PATH updated.")
				return nil
			},
		},
		{
			label: "Installing Bullscript...",
			run: func(log func(string)) error {
				return runCmd(log, "cargo", "install",
					"--git", "https://github.com/The-Bullang-Foundation/Bullscript.git",
					"--force")
			},
		},
		{
			label: "Installing Bullarchy...",
			run: func(log func(string)) error {
				return runCmd(log, "cargo", "install",
					"--git", "https://github.com/The-Bullang-Foundation/Bullarchy.git",
					"--force")
			},
		},
		{
			label: "Running editor setup...",
			run: func(log func(string)) error {
				home, _ := os.UserHomeDir()
				return runCmd(log, home+"/.cargo/bin/bullarchy", "editor-setup")
			},
		},
		{
			label: "Installing target languages (Python, Java, C/C++)...",
			run: func(log func(string)) error {
				switch currentOS {
				case OSLinuxApt:
					_ = priv(log, "apt-get", "update", "-y")
					return priv(log, "apt-get", "install", "-y",
						"python3", "python3-pip", "default-jdk", "gcc", "g++", "build-essential")
				case OSLinuxArch:
					return priv(log, "pacman", "-Sy", "--noconfirm",
						"python", "python-pip", "jdk-openjdk", "gcc")
				case OSMac:
					if err := ensureBrew(log); err != nil {
						return err
					}
					return runCmd(log, "brew", "install", "python", "openjdk", "gcc")
				case OSWindows:
					for _, cmd := range [][]string{
						{"winget", "install", "-e", "--id", "Python.Python.3", "--silent"},
						{"winget", "install", "-e", "--id", "Microsoft.OpenJDK.21", "--silent"},
						{"winget", "install", "-e", "--id", "GnuWin32.Gcc", "--silent"},
					} {
						if err := runCmd(log, cmd[0], cmd[1:]...); err != nil {
							log(fmt.Sprintf("Warning: %v", err))
						}
					}
					return nil
				default:
					return fmt.Errorf("unsupported OS")
				}
			},
		},
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func ensureBrew(log func(string)) error {
	if _, err := exec.LookPath("brew"); err == nil {
		return nil
	}
	log("Installing Homebrew...")
	return runCmd(log, "bash", "-c",
		`/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`)
}

// privExec returns the first available privilege escalation tool.
// Returns "" if none found — callers should handle that case.
func privExec() string {
	for _, tool := range []string{"sudo", "doas", "pkexec"} {
		if _, err := exec.LookPath(tool); err == nil {
			return tool
		}
	}
	return ""
}

// priv runs a command with privilege escalation if available.
// If no escalation tool is found, logs a clear message and returns an error.
func priv(log func(string), args ...string) error {
	tool := privExec()
	if tool == "" {
		log("✗ No privilege escalation tool found (sudo/doas/pkexec).")
		log("  Please install the following packages manually:")
		log("  " + strings.Join(args, " "))
		return fmt.Errorf("no privilege escalation tool available (sudo/doas/pkexec not found)")
	}
	return runCmd(log, tool, args...)
}

func runCmd(log func(string), name string, args ...string) error {
	log(fmt.Sprintf("$ %s %s", name, strings.Join(args, " ")))
	cmd := exec.Command(name, args...)
	cmd.Stdout = &logWriter{log: log}
	cmd.Stderr = &logWriter{log: log}
	return cmd.Run()
}

type logWriter struct{ log func(string) }

func (w *logWriter) Write(p []byte) (n int, err error) {
	for _, line := range strings.Split(strings.TrimRight(string(p), "\n"), "\n") {
		if strings.TrimSpace(line) != "" {
			w.log(line)
		}
	}
	return len(p), nil
}

// ── Result channel ────────────────────────────────────────────────────────────

type result struct {
	err      error
	stepName string
	done     bool
}

// ── UI ────────────────────────────────────────────────────────────────────────

func main() {
	currentOS := detectOS()
	a := app.New()
	a.SetIcon(fyne.NewStaticResource("Icon.png", iconBytes))
	w := a.NewWindow("Bullang Installer")
	w.Resize(fyne.NewSize(660, 580))
	w.SetFixedSize(true)

	logo := canvas.NewImageFromResource(fyne.NewStaticResource("Icon.png", iconBytes))
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(96, 96))

	title := widget.NewLabelWithStyle("Bullang Ecosystem Installer",
		fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabelWithStyle(
		"Installs Go, Rust, Bullscript, Bullarchy, and target languages",
		fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	osNames := map[OS]string{
		OSLinuxApt: "Ubuntu / Debian", OSLinuxArch: "Arch Linux",
		OSMac: "macOS", OSWindows: "Windows", OSUnknown: "Unknown OS",
	}
	osLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("Detected OS: %s", osNames[currentOS]),
		fyne.TextAlignCenter, fyne.TextStyle{})

	progress := widget.NewProgressBar()
	progress.Hide()

	stepLabel := widget.NewLabel("")
	stepLabel.Alignment = fyne.TextAlignCenter
	stepLabel.Hide()

	// Error label — shown inline when a step fails
	errorLabel := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	errorLabel.Hide()

	logOutput := widget.NewMultiLineEntry()
	logOutput.Disable()
	logOutput.Hide()
	logScroll := container.NewScroll(logOutput)
	logScroll.SetMinSize(fyne.NewSize(620, 220))

	// Channel for goroutine → main thread communication
	results := make(chan result, 1)

	appendLog := func(msg string) {
		t := logOutput.Text
		if t == "" {
			logOutput.SetText(msg)
		} else {
			logOutput.SetText(t + "\n" + msg)
		}
		logScroll.ScrollToBottom()
	}

	installBtn := widget.NewButton("  Install  ", nil)
	installBtn.Importance = widget.HighImportance

	// Retry button — hidden until failure
	retryBtn := widget.NewButton("  Retry  ", nil)
	retryBtn.Hide()

	installBtn.OnTapped = func() {
		installBtn.Disable()
		retryBtn.Hide()
		errorLabel.Hide()
		progress.Show()
		stepLabel.Show()
		logOutput.Show()
		logOutput.SetText("")

		steps := buildSteps(currentOS)
		total := float64(len(steps))

		go func() {
			for i, step := range steps {
				// Send step update via channel
				results <- result{stepName: step.label}

				// Update progress directly — Fyne widget methods are
				// goroutine-safe for SetValue/SetText
				stepLabel.SetText(step.label)
				progress.SetValue(float64(i) / total)
				appendLog(fmt.Sprintf("\n── %s", step.label))

				if err := step.run(appendLog); err != nil {
					appendLog(fmt.Sprintf("\n✗ FAILED: %v", err))
					results <- result{err: err, stepName: step.label}
					return
				}
			}
			results <- result{done: true}
		}()
	}

	// Process results on main thread via polling with a goroutine
	// that calls fyne's thread-safe QueueEvent
	go func() {
		for r := range results {
			r := r // capture
			if r.stepName != "" && r.err == nil && !r.done {
				continue // progress already set directly
			}
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Bullang Installer",
				Content: "See installer window",
			})
			if r.err != nil {
				// Show error on main thread
				w.Canvas().Refresh(w.Content())
				errorLabel.SetText(fmt.Sprintf("✗ Step failed: %s\n\nError: %v\n\nCheck the log above for details.", r.stepName, r.err))
				errorLabel.Show()
				progress.Hide()
				stepLabel.Hide()
				installBtn.Enable()
				retryBtn.Show()
				w.Canvas().Refresh(w.Content())
			} else if r.done {
				progress.SetValue(1)
				stepLabel.SetText("✓ Installation complete!")
				appendLog("\n✓ Bullang ecosystem installed successfully.")
				appendLog("  Restart your terminal, then run: bullarchy  or  bullscript")
				w.Canvas().Refresh(w.Content())
			}
		}
	}()

	retryBtn.OnTapped = func() {
		installBtn.OnTapped()
	}

	w.SetContent(container.NewPadded(container.NewVBox(
		container.NewCenter(logo),
		title, subtitle, osLabel,
		widget.NewSeparator(),
		container.NewCenter(installBtn),
		container.NewCenter(retryBtn),
		stepLabel,
		progress,
		errorLabel,
		widget.NewSeparator(),
		logScroll,
	)))

	w.ShowAndRun()
}
