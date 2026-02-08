package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// NewInstallCommand creates the install command
func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install zd to your system PATH",
		Long:  "Install the zd binary to /usr/local/bin so you can run 'zd' from anywhere without './zd'",
		RunE:  runInstall,
	}

	return cmd
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Get the path of the current executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve any symlinks
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	targetPath := "/usr/local/bin/zd"

	// Check if already installed
	if existingPath, err := os.Readlink(targetPath); err == nil {
		if existingPath == exePath {
			color.Green("✓ zd is already installed at %s\n", targetPath)
			return nil
		}
	}

	// Check if file exists at target
	if _, err := os.Stat(targetPath); err == nil {
		color.Yellow("A file already exists at %s\n", targetPath)
		prompt := promptui.Prompt{
			Label:     "Overwrite it",
			IsConfirm: true,
		}
		result, err := prompt.Run()
		if err != nil || result != "y" {
			color.Yellow("Installation cancelled.\n")
			return nil
		}
	}

	color.Cyan("Installing zd to %s...\n", targetPath)
	color.White("This will copy the binary and may require sudo permissions.\n\n")

	// Copy the binary to /usr/local/bin
	sourceFile, err := os.Open(exePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(targetPath)
	if err != nil {
		// If permission denied, provide helpful message
		if os.IsPermission(err) {
			color.Red("✗ Permission denied. This command needs elevated privileges.\n")
			color.White("\nPlease run with sudo:\n")
			color.Cyan("  sudo %s install\n", exePath)
			return nil
		}
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the contents
	sourceFile.Seek(0, 0)
	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		os.Remove(targetPath) // Clean up on error
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Make it executable
	if err := os.Chmod(targetPath, 0755); err != nil {
		os.Remove(targetPath) // Clean up on error
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	color.Green("\n✓ zd installed successfully to %s\n", targetPath)
	color.White("\nYou can now run 'zd' from anywhere!\n")
	color.White("Try: zd --version\n")

	return nil
}
