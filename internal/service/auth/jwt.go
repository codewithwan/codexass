package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"codexass/internal/common"
	"codexass/internal/domain"
)

type jwtOrganization struct {
	ID string `json:"id"`
}

type jwtAuthClaims struct {
	ChatGPTAccountID string `json:"chatgpt_account_id"`
	AccountID        string `json:"account_id"`
	OrganizationID   string `json:"organization_id"`
	PlanType         string `json:"chatgpt_plan_type"`
}

type jwtClaims struct {
	Email            string            `json:"email"`
	Exp              int64             `json:"exp"`
	ChatGPTAccountID string            `json:"chatgpt_account_id"`
	Organizations    []jwtOrganization `json:"organizations"`
	Auth             jwtAuthClaims     `json:"https://api.openai.com/auth"`
}

func codeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func decodeJWTPayload(token string) (jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return jwtClaims{}, fmt.Errorf("invalid jwt")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return jwtClaims{}, err
	}
	var out jwtClaims
	if err := json.Unmarshal(raw, &out); err != nil {
		return jwtClaims{}, err
	}
	return out, nil
}

func extractAccountID(idToken, accessToken string) string {
	for _, token := range []string{idToken, accessToken} {
		if token == "" {
			continue
		}
		payload, err := decodeJWTPayload(token)
		if err != nil {
			continue
		}
		if value := strings.TrimSpace(payload.ChatGPTAccountID); value != "" {
			return value
		}
		if value := strings.TrimSpace(payload.Auth.ChatGPTAccountID); value != "" {
			return value
		}
		if len(payload.Organizations) > 0 {
			if value := strings.TrimSpace(payload.Organizations[0].ID); value != "" {
				return value
			}
		}
	}
	return ""
}

func accessExpiry(accessToken string) string {
	payload, err := decodeJWTPayload(accessToken)
	if err != nil {
		return ""
	}
	if payload.Exp <= 0 {
		return ""
	}
	return common.ISOFormat(time.Unix(payload.Exp, 0).UTC())
}

func parseClaims(idToken, accessToken string) (domain.Claims, error) {
	idPayload, err := decodeJWTPayload(idToken)
	if err != nil {
		return domain.Claims{}, err
	}
	accessPayload, _ := decodeJWTPayload(accessToken)
	email := strings.TrimSpace(idPayload.Email)
	if email == "" {
		return domain.Claims{}, fmt.Errorf("email not found in id_token")
	}
	claims := domain.Claims{
		Email:          email,
		PlanType:       strings.TrimSpace(idPayload.Auth.PlanType),
		AccountID:      strings.TrimSpace(idPayload.Auth.AccountID),
		OrganizationID: strings.TrimSpace(idPayload.Auth.OrganizationID),
	}
	if claims.AccountID == "" {
		claims.AccountID = strings.TrimSpace(accessPayload.Auth.ChatGPTAccountID)
	}
	if claims.OrganizationID == "" {
		claims.OrganizationID = strings.TrimSpace(accessPayload.Auth.OrganizationID)
	}
	return claims, nil
}
