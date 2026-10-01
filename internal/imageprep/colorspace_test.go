package imageprep

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type chunk struct {
	kind string
	data []byte
}

// pngWithChunks returns a valid PNG with extra chunks inserted before its pixel data.
func pngWithChunks(t *testing.T, chunks ...chunk) []byte {
	t.Helper()
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, gradient(8, 8)); err != nil {
		t.Fatal(err)
	}
	data := encoded.Bytes()

	// the 8 byte signature and the 25 byte IHDR chunk come first
	const afterHeader = 8 + 25
	var inserted []byte
	for _, c := range chunks {
		length := make([]byte, 4)
		binary.BigEndian.PutUint32(length, uint32(len(c.data)))
		checksum := make([]byte, 4)
		binary.BigEndian.PutUint32(checksum, crc32.ChecksumIEEE(append([]byte(c.kind), c.data...)))
		inserted = slices.Concat(inserted, length, []byte(c.kind), c.data, checksum)
	}
	return slices.Concat(data[:afterHeader], inserted, data[afterHeader:])
}

// iccProfile is an iCCP chunk with a profile name; the profile itself is not inspected.
func iccProfile(name string) chunk {
	return chunk{"iCCP", slices.Concat([]byte(name), []byte{0, 0}, []byte("profile data"))}
}

func cicp(primaries, transfer byte) chunk {
	return chunk{"cICP", []byte{primaries, transfer, 0, 1}}
}

func TestCheckSRGB(t *testing.T) {
	tests := []struct {
		name    string
		chunks  []chunk
		wantErr string
	}{
		{name: "untagged PNG is assumed to be sRGB"},
		{name: "sRGB chunk", chunks: []chunk{{"sRGB", []byte{0}}}},
		{name: "sRGB profile as Lightroom names it", chunks: []chunk{iccProfile("sRGB IEC61966-2.1")}},
		{name: "profile name is matched ignoring case", chunks: []chunk{iccProfile("SRGB built-in")}},
		{name: "sRGB cICP", chunks: []chunk{cicp(1, 13)}},
		{name: "Display P3 profile", chunks: []chunk{iccProfile("Display P3")}, wantErr: "Display P3"},
		{name: "Adobe RGB profile", chunks: []chunk{iccProfile("Adobe RGB (1998)")}, wantErr: "Adobe RGB (1998)"},
		{name: "ProPhoto profile", chunks: []chunk{iccProfile("ProPhoto RGB")}, wantErr: "ProPhoto RGB"},
		{name: "Display P3 cICP", chunks: []chunk{cicp(12, 13)}, wantErr: "primaries 12"},
		{name: "BT.709 transfer shares the sRGB primaries", chunks: []chunk{cicp(1, 1)}},
		{name: "HDR primaries", chunks: []chunk{cicp(9, 16)}, wantErr: "primaries 9"},
		{name: "HDR transfer function", chunks: []chunk{cicp(1, 16)}, wantErr: "transfer function 16"},
		{
			name:    "wide gamut profile is rejected even when an sRGB chunk is also present",
			chunks:  []chunk{{"sRGB", []byte{0}}, iccProfile("Display P3")},
			wantErr: "Display P3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkSRGB(pngWithChunks(t, tt.chunks...))

			switch {
			case tt.wantErr == "" && err != nil:
				t.Errorf("checkSRGB() = %v, want no error", err)
			case tt.wantErr != "" && err == nil:
				t.Errorf("checkSRGB() = nil, want an error mentioning %q", tt.wantErr)
			case tt.wantErr != "" && !strings.Contains(err.Error(), tt.wantErr):
				t.Errorf("checkSRGB() = %q, want it to mention %q", err, tt.wantErr)
			}
		})
	}
}

func TestCheckSRGBIgnoresDataThatIsNotAPNG(t *testing.T) {
	if err := checkSRGB([]byte("not a png")); err != nil {
		t.Errorf("checkSRGB() = %v, decoding should report invalid files instead", err)
	}
}

func TestGenerateVariantsRefusesANonSRGBOriginal(t *testing.T) {
	originals, out := t.TempDir(), t.TempDir()
	path := filepath.Join(originals, "wide-gamut.png")
	if err := os.WriteFile(path, pngWithChunks(t, iccProfile("Display P3")), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := GenerateVariants(path, out, 50)

	if err == nil {
		t.Fatal("expected an error for a Display P3 original")
	}
	for _, want := range []string{"wide-gamut.png", "Display P3", "sRGB"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q", err, want)
		}
	}
	if got := fileNames(t, out); len(got) != 0 {
		t.Errorf("generated %v from a rejected original, want nothing", got)
	}
}
