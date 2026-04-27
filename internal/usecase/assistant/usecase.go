package assistant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	assistantexternal "github.com/GIT_USER_ID/GIT_REPO_ID/internal/external/http/assistant"
)

var (
	ErrInvalidRequest = errors.New("invalid assistant request")
	ErrUnavailable    = errors.New("assistant is unavailable")
)

const defaultSystemPrompt = `You are an indoor navigation assistant for a university building map.
Always answer in Russian.
Use only the provided map JSON context.
Never invent rooms, ids, floors or routes.
Return ONLY valid JSON object without markdown fences.
Output schema:
{"answer":"short helpful answer in Russian","actions":[{"type":"show_on_map","label":"\u041f\u043e\u0441\u043c\u043e\u0442\u0440\u0435\u0442\u044c \u043d\u0430 \u043a\u0430\u0440\u0442\u0435","mapId":"...","placeId":"...","elementId":"..."}]}
If no exact map target exists, return an empty actions array.`

type LLMClient interface {
	Complete(ctx context.Context, model string, messages []assistantexternal.Message) (string, error)
}

type UseCase struct {
	client          LLMClient
	model           string
	contextMaxChars int
}

func New(client LLMClient, model string, contextMaxChars int) *UseCase {
	if contextMaxChars <= 0 {
		contextMaxChars = 120000
	}

	return &UseCase{
		client:          client,
		model:           strings.TrimSpace(model),
		contextMaxChars: contextMaxChars,
	}
}

func (u *UseCase) Chat(ctx context.Context, request domain.AssistantChatRequest) (domain.AssistantChatResponse, error) {
	query := strings.TrimSpace(request.Query)
	if query == "" {
		return domain.AssistantChatResponse{}, fmt.Errorf("%w: query is required", ErrInvalidRequest)
	}

	if u.client == nil {
		return domain.AssistantChatResponse{}, fmt.Errorf("%w: llm client is not configured", ErrUnavailable)
	}

	contextJSON := compactJSON(request.Context)
	if contextJSON == "" {
		contextJSON = "{}"
	}

	if len(contextJSON) > u.contextMaxChars {
		contextJSON = contextJSON[:u.contextMaxChars] + "...TRUNCATED"
	}

	messages := make([]assistantexternal.Message, 0, len(request.Messages)+2)
	messages = append(messages, assistantexternal.Message{
		Role:    "system",
		Content: defaultSystemPrompt,
	})

	messages = append(messages, normalizeHistory(request.Messages)...)

	finalUserPrompt := strings.Join([]string{
		"User query:",
		query,
		"",
		"Map context JSON:",
		contextJSON,
	}, "\n")

	messages = append(messages, assistantexternal.Message{Role: "user", Content: finalUserPrompt})

	rawContent, err := u.client.Complete(ctx, u.model, messages)
	if err != nil {
		return domain.AssistantChatResponse{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}

	response := parseAssistantResponse(rawContent)
	if strings.TrimSpace(response.Answer) == "" {
		response.Answer = "\u041d\u0435 \u0443\u0434\u0430\u043b\u043e\u0441\u044c \u043f\u043e\u043b\u0443\u0447\u0438\u0442\u044c \u043e\u0442\u0432\u0435\u0442 \u0430\u0441\u0441\u0438\u0441\u0442\u0435\u043d\u0442\u0430."
	}

	response.Actions = sanitizeActions(response.Actions)

	return response, nil
}

func normalizeHistory(history []domain.AssistantMessage) []assistantexternal.Message {
	result := make([]assistantexternal.Message, 0, len(history))

	for _, message := range history {
		role := strings.ToLower(strings.TrimSpace(message.Role))
		switch role {
		case "user", "assistant":
		default:
			continue
		}

		content := strings.TrimSpace(message.Content)
		if content == "" {
			continue
		}

		result = append(result, assistantexternal.Message{Role: role, Content: content})
	}

	return result
}

func compactJSON(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err == nil {
		return compact.String()
	}

	return strings.TrimSpace(string(raw))
}

func parseAssistantResponse(raw string) domain.AssistantChatResponse {
	content := strings.TrimSpace(raw)
	content = trimCodeFence(content)

	var parsed domain.AssistantChatResponse
	if err := json.Unmarshal([]byte(content), &parsed); err == nil {
		return parsed
	}

	return domain.AssistantChatResponse{
		Answer: strings.TrimSpace(raw),
	}
}

func trimCodeFence(content string) string {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}

	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimPrefix(trimmed, "json")
	trimmed = strings.TrimSpace(trimmed)
	trimmed = strings.TrimSuffix(trimmed, "```")

	return strings.TrimSpace(trimmed)
}

func sanitizeActions(actions []domain.AssistantAction) []domain.AssistantAction {
	if len(actions) == 0 {
		return nil
	}

	result := make([]domain.AssistantAction, 0, len(actions))
	for _, action := range actions {
		if strings.TrimSpace(action.Type) != "show_on_map" {
			continue
		}

		action.Label = strings.TrimSpace(action.Label)
		if action.Label == "" {
			action.Label = "\u041f\u043e\u0441\u043c\u043e\u0442\u0440\u0435\u0442\u044c \u043d\u0430 \u043a\u0430\u0440\u0442\u0435"
		}
		action.MapID = strings.TrimSpace(action.MapID)
		action.PlaceID = strings.TrimSpace(action.PlaceID)
		action.ElementID = strings.TrimSpace(action.ElementID)

		if action.MapID == "" && action.PlaceID == "" && action.ElementID == "" {
			continue
		}

		result = append(result, action)
	}

	if len(result) == 0 {
		return nil
	}

	return result
}
