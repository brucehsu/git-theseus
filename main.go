package main

import (
	"fmt"
	gogit "github.com/go-git/go-git/v5"
	gogit_plumbing "github.com/go-git/go-git/v5/plumbing"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type lineOrRange struct {
	start int
	end   int
}

func main() {
	if len(os.Args) != 5 {
		fmt.Println("usage: git-theseus [BASE_SHA] [FILE_PATH] [LINE_OR_RANGE] [COMPARE_SHA]")
		os.Exit(1)
	}
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Unable to get current working directory")
	}

	baseSha := os.Args[1]
	baseFilePath := os.Args[2]
	cmpSha := os.Args[4]
	baseRange := parseLineOrRange(os.Args[3])

	isSha(baseSha)
	isSha(cmpSha)

	repo, err := gogit.PlainOpen(pwd)
	if err != nil {
		log.Fatalf("Failed to open git repo at: %s\n", pwd)
	}

	baseCommit, err := repo.CommitObject(gogit_plumbing.NewHash(baseSha))
	if err != nil {
		log.Fatalf("Failed to get base commit: %s\n", baseSha)
	}

	baseTree, err := repo.TreeObject(baseCommit.TreeHash)
	if err != nil {
		log.Fatalf("Failed to get base tree: %s\n", baseCommit.TreeHash)
	}

	baseFile, err := baseTree.File(baseFilePath)
	if err != nil {
		log.Fatalf("Failed to get base file: %s\n", baseFilePath)
	}

	baseContent, err := baseFile.Lines()
	if err != nil {
		log.Fatalf("Failed to get base file content: %s\n", err)
	}

	for l := baseRange.start - 1; l < baseRange.end; l++ {
		fmt.Println(baseContent[l])
	}
}

func isSha(arg string) {
	re := regexp.MustCompile("[0-9a-f]{7,40}")
	if !re.MatchString(arg) {
		log.Fatalf("Invalid SHA: %s\n", arg)
	}
}

func parseLineOrRange(arg string) *lineOrRange {
	splitted := strings.Split(arg, "-")
	var start, end int

	start, err := strconv.Atoi(splitted[0])
	if err != nil {
		log.Fatalf("Failed to parse given line/range: %v\n", err)
	}

	if len(splitted) > 1 {
		end, err = strconv.Atoi(splitted[1])
		if err != nil {
			log.Fatalf("Failed to parse given line/range: %v\n", err)
		}
	} else {
		end = start
	}

	return &lineOrRange{
		start: start,
		end:   end,
	}
}
