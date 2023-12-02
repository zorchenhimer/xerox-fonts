package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

/*

	abs(BitmapSize) & 0x1FF = height
	abs(BitmapSize >> 9) * 8  = width

*/

type CharacterMeta9700 struct {
	BlanksLeft uint16
	//Spacing byte // nonblank = $00, spacing = $80, null = ??
	GlyphOffset uint16  // halved.  multiply by 2 for byte offset.
	BitmapSize int16 // packed dimensions
	CellWidth uint16
}

func (m CharacterMeta9700) IsSpace() bool {
	if m.BlanksLeft & 0x8000 != 0 {
		return true
	}
	return false
}

func (m CharacterMeta9700) Offset(start int64) int {
	return int(start) + (int(m.GlyphOffset) * 2)
}

func (m CharacterMeta9700) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "BlanksLeft:   $%04X %3d\n", m.BlanksLeft & 0x7FFF, m.BlanksLeft & 0x7FFF)
	fmt.Fprintf(sb, "Spacing:      %t\n", m.IsSpace())
	fmt.Fprintf(sb, "GlyphOffset:  $%04X %4d\n", m.GlyphOffset, m.GlyphOffset)
	fmt.Fprintf(sb, "BitmapSize:   $%04X %4d\n", m.BitmapSize, m.BitmapSize)
	fmt.Fprintf(sb, "CellWidth:    $%04X %3d\n", m.CellWidth, m.CellWidth)
	return sb.String()
}

func (m CharacterMeta9700) Is5Word() bool {
	return false
}

//func (m CharacterMeta9700) IsSpacing() bool {
//	return m.Spacing == 0x80
//}

func parse9700Meta(reader io.Reader, lastChar int) ([]CharacterMeta, error) {
	var err error
	meta := []CharacterMeta{}
	for i := 0; i < lastChar; i++ {
		m := &CharacterMeta9700{}
		err = binary.Read(reader, binary.LittleEndian, m)
		if err != nil {
			return nil, fmt.Errorf("Error reading CharacterMeta9700 at index %d: %w", i, err)
		}
		meta = append(meta, m)
	}

	return meta, nil
}
