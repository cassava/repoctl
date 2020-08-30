// Copyright (c) 2020, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package term

import (
	"io"
	"os"

	"github.com/goulash/color"
)

var (
	Formatter *color.Colorizer = color.New()

	ErrorOut io.Writer = os.Stderr
	WarnOut  io.Writer = os.Stderr
	StdOut   io.Writer = os.Stdout
	DebugOut io.Writer = os.Stderr
)

func SetMode(mode string) {
	Formatter.Set(mode)
}

// ------------------------------------------------------------------ //

func Printf(format string, obj ...interface{}) (int, error) {
	if StdOut == nil {
		return 0, nil
	}
	return Formatter.Fprintf(StdOut, format, obj...)
}

func Sprintf(format string, obj ...interface{}) string {
	return Formatter.Sprintf(format, obj...)
}

func Println(obj ...interface{}) (int, error) {
	return Formatter.Fprintln(StdOut, obj...)
}

// ------------------------------------------------------------------ //

func Errorf(format string, obj ...interface{}) (int, error) {
	return ferrorf(ErrorOut, format, obj...)
}

func Errorff(format string, obj ...interface{}) (int, error) {
	return ferrorff(ErrorOut, format, obj...)
}

func ferrorf(w io.Writer, format string, obj ...interface{}) (int, error) {
	if w == nil {
		return 0, nil
	}
	return Formatter.Fprintf(w, "@{r}"+format, obj...)
}

func ferrorff(w io.Writer, format string, obj ...interface{}) (int, error) {
	if w == nil {
		return 0, nil
	}
	return Formatter.Fprintf(w, "@{.r}"+format, obj...)
}

type ErrorWriter struct {
	File io.Writer
}

func NewErrorWriter(w io.Writer) *ErrorWriter {
	return &ErrorWriter{
		File: w,
	}
}

func (w *ErrorWriter) Write(p []byte) (n int, err error) {
	return ferrorf(w.File, "%s", p)
}

// ------------------------------------------------------------------ //

func Warnf(format string, obj ...interface{}) (int, error) {
	return fwarnf(WarnOut, format, obj...)
}

func Warnff(format string, obj ...interface{}) (int, error) {
	return fwarnff(WarnOut, format, obj...)
}

func fwarnf(w io.Writer, format string, obj ...interface{}) (int, error) {
	if w == nil {
		return 0, nil
	}
	return Formatter.Fprintf(w, "@{y}"+format, obj...)
}

func fwarnff(w io.Writer, format string, obj ...interface{}) (int, error) {
	if w == nil {
		return 0, nil
	}
	return Formatter.Fprintf(w, "@{.y}"+format, obj...)
}

type WarnWriter struct {
	File io.Writer
}

func NewWarnWriter(w io.Writer) *WarnWriter {
	return &WarnWriter{
		File: w,
	}
}

func (w *WarnWriter) Write(p []byte) (n int, err error) {
	return fwarnf(w.File, "%s", p)
}

// ------------------------------------------------------------------ //

func Debugf(format string, obj ...interface{}) (int, error) {
	return fdebugf(DebugOut, format, obj...)
}

func Debugff(format string, obj ...interface{}) (int, error) {
	return fdebugff(DebugOut, format, obj...)
}

func fdebugf(w io.Writer, format string, obj ...interface{}) (int, error) {
	if w == nil {
		return 0, nil
	}
	return Formatter.Fprintf(w, "@{.}"+format, obj...)
}

func fdebugff(w io.Writer, format string, obj ...interface{}) (int, error) {
	if w == nil {
		return 0, nil
	}
	return Formatter.Fprintf(w, "@{.}"+format, obj...)
}

type DebugWriter struct {
	File io.Writer
}

func NewDebugWriter(w io.Writer) *DebugWriter {
	return &DebugWriter{
		File: w,
	}
}

func (w *DebugWriter) Write(p []byte) (n int, err error) {
	return fdebugf(w.File, "%s", p)
}
