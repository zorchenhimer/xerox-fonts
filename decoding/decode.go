package main

import (
	"os"
	"fmt"
	"io"
	"bytes"
	//"bufio"
	"encoding/binary"
	"strings"
	//"unsafe"
	//"image"
	//"image/color"
	//"image/png"
	"path/filepath"
	"image"
	"image/color"
	"image/png"

	"github.com/alexflint/go-arg"
)

type FontFormat byte

const (
	FF_5Word_Portrait   FontFormat = 0xA8
	FF_5Word_Landscape  FontFormat = 0xD0
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

// 80 byte header, only new fonts saved in Elixir
type ExtraHeader struct {
	FontFormat FontFormat
	FontType byte // fixed or proportional
	UnknownC [16]byte

	FontNameA [6]byte
	FontNameB [6]byte
	UnknownA [4]byte
	UnknownB [12]byte

	_ [81]byte
	End byte
}

func (h ExtraHeader) String() string {
	sb := &strings.Builder{}

	fmt.Fprintf(sb, "FontFormat: $%02X %s\n", byte(h.FontFormat), h.FontFormat)
	fmt.Fprintf(sb, "FontType:   %c   $%02X\n", h.FontType, h.FontType)
	fmt.Fprintf(sb, "UnknownC:   $%X\n", h.UnknownC)
	fmt.Fprintf(sb, "FontNameA:  %q\n", bytes.ReplaceAll(h.FontNameA[:], []byte{0x00}, []byte{0x20}))
	fmt.Fprintf(sb, "FontNameB:  %q\n", bytes.ReplaceAll(h.FontNameB[:], []byte{0x00}, []byte{0x20}))
	fmt.Fprintf(sb, "UnknownA:   $%X\n", h.UnknownA)
	fmt.Fprintf(sb, "UnknownB:   $%X\n", h.UnknownB)
	fmt.Fprintf(sb, "End:        %c   $%X\n", h.End, h.End)

	return sb.String()
}

func (h ExtraHeader) Is9700() bool {
	switch h.FontFormat {
	case FF_5Word_Portrait, FF_5Word_Landscape, FF_5Word_IPortrait, FF_5Word_ILandscape, FF_5Word_Unknown:
		return false
	default:
		return true
	}
}

type CharacterMeta interface {
	Is5Word() bool
	IsSpacing() bool
	Offset(start int64) int
}

type CharacterMeta9700 struct {
	BlanksLeft byte
	Spacing byte // nonblank = $00, spacing = $80, null = ??
	//UnknownB byte
	//UnknownC byte
	GlyphOffset uint16  // halved.  multiply by 2 for byte offset.
	BottomOffset int8 // bitmap bottom offset from top of font bounds
	UnknownE int8 // negated bytes per line?
	//UnknownDE uint16
	CellWidth byte
	UnknownG byte // accent?
}

func (m CharacterMeta9700) Offset(start int64) int {
	return int(start) + (int(m.GlyphOffset) * 2)
}

func (m CharacterMeta9700) String() string {
	spaceStr := "Space"
	if m.Spacing == 0x00 {
		spaceStr = "Non-Blank"
	}

	sb := &strings.Builder{}
	fmt.Fprintf(sb, "BlanksLeft:  $%02X %03d\n", m.BlanksLeft, m.BlanksLeft)
	fmt.Fprintf(sb, "Spacing:     $%02X %03d %s\n", m.Spacing, m.Spacing, spaceStr)
	//fmt.Fprintf(sb, "UnknownB:    $%02X %03d\n", m.UnknownB, m.UnknownB)
	//fmt.Fprintf(sb, "UnknownC:    $%02X %03d\n", m.UnknownC, m.UnknownC)
	fmt.Fprintf(sb, "GlyphOffset: $%04X %04d\n", m.GlyphOffset, m.GlyphOffset)
	//fmt.Fprintf(sb, "UnknownDE:    $%04X %04d\n", m.UnknownDE, m.UnknownDE)
	fmt.Fprintf(sb, "BottomOffset:    $%02X %03d\n", uint8(m.BottomOffset), m.BottomOffset)
	fmt.Fprintf(sb, "UnknownE:    $%02X %03d\n", uint8(m.UnknownE), m.UnknownE)
	fmt.Fprintf(sb, "CellWidth:   $%02X %03d\n", m.CellWidth, m.CellWidth)
	fmt.Fprintf(sb, "UnknownG:    $%02X %03d\n", m.UnknownG, m.UnknownG)
	return sb.String()
}

func (m CharacterMeta9700) Is5Word() bool {
	return false
}

func (m CharacterMeta9700) IsSpacing() bool {
	return m.Spacing == 0x80
}

type CharacterMeta5Word struct {
	Unknown0 byte
	Unknown1 byte
	Unknown2 byte
	Unknown3 byte
	Unknown4 byte
	Unknown5 byte
	Unknown6 byte
	Unknown7 byte
	Unknown8 byte
	Unknown9 byte
}

func (m CharacterMeta5Word) Offset(start int64) int {
	return 0
}

func (m CharacterMeta5Word) String() string {
	sb := &strings.Builder{}

	fmt.Fprintf(sb, "Unknown0:   $%X\n", m.Unknown0)
	fmt.Fprintf(sb, "Unknown1:   $%X\n", m.Unknown1)
	fmt.Fprintf(sb, "Unknown2:   $%X\n", m.Unknown2)
	fmt.Fprintf(sb, "Unknown3:   $%X\n", m.Unknown3)
	fmt.Fprintf(sb, "Unknown4:   $%X\n", m.Unknown4)
	fmt.Fprintf(sb, "Unknown5:   $%X\n", m.Unknown5)
	fmt.Fprintf(sb, "Unknown6:   $%X\n", m.Unknown6)
	fmt.Fprintf(sb, "Unknown7:   $%X\n", m.Unknown7)
	fmt.Fprintf(sb, "Unknown8:   $%X\n", m.Unknown8)
	fmt.Fprintf(sb, "Unknown9:   $%X\n", m.Unknown9)

	return sb.String()
}

func (m CharacterMeta5Word) Is5Word() bool {
	return true
}

func (m CharacterMeta5Word) IsSpacing() bool {
	// TODO: figure out if this is correct (probably isn't)
	return m.Unknown1 == 0x80
}

type FontHeader struct {
	Orientation Orientation
	FontType byte // fixed or proportional
	PixelHeight uint16
	LineSpacing uint16
	FixedWidth uint16
	DistanceBelow uint16
	DistanceAbove uint16
	DistanceLeading uint16
	UnknownD uint16
	LastCharacter uint16

	UnknownC [6]byte
	FontName [6]byte
	Revision [2]byte
	_ [2]byte
	Version [2]byte
	Library [10]byte

	_ [210]byte
}

func (h FontHeader) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "FontName: %q\n", bytes.ReplaceAll(h.FontName[:], []byte{0x00}, []byte{0x20}))
	fmt.Fprintf(sb, "Revision:        %q\n", h.Revision)
	fmt.Fprintf(sb, "Version:         %q\n", h.Version)
	fmt.Fprintf(sb, "Library:         %q\n", h.Library)
	fmt.Fprintf(sb, "Orientation:     %s\n", h.Orientation)
	fmt.Fprintf(sb, "FontType:        %c   $%02X\n", h.FontType, h.FontType)
	fmt.Fprintf(sb, "PixelHeight:     %-3d $%04X\n", h.PixelHeight, h.PixelHeight)
	fmt.Fprintf(sb, "LineSpacing:     %-3d $%04X\n", h.LineSpacing, h.LineSpacing)
	fmt.Fprintf(sb, "FixedWidth:      %-3d $%04X\n", h.FixedWidth, h.FixedWidth)
	fmt.Fprintf(sb, "DistanceBelow:   %-3d $%04X\n", h.DistanceBelow, h.DistanceBelow)
	fmt.Fprintf(sb, "DistanceAbove:   %-3d $%04X\n", h.DistanceAbove, h.DistanceAbove)
	fmt.Fprintf(sb, "DistanceLeading: %-3d $%04X\n", h.DistanceLeading, h.DistanceLeading)
	fmt.Fprintf(sb, "LastCharacter:   %-3d $%04X\n", h.LastCharacter, h.LastCharacter)
	return sb.String()
}

type Arguments struct {
	Input string `arg:"positional,required"`
	DataOffset int `arg:"--data-offset,-d"`
	WidthOffset int `arg:"--width-offset,-w"`
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

	outlog, err := os.Create(outputPrefix+"_output.txt")
	if err != nil {
		return fmt.Errorf("unable to create output log: %w", err)
	}

	file, err := os.Open(args.Input)
	if err != nil {
		return fmt.Errorf("Unable to open %s: %w", args.Input, err)
	}
	defer file.Close()

	//reader := bufio.NewReader(file)
	//val, err := reader.ReadByte()
	//if err != nil {
	//	return fmt.Errorf("Error peeking: %w", err)
	//}

	//err = reader.UnreadByte()
	//if err != nil {
	//	return fmt.Errorf("Error unreading: %w", err)
	//}

	var val byte
	err = binary.Read(file, binary.LittleEndian, &val)
	if err != nil {
		return fmt.Errorf("Error reading first byte: %w", err)
	}

	// Reset to beginning of file
	file.Seek(0, 0)


	var t FontFormat = FontFormat(val)
	readOffset := 0

	var exHeader *ExtraHeader
	switch t {
	case FF_5Word_Portrait, FF_5Word_Landscape, FF_5Word_IPortrait, FF_5Word_ILandscape,
		 FF_9700_Portrait, FF_9700_Portrait2, FF_9700_Landscape, FF_9700_IPortrait,
		 FF_9700_ILandscape, FF_5Word_Unknown:

		//fmt.Println("extra header found")

		exHeader = &ExtraHeader{}
		err = binary.Read(file, binary.LittleEndian, exHeader)
		if err != nil {
			return fmt.Errorf("Unable to read extra header: %w", err)
		}

		readOffset += binary.Size(exHeader)
		//fmt.Printf("Filler: %v\n", exHeader.Filler)
		//fmt.Printf("End: $%02X\n", exHeader.End)
		//fmt.Printf("size: %d\n", binary.Size(exHeader))

	case FontFormat(Portrait), FontFormat(Landscape), FontFormat(InvertedPortrait), FontFormat(InvertedLandscape):
		// normal header
	
	default:
		return fmt.Errorf("Unknown font type: $%02X", val)
	}

	header := &FontHeader{}
	//fmt.Printf("headerSize: %d\n", binary.Size(header))
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

	err = writeWidths(outputPrefix+"_widths.txt", widthTable[:])
	if err != nil {
		return fmt.Errorf("Unable to write width table: %w", err)
	}

	var meta []CharacterMeta
	//_ = meta

	is9700 := true
	if exHeader != nil {
		is9700 = exHeader.Is9700()
	}

	metaTableOffset := readOffset
	metaSize := binary.Size(CharacterMeta9700{})
	if is9700 {
		meta, err = parse9700Meta(file, int(header.LastCharacter))
		fmt.Fprintln(outlog, "type: 9700")
	} else {
		meta, err = parse5WordMeta(file, int(header.LastCharacter))
		fmt.Fprintln(outlog, "type: 5Word")
		metaSize = binary.Size(CharacterMeta5Word{})
	}

	fmt.Fprintln(outlog, "meta len:", len(meta))
	readOffset += metaSize * len(meta)

	eot, err := file.Seek(0, 1)
	fmt.Printf("end of table: $%04X\n", eot)
	glyphStarts := make(map[int]int)

	var firstGlyph CharacterMeta
	firstGlyphId := -1
	for i := 0; i < len(meta); i++ {
		fmt.Fprintf(outlog, "[$%04X] $%02x %q\n", metaTableOffset+(i*metaSize), i, string(i))
		fmt.Fprint(outlog, meta[i])
		if !meta[i].IsSpacing() {
			if firstGlyph == nil {
				firstGlyph = meta[i]
				firstGlyphId = i
			}
			glyphStarts[i] = meta[i].Offset(eot)
			fmt.Fprintf(outlog, "glyph addr:  $%04X\n", glyphStarts[i])
		}
		fmt.Fprintln(outlog, "")
	}

	if firstGlyph != nil {
		fmt.Fprintf(outlog, "first glyph: $%02X %q\n", firstGlyphId, firstGlyphId)
	}

	fmt.Fprintf(os.Stderr, "first glyph at offset $%04X\n", readOffset)

	//discarded := 100
	//reader.Discard(discarded) // skip past '!' for now.
	//readOffset += discarded
	fmt.Fprintf(os.Stderr, "reading glyph at offset $%04X\n", readOffset)

	raw := [256]byte{}
	err = binary.Read(file, binary.LittleEndian, &raw)
	if err != nil {
		return fmt.Errorf("Error reading bitmap data: %w", err)
	}
	readOffset += binary.Size(raw)

	// seek to end of table
	//_, err := file.Seek(eot, 0)
	//if err != nil {
	//	return fmt.Errorf("seek eot error: %w", err)
	//}

	err = os.MkdirAll(outputPrefix+"_chars", 0755)
	if err != nil {
		return fmt.Errorf("unable to create output char directory: %w", err)
	}

	for id, m := range meta {
		if m.IsSpacing() {
			continue
		}
		fmt.Println("extracting", id)
		char := m.(*CharacterMeta9700)
		_, err = file.Seek(int64(char.Offset(eot)), 0)
		if err != nil {
			return fmt.Errorf("unable to seek to glyph start $%04X: %w", char.Offset(eot), err)
		}

		l := int(char.BottomOffset*-1) * int(char.UnknownE*-1)
		buff := make([]byte, l)
		_, err = file.Read(buff)
		if err != nil {
			return fmt.Errorf("error reading glyph bytes: %w", err)
		}

		filename := filepath.Join(outputPrefix+"_chars", fmt.Sprintf("%03d.png", id))
		//if id == 34 {
		//	err = writeImage(filename, buff[:], 16, int(char.BottomOffset*-1))
		//} else {
			err = writeImage(filename, buff[:], int(char.UnknownE*-1)*4, int(char.BottomOffset*-1))
		//}
		if err != nil {
			return fmt.Errorf("error writing glyph image: %w", err)
		}
	}

	//char := meta[33].(*CharacterMeta9700)
	//bytesPerLine := int(char.UnknownE*-1)
	//lines := int(char.BottomOffset*-1)
	//fmt.Println("bytesPerLine:", bytesPerLine)
	//fmt.Println("lines:", lines)
	//fmt.Println("char.BottomOffset:", char.BottomOffset)
	//err = writeImage(outputPrefix+"_char.png", raw[:], bytesPerLine*8, lines)

	//fmt.Println(data)

	//for i := 0; i < 256; i++ {
	//	fmt.Printf(">> $%02X %03d %q\n", i, i, i)
	//	fmt.Println(data.Data[i].String())
	//}

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

func writeImage(filename string, raw []byte, width, height int) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	x, y := 0, 0
	on := color.White
	off := color.Black

	for _, b := range raw {
		for i := 7; i > -1; i-- {
			v := (b >> i) & 0x01
			if v == 1 {
				img.Set(x, y, on)
			} else {
				img.Set(x, y, off)
			}
			x++
			if x >= width {
				y++
				x = 0
				break
			}
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

func parse9700Meta(reader io.Reader, lastChar int) ([]CharacterMeta, error) {
	var err error
	meta := []CharacterMeta{}
	for i := 0; i <= lastChar; i++ {
		m := &CharacterMeta9700{}
		err = binary.Read(reader, binary.LittleEndian, m)
		if err != nil {
			return nil, fmt.Errorf("Error reading CharacterMeta9700 at index %d: %w", i, err)
		}
		meta = append(meta, m)
	}

	return meta, nil
}

func parse5WordMeta(reader io.Reader, lastChar int) ([]CharacterMeta, error) {
	var err error
	meta := []CharacterMeta{}
	for i := 0; i <= lastChar; i++ {
		m := &CharacterMeta5Word{}
		err = binary.Read(reader, binary.LittleEndian, m)
		if err != nil {
			return nil, fmt.Errorf("Error reading CharacterMeta5Word at index %d: %w", i, err)
		}
		meta = append(meta, m)
	}

	return meta, nil
}

//func (data XeroxFont) WriteImage(widthOffset, dataOffset int) error {
//	img := image.NewRGBA(image.Rect(0, 0, int(data.CellWidthTable[widthOffset]), int(data.Header.PixelHeight)))
//	x, y := 0, 0
//	on := color.White
//	off := color.Black
//
//	for _, b := range data.Data[dataOffset] {
//		for i := 7; i >= 0; i-- {
//			v := (b >> i) & 0x01
//			if v == 1 {
//				img.Set(x, y, on)
//			} else {
//				img.Set(x, y, off)
//			}
//			x++
//
//			if x >= int(data.CellWidthTable[widthOffset]) {
//				y++
//				x = 0
//			}
//		}
//	}
//
//	dir := filepath.Join("images", fmt.Sprintf("%03d", widthOffset))
//	filename := filepath.Join(dir, fmt.Sprintf("%03d_%03d.png", widthOffset, dataOffset))
//	outfile, err := os.Create(filename)
//	if err != nil {
//		return fmt.Errorf("unable to create %s: %w", filename, err)
//	}
//	defer outfile.Close()
//
//	err = png.Encode(outfile, img)
//	if err != nil {
//		return fmt.Errorf("unable to encode %s: %w", filename, err)
//	}
//
//	return nil
//}
