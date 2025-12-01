package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/initia-labs/OPinit/contrib/launchtools/types"
	"github.com/pkg/errors"
)

// EnsureRelayerBinary checks if the relayer binary is installed and matches the expected version.
func EnsureRelayerBinary() (string, error) {
	version := types.RlyVersion
	binaryName := types.RlyBinaryName

	// Check if rly is already in PATH
	if path, err := exec.LookPath(binaryName); err == nil {
		if checkRelayerVersion(path, version) {
			return path, nil
		}
	}

	// Download to a temporary directory
	destDir := filepath.Join(os.TempDir(), "opinit-bin")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create binary destination directory")
	}

	destBin := filepath.Join(destDir, binaryName)
	if _, err := os.Stat(destBin); err == nil {
		if checkRelayerVersion(destBin, version) {
			return destBin, nil
		}
		// If version mismatch, remove the old binary
		_ = os.Remove(destBin)
	}

	goOS := runtime.GOOS
	goArch := runtime.GOARCH

	var osName string
	switch goOS {
	case "darwin":
		osName = "darwin"
	case "linux":
		osName = "linux"
	default:
		return "", fmt.Errorf("unsupported OS for automatic download: %s", goOS)
	}

	var archName string
	switch goArch {
	case "amd64":
		archName = "amd64"
	case "arm64":
		archName = "arm64"
	default:
		return "", fmt.Errorf("unsupported architecture for automatic download: %s", goArch)
	}

	versionNoV := strings.TrimPrefix(version, "v")
	downloadURL := fmt.Sprintf("https://github.com/cosmos/relayer/releases/download/%s/Cosmos.Relayer_%s_%s_%s.tar.gz", version, versionNoV, osName, archName)
	resp, err := http.Get(downloadURL) //nolint:gosec // G107: URL is constructed from constants and system properties
	if err != nil {
		return "", errors.Wrap(err, "failed to download rly binary")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download rly binary: status %d", resp.StatusCode)
	}

	// Read the body into a buffer to calculate checksum and extract
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read rly binary body")
	}

	// Verify checksum
	checksumURL := fmt.Sprintf("https://github.com/cosmos/relayer/releases/download/%s/SHA256SUMS-%s.txt", version, versionNoV)
	checksumResp, err := http.Get(checksumURL) //nolint:gosec // G107: URL is constructed from constants and system properties
	if err != nil {
		return "", errors.Wrap(err, "failed to download checksum file")
	}
	defer checksumResp.Body.Close()

	if checksumResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download checksum file: status %d", checksumResp.StatusCode)
	}

	checksumBody, err := io.ReadAll(checksumResp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read checksum file body")
	}

	hasher := sha256.New()
	hasher.Write(bodyBytes)
	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))

	if !verifyChecksum(string(checksumBody), calculatedChecksum, osName, archName) {
		return "", fmt.Errorf("checksum mismatch for downloaded binary (calculated: %s)", calculatedChecksum)
	}

	// Extract tar.gz using native Go
	if err := extractTarGz(bodyBytes, destDir); err != nil {
		return "", errors.Wrap(err, "failed to extract rly binary")
	}

	// The binary might be in a subdirectory (e.g. "Cosmos Relayer_2.6.0-rc.2_darwin_amd64/rly")
	// We need to find it and move it to destBin
	err = filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == binaryName {
			if path != destBin {
				return os.Rename(path, destBin)
			}
		}
		return nil
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to locate rly binary after extraction")
	}

	if _, err := os.Stat(destBin); err != nil {
		return "", errors.Wrap(err, "rly binary not found after extraction")
	}

	return destBin, nil
}

func checkRelayerVersion(binPath, expectedVersion string) bool {
	cmd := exec.Command(binPath, "version")
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	// Expected output format: "version: 2.6.0" or similar
	// We'll check if the output contains the version string (without 'v' prefix if present in expectedVersion)
	versionStr := expectedVersion
	if len(versionStr) > 0 && versionStr[0] == 'v' {
		versionStr = versionStr[1:]
	}

	return strings.Contains(string(out), versionStr)
}

func verifyChecksum(checksumsContent, calculatedChecksum, osName, archName string) bool {
	lines := strings.Split(checksumsContent, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		hash := parts[0]
		filename := strings.Join(parts[1:], " ")

		if hash == calculatedChecksum {
			// Verify that the filename corresponds to the requested OS and Arch
			// This handles cases where the filename in checksums differs from the download URL (e.g. spaces, rc versions)
			if strings.Contains(filename, osName) && strings.Contains(filename, archName) {
				return true
			}
		}
	}
	return false
}
