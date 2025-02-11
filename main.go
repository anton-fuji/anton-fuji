package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

type AtomFeed struct {
	Entries []struct {
		Title string `xml:"title"`
		Link  struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
		Published string `xml:"published"`
	} `xml:"entry"`
}

type Post struct {
	Title  string
	Date   time.Time
	URL    string
	Source string
}

func fetchFeed(feedURL, source string) ([]Post, error) {
	resp, err := http.Get(feedURL)
	if err != nil {
		return nil, fmt.Errorf("%s フィード取得エラー: %w", source, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("フィードデータ読み込みエラー: %w", err)
	}

	var feed AtomFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("フィード解析エラー: %w", err)
	}

	var posts []Post
	for _, entry := range feed.Entries {
		date, err := time.Parse(time.RFC3339, entry.Published)
		if err != nil {
			continue
		}
		posts = append(posts, Post{
			Title:  entry.Title,
			Date:   date,
			URL:    entry.Link.Href,
			Source: source,
		})
	}

	// sort.Slice(posts, func(i, j int) bool {
	// 	return posts[i].Date.After(posts[j].Date)
	// })
	return posts, nil
}

func main() {
	const (
		QiitaFeedURL = "https://qiita.com/fujifuji1414/feed.atom"
		ZennFeedURL  = "https://zenn.dev/fuuji/feed"
	)

	qiitaPosts, err := fetchFeed(QiitaFeedURL, "Qiita")
	if err != nil {
		log.Fatalf("フィード取得エラー: %v", err)
	}

	zennPosts, err := fetchFeed(ZennFeedURL, "Zenn")
	if err != nil {
		log.Fatalf("フィード取得エラー: %v", err)
	}

	posts := append(qiitaPosts, zennPosts...)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	distMD := "**Recent Articles**\n"
	for i, post := range posts {
		if i >= 5 {
			break
		}
		icon := "qiita.png"
		if post.Source == "Zenn" {
			icon = "zenn.png"
		}
		distMD += fmt.Sprintf("- ![](img/%s) [%s](%s)\n", icon, post.Title, post.URL)
	}

	readme, err := os.ReadFile("README.md")
	if err != nil {
		log.Fatalf("README読み込みエラー: %v", err)
	}

	newReadme := `<!--[START POSTS]-->` + "\n" + distMD + `<!--[END POSTS]-->`
	readmeContent := string(readme)
	readmeContent = replaceBetween(readmeContent, "<!--[START POSTS]-->", "<!--[END POSTS]-->", newReadme)

	if err := os.WriteFile("README.md", []byte(readmeContent), 0644); err != nil {
		log.Fatalf("README書き込みエラー: %v", err)
	}

	fmt.Println("README.md が更新されました！")
}

func replaceBetween(content, start, end, newContent string) string {
	startIdx := indexOf(content, start) + len(start)
	endIdx := indexOf(content, end)
	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return content // プレースホルダーが見つからない場合はそのまま返す
	}
	return content[:startIdx] + "\n" + newContent + "\n" + content[endIdx:]
}

// 部分文字列の開始インデックスを取得
func indexOf(content, substr string) int {
	return findIndex(content, substr)
}

// 部分文字列を検索して最初に見つかった位置を返す
func findIndex(content, substr string) int {
	idx := -1
	for i := 0; i+len(substr) <= len(content); i++ {
		if content[i:i+len(substr)] == substr {
			idx = i
			break
		}
	}
	return idx
}
