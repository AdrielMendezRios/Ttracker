package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	installPath string
	force       bool
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Ttracker to the system",
	Long: `Install Ttracker to the system, making the 'tt' command available globally.
This will copy the binary to /usr/local/bin (requires sudo) or $HOME/.local/bin.

Example:
  tt install              # Install to system (auto-detect location)
  tt install --path /usr/local/bin  # Install to specific location
  tt install --force     # Force reinstall if already installed`,
	Run: installRun,
}

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Ttracker from the system",
	Long: `Remove Ttracker from the system, making the 'tt' command no longer available globally.
This will remove the binary from /usr/local/bin or $HOME/.local/bin.

Example:
  tt uninstall           # Uninstall from system (auto-detect location)
  tt uninstall --path /usr/local/bin  # Uninstall from specific location`,
	Run: uninstallRun,
}

func init() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)

	// Add flags to install command
	installCmd.Flags().StringVarP(&installPath, "path", "p", "", "Installation path (default: auto-detect)")
	installCmd.Flags().BoolVarP(&force, "force", "f", false, "Force reinstall if already installed")

	// Add flags to uninstall command
	uninstallCmd.Flags().StringVarP(&installPath, "path", "p", "", "Installation path (default: auto-detect)")
}

func installRun(cmd *cobra.Command, args []string) {
	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		return
	}

	// Determine installation path if not specified
	var useSudo bool
	if installPath == "" {
		// Try to use /usr/local/bin first
		if _, err := os.Stat("/usr/local/bin"); err == nil {
			installPath = "/usr/local/bin/tt"
			useSudo = true
		} else {
			// Fall back to $HOME/.local/bin
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("Error getting home directory: %v\n", err)
				return
			}
			localBin := filepath.Join(home, ".local", "bin")

			// Create .local/bin if it doesn't exist
			if err := os.MkdirAll(localBin, 0755); err != nil {
				fmt.Printf("Error creating .local/bin directory: %v\n", err)
				return
			}

			installPath = filepath.Join(localBin, "tt")
			useSudo = false
		}
	} else {
		// Check if the specified path requires sudo
		useSudo = strings.HasPrefix(installPath, "/usr/")
	}

	// Check if already installed
	if _, err := os.Stat(installPath); err == nil {
		if !force {
			fmt.Printf("Ttracker is already installed at %s\n", installPath)
			fmt.Println("Use --force to reinstall")
			return
		}
		fmt.Printf("Reinstalling Ttracker at %s\n", installPath)
	}

	// Copy the binary
	if useSudo {
		// Use sudo to copy to /usr/local/bin
		cmd := exec.Command("sudo", "cp", execPath, installPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error installing with sudo: %v\n", err)
			return
		}
	} else {
		// Copy to user's local bin
		if err := copyFile(execPath, installPath); err != nil {
			fmt.Printf("Error copying binary: %v\n", err)
			return
		}
	}

	// Make the binary executable
	if err := os.Chmod(installPath, 0755); err != nil {
		fmt.Printf("Error setting executable permissions: %v\n", err)
		return
	}

	fmt.Printf("Successfully installed Ttracker to %s\n", installPath)
	fmt.Println("You can now use 'tt' from anywhere!")
}

func uninstallRun(cmd *cobra.Command, args []string) {
	// Determine installation path if not specified
	var useSudo bool
	if installPath == "" {
		// Try to find the installation
		paths := []string{
			"/usr/local/bin/tt",
			"/usr/bin/tt",
		}

		// Add user's local bin
		if home, err := os.UserHomeDir(); err == nil {
			paths = append(paths, filepath.Join(home, ".local", "bin", "tt"))
		}

		// Find the first existing installation
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				installPath = path
				useSudo = strings.HasPrefix(path, "/usr/")
				break
			}
		}

		if installPath == "" {
			fmt.Println("Ttracker is not installed in any standard location")
			fmt.Println("Use --path to specify the installation location")
			return
		}
	} else {
		// Check if the specified path requires sudo
		useSudo = strings.HasPrefix(installPath, "/usr/")
	}

	// Check if the file exists
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		fmt.Printf("Ttracker is not installed at %s\n", installPath)
		return
	}

	// Remove the binary
	if useSudo {
		cmd := exec.Command("sudo", "rm", installPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error removing with sudo: %v\n", err)
			return
		}
	} else {
		if err := os.Remove(installPath); err != nil {
			fmt.Printf("Error removing binary: %v\n", err)
			return
		}
	}

	fmt.Printf("Successfully uninstalled Ttracker from %s\n", installPath)
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}

	return nil
}
