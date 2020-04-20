package main

import (
	"fmt"
	gogit "github.com/go-git/go-git/v5"
	gogit_plumbing "github.com/go-git/go-git/v5/plumbing"
	gogit_diff "github.com/go-git/go-git/v5/plumbing/format/diff"
	gogit_object "github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type lineOrRange struct {
	start int
	end   int
}

type hunk struct {
	contentRange *lineOrRange
	contentLines []string
	opType       gogit_diff.Operation
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

	baseTree := getGitTree(repo, baseSha)

	baseLines := getBaseLines(baseTree, baseFilePath)
	baseLines = baseLines[baseRange.start-1 : baseRange.end]

	cmpTree := getGitTree(repo, cmpSha)

	chunks := buildChunks(baseTree, cmpTree)

	files, hunks := buildFilesAndHunks(chunks)

	findSectionInPatches(baseFilePath, baseLines, baseRange, files, hunks)
}

func findSectionInPatches(path string, baseLines []string, baseRange *lineOrRange, files map[string][]string, hunksMap map[string][]hunk) {
	// check the same file path first
	hunks, existed := hunksMap[path]
	resultRange := &lineOrRange{
		start: baseRange.start,
		end:   baseRange.end,
	}
	if existed {
		delete(hunksMap, path)
		file := files[path]
		// Calculate the up-to-date range if there is modification happening
		// before the base code section
		for _, h := range hunks {
			if h.contentRange.start <= resultRange.start {
				delta := (h.contentRange.end - h.contentRange.start) + 1
				if h.opType == gogit_diff.Add {
					resultRange.start += delta
					resultRange.end += delta
				} else {
					resultRange.start -= delta
					resultRange.end -= delta
				}
			} else {
				break
			}
		}
		if resultRange.start < 1 { // Chances are the code itself got deleted
			resultRange = nil
		} else {
			for i, l := range file[resultRange.start-1 : resultRange.end] {
				if strings.Compare(baseLines[i], l) != 0 { // FIXME: Naive line-by-line comparison
					resultRange = nil
					break
				}
			}
		}
		if resultRange != nil {
			fmt.Printf("Given code found in %s#L%d-%d\n", path, resultRange.start, resultRange.end)
			return
		}
	} else {
		fmt.Printf("Given code found in %s#L%d-%d [File not changed]\n", path, baseRange.start, baseRange.end)
		return
	}

	// check if the given code has been moved to another file
	var wg sync.WaitGroup
	resChan := make(chan string, len(hunksMap))

	for outerHunkPath, outerHunks := range hunksMap {
		wg.Add(1)
		go func(hunkPath string, hunks []hunk, waitgroup *sync.WaitGroup) {
			defer waitgroup.Done()
			for _, h := range hunks {
				if h.opType == gogit_diff.Delete {
					continue
				}

				i := 0
				start := -1
				for j, l := range h.contentLines {
					// Compare line-by-line to handle when the hunk including
					// the target has other changes along with it
					if i < len(baseLines) && strings.Compare(baseLines[i], l) == 0 {
						if start == -1 {
							start = j
						}
						i += 1
					}
				}

				if i == len(baseLines) {
					resChan <- fmt.Sprintf("Given code found in %s#L%d-%d\n", hunkPath, h.contentRange.start+start, h.contentRange.start+start+len(baseLines)-1)
					return
				}
			}
		}(outerHunkPath, outerHunks, &wg)
	}

	wg.Wait()
	if len(resChan) == 0 {
		fmt.Println("Given code is not found")
		return
	}
	for i := 0; i < len(resChan); i++ {
		fmt.Print(<-resChan)
	}
}

func getBaseLines(baseTree *gogit_object.Tree, baseFilePath string) []string {
	baseFile, err := baseTree.File(baseFilePath)
	if err != nil {
		log.Fatalf("Failed to get base file: %s\n", baseFilePath)
	}

	baseContent, err := baseFile.Lines()
	if err != nil {
		log.Fatalf("Failed to get base file content: %s\n", err)
	}

	return baseContent
}

func buildChunks(from, to *gogit_object.Tree) map[string][]gogit_diff.Chunk {
	patch, err := from.Patch(to)
	if err != nil {
		log.Fatalf("Failed to obtain the patch: %s\n", err)
	}

	filePatches := patch.FilePatches()

	chunks := make(map[string][]gogit_diff.Chunk)
	for _, filePatch := range filePatches {
		if filePatch.IsBinary() {
			continue
		}

		_, to := filePatch.Files()
		// skip deleted file
		if to == nil {
			continue
		}
		chunks[to.Path()] = filePatch.Chunks()
	}

	return chunks
}

func buildFilesAndHunks(chunksMap map[string][]gogit_diff.Chunk) (map[string][]string, map[string][]hunk) {
	type result struct {
		hunks []hunk
		file  []string
		path  string
	}
	resChan := make(chan result, len(chunksMap))
	var wg sync.WaitGroup

	for chunksPath, chunks := range chunksMap {
		wg.Add(1)
		go func(path string, cs []gogit_diff.Chunk, waitgroup *sync.WaitGroup) {
			defer waitgroup.Done()
			lineCursor := 0
			h := make([]hunk, 0)
			file := make([]string, 0)
			for _, c := range cs {
				lines := strings.Split(c.Content(), "\n") // naively use UNIX linebreak
				if lines[len(lines)-1] == "" {
					lines = lines[:len(lines)-1]
				}
				switch c.Type() {
				case gogit_diff.Equal:
					lineCursor += len(lines)
					file = append(file, lines...)
					printDiff("=", lines)
				case gogit_diff.Add:
					h = append(h, hunk{
						contentRange: &lineOrRange{
							start: lineCursor + 1,
							end:   lineCursor + len(lines),
						},
						contentLines: lines,
						opType:       gogit_diff.Add,
					})
					lineCursor += len(lines)
					file = append(file, lines...)
					printDiff("+", lines)
				case gogit_diff.Delete:
					h = append(h, hunk{
						contentRange: &lineOrRange{
							start: lineCursor + 1,
							end:   lineCursor + len(lines),
						},
						contentLines: lines,
						opType:       gogit_diff.Delete,
					})
				}
			}
			resChan <- result{
				hunks: h,
				file:  file,
				path:  path,
			}
		}(chunksPath, chunks, &wg)
	}

	wg.Wait()

	hunks := make(map[string][]hunk)
	files := make(map[string][]string)
	for range chunksMap {
		res := <-resChan
		hunks[res.path] = res.hunks
		files[res.path] = res.file
	}

	return files, hunks
}

func printDiff(prefix string, lines []string) {
	if os.Getenv("THESEUS_DEBUG") != "1" {
		return
	}
	for _, l := range lines {
		fmt.Printf("%s %s\n", prefix, l)
	}
}

func getGitTree(repo *gogit.Repository, sha string) *gogit_object.Tree {
	commit, err := repo.CommitObject(gogit_plumbing.NewHash(sha))
	if err != nil {
		log.Fatalf("Failed to get commit: %s\n", sha)
	}

	tree, err := repo.TreeObject(commit.TreeHash)
	if err != nil {
		log.Fatalf("Failed to get tree: %s\n", commit.TreeHash)
	}
	return tree
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

	if start == 0 {
		log.Fatalln("Given range should have a start > 0")
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
