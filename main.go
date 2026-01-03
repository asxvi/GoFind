// package goFind
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// type Stats struct (
// 	FileCount int
// )

func main() {
	startingDirectory := os.Args[1]
	target := "main.go"
	temp := findAndExit(startingDirectory, target, true)
	println(temp)
	return 


	fileCount, dirCount, count := 0, 0, 0
	var files []string
	var dirs []string
	// https://dev.to/rezmoss/file-system-walking-with-walkdir-recursive-tree-traversal-49-dj3
	err := filepath.WalkDir(startingDirectory, func(path string, d fs.DirEntry, err error) error{
		if d.IsDir(){
			dirCount +=1
			dirs = append(dirs, d.Name())
		}else if d.Type().IsRegular(){
			fileCount +=1
			files = append(files, d.Name())
		}
		count +=1
		return nil
	})
	if err == nil{
		fmt.Println(count)
		fmt.Println(fileCount)
		// fmt.Println(files)
		fmt.Println(dirCount)
		// fmt.Println(dirs)
	}

}

func findAndExit(startingDir string, target string, dotFilesSkip bool) string {
	rv := ""
	
	err := filepath.WalkDir(startingDir, findAndExitFunc(target, dotFilesSkip, &rv))
	if err != nil {
		fmt.Println("We have an error on aisle 5")
	}
	return rv
}

// return a WalkDirFunc later called 
func findAndExitFunc(target string, dotfilesSkip bool, rv *string) fs.WalkDirFunc{
	return func(path string, d fs.DirEntry, err error) error {
		if dotfilesSkip && d.IsDir() && strings.HasPrefix(d.Name(), ".") { 
			return filepath.SkipDir
		} else if d.Name() == target {
			(*rv) += path
			return filepath.SkipAll // early break out bc target found
		}
		return nil
	}
}