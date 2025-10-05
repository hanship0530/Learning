package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	links := []string{
		"http://google.com",
		"http://facebook.com",
		"http://stackoverflow.com",
		"http://golang.org",
		"http://amazon.com",
		"http://www.naver.com",
	}

	c := make(chan string)

	for _, link := range links {
		// we only use go keyward in front of function calls
		go checkLink(link, c)
	}

	// for i := 0; i < len(links); i++ {
	// 	fmt.Println(<-c)
	// }

	for l := range c {
		// go checkLink(l, c)
		go func(link string) {
			time.Sleep(5 * time.Second)
			checkLink(link, c)
		}(l)
	}
}

func checkLink(link string, c chan string) {
	// This is not best approach
	// time.Sleep(5 * time.Second)
	_, err := http.Get(link)

	if err != nil {
		fmt.Println(link, "might be down!")
		// c <- "Might be down I think"
		c <- link
		return
	}

	fmt.Println(link, "is up!")
	// c <- "Yep It's up"
	c <- link
}
