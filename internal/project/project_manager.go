package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"network-log-formatter/internal/model"
)

// ProjectManager handles CRUD operations for project records.
// Each project is persisted as an individual JSON file named {id}.json.
type ProjectManager struct {
	storagePath string
}

// NewProjectManager creates a new ProjectManager that stores projects in the given directory.
// It creates the storage directory if it does not exist.
func NewProjectManager(storagePath string) (*ProjectManager, error) {
	if err := os.MkdirAll(storagePath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &ProjectManager{storagePath: storagePath}, nil
}

// Create saves a new project as a JSON file named {id}.json.
// If a project with the same Name already exists, a numeric suffix is appended
// (e.g. "name_2", "name_3") to ensure uniqueness.
func (pm *ProjectManager) Create(project model.Project) error {
	if project.Name != "" {
		existing, _ := pm.List()
		project.Name = pm.uniqueName(project.Name, project.ID, existing)
	}

	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project: %w", err)
	}
	path := pm.filePath(project.ID)
	return os.WriteFile(path, data, 0o644)
}

// List reads all project files and returns them sorted by CreatedAt descending.
func (pm *ProjectManager) List() ([]model.Project, error) {
	entries, err := os.ReadDir(pm.storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var projects []model.Project
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(pm.storagePath, entry.Name()))
		if err != nil {
			continue // skip unreadable files
		}
		var p model.Project
		if err := json.Unmarshal(data, &p); err != nil {
			continue // skip corrupted files
		}
		projects = append(projects, p)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].CreatedAt.After(projects[j].CreatedAt)
	})

	return projects, nil
}

// Get reads a single project by ID.
func (pm *ProjectManager) Get(id string) (*model.Project, error) {
	data, err := os.ReadFile(pm.filePath(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("project not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read project: %w", err)
	}

	var p model.Project
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project: %w", err)
	}
	return &p, nil
}

// Update applies partial updates to an existing project using pointer fields.
func (pm *ProjectManager) Update(id string, updates model.ProjectUpdate) error {
	p, err := pm.Get(id)
	if err != nil {
		return err
	}

	if updates.Name != nil {
		existing, _ := pm.List()
		p.Name = pm.uniqueName(*updates.Name, id, existing)
	}
	if updates.Code != nil {
		p.Code = *updates.Code
	}
	if updates.Status != nil {
		p.Status = *updates.Status
	}
	p.UpdatedAt = time.Now()

	// Write directly to avoid re-checking uniqueness against self
	data, jsonErr := json.MarshalIndent(*p, "", "  ")
	if jsonErr != nil {
		return fmt.Errorf("failed to marshal project: %w", jsonErr)
	}
	return os.WriteFile(pm.filePath(id), data, 0o644)
}

// Delete removes a project file by ID.
func (pm *ProjectManager) Delete(id string) error {
	path := pm.filePath(id)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("project not found: %s", id)
	}
	return os.Remove(path)
}

// uniqueName returns a name that does not conflict with any existing project.
// If baseName is already taken, it appends _2, _3, etc. If baseName already
// ends with a numeric suffix (e.g. "foo_3"), the counter starts from that number.
func (pm *ProjectManager) uniqueName(baseName string, selfID string, projects []model.Project) string {
	names := make(map[string]struct{})
	for _, p := range projects {
		if p.ID != selfID {
			names[p.Name] = struct{}{}
		}
	}

	if _, taken := names[baseName]; !taken {
		return baseName
	}

	// Strip existing numeric suffix to find the root name.
	re := regexp.MustCompile(`^(.+)_(\d+)$`)
	root := baseName
	start := 2
	if m := re.FindStringSubmatch(baseName); m != nil {
		root = m[1]
		start, _ = strconv.Atoi(m[2])
		if start < 2 {
			start = 2
		}
	}

	for i := start; ; i++ {
		candidate := fmt.Sprintf("%s_%d", root, i)
		if _, taken := names[candidate]; !taken {
			return candidate
		}
	}
}

// filePath returns the full file path for a project by ID.
// It validates the ID to prevent path traversal attacks.
func (pm *ProjectManager) filePath(id string) string {
	// Sanitize: use filepath.Base to strip directory components,
	// then reject any result that is empty, ".", or ".."
	clean := filepath.Base(id)
	if clean == "." || clean == ".." || clean == "" {
		clean = "_invalid_"
	}
	return filepath.Join(pm.storagePath, clean+".json")
}
