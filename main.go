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

type QiitaAtom struct {
	Entries []struct {
		Title string `xml:"title"`
		Link  struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
		Published string `xml:"published"`
	} `xml:"entry"`
}

type Post struct {
	Title string
	Date  time.Time
	URL   string
}

func fetchQiitaFeed(feedURL string) ([]Post, error) {
	resp, err := http.Get(feedURL)
	if err != nil {
		return nil, fmt.Errorf("Qiitaフィード取得エラー: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("フィードデータ読み込みエラー: %w", err)
	}

	var feed QiitaAtom
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
			Title: entry.Title,
			Date:  date,
			URL:   entry.Link.Href,
		})
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	return posts, nil
}

func main() {
	const feedURL = "https://qiita.com/fujifuji1414/feed.atom"

	posts, err := fetchQiitaFeed(feedURL)
	if err != nil {
		log.Fatalf("フィード取得エラー: %v", err)
	}

	distMD := "**Recent Qiita Articles**\n"
	for i, post := range posts {
		if i >= 5 {
			break
		}
		distMD += fmt.Sprintf("- ![](img/qiita.png) [%s](%s)\n", post.Title, post.URL)
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
