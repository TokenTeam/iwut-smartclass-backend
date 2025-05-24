package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed *
var staticFiles embed.FS

func GetAssets(filename string) ([]byte, error) {
	return staticFiles.ReadFile(fmt.Sprintf("assets/%s", filename))
}

func GetFileSystem() (http.FileSystem, error) {
	subFS, err := fs.Sub(staticFiles, "assets")
	if err != nil {
		return nil, err
	}
	return http.FS(subFS), nil
}

func ListAssets() ([]string, error) {
	var assets []string
	err := fs.WalkDir(staticFiles, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			assets = append(assets, path)
		}
		return nil
	})
	return assets, err
}
