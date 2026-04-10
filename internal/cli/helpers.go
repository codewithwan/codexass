package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"codexass/internal/common"
	"codexass/internal/config"
)

func printJSON(value interface{}) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(raw))
	return nil
}

func printRootHelp() {
	fmt.Fprintln(os.Stdout, "codexass <command> [flags]")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Commands:")
	fmt.Fprintln(os.Stdout, "  login      Start browser OAuth login")
	fmt.Fprintln(os.Stdout, "  complete   Complete OAuth from pending login")
	fmt.Fprintln(os.Stdout, "  usage      Check usage/quota windows")
	fmt.Fprintln(os.Stdout, "  list       List stored sessions")
	fmt.Fprintln(os.Stdout, "  models     List known Codex OAuth-compatible models")
	fmt.Fprintln(os.Stdout, "  chat       Start terminal streaming chat")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Global flag:")
	fmt.Fprintf(os.Stdout, "  %s <dir> (default: %s)\n", config.StoreDirFlag, config.DefaultStoreDir())
}

func promptCopy(url string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("copy> ")
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	if strings.EqualFold(strings.TrimSpace(line), "c") {
		if err := common.CopyToClipboard(url); err != nil {
			fmt.Println("Clipboard copy failed. Copy the URL manually.")
			return
		}
		fmt.Println("Auth URL copied to clipboard.")
	}
}
