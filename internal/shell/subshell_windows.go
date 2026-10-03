//go:build windows

package shell

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type procInfo struct {
	pid    uint32
	ppid   uint32
	exe    string
}

// DetectActiveShell detects the parent shell process on Windows via Toolhelp32
func DetectActiveShell() (*DetectedShell, error) {
	currPPID := uint32(os.Getppid())

	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return fallbackDetection(int(currPPID))
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	procMap := make(map[uint32]procInfo)
	err = windows.Process32First(snapshot, &entry)
	for err == nil {
		exeName := windows.UTF16ToString(entry.ExeFile[:])
		procMap[entry.ProcessID] = procInfo{
			pid:  entry.ProcessID,
			ppid: entry.ParentProcessID,
			exe:  exeName,
		}
		err = windows.Process32Next(snapshot, &entry)
	}

	// Trace upwards from immediate parent
	targetPID := currPPID
	targetExe := ""

	// Check up to 4 ancestors to penetrate terminal emulators (e.g. conhost, WindowsTerminal, code)
	for i := 0; i < 4; i++ {
		info, exists := procMap[targetPID]
		if !exists {
			break
		}
		exeLower := strings.ToLower(info.exe)

		// If this process is a recognized shell, stop here!
		if isRecognizedShell(exeLower) {
			targetExe = info.exe
			break
		}

		// If this is a terminal wrapper, walk up to parent
		if isTerminalWrapper(exeLower) {
			targetPID = info.ppid
			continue
		}

		// Unknown intermediate, record and break
		targetExe = info.exe
		break
	}

	if targetExe == "" {
		return fallbackDetection(int(currPPID))
	}

	shellType, binPath := classifyWindowsShell(targetExe)
	return &DetectedShell{
		Type:       shellType,
		BinaryPath: binPath,
		ParentPID:  int(targetPID),
		ParentName: targetExe,
	}, nil
}

func isRecognizedShell(exe string) bool {
	switch exe {
	case "pwsh.exe", "powershell.exe", "cmd.exe", "nu.exe", "bash.exe", "zsh.exe":
		return true
	default:
		return false
	}
}

func isTerminalWrapper(exe string) bool {
	switch exe {
	case "conhost.exe", "windowsterminal.exe", "openconsole.exe", "code.exe":
		return true
	default:
		return false
	}
}

func classifyWindowsShell(exe string) (ShellType, string) {
	exeLower := strings.ToLower(exe)
	switch exeLower {
	case "pwsh.exe":
		path, _ := exec.LookPath("pwsh.exe")
		if path == "" {
			path = exe
		}
		return ShellTypePwsh, path
	case "powershell.exe":
		path, _ := exec.LookPath("powershell.exe")
		if path == "" {
			path = exe
		}
		return ShellTypePowerShell, path
	case "cmd.exe":
		path := os.Getenv("COMSPEC")
		if path == "" {
			path, _ = exec.LookPath("cmd.exe")
		}
		return ShellTypeCmd, path
	case "nu.exe":
		path, _ := exec.LookPath("nu.exe")
		return ShellTypeNushell, path
	case "bash.exe":
		path, _ := exec.LookPath("bash.exe")
		return ShellTypeBash, path
	default:
		return ShellTypeUnknown, exe
	}
}

func fallbackDetection(ppid int) (*DetectedShell, error) {
	// Probe pwsh -> powershell -> COMSPEC
	if path, err := exec.LookPath("pwsh.exe"); err == nil {
		return &DetectedShell{Type: ShellTypePwsh, BinaryPath: path, ParentPID: ppid, ParentName: "pwsh.exe"}, nil
	}
	if path, err := exec.LookPath("powershell.exe"); err == nil {
		return &DetectedShell{Type: ShellTypePowerShell, BinaryPath: path, ParentPID: ppid, ParentName: "powershell.exe"}, nil
	}
	comspec := os.Getenv("COMSPEC")
	if comspec != "" {
		return &DetectedShell{Type: ShellTypeCmd, BinaryPath: comspec, ParentPID: ppid, ParentName: filepath.Base(comspec)}, nil
	}
	return &DetectedShell{Type: ShellTypeCmd, BinaryPath: "cmd.exe", ParentPID: ppid, ParentName: "cmd.exe"}, nil
}

// SpawnSubshell spawns an interactive subshell with memory-injected prompt on Windows
func SpawnSubshell(shell *DetectedShell, cfg *SubshellConfig) (int, error) {
	var cmd *exec.Cmd

	switch shell.Type {
	case ShellTypePwsh, ShellTypePowerShell:
		// In-memory prompt injection: updates prompt for this session without writing to disk
		promptCode := fmt.Sprintf(`function prompt { "[(aenv:%s)] " + (Get-Location) + "> " }`, cfg.PresetName)
		cmd = exec.Command(shell.BinaryPath, "-NoLogo", "-NoExit", "-Command", promptCode)

	case ShellTypeCmd:
		promptStr := fmt.Sprintf("[(aenv:%s)] $P$G", cfg.PresetName)
		cmd = exec.Command(shell.BinaryPath, "/k", "prompt", promptStr)

	case ShellTypeBash:
		cmd = exec.Command(shell.BinaryPath, "-i")

	default:
		// Fallback: spawn whatever shell was detected with -i or interactive flag
		cmd = exec.Command(shell.BinaryPath)
	}

	// Forward stdio directly to host terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Construct environment variables
	env := os.Environ()
	env = append(env, cfg.FormatEnvSlice()...)
	if shell.Type == ShellTypeBash {
		env = append(env, fmt.Sprintf("PS1=[(aenv:%s)] \\w$ ", cfg.PresetName))
	}
	cmd.Env = env

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}

	return 0, nil
}
