package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nlpodyssey/openai-agents-go/agents"
)

// FileCreatorTool is a tool that creates a file with specified content.
var FileCreatorTool = agents.NewFunctionTool(
	"file_creator",
	"Create a file with specified content.",
	CreateFile,
)

// CreateFileInput defines the input for the file creation tool.
type CreateFileArgs struct {
	Path    string
	Content string
}

// CreateFile creates a file at the given path with the provided content.
func CreateFile(_ context.Context, args CreateFileArgs) (string, error) {
	dir := filepath.Dir(args.Path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
	}

	if err := os.WriteFile(args.Path, []byte(args.Content), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("created file: %s", args.Path), nil
}
