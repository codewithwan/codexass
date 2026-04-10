package chat

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"codexass/internal/common"
	"codexass/internal/config"
	"codexass/internal/domain"
	"codexass/internal/infra/httpclient"
)

const defaultInstructions = "You are Codex running in terminal chat mode. Be a helpful general-purpose AI assistant."

type State struct {
	SessionID string
	History   []domain.ChatMessage
}

type requestPayload struct {
	Model        string               `json:"model"`
	Input        []domain.ChatMessage `json:"input"`
	Instructions string               `json:"instructions"`
	Stream       bool                 `json:"stream"`
	Store        bool                 `json:"store"`
}

type incompleteDetails struct {
	Reason string `json:"reason"`
}

type responseDetails struct {
	IncompleteDetails *incompleteDetails `json:"incomplete_details,omitempty"`
}

type streamEvent struct {
	Type     string           `json:"type"`
	Delta    string           `json:"delta,omitempty"`
	Message  string           `json:"message,omitempty"`
	Response *responseDetails `json:"response,omitempty"`
}

func NewState() (State, error) {
	id, err := common.RandomString(12)
	if err != nil {
		return State{}, err
	}
	return State{SessionID: id, History: []domain.ChatMessage{}}, nil
}

func buildPayload(message string, state State, model string) requestPayload {
	history := append([]domain.ChatMessage{}, state.History...)
	history = append(history, domain.ChatMessage{
		Role: "user",
		Content: []domain.ChatContentPart{{
			Type: "input_text",
			Text: message,
		}},
	})
	return requestPayload{
		Model:        model,
		Input:        history,
		Instructions: defaultInstructions,
		Stream:       true,
		Store:        false,
	}
}

func buildHeaders(session domain.SessionRecord, state State) map[string]string {
	headers := map[string]string{
		"Authorization": "Bearer " + session.AccessToken,
		"originator":    config.DefaultOriginator,
		"session_id":    state.SessionID,
		"User-Agent":    "codexass/0.1 (go terminal chat)",
	}
	if session.AccountID != "" {
		headers["ChatGPT-Account-Id"] = session.AccountID
	}
	return headers
}

func StreamReply(session domain.SessionRecord, state State, prompt, model string, onText func(string)) (State, error) {
	next := State{SessionID: state.SessionID, History: append([]domain.ChatMessage{}, state.History...)}
	var output strings.Builder
	var failed error
	payloadBytes, err := json.Marshal(buildPayload(prompt, state, model))
	if err != nil {
		return State{}, err
	}
	err = httpclient.StreamSSEJSON(config.CodexResponsesURL, payloadBytes, buildHeaders(session, state), 5*time.Minute, func(event []byte) error {
		var payload streamEvent
		if err := json.Unmarshal(event, &payload); err != nil {
			return nil
		}
		switch payload.Type {
		case "response.output_text.delta":
			if payload.Delta != "" {
				onText(payload.Delta)
				output.WriteString(payload.Delta)
			}
		case "error":
			if strings.TrimSpace(payload.Message) != "" {
				failed = fmt.Errorf(strings.TrimSpace(payload.Message))
			}
		case "response.incomplete":
			reason := ""
			if payload.Response != nil && payload.Response.IncompleteDetails != nil {
				reason = payload.Response.IncompleteDetails.Reason
			}
			if strings.TrimSpace(reason) != "" {
				failed = fmt.Errorf("response incomplete: %s", strings.TrimSpace(reason))
			}
		}
		return nil
	})
	if err != nil {
		return State{}, err
	}
	if failed != nil {
		return State{}, failed
	}
	next.History = append(next.History,
		domain.ChatMessage{Role: "user", Content: []domain.ChatContentPart{{Type: "input_text", Text: prompt}}},
		domain.ChatMessage{Role: "assistant", Content: []domain.ChatContentPart{{Type: "output_text", Text: output.String()}}},
	)
	return next, nil
}
