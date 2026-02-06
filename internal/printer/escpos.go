package printer

import (
	"bytes"

	"golang.org/x/text/encoding/charmap"
)

// ESC/POS command constants
var (
	CmdInit       = []byte{0x1B, 0x40}
	CmdAlignLeft  = []byte{0x1B, 0x61, 0x00}
	CmdAlignCenter = []byte{0x1B, 0x61, 0x01}
	CmdAlignRight = []byte{0x1B, 0x61, 0x02}
	CmdBoldOn     = []byte{0x1B, 0x45, 0x01}
	CmdBoldOff    = []byte{0x1B, 0x45, 0x00}
	CmdFontNormal = []byte{0x1B, 0x21, 0x00}
	CmdFontDouble = []byte{0x1B, 0x21, 0x30} // double height + double width
	CmdLineFeed   = []byte{0x0A}
	CmdPartialCut = []byte{0x1D, 0x56, 0x01}
	CmdFullCut    = []byte{0x1D, 0x56, 0x00}
	CmdOpenDrawer = []byte{0x1B, 0x70, 0x00, 0x19, 0xFA}
	CmdCodePage858 = []byte{0x1B, 0x74, 0x13}
)

func CmdFeedLines(n byte) []byte {
	return []byte{0x1B, 0x64, n}
}

// ReceiptBuilder provides a fluent API for constructing ESC/POS byte sequences.
type ReceiptBuilder struct {
	buf     bytes.Buffer
	encoder *charmap.Charmap
}

func NewReceiptBuilder() *ReceiptBuilder {
	rb := &ReceiptBuilder{
		encoder: charmap.CodePage858,
	}
	rb.buf.Write(CmdInit)
	rb.buf.Write(CmdCodePage858)
	return rb
}

func (rb *ReceiptBuilder) AlignLeft() *ReceiptBuilder {
	rb.buf.Write(CmdAlignLeft)
	return rb
}

func (rb *ReceiptBuilder) AlignCenter() *ReceiptBuilder {
	rb.buf.Write(CmdAlignCenter)
	return rb
}

func (rb *ReceiptBuilder) AlignRight() *ReceiptBuilder {
	rb.buf.Write(CmdAlignRight)
	return rb
}

func (rb *ReceiptBuilder) Bold() *ReceiptBuilder {
	rb.buf.Write(CmdBoldOn)
	return rb
}

func (rb *ReceiptBuilder) NoBold() *ReceiptBuilder {
	rb.buf.Write(CmdBoldOff)
	return rb
}

func (rb *ReceiptBuilder) FontNormal() *ReceiptBuilder {
	rb.buf.Write(CmdFontNormal)
	return rb
}

func (rb *ReceiptBuilder) FontDouble() *ReceiptBuilder {
	rb.buf.Write(CmdFontDouble)
	return rb
}

func (rb *ReceiptBuilder) Text(s string) *ReceiptBuilder {
	encoded, err := rb.encoder.NewEncoder().Bytes([]byte(s))
	if err != nil {
		// Replace unmappable characters with '?'
		var safe []byte
		for _, r := range s {
			b, err := rb.encoder.NewEncoder().Bytes([]byte(string(r)))
			if err != nil {
				safe = append(safe, '?')
			} else {
				safe = append(safe, b...)
			}
		}
		rb.buf.Write(safe)
		return rb
	}
	rb.buf.Write(encoded)
	return rb
}

func (rb *ReceiptBuilder) Line(s string) *ReceiptBuilder {
	rb.Text(s)
	rb.buf.Write(CmdLineFeed)
	return rb
}

func (rb *ReceiptBuilder) Feed(n int) *ReceiptBuilder {
	if n > 0 && n <= 255 {
		rb.buf.Write(CmdFeedLines(byte(n)))
	}
	return rb
}

func (rb *ReceiptBuilder) Separator(char byte, width int) *ReceiptBuilder {
	line := bytes.Repeat([]byte{char}, width)
	rb.buf.Write(line)
	rb.buf.Write(CmdLineFeed)
	return rb
}

func (rb *ReceiptBuilder) Cut() *ReceiptBuilder {
	rb.buf.Write(CmdFullCut)
	return rb
}

func (rb *ReceiptBuilder) PartialCut() *ReceiptBuilder {
	rb.buf.Write(CmdPartialCut)
	return rb
}

func (rb *ReceiptBuilder) OpenDrawer() *ReceiptBuilder {
	rb.buf.Write(CmdOpenDrawer)
	return rb
}

func (rb *ReceiptBuilder) Build() []byte {
	return rb.buf.Bytes()
}

func (rb *ReceiptBuilder) Reset() *ReceiptBuilder {
	rb.buf.Reset()
	rb.buf.Write(CmdInit)
	rb.buf.Write(CmdCodePage858)
	return rb
}
