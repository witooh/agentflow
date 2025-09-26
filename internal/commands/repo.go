package commands

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"agentflow/internal/agents"
	"agentflow/internal/config"
)

type RepoOptions struct {
	ConfigPath string
	SourceDir  string // where to read prior docs (requirements/srs/stories/architecture/entities). If empty, use cfg.IO.OutputDir
	OutputDir  string // where to write repository.md
	Role       string
	DryRun     bool
}

//go:embed repo_prompt.md
var repoPromptTemplate string

func Repo(opts RepoOptions) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.ApplyEnv()
	if opts.OutputDir != "" {
		cfg.IO.OutputDir = opts.OutputDir
	}
	if opts.SourceDir == "" {
		opts.SourceDir = cfg.IO.OutputDir
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := config.EnsureDirs(opts.ConfigPath, cfg); err != nil {
		return err
	}

	systemMessages, err := buildRepoSystemMessage(opts.SourceDir, cfg.IO.OutputDir)
	if err != nil {
		return err
	}

	if opts.DryRun {
		// In dry run mode, write scaffold files without making API calls
		return writeRepoScaffold(cfg.IO.OutputDir)
	}

	_, err = agents.SA.RunInputs(context.Background(), systemMessages)
	if err != nil {
		fmt.Printf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		// Write scaffold as fallback
		if scaffoldErr := writeRepoScaffold(cfg.IO.OutputDir); scaffoldErr != nil {
			return fmt.Errorf("API call failed and scaffold write failed: %v (original: %v)", scaffoldErr, err)
		}
	}

	return err
}

func buildRepoSystemMessage(sourceDir, outputDir string) ([]agents.TResponseInputItem, error) {
	data := struct {
		RequirementsPath string
		SrsPath          string
		StoriesPath      string
		ArchitecturePath string
		EntitiesPath     string
		RepositoryPath   string
	}{
		RequirementsPath: filepath.Join(sourceDir, "requirements.md"),
		SrsPath:          filepath.Join(sourceDir, "srs.md"),
		StoriesPath:      filepath.Join(sourceDir, "stories.md"),
		ArchitecturePath: filepath.Join(sourceDir, "architecture.md"),
		EntitiesPath:     filepath.Join(sourceDir, "entities.md"),
		RepositoryPath:   filepath.Join(outputDir, "repository.md"),
	}

	tmpl, err := template.New("repo").Parse(repoPromptTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse repo template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render repo template: %w", err)
	}

	return agents.InputList(
		agents.SystemMessage(buf.String()),
	), nil
}

func ensureRepository(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		// Provide fallback content when empty
		return `# Repository Interfaces

## Repository Pattern Overview

The Repository pattern encapsulates the logic needed to access data sources. It centralizes common data access functionality, providing better maintainability and decoupling the infrastructure or technology used to access databases from the domain model layer.

### Benefits

- **Separation of Concerns**: Isolates data access logic from business logic
- **Testability**: Easy to mock for unit testing
- **Flexibility**: Can switch between different data sources
- **Maintainability**: Centralized data access logic

## Base Repository Interface

` + "```go" + `
package repository

import (
	"context"
	"errors"
)

// Common errors
var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("record already exists")
)

// ListOptions defines common parameters for list operations
type ListOptions struct {
	Limit  int
	Offset int
	SortBy string
	Order  string // "asc" or "desc"
}

// BaseRepository defines common operations for all repositories
type BaseRepository[T any, ID any] interface {
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id ID) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id ID) error
	List(ctx context.Context, opts ListOptions) ([]*T, error)
	Count(ctx context.Context) (int64, error)
}
` + "```" + `

## Entity-Specific Repositories

### User Repository

` + "```go" + `
package repository

import "context"

type User struct {
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserFilter struct {
	Email  *string
	Name   *string
	Active *bool
}

type UserRepository interface {
	BaseRepository[User, string]
	
	// Business-specific methods
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByName(ctx context.Context, name string) (*User, error)
	Search(ctx context.Context, filter UserFilter, opts ListOptions) ([]*User, error)
	UpdateEmail(ctx context.Context, id string, email string) error
	SetActive(ctx context.Context, id string, active bool) error
	
	// Batch operations
	CreateBatch(ctx context.Context, users []*User) error
	DeleteBatch(ctx context.Context, ids []string) error
}
` + "```" + `

## Implementation Guidelines

### Error Handling

` + "```go" + `
// Standard error handling pattern
func (r *userRepository) GetByID(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	
	user, err := r.db.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	
	return user, nil
}
` + "```" + `

### Transaction Support

` + "```go" + `
type TxRepository interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type UserRepositoryTx interface {
	UserRepository
	TxRepository
}
` + "```" + `

### Context Usage

- Always use ` + "`context.Context`" + ` as the first parameter
- Respect context cancellation and timeout
- Pass context to underlying database operations

### Testing Strategies

` + "```go" + `
// Mock repository for testing
type MockUserRepository struct {
	users map[string]*User
	mu    sync.RWMutex
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.users[user.ID]; exists {
		return ErrConflict
	}
	
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	user, exists := m.users[id]
	if !exists {
		return nil, ErrNotFound
	}
	
	return user, nil
}
` + "```" + `

## Performance Considerations

### Caching Strategy

` + "```go" + `
type CachedUserRepository struct {
	repo  UserRepository
	cache Cache
	ttl   time.Duration
}

func (c *CachedUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	// Try cache first
	if user, found := c.cache.Get(id); found {
		return user.(*User), nil
	}
	
	// Fallback to repository
	user, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	c.cache.Set(id, user, c.ttl)
	return user, nil
}
` + "```" + `

### Batch Operations

Use batch operations for better performance when dealing with multiple records:

` + "```go" + `
func (r *userRepository) CreateBatch(ctx context.Context, users []*User) error {
	const batchSize = 100
	
	for i := 0; i < len(users); i += batchSize {
		end := i + batchSize
		if end > len(users) {
			end = len(users)
		}
		
		batch := users[i:end]
		if err := r.db.InsertBatch(ctx, batch); err != nil {
			return fmt.Errorf("batch insert failed at index %d: %w", i, err)
		}
	}
	
	return nil
}
` + "```" + `

## Database Connection Management

### Connection Pool Configuration

` + "```go" + `
type DatabaseConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func NewRepository(db *sql.DB, config DatabaseConfig) Repository {
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	
	return &repository{db: db}
}
` + "```" + `

## Migration and Schema Management

Repository interfaces should work with proper database schema management:

` + "```go" + `
type SchemaManager interface {
	Migrate(ctx context.Context) error
	Rollback(ctx context.Context, steps int) error
	Version(ctx context.Context) (int, error)
}
` + "```" + `

This repository design provides a solid foundation for data access layer implementation in Go applications.`
	}

	lower := strings.ToLower(s)
	var additions []string

	// Check if Repository Pattern Overview section exists
	if !strings.Contains(lower, "repository pattern overview") {
		additions = append(additions, "\n## Repository Pattern Overview\n\nThe Repository pattern encapsulates data access logic.")
	}

	// Check if Repository Interfaces section exists
	if !strings.Contains(lower, "repository interfaces") {
		additions = append(additions, "\n## Repository Interfaces\n\nInterface definitions for data access operations.")
	}

	// Check if Implementation Guidelines section exists
	if !strings.Contains(lower, "implementation guidelines") {
		additions = append(additions, "\n## Implementation Guidelines\n\nBest practices for repository implementation.")
	}

	// Check if Testing Strategies section exists
	if !strings.Contains(lower, "testing strategies") {
		additions = append(additions, "\n## Testing Strategies\n\nApproaches for testing repository implementations.")
	}

	if len(additions) > 0 {
		return s + strings.Join(additions, "")
	}

	return s
}

func writeRepoScaffold(outputDir string) error {
	// Write repository.md with scaffold content
	repoContent := ensureRepository("")
	repoPath := filepath.Join(outputDir, "repository.md")
	if err := os.WriteFile(repoPath, []byte(repoContent), 0644); err != nil {
		return fmt.Errorf("write repository.md: %w", err)
	}

	return nil
}
