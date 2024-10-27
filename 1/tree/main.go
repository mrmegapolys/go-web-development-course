package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	return dirTreeRecursive(out, path, printFiles, "")
}

func dirTreeRecursive(out io.Writer, path string, printFiles bool, levelPrefix string) error {
	entries, err := getDirEntries(path, printFiles)
	if err != nil {
		return err
	}

	for idx, entry := range entries {
		entryPath := path + string(os.PathSeparator) + entry.Name()
		currentPrefix := buildCurrentPrefix(idx, len(entries))

		if entry.IsDir() {
			_, err = fmt.Fprintln(out, levelPrefix+currentPrefix+entry.Name())
			if err != nil {
				panic(err)
			}

			nextLevelPrefix := buildNextLevelPrefix(idx, len(entries))

			err := dirTreeRecursive(out, entryPath, printFiles, levelPrefix+nextLevelPrefix)
			if err != nil {
				return err
			}
		} else {
			fileInfo, err := getFileInfo(entryPath)
			if err != nil {
				return err
			}

			size := fileInfo.Size()
			var sizeString string
			if size == 0 {
				sizeString = "empty"
			} else {
				sizeString = strconv.FormatInt(size, 10) + "b"
			}

			_, err = fmt.Fprintf(out, "%s (%s)\n", levelPrefix+currentPrefix+fileInfo.Name(), sizeString)
			if err != nil {
				panic(err)
			}
		}
	}

	return nil
}

func getDirEntries(path string, printFiles bool) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list entries in directory %s: %v", path, err)
	}

	if !printFiles {
		entries = filterOutFiles(entries)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	return entries, nil
}

func filterOutFiles(entries []os.DirEntry) []os.DirEntry {
	result := make([]os.DirEntry, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			result = append(result, entry)
		}
	}

	return result
}

func getFileInfo(entryPath string) (os.FileInfo, error) {
	file, err := os.Open(entryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", entryPath, err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get info for file %s: %v", file.Name(), err)
	}

	return fileInfo, nil
}

func buildCurrentPrefix(idx int, size int) string {
	if idx != size-1 {
		return "├───"
	} else {
		return "└───"
	}
}

func buildNextLevelPrefix(idx int, size int) string {
	if idx != size-1 {
		return "│\t"
	} else {
		return "\t"
	}
}
