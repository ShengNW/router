package model

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBuildDefaultChannelTypeCatalog(t *testing.T) {
	items := buildDefaultChannelTypeCatalog(1700000000)
	if len(items) == 0 {
		t.Fatalf("expected default channel types, got empty")
	}
	foundOpenAI := false
	foundOpenAICompatible := false
	for _, item := range items {
		if item.ID == 1 && item.Name == "openai" && item.Label == "OpenAI" {
			foundOpenAI = true
		}
		if item.ID == 50 && item.Name == "openai-compatible" {
			foundOpenAICompatible = true
		}
	}
	if !foundOpenAI {
		t.Fatalf("expected openai channel type in defaults")
	}
	if !foundOpenAICompatible {
		t.Fatalf("expected openai-compatible channel type in defaults")
	}
}

func TestRunChannelTypeCatalogMigrationsWithDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := runChannelTypeCatalogMigrationsWithDB(db); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if err := runChannelTypeCatalogMigrationsWithDB(db); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	var count int64
	if err := db.Model(&ChannelTypeCatalog{}).Count(&count).Error; err != nil {
		t.Fatalf("count channel types failed: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected seeded channel types, got 0")
	}
}
