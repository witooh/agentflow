package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"agentflow/internal/commands"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "help", "-h", "--help":
		usage()
		return
	case "version", "-v", "--version":
		fmt.Println("agentflow v0.1.0")
		return
	case "init":
		initCmd(os.Args[2:])
	case "intake":
		intakeCmd(os.Args[2:])
	case "plan":
		planCmd(os.Args[2:])
	case "design":
		designCmd(os.Args[2:])
	case "uml":
		umlCmd(os.Args[2:])
	case "qa":
		qaCmd(os.Args[2:])
	case "devplan":
		devplanCmd(os.Args[2:])
	case "entity":
		entityCmd(os.Args[2:])
	case "repo":
		repoCmd(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	prog := filepath.Base(os.Args[0])
	fmt.Printf(`AgentFlow CLI

Usage:
  %s <command> [flags]

Commands:
  init        Initialize .agentflow/config.json
  intake      Aggregate input and generate requirements.md
  plan        Generate srs.md, stories.md, and acceptance_criteria.md from requirements.md
  design      Generate architecture.md and uml.md from prior docs
  uml         Generate uml.md from requirements/srs/stories using uml template
  qa          Generate test-plan.md from prior docs
  devplan     Generate task list and per-task context
  entity      Generate entities.md with data models and relationships
  repo        Generate repository.md with Golang repository interfaces
  help        Show this help
  version     Show version

Use "%s <command> -h" for command-specific help.
`, prog, prog)
}

func initCmd(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	projectName := fs.String("project-name", "MyProject", "Project name to store in config")
	model := fs.String("model", "gpt-5", "Default LLM model")
	configPath := fs.String("config", ".agentflow/config.json", "Path to config file to create")
	_ = fs.Parse(args)

	if err := commands.Init(*configPath, *projectName, *model); err != nil {
		log.Fatalf("init failed: %v", err)
	}
	fmt.Printf("Initialized %s\n", *configPath)
}

func intakeCmd(args []string) {
	fs := flag.NewFlagSet("intake", flag.ExitOnError)
	configPath := fs.String("config", ".agentflow/config.json", "Path to config file")
	inputsDir := fs.String("input", ".agentflow/input", "Input directory with .md files")
	outputDir := fs.String("output", ".agentflow/output", "Output directory")
	role := fs.String("role", "po_pm", "Role to use for prompt building (po_pm)")
	dryRun := fs.Bool("dry-run", false, "Do not call OpenAI, just scaffold output")
	_ = fs.Parse(args)

	if err := commands.Intake(commands.IntakeOptions{
		ConfigPath: *configPath,
		InputsDir:  *inputsDir,
		OutputDir:  *outputDir,
		Role:       *role,
		DryRun:     *dryRun,
	}); err != nil {
		if errors.Is(err, commands.ErrNoInputs) {
			fmt.Fprintln(os.Stderr, "warning: no input markdown files found; creating empty requirements.md")
		} else {
			log.Fatalf("intake failed: %v", err)
		}
	}
	fmt.Printf("Wrote %s\n", filepath.Join(*outputDir, "requirements.md"))
}

func planCmd(args []string) {
	fs := flag.NewFlagSet("plan", flag.ExitOnError)
	configPath := fs.String("config", ".agentflow/config.json", "Path to config file")
	reqPath := fs.String("requirements", ".agentflow/output/requirements.md", "Path to requirements.md")
	outputDir := fs.String("output", ".agentflow/output", "Output directory")
	role := fs.String("role", "sa", "Role to use for planning (sa)")
	dryRun := fs.Bool("dry-run", false, "Do not call OpenAI, just scaffold output")
	_ = fs.Parse(args)

	if err := commands.Plan(commands.PlanOptions{
		ConfigPath:   *configPath,
		Requirements: *reqPath,
		OutputDir:    *outputDir,
		Role:         *role,
		DryRun:       *dryRun,
	}); err != nil {
		if errors.Is(err, commands.ErrNoRequirements) {
			log.Fatalf("plan failed: requirements.md not found at %s", *reqPath)
		}
		log.Fatalf("plan failed: %v", err)
	}
	fmt.Printf("Wrote %s, %s, %s\n", filepath.Join(*outputDir, "srs.md"), filepath.Join(*outputDir, "stories.md"), filepath.Join(*outputDir, "acceptance_criteria.md"))
}

func qaCmd(args []string) {
	fs := flag.NewFlagSet("qa", flag.ExitOnError)
	configPath := fs.String("config", ".agentflow/config.json", "Path to config file")
	sourceDir := fs.String("source", ".agentflow/output", "Directory with prior docs (srs/stories/AC)")
	outputDir := fs.String("output", ".agentflow/output", "Output directory")
	role := fs.String("role", "qa", "Role to use for QA (qa)")
	dryRun := fs.Bool("dry-run", false, "Do not call OpenAI, just scaffold output")
	_ = fs.Parse(args)

	if err := commands.QA(commands.QAOptions{
		ConfigPath: *configPath,
		SourceDir:  *sourceDir,
		OutputDir:  *outputDir,
		Role:       *role,
		DryRun:     *dryRun,
	}); err != nil {
		log.Fatalf("qa failed: %v", err)
	}
	fmt.Printf("Wrote %s\n", filepath.Join(*outputDir, "test-plan.md"))
}

func designCmd(args []string) {
	fs := flag.NewFlagSet("design", flag.ExitOnError)
	configPath := fs.String("config", ".agentflow/config.json", "Path to config file")
	sourceDir := fs.String("source", ".agentflow/output", "Directory with prior docs (requirements/srs/stories/AC)")
	outputDir := fs.String("output", ".agentflow/output", "Output directory")
	role := fs.String("role", "sa", "Role to use for design (sa)")
	dryRun := fs.Bool("dry-run", false, "Do not call OpenAI, just scaffold output")
	_ = fs.Parse(args)

	if err := commands.Design(commands.DesignOptions{
		ConfigPath: *configPath,
		SourceDir:  *sourceDir,
		OutputDir:  *outputDir,
		Role:       *role,
		DryRun:     *dryRun,
	}); err != nil {
		log.Fatalf("design failed: %v", err)
	}
	fmt.Printf("Wrote %s\n", filepath.Join(*outputDir, "architecture.md"))
}

func umlCmd(args []string) {
	fs := flag.NewFlagSet("uml", flag.ExitOnError)
	configPath := fs.String("config", ".agentflow/config.json", "Path to config file")
	sourceDir := fs.String("source", ".agentflow/output", "Directory with prior docs (requirements/srs/stories)")
	outputDir := fs.String("output", ".agentflow/output", "Output directory")
	role := fs.String("role", "sa", "Role to use for uml (sa)")
	dryRun := fs.Bool("dry-run", false, "Do not call OpenAI, just scaffold output")
	_ = fs.Parse(args)

	if err := commands.Uml(commands.UmlOptions{
		ConfigPath: *configPath,
		SourceDir:  *sourceDir,
		OutputDir:  *outputDir,
		Role:       *role,
		DryRun:     *dryRun,
	}); err != nil {
		log.Fatalf("uml failed: %v", err)
	}
	fmt.Printf("Wrote %s\n", filepath.Join(*outputDir, "uml.md"))
}

func devplanCmd(args []string) {
	fs := flag.NewFlagSet("devplan", flag.ExitOnError)
	configPath := fs.String("config", ".agentflow/config.json", "Path to config file")
	sourceDir := fs.String("source", ".agentflow/output", "Directory with prior generated docs (requirements/srs/stories/...)")
	outputDir := fs.String("output", ".agentflow/output", "Output directory for task_list.md and tasks/")
	role := fs.String("role", "dev", "Role to use for devplanning (dev)")
	dryRun := fs.Bool("dry-run", false, "Do not call OpenAI, just scaffold output")
	_ = fs.Parse(args)

	if err := commands.DevPlan(commands.DevPlanOptions{
		ConfigPath: *configPath,
		SourceDir:  *sourceDir,
		OutputDir:  *outputDir,
		Role:       *role,
		DryRun:     *dryRun,
	}); err != nil {
		log.Fatalf("devplan failed: %v", err)
	}
	fmt.Printf("Wrote %s and %s/*.md\n", filepath.Join(*outputDir, "task_list.md"), filepath.Join(*outputDir, "tasks"))
}

func entityCmd(args []string) {
	fs := flag.NewFlagSet("entity", flag.ExitOnError)
	configPath := fs.String("config", ".agentflow/config.json", "Path to config file")
	sourceDir := fs.String("source", ".agentflow/output", "Directory with prior docs (requirements/srs/stories/architecture)")
	outputDir := fs.String("output", ".agentflow/output", "Output directory")
	role := fs.String("role", "sa", "Role to use for entity design (sa)")
	dryRun := fs.Bool("dry-run", false, "Do not call OpenAI, just scaffold output")
	_ = fs.Parse(args)

	if err := commands.Entity(commands.EntityOptions{
		ConfigPath: *configPath,
		SourceDir:  *sourceDir,
		OutputDir:  *outputDir,
		Role:       *role,
		DryRun:     *dryRun,
	}); err != nil {
		log.Fatalf("entity failed: %v", err)
	}
	fmt.Printf("Wrote %s\n", filepath.Join(*outputDir, "entities.md"))
}

func repoCmd(args []string) {
	fs := flag.NewFlagSet("repo", flag.ExitOnError)
	configPath := fs.String("config", ".agentflow/config.json", "Path to config file")
	sourceDir := fs.String("source", ".agentflow/output", "Directory with prior docs (requirements/srs/stories/architecture/entities)")
	outputDir := fs.String("output", ".agentflow/output", "Output directory")
	role := fs.String("role", "sa", "Role to use for repository design (sa)")
	dryRun := fs.Bool("dry-run", false, "Do not call OpenAI, just scaffold output")
	_ = fs.Parse(args)

	if err := commands.Repo(commands.RepoOptions{
		ConfigPath: *configPath,
		SourceDir:  *sourceDir,
		OutputDir:  *outputDir,
		Role:       *role,
		DryRun:     *dryRun,
	}); err != nil {
		log.Fatalf("repo failed: %v", err)
	}
	fmt.Printf("Wrote %s\n", filepath.Join(*outputDir, "repository.md"))
}
