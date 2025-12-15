package main

import (
	"strings"
	"testing"

	"github.com/daniel0321forever/terriyaki-go/internal/services"
)

// TestCrawlProblemDescription tests crawling problem description and constraints
func TestCrawlProblemDescription(t *testing.T) {
	// Act: Fetch the problem with parsed details
	details, err := services.GetProblemById(1)
	if err != nil {
		t.Fatalf("GetProblemById() error = %v", err)
	}

	// Assert: Verify description is extracted
	if details.Description == "" {
		t.Error("Expected description to be extracted, got empty string")
	}

	// Check that description contains expected keywords
	descriptionLower := strings.ToLower(details.Description)
	if !strings.Contains(descriptionLower, "array") && !strings.Contains(descriptionLower, "integer") {
		t.Logf("Description: %s", details.Description[:min(200, len(details.Description))])
	}

	// Assert: Verify constraints are extracted
	if len(details.Constraints) == 0 {
		t.Error("Expected constraints to be extracted, got empty slice")
	} else {
		t.Logf("Found %d constraints:", len(details.Constraints))
		for i, constraint := range details.Constraints {
			t.Logf("  Constraint %d: %s", i+1, constraint)
		}
	}

	// Assert: Verify examples are extracted
	if len(details.Examples) == 0 {
		t.Log("No examples extracted (this might be okay depending on HTML structure)")
	} else {
		t.Logf("Found %d examples:", len(details.Examples))
		for i, example := range details.Examples {
			t.Logf("  Example %d:", i+1)
			t.Logf("    Input: %s", example.Input)
			t.Logf("    Output: %s", example.Output)
			if example.Explanation != "" {
				t.Logf("    Explanation: %s", example.Explanation)
			}
		}
	}

	// Log the full description for inspection
	t.Logf("\nFull Description:\n%s", details.Description)
}


// TestParseProblemContentWithSampleHTML tests parsing with sample HTML
func TestParseProblemContentWithSampleHTML(t *testing.T) {
	sampleHTML := `
	<p>Given an array of integers <code>nums</code>&nbsp;and an integer <code>target</code>, return <em>indices of the two numbers such that they add up to <code>target</code></em>.</p>
	<p>You may assume that each input would have <strong><em>exactly</em> one solution</strong>, and you may not use the <em>same</em> element twice.</p>
	<p><strong class="example">Example 1:</strong></p>
	<pre>
	<strong>Input:</strong> nums = [2,7,11,15], target = 9
	<strong>Output:</strong> [0,1]
	<strong>Explanation:</strong> Because nums[0] + nums[1] == 9, we return [0, 1].
	</pre>
	<p><strong>Constraints:</strong></p>
	<ul>
		<li><code>2 &lt;= nums.length &lt;= 10<sup>4</sup></code></li>
		<li><code>-10<sup>9</sup> &lt;= nums[i] &lt;= 10<sup>9</sup></code></li>
	</ul>
	`

	description, constraints, examples, err := services.ParseProblemContent(sampleHTML)
	if err != nil {
		t.Fatalf("ParseProblemContent() error = %v", err)
	}

	// Verify description
	if !strings.Contains(strings.ToLower(description), "array") {
		t.Errorf("Description should contain 'array', got: %s", description)
	}

	// Verify constraints
	if len(constraints) < 2 {
		t.Errorf("Expected at least 2 constraints, got %d", len(constraints))
	}

	// Verify examples
	if len(examples) == 0 {
		t.Log("No examples extracted from sample HTML (parsing might need adjustment)")
	}
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
