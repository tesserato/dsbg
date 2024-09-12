package article

import "time"

type Article struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Created     time.Time `yaml:"created"`
	Updated     time.Time `yaml:"updated"`
	Tags        []string  `yaml:"tags"`
	Content     string    `yaml:"-"`
	Files       []string  `yaml:"-"`
	IsPage      bool      `yaml:"-"`
	Path        string    `yaml:"-"`
}
