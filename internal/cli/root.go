package cli

import (
	"flag"
	"fmt"
	"time"

	"codexass/internal/common"
	"codexass/internal/config"
	"codexass/internal/domain"
	"codexass/internal/infra/oauth"
	"codexass/internal/infra/storage"
	authservice "codexass/internal/service/auth"
	modelservice "codexass/internal/service/models"
	usageservice "codexass/internal/service/usage"
	"codexass/internal/ui"
)

func Run(args []string) error {
	if len(args) == 0 {
		printRootHelp()
		return nil
	}
	storeDir, stripped := peelStoreDir(args)
	store, err := storage.New(storeDir)
	if err != nil {
		return err
	}
	switch stripped[0] {
	case "login":
		return runLogin(store, stripped[1:])
	case "complete":
		return runComplete(store, stripped[1:])
	case "usage":
		return runUsage(store, stripped[1:])
	case "list":
		return runList(store)
	case "chat":
		return runChat(store, stripped[1:])
	case "models":
		return runModels(stripped[1:])
	default:
		printRootHelp()
		return fmt.Errorf("unknown command %q", stripped[0])
	}
}

func peelStoreDir(args []string) (string, []string) {
	if len(args) >= 2 && args[0] == config.StoreDirFlag {
		return args[1], args[2:]
	}
	return "", args
}

func runLogin(store *storage.Store, args []string) error {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	alias := fs.String("alias", "", "optional session alias")
	originator := fs.String("originator", config.DefaultOriginator, "oauth originator")
	timeout := fs.Int("timeout", 180, "seconds to wait for callback")
	openBrowser := fs.Bool("open-browser", false, "open browser automatically")
	if err := fs.Parse(args); err != nil {
		return err
	}
	pending, err := authservice.NewPending(*alias)
	if err != nil {
		return err
	}
	if err := store.SavePending(pending); err != nil {
		return err
	}
	authURL := authservice.BuildAuthURL(pending, *originator)
	fmt.Println("Login ID :", pending.LoginID)
	fmt.Println("Auth URL  :", authURL)
	fmt.Println("")
	fmt.Println("Manual login mode is the default.")
	fmt.Println("Type 'c' then Enter to copy the URL to clipboard, or just open it manually.")
	fmt.Println("Use --open-browser if you explicitly want auto-open.")
	if *openBrowser {
		_ = common.OpenBrowser(authURL)
	} else {
		promptCopy(authURL)
	}
	server := oauth.NewCallbackServer(pending.State)
	defer server.Close()
	if err := server.Start(); err != nil {
		return err
	}
	callback, err := server.Wait(time.Duration(*timeout) * time.Second)
	if err != nil {
		return err
	}
	session, err := authservice.CompleteLogin(store, pending.LoginID, callback)
	if err != nil {
		return err
	}
	return printJSON(session)
}

func runComplete(store *storage.Store, args []string) error {
	fs := flag.NewFlagSet("complete", flag.ContinueOnError)
	loginID := fs.String("login-id", "", "pending login id")
	callbackURL := fs.String("callback-url", "", "callback URL")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *loginID == "" || *callbackURL == "" {
		return fmt.Errorf("--login-id and --callback-url are required")
	}
	session, err := authservice.CompleteLogin(store, *loginID, *callbackURL)
	if err != nil {
		return err
	}
	return printJSON(session)
}

func runUsage(store *storage.Store, args []string) error {
	fs := flag.NewFlagSet("usage", flag.ContinueOnError)
	selector := fs.String("session", "", "session id, alias, or email")
	jsonMode := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	session, err := authservice.EnsureFreshSession(store, *selector)
	if err != nil {
		return err
	}
	snapshot, err := usageservice.Fetch(session)
	if err != nil {
		return err
	}
	session.LastUsageCheckAt = snapshot.FetchedAt
	state, _ := store.LoadState()
	_, _ = store.UpsertSession(session, session.ID == state.ActiveSessionID)
	if *jsonMode {
		return printJSON(snapshot)
	}
	fmt.Println("Session    :", session.DisplayName())
	fmt.Println("Account ID :", common.FirstNonEmpty(session.AccountID, "-"))
	fmt.Println("Plan       :", common.FirstNonEmpty(session.PlanType, "-"))
	fmt.Println("Fetched At :", snapshot.FetchedAt)
	fmt.Println()
	if len(snapshot.Windows) == 0 {
		fmt.Println("No quota windows found in usage response.")
		return nil
	}
	for _, item := range snapshot.Windows {
		fmt.Printf("- %s: remaining=%d%% used=%d%% window=%s reset_at=%s\n", common.FirstNonEmpty(item.Name, item.Key), item.RemainingPct, item.UsedPct, common.FirstNonEmpty(item.WindowLabel, "-"), common.FirstNonEmpty(item.ResetAt, "-"))
	}
	return nil
}

func runList(store *storage.Store) error {
	sessions, activeID, err := store.ListSessions()
	if err != nil {
		return err
	}
	rows := make([]domain.SessionListRow, 0, len(sessions))
	for _, session := range sessions {
		rows = append(rows, domain.SessionListRow{
			Active:    session.ID == activeID,
			ID:        session.ID,
			Alias:     session.Alias,
			Email:     session.Email,
			PlanType:  session.PlanType,
			AccountID: session.AccountID,
			ExpiresAt: session.ExpiresAt,
		})
	}
	return printJSON(rows)
}

func runChat(store *storage.Store, args []string) error {
	fs := flag.NewFlagSet("chat", flag.ContinueOnError)
	selector := fs.String("session", "", "session id, alias, or email")
	model := fs.String("model", config.DefaultModel, "chat model")
	if err := fs.Parse(args); err != nil {
		return err
	}
	session, err := authservice.EnsureFreshSession(store, *selector)
	if err != nil {
		return err
	}
	fmt.Println("Using session:", session.DisplayName())
	return ui.RunChat(session, *model)
}

func runModels(args []string) error {
	fs := flag.NewFlagSet("models", flag.ContinueOnError)
	jsonMode := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	models := modelservice.KnownModels()
	if *jsonMode {
		return printJSON(models)
	}
	fmt.Println("Known Codex OAuth-compatible models")
	fmt.Println("")
	for _, model := range models {
		tag := ""
		if model.Recommended {
			tag = " (recommended)"
		}
		fmt.Printf("- %s%s\n  family: %s\n  source: %s\n", model.ID, tag, model.Family, model.Source)
		if model.Notes != "" {
			fmt.Printf("  notes : %s\n", model.Notes)
		}
	}
	return nil
}
