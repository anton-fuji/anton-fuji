package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"time"
)

type QiitaFeed struct {
	Entries []struct {
		Title     string `xml:"title"`
		Published string `xml:"pubDate"`
		Link      string `xml:"link"`
	} `xml:"entry"`
}

type Post struct {
	Title string
	Date  time.Time
	Type  string
	URL   string
}

// QiitaのURL取得
func fetchURLContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func parseQiitaFeed(data []byte) ([]Post, error) {
	var feed QiitaFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	var posts []Post
	for _, entry := range feed.Entries {
		date, err := time.Parse(time.RFC3339, entry.Published)
		if err != nil {
			return nil, err
		}
		posts = append(posts, Post{
			Title: entry.Title,
			Date:  date,
			Type:  "qiita",
			URL:   entry.Link,
		})
	}
	return posts, nil
}

func updateReadme(posts []Post, readPath string) error {
	//日付でソート
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	markdown := ""
	for i, post := range posts {
		if i >= 5 {
			break
		}
		markdown += fmt.Sprintf("- ![](img/%s.png) [%s](%s)\n", post.Type, post.Title, post.URL)
	}
	// README.mdを読み込み、更新
	readme, err := os.ReadFile(readPath)
	if err != nil {
		return err
	}

	// <!--[START POSTS]--> と <!--[END POSTS]--> の間を置換
	re := regexp.MustCompile(`<!--\[START POSTS\]-->.*<!--\[END POSTS\]-->`)
	updated := re.ReplaceAllString(string(readme), fmt.Sprintf("<!--[START POSTS]-->\n%s<!--[END POSTS]-->", markdown))

	return os.WriteFile(readPath, []byte(updated), 0644)
}

func main() {
	//Qiitaのフィード取得
	qiitaData, err := fetchURLContent("https://qiita.com/fujifuji1414/feed.atom")
	if err != nil {
		fmt.Println("Qiitaフィードの取得中にエラーが発生しました:", err)
		return
	}

	qiitaPosts, err := parseQiitaFeed(qiitaData)
	if err != nil {
		fmt.Println("Qiitaフィードの解析中にエラーが発生しました:", err)
		return
	}

	if err := updateReadme(qiitaPosts, "README.md"); err != nil {
		fmt.Println("README.mdの更新中にエラーが発生しました:", err)
	} else {
		fmt.Println("README.mdが正常に更新されました！")
	}
}
