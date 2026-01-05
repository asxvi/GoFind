package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	// "time"
)

type Stats struct {
	totCount    int64
	fileCount   int64
	dirCount    int64
	totByteSize uint64
	// maxDepth    int64
}

type ScannedEntry struct {
	Path string
	Info fs.FileInfo
	NumFuncCalls int
}

type ScannedResult struct {
	Files []ScannedEntry
	sync.Mutex
}

// mutex add file
func (res *ScannedResult) AddEntrySafe(entry ScannedEntry) {
	res.Lock()
	defer res.Unlock()
	res.Files = append(res.Files, entry)
}

func printStats(stats Stats) {
	fmt.Printf(`
  Total Count: %d
  File Count:  %d
  Dir Count:   %d
  Total Size:  %d bytes or __ gb`, stats.totCount, stats.fileCount, stats.dirCount, stats.totByteSize)
}

// gofind starting dir [-t target] [-s stats] [-h don't ignore hidden]
func main() {
	cliUtility()
	
	// stats, err := findDirStats("/Users/asxvi/Desktop/", false)
	// if err == nil{
	// 	printStats(stats)
	// }

	// for i:=1; i<=10;i++ {
	// 	startTime := time.Now()
	// 	res, e := findAndExitGoRoutine("/Users/asxvi/Desktop/", "Assessment1SampleQuestions.pdf", i)

	// 	if e != nil {
	// 		fmt.Printf("error: %s\n", e)
	// 	} else {
	// 		fmt.Printf("%s\n", res.Path)
	// 	}
	// 	endTime := time.Since(startTime)

	// 	fmt.Printf("Time for %d go routines: %s\n", i, endTime)
	// }

	b, e := goFind("/Users/asxvi/Desktop/projects", "main.go", 1, true)
	if e == nil{
		fmt.Println(b)
		fmt.Printf("b: %v\n", b)
	}else{
		fmt.Println(e)
	}

}

// have to implement our own recursive function using go routines for cc
// returns error if error, the path if found, or empty struct if nothing found after full traversal
func findAndExitGoRoutine(startingDir string, target string, numConcurrent int) (ScannedEntry, error) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // context to cancel traversing once found

	/*		// should maybe be dynamic by end user		*/
	numCPU := 1

	// use struct bc its smallest data type, literally nothing, but can use other type
	// https://stackoverflow.com/questions/52035390/why-using-chan-struct-when-wait-something-done-not-chan-interface
	semaphore := make(chan struct{}, (numCPU * numConcurrent))
	errorChan := make(chan error, 1) // capture errors and break out
	resultChan := make(chan ScannedEntry, 1)
	var numFuncCalls int64		// just outta curiosity

	// nested definition bc need to use functions vars
	var myCCWalkDir func(string)
	myCCWalkDir = func(currDir string) {
		// defer wg.Done() // on return decrement wg counter		// this was causing an error of negative wg count
		atomic.AddInt64(&numFuncCalls, 1)

		// check if another goroutine already found it
		select {
			case <-ctx.Done():
				return
			default:
		}

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
			fullpath := filepath.Join(currDir, entry.Name())
			if entry.Name() == target { // FOUND and exit
				// fmt.Println("FOUND THE TARGET AT ")
				// fmt.Println(ScannedEntry{Path: fullpath, Info: info})
				
				info, _ := entry.Info()
				select {
				case resultChan <- ScannedEntry{Path: fullpath, Info: info, NumFuncCalls: int(numFuncCalls)}:
					cancel() // FOUND and exit and stop all go routines
				default:
				}
				return
			}

			// semaphor controls traffic. if too many goroutines open at once. THIS one will just hang an wait until other one closes with the deffered slot thats released
			// this ensures that WG and sema counters are synced and complete at the same time
			if entry.IsDir() { 	//recursive traverse this dir
				wg.Add(1)
				go func(fp string) {
					semaphore <- struct{}{}
					defer func() {
						<-semaphore
						wg.Done()
						}() // release slot and decrement WG at SAME time
					
					myCCWalkDir(fp)		// repeat til any of the 3 cases are reached
				}(fullpath)
			}
		}
	}

	// base cc call from startingDir
	wg.Add(1)
	// we add to semaphore and then start from root releasing both semaphore and wg at same time making sure wg == sema
	go func() {
		semaphore <- struct{}{}
		defer func() {
			<-semaphore
			wg.Done()
		}() // release slot and decrement WG at SAME time
		myCCWalkDir(startingDir)	//root
	}()
	
	
	// let all goroutines cook or finish without success
	finishedChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(finishedChan)
	}()

	// https://www.geeksforgeeks.org/go-language/select-statement-in-go-language/ // block until one of the channels fill
	select {
	case rv := <-resultChan:
		return rv, nil
	case err := <-errorChan:
		return ScannedEntry{}, err
	case <-finishedChan:
		return ScannedEntry{}, nil
	}
}

// takes in startingDir, target and finds the stats for entire tree, as well as all the locations of a specific target
// also can take in number of concurrent nodes and skipdotfiles
// return an array of ScannedEntry, a ScannedEntry of target file(s), and error

// have to implement our own recursive function using go routines for cc
// returns error if error, the path if found, or empty struct if nothing found after full traversal
func goFind(startingDir string, target string, numConcurrent int, ignoreHiddenFiles bool) (bool, error) {
	var wg sync.WaitGroup
	
	var resPathRV ScannedResult

	numCPU := 1
	semaphore := make(chan struct{}, numCPU * numConcurrent)
	errorChan := make(chan error)
	// resPathChan := make(chan []ScannedEntry, 1)
	// resArrChan := make(chan []ScannedEntry,1)
	var numFuncCalls int64		// just outta curiosity

	var myCCWalkDir func(string)
	myCCWalkDir = func (currDir string){
		atomic.AddInt64(&numFuncCalls,1)
		semaphore<- struct{}{}
		defer wg.Done()
		defer func () {<-semaphore}()

		entries, err := os.ReadDir(currDir)
		if err != nil{
			select {
			case errorChan <- err:
			default:
			}
			return
		}
		
		resPathBuffer := make([]ScannedEntry, 0, len(entries))
		for _, entry := range entries{
			fullpath := path.Join(currDir, entry.Name())			
			if entry.Name() == target || fullpath == target {	// found goal, so keep track of full path
				info, _ := entry.Info()
				// resPathChan <- ScannedEntry{Path: fullpath, Info: info, NumFuncCalls: int(numFuncCalls)}
				resPathBuffer = append(resPathBuffer, ScannedEntry{Path: fullpath, Info: info, NumFuncCalls: int(numFuncCalls)})
			}
			
			// // recurse here
			// if entry.IsDir(){
			// 	wg.Add(1)
				
			// }
		}

		resPathRV.Lock()
		resPathRV.Files = append(resPathRV.Files, resPathBuffer...)
		resPathRV.Unlock()
	}

	wg.Add(1)
	go myCCWalkDir(startingDir)
	doneChan := make(chan struct{})
	
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
		case <- doneChan:
			return true, nil
		case err := <-errorChan:
			return false, err
	}
}

// func goFind2(startingDir string, target string, numConcurrent int) ([]ScannedEntry, error) {
// 	var wg sync.WaitGroup
// 	var resArr ScannedResult
// 	// var resPath ScannedEntry

// 	/*		// should maybe be dynamic by end user		*/
// 	numCPU := 1

// 	// use struct bc its smallest data type, literally nothing, but can use other type
// 	// https://stackoverflow.com/questions/52035390/why-using-chan-struct-when-wait-something-done-not-chan-interface
// 	semaphore := make(chan struct{}, (numCPU * numConcurrent))
// 	errorChan := make(chan error, 1) // capture errors and break out
// 	doneChan := make(chan struct{})                                // Channel to signal completion.
// 	resPathChan := make(chan ScannedEntry)
// 	var numFuncCalls int64		// just outta curiosity
	
// 	// nested definition bc need to use functions vars
// 	var myCCWalkDir func(string)
// 	myCCWalkDir = func(currDir string) {
// 		atomic.AddInt64(&numFuncCalls, 1)

// 		defer wg.Done()
// 		semaphore<-struct{}{}
// 		defer func ()  {<-semaphore}()

// 		// break out on error, otherwise try to find data
// 		entries, err := os.ReadDir(currDir)
// 		if err != nil {
// 			select {
// 				case errorChan <- err:	//just mark as error	
// 				default:
// 			}
// 			return
// 		}

// 		// go thru every subfile and subdir and search for file
// 		for _, entry := range entries {
// 			fullpath := filepath.Join(currDir, entry.Name())
// 			info, _ := entry.Info()
// 			if entry.Name() == target { // FOUND and exit
// 				resPathChan <-ScannedEntry{Path: fullpath, Info: info, NumFuncCalls: int(numFuncCalls)}
// 			}

// 			// semaphor controls traffic. if too many goroutines open at once. THIS one will just hang an wait until other one closes with the deffered slot thats released
// 			// this ensures that WG and sema counters are synced and complete at the same time
// 			if entry.IsDir() { 	//recursive traverse this dir
// 				wg.Add(1)
// 				go myCCWalkDir(fullpath) // Scan the subdirectory.
// 			}

// 			resArr.Lock()
// 			resArr.Files = append(resArr.Files, ScannedEntry{Path: fullpath, Info: info, NumFuncCalls: int(numFuncCalls)})
// 			resArr.Unlock()
// 		}
// 	}

// 	startingDirInfo, err := os.Stat(startingDir)
// 	if err != nil {
// 		return nil, err
// 	}
// 	resArr.Lock()
// 	resArr.Files = append(resArr.Files, ScannedEntry{Path: startingDir, Info: startingDirInfo, NumFuncCalls: 1})
// 	resArr.Unlock()

// 	// base cc call from startingDir
// 	wg.Add(1)
// 	// we add to semaphore and then start from root releasing both semaphore and wg at same time making sure wg == sema
// 	go func() {
// 		wg.Wait()
// 		close(doneChan)
// 	}()
	
// 	// let all goroutines cook or finish without success
// 	finishedChan := make(chan struct{})
// 	go func() {
// 		wg.Wait()
// 		close(finishedChan)
// 	}()

// 	// https://www.geeksforgeeks.org/go-language/select-statement-in-go-language/ // block until one of the channels fill
// 	select {
// 	case rv := <-resultChan:
// 		return rv, nil
// 	case err := <-errorChan:
// 		return ScannedEntry{}, err
// 	case <-finishedChan:
// 		return ScannedEntry{}, nil
// 	}
// }







// unoptimized search for file within directory tree. Exits when the file is found, or continues until exhausts search options.
// returns the path of file and error
// unused after implementing concurrent version, but keep as backup and comparison between the two 
func findAndExit(startingDir string, target string, dotFilesSkip bool) (string, error) {
	var rv string
	err := filepath.WalkDir(startingDir, findAndExitFunc(target, dotFilesSkip, &rv))
	if err != nil {
		fmt.Println("Error trying to find target: ", err)
		return rv, err
	}
	return rv, err
}

// return a WalkDirFunc that is called later
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

// unoptimized aggregation of directory stats. Computes a Stats struct without using concurrency
// returns Stats struct and error
// unused after implementing concurrent version, but keep as backup and comparison between the two 
func findDirStats(startingDir string, dotFilesSkip bool) (Stats, error) {
	var rv Stats
	err := filepath.WalkDir(startingDir, findDirStatsFunc(dotFilesSkip, &rv))
	if err != nil {
		fmt.Println("Error trying to find directory stats: ", err)
		return rv, err
	}
	return rv, err
}

// the fs.WalkDirFunc helper that is passed in as second parameter in WalkDir method
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

// cli utility to be optimized using cobra and viper 
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

	fmt.Println("CLI util\n\n\n")
}

// useless and unused because nobody want to navigate an ugly terminal gui
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
