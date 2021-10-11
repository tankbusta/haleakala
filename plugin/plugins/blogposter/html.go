package blogposter

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var ErrNoBlogs = errors.New("no blogs found")

const (
	BaseURL     = "https://www.mandiant.com"
	BlogPostURL = "https://www.mandiant.com/resources?f[0]=layout:article_blog"
)

type BlogPost struct {
	URL        string
	Title      string
	TimeToRead int
	PostedOn   time.Time
	IndexedOn  time.Time
}

type BlogPosts []BlogPost

// Len implements sort.Interface
func (p BlogPosts) Len() int {
	return len(p)
}

// Less implements sort.Interface
func (p BlogPosts) Less(i, j int) bool {
	return p[i].PostedOn.After(p[j].PostedOn)
}

// Swap implements sort.Interface
func (p BlogPosts) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func getSpanChildData(klass string, parent *html.Node) (string, error) {
	node := htmlquery.FindOne(parent, fmt.Sprintf("//small[@class=\"info\"]/span[@class=\"%s\"]", klass))
	if node != nil && node.FirstChild != nil {
		return strings.TrimSpace(node.FirstChild.Data), nil
	}

	return "", fmt.Errorf("%s not found", klass)
}

func getTimeToRead(parent *html.Node) (int, error) {
	timeRaw, err := getSpanChildData("time", parent)
	if err == nil {
		parts := strings.SplitN(timeRaw, " ", 2)
		if len(parts) == 2 {
			return strconv.Atoi(parts[0])
		}
	}

	return 0, fmt.Errorf("estimated time not found")
}

// GetMandiantBlogs parses the HTML contents and returns a list of available blog posts
//
// Someone should really talk to the marketing department about implementing an RSS feed...
func GetMandiantBlogs(body *html.Node) (BlogPosts, error) {
	list := htmlquery.Find(body, "//div[@class=\"cols cols-3\"]/a")
	if list == nil {
		return nil, ErrNoBlogs
	}

	blogs := make(BlogPosts, len(list))
	for i, elem := range list {
		blogTitle := htmlquery.SelectAttr(elem, "title")
		if blogTitle == "" {
			return nil, fmt.Errorf("missing blog post title at /a[%d]", i)
		}

		timeToRead, err := getTimeToRead(elem)
		if err != nil {
			return nil, fmt.Errorf("error reading time estimate at /a[%d]", i)
		}

		postedOnRaw, err := getSpanChildData("date string", elem)
		if err != nil {
			return nil, fmt.Errorf("error reading blog created on at /a[%d]", i)
		}

		postedOn, err := time.Parse("Jan 02, 2006", postedOnRaw)
		if err != nil {
			return nil, fmt.Errorf("error converting postedOn /a[%d]", i)
		}

		blogs[i] = BlogPost{
			Title:      blogTitle,
			URL:        BaseURL + htmlquery.SelectAttr(elem, "href"),
			TimeToRead: timeToRead,
			PostedOn:   postedOn,
			IndexedOn:  time.Now().UTC(),
		}
	}

	// Sort with newest blogs first
	sort.Sort(blogs)

	return blogs, nil
}

func getBlogContent() (*html.Node, error) {
	resp, err := http.DefaultClient.Get(BlogPostURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	return htmlquery.Parse(resp.Body)
}
