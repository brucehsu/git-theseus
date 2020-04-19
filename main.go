package main

import (
	"fmt"
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
	if len(os.Args) != 4 {
		fmt.Println("usage: git-theseus [BASE_SHA] [LINE_OR_RANGE] [COMPARE_SHA]")
		os.Exit(1)
	}
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Unable to get current working directory")
	}

	baseSha := os.Args[1]
	cmpSha := os.Args[3]
	baseRange := parseLineOrRange(os.Args[2])

	isSha(baseSha)
	isSha(cmpSha)

	fmt.Println(pwd, baseSha, cmpSha, baseRange)
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
	}

	return &lineOrRange{
		start: start,
		end:   end,
	}
}
