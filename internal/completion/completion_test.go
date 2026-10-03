package completion

import (
	"bytes"
	"testing"
)

func TestDecodeRecords(t *testing.T) {
	input := []byte("git\x00commit\x00record changes\x00 \x00git\x00checkout\x00switch branch\x00 \x00")
	got, err := DecodeRecords(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d candidates", len(got))
	}
	if got[0].Value != "commit" || got[0].Description != "record changes" || got[0].Suffix != " " {
		t.Fatalf("unexpected first candidate: %+v", got[0])
	}
}

func TestDecodeRecordsRejectsTruncatedInput(t *testing.T) {
	_, err := DecodeRecords([]byte("git\x00commit\x00"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFilterPreservesOrderAndDescriptions(t *testing.T) {
	all := []Candidate{
		{Value: "checkout"},
		{Value: "commit", Description: "record changes"},
		{Value: "checkout", Description: "switch branch"},
		{Value: "add", Description: "stage files"},
	}
	got := Filter(all, "c")
	if len(got) != 2 || got[0].Value != "checkout" || got[1].Value != "commit" {
		t.Fatalf("unexpected candidates: %+v", got)
	}
	if got[0].Description != "switch branch" {
		t.Fatalf("duplicate did not enrich description: %+v", got[0])
	}
}

func TestFormatProtocol(t *testing.T) {
	got := EncodeRecords([]Candidate{{Group: "git", Value: "commit", Description: "record changes", Suffix: " "}})
	want := []byte("git\x00commit\x00record changes\x00 \x00")
	if !bytes.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}
