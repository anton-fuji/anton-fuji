package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"text/template"
	"time"
)

// 記事構造体
type Post struct {
	Title string
	Date  time.Time
	URL   string
}

// RSS構造体
type RSS struct {
	Items []struct {
		Title   string `xml:"title"`
		PubDate string `xml:"pubDate"`
		Link    string `xml:"link"`
	} `xml:"item"`
}

// RSSフィードを取得
func fetchRSS(url string) ([]Post, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("RSSデータ取得エラー: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("RSSデータ読み込みエラー: %w", err)
	}

	// デバッグ: RSSデータを確認
	fmt.Printf("取得したRSSデータ:\n%s\n", data)

	var feed RSS
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("RSS解析エラー: %w", err)
	}

	// 投稿データに変換
	var posts []Post
	for _, item := range feed.Items {
		date, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			fmt.Printf("日付解析エラー: %s (%v)\n", item.PubDate, err)
			continue
		}
		posts = append(posts, Post{
			Title: item.Title,
			Date:  date,
			URL:   item.Link,
		})
	}

	// デバッグ: パースされた投稿データ
	fmt.Printf("パースされた投稿データ:\n%+v\n", posts)

	return posts, nil
}

// READMEを更新
func updateReadme(posts []Post, templateText, readmePath string, limit int) error {
	// 投稿を日付順に並べ替え
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	if len(posts) > limit {
		posts = posts[:limit]
	}

	// テンプレートを適用
	tmpl, err := template.New("readme").Parse(templateText)
	if err != nil {
		return fmt.Errorf("テンプレート解析エラー: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, posts); err != nil {
		return fmt.Errorf("テンプレート生成エラー: %w", err)
	}
	markdown := buf.String()

	// デバッグ: 生成されたMarkdownを出力
	fmt.Printf("生成されたMarkdown:\n%s\n", markdown)

	// READMEを読み込み
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("README読み込みエラー: %w", err)
	}

	// プレースホルダー部分を置換
	re := regexp.MustCompile(`<!--\[START POSTS\]-->.*<!--\[END POSTS\]-->`)
	updated := re.ReplaceAllString(string(readme), fmt.Sprintf("<!--[START POSTS]-->\n%s\n<!--[END POSTS]-->", markdown))

	// デバッグ: 置換後のREADME内容
	fmt.Printf("置換後のREADME内容:\n%s\n", updated)

	// 書き込み
	return os.WriteFile(readmePath, []byte(updated), 0644)
}

func main() {
	const feedURL = "https://qiita.com/fujifuji1414/feed"
	const readmePath = "README.md"
	const maxPosts = 5
	const templateText = `**Qiita**
{{range . -}}
- ![](img/qiita.png) [{{.Title}}]({{.URL}})
{{end}}
`

	// RSSを取得
	posts, err := fetchRSS(feedURL)
	if err != nil {
		log.Fatalf("RSSの取得中にエラー: %v", err)
	}

	// READMEを更新
	if err := updateReadme(posts, templateText, readmePath, maxPosts); err != nil {
		log.Fatalf("READMEの更新中にエラー: %v", err)
	}

	fmt.Println("README.md が正常に更新されました！")
}
