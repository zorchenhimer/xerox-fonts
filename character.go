package xeroxfont

import (
	"fmt"
	//"io"
	"os"
	"image"
	"image/png"
	"image/color"
	"encoding/json"
)

type Character struct {
	IsSpace bool
	BlanksLeft int
	GlyphOffset int
	CellWidth int
	Value rune

	BitmapSize int16
	GlyphCount int
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

func (c *Character) MarshalJSON() ([]byte, error) {
	data := struct {
		Value rune
		IsSpace bool
		BlanksLeft int
		CellWidth int
		Width int
		Height int
	}{
		Value: c.Value,
		IsSpace: c.IsSpace,
		BlanksLeft: c.BlanksLeft,
		CellWidth: c.CellWidth,
		Width: c.Width(),
		Height: c.Height(),
	}

	return json.MarshalIndent(data, "", "    ")
}

func (c *Character) Height() int {
	return int(abs(c.BitmapSize >> 9))*8
}

func (c *Character) Width() int {
	return int(abs(c.BitmapSize) & 0x1FF)
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

	height := c.Height()
	width  := c.Width()

	var img *image.RGBA
	if height % 8 != 0 {
		img = image.NewRGBA(image.Rect(0, 0, width, height+4))
	} else {
		img = image.NewRGBA(image.Rect(0, 0, width, height))
	}

	if c.IsSpace {
		return img
	}

	x, y := 0, height
	on := color.White

	c.GlyphCount = 0
	for _, b := range c.glyph {
		for i := 7; i > -1; i-- {
			v := (b >> i) & 0x01
			if v == 1 {
				img.Set(height-y, height-x-1, on)
			}
			x++
		}
		if x >= height {
			y--
			x = 0
		}
		c.GlyphCount++
	}

	c.img = img
	return img
}

func (c *Character) Image() image.Image {
	if c.img != nil {
		return c.img
	}

	height := c.Height()
	width  := c.Width()

	var img *image.RGBA
	if height % 8 != 0 {
		img = image.NewRGBA(image.Rect(0, 0, width, height+4))
	} else {
		img = image.NewRGBA(image.Rect(0, 0, width, height))
	}

	if c.IsSpace {
		return img
	}

	x, y := 0, height
	on := color.White
	off := color.Black

	c.GlyphCount = 0
	for _, b := range c.glyph {
		for i := 7; i > -1; i-- {
			v := (b >> i) & 0x01
			if v == 1 {
				img.Set(height-y, height-x-1, on)
			} else {
				img.Set(height-y, height-x-1, off)
			}
			x++
		}
		if x >= height {
			y--
			x = 0
		}
		c.GlyphCount++
	}

	c.img = img
	return img
}

func (c *Character) RawGlyph() []byte {
	return c.glyph
}
