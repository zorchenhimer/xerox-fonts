package main

import (
	"os"
	"fmt"
	//"path/filepath"

	"github.com/alexflint/go-arg"
	xf "github.com/zorchenhimer/xeroxfont"
)

type Arguments struct {
	Input string `arg:"positional,required"`
	Output string `arg:"positional"`
}

func run(args *Arguments) error {
	font, err := xf.LoadFontFromFile(args.Input)
	if err != nil {
		return fmt.Errorf("Unable to load font: %w", err)
	}

	if len(font.Characters) == 0 {
		return fmt.Errorf("No characters loaded!")
	}

	outfile := os.Stdout
	if args.Output != "" {
		file, err := os.Create(args.Output)
		if err != nil {
			return err
		}
		defer file.Close()
		outfile = file
	}

	fmt.Fprintln(outfile, font.BDF(10))
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
