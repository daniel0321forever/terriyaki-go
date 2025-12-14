package services

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/dustyRAIN/leetcode-api-go/leetcodeapi"
	"golang.org/x/net/html"
)

// additional (externally defined) metadata from CSV (e.g., leetcode id, topic)
type CustomProblem struct {
	ID   int
	Slug string
	Tag  string
}

// ProblemDetails represents parsed problem information from LeetCode HTML
type ProblemDetails struct {
	Description string    `json:"description"`
	Constraints []string  `json:"constraints"`
	Examples    []Example `json:"examples"`
	// TODO:
	// SimilarProblems []SimilarProblem `json:"similar_problems"`
}

// Example represents a problem example with input, output, and explanation
type Example struct {
	Input       string `json:"input"`
	Output      string `json:"output"`
	Explanation string `json:"explanation"`
}

// SimilarProblem represents a similar problem reference
type SimilarProblem struct {
	TitleSlug  string `json:"title_slug"`
	Title      string `json:"title"`
	Difficulty string `json:"difficulty"`
}

// ProblemMetadata is kept for backward compatibility but deprecated
// Use ProblemDetails instead
type ProblemMetadata struct {
	descriptions    string
	constraint      string
	testcases       string
	similarProblems []string
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

// GetLeetCodeProblemContents fetches the raw HTML content from LeetCode API
func GetLeetCodeProblemContents(id int) (*leetcodeapi.ProblemContent, error) {
	problems, err := leetcodeapi.GetAllProblems(id-1, 1) // transform id to zero-index

	if err != nil {
		return nil, err
	}

	problem := problems.Problems[0]

	content, err := leetcodeapi.GetProblemContentByTitleSlug(problem.TitleSlug)

	return &content, err
}

// ParseProblemContent parses HTML content into structured ProblemDetails
func ParseProblemContent(htmlContent string) (*ProblemDetails, error) {
	details := &ProblemDetails{
		Constraints: []string{},
		Examples:    []Example{},
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract description
	details.Description = extractDescription(doc)

	// Extract constraints
	details.Constraints = extractConstraints(doc)

	// Extract examples
	details.Examples = extractExamples(doc)

	return details, nil
}

// GetParsedProblemDetails fetches and parses problem content into structured format
func GetParsedProblemDetails(id int) (*ProblemDetails, error) {
	content, err := GetLeetCodeProblemContents(id)
	if err != nil {
		return nil, err
	}

	return ParseProblemContent(content.Content)
}

// extractDescription extracts the main problem description from HTML
func extractDescription(doc *html.Node) string {
	var description strings.Builder
	var inParagraph bool
	var skipNext bool

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "p" {
				inParagraph = true
			} else if n.Data == "strong" && getAttr(n, "class") == "example" {
				// Skip example headers
				skipNext = true
			} else if n.Data == "ul" && strings.Contains(getAttr(n, "class"), "constraint") {
				// Stop at constraints section
				return
			}
		}

		if n.Type == html.TextNode && inParagraph && !skipNext {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				if description.Len() > 0 {
					description.WriteString("\n\n")
				}
				description.WriteString(text)
			}
		}

		if n.Type == html.ElementNode && n.Data == "p" && n.FirstChild == nil {
			inParagraph = false
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}

		if n.Type == html.ElementNode && n.Data == "p" {
			inParagraph = false
		}
		if skipNext {
			skipNext = false
		}
	}

	traverse(doc)

	// Clean up the description - remove HTML entities and extra whitespace
	desc := description.String()
	desc = strings.ReplaceAll(desc, "&nbsp;", " ")
	desc = strings.ReplaceAll(desc, "&lt;", "<")
	desc = strings.ReplaceAll(desc, "&gt;", ">")
	desc = strings.ReplaceAll(desc, "&amp;", "&")
	desc = regexp.MustCompile(`\s+`).ReplaceAllString(desc, " ")
	desc = strings.TrimSpace(desc)

	return desc
}

// extractConstraints extracts constraints from the HTML
func extractConstraints(doc *html.Node) []string {
	var constraints []string
	var inConstraintsList bool

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "p" {
				text := getTextContent(n)
				if strings.Contains(text, "Constraints:") {
					inConstraintsList = true
				}
			} else if n.Data == "ul" && inConstraintsList {
				// Found constraints list
				for li := n.FirstChild; li != nil; li = li.NextSibling {
					if li.Type == html.ElementNode && li.Data == "li" {
						constraint := getTextContent(li)
						constraint = cleanHTMLText(constraint)
						if constraint != "" {
							constraints = append(constraints, constraint)
						}
					}
				}
				inConstraintsList = false
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	// Fallback: try regex pattern if structured parsing didn't work
	if len(constraints) == 0 {
		constraintPattern := regexp.MustCompile(`<li><code>([^<]+)</code></li>`)
		matches := constraintPattern.FindAllStringSubmatch(doc.Data, -1)
		for _, match := range matches {
			if len(match) > 1 {
				constraints = append(constraints, cleanHTMLText(match[1]))
			}
		}
	}

	return constraints
}

// extractExamples extracts examples from the HTML
func extractExamples(doc *html.Node) []Example {
	var examples []Example

	// First try: extract from <pre> tags which usually contain formatted examples
	var traversePre func(*html.Node)
	traversePre = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "pre" {
			text := getTextContent(n)
			text = cleanHTMLText(text)

			// Parse the pre-formatted text
			ex := parseExampleFromText(text)
			if ex.Input != "" || ex.Output != "" {
				examples = append(examples, ex)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traversePre(c)
		}
	}

	traversePre(doc)

	// Clean up examples - remove duplicate labels
	for i := range examples {
		examples[i].Input = cleanExampleField(examples[i].Input, "Input:")
		examples[i].Output = cleanExampleField(examples[i].Output, "Output:")
		examples[i].Explanation = cleanExampleField(examples[i].Explanation, "Explanation:")
	}

	// If no examples found, try regex fallback
	if len(examples) == 0 {
		examples = extractExamplesWithRegex(doc)
	}

	return examples
}

// parseExampleFromText parses example data from a text string
func parseExampleFromText(text string) Example {
	ex := Example{}

	// Split by common delimiters
	parts := strings.Split(text, "Output:")
	if len(parts) > 0 {
		// Extract input
		inputPart := strings.TrimSpace(parts[0])
		inputPart = strings.TrimPrefix(inputPart, "Input:")
		inputPart = strings.TrimSpace(inputPart)
		ex.Input = inputPart
	}

	if len(parts) > 1 {
		// Split output and explanation
		outputParts := strings.Split(parts[1], "Explanation:")
		if len(outputParts) > 0 {
			outputPart := strings.TrimSpace(outputParts[0])
			outputPart = strings.TrimPrefix(outputPart, "Output:")
			outputPart = strings.TrimSpace(outputPart)
			ex.Output = outputPart
		}
		if len(outputParts) > 1 {
			ex.Explanation = strings.TrimSpace(outputParts[1])
		}
	}

	return ex
}

// cleanExampleField removes duplicate labels and cleans up the field
func cleanExampleField(field, label string) string {
	field = strings.TrimSpace(field)
	if field == "" {
		return field
	}

	// Remove duplicate labels at the start
	lowerLabel := strings.ToLower(label)
	lowerField := strings.ToLower(field)

	for strings.HasPrefix(lowerField, lowerLabel) {
		field = strings.TrimSpace(strings.TrimPrefix(field, label))
		field = strings.TrimSpace(strings.TrimPrefix(field, strings.ToLower(label)))
		lowerField = strings.ToLower(field)
	}

	return strings.TrimSpace(field)
}

// extractExamplesWithRegex uses regex as fallback to extract examples
func extractExamplesWithRegex(doc *html.Node) []Example {
	var examples []Example
	htmlStr := renderNode(doc)

	// Pattern to match examples
	examplePattern := regexp.MustCompile(`<strong>Input:</strong>\s*([^<]+)<br/>\s*<strong>Output:</strong>\s*([^<]+)(?:<br/>\s*<strong>Explanation:</strong>\s*([^<]+))?`)
	matches := examplePattern.FindAllStringSubmatch(htmlStr, -1)

	for _, match := range matches {
		ex := Example{}
		if len(match) > 1 {
			ex.Input = cleanHTMLText(match[1])
		}
		if len(match) > 2 {
			ex.Output = cleanHTMLText(match[2])
		}
		if len(match) > 3 {
			ex.Explanation = cleanHTMLText(match[3])
		}
		if ex.Input != "" || ex.Output != "" {
			examples = append(examples, ex)
		}
	}

	return examples
}

// Helper functions

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func getTextContent(n *html.Node) string {
	var text strings.Builder
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(n)
	return text.String()
}

func cleanHTMLText(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func renderNode(n *html.Node) string {
	var buf strings.Builder
	html.Render(&buf, n)
	return buf.String()
}
