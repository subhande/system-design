package state

import "testing"

func TestFileRoundTrip(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer st.Close()

	rec := FileRecord{
		RemoteID:      "file-1",
		Name:          "a.txt",
		RelPath:       "docs/a.txt",
		SizeBytes:     42,
		Checksum:      "abc",
		LocalMtime:    1000,
		LocalChecksum: "abc",
		Status:        "uploaded",
	}
	if err := st.UpsertFile(rec); err != nil {
		t.Fatalf("UpsertFile: %v", err)
	}

	byPath, err := st.FileByPath("docs/a.txt")
	if err != nil || byPath == nil {
		t.Fatalf("FileByPath: %v rec=%v", err, byPath)
	}
	if byPath.RemoteID != "file-1" || byPath.SizeBytes != 42 {
		t.Errorf("unexpected record: %+v", byPath)
	}

	// Upsert with new size should update in place, not duplicate.
	rec.SizeBytes = 99
	if err := st.UpsertFile(rec); err != nil {
		t.Fatalf("UpsertFile update: %v", err)
	}
	byID, err := st.FileByID("file-1")
	if err != nil || byID == nil {
		t.Fatalf("FileByID: %v", err)
	}
	if byID.SizeBytes != 99 {
		t.Errorf("size = %d, want 99", byID.SizeBytes)
	}

	all, err := st.AllFiles()
	if err != nil {
		t.Fatalf("AllFiles: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("AllFiles len = %d, want 1", len(all))
	}
}

func TestMetaAndChangeID(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer st.Close()

	id, err := st.LastChangeID()
	if err != nil || id != 0 {
		t.Fatalf("initial LastChangeID = %d, %v", id, err)
	}
	if err := st.SetLastChangeID(123); err != nil {
		t.Fatalf("SetLastChangeID: %v", err)
	}
	id, err = st.LastChangeID()
	if err != nil || id != 123 {
		t.Fatalf("LastChangeID = %d, %v; want 123", id, err)
	}
}

func TestMissingLookupsReturnNil(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer st.Close()

	if f, err := st.FileByPath("nope"); err != nil || f != nil {
		t.Errorf("FileByPath(nope) = %v, %v; want nil, nil", f, err)
	}
	if f, err := st.FolderByID("nope"); err != nil || f != nil {
		t.Errorf("FolderByID(nope) = %v, %v; want nil, nil", f, err)
	}
}
