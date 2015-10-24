package main

import (
	"fmt"
	"net/http"
	"os"
//	"io/ioutil"
	"golang.org/x/net/html"
	"io"
)

const START = "http://www.yahoo.com"

func main() {

	crawl(START)

}

func crawl(url string) {
	response, err := http.Get(url)

	if err != nil {
		fmt.Println("Error fetching the file: ", err)
		os.Exit(1)
	}

	b := response.Body
	links := parse(b)

	for _, l := range links {
		println(l)
	}
}

func parse(r io.Reader) []string {
	z := html.NewTokenizer(r)
	links := make([]string, 0)

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
					if attr.Key == "href" && len(attr.Val) >= 4 && attr.Val[:4] == "http" {
						links = append(links, attr.Val)
					}
				}

			}


		}

	}

	return links
}