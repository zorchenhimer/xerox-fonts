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
	Value rune

	bitmapSize int16
	width int
	height int

	glyph []byte
	img image.Image
	mask image.Image
}

func From5Word(raw *CharacterMeta5Word) (*Character, error) {
	return nil, fmt.Errorf("From5Word() not implemented")
}

func abs(x int16) int16 {
	if x < 0 {
		return -x
	}
	return x
}

func (c *Character) Width() int {
	return int(abs(c.bitmapSize >> 9))*8
}

func (c *Character) Height() int {
	return int(abs(c.bitmapSize) & 0x1FF)
}

func From9700(raw *CharacterMeta9700, reader io.ReadSeeker, eot int64) (*Character, error) {
	c := &Character{
		BlanksLeft: int(raw.BlanksLeft & 0x7FFF),
		GlyphOffset: int(raw.GlyphOffset),
		CellWidth: int(raw.CellWidth),
		bitmapSize: raw.BitmapSize,
	}

	c.IsSpace = raw.IsSpace()
	if c.IsSpace {
		return c, nil
	}

	_, err := reader.Seek(int64(raw.Offset(eot)), 0)
	if err != nil {
		return nil, fmt.Errorf("unable to seek to glyph start $%04X: %w", raw.Offset(eot), err)
	}

	c.width = c.Width()
	c.height = c.Height()

	c.glyph = make([]byte, c.width*c.height)
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

	return c, nil
}

func (c *Character) WriteImage(filename string) error {
	outfile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("unable to create %s: %w", filename, err)
	}
	defer outfile.Close()

	err = png.Encode(outfile, c.Image())
	if err != nil {
		return fmt.Errorf("Error encoding PNG for %s: %w", filename, err)
	}

	return nil
}

func (c *Character) Mask() image.Image {
	if c.mask != nil {
		return c.mask
	}

	var img *image.RGBA
	if c.width % 8 != 0 {
		img = image.NewRGBA(image.Rect(0, 0, c.height, c.width+4))
	} else {
		img = image.NewRGBA(image.Rect(0, 0, c.height, c.width))
	}

	if c.IsSpace {
		return img
	}

	x, y := 0, c.height
	on := color.Black

	for _, b := range c.glyph {
		for i := 7; i > -1; i-- {
			v := (b >> i) & 0x01
			if v == 1 {
				img.Set(c.height-y, c.width-x-1, on)
			}
			x++
		}
		if x >= c.width {
			y--
			x = 0
		}
	}

	c.img = img
	return img
}

func (c *Character) Image() image.Image {
	if c.img != nil {
		return c.img
	}

	var img *image.RGBA
	if c.width % 8 != 0 {
		img = image.NewRGBA(image.Rect(0, 0, c.height, c.width+4))
	} else {
		img = image.NewRGBA(image.Rect(0, 0, c.height, c.width))
	}

	if c.IsSpace {
		return img
	}

	x, y := 0, c.height
	on := color.White
	off := color.Black

	for _, b := range c.glyph {
		for i := 7; i > -1; i-- {
			v := (b >> i) & 0x01
			if v == 1 {
				img.Set(c.height-y, c.width-x-1, on)
			} else {
				img.Set(c.height-y, c.width-x-1, off)
			}
			x++
		}
		if x >= c.width {
			y--
			x = 0
		}
	}

	c.img = img
	return img
}
