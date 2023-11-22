package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sort"

	"github.com/alexflint/go-arg"
	_ "github.com/llgcode/draw2d/draw2dimg"
)

type SortedChars []*Character

func (s SortedChars) Len() int { return len(s) }
func (s SortedChars) Less(i, j int) bool { return s[i].Value < s[j].Value }
func (s SortedChars) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type Character struct {
	Value rune
	Name string
	Width int
	Height int
	BlanksLeft int
	BaselineOffset int
	BoundX int
	BoundY int
	CellWidth int

	Bitmap []byte
}

func (c Character) String() string {
	return fmt.Sprintf("%-15s Width:%03d Height:%03d BlanksLeft:%03d BaselineOffset:%03d BoundX:%03d BoundY:%03d CellWidth:%03d",
		c.Name,
		c.Width,
		c.Height,
		c.BlanksLeft,
		c.BaselineOffset,
		c.BoundX,
		c.BoundY,
		c.CellWidth,
	)
}

func (c Character) DebugString() string {
	str := []string{}
	for i, b := range c.Bitmap {
		if i % (c.Width/8) == 0 {
			str = append(str, "\n")
		}
		str = append(str, fmt.Sprintf("%08b", b))
	}

	s := strings.Join(str, "")
	s = strings.ReplaceAll(s, "0", " ")
	s = strings.ReplaceAll(s, "1", "X")

	return fmt.Sprintf("%s %dx%d\n%s",
		c.Name,
		c.Width,
		c.Height,
		s,
	)
}

func (c Character) Img(bounds image.Rectangle) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	x, y := 0, 0
	on := color.White
	off := color.Black

	for _, b := range c.Bitmap {
		for i := 7; i >= 0; i-- {
			v := (b >> i) & 0x01
			if v == 1 {
				img.Set(x, y, on)
			} else {
				img.Set(x, y, off)
			}
			x++

			if x >= c.Width {
				y++
				x = 0
			}
		}
	}

	char := image.NewRGBA(bounds)
	draw.Draw(char, char.Bounds(), image.NewUniform(color.RGBA{0xF9, 0x8F, 0xFF, 255}), image.Point{0, 0}, draw.Over)
	draw.Draw(char, image.Rect(c.BlanksLeft, 0, bounds.Max.X, bounds.Max.Y), img, image.Point{0, 0}, draw.Over)

	lines := image.NewRGBA(bounds)
	red := color.RGBA{255, 0, 0, 128}
	//for i := 0; i < c.Height; i++ {
	for i := 0; i < bounds.Max.Y; i++ {
		lines.Set(c.BlanksLeft, i, red)
	}

	blue := color.RGBA{0, 0, 255, 128}
	for i := 0; i < bounds.Max.X; i++ {
		lines.Set(i, c.Height+c.BaselineOffset, blue)
	}

	green := color.RGBA{0, 255, 0, 128}
	for i := 0; i < c.Height; i++ {
		lines.Set(c.Width, i, green)
	}

	draw.Draw(char, image.Rect(0, 0, bounds.Max.X, bounds.Max.Y), lines, image.Point{0, 0}, draw.Over)

	//gc := draw2dimg.NewGraphicContext(char)

	//gc.SetStrokeColor(color.RGBA{255, 0, 0, 0})
	//gc.SetFillColor(color.RGBA{255, 0, 0, 128})
	//gc.SetLineWidth(1)
	//gc.BeginPath()
	//gc.MoveTo(float64(c.BlanksLeft), 0)
	//gc.LineTo(float64(c.BlanksLeft), float64(c.Height))
	//gc.Close()
	//gc.FillStroke()

	//gc.SetStrokeColor(color.RGBA{0, 0, 255, 0})
	//gc.SetFillColor(color.RGBA{0, 0, 255, 128})
	//gc.SetLineWidth(1)
	//gc.BeginPath()
	//gc.MoveTo(0, float64(c.Height + c.BaselineOffset))
	//gc.LineTo(float64(c.Width), float64(c.Height + c.BaselineOffset))
	//gc.Close()
	//gc.FillStroke()

	//gc.SetStrokeColor(color.RGBA{0, 255, 0, 0})
	//gc.SetFillColor(color.RGBA{0, 255, 0, 128})
	//gc.SetLineWidth(1)
	//gc.BeginPath()
	//gc.MoveTo(float64(c.Width), 0)
	//gc.LineTo(float64(c.Width), float64(c.Height))
	//gc.Close()
	//gc.FillStroke()

	return char
}

func (c Character) WriteImage(bounds image.Rectangle, filename string) error {
	outfile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("unable to create %s: %w", filename, err)
	}
	defer outfile.Close()

	err = png.Encode(outfile, c.Img(bounds))
	if err != nil {
		return fmt.Errorf("error encoding %s: %w", filename, err)
	}

	return nil
}

type Font struct {
	Name string
	Bounds image.Rectangle
	GlyphCount int
	Characters map[int]*Character
}

func (f Font) String() string {
	return fmt.Sprintf("Font:%s GlyphCount:%d Bounds:%v", f.Name, f.GlyphCount, f.Bounds)
}

func NewFont(name string) *Font {
	return &Font{
		Name: name,
		Characters: make(map[int]*Character),
	}
}

type Arguments struct {
	Input []string `arg:"positional,required" help:"input PostScript file"`
	//Output string?
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
	for _, file := range args.Input {
		err := processFile(file)
		if err != nil {
			return fmt.Errorf("Error processing %s: %w", file, err)
		}
	}
	return nil
}

func processFile(filename string) error {
	infile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Unable to open %s: %w", filename, err)
	}
	defer infile.Close()

	reader := bufio.NewReader(infile)

	fonts := make(map[string]*Font)
	var currentFont *Font

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("Unable to read line: %w", err)
		}
		line = strings.ReplaceAll(strings.ReplaceAll(line, "\n", ""), "\r", "")

		if strings.HasPrefix(line, "%%BeginResource: font ") {
			currentFont = NewFont(strings.Replace(line, "%%BeginResource: font ", "", 1))
			fonts[currentFont.Name] = currentFont
			//fmt.Println(line)

		} else if strings.HasPrefix(line, "/FontBBox") && currentFont != nil {
			start := strings.Index(line, "[")
			end := strings.Index(line, "]")
			data := strings.Split(line[start+1:end], " ")
			//fmt.Printf("FontBBox: %s\n", data)

			intData := []int{}
			for i := 0; i < 4; i++ {
				i, err := strconv.Atoi(data[i])
				if err != nil {
					return fmt.Errorf("Non-numeric in /FontBBox %q: %s", data[i], line)
				}
				intData = append(intData, i)
			}

			currentFont.Bounds = image.Rect(intData[0], intData[1], intData[2], intData[3])

		} else if strings.HasPrefix(line, "/Encoding") && currentFont != nil {
			currentFont.GlyphCount, err = strconv.Atoi(strings.Split(line, " ")[1])
			if err != nil {
				return fmt.Errorf("Non-numeric /Encoding array size: %s", line)
			}

			for i := 0; i <= currentFont.GlyphCount; i++ {
				line, err = reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						return fmt.Errorf("Premature EOF in font %s", currentFont.Name)
					}
					return fmt.Errorf("Error reading line in /Encoding array: %w", err)
				}
				line = strings.ReplaceAll(strings.ReplaceAll(line, "\n", ""), "\r", "")

				if i == 0 {
					continue
				}

				data := strings.Split(line, " ")
				id, err := strconv.Atoi(data[1])
				if err != nil {
					return fmt.Errorf("Non-numeric glyph id %q in font %s: %s", data[1], currentFont.Name, line)
				}

				currentFont.Characters[id] = &Character{
					Value: rune(id),
					Name: data[2][1:len(data[2])],
				}
			}

		} else if strings.HasPrefix(line, "/CharProcs") && currentFont != nil {
			if currentFont.GlyphCount == 0 {
				return fmt.Errorf("/CharProcs before /Encoding in font %s", currentFont.Name)
			}

			for i := 0; i < currentFont.GlyphCount; i++ {
				//fmt.Printf(">> begin %d/%d\n", i, currentFont.GlyphCount-1)
				// opening slash
				_, _ = reader.ReadString('/')
				name, _ := reader.ReadString('{')
				//fmt.Printf("name: %q", name)
				currentFont.Characters[i].Name = name[0:len(name)-1]

				vals, _ := reader.ReadString('{')
				valsInt := strings.Split(vals[0:len(vals)-1], " ")
				//fmt.Printf("%s %s\n", currentFont.Characters[i].Name, valsInt)
				currentFont.Characters[i].Width, _ = strconv.Atoi(valsInt[0])
				currentFont.Characters[i].Height, _ = strconv.Atoi(valsInt[1])
				currentFont.Characters[i].BlanksLeft, _ = strconv.Atoi(valsInt[2])
				currentFont.Characters[i].BaselineOffset, _ = strconv.Atoi(valsInt[3])
				currentFont.Characters[i].BoundX, _ = strconv.Atoi(valsInt[4])
				currentFont.Characters[i].BoundY, _ = strconv.Atoi(valsInt[5])
				currentFont.Characters[i].CellWidth, _ = strconv.Atoi(valsInt[6])

				data, _ := reader.ReadString('}')
				if len(data) > 1 {
					data = strings.ReplaceAll(data[1:len(data)-2], "\n", "")
					data = strings.ReplaceAll(data, "\r", "")
					//data = strings.ReplaceAll(data, "<", "")
					//data = strings.ReplaceAll(data, ">", "")

					currentFont.Characters[i].Bitmap, err = hex.DecodeString(data)
					if err != nil {
						return fmt.Errorf("Error parsing bin data for %s: %w", currentFont.Characters[i].Name, err)
					}
				}

				peeked, _ := reader.Peek(20)
				if bytes.Contains(peeked, []byte("end def")) {
					break
				}
			}
		} else if strings.HasPrefix(line, "%%EndResource") {
			currentFont = nil
		}
	}

	for _, font := range fonts {
		for id, char := range font.Characters {
			if len(char.Bitmap) == 0 {
				delete(font.Characters, id)
			}
		}
		font.GlyphCount = len(font.Characters)
	}

	//fmt.Println("Fonts found:")
	for _, fnt := range fonts {
		fmt.Println(fnt)
	}
	//fmt.Println("count:", len(fonts))

	//fmt.Println(fonts["AR11NP"].Characters[int('A')].String())
	//for _, c := range fonts["PIPT9B"].Characters {
	//	fmt.Println(c.String())
	//}

	for name, fnt := range fonts {
		os.MkdirAll(name, 0755)
		chars := []*Character{}
		for _, char := range fnt.Characters {
			chars = append(chars, char)
		}

		sort.Sort(SortedChars(chars))

		for _, char := range chars { //fnt.Characters {
			//char := fnt.Characters[id]
			fmt.Println(char.String())
			err = char.WriteImage(fnt.Bounds, filepath.Join(name, fmt.Sprintf("%03d_%s.png", char.Value, char.Name)))
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	return nil
}
