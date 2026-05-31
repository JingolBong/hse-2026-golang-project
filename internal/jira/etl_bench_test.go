package connector

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
)

// BenchmarkTransformIssue measures the cost of mapping one heavy Jira issue
// (with optional fields, two users and a changelog) into the domain model —
// the parsing path the curator called out specifically.
func BenchmarkTransformIssue(b *testing.B) {
	ji := sampleJiraIssue()
	b.ReportAllocs()
	for b.Loop() {
		_, _, _ = transformIssue(ji, 77)
	}
}

func BenchmarkParseJiraTime(b *testing.B) {
	const s = "2025-01-02T15:04:05.000+0000"
	b.ReportAllocs()
	for b.Loop() {
		_, _ = parseJiraTime(s)
	}
}

func BenchmarkHashUsername(b *testing.B) {
	const name = "some.user@example.com"
	b.ReportAllocs()
	for b.Loop() {
		_ = hashUsername(name)
	}
}

// genJiraPayload marshals a realistic Jira search response of n issues once, so
// the benchmark measures only json.Unmarshal (the "heavy JSON parsing" cost).
func genJiraPayload(b *testing.B, n int) []byte {
	b.Helper()
	issues := make([]JiraIssue, 0, n)
	for i := 0; i < n; i++ {
		ji := sampleJiraIssue()
		ji.ID = strconv.Itoa(1000 + i)
		ji.Key = "ABC-" + strconv.Itoa(i)
		issues = append(issues, ji)
	}
	payload := JiraIssues{StartAt: 0, MaxResults: n, Total: n, Issues: issues}
	data, err := json.Marshal(payload)
	if err != nil {
		b.Fatalf("marshal payload: %v", err)
	}
	return data
}

func BenchmarkUnmarshalIssues(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		data := genJiraPayload(b, n)
		b.Run(fmt.Sprintf("issues=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				var out JiraIssues
				if err := json.Unmarshal(data, &out); err != nil {
					b.Fatalf("unmarshal: %v", err)
				}
			}
		})
	}
}
