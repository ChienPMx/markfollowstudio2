package static

import "embed"

//go:embed background.jpg dist/*
var EmbeddedFiles embed.FS
