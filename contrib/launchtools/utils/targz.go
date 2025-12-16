package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// extractTarGz extracts a .tar.gz archive from the given byte slice into the specified destination directory.
func extractTarGz(data []byte, destDir string) error {
	uncompressedStream, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Zip Slip protection
		target := filepath.Join(destDir, header.Name) //nolint:gosec // G305: Standard Zip Slip protection is implemented
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", target)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode)) //nolint: gosec
			if err != nil {
				return err
			}
			// G110: Potential DoS vulnerability via decompression bomb
			// Limit the size of the decompressed file to 1GB
			const maxFileSize = 1 * 1024 * 1024 * 1024 // 1GB
			limitReader := io.LimitReader(tarReader, maxFileSize)

			if _, err := io.Copy(outFile, limitReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}
