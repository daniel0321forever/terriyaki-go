package services

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"

	"github.com/dustyRAIN/leetcode-api-go/leetcodeapi"
)

// additional (externally defined) metadata from CSV (e.g., leetcode id, topic)
type CustomProblem struct {
	ID   int
	Slug string
	Tag  string
}

var (
	// loadedProblemLists stores metadata for loaded problem lists
	loadedProblemLists = make(map[string][]CustomProblem)
	listsMutex         sync.RWMutex
)

// LoadExternalProblemList loads a problem list from a CSV file and return an array of CustomProblem
// The CSV should have format: id,slug,tag
func LoadExternalProblemList(listName string) ([]CustomProblem, error) {
	// Check if already loaded
	listsMutex.RLock()
	if list, exists := loadedProblemLists[listName]; exists {
		listsMutex.RUnlock()
		return list, nil
	}
	listsMutex.RUnlock()

	// Load from file
	listsMutex.Lock()
	defer listsMutex.Unlock()

	// Double-check after acquiring write lock
	if list, exists := loadedProblemLists[listName]; exists {
		return list, nil
	}

	// Get the directory where this file is located
	_, filename, _, _ := runtime.Caller(0)
	utilsDir := filepath.Dir(filename)
	csvPath := filepath.Join(utilsDir, listName+".csv")

	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s.csv: %w", listName, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	var problems []CustomProblem
	for i, record := range records {
		// Skip header row
		if i == 0 {
			continue
		}

		if len(record) < 3 {
			continue
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}

		problems = append(problems, CustomProblem{
			ID:   id,
			Slug: record[1],
			Tag:  record[2],
		})
	}

	if len(problems) == 0 {
		return nil, fmt.Errorf("no problems found in %s.csv", listName)
	}

	loadedProblemLists[listName] = problems
	return problems, nil
}

// return a problem by problem id shown on LeetCode
func GetProblemById(id int) (*leetcodeapi.Problem, error) {
	problems, err := leetcodeapi.GetAllProblems(id-1, 1) // transform id to zero-index

	if err != nil {
		return nil, err
	}

	problem := problems.Problems[0]
	return &problem, nil
}

// returns a random problem from the specified list
func GetRandomProblemFromList(listName string) (*leetcodeapi.Problem, error) {
	problems, err := LoadExternalProblemList(listName)
	if err != nil {
		return nil, err
	}

	if len(problems) == 0 {
		return nil, errors.New("no leetcode problems found")
	}

	randomIndex := rand.Intn(len(problems))

	problemId := problems[randomIndex].ID
	problem, err := GetProblemById(problemId)

	if err != nil {
		return nil, err
	}

	return problem, nil
}

// returns a random problem from all LeetCode problems
func GetRandomLeetCodeProblem() (*leetcodeapi.Problem, error) {
	offset := rand.Intn(2000) // random int for offset (make sure to seed rand in init or main if needed)

	problem, err := GetProblemById(offset)
	if err != nil {
		return nil, err
	}

	return problem, nil
}
