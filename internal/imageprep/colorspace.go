package imageprep

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	pngSignature = "\x89PNG\r\n\x1a\n"

	// cICP code points for the sRGB primaries and the transfer functions of
	// sRGB and BT.709, which share the sRGB primaries.
	cicpSRGBPrimaries     = 1
	cicpSRGBTransfer      = 13
	cicpBT709Transfer     = 1
	chunkHeaderLength     = 8
	chunkChecksumLength   = 4
	minimumCICPDataLength = 2
)

// checkSRGB returns an error if the PNG declares a color space other than sRGB.
// The image pipeline neither reads nor writes color profiles, so a wide gamut
// original would silently be treated as sRGB and come out desaturated. PNGs
// without a color declaration are assumed to be sRGB. Data that is not a PNG is
// ignored so decoding can report it.
func checkSRGB(data []byte) error {
	if !bytes.HasPrefix(data, []byte(pngSignature)) {
		return nil
	}

	// color chunks precede the pixel data, so stop at the first IDAT
	for pos := len(pngSignature); pos+chunkHeaderLength <= len(data); {
		length := int(binary.BigEndian.Uint32(data[pos:]))
		kind := string(data[pos+4 : pos+chunkHeaderLength])
		start := pos + chunkHeaderLength
		end := start + length
		if kind == "IDAT" || end+chunkChecksumLength > len(data) {
			return nil
		}

		switch kind {
		case "iCCP":
			// the profile name is a NUL terminated keyword
			name, _, _ := bytes.Cut(data[start:end], []byte{0})
			if !strings.Contains(strings.ToLower(string(name)), "srgb") {
				return fmt.Errorf("embedded color profile %q", name)
			}
		case "cICP":
			if length >= minimumCICPDataLength {
				primaries, transfer := data[start], data[start+1]
				if primaries != cicpSRGBPrimaries {
					return fmt.Errorf("color primaries %d", primaries)
				}
				if transfer != cicpSRGBTransfer && transfer != cicpBT709Transfer {
					return fmt.Errorf("transfer function %d", transfer)
				}
			}
		}
		pos = end + chunkChecksumLength
	}
	return nil
}
