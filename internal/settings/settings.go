package settings

type Settings struct {
	Title            string
	InputDirectory  string
	OutputDirectory string
}

func NewSettings() Settings {
	settings := Settings{}
	settings.Title = "Blog"
	settings.InputDirectory = "."
	settings.OutputDirectory = "content"
	return settings
}
