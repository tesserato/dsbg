package parse

type Settings struct {
	Title           string
	InputDirectory  string
	OutputDirectory string
	DateFormat      string
	IndexName       string
}

func NewSettings() Settings {
	settings := Settings{}
	settings.Title = "Blog"
	settings.InputDirectory = "."
	settings.OutputDirectory = "content"
	settings.DateFormat = "2006-01-02"
	return settings
}
