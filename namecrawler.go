package main

import (
	"net/http"
//	"io/ioutil"
	"golang.org/x/net/html"
	"io"
	"runtime"
	"fmt"
)

const START = "https://ece.osu.edu/news/2015/10/texas-inst.-scholars-program"
var visited = make(map[string]bool)
var queue = make(chan string)

func main() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	go func() {
		queue <- START
	}()

	for url := range queue {
		go crawl(url)
	}

}

func crawl(url string) {
	fmt.Printf("Crawling %s\n", url)
	visited[url] = true;
	response, err := http.Get(url)

	if err != nil {
		fmt.Println("Error fetching the file: ", err)
		return
	}

	b := response.Body
	parse(b, url)
}

func parse(r io.Reader, url string) {
	z := html.NewTokenizer(r)

	Loop:
	for {

		page := z.Next()

		switch page {

		case html.ErrorToken:
			break Loop

		case html.StartTagToken:
			token := z.Token()

			if token.Data == "a" {
				for _, attr := range token.Attr {
					link := attr.Val
					if attr.Key == "href" && len(link) >= 4 && link[:4] == "http" {
						fmt.Printf("Found Link %s from url %s\n", link, url)
						if !visited[link] {
							go func() { queue <- link }()
						}
					}
				}
			}
		}
	}
}