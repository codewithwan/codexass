package domain

import "encoding/json"

type Claims struct {
	Email          string `json:"email"`
	PlanType       string `json:"plan_type,omitempty"`
	AccountID      string `json:"account_id,omitempty"`
	OrganizationID string `json:"organization_id,omitempty"`
}

type PendingLogin struct {
	LoginID      string `json:"login_id"`
	State        string `json:"state"`
	CodeVerifier string `json:"code_verifier"`
	RedirectURI  string `json:"redirect_uri"`
	Alias        string `json:"alias,omitempty"`
	CreatedAt    string `json:"created_at"`
}

type SessionRecord struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	IDToken          string `json:"id_token"`
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	Alias            string `json:"alias,omitempty"`
	PlanType         string `json:"plan_type,omitempty"`
	AccountID        string `json:"account_id,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	ExpiresAt        string `json:"expires_at,omitempty"`
	Source           string `json:"source,omitempty"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	LastRefreshAt    string `json:"last_refresh_at,omitempty"`
	LastUsageCheckAt string `json:"last_usage_check_at,omitempty"`
}

func (s SessionRecord) DisplayName() string {
	if s.Alias != "" {
		return s.Alias
	}
	if s.Email != "" {
		return s.Email
	}
	return s.ID
}

type StateFile struct {
	Version         int             `json:"version"`
	ActiveSessionID string          `json:"active_session_id,omitempty"`
	Sessions        []SessionRecord `json:"sessions"`
}

type TokenSet struct {
	IDToken      string
	AccessToken  string
	RefreshToken string
	AccountID    string
	ExpiresAt    string
}

type UsageWindow struct {
	Group        string          `json:"group"`
	Key          string          `json:"key"`
	Name         string          `json:"name"`
	RemainingPct int             `json:"remaining_pct"`
	UsedPct      int             `json:"used_pct"`
	ResetAt      string          `json:"reset_at,omitempty"`
	WindowSecs   int             `json:"window_seconds"`
	WindowLabel  string          `json:"window_label,omitempty"`
	Raw          json.RawMessage `json:"raw"`
}

type UsageSnapshot struct {
	AccountID string        `json:"account_id,omitempty"`
	FetchedAt string        `json:"fetched_at"`
	RawJSON   string        `json:"raw_json"`
	Windows   []UsageWindow `json:"windows"`
}

type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Family      string `json:"family"`
	Recommended bool   `json:"recommended"`
	Source      string `json:"source"`
	Notes       string `json:"notes,omitempty"`
}

type ChatContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ChatMessage struct {
	Role    string            `json:"role"`
	Content []ChatContentPart `json:"content"`
}

type SessionListRow struct {
	Active    bool   `json:"active"`
	ID        string `json:"id"`
	Alias     string `json:"alias,omitempty"`
	Email     string `json:"email"`
	PlanType  string `json:"plan_type,omitempty"`
	AccountID string `json:"account_id,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}
