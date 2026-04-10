package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"codexass/internal/domain"
	chatservice "codexass/internal/service/chat"
)

var style = struct {
	Reset  string
	Bold   string
	Dim    string
	User   string
	Bot    string
	System string
	Error  string
	Border string
}{
	Reset:  "\033[0m",
	Bold:   "\033[1m",
	Dim:    "\033[2m",
	User:   "\033[38;5;114m",
	Bot:    "\033[38;5;117m",
	System: "\033[38;5;221m",
	Error:  "\033[38;5;203m",
	Border: "\033[38;5;239m",
}

func printChunk(chunk string) {
	_, _ = os.Stdout.Write([]byte(chunk))
}

func RunChat(session domain.SessionRecord, model string) error {
	state, err := chatservice.NewState()
	if err != nil {
		return err
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\n%s+----------------------------------------------+%s\n", style.Border, style.Reset)
	fmt.Printf("%s|             %s%sCODEX TERMINAL CHAT%s%s              |%s\n", style.Border, style.Bold, style.Bot, style.Reset, style.Border, style.Reset)
	fmt.Printf("%s+----------------------------------------------+%s\n", style.Border, style.Reset)
	fmt.Printf(" %sModel: %s%s%s\n", style.System, style.Dim, model, style.Reset)
	fmt.Printf(" %sCommands: %s/exit, /quit, /reset%s\n\n", style.System, style.Dim, style.Reset)
	for {
		fmt.Printf("%syou%s%s>%s ", style.User, style.Reset, style.Border, style.Reset)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("\n%sGoodbye!%s\n", style.Dim, style.Reset)
			return nil
		}
		prompt := strings.TrimSpace(line)
		if prompt == "" {
			continue
		}
		switch prompt {
		case "/exit", "/quit":
			fmt.Printf("%sGoodbye!%s\n", style.Dim, style.Reset)
			return nil
		case "/reset":
			state, err = chatservice.NewState()
			if err != nil {
				return err
			}
			fmt.Printf("%s[reset] Conversation reset.%s\n\n", style.System, style.Reset)
			continue
		}
		fmt.Printf("%scodex%s%s>%s ", style.Bot, style.Reset, style.Border, style.Reset)
		next, err := chatservice.StreamReply(session, state, prompt, model, printChunk)
		if err != nil {
			fmt.Printf("\n%sError: %v%s\n\n", style.Error, err, style.Reset)
			continue
		}
		state = next
		fmt.Print("\n\n")
	}
}
