package domain

import "time"

type NewsResponse struct {
	Items []NewsItem `json:"items"`
}

type NewsItem struct {
	Slug         string      `json:"slug"`
	Title        string      `json:"title"`
	PreviewText  string      `json:"preview_text"`
	PublishedAt  PublishedAt `json:"published_at"`
	ImagePreview string      `json:"imagePreview"`
	Tags         []Tag       `json:"tags"`
	PageURL      string      `json:"page_url"`
}

type PublishedAt struct {
	Day   string `json:"day"`
	Month string `json:"month"`
	Year  string `json:"year"`
}

type Tag struct {
	ID    int    `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Color string `json:"color"`
}

type NewsRecord struct {
	Slug          string
	Title         string
	PreviewText   string
	PublishedDate time.Time
	ImagePreview  string
	PageURL       string
	Tags          []Tag
}