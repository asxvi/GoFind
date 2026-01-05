package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// flags
	startDir string
	target   string
	workers  int
	allFiles bool
	stats    bool

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "gofind",
		Short: "A high-performance concurrent file crawler",
		Long: `
	When working on larger projects, unfamiliar codebases, or just learning to program, this utility can assist in getting a quick overview of directory stucture, as well as where exactly certain files may be located (non fzf for now). Combines functionality of popular commands like du, and find while using go routines to decrease search time.`,
		Run: runCommand,

		Example: `  # Search for a file in the current directory
     gofind -f main.go

  # Search in a specific path with extra workers
     gofind --dir /etc --find hosts --workers 20

  # Get directory statistics including hidden files
    gofind -d ./projects -s -a
`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&startDir, "dir", "d", ".", "Starting directory")
	rootCmd.Flags().StringVarP(&target, "find", "f", "", "Target filename to search for")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", runtime.NumCPU(), "Parallel workers")
	rootCmd.Flags().BoolVarP(&allFiles, "all", "a", false, "Include hidden files")
	rootCmd.Flags().BoolVarP(&stats, "stats", "s", false, "Show Directory Statistics")
}

func runCommand(cmd *cobra.Command, args []string) {
	// fmt.Println(startDir, target, workers, allFiles, stats)
	var foundPaths ScannedResult
	var data ScannedResult
	var err error

	foundPaths.Files, data.Files, err = goFind(startDir, target, workers, allFiles)
	if err != nil {
		fmt.Printf("goFind() returned Error: %v", err)
	}

	foundPaths.OutputPaths(target)

	if stats {
		stats, err := convertResultToStats(data.Files)
		if err != nil {
			fmt.Printf("Error converting Stats: %v", err)
		}
		stats.printStats(startDir)
	}
}
