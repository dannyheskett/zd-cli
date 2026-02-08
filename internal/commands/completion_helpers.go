package commands

import (
	"os"
	"strings"
)

func addBlockToFile(filepath, block, marker string) error {
	// Read existing content
	content, err := os.ReadFile(filepath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Check if marker already exists (avoid duplicates)
	if strings.Contains(string(content), marker) {
		return nil // Already installed
	}

	// Append to file
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Add block
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		f.WriteString("\n")
	}
	f.WriteString(block)

	return nil
}
