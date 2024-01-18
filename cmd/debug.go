package main

import (
	"os"
	"fmt"
	"image/color"
	"image/png"
	"path/filepath"
	"encoding/json"

	"github.com/alexflint/go-arg"
	xf "github.com/zorchenhimer/xeroxfont"
)

type Arguments struct {
	InputFile string `arg:"positional,required" help:"Xerox .FNT file"`

	//OutputPrefix string `arg:"-p,--prefix"   help:"Prefix string for all output files."`
	MetadataFile string `arg:"-m,--metadata" help:"File to write font metadata to."`
	ImageDir     string `arg:"-d,--image-dir"      help:"Write font glyphs to this directory."`

	SampleTextFile string `arg:"--sample-text-file" help:"File that contains sample text."`
	SampleText     string `arg:"--sample-text"      help:"Sample text string."`
	SampleOutput   string `arg:"--sample-output"    help:"Output filename for sample."`
}

const (
	DefaultSampleText string = "The quick brown fox jumps over the lazy dog."
)

func run(args *Arguments) error {
	font, err := xf.LoadFontFromFile(args.InputFile)
	if err != nil {
		return fmt.Errorf("Unable to load font: %w", err)
	}

	if len(font.Characters) == 0 {
		return fmt.Errorf("No characters loaded!")
	}

	if args.ImageDir != "" {
		err = os.MkdirAll(args.ImageDir, 0775)
		if err != nil {
			return fmt.Errorf("MkdirAll error: %w", err)
		}

		for id, chr := range font.Characters {
			if chr.IsSpace {
				continue
			}

			filename := filepath.Join(args.ImageDir, fmt.Sprintf("%03d_0x%02X.png", id, id))
			err = chr.WriteImage(filename)
			if err != nil {
				return fmt.Errorf("Error writing glyph bitmap: %w", err)
			}
		}
	}

	if args.MetadataFile != "" {
		ext := filepath.Ext(args.MetadataFile)
		switch ext {
		case ".txt":
			return fmt.Errorf("// TODO: .txt metadata")

		case ".json":
			raw, err := json.MarshalIndent(font, "", "    ")
			if err != nil {
				return fmt.Errorf("JSON marshal error: %w", err)
			}

			err = os.WriteFile(args.MetadataFile, raw, 0664)
			if err != nil {
				return fmt.Errorf("Error writing metadata to %s: %w", args.MetadataFile, err)
			}

		default:
			return fmt.Errorf("Unknown metadata file format: %s", ext)
		}
	}

	if args.SampleOutput != "" {
		if args.SampleTextFile != "" && args.SampleText != "" {
			return fmt.Errorf("--sample-text-file and --sample-text cannot be used at the same time")
		}

		var sampleText string
		if args.SampleTextFile != "" {
			raw, err := os.ReadFile(args.SampleTextFile)
			if err != nil {
				return fmt.Errorf("Unable to read sample text from %s: %w", args.SampleTextFile, err)
			}
			sampleText = string(raw)

		} else if args.SampleText != "" {
			sampleText = args.SampleText

		} else {
			sampleText = DefaultSampleText
		}

		err = drawText(args.SampleOutput, font, sampleText)
		if err != nil {
			return fmt.Errorf("Unable to write sample text: %w", err)
		}
	}

	return nil
}

func drawText(filename string, font *xf.Font, text string) error {
	img := font.Render(color.Black, text)

	outfile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("os.Create() error: %w", err)
	}
	defer outfile.Close()

	err = png.Encode(outfile, img)
	if err != nil {
		return fmt.Errorf("PNG encode error: %w", err)
	}
	return nil
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
