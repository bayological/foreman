package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// FeatureState represents a feature's persisted state
type FeatureState struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Branch      string            `json:"branch"`
	Phase       string            `json:"phase"`
	TechStack   string            `json:"tech_stack,omitempty"`
	Constraints string            `json:"constraints,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Tasks       []TaskState       `json:"tasks,omitempty"`
	Answers     map[string]string `json:"answers,omitempty"`
}

// TaskState represents a task's persisted state
type TaskState struct {
	ID         string `json:"id"`
	Spec       string `json:"spec"`
	Status     string `json:"status"`
	Branch     string `json:"branch"`
	AgentName  string `json:"agent_name"`
	IsParallel bool   `json:"is_parallel"`
	Attempt    int    `json:"attempt"`
	FeatureID  string `json:"feature_id"`
}

// Store represents the persistence store data
type Store struct {
	Features  map[string]*FeatureState `json:"features"`
	UpdatedAt time.Time                `json:"updated_at"`
}

// FileStorage provides JSON file-based persistence
type FileStorage struct {
	path  string
	mu    sync.RWMutex
	store *Store
}

// New creates a new file storage instance
func New(path string) (*FileStorage, error) {
	fs := &FileStorage{
		path: path,
		store: &Store{
			Features: make(map[string]*FeatureState),
		},
	}

	// Try to load existing data
	if err := fs.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("loading storage: %w", err)
	}

	return fs, nil
}

func (fs *FileStorage) load() error {
	data, err := os.ReadFile(fs.path)
	if err != nil {
		return err
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err := json.Unmarshal(data, &fs.store); err != nil {
		return fmt.Errorf("parsing storage file: %w", err)
	}

	return nil
}

func (fs *FileStorage) save() error {
	fs.store.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(fs.store, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling storage: %w", err)
	}

	if err := os.WriteFile(fs.path, data, 0644); err != nil {
		return fmt.Errorf("writing storage file: %w", err)
	}

	return nil
}

// SaveFeature persists a feature state
func (fs *FileStorage) SaveFeature(state *FeatureState) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	state.UpdatedAt = time.Now()
	fs.store.Features[state.ID] = state

	return fs.save()
}

// LoadFeature retrieves a feature state
func (fs *FileStorage) LoadFeature(id string) (*FeatureState, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	state, ok := fs.store.Features[id]
	if !ok {
		return nil, fmt.Errorf("feature %s not found", id)
	}

	return state, nil
}

// LoadAllFeatures retrieves all feature states
func (fs *FileStorage) LoadAllFeatures() ([]*FeatureState, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	features := make([]*FeatureState, 0, len(fs.store.Features))
	for _, f := range fs.store.Features {
		features = append(features, f)
	}

	return features, nil
}

// DeleteFeature removes a feature from storage
func (fs *FileStorage) DeleteFeature(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	delete(fs.store.Features, id)

	return fs.save()
}
