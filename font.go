package xeroxfont

import (
	"fmt"
	"io"
	"os"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"strings"
	"log"
	//"image/png"
)

type Font struct {
	Header *FontHeader
	Characters map[rune]*Character
	Widths [256]uint8
}

func LoadFont(reader io.ReadSeeker) (*Font, error) {
	var val byte
	err := binary.Read(reader, binary.LittleEndian, &val)
	if err != nil {
		return nil, fmt.Errorf("Error reading first byte: %w", err)
	}

	reader.Seek(0, 0)
	readOffset := 0
	font := &Font{
		Header: &FontHeader{},
		Characters: make(map[rune]*Character),
		Widths: [256]uint8{},
	}

	// Just ignore the extra header for now.
	if !IsOrientation(val) {
		log.Println("skipping extra header")
		_, err = reader.Seek(128, io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("Error skipping extra header: %w", err)
		}
		readOffset += 128
	}

	err = binary.Read(reader, binary.LittleEndian, font.Header)
	if err != nil {
		return nil, fmt.Errorf("Error reading main header: %w", err)
	}
	readOffset += binary.Size(font.Header)
	log.Println(font.Header)

	err = binary.Read(reader, binary.LittleEndian, &font.Widths)
	if err != nil {
		return nil, fmt.Errorf("Error reading width table: %w", err)
	}
	readOffset += binary.Size(font.Widths)

	var metaCount int = int(font.Header.LastCharacter)
	if font.Header.LastCharacter < 128 {
		metaCount = 128
	} else if font.Header.LastCharacter % 128 != 0 {
		mod := int(font.Header.LastCharacter) % 128
		metaCount = int(font.Header.LastCharacter) + (128 - mod)
	}
	//metaTableOffset := readOffset
	log.Printf("metaCount: %d\n", metaCount)

	var meta []CharacterMeta
	if font.Header.Is9700() {
		meta, err = MetaFrom9700(reader, metaCount)
		readOffset += binary.Size(CharacterMeta9700{}) * metaCount
	} else {
		meta, err = MetaFrom5Word(reader, metaCount)
		readOffset += binary.Size(CharacterMeta5Word{}) * metaCount
	}

	if err != nil {
		return nil, fmt.Errorf("Error parsing metadata: %w", err)
	}

	for id, m := range meta {
		//log.Printf("[font] %d: %s\n", id, m)
		char, err := m.Character(reader, int64(readOffset))
		if err != nil {
			return nil, fmt.Errorf("Error reading character data for 0x%02X: %w", id, err)
		}

		char.Value = rune(id)
		font.Characters[rune(id)] = char
	}

	return font, nil
}

func LoadFontFromFile(filename string) (*Font, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return LoadFont(file)
}

/*
	destImg: destination image
	destPt:  destination start point.  this is the baseline on the destination image.
	c:       text fill color
*/
func (f *Font) DrawString(destImg draw.Image, destPt image.Point, cl color.Color, text string) {
	uni := image.NewUniform(cl)
	maxHeight := int(f.Header.DistanceBelow) + int(f.Header.DistanceAbove)
	offset := destPt.X
	//if maxHeight != int(f.Header.PixelHeight) {
	//	fmt.Printf("maxHeight != PixelHeight: %d != %d\n", maxHeight, f.Header.PixelHeight)
	//}

	for _, r := range []rune(text) {
		c, ok := f.Characters[r]
		if !ok {
			continue
		}

		top := destPt.Y - (c.Height() - (maxHeight - c.BlanksLeft)) - int(f.Header.DistanceAbove)
		draw.DrawMask(
			destImg,
			image.Rect(offset, top, destImg.Bounds().Max.X, destImg.Bounds().Max.Y),
			uni,
			image.Pt(0, 0),
			c.Mask(),
			image.Pt(0, 0),
			draw.Over,
		)

		offset += c.CellWidth
	}
}

func (f *Font) Render(cl color.Color, text string) image.Image {
	lines := strings.Split(text, "\n")
	maxWidth := 0
	pxHeight := len(lines) * int(f.Header.PixelHeight)
	//maxHeight := pxHeight + ((len(lines)-1) * int(f.Header.LineSpacing))
	maxHeight := (len(lines)*int(f.Header.LineSpacing))

	log.Println("pxHeight:", pxHeight)
	log.Println("maxHeight:", maxHeight)
	log.Println("LineSpacing:", f.Header.LineSpacing)
	log.Println("len(lines):", len(lines))

	for _, l := range lines {
		w := f.textWidth(l)
		if w > maxWidth {
			maxWidth = w
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, maxWidth, maxHeight))
	for i, l := range lines {
		//baseline := ((i+1)*int(f.Header.DistanceAbove)) + (i*(int(f.Header.DistanceBelow)+int(f.Header.LineSpacing)))
		baseline := int(f.Header.DistanceAbove) + (i*int(f.Header.LineSpacing))
		log.Printf("[%d] baseline: %d\n", i, baseline)
		f.DrawString(img, image.Pt(0, baseline), cl, l)
	}

	return img
}

func (f *Font) textWidth(text string) int {
	l := 0
	for _, r := range []rune(text) {
		if chr, ok := f.Characters[r]; ok {
			l += chr.CellWidth
		}
	}
	return l
}
