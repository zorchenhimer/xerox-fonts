package xeroxfont

import (
	"bytes"
	"fmt"
	"strings"
	"encoding/json"
)

type FontFormat byte

const (
	FF_5Word_Portrait   FontFormat = 0xA8
	FF_5Word_Landscape  FontFormat = 0xD0
	FF_5Word_Landscape2 FontFormat = 0x2F
	FF_5Word_IPortrait  FontFormat = 0x58
	FF_5Word_ILandscape FontFormat = 0xF8

	FF_5Word_Unknown FontFormat = 0xE6

	FF_9700_Portrait   FontFormat = 0x20
	FF_9700_Portrait2  FontFormat = 0x98
	FF_9700_Landscape  FontFormat = 0x48
	FF_9700_IPortrait  FontFormat = 0x80
	FF_9700_ILandscape FontFormat = 0x70
)

func (f FontFormat) String() string {
	switch f {
	case FF_5Word_Portrait:
		return "5Word Portrait"
	case FF_5Word_Landscape:
		return "5Word Landscape"
	case FF_5Word_IPortrait:
		return "5Word Inverted"
	case FF_5Word_ILandscape:
		return "5Word Inverted Landscape"

	case FF_9700_Portrait, FF_9700_Portrait2:
		return "9700 Portrait"
	case FF_9700_Landscape:
		return "9700 Landscape"
	case FF_9700_IPortrait:
		return "9700 Inverted Portrait"
	case FF_9700_ILandscape:
		return "9700 Inverted Landscape"
	}

	return "Unknown"
}

type Orientation byte

const (
	Portrait          Orientation = 0x50
	Landscape         Orientation = 0x4C
	InvertedPortrait  Orientation = 0x49
	InvertedLandscape Orientation = 0x4A
)

func (o Orientation) String() string {
	switch o {
	case Portrait:          return "Portrait"
	case Landscape:         return "Landscape"
	case InvertedPortrait:  return "Inverted Portrait"
	case InvertedLandscape: return "Inverted Landscape"
	}

	return "Unknown"
}

func IsOrientation(val byte) bool {
	switch Orientation(val) {
	case Portrait, Landscape, InvertedPortrait, InvertedLandscape:
		return true
	default:
		return false
	}
}

// 80 byte header, only new fonts saved in Elixir
type ExtraHeader struct {
	FontFormat FontFormat
	FontType byte // fixed or proportional
	UnknownC [16]byte

	FontNameA [6]byte
	FontNameB [6]byte
	UnknownA [4]byte
	UnknownB [12]byte

	_ [81]byte
	End byte
}

func (h ExtraHeader) String() string {
	sb := &strings.Builder{}

	fmt.Fprintf(sb, "FontFormat: $%02X %s\n", byte(h.FontFormat), h.FontFormat)
	fmt.Fprintf(sb, "FontType:   %c   $%02X\n", h.FontType, h.FontType)
	fmt.Fprintf(sb, "UnknownC:   $%X\n", h.UnknownC)
	fmt.Fprintf(sb, "FontNameA:  %q\n", bytes.ReplaceAll(h.FontNameA[:], []byte{0x00}, []byte{0x20}))
	fmt.Fprintf(sb, "FontNameB:  %q\n", bytes.ReplaceAll(h.FontNameB[:], []byte{0x00}, []byte{0x20}))
	fmt.Fprintf(sb, "UnknownA:   $%X\n", h.UnknownA)
	fmt.Fprintf(sb, "UnknownB:   $%X\n", h.UnknownB)
	fmt.Fprintf(sb, "End:        %c   $%X\n", h.End, h.End)

	return sb.String()
}

func (h ExtraHeader) Is9700() bool {
	switch h.FontFormat {
	case FF_5Word_Portrait, FF_5Word_Landscape, FF_5Word_Landscape2, FF_5Word_IPortrait, FF_5Word_ILandscape, FF_5Word_Unknown:
		return false
	default:
		return true
	}

	//if h.BitmapSize == 0 {
	//	return false
	//}
	//return true
}

// Standard header
type FontHeader struct {
	Orientation Orientation
	FontType byte // fixed or proportional
	PixelHeight uint16
	LineSpacing uint16
	FixedWidth uint16
	DistanceBelow uint16
	DistanceAbove uint16
	DistanceLeading uint16
	_ uint16
	LastCharacter uint16

	// BitmapSize and Unknown5Word don't seem to ever both have a value.  9700
	// fonts will fill BitmapSize, while 5Word fonts will fill Unknown5Word.
	BitmapSize uint16
	_ [2]byte
	Unknown5Word uint16

	FontName [6]byte
	Revision [2]byte
	_ [2]byte
	Version [2]byte
	Library [10]byte

	_ [210]byte
}

func (h FontHeader) MarshalJSON() ([]byte, error) {
	t := string(h.FontType)

	switch h.FontType {
	case 'P':
		t = "Proportional"
	case 'F':
		t = "Fixed"
	}

	data := struct {
		Orientation string
		FontType string
		PixelHeight int
		LineSpacing int
		FixedWidth int
		DistanceBelow int
		DistanceAbove int
		DistanceLeading int
		LastCharacter int

		BitmapSize int
		Unknown5Word int

		FontName string
		Revision string
		Version string
		Library string
	}{
		Orientation: h.Orientation.String(),
		FontType: t,
		PixelHeight: int(h.PixelHeight),
		LineSpacing: int(h.LineSpacing),
		FixedWidth: int(h.FixedWidth),
		DistanceBelow: int(h.DistanceBelow),
		DistanceAbove: int(h.DistanceAbove),
		DistanceLeading: int(h.DistanceLeading),
		LastCharacter: int(h.LastCharacter),

		BitmapSize: int(h.BitmapSize),
		Unknown5Word: int(h.Unknown5Word),

		FontName: fmt.Sprintf("%s", h.FontName),
		Revision: fmt.Sprintf("%s", h.Revision),
		Version: fmt.Sprintf("%s", h.Version),
		Library: fmt.Sprintf("%s", h.Library),
	}

	return json.MarshalIndent(data, "", "    ")
}

func (h FontHeader) Is9700() bool {
	if h.BitmapSize == 0 {
		return false
	}
	return true
}

func (h FontHeader) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "FontName: %q\n", bytes.ReplaceAll(h.FontName[:], []byte{0x00}, []byte{0x20}))
	fmt.Fprintf(sb, "Revision:        %q\n", h.Revision)
	fmt.Fprintf(sb, "Version:         %q\n", h.Version)
	fmt.Fprintf(sb, "Library:         %q\n", h.Library)
	fmt.Fprintf(sb, "Orientation:     %s\n", h.Orientation)
	fmt.Fprintf(sb, "FontType:        %c   $%02X\n", h.FontType, h.FontType)
	fmt.Fprintf(sb, "PixelHeight:     %-3d $%04X\n", h.PixelHeight, h.PixelHeight)
	fmt.Fprintf(sb, "LineSpacing:     %-3d $%04X\n", h.LineSpacing, h.LineSpacing)
	fmt.Fprintf(sb, "FixedWidth:      %-3d $%04X\n", h.FixedWidth, h.FixedWidth)
	fmt.Fprintf(sb, "DistanceBelow:   %-3d $%04X\n", h.DistanceBelow, h.DistanceBelow)
	fmt.Fprintf(sb, "DistanceAbove:   %-3d $%04X\n", h.DistanceAbove, h.DistanceAbove)
	fmt.Fprintf(sb, "DistanceLeading: %-3d $%04X\n", h.DistanceLeading, h.DistanceLeading)
	fmt.Fprintf(sb, "LastCharacter:   %-3d $%04X\n", h.LastCharacter, h.LastCharacter)
	return sb.String()
}

