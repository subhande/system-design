// Package transfer implements chunked/multipart uploads and downloads against
// the lumo-drive server + S3 presigned URLs.
package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// ChunkSize is the fixed multipart chunk size. It MUST match the server's
// chunk size (5 MiB), since the server derives total_chunks from file_size with
// this value and feeds chunk_index straight to the S3 part number.
const ChunkSize = 5 * 1024 * 1024

// Chunk describes one part of a file. Index is 1-based (S3 part numbers start
// at 1, and the server passes the index through unchanged).
type Chunk struct {
	Index    int
	Offset   int64
	Size     int64
	Checksum string // SHA-256 hex of the chunk's bytes
}

// FilePlan is the chunking plan plus whole-file checksum for one file.
type FilePlan struct {
	Size     int64
	Checksum string // SHA-256 hex of the whole file
	Chunks   []Chunk
}

// PlanFile reads the file once and returns its size, whole-file SHA-256, and a
// 1-based list of chunks each with its own SHA-256. A zero-byte file produces a
// single empty chunk so it still round-trips through the multipart flow.
func PlanFile(path string) (*FilePlan, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fileHash := sha256.New()
	buf := make([]byte, ChunkSize)

	plan := &FilePlan{}
	index := 1
	var offset int64
	for {
		n, readErr := io.ReadFull(f, buf)
		if n > 0 {
			fileHash.Write(buf[:n])

			chunkHash := sha256.Sum256(buf[:n])
			plan.Chunks = append(plan.Chunks, Chunk{
				Index:    index,
				Offset:   offset,
				Size:     int64(n),
				Checksum: hex.EncodeToString(chunkHash[:]),
			})
			index++
			offset += int64(n)
		}
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
		if readErr != nil {
			return nil, readErr
		}
	}

	plan.Size = offset
	plan.Checksum = hex.EncodeToString(fileHash.Sum(nil))

	if len(plan.Chunks) == 0 {
		// Empty file: represent as one zero-length chunk.
		empty := sha256.Sum256(nil)
		plan.Chunks = []Chunk{{Index: 1, Offset: 0, Size: 0, Checksum: hex.EncodeToString(empty[:])}}
	}
	return plan, nil
}

// FileChecksum returns the SHA-256 hex of a file's contents.
func FileChecksum(path string) (string, error) {
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

// openForRead opens a file for reading.
func openForRead(path string) (*os.File, error) {
	return os.Open(path)
}

// readChunk reads a chunk's bytes from an open file.
func readChunk(f *os.File, c Chunk) ([]byte, error) {
	buf := make([]byte, c.Size)
	if c.Size == 0 {
		return buf, nil
	}
	if _, err := f.ReadAt(buf, c.Offset); err != nil && err != io.EOF {
		return nil, fmt.Errorf("read chunk %d: %w", c.Index, err)
	}
	return buf, nil
}
