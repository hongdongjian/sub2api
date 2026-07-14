package domain

// OpenAIRequestOverrides contains group-level request parameter overrides for
// OpenAI gateway requests. Empty fields are ignored by the service layer.
type OpenAIRequestOverrides struct {
	ServiceTier     string `json:"service_tier,omitempty"`
	ReasoningEffort string `json:"reasoning_effort,omitempty"`
	TextVerbosity   string `json:"text_verbosity,omitempty"`
}
