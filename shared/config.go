package cursheet

// Column defines the attributes of the individual columns in the speadsheet
type Column struct {
	Name	string
	Ctype	string
	Size	int
	LogPos	int
}

// Config defines the mapping between the cursor and the output sheet
type Config struct {
	Title		string
	Schema		string
	Procedure	string 
	Cols		[]Column
}
