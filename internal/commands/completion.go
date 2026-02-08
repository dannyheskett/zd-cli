package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// NewCompletionCommand creates the completion command with auto-install
func NewCompletionCommand(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Install shell completion for zd",
		Long: `Automatically detects your shell and installs tab completion.

After installation, you'll be able to use TAB to auto-complete commands:
  - zd inst<TAB>      → zd instance
  - zd instance <TAB> → shows add, list, switch, remove, current
  - zd --<TAB>        → shows all available flags

Supported shells: bash, zsh, fish, powershell`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompletion(cmd, rootCmd)
		},
	}

	// Keep the subcommands for manual generation if needed
	cmd.AddCommand(&cobra.Command{
		Use:   "bash",
		Short: "Generate bash completion script",
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootCmd.GenBashCompletion(os.Stdout)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "zsh",
		Short: "Generate zsh completion script",
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootCmd.GenZshCompletion(os.Stdout)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "fish",
		Short: "Generate fish completion script",
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootCmd.GenFishCompletion(os.Stdout, true)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "powershell",
		Short: "Generate powershell completion script",
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootCmd.GenPowerShellCompletion(os.Stdout)
		},
	})

	return cmd
}

func runCompletion(cmd *cobra.Command, rootCmd *cobra.Command) error {
	// Detect current shell
	shell := detectShell()

	if shell == "" {
		color.Yellow("Could not detect your shell automatically.\n")
		color.White("Please run one of:\n")
		color.White("  zd completion bash\n")
		color.White("  zd completion zsh\n")
		color.White("  zd completion fish\n")
		color.White("  zd completion powershell\n")
		return nil
	}

	color.Cyan("Detected shell: %s\n\n", shell)
	color.White("This will install tab completion for the 'zd' command.\n")
	color.White("You'll need to restart your shell or run 'source ~/%s' after installation.\n\n", getShellRC(shell))

	// Confirm installation
	prompt := promptui.Prompt{
		Label:     "Install completion",
		IsConfirm: true,
	}
	result, err := prompt.Run()
	if err != nil || strings.ToLower(result) != "y" {
		color.Yellow("Installation cancelled.\n")
		return nil
	}

	// Install completion
	if err := installCompletion(shell, rootCmd); err != nil {
		color.Red("✗ Failed to install completion: %v\n", err)
		return err
	}

	color.Green("✓ Completion installed successfully!\n\n")

	// Show next steps
	shellRC := getShellRC(shell)
	color.White("To activate completion, run:\n")
	color.Cyan("  source ~/%s\n\n", shellRC)
	color.White("Or restart your shell.\n")

	return nil
}

func detectShell() string {
	// Check SHELL environment variable
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return ""
	}

	// Extract shell name from path
	shellName := filepath.Base(shellPath)

	// Normalize shell name
	switch {
	case strings.Contains(shellName, "bash"):
		return "bash"
	case strings.Contains(shellName, "zsh"):
		return "zsh"
	case strings.Contains(shellName, "fish"):
		return "fish"
	case strings.Contains(shellName, "pwsh") || strings.Contains(shellName, "powershell"):
		return "powershell"
	default:
		return ""
	}
}

func getShellRC(shell string) string {
	switch shell {
	case "bash":
		return ".bashrc"
	case "zsh":
		return ".zshrc"
	case "fish":
		return ".config/fish/config.fish"
	case "powershell":
		return "Documents/PowerShell/Microsoft.PowerShell_profile.ps1"
	default:
		return ".bashrc"
	}
}

func installCompletion(shell string, rootCmd *cobra.Command) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	switch shell {
	case "bash":
		return installBashCompletion(home, rootCmd)
	case "zsh":
		return installZshCompletion(home, rootCmd)
	case "fish":
		return installFishCompletion(home, rootCmd)
	case "powershell":
		return installPowerShellCompletion(home, rootCmd)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

func installBashCompletion(home string, rootCmd *cobra.Command) error {
	// Create bash completions directory
	completionDir := filepath.Join(home, ".bash_completion.d")
	if err := os.MkdirAll(completionDir, 0755); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	// Write completion script
	completionFile := filepath.Join(completionDir, "zd")
	f, err := os.Create(completionFile)
	if err != nil {
		return fmt.Errorf("failed to create completion file: %w", err)
	}
	defer f.Close()

	if err := rootCmd.GenBashCompletion(f); err != nil {
		return fmt.Errorf("failed to generate completion script: %w", err)
	}

	color.White("Completion script installed to: %s\n", completionFile)

	// Add to bashrc with bash-completion loading included
	bashrc := filepath.Join(home, ".bashrc")

	// Direct completion block - always load bash-completion first, then our completion
	sourceBlock := fmt.Sprintf(`
# zd completion - load bash-completion framework first
if [ -f /usr/share/bash-completion/bash_completion ] && [ -f %s ]; then
  . /usr/share/bash-completion/bash_completion
  . %s
elif [ -f /etc/bash_completion ] && [ -f %s ]; then
  . /etc/bash_completion
  . %s
fi
`, completionFile, completionFile, completionFile, completionFile)

	if err := addBlockToFile(bashrc, sourceBlock, "# zd completion"); err != nil {
		color.Yellow("⚠ Could not automatically update .bashrc\n")
		color.White("Please add this to your .bashrc:\n")
		color.Cyan("%s\n", sourceBlock)
	}

	return nil
}

func installZshCompletion(home string, rootCmd *cobra.Command) error {
	// Create completion directory
	completionDir := filepath.Join(home, ".zsh", "completion")
	if err := os.MkdirAll(completionDir, 0755); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	// Write completion script
	completionFile := filepath.Join(completionDir, "_zd")
	f, err := os.Create(completionFile)
	if err != nil {
		return fmt.Errorf("failed to create completion file: %w", err)
	}
	defer f.Close()

	if err := rootCmd.GenZshCompletion(f); err != nil {
		return fmt.Errorf("failed to generate completion script: %w", err)
	}

	// Add fpath line to .zshrc if not already present
	zshrc := filepath.Join(home, ".zshrc")
	fpathLine := fmt.Sprintf("fpath=(%s $fpath)\n", completionDir)
	autoloadLine := "autoload -Uz compinit && compinit\n"

	if err := addLineToFile(zshrc, fpathLine, "# zd completion"); err != nil {
		color.Yellow("⚠ Could not automatically add to .zshrc\n")
		color.White("Please add these lines to your .zshrc:\n")
		color.Cyan("  %s", fpathLine)
		color.Cyan("  %s\n", autoloadLine)
		return nil
	}

	addLineToFile(zshrc, autoloadLine, "# zd completion")

	return nil
}

func installFishCompletion(home string, rootCmd *cobra.Command) error {
	// Create completion directory
	completionDir := filepath.Join(home, ".config", "fish", "completions")
	if err := os.MkdirAll(completionDir, 0755); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	// Write completion script
	completionFile := filepath.Join(completionDir, "zd.fish")
	f, err := os.Create(completionFile)
	if err != nil {
		return fmt.Errorf("failed to create completion file: %w", err)
	}
	defer f.Close()

	if err := rootCmd.GenFishCompletion(f, true); err != nil {
		return fmt.Errorf("failed to generate completion script: %w", err)
	}

	return nil
}

func installPowerShellCompletion(home string, rootCmd *cobra.Command) error {
	// Create profile directory
	profileDir := filepath.Join(home, "Documents", "PowerShell")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	// Write completion script
	completionFile := filepath.Join(profileDir, "zd-completion.ps1")
	f, err := os.Create(completionFile)
	if err != nil {
		return fmt.Errorf("failed to create completion file: %w", err)
	}
	defer f.Close()

	if err := rootCmd.GenPowerShellCompletion(f); err != nil {
		return fmt.Errorf("failed to generate completion script: %w", err)
	}

	// Add source line to profile
	profileFile := filepath.Join(profileDir, "Microsoft.PowerShell_profile.ps1")
	sourceLine := fmt.Sprintf(". %s\n", completionFile)

	if err := addLineToFile(profileFile, sourceLine, "# zd completion"); err != nil {
		color.Yellow("⚠ Could not automatically add to PowerShell profile\n")
		color.White("Please add this line to your profile:\n")
		color.Cyan("  %s\n", sourceLine)
	}

	return nil
}

func addLineToFile(filepath, line, comment string) error {
	// Read existing content
	content, err := os.ReadFile(filepath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Check if line already exists
	if strings.Contains(string(content), line) {
		return nil // Already installed
	}

	// Append to file
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Add comment and line
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		f.WriteString("\n")
	}
	f.WriteString("\n" + comment + "\n")
	f.WriteString(line)

	return nil
}
