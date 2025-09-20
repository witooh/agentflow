package tools

import (
	"context"
	"os"

	"github.com/nlpodyssey/openai-agents-go/agents"
)

// FileReaderTool is a tool that reads the contents of a file at a given path.
var FileReaderTool = agents.NewFunctionTool(
	"file_reader",
	"Read the contents of a file at the specified path.",
	ReadFile,
)

// ReadFileArgs defines the input for the file reader tool.
type ReadFileArgs struct {
	Path string
}

// ReadFile reads a file at the given path and returns its contents as a string.
func ReadFile(_ context.Context, args ReadFileArgs) (string, error) {
	data, err := os.ReadFile(args.Path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
