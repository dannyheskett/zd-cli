package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"zd-cli/internal/cache"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewCacheCommand creates the cache management command
func NewCacheCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage API response cache",
		Long:  "View cache statistics and clear cached API responses.",
	}

	cmd.AddCommand(newCacheInfoCommand())
	cmd.AddCommand(newCacheClearCommand())

	return cmd
}

func newCacheInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show cache information",
		RunE:  runCacheInfo,
	}
}

func newCacheClearCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear all cached data",
		RunE:  runCacheClear,
	}
}

func runCacheInfo(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(home, ".zd", "cache")

	// Check if cache directory exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		color.Yellow("Cache directory does not exist yet.\n")
		color.White("Cache will be created when you run commands that access the API.\n")
		return nil
	}

	// Read cache entries
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	if len(entries) == 0 {
		color.Yellow("Cache is empty.\n")
		return nil
	}

	var totalSize int64
	validEntries := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		totalSize += info.Size()
		validEntries++
	}

	color.Cyan("Cache Information\n")
	color.White("─────────────────\n")
	color.White("Location:     %s\n", cacheDir)
	color.White("Entries:      %d\n", validEntries)
	color.White("Total size:   %.2f KB\n", float64(totalSize)/1024)
	color.White("Default TTL:  10 minutes\n")

	return nil
}

func runCacheClear(cmd *cobra.Command, args []string) error {
	c, err := cache.New(15 * time.Minute)
	if err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}

	if err := c.Clear(); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	color.Green("✓ Cache cleared successfully!\n")

	return nil
}
