package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

type CharacterMeta9700 struct {
	BlanksLeft int8
	Spacing byte // nonblank = $00, spacing = $80, null = ??
	GlyphOffset uint16  // halved.  multiply by 2 for byte offset.
	BottomOffset int8 // bitmap bottom offset from top of font bounds
	UnknownE int8 // negated bytes per line?
	//BitmapSize int16
	CellWidth uint16
	//UnknownG byte // accent?
}

func (m CharacterMeta9700) Offset(start int64) int {
	return int(start) + (int(m.GlyphOffset) * 2)
}

func (m CharacterMeta9700) String() string {
	spaceStr := "Space"
	if m.Spacing == 0x00 {
		spaceStr = "Non-Blank"
	}

	sb := &strings.Builder{}
	fmt.Fprintf(sb, "BlanksLeft:   $%02X %3d\n", m.BlanksLeft, m.BlanksLeft)
	fmt.Fprintf(sb, "Spacing:      $%02X %3d %s\n", m.Spacing, m.Spacing, spaceStr)
	fmt.Fprintf(sb, "GlyphOffset:  $%04X %4d\n", m.GlyphOffset, m.GlyphOffset)
	fmt.Fprintf(sb, "BottomOffset: $%02X %3d\n", uint8(m.BottomOffset), m.BottomOffset)
	fmt.Fprintf(sb, "UnknownE:     $%02X %3d\n", uint8(m.UnknownE), m.UnknownE)
	//fmt.Fprintf(sb, "BitmapSize:   $%02X %03d\n", m.BitmapSize, m.BitmapSize)
	fmt.Fprintf(sb, "CellWidth:    $%04X %3d\n", m.CellWidth, m.CellWidth)
	//fmt.Fprintf(sb, "UnknownG:    $%02X %03d\n", m.UnknownG, m.UnknownG)
	return sb.String()
}

func (m CharacterMeta9700) Is5Word() bool {
	return false
}

func (m CharacterMeta9700) IsSpacing() bool {
	return m.Spacing == 0x80
}

func parse9700Meta(reader io.Reader, lastChar int) ([]CharacterMeta, error) {
	var err error
	meta := []CharacterMeta{}
	for i := 0; i <= lastChar; i++ {
		m := &CharacterMeta9700{}
		err = binary.Read(reader, binary.LittleEndian, m)
		if err != nil {
			return nil, fmt.Errorf("Error reading CharacterMeta9700 at index %d: %w", i, err)
		}
		meta = append(meta, m)
	}

	return meta, nil
}
