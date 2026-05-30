package transfer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// writeTemp writes data to a temp file and returns its path.
func writeTemp(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "blob")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return path
}

func sha(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestPlanFile(t *testing.T) {
	cases := []struct {
		name       string
		size       int
		wantChunks int
	}{
		{"empty", 0, 1},
		{"small", 10, 1},
		{"exactly one chunk", ChunkSize, 1},
		{"one byte over", ChunkSize + 1, 2},
		{"two and a half chunks", 2*ChunkSize + 100, 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := bytes.Repeat([]byte{0xAB}, tc.size)
			path := writeTemp(t, data)

			plan, err := PlanFile(path)
			if err != nil {
				t.Fatalf("PlanFile: %v", err)
			}

			if plan.Size != int64(tc.size) {
				t.Errorf("size = %d, want %d", plan.Size, tc.size)
			}
			if len(plan.Chunks) != tc.wantChunks {
				t.Fatalf("chunks = %d, want %d", len(plan.Chunks), tc.wantChunks)
			}
			if plan.Checksum != sha(data) {
				t.Errorf("file checksum mismatch")
			}

			// Indices must be 1-based and contiguous; sizes must sum to total.
			var total int64
			for i, ch := range plan.Chunks {
				if ch.Index != i+1 {
					t.Errorf("chunk %d has index %d, want %d", i, ch.Index, i+1)
				}
				total += ch.Size
			}
			if tc.size > 0 && total != int64(tc.size) {
				t.Errorf("chunk sizes sum = %d, want %d", total, tc.size)
			}

			// First chunk checksum must match the bytes at its offset.
			if tc.size > 0 {
				want := sha(data[:plan.Chunks[0].Size])
				if plan.Chunks[0].Checksum != want {
					t.Errorf("chunk 0 checksum mismatch")
				}
			}
		})
	}
}

func TestFileChecksum(t *testing.T) {
	data := []byte("hello lumo")
	path := writeTemp(t, data)
	got, err := FileChecksum(path)
	if err != nil {
		t.Fatalf("FileChecksum: %v", err)
	}
	if got != sha(data) {
		t.Errorf("checksum = %s, want %s", got, sha(data))
	}
}
