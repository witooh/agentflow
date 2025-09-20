package prompt

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nlpodyssey/openai-agents-go/agents"
)

// getPromptFromFile reads the content of the given file path and returns it as a string.
// If the file does not exist or cannot be read, it returns an error.
func getPromptFromFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GetPromptFromFiles reads the content of the given files and returns them as a slice of TResponseInputItem.
// If the file does not exist or cannot be read, it returns an error.
func GetPromptFromFiles(paths []string) ([]agents.TResponseInputItem, error) {
	items := []agents.TResponseInputItem{}
	for _, path := range paths {
		prompt, err := getPromptFromFile(path)
		if err != nil {
			return nil, err
		}
		items = append(items, agents.UserMessage(path+"\n\n"+prompt))
	}
	return items, nil
}

type BuildOptions struct {
	RoleTemplate string
	InputsDir    string
	ExtraContext string
}

func ListInputFiles(inputsDir string) ([]string, error) {
	files, err := readMarkdownFiles(inputsDir)
	if err != nil {
		return nil, err
	}
	return filesToNames(files), nil
}

func GetInputFiles(dir string, filenames []string) ([]string, error) {
	files := []fileData{}
	for _, filename := range filenames {
		files = append(files, fileData{Path: filepath.Join(dir, filename), Content: ""})
	}
	return filesToNames(files), nil
}

type fileData struct {
	Path    string
	Content string
}

func readMarkdownFiles(dir string) ([]fileData, error) {
	var out []fileData
	// If dir does not exist, return no files, no error
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return out, nil
	}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		low := strings.ToLower(d.Name())
		if strings.HasSuffix(low, ".md") || strings.HasSuffix(low, ".txt") {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			out = append(out, fileData{Path: path, Content: string(b)})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// stable order
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func filesToNames(files []fileData) []string {
	n := make([]string, 0, len(files))
	for _, f := range files {
		n = append(n, f.Path)
	}
	return n
}

// BuildTimelineSummary reads input markdown files and composes a concise timeline
// by sorting files whose base name starts with YYYY-MM-DD by date ascending.
// It returns the timeline summary string and the sorted file names that contributed.
func BuildTimelineSummary(dir string) (string, []string, error) {
	files, err := readMarkdownFiles(dir)
	if err != nil {
		return "", nil, err
	}
	type item struct {
		date      time.Time
		dateStr   string
		name      string
		firstLine string
	}
	var items []item
	for _, f := range files {
		base := filepath.Base(f.Path)
		if len(base) >= 10 && base[4] == '-' && base[7] == '-' {
			if dt, err := time.Parse("2006-01-02", base[:10]); err == nil {
				fl := firstNonEmptyLine(f.Content)
				if len(fl) > 200 {
					fl = fl[:200] + "…"
				}
				items = append(items, item{date: dt, dateStr: base[:10], name: base, firstLine: fl})
			}
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].date.Before(items[j].date) })
	if len(items) == 0 {
		return "", []string{}, nil
	}
	b := &strings.Builder{}
	b.WriteString("A chronological summary of idea evolution by date:\n")
	for _, it := range items {
		fmt.Fprintf(b, "- %s — %s\n", it.dateStr, strings.TrimSpace(it.firstLine))
	}
	sortedNames := make([]string, 0, len(items))
	for _, it := range items {
		sortedNames = append(sortedNames, it.name)
	}
	return b.String(), sortedNames, nil
}

func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}
