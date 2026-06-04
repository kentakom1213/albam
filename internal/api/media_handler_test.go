package api

import "testing"

func TestParseMediaPathSupportsOriginalShortcut(t *testing.T) {
	photoID, kind, ok := parseMediaPath("/media/photo-001/original")
	if !ok {
		t.Fatal("parseMediaPath did not match")
	}
	if photoID != "photo-001" {
		t.Fatalf("photoID = %q, want photo-001", photoID)
	}
	if kind != VariantOriginal {
		t.Fatalf("kind = %q, want %q", kind, VariantOriginal)
	}
}

func TestParseMediaPathSupportsPhotoVariantPath(t *testing.T) {
	photoID, kind, ok := parseMediaPath("/media/photos/photo-001/original")
	if !ok {
		t.Fatal("parseMediaPath did not match")
	}
	if photoID != "photo-001" {
		t.Fatalf("photoID = %q, want photo-001", photoID)
	}
	if kind != VariantOriginal {
		t.Fatalf("kind = %q, want %q", kind, VariantOriginal)
	}
}
