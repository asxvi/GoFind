// package goFind
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"io/fs"
)

type Stats struct (
	FileCount int
)

func main() {
	fileCount, dirCount, count := 0, 0, 0
	startingDirectory := os.Args[1]

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
		fmt.Println(files)
		fmt.Println(dirCount)
		fmt.Println(dirs)
		// for _, name := range(files){
		// 	fmt.pr
		// }
	}

}

func incCount(count *int)  {
	(*count) += 1
	return
}