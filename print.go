// Package numprint pretty prints sequences of digits.
//
// For the github.com/keep94/sqrt package, sqrt.Sequence implements Printable
// and sqrt.FiniteSequence implements both Writable and Printable.
//
// While this package is meant to be used with the data structures in the
// github.com/keep94/sqrt package, it will work with anything that
// implements the Printable or Writable interface.
package numprint

import (
	"io"
	"iter"
	"os"
	"strings"
)

// Printable represents a sequence of digits between 0-9 with contiguous
// positions that can be printed with Print(), Fprint(), or Sprint().
type Printable interface {

	// AllInRange returns the 0 based position and value of each digit in
	// this Printable from position start up to but not including position
	// end.
	AllInRange(start, end int) iter.Seq2[int, int]
}

// Writable represents a sequence of digits between 0-9 with contiguous
// positions that can be printed with Write(), Fwrite(), or Swrite().
type Writable interface {

	// All returns the 0 based position and value of each digit in this
	// Writable from beginning to end.
	All() iter.Seq2[int, int]

	// Backward returns the 0 based position and value of each digit in this
	// Writable from end to beginning.
	Backward() iter.Seq2[int, int]
}

// Option represents an option for printing.
type Option interface {
	mutate(p *printerSettings)
}

// DigitsPerRow sets the number of digits per row. Zero or negative means no
// separate rows.
func DigitsPerRow(count int) Option {
	return optionFunc(func(p *printerSettings) {
		p.digitsPerRow = count
	})
}

// DigitsPerColumn sets the number of digits per column. Zero or negative
// means no separate columns.
func DigitsPerColumn(count int) Option {
	return optionFunc(func(p *printerSettings) {
		p.digitsPerColumn = count
	})
}

// ShowCount shows the digit count in the left margin if on is true.
func ShowCount(on bool) Option {
	return optionFunc(func(p *printerSettings) {
		p.showCount = on
	})
}

// MissingDigit sets the character to represent a missing digit.
func MissingDigit(missingDigit rune) Option {
	return optionFunc(func(p *printerSettings) {
		p.missingDigit = missingDigit
	})
}

// TrailingLF adds a trailing line feed to what is printed if on is true.
func TrailingLF(on bool) Option {
	return optionFunc(func(p *printerSettings) {
		p.trailingLineFeed = on
	})
}

// LeadingDecimal prints "0." before the first digit if on is true.
func LeadingDecimal(on bool) Option {
	return optionFunc(func(p *printerSettings) {
		p.leadingDecimal = on
	})
}

func bufferSize(size int) Option {
	return optionFunc(func(p *printerSettings) {
		p.bufferSize = size
	})
}

// Fprint prints digits of s to w. Unless using advanced functionality,
// prefer Fwrite, Write, and Swrite to Fprint, Print, and Sprint.
// Fprint returns the number of bytes written and any error encountered.
// p contains the positions of the digits to print.
// For options, the default is 50 digits per row, 5 digits per column,
// show digit count, period (.) for missing digits, don't write a trailing
// line feed, and show the leading decimal point.
func Fprint(w io.Writer, s Printable, p Positions, options ...Option) (
	written int, err error) {
	settings := &printerSettings{
		digitsPerRow:    50,
		digitsPerColumn: 5,
		showCount:       true,
		missingDigit:    '.',
		leadingDecimal:  true,
	}
	printer := newPrinter(w, p.End(), mutateSettings(options, settings))
	fromSequenceWithPositions(s, p, printer)
	printer.Finish()
	return printer.BytesWritten(), printer.Err()
}

// Fwrite writes all the digits of s to w. Fwrite returns the number of bytes
// written and any error encountered. For options, the default is 50 digits
// per row, 5 digits per column, show digit count, period (.) for missing
// digits, write a trailing line feed, and don't show the leading decimal
// point.
func Fwrite(w io.Writer, s Writable, options ...Option) (
	written int, err error) {
	settings := &printerSettings{
		digitsPerRow:     50,
		digitsPerColumn:  5,
		showCount:        true,
		missingDigit:     '.',
		trailingLineFeed: true,
	}
	printer := newPrinter(w, endOf(s), mutateSettings(options, settings))
	fromIterator(s.All(), printer)
	printer.Finish()
	return printer.BytesWritten(), printer.Err()
}

// Sprint works like Fprint and prints digits of s to a string.
func Sprint(s Printable, p Positions, options ...Option) string {
	var builder strings.Builder
	Fprint(&builder, s, p, options...)
	return builder.String()
}

// Swrite works like Fwrite and writes all the digits of s to returned string.
func Swrite(s Writable, options ...Option) string {
	var builder strings.Builder
	Fwrite(&builder, s, options...)
	return builder.String()
}

// Print works like Fprint and prints digits of s to stdout.
func Print(s Printable, p Positions, options ...Option) (
	written int, err error) {
	return Fprint(os.Stdout, s, p, options...)
}

// Write works like Fwrite and writes all the digits of s to stdout.
func Write(s Writable, options ...Option) (
	written int, err error) {
	return Fwrite(os.Stdout, s, options...)
}

func endOf(s Writable) int {
	for index := range s.Backward() {
		return index + 1
	}
	return 0
}

func fromSequenceWithPositions(s Printable, p Positions, printer *printer) {
	for pr := range p.All() {
		fromIterator(s.AllInRange(pr.Start, pr.End), printer)
	}
}

func fromIterator(it iter.Seq2[int, int], printer *printer) {
	if !printer.CanConsume() {
		return
	}
	for posit, digit := range it {
		printer.Consume(posit, digit)
		if !printer.CanConsume() {
			return
		}
	}
}

type optionFunc func(p *printerSettings)

func (o optionFunc) mutate(p *printerSettings) {
	o(p)
}

func mutateSettings(
	options []Option, settings *printerSettings) *printerSettings {
	for _, option := range options {
		option.mutate(settings)
	}
	return settings
}
