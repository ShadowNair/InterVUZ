package domain

import "encoding/json"

type AssistantMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AssistantAction struct {
	Type      string `json:"type"`
	Label     string `json:"label,omitempty"`
	MapID     string `json:"mapId,omitempty"`
	PlaceID   string `json:"placeId,omitempty"`
	ElementID string `json:"elementId,omitempty"`
}

type AssistantChatRequest struct {
	Query        string             `json:"query"`
	Messages     []AssistantMessage `json:"messages,omitempty"`
	Context      json.RawMessage    `json:"context,omitempty"`
	SystemPrompt string             `json:"systemPrompt,omitempty"`
}

type AssistantChatResponse struct {
	Answer  string            `json:"answer"`
	Actions []AssistantAction `json:"actions,omitempty"`
}
