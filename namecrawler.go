package main

import (
	"net/http"
	"net/url"
//	"io/ioutil"
	"golang.org/x/net/html"
	"io"
	"runtime"
	"fmt"
	"time"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const START = "https://ece.osu.edu/news/2015/10/texas-inst.-scholars-program"

var visited = make(map[string]bool)
var url_queue = make(chan string)
var notifications = make(chan bool)

var hostname = "ece.osu.edu"

type page struct {
	Id           bson.ObjectId `bson:"_id,omitempty"`
	Url          string
	ResponseTime float64
	FetchDate	 time.Time
}

func main() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	// Set the max threads to the number of cores
	var free_threads int = numCPU

	go func() {
		url_queue <- START
	}()

	for url := range url_queue {

		if free_threads > 0 {
			go crawl(url)
			free_threads--
		}

		go func() {
			if <- notifications {
				free_threads++
			}
		}()
	}

}

func crawl(url string) {
	defer func () {
		notifications <- true
	}()

	start := time.Now()
	response, err := http.Get(url)
	elapsed := time.Since(start).Seconds()

	if err != nil {
		fmt.Println("Error fetching the file: ", err)
		return
	}

	visited[url] = true
	savePage(url, elapsed)

	b := response.Body
	parse(b)
}

func parse(r io.Reader) {
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
						s, _ := url.Parse(link)

						if hostname != s.Host {
							continue
						}


						if !visited[link] {
							go func() { url_queue <- link }()
						}
					}
				}
			}
		}
	}
}

func savePage(url string, response float64) bool {
	db, err := mgo.Dial("mongodb://localhost")

	if err != nil {
		println("Eror Connecting to Mongo DB");
		return false
	} else {
		defer db.Close()
	}

	c := db.DB("test").C("pages")

	err = c.Insert(&page{Url: url, ResponseTime: response, FetchDate: time.Now()})

	if err != nil {
		fmt.Printf("Error Inserting to Mongo DB error: %s", err)
		return false
	}

	return true
}