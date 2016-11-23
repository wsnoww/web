
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
	"sort"
	"github.com/russross/blackfriday"
)

// todo: serve all static content
//	fix date parsing

type Post struct {
	Filename string
	Title string
	PublishDate time.Time
	Tags []string
	Body []byte
	Html template.HTML
}

// interface to Sort()
type PostSort []*Post

func (posts PostSort) Len() int {
	return len(posts)
}

func (posts PostSort) Swap(i, j int) {
	posts[i], posts[j] = posts[j], posts[i]
}

func (posts PostSort) Less(i, j int) bool {
	return posts[i].PublishDate.Before(posts[j].PublishDate)
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
	p.PublishDate, _ = parsedate(date)
	//fmt.Print(p.PublishDate)
	if err != nil {
		return nil, err
	}
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
		return nil, err
	}
	for _, fn := range filenames {
		p, err := loadpost(fn)
		if err != nil {
			fmt.Print(err)
			return nil, err
		}
		posts = append(posts, p)
	}
	sort.Sort(PostSort(posts))
	return posts, nil
}

func pagehandler(w http.ResponseWriter, r *http.Request) {
	posts, _ := getposts()
	t := template.New("index.html")
	t, _ = t.ParseFiles("index.html")
	t.Execute(w, posts)
}
func resumehandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "resume.pdf")
}

func main() {
	http.HandleFunc("/", pagehandler)
	http.HandleFunc("/resume.pdf", resumehandler)
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img/"))))
	http.ListenAndServe(":80", nil)
}

