package auth

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"codexass/internal/common"
	"codexass/internal/config"
	"codexass/internal/domain"
	"codexass/internal/infra/httpclient"
	"codexass/internal/infra/storage"
)

type tokenResponse struct {
	IDToken      string  `json:"id_token"`
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresIn    float64 `json:"expires_in"`
}

func callbackURL() string {
	return fmt.Sprintf("http://localhost:%d%s", config.CallbackPort, config.CallbackPath)
}

func NewPending(alias string) (domain.PendingLogin, error) {
	loginID, err := common.RandomString(16)
	if err != nil {
		return domain.PendingLogin{}, err
	}
	state, err := common.RandomString(24)
	if err != nil {
		return domain.PendingLogin{}, err
	}
	verifier, err := common.RandomString(32)
	if err != nil {
		return domain.PendingLogin{}, err
	}
	return domain.PendingLogin{
		LoginID:      loginID,
		State:        state,
		CodeVerifier: verifier,
		RedirectURI:  callbackURL(),
		Alias:        alias,
		CreatedAt:    common.ISOFormat(common.UTCNow()),
	}, nil
}

func BuildAuthURL(pending domain.PendingLogin, originator string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", config.ClientID)
	params.Set("redirect_uri", pending.RedirectURI)
	params.Set("scope", "openid profile email offline_access")
	params.Set("code_challenge", codeChallenge(pending.CodeVerifier))
	params.Set("code_challenge_method", "S256")
	params.Set("id_token_add_organizations", "true")
	params.Set("codex_cli_simplified_flow", "true")
	params.Set("state", pending.State)
	params.Set("originator", originator)
	return config.AuthEndpoint + "?" + params.Encode()
}

func parseTokens(payload tokenResponse) domain.TokenSet {
	expiresAt := ""
	if payload.ExpiresIn > 0 {
		expiresAt = common.ISOFormat(common.UTCNow().Add(time.Duration(payload.ExpiresIn) * time.Second))
	}
	if expiresAt == "" {
		expiresAt = accessExpiry(payload.AccessToken)
	}
	return domain.TokenSet{
		IDToken:      payload.IDToken,
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		AccountID:    extractAccountID(payload.IDToken, payload.AccessToken),
		ExpiresAt:    expiresAt,
	}
}

func exchangeCode(code string, pending domain.PendingLogin) (domain.TokenSet, error) {
	raw, err := httpclient.RequestFormJSON(config.TokenEndpoint, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {pending.RedirectURI},
		"client_id":     {config.ClientID},
		"code_verifier": {pending.CodeVerifier},
	})
	if err != nil {
		return domain.TokenSet{}, err
	}
	var payload tokenResponse
	if err := json.Unmarshal(raw, &payload); err != nil {
		return domain.TokenSet{}, err
	}
	tokens := parseTokens(payload)
	if tokens.IDToken == "" || tokens.AccessToken == "" {
		return domain.TokenSet{}, fmt.Errorf("token response missing id_token/access_token")
	}
	return tokens, nil
}

func refreshSession(session domain.SessionRecord) (domain.SessionRecord, error) {
	if session.RefreshToken == "" {
		return domain.SessionRecord{}, fmt.Errorf("refresh token missing for %s", session.ID)
	}
	raw, err := httpclient.RequestFormJSON(config.TokenEndpoint, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {session.RefreshToken},
		"client_id":     {config.ClientID},
	})
	if err != nil {
		return domain.SessionRecord{}, err
	}
	var payload tokenResponse
	if err := json.Unmarshal(raw, &payload); err != nil {
		return domain.SessionRecord{}, err
	}
	tokens := parseTokens(payload)
	claims, err := parseClaims(common.FirstNonEmpty(tokens.IDToken, session.IDToken), common.FirstNonEmpty(tokens.AccessToken, session.AccessToken))
	if err != nil {
		return domain.SessionRecord{}, err
	}
	updated := session
	updated.Email = claims.Email
	updated.PlanType = claims.PlanType
	updated.AccountID = common.FirstNonEmpty(tokens.AccountID, session.AccountID, claims.AccountID)
	updated.OrganizationID = common.FirstNonEmpty(claims.OrganizationID, session.OrganizationID)
	updated.IDToken = common.FirstNonEmpty(tokens.IDToken, session.IDToken)
	updated.AccessToken = common.FirstNonEmpty(tokens.AccessToken, session.AccessToken)
	updated.RefreshToken = common.FirstNonEmpty(tokens.RefreshToken, session.RefreshToken)
	updated.ExpiresAt = common.FirstNonEmpty(tokens.ExpiresAt, session.ExpiresAt)
	updated.LastRefreshAt = common.ISOFormat(common.UTCNow())
	return updated, nil
}

func shouldRefresh(session domain.SessionRecord) bool {
	expiresAt, ok := common.ParseTime(session.ExpiresAt)
	if !ok {
		return false
	}
	return time.Until(expiresAt) <= 2*time.Minute
}

func buildSession(tokens domain.TokenSet, alias string) (domain.SessionRecord, error) {
	claims, err := parseClaims(tokens.IDToken, tokens.AccessToken)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	now := common.ISOFormat(common.UTCNow())
	accountID := common.FirstNonEmpty(tokens.AccountID, claims.AccountID)
	return domain.SessionRecord{
		ID:             common.SessionID(claims.Email, accountID, claims.OrganizationID),
		Alias:          alias,
		Email:          claims.Email,
		PlanType:       claims.PlanType,
		AccountID:      accountID,
		OrganizationID: claims.OrganizationID,
		IDToken:        tokens.IDToken,
		AccessToken:    tokens.AccessToken,
		RefreshToken:   tokens.RefreshToken,
		ExpiresAt:      tokens.ExpiresAt,
		Source:         "browser_oauth",
		CreatedAt:      now,
		UpdatedAt:      now,
		LastRefreshAt:  now,
	}, nil
}

func parseCallbackURL(callback string, pending domain.PendingLogin) (string, error) {
	raw := common.FirstNonEmpty(callback)
	if len(raw) >= len(config.CallbackPath) && raw[:len(config.CallbackPath)] == config.CallbackPath {
		raw = "http://localhost" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	state := parsed.Query().Get("state")
	code := parsed.Query().Get("code")
	if state != pending.State {
		return "", fmt.Errorf("state mismatch")
	}
	if code == "" {
		return "", fmt.Errorf("missing code")
	}
	return code, nil
}

func CompleteLogin(store *storage.Store, loginID, callback string) (domain.SessionRecord, error) {
	pending, err := store.LoadPending(loginID)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	code, err := parseCallbackURL(callback, pending)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	tokens, err := exchangeCode(code, pending)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	session, err := buildSession(tokens, pending.Alias)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	saved, err := store.UpsertSession(session, true)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	store.DeletePending(loginID)
	return saved, nil
}

func EnsureFreshSession(store *storage.Store, selector string) (domain.SessionRecord, error) {
	session, err := store.FindSession(selector)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	state, err := store.LoadState()
	if err != nil {
		return domain.SessionRecord{}, err
	}
	if shouldRefresh(session) {
		refreshed, err := refreshSession(session)
		if err != nil {
			return domain.SessionRecord{}, err
		}
		return store.UpsertSession(refreshed, session.ID == state.ActiveSessionID)
	}
	return session, nil
}
