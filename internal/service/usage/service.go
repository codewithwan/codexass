package usage

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"codexass/internal/common"
	"codexass/internal/config"
	"codexass/internal/domain"
	"codexass/internal/infra/httpclient"
)

func classifyWindow(key string, seconds int) string {
	switch {
	case seconds == 18000 || key == "primary_window":
		return "5h"
	case seconds == 604800 || key == "secondary_window":
		return "weekly"
	default:
		return key
	}
}

func windowLabel(seconds int) string {
	if seconds <= 0 {
		return ""
	}
	if seconds%604800 == 0 {
		return fmt.Sprintf("%dw", seconds/604800)
	}
	if seconds%86400 == 0 {
		return fmt.Sprintf("%dd", seconds/86400)
	}
	if seconds%3600 == 0 {
		return fmt.Sprintf("%dh", seconds/3600)
	}
	if seconds%60 == 0 {
		return fmt.Sprintf("%dm", seconds/60)
	}
	return fmt.Sprintf("%ds", seconds)
}

func parseReset(raw common.JSONObject) string {
	if resetAt, ok := common.FloatField(raw, "reset_at"); ok && resetAt > 0 {
		return common.ISOFormat(time.Unix(int64(resetAt), 0).UTC())
	}
	if after, ok := common.FloatField(raw, "reset_after_seconds"); ok && after > 0 {
		return common.ISOFormat(common.UTCNow().Add(time.Duration(after) * time.Second))
	}
	return ""
}

func looksLikeWindow(raw common.JSONObject) bool {
	a := common.HasField(raw, "used_percent")
	b := common.HasField(raw, "reset_at")
	c := common.HasField(raw, "reset_after_seconds")
	d := common.HasField(raw, "limit_window_seconds")
	return a || b || c || d
}

func hasWindowChildren(raw common.JSONObject) bool {
	for _, value := range raw {
		child, ok := common.DecodeObjectRaw(value)
		if ok && looksLikeWindow(child) {
			return true
		}
	}
	return false
}

func parseGroup(group string, raw common.JSONObject) []domain.UsageWindow {
	var windows []domain.UsageWindow
	for key, childRaw := range raw {
		child, ok := common.DecodeObjectRaw(childRaw)
		if !ok || !looksLikeWindow(child) {
			continue
		}
		used, _ := common.FloatField(child, "used_percent")
		seconds, _ := common.FloatField(child, "limit_window_seconds")
		usedPct := int(math.Max(0, math.Min(100, used)))
		rawJSON, err := json.Marshal(child)
		if err != nil {
			rawJSON = json.RawMessage("{}")
		}
		windows = append(windows, domain.UsageWindow{
			Group:        group,
			Key:          key,
			Name:         classifyWindow(key, int(seconds)),
			RemainingPct: max(0, 100-usedPct),
			UsedPct:      usedPct,
			ResetAt:      parseReset(child),
			WindowSecs:   int(seconds),
			WindowLabel:  windowLabel(int(seconds)),
			Raw:          rawJSON,
		})
	}
	return windows
}

func walkWindows(path string, raw json.RawMessage, out *[]domain.UsageWindow) {
	if value, ok := common.DecodeObjectRaw(raw); ok {
		if hasWindowChildren(value) {
			*out = append(*out, parseGroup(common.FirstNonEmpty(path, "rate_limit"), value)...)
		}
		for key, child := range value {
			next := key
			if path != "" {
				next = path + "." + key
			}
			walkWindows(next, child, out)
		}
		return
	}
	if value, ok := common.DecodeArrayRaw(raw); ok {
		for _, child := range value {
			walkWindows(path, child, out)
		}
	}
}

func Fetch(session domain.SessionRecord) (domain.UsageSnapshot, error) {
	headers := map[string]string{"Authorization": "Bearer " + session.AccessToken}
	if session.AccountID != "" {
		headers["ChatGPT-Account-Id"] = session.AccountID
	}
	raw, err := httpclient.RequestNoBodyJSON(config.UsageURL, headers)
	if err != nil {
		return domain.UsageSnapshot{}, err
	}
	var windows []domain.UsageWindow
	walkWindows("", raw, &windows)
	sort.Slice(windows, func(i, j int) bool {
		return windows[i].WindowSecs < windows[j].WindowSecs
	})
	return domain.UsageSnapshot{
		AccountID: session.AccountID,
		FetchedAt: common.ISOFormat(common.UTCNow()),
		RawJSON:   string(raw),
		Windows:   windows,
	}, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
