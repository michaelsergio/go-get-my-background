package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/jzelinskie/reddit"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func info() {
	sub, err := reddit.AboutSubreddit("spaceporn")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(sub)
}

func getMimeFromUrl(url string) (string, error) {
	extPos := strings.LastIndex(url, ".")
	if extPos == -1 {
		return url, errors.New("Not a real mimetype")
	}
	return url[extPos+1:], nil

}

func getFilename(url string) (string, error) {
	const folder = "images"
	hash := md5.New()
	io.WriteString(hash, url)
	filename := hex.EncodeToString(hash.Sum(nil))
	mime, err := getMimeFromUrl(url)
	if err != nil {
		fmt.Printf("\033[31m%v\033[0m\n", err)
		return "", errors.New("Bad filetype")
	}
	filepath := fmt.Sprintf("%v/%v.%v", folder, filename, mime)
	return filepath, nil
}

func downloadFile(from, to string) {
	out, err := os.Create(to)
	if err != nil {
		return
	}
	fmt.Printf("\033[33mFrom:%v To:%v\033[0m\n", from, to)

	resp, err := http.Get(from)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	io.Copy(out, resp.Body)
	out.Close()
}

func isWhitelistedSite(url string) bool {
	whitelist := []string{"imgur"}
	for _, entry := range whitelist {
		if strings.Contains(url, entry) {
			return true
		}
	}
	return false
}

func listStories() {
	downloads := make(chan string)
	done := make(chan int)
	headlines, err := reddit.SubredditHeadlines("spaceporn")
	if err != nil {
		fmt.Println("OHOH! %v", err)
		return
	}
	for _, headline := range headlines {
		fmt.Println(headline.Title, headline.URL)
		if isWhitelistedSite(headline.URL) {
			filepath, err := getFilename(headline.URL)
			if err != nil {
				fmt.Println("error with filename")
				return
			}

			go func(url, path string) {
				//fmt.Printf("\033[32mFrom:%v\033[0m\n", filepath)
				downloadFile(url, path)
				downloads <- url // Download Finished
			}(headline.URL, filepath)
		}
	}
	time.AfterFunc(5*time.Second, func() {
		done <- 1
	})

	for {
		select {
		case url := <-downloads:
			fmt.Printf("Downloaded: %v\n", url)
		case <-done:
			return
		}
	}
}

func main() {
	info()
	listStories()
	fmt.Println("Done")
	//defaults write com.apple.desktop Background '{default = {ImageFilePath = "/Library/Desktop Pictures/Black & White/Lightning.jpg"; };}'

}
