package utils

import (
	"errors"
	"math/rand"

	"github.com/dustyRAIN/leetcode-api-go/leetcodeapi"
)

func GetRandomLeetCodeProblem() (*leetcodeapi.Problem, error) {
	offset := rand.Intn(2000) // random int for offset (make sure to seed rand in init or main if needed)

	problems, err := leetcodeapi.GetAllProblems(offset, 1)
	if err != nil {
		return nil, err
	}

	if len(problems.Problems) == 0 {
		return nil, errors.New("no leetcode problems found")
	}

	problem := problems.Problems[rand.Intn(len(problems.Problems))]

	return &problem, nil
}
