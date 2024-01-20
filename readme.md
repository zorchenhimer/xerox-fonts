# Xerox Legacy Font (.FNT)

The Xerox font format is really two separate formats: 9700 and 5Word.  The only
difference between these two formats is the data size in the character metadata
table.  The 9700 format's table is 8 bytes per character, while the 5Word is 10
bytes per character.

| Section | Size (bytes) |
| ----------- | ------ |
| Extra Header (optional)   | 128 |
| Main Header               | 256 |
| Width Table               | 256 |
| Character Metadata        | Multiple of 128 |
| Character Glyph Bitmaps   | Variable        |

The width table is a `uint8` table of character widths.  Most likely for
kerning purposes.  This table can be ignored as each character has a CellWidth
filed in the metadata table that can be used for spacing characters.

The character metadata table size is the last character value rounded up to the
nearest 128.  The glyph bitmaps start immediately after this table.

The bitmap table only contains bitmaps for characters that do not have the
spacing property set.  Spacing characters are skipped over.

## Extra Header

This is an optional header padded to 128 bytes.  If present, it will start at
`0x00` bumping the main header to `0x80`.  It can be completely ignored.

| Type       | Description  |
| ---------- | ------------ |
| `[2]byte`  | FontFormat   |
| `[16]byte` | Unknown      |
| `[6]byte`  | Font Name A  |
| `[6]byte`  | Font Name B  |
| `[4]byte`  | Unknown      |
| `[12]byte` | Unknown      |
| `[81]byte` | Padding      |
| `byte`     | End of header (always `0x2A`) |

### Extra Header - Font Formats

This is most likely information about the type of metadata table (9700 or
5Word), orientation, and fixed/proportional properties.  I haven't been able to
come to any solid conclusions about the values yet.

## Main Header

This header will be at `0x00` or `0x80` if the Extra header is present.  This
header is padded to 256 bytes.

| Type      | Description |
| --------- | -------- |
| `byte`      | Orientation (see table) |
| `byte`      | Font Type (see table)   |
| `uint16`    | Pixel height            |
| `uint16`    | Line spacing            |
| `uint16`    | Fixed width (ignored if proportional font) |
| `uint16`    | Distance below   |
| `uint16`    | Distance above   |
| `uint16`    | Distance leading |
| `uint16`    | Unknown          |
| `uint16`    | Last character   |
| `uint16`    | BitmapSize       |
| `[2]byte`   | Unknown          |
| `uint16`    | Unknown5Word     |
| `[6]byte`   | Font name        |
| `[2]byte`   | Revision         |
| `[2]byte`   | Unknown          |
| `[2]byte`   | Version          |
| `[10]byte`  | Library          |
| `[210]byte` | Padding          |

### Main Header - Orientations

| Value    | ASCII | Description |
| -------- | ----- | -------- |
| `0x50`   | P     | Portrait           |
| `0x4C`   | L     | Landscape          |
| `0x49`   | I     | Inverted Portrait  |
| `0x4A`   | J     | Inverted Landscape |

### Main Header - Fonts Types

| Value    | ASCII | Description  |
| -------- | ----- | --------     |
| `0x50`   | P     | Proportional |
| `0x46`   | F     | Fixed        |

## Character Metadata

### 9700

| Type      | Description |
| --------- | -------- |
| uint16 | Blanks left & Spacing |
| uint16 | Glyph offset |
| int16  | Bitmap Size |
| uint16 | Cell width |

### 5Word

| Type      | Description |
| --------- | -------- |
| uint16 | Blanks left & Spacing |
| uint16 | Glyph offset |
| uint16 | Unknown |
| int16  | Bitmap Size |
| uint16 | Cell width |

Bitmap size is a packed field that contains both width and height.  Width is
`abs(BitmapSize) & 0x1FF)` and height is `abs(BitmapSize >> 9) * 8`.  Dimensions
are in pixels.

Blanks left is `BlanksLeft & 0x7FFF` and Spacing is a boolean determited by
`BlanksLeft & 0x8000`.  The baseline is `BlanksLeft & 0x7FFF` pixels from the
bottom of the glyph.

Cell width is the offset from the left side of the bitmap to the current glyph
to the left side of the next glyph.

No idea what the `Unknown` field in the 5Word format is.  The only value I've
seen there is `0xC000`.

## Glyph Bitmaps

The size, in bytes, of each glyph is `abs(BitmapSize >> 9) *
(abs(BitmapSize) & 0x1FF)`.  Each glyph is stored rotated 90 degrees clockwise.
