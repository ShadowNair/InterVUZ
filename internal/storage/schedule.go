package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type StubScheduleRepository struct {
	basePath   string 
	groupCatalog domain.GroupCatalogResponse
	groupData    map[string]domain.GroupScheduleResponse
	eventData    map[string]domain.ScheduleEventResponse
}

func hasGroupInCatalog(node domain.GroupCatalogNode, groupID string) bool {
	if node.NodeType == "group" && node.UUID == groupID {
		return true
	}

	for _, child := range node.Children {
		if hasGroupInCatalog(child, groupID) {
			return true
		}
	}

	return false
}

func pruneCatalogToKnownGroups(node domain.GroupCatalogNode, validGroupIDs map[string]struct{}) (domain.GroupCatalogNode, bool) {
	prunedChildren := make([]domain.GroupCatalogNode, 0, len(node.Children))
	for _, child := range node.Children {
		prunedChild, keep := pruneCatalogToKnownGroups(child, validGroupIDs)
		if keep {
			prunedChildren = append(prunedChildren, prunedChild)
		}
	}
	node.Children = prunedChildren

	if node.NodeType == "group" {
		_, keep := validGroupIDs[node.UUID]
		return node, keep
	}

	// Keep non-group nodes only when they still contain at least one valid group.
	if len(node.Children) > 0 {
		return node, true
	}

	return node, false
}

func NewStubScheduleRepository(dataDir string) (*StubScheduleRepository, error) {
	groupCatalogPath := filepath.Join(dataDir, "schedule_ID.json")
	groupSchedulePath := filepath.Join(dataDir, "schedule_group.json")

	groupCatalogBytes, err := os.ReadFile(groupCatalogPath)
	if err != nil {
		return nil, fmt.Errorf("read group catalog stub: %w", err)
	}

	groupScheduleBytes, err := os.ReadFile(groupSchedulePath)
	if err != nil {
		return nil, fmt.Errorf("read group schedule stub: %w", err)
	}

	var catalog domain.GroupCatalogResponse
	if err := json.Unmarshal(groupCatalogBytes, &catalog); err != nil {
		return nil, fmt.Errorf("decode group catalog stub: %w", err)
	}

	var groupSchedule domain.GroupScheduleResponse
	if err := json.Unmarshal(groupScheduleBytes, &groupSchedule); err != nil {
		return nil, fmt.Errorf("decode group schedule stub: %w", err)
	}

	repo := &StubScheduleRepository{
		basePath: dataDir,
		groupCatalog: catalog,
		groupData: map[string]domain.GroupScheduleResponse{
			groupSchedule.Data.UUID: groupSchedule,
		},
		eventData: make(map[string]domain.ScheduleEventResponse),
	}

	for index, event := range groupSchedule.Data.Schedule {
		eventID := fmt.Sprintf("event-%d", index+1)
		repo.eventData[eventID] = domain.ScheduleEventResponse{Data: event}
	}

	// The stub catalog and stub schedule files can be out of sync.
	// Ensure every group with available schedule is present in group catalog.
	for groupID, schedule := range repo.groupData {
		if hasGroupInCatalog(repo.groupCatalog.Data, groupID) {
			continue
		}

		repo.groupCatalog.Data.Children = append(
			repo.groupCatalog.Data.Children,
			domain.GroupCatalogNode{
				Abbr:       schedule.Data.Title,
				Name:       schedule.Data.Title,
				UUID:       groupID,
				NodeType:   "group",
				ParentUUID: repo.groupCatalog.Data.UUID,
			},
		)
	}

	validGroupIDs := make(map[string]struct{}, len(repo.groupData))
	for groupID := range repo.groupData {
		validGroupIDs[groupID] = struct{}{}
	}

	if prunedRoot, ok := pruneCatalogToKnownGroups(repo.groupCatalog.Data, validGroupIDs); ok {
		repo.groupCatalog.Data = prunedRoot
	}

	return repo, nil
}

func (r *StubScheduleRepository) Import(ctx context.Context, request domain.ScheduleImportRequest) (*domain.ScheduleImportResult, error) {
	inputPath := filepath.Join(r.basePath, "schedule_ID.json")
	
	// 1. Читаем иерархию групп
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("read hierarchy file: %w", err)
	}

	var root domain.GroupCatalogResponse
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse hierarchy JSON: %w", err)
	}

	// 2. Собираем UUID групп
	var groupUUIDs []string
	collectGroupUUIDs(root.Data, &groupUUIDs)

	client := &http.Client{Timeout: 30 * time.Second}
	importedCount := 0

	for _, uuid := range groupUUIDs {
		// Проверяем отмену контекста
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if err := r.fetchAndSaveSchedule(ctx, client, uuid); err != nil {
			// Логируем, но продолжаем загрузку остальных
			// В продакшене: r.logger.Warn("failed to fetch", "uuid", uuid, "err", err)
			fmt.Printf("failed to fetch %s: %v\n", uuid, err)
			continue
		}
		importedCount++
		time.Sleep(100 * time.Millisecond) // rate limiting
	}

	return &domain.ScheduleImportResult{
		ImportID:       fmt.Sprintf("import-%d", time.Now().Unix()),
		Status:         "completed",
		ImportedEvents: importedCount,
	}, nil
}

func (r *StubScheduleRepository) GetGroupCatalog(_ context.Context) (*domain.GroupCatalogResponse, error) {
	response := r.groupCatalog
	return &response, nil
}

func (r *StubScheduleRepository) GetGroupSchedule(_ context.Context, groupID string) (*domain.GroupScheduleResponse, error) {
	schedule, ok := r.groupData[groupID]
	if !ok {
		return nil, domain.ErrNotFound
	}

	response := schedule
	return &response, nil
}

func (r *StubScheduleRepository) GetEvent(_ context.Context, eventID string) (*domain.ScheduleEventResponse, error) {
	event, ok := r.eventData[eventID]
	if !ok {
		return nil, domain.ErrNotFound
	}

	response := event
	return &response, nil
}

func collectGroupUUIDs(node domain.GroupCatalogNode, result *[]string) {
	if node.NodeType == "group" && node.UUID != "" {
		*result = append(*result, node.UUID)
	}
	for _, child := range node.Children {
		collectGroupUUIDs(child, result)
	}
}

// fetchAndSaveSchedule загружает расписание по UUID и сохраняет в файл
func (r *StubScheduleRepository) fetchAndSaveSchedule(ctx context.Context, client *http.Client, uuid string) error {
	url := fmt.Sprintf("https://lks.bmstu.ru/lks-back/srv/v2/ics/%s", uuid)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Если нужна авторизация — раскомментируйте и добавьте токен
	// req.Header.Set("Authorization", "Bearer YOUR_TOKEN_HERE")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bad status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Сохраняем в файл schedule_{uuid}.json
	outputPath := filepath.Join(r.basePath, fmt.Sprintf("schedule_%s.json", uuid))
	if err := os.WriteFile(outputPath, body, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// Опционально: обновляем in-memory кеш, чтобы новые данные были доступны сразу
	var schedule domain.GroupScheduleResponse
	if err := json.Unmarshal(body, &schedule); err == nil {
		r.groupData[uuid] = schedule
		// Обновляем eventData для новых событий
		for index, event := range schedule.Data.Schedule {
			eventID := fmt.Sprintf("event-%s-%d", uuid, index+1)
			r.eventData[eventID] = domain.ScheduleEventResponse{Data: event}
		}
	}

	return nil
}