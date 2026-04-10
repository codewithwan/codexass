package common

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func UTCNow() time.Time {
	return time.Now().UTC()
}

func ISOFormat(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func ParseTime(raw string) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, false
	}
	return t.UTC(), true
}

func RandomString(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func SessionID(email, accountID, orgID string) string {
	sum := md5.Sum([]byte(email + "|" + accountID + "|" + orgID))
	return "sess_" + hex.EncodeToString(sum[:])
}

func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func CopyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("powershell", "-NoProfile", "-Command", "Set-Clipboard -Value @'\n"+text+"\n'@")
	case "darwin":
		cmd = exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
	default:
		cmd = exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
	}
	return cmd.Run()
}

func MaybeMakePrivate(path string, mode os.FileMode) {
	if runtime.GOOS == "windows" {
		return
	}
	_ = os.Chmod(path, mode)
}
