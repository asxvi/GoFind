# goFind
<!-- #### A CLI utility for getting overview of projects and directories. -->
### A safe, readable alternative to find + du

## Motivation
When working on larger projects, unfamiliar codebases, or just learning to program this utility can assist in understanding what you are working with.

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


## Improvements
- [] concurrency in go
- [X] add CLI flags and remove menu
- [] add more features i.e JSON tree
- [] early exit 
