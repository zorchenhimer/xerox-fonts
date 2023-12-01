package main

import (
	"fmt"
	"io"
	"os"
	"image"
	"image/png"
	"image/color"
)

type Character struct {
	IsSpace bool
	BlanksLeft int
	GlyphOffset int
	CellWidth int

	width int
	height int

	glyph []byte
}

func From5Word(raw *CharacterMeta5Word) (*Character, error) {
	return nil, fmt.Errorf("From5Word() not implemented")
}

func From9700(raw *CharacterMeta9700, reader io.ReadSeeker, eot int64) (*Character, error) {
	c := &Character{
		BlanksLeft: int(raw.BlanksLeft),
		GlyphOffset: int(raw.GlyphOffset),
	}

	if raw.Spacing == 0x80 {
		c.IsSpace = true
		return c, nil
	}

	_, err := reader.Seek(int64(raw.Offset(eot)), 0)
	if err != nil {
		return nil, fmt.Errorf("unable to seek to glyph start $%04X: %w", raw.Offset(eot), err)
	}

	c.width = int((int8(raw.UnknownE) >> 1)) * -8
	//fmt.Println("bitWidth:", bitWidth)
	//c.width = bitwidth

	l := int(raw.BottomOffset*-1) * int(raw.UnknownE*-1)
	if l <= 0 {
		return nil, fmt.Errorf("invalid length: (%d*-1) * (%d*-1) %d", raw.BottomOffset, raw.UnknownE, l)
	}

	c.glyph = make([]byte, l)
	_, err = reader.Read(c.glyph)
	if err != nil {
		return nil, fmt.Errorf("error reading glyph bytes: %w", err)
	}

	if len(c.glyph)-1 % 2 != 0 {
		c.glyph = append(c.glyph, 0x00)
	}

	for i := 0; i < len(c.glyph)-1; i+=2 {
		c.glyph[i], c.glyph[i+1] = c.glyph[i+1], c.glyph[i]
	}

	//c.width = int(raw.UnknownE*-1)*4
	c.height = int(raw.BottomOffset*-1)

	return c, nil
}

func (c *Character) WriteImage(filename string) error {
	var img *image.RGBA
	if c.width % 8 != 0 {
		img = image.NewRGBA(image.Rect(0, 0, c.height, c.width+4))
	} else {
		img = image.NewRGBA(image.Rect(0, 0, c.height, c.width))
	}
	x, y := 0, c.height
	on := color.White
	off := color.Black

	for _, b := range c.glyph {
		for i := 7; i > -1; i-- {
			v := (b >> i) & 0x01
			if v == 1 {
				img.Set(c.height-y, c.width-x, on)
			} else {
				img.Set(c.height-y, c.width-x, off)
			}
			x++
		}
		if x >= c.width {
			y--
			x = 0
		}
	}

	outfile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("unable to create %s: %w", filename, err)
	}
	defer outfile.Close()
	err = png.Encode(outfile, img)
	if err != nil {
		return fmt.Errorf("Error encoding PNG for %s: %w", filename, err)
	}

	return nil
}