package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/mmcdole/gofeed"
)

type Post struct {
	Title  string
	Date   string
	URL    string
	Source string
}

func fetchFeed(feedURL, source string) ([]Post, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return nil, fmt.Errorf("%s フィード取得エラー: %w", source, err)
	}

	var posts []Post
	for _, item := range feed.Items {
		date := item.Published // `gofeed` が自動的に日付を解析
		if item.UpdatedParsed != nil {
			date = item.UpdatedParsed.Format("2006-01-02 15:04:05")
		} else if item.PublishedParsed != nil {
			date = item.PublishedParsed.Format("2006-01-02 15:04:05")
		}

		posts = append(posts, Post{
			Title:  item.Title,
			Date:   date,
			URL:    item.Link,
			Source: source,
		})
	}

	fmt.Printf("✅ %s の記事数: %d\n", source, len(posts))
	return posts, nil
}

func main() {
	const (
		QiitaFeedURL = "https://qiita.com/fujifuji1414/feed.atom"
		ZennFeedURL  = "https://zenn.dev/fuuji/feed"
	)

	qiitaPosts, err := fetchFeed(QiitaFeedURL, "Qiita")
	if err != nil {
		log.Fatalf("Qiita フィード取得エラー: %v", err)
	}

	zennPosts, err := fetchFeed(ZennFeedURL, "Zenn")
	if err != nil {
		log.Fatalf("Zenn フィード取得エラー: %v", err)
	}

	posts := append(qiitaPosts, zennPosts...)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date > posts[j].Date
	})

	distMD := "**Recent Articles**\n"
	for i, post := range posts {
		if i >= 7 {
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

	if readmeContent != string(readme) {
		if err := os.WriteFile("README.md", []byte(readmeContent), 0644); err != nil {
			log.Fatalf("README書き込みエラー: %v", err)
		}
		fmt.Println("README.md が更新されました！")
	} else {
		fmt.Println("README.md に変更はありません。")
	}
}

// 指定されたプレースホルダーの間の内容を置き換える
func replaceBetween(content, start, end, newContent string) string {
	startIdx := indexOf(content, start)
	endIdx := indexOf(content, end)

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		fmt.Println("⚠️ プレースホルダーが見つからないため、README は変更されません。")
		return content
	}

	// `start` と `end` を含む範囲を置き換え
	result := content[:startIdx+len(start)] + "\n" + newContent + "\n" + content[endIdx:]

	// デバッグ: 変更があるかチェック
	if result == content {
		fmt.Println("⚠️ `replaceBetween` の結果に変更がありません！")
	} else {
		fmt.Println("✅ `replaceBetween` で変更が適用されました！")
	}

	return result
}

// 部分文字列の開始インデックスを取得
func indexOf(content, substr string) int {
	return strings.Index(content, substr)
}
