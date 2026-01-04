package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type Stats struct {
	totCount    int64
	fileCount   int64
	dirCount    int64
	totByteSize uint64
	maxDepth    int64
}

func printStats(stats Stats) {
	fmt.Printf(`
		Total Count: %d
		File Count:  %d
		Dir Count:   %d
		Total Size:  %d bytes or __ gb
	`, stats.totCount, stats.fileCount, stats.dirCount, stats.totByteSize)
}

func main() {
	cliUtility()
}

func cliUtility() {
	var startingDir string
	fmt.Print("Enter Starting directory: ")
	fmt.Scanf("%s", &startingDir)

	// ADD functionality for $HOME and other CLI functionality

	input := "0"
	for input != "-1" {

		fmt.Print("Enter\n  * 1 to find a file\n  * 2 for directory stats\n  * -1 to quit\n")
		fmt.Scanf("%s", &input)

		switch input {
		case "1":
			var target string
			fmt.Print("Enter target: ")
			fmt.Scanf("%s", &target)
			var dotFilesSkip bool
			var dotFileChar string
			fmt.Print("skip hidden stuff (y/n): ")
			fmt.Scanf("%s", &dotFileChar)

			if dotFileChar == "y" {
				dotFilesSkip = true
			} else {
				dotFilesSkip = false
			}

			targetPath, err := findAndExit(startingDir, target, dotFilesSkip)
			if err != nil {
				fmt.Println("Error: ", err)
			} else if targetPath == "" {
				fmt.Printf("Target %s not found", target)
			} else {
				fmt.Printf("Target %s found at path: %s", target, targetPath)
			}
		case "2":
			dirStats, err := findDirStats(startingDir, true)
			if err != nil {
				fmt.Println("Error: ", err)
			} else {
				printStats(dirStats)
			}
		default:
			fmt.Println("invalid input. try again")
		}
	}
}

func findAndExit(startingDir string, target string, dotFilesSkip bool) (string, error) {
	var rv string
	err := filepath.WalkDir(startingDir, findAndExitFunc(target, dotFilesSkip, &rv))
	if err != nil {
		fmt.Println("Error trying to find target: ", err)
		return rv, err
	}
	return rv, err
}

// return a WalkDirFunc later called
func findAndExitFunc(target string, dotfilesSkip bool, rv *string) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dotfilesSkip && d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		} else if d.Name() == target {
			*rv = path
			return filepath.SkipAll // early break out bc target found
		}
		return nil
	}
}

func findDirStats(startingDir string, dotFilesSkip bool) (Stats, error) {
	var rv Stats
	err := filepath.WalkDir(startingDir, findDirStatsFunc(dotFilesSkip, &rv))
	if err != nil {
		fmt.Println("Error trying to find directory stats: ", err)
		return rv, err
	}
	return rv, err
}

func findDirStatsFunc(dotfilesSkip bool, stats *Stats) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if dotfilesSkip && d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		} else {
			if d.IsDir() {
				stats.dirCount++
				stats.totCount++
				return nil
			} else if d.Type().IsRegular() {
				stats.fileCount++
				stats.totCount++

				info, err := d.Info()
				if err != nil {
					return err
				}
				stats.totByteSize += uint64(info.Size())
				return nil
			}
		}
		return nil
	}
}
