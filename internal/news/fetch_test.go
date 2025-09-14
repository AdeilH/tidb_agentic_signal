package news

import (
	"testing"
	"time"
)

func TestLatest(t *testing.T) {
	// This is an integration test that requires network access
	// Skip in CI environments where network might be restricted
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	stories, err := Latest()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(stories) == 0 {
		t.Fatal("expected at least some stories")
	}

	// Check first story has required fields
	story := stories[0]
	if story.Title == "" {
		t.Error("expected story to have title")
	}
	if story.Source == "" {
		t.Error("expected story to have source")
	}

	// Check published time is valid
	if story.Published == "" {
		t.Error("expected story to have published time")
	} else {
		_, err := time.Parse(time.RFC3339, story.Published)
		if err != nil {
			t.Errorf("expected valid RFC3339 time, got: %s", story.Published)
		}
	}

	t.Logf("Successfully fetched %d stories", len(stories))
}
