package models

import "codexass/internal/domain"

func KnownModels() []domain.ModelInfo {
	return []domain.ModelInfo{
		{
			ID:          "gpt-5.3-codex",
			Name:        "GPT-5.3 Codex",
			Family:      "gpt-codex",
			Recommended: true,
			Source:      "known-codex-oauth",
			Notes:       "Best default for Codex OAuth chat in this project.",
		},
		{
			ID:          "gpt-5.4",
			Name:        "GPT-5.4",
			Family:      "gpt-5",
			Recommended: false,
			Source:      "known-codex-oauth",
			Notes:       "Known in Codex OAuth-compatible model sets; account availability may vary.",
		},
		{
			ID:          "gpt-5.2",
			Name:        "GPT-5.2",
			Family:      "gpt-5",
			Recommended: false,
			Source:      "known-codex-oauth",
			Notes:       "Known in Codex OAuth-compatible model sets; account availability may vary.",
		},
		{
			ID:          "gpt-5.2-codex",
			Name:        "GPT-5.2 Codex",
			Family:      "gpt-codex",
			Recommended: false,
			Source:      "known-codex-oauth",
			Notes:       "Known in Codex OAuth-compatible model sets; account availability may vary.",
		},
		{
			ID:          "gpt-5.1-codex",
			Name:        "GPT-5.1 Codex",
			Family:      "gpt-codex",
			Recommended: false,
			Source:      "known-codex-oauth",
			Notes:       "Known in Codex OAuth-compatible model sets; account availability may vary.",
		},
		{
			ID:          "gpt-5.1-codex-mini",
			Name:        "GPT-5.1 Codex Mini",
			Family:      "gpt-codex",
			Recommended: false,
			Source:      "known-codex-oauth",
			Notes:       "Known in Codex OAuth-compatible model sets; account availability may vary.",
		},
		{
			ID:          "gpt-5.1-codex-max",
			Name:        "GPT-5.1 Codex Max",
			Family:      "gpt-codex",
			Recommended: false,
			Source:      "known-codex-oauth",
			Notes:       "Known in Codex OAuth-compatible model sets; account availability may vary.",
		},
	}
}
