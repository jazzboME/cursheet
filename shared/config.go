package cursheet

import (
	"github.com/tealeg/xlsx"
)

// Column defines the attributes of the individual columns in the speadsheet
type Column struct {
	Name		string
	Ctype		string
	Size		float64
	LogPos		int
	ShowPos 	int
	Sum			bool
	Style		xlsx.Style
}

// Config defines the mapping between the cursor and the output sheet
type Config struct {
	Title		string
	Schema		string
	Procedure	string
	Typeface	string
	Typesize	int
	HeadItalic	bool
	HeadBold	bool 
	Cols		[]Column
}
