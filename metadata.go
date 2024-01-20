package xeroxfont

import (
	"encoding/binary"
	"fmt"
	"io"
	//"os"
	//"strings"
)

type CharacterMeta struct {
	BlanksLeft int
	GlyphOffset int
	Unknown uint16
	BitmapSize int16
	CellWidth int
	Spacing bool
}

func (m CharacterMeta) IsSpace() bool {
	return m.Spacing
}

func (m CharacterMeta) Offset(start int64) int {
	return int(start) + (int(m.GlyphOffset) * 2)
}

func (m CharacterMeta) String() string {
	return fmt.Sprintf("{CharacterMeta BlanksLeft:%04X (%d) GlyphOffset:%d BitmapSize:0x%04X CellWidth:%d Spacing:%t}",
		m.BlanksLeft,
		m.BlanksLeft & 0x7FFF,
		m.GlyphOffset,
		uint16(m.BitmapSize),
		m.CellWidth,
		m.Spacing,
	)
}

func (m CharacterMeta) Character(reader io.ReadSeeker, eot int64) (*Character, error) {
	c := &Character{
		BlanksLeft: int(m.BlanksLeft & 0x7FFF),
		GlyphOffset: int(m.GlyphOffset),
		CellWidth: int(m.CellWidth),
		BitmapSize: m.BitmapSize,
	}

	c.IsSpace = m.IsSpace()
	if c.IsSpace {
		return c, nil
	}

	currentOffset, err := reader.Seek(int64(m.Offset(eot)), 0)
	if err != nil {
		return nil, fmt.Errorf("unable to seek to glyph start $%04X: %w", m.Offset(eot), err)
	}

	//c.width = c.Width()
	//c.height = c.Height()
	//fmt.Fprintf(os.Stderr, "[meta char] 0x%06X: %s\n", currentOffset, m)

	c.glyph = make([]byte, c.Width()*c.Height())
	_, err = reader.Read(c.glyph)
	if err != nil {
		return nil, fmt.Errorf("error reading glyph bytes at offset %d with length of %d (%dx%d): %w", currentOffset, len(c.glyph), c.Width(), c.Height(), err)
	}

	if len(c.glyph)-1 % 2 != 0 {
		c.glyph = append(c.glyph, 0x00)
	}

	for i := 0; i < len(c.glyph)-1; i+=2 {
		c.glyph[i], c.glyph[i+1] = c.glyph[i+1], c.glyph[i]
	}

	return c, nil
}

func MetaFrom9700(reader io.Reader, lastChar int) ([]CharacterMeta, error) {
	var err error
	meta := []CharacterMeta{}
	for i := 0; i < lastChar; i++ {
		m := &CharacterMeta9700{}
		err = binary.Read(reader, binary.LittleEndian, m)
		if err != nil {
			return nil, fmt.Errorf("Error reading CharacterMeta9700 at index %d: %w", i, err)
		}
		meta = append(meta, m.Meta())
	}

	return meta, nil
}

func MetaFrom5Word(reader io.Reader, lastChar int) ([]CharacterMeta, error) {
	var err error
	meta := []CharacterMeta{}
	for i := 0; i < lastChar; i++ {
		m := &CharacterMeta5Word{}
		err = binary.Read(reader, binary.LittleEndian, m)
		if err != nil {
			return nil, fmt.Errorf("Error reading CharacterMeta5Word at index %d: %w", i, err)
		}
		meta = append(meta, m.Meta())
	}

	return meta, nil
}
