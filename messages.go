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

// UserMessage creates a new user message with the given content.
func UserMessage(content Content) ChatCompletionMessage {
	return ChatCompletionMessage{
		Role:    ChatMessageRoleUser,
		Content: content,
	}
}

// AssistantMessage creates a new assistant message with content and optional tool calls.
func AssistantMessage(content Content, toolCalls []ToolCall) ChatCompletionMessage {
	return ChatCompletionMessage{
		Role:      ChatMessageRoleAssistant,
		Content:   content,
		ToolCalls: toolCalls,
	}
}

// ToolMessage creates a new tool (response) message with a call ID and content.
func ToolMessage(callID string, content Content) ChatCompletionMessage {
	return ChatCompletionMessage{
		Role:       ChatMessageRoleTool,
		Content:    content,
		ToolCallID: callID,
	}
}
