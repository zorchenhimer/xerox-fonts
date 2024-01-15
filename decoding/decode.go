package main

import (
	"os"
	"fmt"
	"encoding/binary"
	"path/filepath"
	"image"
	"image/draw"
	"image/color"
	"image/png"

	"github.com/alexflint/go-arg"
)

type FontFormat byte

const (
	FF_5Word_Portrait   FontFormat = 0xA8
	FF_5Word_Landscape  FontFormat = 0xD0
	FF_5Word_Landscape2 FontFormat = 0x2F
	FF_5Word_IPortrait  FontFormat = 0x58
	FF_5Word_ILandscape FontFormat = 0xF8

	FF_5Word_Unknown FontFormat = 0xE6

	FF_9700_Portrait   FontFormat = 0x20
	FF_9700_Portrait2  FontFormat = 0x98
	FF_9700_Landscape  FontFormat = 0x48
	FF_9700_IPortrait  FontFormat = 0x80
	FF_9700_ILandscape FontFormat = 0x70
)

func (f FontFormat) String() string {
	switch f {
	case FF_5Word_Portrait:
		return "5Word Portrait"
	case FF_5Word_Landscape:
		return "5Word Landscape"
	case FF_5Word_IPortrait:
		return "5Word Inverted"
	case FF_5Word_ILandscape:
		return "5Word Inverted Landscape"

	case FF_9700_Portrait, FF_9700_Portrait2:
		return "9700 Portrait"
	case FF_9700_Landscape:
		return "9700 Landscape"
	case FF_9700_IPortrait:
		return "9700 Inverted Portrait"
	case FF_9700_ILandscape:
		return "9700 Inverted Landscape"
	}

	return "Unknown"
}

type Orientation byte

const (
	Portrait          Orientation = 0x50
	Landscape         Orientation = 0x4C
	InvertedPortrait  Orientation = 0x49
	InvertedLandscape Orientation = 0x4A
)

func (o Orientation) String() string {
	switch o {
	case Portrait:          return "Portrait"
	case Landscape:         return "Landscape"
	case InvertedPortrait:  return "Inverted Portrait"
	case InvertedLandscape: return "Inverted Landscape"
	}

	return "Unknown"
}

func IsOrientation(val byte) bool {
	switch Orientation(val) {
	case Portrait, Landscape, InvertedPortrait, InvertedLandscape:
		return true
	default:
		return false
	}
}

type CharacterMeta interface {
	Is5Word() bool
	IsSpace() bool
	Offset(start int64) int
}

type Arguments struct {
	Input string `arg:"positional,required"`
	OutputDir string `arg:"--dir"`
	IgnoreSpaces bool `arg:"--ignore-spaces"`
}

func main() {
	args := &Arguments{}
	arg.MustParse(args)

	err := run(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args *Arguments) error {
	outputPrefix := filepath.Base(args.Input)
	outputPrefix = outputPrefix[0:len(outputPrefix)-len(filepath.Ext(outputPrefix))]
	//fmt.Fprintf(os.Stderr, "outputPrefix: %s\n", outputPrefix)
	if args.OutputDir != "" {
		outputPrefix = filepath.Join(args.OutputDir, outputPrefix)
		err := os.MkdirAll(args.OutputDir, 0755)
		if err != nil {
			return fmt.Errorf("unable to create output directory: %w", err)
		}
	}

	outlog, err := os.Create(outputPrefix+"_output.txt")
	if err != nil {
		return fmt.Errorf("unable to create output log: %w", err)
	}

	file, err := os.Open(args.Input)
	if err != nil {
		return fmt.Errorf("Unable to open %s: %w", args.Input, err)
	}
	defer file.Close()

	var val byte
	err = binary.Read(file, binary.LittleEndian, &val)
	if err != nil {
		return fmt.Errorf("Error reading first byte: %w", err)
	}

	// Reset to beginning of file
	file.Seek(0, 0)

	readOffset := 0

	var exHeader *ExtraHeader
	if !IsOrientation(val) {
		exHeader = &ExtraHeader{}
		err = binary.Read(file, binary.LittleEndian, exHeader)
		if err != nil {
			return fmt.Errorf("Unable to read extra header: %w", err)
		}

		readOffset += binary.Size(exHeader)
	}

	header := &FontHeader{}
	err = binary.Read(file, binary.LittleEndian, header)
	if err != nil {
		return fmt.Errorf("Unable to read header: %w", err)
	}
	readOffset += binary.Size(header)
	if exHeader != nil {
		fmt.Fprintln(outlog, "exHeader:", binary.Size(exHeader))
	}
	fmt.Fprintln(outlog, "header:", binary.Size(header))

	fmt.Fprintln(outlog, "")

	if exHeader != nil {
		fmt.Fprintln(outlog, "= Extra header")
		fmt.Fprintln(outlog, exHeader)
	}
	fmt.Fprintln(outlog, "= Standard header")
	fmt.Fprintln(outlog, header)

	widthTable := [256]byte{}
	err = binary.Read(file, binary.LittleEndian, &widthTable)
	if err != nil {
		return fmt.Errorf("Unable to read width table: %w", err)
	}
	readOffset += binary.Size(widthTable)

	//err = writeWidths(outputPrefix+"_widths.txt", widthTable[:])
	//if err != nil {
	//	return fmt.Errorf("Unable to write width table: %w", err)
	//}

	var meta []CharacterMeta

	is9700 := true
	if exHeader != nil {
		is9700 = exHeader.Is9700()
	}

	var metaCount int = int(header.LastCharacter)
	if header.LastCharacter < 128 {
		metaCount = 128
	} else if header.LastCharacter % 128 != 0 {
		mod := int(header.LastCharacter) % 128
		metaCount = int(header.LastCharacter) + (128 - mod)
	}

	//metaCount := int(header.LastCharacter % 128) + int(header.LastCharacter)
	fmt.Println("metaCount:", metaCount)

	metaTableOffset := readOffset
	metaSize := binary.Size(CharacterMeta9700{})
	if is9700 {
		//meta, err = parse9700Meta(file, int(header.LastCharacter))
		//meta, err = parse9700Meta(file, lastWidth)
		meta, err = parse9700Meta(file, metaCount)
		fmt.Fprintln(outlog, "type: 9700")
	} else {
		//meta, err = parse5WordMeta(file, int(header.LastCharacter))
		//meta, err = parse5WordMeta(file, lastWidth)
		meta, err = parse5WordMeta(file, metaCount)
		fmt.Fprintln(outlog, "type: 5Word")
		metaSize = binary.Size(CharacterMeta5Word{})
	}

	fmt.Fprintln(outlog, "meta len:", len(meta))
	fmt.Fprintln(outlog, "")

	//eot, err := file.Seek(0, 1)
	//fmt.Printf("end of table: $%04X\n", eot)
	//fmt.Printf("readOffset = readOffset + (metaSize * lastWidth): $%04X = $%04X + (%d * %d)\n",
	//	readOffset + (metaSize*lastWidth), readOffset, metaSize, lastWidth)
	glyphStarts := make(map[int]int)

	//readOffset += metaSize * lastWidth
	readOffset += metaSize * metaCount

	//for i := 0; i < len(meta); i++ {
	for i := 0; i < int(header.LastCharacter); i++ {
		if args.IgnoreSpaces && meta[i].IsSpace() {
			continue
		}

		fmt.Fprintf(outlog, "[$%04X] $%02x %q\n", metaTableOffset+(i*metaSize), i, string(i))
		fmt.Fprint(outlog, meta[i])

		if !meta[i].IsSpace() {
			glyphStarts[i] = meta[i].Offset(int64(readOffset))
			fmt.Fprintf(outlog, "glyph addr:  $%04X\n", glyphStarts[i])
		}
		fmt.Fprintln(outlog, "")
	}

	fmt.Fprintf(os.Stderr, "first glyph at offset $%04X\n", readOffset)

	if !is9700 {
		return nil
	}

	err = os.MkdirAll(outputPrefix+"_chars", 0755)
	if err != nil {
		return fmt.Errorf("unable to create output char directory: %w", err)
	}

	chars := make(map[rune]*Character)

	for id, m := range meta {
		//fmt.Println("extracting", id)
		char := m.(*CharacterMeta9700)
		character, err := From9700(char, file, int64(readOffset))
		if err != nil {
			return fmt.Errorf("error reading character data: %w", err)
		}

		character.Value = rune(id)
		chars[rune(id)] = character

		if m.IsSpace() {
			continue
		}

		filename := filepath.Join(outputPrefix+"_chars", fmt.Sprintf("%03d_0x%02X.png", id, id))
		err = character.WriteImage(filename)
		if err != nil {
			return fmt.Errorf("error writing glyph bitmap: %w", err)
		}
	}

	err = drawText(outputPrefix+"_sample.png", chars, header, "The quick brown fox jumps over the lazy dog")
	if err != nil {
		return fmt.Errorf("Error drawing text: %w", err)
	}

	return nil
}

func fromPng(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return png.Decode(file)
}

func drawText(filename string, chars map[rune]*Character, header *FontHeader, text string) error {
	fmt.Println("drawing sample to", filename)
	line := []*Character{}

	for _, r := range []rune(text) {
		c, ok := chars[r]
		if !ok {
			return fmt.Errorf("Character for rune %c (0x%02X) doesn't exist", r, r)
		}

		line = append(line, c)
	}

	//gr, err := fromPng("gradient.png")
	//if err != nil {
	//	return err
	//}

	img := image.NewRGBA(image.Rect(0, 0, 1200, 200))
	//offset := image.Rect(0, 100, 1200, 200)
	baseline := 100
	offset := 100

	red := color.RGBA{255, 0, 0, 255}
	//blue := color.RGBA{0, 0, 255, 255}
	for i := 0; i < img.Bounds().Max.X; i++ {
		img.Set(i, baseline, red)
	}

	maxHeight := int(header.DistanceBelow) + int(header.DistanceAbove)

	for _, c := range line {
		top := baseline - (c.Width() - (maxHeight - c.BlanksLeft)) - int(header.DistanceAbove)

		draw.Draw(img, image.Rect(offset, top, img.Bounds().Max.X, img.Bounds().Max.Y), c.Image(), image.Pt(0, 0), draw.Over)
		//draw.DrawMask(img, image.Rect(offset, top, img.Bounds().Max.X, img.Bounds().Max.Y), gr, image.Pt(200+offset, 0), c.Image(), image.Pt(0, 0), draw.Over)

		offset += c.CellWidth
	}

	outfile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Unable to create %s: %w", filename, err)
	}
	defer outfile.Close()

	err = png.Encode(outfile, img)
	if err != nil {
		return fmt.Errorf("PNG encode error: %w", err)
	}
	return nil
}

func writeWidths(filename string, table []byte) error {
	outfile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Unable to create width table %s: %w", filename, err)
	}
	defer outfile.Close()

	for i, val := range table {
		fmt.Fprintf(outfile, "%03d [$%02X] %d\n", i, i, val)
	}
	return nil
}
