package server

import "embed"

//go:embed static
var assets embed.FS

func Asset(name string) ([]byte, error) {
	return assets.ReadFile(name)
}
