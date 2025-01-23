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

	"gopkg.in/yaml.v2"
)

type Config struct {
	URLs     []string `yaml:"urls"`
	Template string   `yaml:"template"`
	Limit    int      `yaml:"limit"`
}

type QiitaFeed struct {
	Items []struct {
		Title   string `xml:"title"`
		PubDate string `xml:"pubDate"`
		Link    string `xml:"link"`
	} `xml:"item"`
}

type Post struct {
	Title string
	Date  time.Time
	URL   string
}

// 設定ファイルを読み込む
func loadConfig(filename string) ([]Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var configs []Config
	if err := yaml.Unmarshal(data, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

// RSSフィードを取得
func fetchRSS(url string) ([]Post, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed QiitaFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	var posts []Post
	for _, item := range feed.Items {
		date, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			continue
		}
		posts = append(posts, Post{Title: item.Title, Date: date, URL: item.Link})
	}
	return posts, nil
}

// READMEを更新
func updateReadme(posts []Post, templateText, readmePath string, limit int) error {
	// 最新の投稿のみ取得
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	if len(posts) > limit {
		posts = posts[:limit]
	}

	// テンプレートを適用
	tmpl, err := template.New("readme").Parse(templateText)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, posts); err != nil {
		return err
	}
	markdown := buf.String() // バッファから文字列を取得

	// READMEを読み込み
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}

	// <!--[START POSTS]--> と <!--[END POSTS]--> の間を置換
	re := regexp.MustCompile(`<!--\[START POSTS\]-->.*<!--\[END POSTS]-->`)
	updated := re.ReplaceAllString(string(readme), fmt.Sprintf("<!--[START POSTS]-->\n%s\n<!--[END POSTS]-->", markdown))

	// 書き込み
	return os.WriteFile(readmePath, []byte(updated), 0644)
}

func main() {
	// 設定ファイルを読み込み
	configs, err := loadConfig("config.yml")
	if err != nil {
		log.Fatalf("設定ファイルの読み込みに失敗しました: %v", err)
	}

	for _, config := range configs {
		var allPosts []Post

		// 各URLのデータを取得
		for _, url := range config.URLs {
			posts, err := fetchRSS(url)
			if err != nil {
				log.Printf("RSSフィードの取得中にエラーが発生しました: %v", err)
				continue
			}
			allPosts = append(allPosts, posts...)
		}

		// READMEを更新
		if err := updateReadme(allPosts, config.Template, "README.md", config.Limit); err != nil {
			log.Fatalf("READMEの更新中にエラーが発生しました: %v", err)
		}
	}

	fmt.Println("READMEが正常に更新されました！")
}
