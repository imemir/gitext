package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gitext/gitext/pkg/ui"
	"github.com/spf13/cobra"
)

const (
	githubRepoOwner = "imemir"
	githubRepoName  = "gitext"
	githubAPIURL    = "https://api.github.com/repos/%s/%s/releases/latest"
	downloadURL     = "https://github.com/%s/%s/releases/download/%s/%s"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func NewSelfUpdateCmd(opts *Options) *cobra.Command {
	var yesFlag bool

	cmd := &cobra.Command{
		Use:   "self-update",
		Short: "Update gitext to the latest version",
		Long: `Check for the latest version of gitext and update the binary if a newer version is available.
This command downloads the latest release from GitHub and replaces the current binary.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			output := ui.NewOutput(opts.Verbose)

			currentVersion := opts.Version
			if currentVersion == "" {
				currentVersion = "dev"
			}

			output.Info("Current version: %s", currentVersion)

			// Get latest release from GitHub
			output.Doing("Checking for latest version...")
			latestRelease, err := getLatestRelease()
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			latestVersion := latestRelease.TagName
			output.Info("Latest version: %s", latestVersion)

			// Compare versions
			if !isNewerVersion(latestVersion, currentVersion) {
				output.Success("You are already using the latest version (%s)", currentVersion)
				return nil
			}

			// Confirm update
			if !yesFlag {
				output.Info("Update available: %s -> %s", currentVersion, latestVersion)
				fmt.Print("Do you want to update? [Y/n]: ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) == "n" {
					output.Info("Update cancelled")
					return nil
				}
			}

			// Detect OS and architecture
			goos := runtime.GOOS
			goarch := runtime.GOARCH

			// Map Go OS/arch to release asset naming
			var assetName string
			switch goos {
			case "linux":
				assetName = fmt.Sprintf("gitext-linux-%s", goarch)
			case "darwin":
				assetName = fmt.Sprintf("gitext-darwin-%s", goarch)
			case "windows":
				assetName = fmt.Sprintf("gitext-windows-%s.exe", goarch)
			default:
				return fmt.Errorf("unsupported operating system: %s", goos)
			}

			// Find the asset in the release
			var downloadURL string
			for _, asset := range latestRelease.Assets {
				if asset.Name == assetName {
					downloadURL = asset.BrowserDownloadURL
					break
				}
			}

			if downloadURL == "" {
				return fmt.Errorf("binary not found for %s/%s. Available assets: %v", goos, goarch, getAssetNames(latestRelease.Assets))
			}

			// Get current executable path
			execPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("failed to get executable path: %w", err)
			}

			// Resolve symlinks to get actual path
			execPath, err = filepath.EvalSymlinks(execPath)
			if err != nil {
				return fmt.Errorf("failed to resolve executable path: %w", err)
			}

			output.Doing("Downloading %s...", latestVersion)

			// Download to temporary file
			tempFile := execPath + ".tmp"
			if err := downloadFile(downloadURL, tempFile); err != nil {
				return fmt.Errorf("failed to download binary: %w", err)
			}

			// Verify downloaded file
			info, err := os.Stat(tempFile)
			if err != nil {
				os.Remove(tempFile)
				return fmt.Errorf("failed to verify downloaded file: %w", err)
			}
			if info.Size() == 0 {
				os.Remove(tempFile)
				return fmt.Errorf("downloaded file is empty")
			}

			// Set executable permissions (Unix)
			if goos != "windows" {
				if err := os.Chmod(tempFile, 0755); err != nil {
					os.Remove(tempFile)
					return fmt.Errorf("failed to set executable permissions: %w", err)
				}
			}

			// Replace current binary
			output.Doing("Installing new version...")
			if err := replaceBinary(tempFile, execPath, goos); err != nil {
				os.Remove(tempFile)
				return fmt.Errorf("failed to replace binary: %w", err)
			}

			output.Success("Successfully updated to version %s", latestVersion)
			output.Next("restart your terminal or run: gitext --version")

			return nil
		},
	}

	cmd.Flags().BoolVar(&yesFlag, "yes", false, "Skip confirmation prompt")

	return cmd
}

func getLatestRelease() (*githubRelease, error) {
	url := fmt.Sprintf(githubAPIURL, githubRepoOwner, githubRepoName)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func isNewerVersion(latest, current string) bool {
	// Handle "dev" version - always consider update available
	if current == "dev" || current == "" {
		return true
	}

	// Remove "v" prefix if present
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	// Simple string comparison for semantic versions
	// This works because semantic versions sort lexicographically when formatted correctly
	return latest > current
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func replaceBinary(tempFile, targetFile string, goos string) error {
	// On Windows, we need to handle file locking differently
	if goos == "windows" {
		// Try to remove the old file first (may fail if in use)
		if err := os.Remove(targetFile); err != nil {
			// If removal fails, try renaming the old file
			backupFile := targetFile + ".old"
			os.Remove(backupFile) // Remove any existing backup
			if err := os.Rename(targetFile, backupFile); err != nil {
				return fmt.Errorf("failed to backup old binary: %w. You may need to close gitext and try again", err)
			}
		}
	}

	// Atomic rename (works on Unix, and on Windows after removing/renaming old file)
	if err := os.Rename(tempFile, targetFile); err != nil {
		return fmt.Errorf("failed to replace binary: %w. You may need to run with sudo/admin privileges", err)
	}

	return nil
}

func getAssetNames(assets []struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}) []string {
	names := make([]string, len(assets))
	for i, asset := range assets {
		names[i] = asset.Name
	}
	return names
}
