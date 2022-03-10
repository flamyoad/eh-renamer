package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Key: "24738.jpg", Value: "001"
var pageNumberByName = make(map[string]string)

func main() {
	directoryArgs := flag.String("dir", "", "Directory of the doujin")
	linkArgs := flag.String("link", "", "Link of the site")
	flag.Parse()

	var dir string
	var link string

	if *directoryArgs == "" {
		path, err := os.Getwd()
		if err != nil {
			return
		}
		dir = path
	}

	link = *linkArgs
	link = "https://e-hentai.org/g/2159975/d5518a8e59/"
	dir = `C:\Users\user\Desktop\hitomi_downloaded\[Fukuro Daizi] [Fanbox] Fukuro (2159975)`

	fmt.Println(dir)
	fmt.Println(link)

	res, err := fetchHtml(link)
	defer res.Body.Close()
	if err != nil {
		log.Fatalf("Not able to fetch the website, error: %s\n", err)
	}

	largestPaginationNum, err := getLargestPaginationNumber(res)
	if err != nil {
		log.Fatalf("Failed when parsing for total number of pages, error: %s\n", err)
	}

	for i := 0; i < largestPaginationNum; i++ {
		paginationLink := link + "?p=" + strconv.Itoa(i)

		res, err := fetchHtml(paginationLink)
		defer res.Body.Close()
		if err != nil {
			continue
		}

		err = getImageNames(res)
		if err != nil {
			continue
		}
	}

	renameFiles(dir)
}

func fetchHtml(url string) (*http.Response, error) {
	fmt.Printf("Retrieving HTML from %s\n", url)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Fetches 2..3..4..5..6 except 1
func getLargestPaginationNumber(res *http.Response) (int, error) {
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return -1, err
	}

	largestNum := 1

	doc.Find("td").Each(func(i int, s *goquery.Selection) {
		onClick, _ := s.Attr("onclick")
		if onClick != "document.location=this.firstChild.href" {
			return
		}

		anchor := s.Find("a").Text()
		number, err := strconv.Atoi(anchor)
		if err != nil {
			return
		}
		if number > largestNum {
			largestNum = number
		}
	})
	return largestNum, nil
}

func getImageNames(res *http.Response) error {
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		alt, exists := s.Attr("alt")
		if !exists {
			return
		}

		_, err := strconv.Atoi(alt)
		if err != nil {
			return
		}

		title, exists := s.Attr("title")
		if !exists {
			return
		}

		arr := strings.SplitAfter(title, ":")
		formattedTitle := strings.TrimSpace(arr[1])
		pageNumberByName[formattedTitle] = alt
	})
	return nil
}

func renameFiles(pathArgs string) {
	if len(pageNumberByName) == 0 {
		log.Fatal("No images found")
	}

	path, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cannot get current path, err: %s\n", err)
	}

	if pathArgs != "" {
		path = pathArgs
	}

	fmt.Println(path)
	files, err := ioutil.ReadDir(path)
	for _, file := range files {
		if file.IsDir() {
			return
		}

		number, exists := pageNumberByName[file.Name()]
		if !exists {
			return
		}

		oldPath := filepath.Join(path, file.Name())
		fileExtension := filepath.Ext(oldPath)
		newPath := filepath.Join(path, number+fileExtension)

		err = os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Println("Error renaming: ", oldPath)
		} else {
			fmt.Println("Old path: ", oldPath)
			fmt.Println("New path: ", newPath)
			fmt.Println()
		}
	}
}
