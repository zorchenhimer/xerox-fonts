package xeroxfont

import (
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

func (m CharacterMeta9700) Meta() CharacterMeta {
	return CharacterMeta{
		BlanksLeft: int(m.BlanksLeft & 0x7FFF),
		GlyphOffset: int(m.GlyphOffset),
		BitmapSize: m.BitmapSize,
		CellWidth: int(m.CellWidth),
		Spacing: m.BlanksLeft & 0x8000 == 0x8000,
	}
}
