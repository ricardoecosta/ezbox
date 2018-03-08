package media

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/pkg/errors"
)

// todo: offload index building to go routine

type Index struct {
	files []string
}

var (
	index = &Index{files: make([]string, 0)}
)

func InitCollection(mediaDirectories []string) {
	for _, dir := range mediaDirectories {
		log.Printf("indexing media collection, dir=%v", dir)
		indexDir(dir, index)
	}
	log.Printf("indexing complete, entries=%v", len(index.files))
}

func PrintIndexedMediaCollection() {
	log.Println("media collection:")
	for _, file := range index.files {
		log.Printf("\t-> %v", file)
	}
}

func indexDir(baseDir string, index *Index) error {
	dirs, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return errors.WithMessage(
			err, fmt.Sprintf("error indexing media directory, dir=%v", baseDir))
	}
	for _, dir := range dirs {
		absoluteDir := filepath.Join(baseDir, dir.Name())
		if dir.IsDir() {
			indexDir(absoluteDir, index)
		} else if FileIsSupported(absoluteDir) {
			index.files = append(index.files, absoluteDir)
		}
	}
	return nil
}
