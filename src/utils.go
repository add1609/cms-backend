package src

import (
	"os"
	"path/filepath"
	"time"
)

type File struct {
	Name         string    `json:"Name"`
	Path         string    `json:"Path"`
	IsDir        bool      `json:"IsDir"`
	Size         int64     `json:"Size"`
	ModifiedTime time.Time `json:"ModifiedTime"`
	Children     []*File   `json:"Children"`
}

func buildJSONTree(rootPath string) (File, error) {
	rootOSFile, err := os.Stat(rootPath)
	if err != nil {
		return File{}, err
	}
	rootFile := toFile(rootOSFile, rootPath)
	stack := []*File{rootFile}
	for 0 < len(stack) {
		file := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		children, _ := os.ReadDir(file.Path)
		for _, currChild := range children {
			currChildInfo, err := currChild.Info()
			if err != nil {
				return File{}, err
			}
			child := toFile(currChildInfo, filepath.Join(file.Path, currChild.Name()))
			file.Children = append(file.Children, child)
			stack = append(stack, child)
		}
	}
	return *rootFile, nil
}

func toFile(file os.FileInfo, path string) *File {
	JSONFile := File{
		Name:         file.Name(),
		Path:         path,
		IsDir:        file.IsDir(),
		Size:         file.Size(),
		ModifiedTime: file.ModTime(),
		Children:     []*File{},
	}
	return &JSONFile
}
