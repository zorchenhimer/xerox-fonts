package xeroxfont

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

type CharacterMeta5Word struct {
	BlanksLeft uint16
	//Spacing byte
	GlyphOffset uint16
	Unknown uint16
	BitmapSize int16
	CellWidth uint16
}

func (m CharacterMeta5Word) Meta() CharacterMeta {
	return CharacterMeta{
		BlanksLeft: int(m.BlanksLeft & 0x7FFF),
		GlyphOffset: int(m.GlyphOffset),
		Unknown: m.Unknown,
		BitmapSize: m.BitmapSize,
		CellWidth: int(m.CellWidth),
		Spacing: m.BlanksLeft & 0x8000 == 0x8000,
	}
}

func (m CharacterMeta5Word) IsSpace() bool {
	if m.BlanksLeft & 0x8000 != 0 {
		return true
	}
	return false
}

func (m CharacterMeta5Word) Offset(start int64) int {
	return int(start) + (int(m.GlyphOffset) * 2)
}

func (m CharacterMeta5Word) String() string {
	sb := &strings.Builder{}

	fmt.Fprintf(sb, "BlanksLeft:   $%04X %3d\n", m.BlanksLeft & 0x7FFF, m.BlanksLeft & 0x7FFF)
	fmt.Fprintf(sb, "Spacing:      %t\n", m.IsSpace())
	fmt.Fprintf(sb, "GlyphOffset:  $%04X %4d\n", m.GlyphOffset, m.GlyphOffset)
	fmt.Fprintf(sb, "BitmapSize:   $%04X %4d\n", m.BitmapSize, m.BitmapSize)
	fmt.Fprintf(sb, "Width:        %4d\n", int(abs(m.BitmapSize >> 9)) * 8)
	fmt.Fprintf(sb, "Height:       %4d\n", int(abs(m.BitmapSize) & 0x1FF))
	fmt.Fprintf(sb, "CellWidth:    $%04X %3d\n", m.CellWidth, m.CellWidth)

	//fmt.Fprintf(sb, "BlanksLeft:  $%02X\n", m.BlanksLeft)
	//fmt.Fprintf(sb, "Spacing:     $%02X %s\n", m.Spacing, spaceStr)
	//fmt.Fprintf(sb, "GlyphOffset: $%04X\n", m.GlyphOffset)
	//fmt.Fprintf(sb, "Unknown45:   $%04X %3d\n", m.Unknown45, int16(m.Unknown45))
	////fmt.Fprintf(sb, "Unknown5:    $%02X\n", m.Unknown5)
	//fmt.Fprintf(sb, "Unknown67:   $%04X %03d\n", m.Unknown67, int16(m.Unknown67))
	////fmt.Fprintf(sb, "Unknown7:    $%02X\n", m.Unknown7)
	//fmt.Fprintf(sb, "CellWidth:   $%04X\n", m.CellWidth)

	return sb.String()
}

func (m CharacterMeta5Word) Is5Word() bool {
	return true
}

func parse5WordMeta(reader io.Reader, lastChar int) ([]CharacterMeta, error) {
	var err error
	meta := []CharacterMeta{}
	for i := 0; i <= lastChar; i++ {
		m := &CharacterMeta5Word{}
		err = binary.Read(reader, binary.LittleEndian, m)
		if err != nil {
			return nil, fmt.Errorf("Error reading CharacterMeta5Word at index %d: %w", i, err)
		}
		meta = append(meta, m.Meta())
	}

	return meta, nil
}
