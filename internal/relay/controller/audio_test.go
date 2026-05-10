package controller

import "testing"

func TestGetTextFromJSON(t *testing.T) {
	got, err := getTextFromJSON([]byte(`{"text":"hello world"}`))
	if err != nil {
		t.Fatalf("getTextFromJSON() error = %v", err)
	}
	if got != "hello world" {
		t.Fatalf("getTextFromJSON() = %q, want %q", got, "hello world")
	}
}

func TestGetTextFromVerboseJSON(t *testing.T) {
	got, err := getTextFromVerboseJSON([]byte(`{"text":"verbose text","segments":[]}`))
	if err != nil {
		t.Fatalf("getTextFromVerboseJSON() error = %v", err)
	}
	if got != "verbose text" {
		t.Fatalf("getTextFromVerboseJSON() = %q, want %q", got, "verbose text")
	}
}

func TestGetTextFromText(t *testing.T) {
	got, err := getTextFromText([]byte("plain text\n"))
	if err != nil {
		t.Fatalf("getTextFromText() error = %v", err)
	}
	if got != "plain text" {
		t.Fatalf("getTextFromText() = %q, want %q", got, "plain text")
	}
}

func TestGetTextFromSRT(t *testing.T) {
	body := []byte("1\n00:00:00,000 --> 00:00:01,000\nhello\n\n2\n00:00:01,000 --> 00:00:02,000\nworld\n")
	got, err := getTextFromSRT(body)
	if err != nil {
		t.Fatalf("getTextFromSRT() error = %v", err)
	}
	if got != "helloworld" {
		t.Fatalf("getTextFromSRT() = %q, want %q", got, "helloworld")
	}
}

func TestGetTextFromVTT(t *testing.T) {
	body := []byte("WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nhello\n\n00:00:01.000 --> 00:00:02.000\nworld\n")
	got, err := getTextFromVTT(body)
	if err != nil {
		t.Fatalf("getTextFromVTT() error = %v", err)
	}
	if got != "helloworld" {
		t.Fatalf("getTextFromVTT() = %q, want %q", got, "helloworld")
	}
}
