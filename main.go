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

// QiitaのAtom構造体
type QiitaAtom struct {
	Entries []struct {
		Title     string `xml:"title"`
		Link      string `xml:"link"`
		Published string `xml:"published"`
	} `xml:"entry"`
}

// Post構造体
type Post struct {
	Title string
	Date  time.Time
	URL   string
}

// Qiitaのフィードを取得して解析
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
			URL:   entry.Link,
		})
	}

	// 日付順にソート
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	return posts, nil
}

func main() {
	const feedURL = "https://qiita.com/fujifuji1414/feed.atom"

	// Qiitaフィードを取得
	posts, err := fetchQiitaFeed(feedURL)
	if err != nil {
		log.Fatalf("フィード取得エラー: %v", err)
	}

	// 上位5件をMarkdown形式で整形
	distMD := "**Recent Qiita Articles**\n"
	for i, post := range posts {
		if i >= 5 {
			break
		}
		distMD += fmt.Sprintf("- [%s](%s)\n", post.Title, post.URL)
	}

	// README.mdの更新
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

// 指定したプレースホルダー間の文字列を置換
func replaceBetween(content, start, end, newContent string) string {
	startIdx := len(start) + len(content[:len(content)-len(start)])
	endIdx := len(content) - len(end)
	return content[:startIdx] + newContent + content[endIdx:]
}
