package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Node представляет элемент иерархии из schedule_ID.json
type Node struct {
	Abbr     string  `json:"abbr"`
	Name     string  `json:"name"`
	UUID     string  `json:"uuid"`
	Course   int     `json:"course,omitempty"`
	NodeType string  `json:"nodeType,omitempty"`
	Semester int     `json:"semester,omitempty"`
	Parent   string  `json:"parentUuid,omitempty"`
	Children []Node  `json:"children,omitempty"`
}

// Root структура корневого элемента
type Root struct {
	Data Node `json:"data"`
}

func GetJson() {
	inputPath := filepath.Join("schedule", "lksJSON", "schedule_ID.json")
	outputDir := filepath.Join("schedule", "lksJSON")

	// Читаем входной файл
	data, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("read input file: %v", err)
	}

	var root Root
	if err := json.Unmarshal(data, &root); err != nil {
		log.Fatalf("parse JSON: %v", err)
	}

	// Собираем все UUID групп
	var groupUUIDs []string
	collectGroupUUIDs(root.Data, &groupUUIDs)

	log.Printf("found %d groups", len(groupUUIDs))

	// Создаем HTTP-клиент с таймаутами
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Загружаем расписания для каждой группы
	for i, uuid := range groupUUIDs {
		log.Printf("[%d/%d] fetching schedule for %s", i+1, len(groupUUIDs), uuid)

		url := fmt.Sprintf("https://lks.bmstu.ru/lks-back/srv/v2/ics/%s", uuid)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("create request for %s: %v", uuid, err)
			continue
		}

		// Опционально: добавить заголовки, если требуется авторизация
		// req.Header.Set("Authorization", "Bearer ...")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("request %s: %v", uuid, err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("read response for %s: %v", uuid, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("bad status for %s: %d %s", uuid, resp.StatusCode, string(body))
			continue
		}

		// Сохраняем в файл
		outputPath := filepath.Join(outputDir, fmt.Sprintf("schedule_%s.json", uuid))
		if err := os.WriteFile(outputPath, body, 0644); err != nil {
			log.Printf("write file for %s: %v", uuid, err)
			continue
		}

		log.Printf("saved: %s", outputPath)
		time.Sleep(100 * time.Millisecond) // небольшая задержка, чтобы не спамить запросами
	}

	log.Println("done")
}

// collectGroupUUIDs рекурсивно обходит дерево и собирает UUID групп
func collectGroupUUIDs(node Node, result *[]string) {
	if node.NodeType == "group" && node.UUID != "" {
		*result = append(*result, node.UUID)
	}
	for _, child := range node.Children {
		collectGroupUUIDs(child, result)
	}
}