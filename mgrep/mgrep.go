package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	numWorkers        = 10
	matchesBufferSz   = 100
	filenamesBufferSz = 50
)

type matchInfo struct {
	fileName string
	lineNum  int
	line     string
}

func isValidPath(fp string) bool {
	if _, err := os.Stat(fp); err == nil {
		return true
	}
	return false
}

func isDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func sendFilenames(path string, out chan<- string) {
	if isDir(path) {
		err := filepath.WalkDir(path,
			func(path string, d fs.DirEntry, err error) error {
				if !d.IsDir() {
					out <- path
				}
				return nil
			})
		if err != nil {
			log.Println(err)
		}
	} else {
		out <- path
	}
	close(out)
}

func searchInFile(filename string, pattern string, out chan<- matchInfo) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Can't open file %s\n", filename)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		if strings.Contains(line, pattern) {
			out <- matchInfo{
				fileName: filename,
				lineNum:  lineCount,
				line:     line,
			}
		}
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: mgrep <pattern> <file>")
		os.Exit(1)
	}

	pattern := os.Args[1]
	path := os.Args[2]

	if !isValidPath(path) {
		fmt.Println("Invalid path or you don't have permission to access it")
		return
	}

	wgWorkers := sync.WaitGroup{}

	filenames := make(chan string, filenamesBufferSz)
	go func() {
		defer wgWorkers.Done()
		sendFilenames(path, filenames)
	}()

	matchSlices := make(chan matchInfo, matchesBufferSz)

	wgWorkers.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wgWorkers.Done()
			for filename := range filenames {
				searchInFile(filename, pattern, matchSlices)
			}
		}()
	}

	chanFinished := make(chan struct{})
	go func() {
		wgWorkers.Wait()
		chanFinished <- struct{}{}
	}()

	for {
		select {
		case match := <-matchSlices:
			fmt.Printf("%s:%d:%s\n", match.fileName, match.lineNum, match.line)
		case <-chanFinished:
			fmt.Println("Done!")
			return
		}
	}
}
