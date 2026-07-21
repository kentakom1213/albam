package scanner

import "testing"

func TestOrientedImageSize(t *testing.T) {
	tests := []struct {
		name         string
		orientation  *int
		wantW, wantH int
	}{
		{
			name:  "nil orientation keeps dimensions",
			wantW: 4000,
			wantH: 3000,
		},
		{
			name:        "normal orientation keeps dimensions",
			orientation: intPtr(1),
			wantW:       4000,
			wantH:       3000,
		},
		{
			name:        "rotated clockwise swaps dimensions",
			orientation: intPtr(6),
			wantW:       3000,
			wantH:       4000,
		},
		{
			name:        "rotated counterclockwise swaps dimensions",
			orientation: intPtr(8),
			wantW:       3000,
			wantH:       4000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := orientedImageSize(4000, 3000, tt.orientation)
			if gotW != tt.wantW || gotH != tt.wantH {
				t.Fatalf("orientedImageSize() = (%d, %d), want (%d, %d)", gotW, gotH, tt.wantW, tt.wantH)
			}
		})
	}
}

func intPtr(value int) *int {
	return &value
}
