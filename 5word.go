package xeroxfont

import (
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
