package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStorage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storagePath := filepath.Join(tmpDir, "features.json")

	store, err := New(storagePath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if store == nil {
		t.Error("New() should return non-nil storage")
	}
}

func TestNewStorage_LoadExisting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storagePath := filepath.Join(tmpDir, "features.json")

	// Create initial storage with a feature
	store1, err := New(storagePath)
	if err != nil {
		t.Fatal(err)
	}

	feature := &FeatureState{
		ID:          "test-123",
		Name:        "Test Feature",
		Description: "A test feature",
		Phase:       "implementing",
		CreatedAt:   time.Now(),
	}

	if err := store1.SaveFeature(feature); err != nil {
		t.Fatalf("SaveFeature() error = %v", err)
	}

	// Create new storage instance and verify data is loaded
	store2, err := New(storagePath)
	if err != nil {
		t.Fatalf("New() with existing data error = %v", err)
	}

	loaded, err := store2.LoadFeature("test-123")
	if err != nil {
		t.Fatalf("LoadFeature() error = %v", err)
	}

	if loaded.Name != "Test Feature" {
		t.Errorf("Loaded feature name = %q, want %q", loaded.Name, "Test Feature")
	}
}

func TestSaveAndLoadFeature(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := New(filepath.Join(tmpDir, "features.json"))
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	feature := &FeatureState{
		ID:          "feat-001",
		Name:        "User Auth",
		Description: "Add authentication",
		Branch:      "feature/feat-001-User-Auth",
		Phase:       "planning",
		TechStack:   "Go, PostgreSQL",
		CreatedAt:   now,
		UpdatedAt:   now,
		Tasks: []TaskState{
			{ID: "T-001", Spec: "Create schema", Status: "pending"},
			{ID: "T-002", Spec: "Add endpoints", Status: "pending"},
		},
		Answers: map[string]string{
			"Q1": "Use JWT",
			"Q2": "Yes, email verification",
		},
	}

	// Save
	if err := store.SaveFeature(feature); err != nil {
		t.Fatalf("SaveFeature() error = %v", err)
	}

	// Load
	loaded, err := store.LoadFeature("feat-001")
	if err != nil {
		t.Fatalf("LoadFeature() error = %v", err)
	}

	// Verify fields
	if loaded.ID != feature.ID {
		t.Errorf("ID = %q, want %q", loaded.ID, feature.ID)
	}
	if loaded.Name != feature.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, feature.Name)
	}
	if loaded.Phase != feature.Phase {
		t.Errorf("Phase = %q, want %q", loaded.Phase, feature.Phase)
	}
	if len(loaded.Tasks) != 2 {
		t.Errorf("Tasks count = %d, want 2", len(loaded.Tasks))
	}
	if loaded.Answers["Q1"] != "Use JWT" {
		t.Errorf("Answers[Q1] = %q, want %q", loaded.Answers["Q1"], "Use JWT")
	}
}

func TestLoadAllFeatures(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := New(filepath.Join(tmpDir, "features.json"))
	if err != nil {
		t.Fatal(err)
	}

	// Save multiple features
	features := []*FeatureState{
		{ID: "f1", Name: "Feature 1", CreatedAt: time.Now()},
		{ID: "f2", Name: "Feature 2", CreatedAt: time.Now()},
		{ID: "f3", Name: "Feature 3", CreatedAt: time.Now()},
	}

	for _, f := range features {
		if err := store.SaveFeature(f); err != nil {
			t.Fatal(err)
		}
	}

	// Load all
	loaded, err := store.LoadAllFeatures()
	if err != nil {
		t.Fatalf("LoadAllFeatures() error = %v", err)
	}

	if len(loaded) != 3 {
		t.Errorf("LoadAllFeatures() count = %d, want 3", len(loaded))
	}
}

func TestDeleteFeature(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := New(filepath.Join(tmpDir, "features.json"))
	if err != nil {
		t.Fatal(err)
	}

	// Save a feature
	feature := &FeatureState{ID: "delete-me", Name: "To Delete", CreatedAt: time.Now()}
	if err := store.SaveFeature(feature); err != nil {
		t.Fatal(err)
	}

	// Verify it exists
	_, err = store.LoadFeature("delete-me")
	if err != nil {
		t.Fatal("Feature should exist before deletion")
	}

	// Delete
	if err := store.DeleteFeature("delete-me"); err != nil {
		t.Fatalf("DeleteFeature() error = %v", err)
	}

	// Verify it's gone
	_, err = store.LoadFeature("delete-me")
	if err == nil {
		t.Error("Feature should not exist after deletion")
	}
}

func TestLoadFeature_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := New(filepath.Join(tmpDir, "features.json"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.LoadFeature("nonexistent")
	if err == nil {
		t.Error("LoadFeature() should return error for missing feature")
	}
}

func TestFeatureStateWithTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := New(filepath.Join(tmpDir, "features.json"))
	if err != nil {
		t.Fatal(err)
	}

	feature := &FeatureState{
		ID:        "with-tasks",
		Name:      "Feature with Tasks",
		CreatedAt: time.Now(),
		Tasks: []TaskState{
			{
				ID:         "T-001",
				Spec:       "Task 1",
				Status:     "complete",
				Branch:     "task/T-001",
				AgentName:  "claude-code",
				IsParallel: false,
				Attempt:    2,
				FeatureID:  "with-tasks",
			},
			{
				ID:         "T-002",
				Spec:       "Task 2",
				Status:     "running",
				Branch:     "task/T-002",
				AgentName:  "claude-code",
				IsParallel: true,
				Attempt:    0,
				FeatureID:  "with-tasks",
			},
		},
	}

	if err := store.SaveFeature(feature); err != nil {
		t.Fatal(err)
	}

	loaded, err := store.LoadFeature("with-tasks")
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Tasks) != 2 {
		t.Errorf("Tasks count = %d, want 2", len(loaded.Tasks))
	}

	// Verify task details
	task1 := loaded.Tasks[0]
	if task1.Status != "complete" {
		t.Errorf("Task 1 status = %q, want %q", task1.Status, "complete")
	}
	if task1.Attempt != 2 {
		t.Errorf("Task 1 attempt = %d, want 2", task1.Attempt)
	}

	task2 := loaded.Tasks[1]
	if !task2.IsParallel {
		t.Error("Task 2 should be parallel")
	}
}

func TestStorageUpdatedAt(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := New(filepath.Join(tmpDir, "features.json"))
	if err != nil {
		t.Fatal(err)
	}

	feature := &FeatureState{
		ID:        "update-test",
		Name:      "Test",
		CreatedAt: time.Now(),
	}

	beforeSave := time.Now()
	time.Sleep(10 * time.Millisecond) // Small delay to ensure time difference

	if err := store.SaveFeature(feature); err != nil {
		t.Fatal(err)
	}

	loaded, err := store.LoadFeature("update-test")
	if err != nil {
		t.Fatal(err)
	}

	if loaded.UpdatedAt.Before(beforeSave) {
		t.Error("UpdatedAt should be set when saving")
	}
}
