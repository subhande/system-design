package transfer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/desubhan/system-design/lumo-drive/client/internal/api"
)

// DownloadFile downloads a remote file to destPath, resuming a partial transfer
// from a sibling ".part" file via an HTTP Range request when possible. If
// expectedChecksum is non-empty the downloaded content's SHA-256 is verified
// before the file is moved into place.
func DownloadFile(ctx context.Context, c *api.Client, fileID, destPath, expectedChecksum string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}

	url, err := c.GetDownloadURL(ctx, fileID)
	if err != nil {
		return fmt.Errorf("get download url: %w", err)
	}

	partPath := destPath + ".part"

	// Determine how many bytes we already have to resume from.
	var resumeFrom int64
	if info, statErr := os.Stat(partPath); statErr == nil {
		resumeFrom = info.Size()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if resumeFrom > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeFrom))
	}

	resp, err := s3HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If the server ignored the Range header (200 instead of 206), restart.
	flag := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if resp.StatusCode == http.StatusOK && resumeFrom > 0 {
		resumeFrom = 0
		flag = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	out, err := os.OpenFile(partPath, flag, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}

	if expectedChecksum != "" {
		sum, err := fileSHA256(partPath)
		if err != nil {
			return err
		}
		if sum != expectedChecksum {
			return fmt.Errorf("checksum mismatch for %s: got %s want %s", destPath, sum, expectedChecksum)
		}
	}

	return os.Rename(partPath, destPath)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
