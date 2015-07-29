// Copyright 2013, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Copyright 2013, Meng Zhang. All rights reserved.
// URL: https://github.com/wsxiaoys/terminal
// File URL: https://github.com/wsxiaoys/terminal/blob/decf4e097e2e3471b254da8d30c3599d330fe7ba/color/color.go

package pr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

// Mapping from character to concrete escape code.
var codeMap = map[int]int{
	'|': 0,
	'!': 1,
	'.': 2,
	'/': 3,
	'_': 4,
	'^': 5,
	'&': 6,
	'?': 7,
	'-': 8,

	'k': 30,
	'r': 31,
	'g': 32,
	'y': 33,
	'b': 34,
	'm': 35,
	'c': 36,
	'w': 37,
	'd': 39,

	'K': 40,
	'R': 41,
	'G': 42,
	'Y': 43,
	'B': 44,
	'M': 45,
	'C': 46,
	'W': 47,
	'D': 49,
}

// ErrInvalidEscape is an error that is used when the parser panics.
var ErrInvalidEscape = errors.New("invalid escape rune")
var ErrUnexpectedEOF = errors.New("unexpected EOF while parsing")

// ColorReset is the string that resets the text to default style.
const ColorReset = "\033[0m"

// ColorCode compiles a color syntax string like "rG" to escape code.
func ColorCode(s string) string {
	attr := 0
	fg := 39
	bg := 49

	for _, key := range s {
		c, ok := codeMap[int(key)]
		if !ok {
			panic("wrong color syntax: " + string(key))
		}

		switch {
		case 0 <= c && c <= 8:
			attr = c
		case 30 <= c && c <= 37:
			fg = c
		case 40 <= c && c <= 47:
			bg = c
		}
	}
	return fmt.Sprintf("\033[%d;%d;%dm", attr, fg, bg)
}

// Color translates a string into an escaped string.
//
// This example will output the text with a Blue foreground and a Black background
//      color.Println("@{bK}Example Text")
//
// This one will output the text with a red foreground
//      color.Println("@rExample Text")
//
// This one will escape the @
//      color.Println("@@")
//
// Full color syntax code
//      @{rgbcmykwRGBCMYKW}  foreground/background color
//        r/R:  Red
//        g/G:  Green
//        b/B:  Blue
//        c/C:  Cyan
//        m/M:  Magenta
//        y/Y:  Yellow
//        k/K:  Black
//        w/W:  White
//      @{|}  Reset format style
//      @{!./_} Bold / Dim / Italic / Underline
//      @{^&} Blink / Fast blink
//      @{?} Reverse the foreground and background color
//      @{-} Hide the text
// Note some of the functions are not widely supported, like "Fast blink" and "Italic".
func Color(s string, escape rune) string {
	return newParser(escape, true).translateReset(s)
}

// Decolor cleans a string of @x color codes.
func Decolor(s string, escape rune) string {
	return newParser(escape, false).translateReset(s)
}

// Uncolor cleans a string of ANSI color codes.
func Uncolor(s string) string {
	panic("not implemented")
}

type Colorizer struct {
	w io.Writer
	*parser
}

func NewColorizer() *Colorizer {
	return &Colorizer{
		w:      os.Stdout,
		parser: newParser('@', true),
	}
}

func (c *Colorizer) EscapeChar() rune {
	return c.parser.escape
}

// SetEscapeChar sets the escape character, which can be one of the following characters:
//
//		@ * + = ~
//
// If it is none of these characters, then this function panics with ErrInvalidEscape.
func (c *Colorizer) SetEscapeChar(r rune) {
	if c.EscapeChar() == r {
		return
	}

	for _, q := range []rune{'*', '@', '+', '=', '~'} {
		if r == q {
			c.parser.escape = r
			return
		}
	}

	panic(ErrInvalidEscape)
}

func (c *Colorizer) Enabled() bool {
	return c.parser.color
}

func (c *Colorizer) SetEnabled(b bool) {
	if c.Enabled() == b {
		return
	}
	c.parser.color = b
}

func (c *Colorizer) SetOutput(w io.Writer) {
	c.w = w
}

func (c *Colorizer) SetFile(f *os.File) {
	c.SetEnabled(terminal.IsTerminal(int(f.Fd())))
	c.w = f
}

func (c *Colorizer) Color(s string) string {
	return c.translateReset(s)
}

func (c *Colorizer) colorAny(args []interface{}) []interface{} {
	n := len(args)
	r := make([]interface{}, n, n+1)
	for i, x := range args {
		if str, ok := x.(string); ok {
			x = c.translateOnly(str)
		}
		r[i] = x
	}
	if c.Enabled() {
		r = append(r, ColorReset)
	}
	return r
}

func (c *Colorizer) Print(a ...interface{}) (int, error) {
	return fmt.Fprint(c.w, c.colorAny(a)...)
}
func (c *Colorizer) Println(a ...interface{}) (int, error) {
	return fmt.Fprintln(c.w, c.colorAny(a)...)
}
func (c *Colorizer) Printf(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(c.w, c.translateReset(format), a...)
}
func (c *Colorizer) Fprint(w io.Writer, a ...interface{}) (int, error) {
	return fmt.Fprint(w, c.colorAny(a)...)
}
func (c *Colorizer) Fprintln(w io.Writer, a ...interface{}) (int, error) {
	return fmt.Fprintln(w, c.colorAny(a)...)
}
func (c *Colorizer) Fprintf(w io.Writer, format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(w, c.translateReset(format), a...)
}
func (c *Colorizer) Sprint(a ...interface{}) string {
	return fmt.Sprint(c.colorAny(a)...)
}
func (c *Colorizer) Sprintln(a ...interface{}) string {
	return fmt.Sprintln(c.colorAny(a)...)
}
func (c *Colorizer) Sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(c.translateReset(format), a...)
}

type parser struct {
	escape rune
	color  bool
}

func newParser(escape rune, color bool) *parser {
	return &parser{
		escape: escape,
		color:  color,
	}
}

type handler func(p *parser, in, out *bytes.Buffer) (handler, error)

func (p *parser) translateReset(s string) string {
	in := bytes.NewBufferString(s)
	out := bytes.NewBufferString("")

	var h = handleRegular
	var err error
	for {
		h, err = h(p, in, out)
		if err != nil {
			panic(err)
		}
		if h == nil {
			break
		}
	}
	if p.color {
		out.WriteString(ColorReset)
	}
	return out.String()
}

func (p *parser) translateOnly(s string) string {
	in := bytes.NewBufferString(s)
	out := bytes.NewBufferString("")

	var h = handleRegular
	var err error
	for {
		h, err = h(p, in, out)
		if err != nil {
			panic(err)
		}
		if h == nil {
			break
		}
	}
	return out.String()
}

func handleRegular(p *parser, in, out *bytes.Buffer) (handler, error) {
	for {
		r, _, err := in.ReadRune()
		// The only error that can happen here is that we have reached the end of file,
		// or that a rune is messed up. If the rune is messed up, we treat it normally.
		// This is why we only check for io.EOF.
		if err == io.EOF {
			break
		}

		if r == p.escape {
			return handleEscape, nil
		}
		out.WriteRune(r)
	}
	return nil, nil
}

func handleEscape(p *parser, in, out *bytes.Buffer) (handler, error) {
	r, _, err := in.ReadRune()
	if err == io.EOF {
		return nil, ErrUnexpectedEOF
	}

	if r == '{' {
		return handleEscapeClause, nil
	} else if r == p.escape {
		out.WriteRune(p.escape)
	} else if p.color {
		out.WriteString(ColorCode(string(r)))
	}
	return handleRegular, nil
}

func handleEscapeClause(p *parser, in, out *bytes.Buffer) (handler, error) {
	bs := bytes.NewBufferString("")
	for {
		r, _, err := in.ReadRune()
		if err == io.EOF {
			return nil, ErrUnexpectedEOF
		}

		if r == '}' {
			break
		}
		bs.WriteRune(r)
	}

	if p.color {
		out.WriteString(ColorCode(bs.String()))
	}
	return handleRegular, nil
}
