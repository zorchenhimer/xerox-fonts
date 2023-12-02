package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

type CharacterMeta5Word struct {
	BlanksLeft uint8
	Spacing byte
	GlyphOffset uint16
	//Unknown3 byte
	Unknown45 uint16
	//Unknown5 byte
	Unknown67 uint16
	//Unknown7 byte
	CellWidth uint16
	//Unknown9 byte
}

func (m CharacterMeta5Word) Offset(start int64) int {
	return int(start) + (int(m.GlyphOffset) * 2)
}

func (m CharacterMeta5Word) String() string {
	sb := &strings.Builder{}

	spaceStr := "Space"
	if m.Spacing == 0x00 {
		spaceStr = "Non-Blank"
	}

	fmt.Fprintf(sb, "BlanksLeft:  $%02X\n", m.BlanksLeft)
	fmt.Fprintf(sb, "Spacing:     $%02X %s\n", m.Spacing, spaceStr)
	fmt.Fprintf(sb, "GlyphOffset: $%04X\n", m.GlyphOffset)
	fmt.Fprintf(sb, "Unknown45:   $%04X %3d\n", m.Unknown45, int16(m.Unknown45))
	//fmt.Fprintf(sb, "Unknown5:    $%02X\n", m.Unknown5)
	fmt.Fprintf(sb, "Unknown67:   $%04X %03d\n", m.Unknown67, int16(m.Unknown67))
	//fmt.Fprintf(sb, "Unknown7:    $%02X\n", m.Unknown7)
	fmt.Fprintf(sb, "CellWidth:   $%04X\n", m.CellWidth)

	return sb.String()
}

func (m CharacterMeta5Word) Is5Word() bool {
	return true
}

func (m CharacterMeta5Word) IsSpace() bool {
	return m.Spacing == 0x80
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
		meta = append(meta, m)
	}

	return meta, nil
}
