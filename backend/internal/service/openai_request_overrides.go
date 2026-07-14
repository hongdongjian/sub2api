package service

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func normalizeOpenAIRequestOverrides(in OpenAIRequestOverrides) OpenAIRequestOverrides {
	return OpenAIRequestOverrides{
		ServiceTier:     normalizedOpenAIRequestOverrideServiceTier(in.ServiceTier),
		ReasoningEffort: normalizeOpenAIRequestOverrideReasoningEffort(in.ReasoningEffort),
		TextVerbosity:   normalizeOpenAITextVerbosity(in.TextVerbosity),
	}
}

func normalizeOpenAIRequestOverrideReasoningEffort(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "max") {
		return "max"
	}
	return normalizeOpenAIReasoningEffort(raw)
}

func normalizedOpenAIRequestOverrideServiceTier(raw string) string {
	normalized := normalizeOpenAIServiceTier(raw)
	if normalized == nil {
		return ""
	}
	return *normalized
}

func normalizeOpenAITextVerbosity(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "low", "medium", "high":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func groupOpenAIRequestOverrides(group *Group) OpenAIRequestOverrides {
	if group == nil || group.Platform != PlatformOpenAI {
		return OpenAIRequestOverrides{}
	}
	return normalizeOpenAIRequestOverrides(group.OpenAIRequestOverrides)
}

func openAIRequestOverridesIsZero(o OpenAIRequestOverrides) bool {
	return strings.TrimSpace(o.ServiceTier) == "" &&
		strings.TrimSpace(o.ReasoningEffort) == "" &&
		strings.TrimSpace(o.TextVerbosity) == ""
}

func applyOpenAIRequestOverridesToMap(reqBody map[string]any, overrides OpenAIRequestOverrides, upstreamModel string) bool {
	if reqBody == nil {
		return false
	}
	overrides = normalizeOpenAIRequestOverrides(overrides)
	if openAIRequestOverridesIsZero(overrides) {
		return false
	}

	changed := false
	if overrides.ServiceTier != "" {
		if reqBody["service_tier"] != overrides.ServiceTier {
			reqBody["service_tier"] = overrides.ServiceTier
			changed = true
		}
	}
	if overrides.ReasoningEffort != "" {
		reasoning, _ := reqBody["reasoning"].(map[string]any)
		if reasoning == nil {
			reasoning = map[string]any{}
			reqBody["reasoning"] = reasoning
			changed = true
		}
		if reasoning["effort"] != overrides.ReasoningEffort {
			reasoning["effort"] = overrides.ReasoningEffort
			changed = true
		}
	}
	if overrides.TextVerbosity != "" && SupportsVerbosity(upstreamModel) {
		text, _ := reqBody["text"].(map[string]any)
		if text == nil {
			text = map[string]any{}
			reqBody["text"] = text
			changed = true
		}
		if text["verbosity"] != overrides.TextVerbosity {
			text["verbosity"] = overrides.TextVerbosity
			changed = true
		}
	}
	return changed
}

func applyOpenAIRequestOverridesToBody(body []byte, overrides OpenAIRequestOverrides, upstreamModel string) ([]byte, bool, error) {
	if len(body) == 0 {
		return body, false, nil
	}
	overrides = normalizeOpenAIRequestOverrides(overrides)
	if openAIRequestOverridesIsZero(overrides) {
		return body, false, nil
	}

	updated := body
	changed := false
	var err error
	if overrides.ServiceTier != "" {
		if gjson.GetBytes(updated, "service_tier").String() != overrides.ServiceTier {
			updated, err = sjson.SetBytes(updated, "service_tier", overrides.ServiceTier)
			if err != nil {
				return body, false, fmt.Errorf("set openai request override service_tier: %w", err)
			}
			changed = true
		}
	}
	if overrides.ReasoningEffort != "" {
		if gjson.GetBytes(updated, "reasoning.effort").String() != overrides.ReasoningEffort {
			updated, err = sjson.SetBytes(updated, "reasoning.effort", overrides.ReasoningEffort)
			if err != nil {
				return body, false, fmt.Errorf("set openai request override reasoning.effort: %w", err)
			}
			changed = true
		}
	}
	if overrides.TextVerbosity != "" && SupportsVerbosity(upstreamModel) {
		if gjson.GetBytes(updated, "text.verbosity").String() != overrides.TextVerbosity {
			updated, err = sjson.SetBytes(updated, "text.verbosity", overrides.TextVerbosity)
			if err != nil {
				return body, false, fmt.Errorf("set openai request override text.verbosity: %w", err)
			}
			changed = true
		}
	}
	return updated, changed, nil
}

func applyOpenAIRequestOverridesToChatCompletionsBody(body []byte, overrides OpenAIRequestOverrides) ([]byte, bool, error) {
	if len(body) == 0 {
		return body, false, nil
	}
	overrides = normalizeOpenAIRequestOverrides(overrides)
	if overrides.ServiceTier == "" && overrides.ReasoningEffort == "" {
		return body, false, nil
	}

	updated := body
	changed := false
	var err error
	if overrides.ServiceTier != "" {
		if gjson.GetBytes(updated, "service_tier").String() != overrides.ServiceTier {
			updated, err = sjson.SetBytes(updated, "service_tier", overrides.ServiceTier)
			if err != nil {
				return body, false, fmt.Errorf("set openai chat override service_tier: %w", err)
			}
			changed = true
		}
	}
	if overrides.ReasoningEffort != "" {
		if gjson.GetBytes(updated, "reasoning_effort").String() != overrides.ReasoningEffort {
			updated, err = sjson.SetBytes(updated, "reasoning_effort", overrides.ReasoningEffort)
			if err != nil {
				return body, false, fmt.Errorf("set openai chat override reasoning_effort: %w", err)
			}
			changed = true
		}
	}
	return updated, changed, nil
}

func applyOpenAIRequestOverridesToWSResponseCreate(frame []byte, overrides OpenAIRequestOverrides, upstreamModel string) ([]byte, bool, error) {
	if len(frame) == 0 || !gjson.ValidBytes(frame) {
		return frame, false, nil
	}
	if strings.TrimSpace(gjson.GetBytes(frame, "type").String()) != "response.create" {
		return frame, false, nil
	}
	return applyOpenAIRequestOverridesToBody(frame, overrides, upstreamModel)
}
