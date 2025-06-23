package telegram

import (
	"fmt"
	"os/exec"
	"runtime"
)

func openFile(path string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", path).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
