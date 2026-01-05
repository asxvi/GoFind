# goFind
#### A concurrent CLI utility for getting overview of projects and directories.

## Motivation
When working on larger projects, unfamiliar codebases, or just learning to program this utility can assist in getting a quick overview of directory stucture, as well as where exactly certain files may be located (non fzf for now). Combines functionality of popular commands like du, and find while using go routines to decrease search time.

## Features
- Recursive directory traversal
- File and directory counts
- Total size calculation
- Optional skipping of hidden directories
- Early-exit file search

Combines functionality of popular commands like du, and find

Answers questions like:
- How many files and directories are here?
- How large is this directory?
- Does a specific file exist somewhere in the tree?

## Installation
go install github.com/asxvi/GoFind@latest

## Examples
Show an overview of etc/ directory, and find all paths for 'hosts' in etc/ directory.

`gofind -dir /etc -find hosts -stats`

Concurrently find all paths for 'main.go' within current directory with 20 nodes and without ignoring hidden folders/files.

`gofind -f main.go -w 20 -a`

View the CLI man page

`gofind -h`

## Improvements
- [X] concurrency in go
- [X] add CLI flags and remove menu
- [X] early exit 
- [X] switch CLI tool to use cobra and viper rather than flag
- [] handle system errors
- [] add Comparison and benchmarks
- [] add more features i.e JSON tree (idk if logical when working with 50k files)
