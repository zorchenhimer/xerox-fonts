package xeroxfont

import (
	"fmt"
	"strings"
	"slices"
	//"image/color"
	//"os"
)

func (f *Font) BDF(ptSize int) string {
	sb := &strings.Builder{}

	name := []string{
		strings.TrimSpace(string(f.Header.FontName[:])),
		strings.TrimSpace(string(f.Header.Revision[:])),
		strings.TrimSpace(string(f.Header.Version[:])),
		strings.TrimSpace(string(f.Header.Library[:])),
	}

	name = slices.DeleteFunc(name, func(s string) bool { return s == "" })

	fmt.Fprintln(sb, "STARTFONT 2.2")
	fmt.Fprintf(sb, "FONT %s\n", strings.Join(name, " "))
	fmt.Fprintf(sb, "SIZE %d 300 300\n", ptSize)
	// TODO: width, height, offset X, offset Y
	//fmt.Fprintf(sb, "FONTBOUNDINGBOX %d %d %d %d\n", f.Header.FixedWidth, int(f.Header.DistanceAbove)+int(f.Header.DistanceBelow), 0, int(f.Header.DistanceBelow)*-1)
	fmt.Fprintf(sb, "FONTBOUNDINGBOX %d %d %d %d\n", f.Header.FixedWidth, f.Header.PixelHeight, 0, int(f.Header.DistanceBelow)*-1)

	props := make(map[string]any)
	props["FONT_ASCENT"] = int(f.Header.DistanceAbove)
	props["FONT_DESCENT"] = int(f.Header.DistanceBelow)

	fmt.Fprintf(sb, "STARTPROPERTIES %d\n", len(props))
	for k, v := range props {
		fmt.Fprintln(sb, k, v)
	}
	fmt.Fprintln(sb, "ENDPROPERTIES")

	l := 0
	for _, c := range f.Characters {
		if !c.IsSpace && c.Value != ' ' {
			l++
		}
	}
	fmt.Fprintf(sb, "CHARS %d\n", l)

	for _, c := range f.Characters {
		c.bdf(sb, f.Header)
	}

	fmt.Fprintln(sb, "ENDFONT")
	return sb.String()
}

func (c *Character) bdf(sb *strings.Builder, h *FontHeader) {
	if c.IsSpace && c.Value != ' ' {
		return
	}

	fmt.Fprintf(sb, "STARTCHAR %s\n", PostscriptNames[c.Value])
	fmt.Fprintf(sb, "ENCODING %d\n", int(c.Value))
	//fmt.Printf("\nCharacter %s (0x%02X)\n", PostscriptNames[c.Value], c.Value)
	// BBX BBw BBh BBxoff0x BByoff0y
	//yoff := c.Height() - ((int(h.DistanceAbove) + int(h.DistanceBelow)) - c.BlanksLeft)
	fmt.Fprintf(sb, "BBX %d %d %d %d\n", c.Width(), c.Height(), 0, c.BlanksLeft-int(h.DistanceBelow))
	//fmt.Fprintf(sb, "BBX %d %d %d %d\n", c.CellWidth, c.Height(), 0, yoff)

	fmt.Fprintln(sb, "SWIDTH 300 0")
	fmt.Fprintf(sb, "DWIDTH %d 0\n", c.CellWidth)

	fmt.Fprintln(sb, "BITMAP")
	w := c.Width()

	if w % 8 != 0 {
		w += 8 - (w % 8)
	}

	img := c.Mask()

	for y := 0; y < c.Height(); y++ {
		line := 0
		// TODO: make this better
		for x := 0; x < w; x++ {
			if x % 8 == 0 && x != 0 {
				fmt.Fprintf(sb, "%02X", line)
				//fmt.Printf(strings.ReplaceAll(fmt.Sprintf("%08b", line), "0", " "))
				line = 0
			}

			cl := img.At(x, y)
			r, _, _, _ := cl.RGBA()
			if r != 0 {
				line = (line << 1) | 1
			} else {
				line <<= 1
			}

		}
		//fmt.Println(strings.ReplaceAll(fmt.Sprintf("%08b", line), "0", " "))
		fmt.Fprintf(sb, "%02X\n", line)
	}


	fmt.Fprintln(sb, "ENDCHAR")
}

var PostscriptNames = map[rune]string {
	0: "U0",
	1: "controlSTX",
	2: "controlSOT",
	3: "controlETX",
	4: "controlEOT",
	5: "controlENQ",
	6: "controlACK",
	7: "controlBEL",
	8: "controlBS",
	9: "controlHT",
	10: "controlLF",
	11: "controlVT",
	12: "controlFF",
	13: "controlCR",
	14: "controlSO",
	15: "controlSI",
	16: "controlDLE",
	17: "controlDC1",
	18: "controlDC2",
	19: "controlDC3",
	20: "controlDC4",
	21: "controlNAK",
	22: "controlSYN",
	23: "controlETB",
	24: "controlCAN",
	25: "controlEM",
	26: "controlSUB",
	27: "controlESC",
	28: "controlFS",
	29: "controlGS",
	30: "controlRS",
	31: "controlUS",

	' ': "space",
	'!': "exclam",
	'"': "quotedbl",
	'#': "numbersign",
	'$': "dollar",
	'%': "percent",
	'&': "ampersand",
	'\'': "quotesingle",
	'(': "parenleft",
	')': "parenright",
	'*': "asterisk",
	'+': "plus",
	',': "comma",
	'-': "hyphen",
	'.': "period",
	'/': "slash",
	'0': "zero",
	'1': "one",
	'2': "two",
	'3': "three",
	'4': "four",
	'5': "five",
	'6': "six",
	'7': "seven",
	'8': "eight",
	'9': "nine",
	':': "colon",
	';': "semicolon",
	'<': "less",
	'=': "equal",
	'>': "greater",
	'?': "question",
	'@': "at",

	'[': "bracketleft",
	'\\': "backslash",
	']': "bracketright",
	'^': "asciicircum",
	'_': "underscore",
	'`': "grave",

	'{': "braceleft",
	'|': "bar",
	'}': "braceright",
	'~': "asciitilde",
	127: "controlDEL",

	'a': "a",
	'b': "b",
	'c': "c",
	'd': "d",
	'e': "e",
	'f': "f",
	'g': "g",
	'h': "h",
	'i': "i",
	'j': "j",
	'k': "k",
	'l': "l",
	'm': "m",
	'n': "n",
	'o': "o",
	'p': "p",
	'q': "q",
	'r': "r",
	's': "s",
	't': "t",
	'u': "u",
	'v': "v",
	'w': "w",
	'x': "x",
	'y': "y",
	'z': "z",

	'A': "A",
	'B': "B",
	'C': "C",
	'D': "D",
	'E': "E",
	'F': "F",
	'G': "G",
	'H': "H",
	'I': "I",
	'J': "J",
	'K': "K",
	'L': "L",
	'M': "M",
	'N': "N",
	'O': "O",
	'P': "P",
	'Q': "Q",
	'R': "R",
	'S': "S",
	'T': "T",
	'U': "U",
	'V': "V",
	'W': "W",
	'X': "X",
	'Y': "Y",
	'Z': "Z",

	128: "Euro",
	129: "U81",
	130: "quotesinglbase",
	131: "florin",
	132: "quotedblbase",
	133: "ellipsis",
	134: "dagger",
	135: "daggerdbl",
	136: "circumflex",
	137: "perthousand",
	138: "Scaron",
	139: "guilsinglleft",
	140: "OE",
	141: "U8D",
	142: "Zcaron",
	143: "U8F",
	144: "U90",
	145: "quoteleft",
	146: "quoteright",
	147: "quotedblleft",
	148: "quotedblright",
	149: "bullet",
	150: "endash",
	151: "emdash",
	152: "tilde",
	153: "trademark",
	154: "scaron",
	155: "guilsinglright",
	156: "oe",
	157: "U9D",
	158: "zcaron",
	159: "Ydieresis",
	160: "nbspace",
	161: "exclamdown",
	162: "cent",
	163: "sterling",
	164: "currency",
	165: "yen",
	166: "brokenbar",
	167: "section",
	168: "dieresis",
	169: "copyright",
	170: "ordfeminine",
	171: "guillemotleft",
	172: "logicalnot",
	173: "sfthyphen",
	174: "registered",
	175: "macron",
	176: "degree",
	177: "plusminus",
	178: "twosuperior",
	179: "threesuperior",
	180: "acute",
	181: "mu",
	182: "paragraph",
	183: "middot",
	184: "cedilla",
	185: "onesuperior",
	186: "ordmasculine",
	187: "guillemotright",
	188: "onequarter",
	189: "onehalf",
	190: "threequarters",
	191: "questiondown",
	192: "Agrave",
	193: "Aacute",
	194: "Acircumflex",
	195: "Atilde",
	196: "Adieresis",
	197: "Aring",
	198: "AE",
	199: "Ccedilla",
	200: "Egrave",
	201: "Eacute",
	202: "Ecircumflex",
	203: "Edieresis",
	204: "Igrave",
	205: "Iacute",
	206: "Icircumflex",
	207: "Idieresis",
	208: "Eth",
	209: "Ntilde",
	210: "Ograve",
	211: "Oacute",
	212: "Ocircumflex",
	213: "Otilde",
	214: "Odieresis",
	215: "multiply",
	216: "Oslash",
	217: "Ugrave",
	218: "Uacute",
	219: "Ucircumflex",
	220: "Udieresis",
	221: "Yacute",
	222: "Thorn",
	223: "germandbls",
	224: "agrave",
	225: "aacute",
	226: "acircumflex",
	227: "atilde",
	228: "adieresis",
	229: "aring",
	230: "ae",
	231: "ccedilla",
	232: "egrave",
	233: "eacute",
	234: "ecircumflex",
	235: "edieresis",
	236: "igrave",
	237: "iacute",
	238: "icircumflex",
	239: "idieresis",
	240: "eth",
	241: "ntilde",
	242: "ograve",
	243: "oacute",
	244: "ocircumflex",
	245: "otilde",
	246: "odieresis",
	247: "divide",
	248: "oslash",
	249: "ugrave",
	250: "uacute",
	251: "ucircumflex",
	252: "udieresis",
	253: "yacute",
	254: "thorn",
	255: "ydieresis",
}
