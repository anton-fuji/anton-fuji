package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

const feedRequestTimeout = 15 * time.Second

type FeedSource struct {
	Name  string
	URL   string
	Icon  string
	Limit int
}

type Post struct {
	Title       string
	PublishedAt time.Time
	URL         string
	Source      FeedSource
}

var feedSources = []FeedSource{
	{
		Name:  "Qiita",
		URL:   "https://qiita.com/fujifuji1414/feed.atom",
		Icon:  "![](img/qiita.png)",
		Limit: 4,
	},
	{
		Name:  "Zenn",
		URL:   "https://zenn.dev/fuuji/feed",
		Icon:  "![](img/zenn.png)",
		Limit: 4,
	},
	{
		Name:  "Fuji Blog",
		URL:   "https://fuji-blog.netlify.app/rss.xml",
		Icon:  "![](img/fuji-blog.svg)",
		Limit: 3,
	},
}

func fetchFeed(source FeedSource) ([]Post, error) {
	fp := gofeed.NewParser()
	fp.Client = &http.Client{Timeout: feedRequestTimeout}
	feed, err := fp.ParseURL(source.URL)
	if err != nil {
		return nil, fmt.Errorf("%s フィード取得エラー: %w", source.Name, err)
	}

	posts := make([]Post, 0, len(feed.Items))
	for _, item := range feed.Items {
		publishedAt := item.PublishedParsed
		if publishedAt == nil {
			publishedAt = item.UpdatedParsed
		}
		if publishedAt == nil {
			log.Printf("%s の記事をスキップします（日付なし）: %s", source.Name, item.Title)
			continue
		}

		posts = append(posts, Post{
			Title:       item.Title,
			PublishedAt: *publishedAt,
			URL:         item.Link,
			Source:      source,
		})
	}
	posts = sortAndLimitPosts(posts, source.Limit)

	fmt.Printf("✅ %s の表示記事数: %d\n", source.Name, len(posts))
	return posts, nil
}

func main() {
	posts := make([]Post, 0)
	for _, source := range feedSources {
		feedPosts, err := fetchFeed(source)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, feedPosts...)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].PublishedAt.After(posts[j].PublishedAt)
	})
	readme, err := os.ReadFile("README.md")
	if err != nil {
		log.Fatalf("README読み込みエラー: %v", err)
	}

	readmeContent := replaceBetween(
		string(readme),
		"<!--[START POSTS]-->",
		"<!--[END POSTS]-->",
		renderPosts(posts),
	)
	readmeContent = updateWritingLogCount(readmeContent, len(posts))

	if readmeContent != string(readme) {
		if err := os.WriteFile("README.md", []byte(readmeContent), 0o644); err != nil {
			log.Fatalf("README書き込みエラー: %v", err)
		}
		fmt.Println("README.md が更新されました")
		return
	}

	fmt.Println("README.md に変更はありません。")
}

func renderPosts(posts []Post) string {
	var output strings.Builder
	for _, post := range posts {
		fmt.Fprintf(&output, "- %s [%s](%s)\n", post.Source.Icon, post.Title, post.URL)
	}
	return strings.TrimSuffix(output.String(), "\n")
}

func sortAndLimitPosts(posts []Post, limit int) []Post {
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].PublishedAt.After(posts[j].PublishedAt)
	})
	if len(posts) > limit {
		return posts[:limit]
	}
	return posts
}

func updateWritingLogCount(content string, count int) string {
	pattern := regexp.MustCompile(`(<summary>tail -n )\d+( ~/writing\.log</summary>)`)
	return pattern.ReplaceAllString(content, fmt.Sprintf("${1}%d${2}", count))
}

// replaceBetween replaces the content enclosed by the two marker comments.
func replaceBetween(content, start, end, newContent string) string {
	startIdx := strings.Index(content, start)
	endIdx := strings.Index(content, end)
	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		log.Printf("プレースホルダーが見つからないため、README は変更されません。")
		return content
	}

	return content[:startIdx+len(start)] + "\n" + newContent + "\n" + content[endIdx:]
}
