package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	// Define command line flags
	dirPath := flag.String("dir", ".", "Directory to scan")
	outputFileName := flag.String("output", "codebase.md", "Output file name")
	ignorePatterns := flag.String("ignore", "", "Comma-separated list of path patterns to ignore")

	flag.Parse()

	// Split ignore patterns string into a slice
	ignoreList := strings.Split(*ignorePatterns, ",")
	fmt.Println("Patterns to ignore:")
	for i, pattern := range ignoreList {
		ignoreList[i] = strings.TrimSpace(pattern)
		fmt.Println(ignoreList[i])
	}

	// Create the output file
	outputFile, err := os.Create(*outputFileName)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	// Write the codebase tree to the output file
	fmt.Fprintln(outputFile, "# Codebase Structure\n")
	err = printTree(*dirPath, "", ignoreList, outputFile)
	if err != nil {
		fmt.Println("Error printing codebase tree:", err)
		return
	}

	// Write the code content to the output file
	fmt.Fprintln(outputFile, "# Code Content\n")
	err = writeCodeContent(*dirPath, ignoreList, outputFile)
	if err != nil {
		fmt.Println("Error writing code content:", err)
		return
	}

	fmt.Println("Codebase documentation generated successfully!")
}

// printTree recursively walks the directory tree and prints the structure to the output file
func printTree(dirPath string, indent string, ignoreList []string, outputFile *os.File) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		filePath := filepath.Join(dirPath, file.Name())
		relPath, _ := filepath.Rel(".", filePath)

		// Check if the file/directory should be ignored
		if shouldIgnore(relPath, ignoreList) {
			continue
		}

		if file.IsDir() {
			fmt.Fprintf(outputFile, "%s- **%s/**\n", indent, file.Name())
			printTree(filePath, indent+"  ", ignoreList, outputFile)
		} else {
			fmt.Fprintf(outputFile, "%s- %s\n", indent, file.Name())
		}
	}

	return nil
}

// writeCodeContent reads the content of each file and writes it to the output file within a code block
func writeCodeContent(dirPath string, ignoreList []string, outputFile *os.File) error {
	var Red = "\033[31m"
	var Green = "\033[32m"
	var Reset = "\033[0m"
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if the file should be ignored
		relPath, _ := filepath.Rel(".", path)
		if shouldIgnore(relPath, ignoreList) {
			fmt.Println(Red + "- " + path + Reset)
			return nil
		}
		fmt.Println(Green + "+ " + path + Reset)

		if !d.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			fmt.Fprintf(outputFile, "## %s\n", path)
			fmt.Fprintf(outputFile, "```\n%s\n```\n\n", content)
		}

		return nil
	})

	return err
}

// shouldIgnore checks if a given path should be ignored based on the ignore patterns
func shouldIgnore(path string, ignoreList []string) bool {
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
