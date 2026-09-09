package distribution

import "testing"

func TestOutputTextReadsRawResponsesContent(t *testing.T) {
	var out responseBody
	out.Output = append(out.Output, struct {
		Type    string `json:"type"`
		Status  string `json:"status"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}{
		Type:   "message",
		Status: "completed",
		Role:   "assistant",
		Content: []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{
			{Type: "output_text", Text: "1. Πρώτη επιλογή"},
		},
	})

	if got, want := outputText(out), "1. Πρώτη επιλογή"; got != want {
		t.Fatalf("outputText = %q, want %q", got, want)
	}
}
