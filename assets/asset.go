package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed *
var staticFiles embed.FS

// GetAssets 获取静态资源文件内容
func GetAssets(filename string) ([]byte, error) {
	return staticFiles.ReadFile(fmt.Sprintf("assets/%s", filename))
}

// GetFileSystem 获取可供HTTP服务使用的文件系统
func GetFileSystem() (http.FileSystem, error) {
	subFS, err := fs.Sub(staticFiles, "assets")
	if err != nil {
		return nil, err
	}
	return http.FS(subFS), nil
}

// ListAssets 列出所有可用的静态资源
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
