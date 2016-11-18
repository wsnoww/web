
package main

import (
	"strings"
	"io/ioutil"
	"net/http"
	"html/template"
	"time"
	"bufio"
	"os"
	"path/filepath"
	"fmt"
	"github.com/russross/blackfriday"
)

type Post struct {
	Filename string
	Title string
	Date time.Time
	Tags []string
	Body []byte
	Html template.HTML
}

func parsedate(d string) (time.Time, error) {
	t, err := time.Parse("January 2, 2006", d)
	if err == nil {
		return t, nil
	}
	return time.Now(), err
}

func parsetags(t string) []string {
	tags := strings.Split(t, ",")
	for i, tag := range tags {
		tag = strings.TrimSpace(tag)
		tag = strings.ToLower(tag)
		tags[i] = tag
	}
	return tags
}

func markdown(body []byte) []byte {
	return blackfriday.MarkdownCommon(body)
}

func loadpost(path string) (*Post, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	p := &Post{}
	p.Filename = path
	reader := bufio.NewReader(f)
	p.Title, err = reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	date, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	p.Date, _ = parsedate(date)
	//if err != nil {
		//return nil, err
	//}
	tags, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	p.Tags = parsetags(tags)
	p.Body, err = ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	p.Html = template.HTML(markdown(p.Body))
	return p, nil
}

func getposts() ([]*Post, error) {
	var posts []*Post
	filenames, err := filepath.Glob("posts/*.txt")
	if err != nil {
		fmt.Println("error reading filenames")
		return nil, err
	}
	for _, fn := range filenames {
		p, err := loadpost(fn)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func pagehandler(w http.ResponseWriter, r *http.Request) {
	posts, _ := getposts()
	t := template.New("index.html")
	t, _ = t.ParseFiles("index.html")
	t.Execute(w, posts)
}
func avatarhandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "avatar.jpg")
}

func main() {
	http.HandleFunc("/avatar.jpg", avatarhandler)
	http.HandleFunc("/", pagehandler)
	http.ListenAndServe(":8080", nil)
}

