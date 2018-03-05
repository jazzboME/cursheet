package cursheet

// Column defines the attributes of the individual columns in the speadsheet
type Column struct {
	Name	string
	Ctype	string
	Size	uint32
	LogPos	int
	ShowPos int
}

// Config defines the mapping between the cursor and the output sheet
type Config struct {
	Title		string
	Schema		string
	Procedure	string 
	Cols		[]Column
}
