package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
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

func dirTree(out io.Writer, path string, files bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	fullPath := cwd + string(os.PathSeparator) + path

	content, err := getDirContent(fullPath, files)
	if err != nil {
		return err
	}

	printDirTree(content, out, fullPath, "", files)
	return nil
}

func printDirTree(content []os.FileInfo, out io.Writer, basePath string, prefix string, files bool) error {
	for ind, val := range content {
		line, preff := getLine(prefix, isLastItem(ind, content), val)
		fmt.Fprintln(out, line)

		fullPath := basePath + string(os.PathSeparator) + val.Name()
		content1, err := getDirContent(fullPath, files)

		if e, ok := err.(*os.SyscallError); ok && e.Err == syscall.ENOTDIR {
			continue
		} else if err != nil {
			return err
		}
		printDirTree(content1, out, fullPath, preff, files)
	}
	return nil
}

func getLine(prevPrefix string, isLast bool, fileInfo os.FileInfo) (string, string) {

	if strings.Contains(prevPrefix, "├───") {
		prevPrefix = strings.ReplaceAll(prevPrefix, "├───", "│\t")
	}
	if strings.Contains(prevPrefix, "└───") {
		prevPrefix = strings.ReplaceAll(prevPrefix, "└───", "\t")
	}

	prefix := prevPrefix + "├───"

	if isLast {
		prefix = prevPrefix + "└───"
	}

	result := prefix + fileInfo.Name()

	if fileInfo.IsDir() {
		return result, prefix
	}

	result = result + getSizeString(fileInfo.Size())

	return result, prefix
}

func getSizeString(fileSize int64) string {
	var size string

	if fileSize == 0 {
		size = " (empty)"
	} else {
		size = " (" + strconv.Itoa(int(fileSize)) + "b)"
	}
	return size
}

func isLastItem(ind int, slice []os.FileInfo) bool {
	return ind == len(slice)-1
}

func getDirContent(path string, files bool) ([]os.FileInfo, error) {
	list, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	if files {
		return list, nil
	}
	var filteredList []os.FileInfo
	for _, val := range list {
		if val.IsDir() {
			filteredList = append(filteredList, val)
		}
	}
	return filteredList, nil
}
