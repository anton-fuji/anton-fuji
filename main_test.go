package main

import (
	"strings"
	"testing"
	"time"
)

func TestFeedSourceLimits(t *testing.T) {
	want := map[string]int{
		"Qiita":     4,
		"Zenn":      4,
		"Fuji Blog": 3,
	}

	for _, source := range feedSources {
		if source.Limit != want[source.Name] {
			t.Errorf("%s: limit = %d, want %d", source.Name, source.Limit, want[source.Name])
		}
		delete(want, source.Name)
	}
	if len(want) != 0 {
		t.Errorf("missing sources: %v", want)
	}
}

func TestSortAndLimitPosts(t *testing.T) {
	source := FeedSource{Name: "test"}
	posts := []Post{
		{Title: "oldest", PublishedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), Source: source},
		{Title: "newest", PublishedAt: time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC), Source: source},
		{Title: "middle", PublishedAt: time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC), Source: source},
	}

	got := sortAndLimitPosts(posts, 2)
	if len(got) != 2 {
		t.Fatalf("post count = %d, want 2", len(got))
	}
	if got[0].Title != "newest" || got[1].Title != "middle" {
		t.Errorf("posts = [%s, %s], want [newest, middle]", got[0].Title, got[1].Title)
	}
}

func TestUpdateWritingLogCount(t *testing.T) {
	content := "<summary>tail -n 9 ~/writing.log</summary>"
	got := updateWritingLogCount(content, 11)
	want := "<summary>tail -n 11 ~/writing.log</summary>"
	if got != want {
		t.Errorf("updated summary = %q, want %q", got, want)
	}
}

func TestRenderPosts(t *testing.T) {
	posts := []Post{{
		Title:  "Example",
		URL:    "https://example.com/post",
		Source: FeedSource{Icon: "![](img/example.png)"},
	}}

	got := renderPosts(posts)
	want := "- ![](img/example.png) [Example](https://example.com/post)"
	if strings.TrimSpace(got) != want {
		t.Errorf("rendered post = %q, want %q", got, want)
	}
}

func TestReplaceBetweenRejectsInvalidMarkers(t *testing.T) {
	_, err := replaceBetween("README", "<!--[START POSTS]-->", "<!--[END POSTS]-->", "posts")
	if err == nil {
		t.Fatal("replaceBetween() error = nil, want an error for missing markers")
	}
}
