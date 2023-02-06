package main

import (
	"regexp"
	"testing"
)

func TestParser(t *testing.T) {
	links, images, err := parseHTML(html)
	if err != nil {
		t.Fatal(err)
	}

	if len(images) != 1 || images[0] != "/img/0.png" {
		t.Fatalf("Invalid images: %#v", images)
	}

	if len(links) != 2 || links[0] != "/html/1.html" {
		t.Fatalf("Invalid links: %#v", links)
	}
}

func BenchmarkRegExp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		links, images = parseHTMLRe(benchHtml)
	}
}

func BenchmarkParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		links, images, _ = parseHTML(benchHtml)
	}
}

// RegExp parser
var (
	reImg  = regexp.MustCompile(`<img src="(/[^"]+.png)"/>`)
	reLink = regexp.MustCompile(`<a.*href="([^"]+)".*>`)
)

func parseHTMLRe(s string) ([]string, []string) {
	images := make([]string, 0, 1)
	links := make([]string, 0, 2)

	for _, link := range reLink.FindAllStringSubmatch(s, -1) {
		links = append(links, link[0])
	}

	for _, image := range reImg.FindAllStringSubmatch(s, -1) {
		images = append(images, image[0])
	}

	return images, links
}

var html = `<!DOCTYPE html>
	<html lang="en">
	<head>
	<title>0</title>
	</head>
	<body>
		<!-- Multiline comment with images and links
			<img src="/img/0.png"/>
			<p><a href="/html/1.html">Left</a></p>
			<p><a href="/html/1.html">Right</a></p>
		-->

		<h1>0</h1>
		<img src=""/>
		<iMG src="/img/0.png"/>
		
			<p><a href="">Bad link</a></p>
			<p><A href="/html/1.html">Left</a></p>
			<p><a href="/html/1.html">Right</a></p>
		
	</body>
	</html>`

var benchHtml = `
<!DOCTYPE html>
	<html lang="en">
	<head>
	<title>0</title>
	</head>
	<body>
		<h1>0</h1>
		<img src="/img/0.png"/>
		
			<p><a href="/html/1.html">Left</a></p>
			<p><a href="/html/1.html">Right</a></p>
		
	</body>
	</html>`

var (
	images []string
	links  []string
)
