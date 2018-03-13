package cursheet

import (
	"github.com/tealeg/xlsx"
)

// Flag holds any multi-variable fields
type Flag struct {
	Flag		string	
}

// Column defines the attributes of the individual columns in the speadsheet
type Column struct {
	Name		string
	Ctype		string
	Format		string
	Size		float64
	LogPos		int
	ShowPos 	int
	Sum			bool
	Subtotal	bool
	Style		xlsx.Style
}

// Config defines the mapping between the cursor and the output sheet
type Config struct {
	Title		string
	SubjLine	string
	Filename	string
	Schema		string
	Procedure	string
	Typeface	string
	Typesize	int
	HeadItalic	bool
	HeadBold	bool
	FreezeRows	float64
	Subtotal	bool
	SubCol		int
	Subflags	[]Flag
	Cols		[]Column
}
