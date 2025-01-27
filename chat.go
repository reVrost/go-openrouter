package openrouter

type Hate struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity,omitempty"`
}
type SelfHarm struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity,omitempty"`
}
type Sexual struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity,omitempty"`
}
type Violence struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity,omitempty"`
}

type JailBreak struct {
	Filtered bool `json:"filtered"`
	Detected bool `json:"detected"`
}

type Profanity struct {
	Filtered bool `json:"filtered"`
	Detected bool `json:"detected"`
}

type ContentFilterResults struct {
	Hate      Hate      `json:"hate,omitempty"`
	SelfHarm  SelfHarm  `json:"self_harm,omitempty"`
	Sexual    Sexual    `json:"sexual,omitempty"`
	Violence  Violence  `json:"violence,omitempty"`
	JailBreak JailBreak `json:"jailbreak,omitempty"`
	Profanity Profanity `json:"profanity,omitempty"`
}
