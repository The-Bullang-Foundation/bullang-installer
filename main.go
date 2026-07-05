package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

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
					_ = runCmd(log, "sudo", "apt-get", "update", "-y")
					return runCmd(log, "sudo", "apt-get", "install", "-y", "golang")
				case OSLinuxArch:
					return runCmd(log, "sudo", "pacman", "-Sy", "--noconfirm", "go")
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
					_ = runCmd(log, "sudo", "apt-get", "update", "-y")
					return runCmd(log, "sudo", "apt-get", "install", "-y",
						"python3", "python3-pip", "default-jdk", "gcc", "g++", "build-essential")
				case OSLinuxArch:
					return runCmd(log, "sudo", "pacman", "-Sy", "--noconfirm",
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

func ensureBrew(log func(string)) error {
	if _, err := exec.LookPath("brew"); err == nil {
		return nil
	}
	log("Installing Homebrew...")
	return runCmd(log, "bash", "-c",
		`/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`)
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

func main() {
	currentOS := detectOS()
	a := app.New()
	w := a.NewWindow("Bullang Installer")
	w.Resize(fyne.NewSize(660, 560))
	w.SetFixedSize(true)

	logo := canvas.NewImageFromFile("Icon.png")
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

	logOutput := widget.NewMultiLineEntry()
	logOutput.Disable()
	logOutput.Hide()
	logScroll := container.NewScroll(logOutput)
	logScroll.SetMinSize(fyne.NewSize(620, 220))

	log := func(msg string) {
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

	installBtn.OnTapped = func() {
		installBtn.Disable()
		progress.Show()
		stepLabel.Show()
		logOutput.Show()
		steps := buildSteps(currentOS)
		total := float64(len(steps))

		go func() {
			for i, step := range steps {
				stepLabel.SetText(step.label)
				progress.SetValue(float64(i) / total)
				log(fmt.Sprintf("\n── %s", step.label))
				if err := step.run(log); err != nil {
					log(fmt.Sprintf("ERROR: %v", err))
					dialog.ShowError(fmt.Errorf("%s\n\n%v", step.label, err), w)
					installBtn.Enable()
					return
				}
			}
			progress.SetValue(1)
			stepLabel.SetText("✓ Installation complete!")
			log("\n✓ Bullang ecosystem installed successfully.")
			log("  Restart your terminal, then run: bullarchy  or  bullscript")
			dialog.ShowInformation("Installation Complete",
				"Bullang ecosystem installed successfully!\n\nRestart your terminal, then:\n  bullarchy\n  bullscript", w)
		}()
	}

	w.SetContent(container.NewPadded(container.NewVBox(
		container.NewCenter(logo),
		title, subtitle, osLabel,
		widget.NewSeparator(),
		container.NewCenter(installBtn),
		stepLabel, progress,
		widget.NewSeparator(),
		logScroll,
	)))
	w.ShowAndRun()
}
