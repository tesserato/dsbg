package main

import (
	"flag"
	"fmt"
	// "io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Constants for JS and CSS (can be moved to separate files)
const JS = `
// ... (JavaScript code remains the same) 
`

const CSS = `
// ... (CSS code remains the same)
`

// Article struct represents a single article
type Article struct {
	Title        string    `yaml:"title"`
	Description  string    `yaml:"description"`
	Created      time.Time `yaml:"created"`
	LastUpdated  time.Time `yaml:"last_updated"`
	Tags         []string  `yaml:"tags"`
	Content      string
	OriginalPath string
}

// shouldIgnore checks if a given path should be ignored based on the ignore patterns
func shouldIgnore(path string, ignoreList []string) bool {
	if path == "." {
		return true
	}
	for _, pattern := range ignoreList {
		if pattern == "" {
			continue
		}
		matched, err := regexp.MatchString(pattern, path)
		if err != nil {
			fmt.Printf("Error matching pattern '%s': %v\n", pattern, err)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

// printHelp prints the help message
func printHelp() {
	fmt.Println("Usage: go run codemerge.go [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
}

// savePathsToFile saves a list of paths to a file, one per line
func savePathsToFile(filename string, paths []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, path := range paths {
		_, err := fmt.Fprintln(file, path)
		if err != nil {
			return err
		}
	}

	return nil
}

func processFile(path string, ignoreList []string) (*Article, error) {
	// Check if the file should be ignored
	relPath, _ := filepath.Rel(".", path)
	if shouldIgnore(relPath, ignoreList) {
		return nil, nil // Skip ignored files
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", path, err)
	}

	// Extract frontmatter and content
	frontmatter, contentStr, err := extractFrontmatter(string(content))
	if err != nil {
		return nil, fmt.Errorf("error extracting frontmatter from %s: %w", path, err)
	}

	// Parse frontmatter
	var article Article
	err = yaml.Unmarshal([]byte(frontmatter), &article)
	if err != nil {
		return nil, fmt.Errorf("error parsing frontmatter from %s: %w", path, err)
	}
	article.Content = contentStr
	article.OriginalPath = path
	return &article, nil
}

func main() {
	// Define command line flags
	dirPath := flag.String("dir", ".", "Directory to scan")
	// outputFileName := flag.String("output", "codebase.md", "Output file name")
	ignorePatterns := flag.String("ignore", `\.git.*`, "Comma-separated list of regular expression patterns that match the paths to be ignored")
	includedPathsFile := flag.String("included-paths-file", "", "File to save included paths (optional). If provided, the included paths will be saved to the file and not printed to the console.")
	excludedPathsFile := flag.String("excluded-paths-file", "", "File to save excluded paths (optional). If provided, the excluded paths will be saved to the file and not printed to the console.")
	showHelp := flag.Bool("help", false, "Show help message")

	flag.Parse()

	// Check if help flag is set or no arguments are provided
	if *showHelp || len(os.Args) == 1 {
		printHelp()
		return
	}

	// Split ignore patterns string into a slice
	ignoreList := strings.Split(*ignorePatterns, ",")
	fmt.Println("Patterns to ignore:")
	for i, pattern := range ignoreList {
		ignoreList[i] = strings.TrimSpace(pattern)
		fmt.Println(ignoreList[i])
	}

}

