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

type EntityOptions struct {
	ConfigPath string
	SourceDir  string // where to read prior docs (requirements/srs/stories/architecture). If empty, use cfg.IO.OutputDir
	OutputDir  string // where to write entities.md
	Role       string
	DryRun     bool
}

//go:embed entity_prompt.md
var entityPromptTemplate string

func Entity(opts EntityOptions) error {
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

	systemMessages, err := buildEntitySystemMessage(opts.SourceDir, cfg.IO.OutputDir)
	if err != nil {
		return err
	}

	if opts.DryRun {
		// In dry run mode, write scaffold files without making API calls
		return writeEntityScaffold(cfg.IO.OutputDir)
	}

	_, err = agents.SA.RunInputs(context.Background(), systemMessages)
	if err != nil {
		fmt.Printf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		// Write scaffold as fallback
		if scaffoldErr := writeEntityScaffold(cfg.IO.OutputDir); scaffoldErr != nil {
			return fmt.Errorf("API call failed and scaffold write failed: %v (original: %v)", scaffoldErr, err)
		}
	}

	return err
}

func buildEntitySystemMessage(sourceDir, outputDir string) ([]agents.TResponseInputItem, error) {
	data := struct {
		RequirementsPath string
		SrsPath          string
		StoriesPath      string
		ArchitecturePath string
		EntitiesPath     string
	}{
		RequirementsPath: filepath.Join(sourceDir, "requirements.md"),
		SrsPath:          filepath.Join(sourceDir, "srs.md"),
		StoriesPath:      filepath.Join(sourceDir, "stories.md"),
		ArchitecturePath: filepath.Join(sourceDir, "architecture.md"),
		EntitiesPath:     filepath.Join(outputDir, "entities.md"),
	}

	tmpl, err := template.New("entity").Parse(entityPromptTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse entity template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render entity template: %w", err)
	}

	return agents.InputList(
		agents.SystemMessage(buf.String()),
	), nil
}

func ensureEntities(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		// Provide fallback content when empty
		return `# Entities and Data Models

## Domain Entities

This section describes the core domain entities and their relationships.

### Entity Overview

List of main entities in the system:

- **Entity1**: Brief description
- **Entity2**: Brief description

## Data Models

### Entity Schemas

` + "```" + `
Entity1:
  id: UUID (Primary Key)
  name: String
  createdAt: DateTime
  updatedAt: DateTime

Entity2:
  id: UUID (Primary Key)
  entity1Id: UUID (Foreign Key -> Entity1.id)
  description: String
  status: Enum [ACTIVE, INACTIVE]
` + "```" + `

## Relationships

### Entity Relationship Diagram

` + "```plantuml" + `
@startuml
!theme plain
title Entity Relationship Diagram

entity "Entity1" as e1 {
  * id : UUID <<PK>>
  --
  name : String
  createdAt : DateTime
  updatedAt : DateTime
}

entity "Entity2" as e2 {
  * id : UUID <<PK>>
  --
  * entity1Id : UUID <<FK>>
  description : String
  status : Enum
}

e1 ||--o{ e2
@enduml
` + "```" + `

## Database Design

### Constraints and Indexes

- Primary keys on all entities
- Foreign key constraints for relationships
- Indexes on frequently queried fields
- Unique constraints where applicable

### Data Lifecycle

- Entity creation and validation rules
- Update patterns and constraints
- Deletion policies and cascading rules
`
	}

	lower := strings.ToLower(s)
	var additions []string

	// Check if Domain Entities section exists
	if !strings.Contains(lower, "domain entities") {
		additions = append(additions, "\n## Domain Entities\n\nThis section describes the core domain entities and their relationships.")
	}

	// Check if Data Models section exists
	if !strings.Contains(lower, "data models") {
		additions = append(additions, "\n## Data Models\n\nDetailed schemas and data structures.")
	}

	// Check if Relationships section exists
	if !strings.Contains(lower, "relationships") {
		additions = append(additions, "\n## Relationships\n\nEntity relationships and dependencies.")
	}

	// Check if Database Design section exists
	if !strings.Contains(lower, "database design") {
		additions = append(additions, "\n## Database Design\n\nDatabase-specific design considerations.")
	}

	if len(additions) > 0 {
		return s + strings.Join(additions, "")
	}

	return s
}

func writeEntityScaffold(outputDir string) error {
	// Write entities.md with scaffold content
	entityContent := ensureEntities("")
	entityPath := filepath.Join(outputDir, "entities.md")
	if err := os.WriteFile(entityPath, []byte(entityContent), 0644); err != nil {
		return fmt.Errorf("write entities.md: %w", err)
	}

	return nil
}
