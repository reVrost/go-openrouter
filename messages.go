package openrouter

// SystemMessage creates a new system message with the given text content.
func SystemMessage(content string) ChatCompletionMessage {
	return ChatCompletionMessage{
		Role: ChatMessageRoleSystem,
		Content: Content{
			Text: content,
		},
	}
}

// UserMessage creates a new user message with the given text content.
func UserMessage(content string) ChatCompletionMessage {
	return ChatCompletionMessage{
		Role: ChatMessageRoleUser,
		Content: Content{
			Text: content,
		},
	}
}

// AssistantMessage creates a new assistant message with the given text content.
func AssistantMessage(content string) ChatCompletionMessage {
	return ChatCompletionMessage{
		Role: ChatMessageRoleAssistant,
		Content: Content{
			Text: content,
		},
	}
}

// ToolMessage creates a new tool (response) message with a call ID and content.
func ToolMessage(callID string, content string) ChatCompletionMessage {
	return ChatCompletionMessage{
		Role: ChatMessageRoleTool,
		Content: Content{
			Text: content,
		},
		ToolCallID: callID,
	}
}
