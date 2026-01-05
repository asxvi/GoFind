package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Stats struct {
	totCount    int64
	fileCount   int64
	dirCount    int64
	totByteSize uint64
	maxDepth    int64
}

type ScannedFile2 struct {
	Path string
	Info fs.FileInfo
	sync.Mutex
}

type ScannedFile struct {
	Path string
	Info fs.FileInfo
}
type ScanResult struct {
	Files []ScannedFile
	sync.Mutex
}

func printStats(stats Stats) {
	fmt.Printf(`
  Total Count: %d
  File Count:  %d
  Dir Count:   %d
  Total Size:  %d bytes or __ gb
	`, stats.totCount, stats.fileCount, stats.dirCount, stats.totByteSize)
}

// gofind starting dir [-t target] [-s stats] [-h don't ignore hidden]
func main() {
	cliUtility()

	for i:=1; i<=10;i++ {
		startTime := time.Now()
		res, e := findAndExitGoRoutine("/Users/asxvi/Desktop/", "Assessment1SampleQuestions.pdf", i)

		// /Users/asxvi/Downloads/Assessment1SampleQuestions.pdf
		if e != nil {
			fmt.Printf("error: %s\n", e)
		} else {
			fmt.Printf("%s\n", res.Path)
		}
		endTime := time.Since(startTime)

		fmt.Printf("Time for %d go routines: %s\n", i, endTime)
	}
}

// have to implement our own recursive function using go routines for cc
// returns error if error, the path if found, or "" if nothing found after full traversal
func findAndExitGoRoutine(startingDir string, target string, numConcurrent int) (ScannedFile2, error) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // context to cancel traversing once found

	/*		// should be dynamic by end user		*/
	numCPU := 1
	// numConcurrent := 4

	// use struct bc its smallest data type, literally nothing, but can use other type
	// https://stackoverflow.com/questions/52035390/why-using-chan-struct-when-wait-something-done-not-chan-interface
	semaphore := make(chan struct{}, (numCPU * numConcurrent))
	errorChan := make(chan error, 1) // capture errors and break out
	resultChan := make(chan ScannedFile2, 1)

	var numFuncCalls int64
	var myCCWalkDir func(string)
	// can perhaps declare this outside or nah bc vars wont be local and will have to pass by ref
	myCCWalkDir = func(currDir string) {
		// defer wg.Done() // on return decrement wg counter		// this was causing an error
		atomic.AddInt64(&numFuncCalls, 1)
		// fmt.Println(wg)

		// check if another goroutine already found it
		select {
		case <-ctx.Done():
			return
		default:
		}

		// semaphore <- struct{}{}	// read in / aquire slot
		// defer func() {<-semaphore}()	// release slot

		// break out on error, otherwise try to find data
		entries, err := os.ReadDir(currDir)
		if err != nil {
			select {
			case errorChan <- err:
				cancel() // stop all other go rutines on error
			default:
			}
			return
		}
		// go thru every subfile and subdir and search for file
		for _, entry := range entries {
			// if err != nil{
			// 	select {
			// 		case errorChan <-err:
			// 			cancel()		// stop all other go rutines on error
			// 		default:
			// 	}
			// 	continue
			// }

			fullpath := filepath.Join(currDir, entry.Name())
			if entry.Name() == target { // FOUND and exit
				// fmt.Println("FOUND THE TARGET AT ")
				// fmt.Println(ScannedFile2{Path: fullpath, Info: info})
				
				info, _ := entry.Info()
				select {
				case resultChan <- ScannedFile2{Path: fullpath, Info: info}:
					cancel() // FOUND and exit and stop all go routines
				default:
				}
				return
			}

			// add to wg. add slot to sema. run recrusive (releases wg) routine and unblock sema
			if entry.IsDir() { //recursive traverse this dir
				// fmt.Println("another dir")
				wg.Add(1)

				go func(p string) {
					semaphore <- struct{}{}
					// release slot and degrement WG
					defer func() {
						<-semaphore
						wg.Done()
					}() // release slot
					myCCWalkDir(p)
				}(fullpath)
			}
		}
	} // sempahore slot released

	// base cc call from startingDir
	wg.Add(1)
	// go myCCWalkDir(startingDir)
	go func() {
		semaphore <- struct{}{}
		// release slot and degrement WG
		defer func() {
			<-semaphore
			wg.Done()
		}() // release slot
		myCCWalkDir(startingDir)
	}()
	
	
	finishedChan := make(chan struct{}) // signal done and exit
	// let all goroutines cook
	go func() {
		// fmt.Println("waiting")
		wg.Wait()
		close(finishedChan)
	}()

	// https://www.geeksforgeeks.org/go-language/select-statement-in-go-language/ // block until one of the channels fill
	select {
	case rv := <-resultChan:
		// fmt.Println("rChan")
		return rv, nil
	case err := <-errorChan:
		// fmt.Println("eChan")
		return ScannedFile2{}, err
	case <-finishedChan:
		// fmt.Println("doneChan")
		return ScannedFile2{}, nil
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

func cliUtility() {
	target := flag.String("t", "", "File or directory to look for")
	stats := flag.Bool("s", false, "Show stats of src directory")
	hiddenFiles := flag.Bool("h", false, "Include hidden files. Default No")
	// flag.Bool("j", false, "Output JSON")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: gofind [flags] <startingDir>")
		return
	}
	startingDir := args[0]

	if *target != "" {
		targetPath, err := findAndExit(startingDir, *target, *hiddenFiles)
		if err != nil {
			fmt.Println("Error: ", err)
		} else if targetPath == "" {
			fmt.Printf("Target %s not found\n", *target)
		} else {
			fmt.Printf("Target %s found at path: %s\n", *target, targetPath)
		}
	}
	if *stats == true {
		dirStats, err := findDirStats(startingDir, *hiddenFiles)
		if err != nil {
			fmt.Println("Error: ", err)
		} else {
			printStats(dirStats)
		}
	}
}

// useless
func cliMenu() {
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
				fmt.Printf("Target %s found at path: %s\n", target, targetPath)
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
