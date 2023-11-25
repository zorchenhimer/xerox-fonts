package main

import (
	"bytes"
	"fmt"
	"strings"
)

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
	UnknownD uint16
	LastCharacter uint16

	UnknownC [6]byte
	FontName [6]byte
	Revision [2]byte
	_ [2]byte
	Version [2]byte
	Library [10]byte

	_ [210]byte
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

