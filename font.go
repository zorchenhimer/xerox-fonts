package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"
	//"image/png"

)

type Font struct {
	Header *FontHeader
	Characters map[rune]*Character
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

	for _, r := range []rune(text) {
		c, ok := f.Characters[r]
		if !ok {
			continue
		}

		top := destPt.Y - (c.Width() - (maxHeight - c.BlanksLeft)) - int(f.Header.DistanceAbove)
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

	fmt.Println("pxHeight:", pxHeight)
	fmt.Println("maxHeight:", maxHeight)
	fmt.Println("LineSpacing:", f.Header.LineSpacing)
	fmt.Println("len(lines):", len(lines))

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
		fmt.Printf("[%d] baseline: %d\n", i, baseline)
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
