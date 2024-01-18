//go:build ignore

package main

import (
	"os"
	"fmt"
	"encoding/binary"
	"path/filepath"
	"image/color"
	"image/png"

	"github.com/alexflint/go-arg"
)

//type CharacterMeta interface {
//	Is5Word() bool
//	IsSpace() bool
//
//	// Given the start address, what is the offset of the glyph.
//	Offset(start int64) int
//}

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

	font := &Font{
		Header: header,
		Characters: chars,
	}

	//err = drawText(outputPrefix+"_sample.png", font, "The quick brown fox jumps over the lazy dog.\nThe quick brown fox jumps over the lazy dog!")
	err = drawText(outputPrefix+"_sample.png", font, testString)
	if err != nil {
		return fmt.Errorf("Error drawing text: %w", err)
	}

	return nil
}

func drawText(filename string, font *Font, text string) error {
	fmt.Println("drawing sample to", filename)
	img := font.Render(color.Black, text)

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

var testString = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam sollicitudin arcu
sed pulvinar tristique. Maecenas consectetur pulvinar tempus. Pellentesque ac
auctor ex. Etiam vel molestie felis. Nullam scelerisque nisi et egestas
bibendum. Proin rutrum, ante sit amet imperdiet volutpat, urna ex convallis
velit, ut porta nisl ex molestie dolor. Aliquam rhoncus pretium felis sed
cursus. Quisque ornare, turpis et dignissim tincidunt, velit mi rutrum mauris,
et condimentum arcu ipsum vitae lacus. Cras a purus ac urna interdum pharetra
interdum et est. Fusce condimentum neque mauris, id elementum erat dapibus sit
amet. Fusce eu diam viverra, efficitur augue et, mollis ex. Phasellus facilisis
enim non mi fermentum, sit amet sagittis lacus ultricies. Sed maximus, justo a
euismod volutpat, est augue venenatis lacus, eget facilisis dui felis fermentum
libero. Aenean sollicitudin consectetur sodales.

Ut eget euismod velit, non pellentesque nunc. Pellentesque volutpat tellus
interdum rhoncus fringilla. Quisque luctus leo vel lectus malesuada pretium.
Mauris semper pretium metus, at eleifend metus mattis in. Duis sit amet nulla
suscipit, cursus nisl eu, laoreet nisl. Nunc quis urna sollicitudin, imperdiet
mauris in, volutpat felis. Vivamus imperdiet massa et risus euismod volutpat.
Ut cursus elit a dolor condimentum, in posuere orci ornare. Donec faucibus eros
at molestie elementum. In porttitor nunc id nisi tempus dapibus. Phasellus a
lorem quam. Class aptent taciti sociosqu ad litora torquent per conubia nostra,
per inceptos himenaeos. Nulla felis nibh, luctus sit amet purus vel, lobortis
vestibulum nulla.

Proin pretium sem ac bibendum venenatis. Aliquam non velit a nibh finibus
facilisis et at quam. Donec ullamcorper cursus metus egestas laoreet. Phasellus
tellus lacus, luctus eget nunc eget, porta pretium tortor. Nulla facilisi. Nam
cursus, justo eu dictum efficitur, justo nisi consequat arcu, nec viverra
sapien libero vitae eros. Pellentesque interdum diam quis ante tempor, in
rutrum metus tempus. Morbi posuere elit vel dolor congue, et posuere purus
mollis. Pellentesque habitant morbi tristique senectus et netus et malesuada
fames ac turpis egestas. Donec pulvinar fermentum mauris.

Curabitur laoreet augue at velit semper maximus. Cras tempor purus lorem, eu
viverra magna mollis sed. Mauris venenatis neque non arcu commodo, non
ultricies odio cursus. Duis ut erat ipsum. Donec dapibus mattis lorem quis
sagittis. Vivamus vel dapibus risus. Donec a dolor pulvinar, ornare eros et,
tempus magna. In vel nibh mollis, suscipit nibh nec, sagittis tellus. Nulla
placerat diam ipsum, eget auctor sapien suscipit et. Vestibulum posuere
pulvinar mauris, eu congue est interdum sed. Sed pretium dui quis ex tempus,
eget gravida magna consequat. Phasellus eros velit, rhoncus consectetur
ultricies et, blandit id mauris. Suspendisse dapibus lorem eu consequat
sodales. Integer non nisl sodales, fermentum mauris ut, cursus ante. Vestibulum
congue orci ut porttitor elementum. Sed nec neque commodo, pellentesque urna
at, tempor ligula.

Suspendisse ac lacus suscipit, iaculis lorem eu, consectetur lectus. Aenean
efficitur maximus leo quis maximus. Nullam aliquet lacus condimentum, accumsan
erat ut, eleifend dui. Phasellus a gravida sapien, sed iaculis quam. In et orci
et nulla eleifend tincidunt. Fusce id tempor lorem, rutrum pellentesque felis.
Ut rutrum lacus a sagittis scelerisque. Ut varius cursus volutpat. Aenean diam
massa, efficitur sit amet vehicula at, mattis pulvinar nunc. Vestibulum turpis
ante, fermentum ac ullamcorper at, pulvinar ac lorem. Duis venenatis sagittis
turpis, et semper orci. Nulla convallis augue dolor, vel efficitur metus
convallis vitae. Maecenas lacinia laoreet vulputate. Donec dignissim elit ut
velit molestie, a blandit quam laoreet. Class aptent taciti sociosqu ad litora
torquent per conubia nostra, per inceptos himenaeos. Quisque ac mollis diam.`
