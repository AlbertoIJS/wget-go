package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"mime"
	"mvdan.cc/xurls/v2"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
)

func main() {
	file, err := os.Open("index.html")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	rxRelaxed := xurls.Relaxed()
	var wg sync.WaitGroup
	var urls []string

	for scanner.Scan() {
		line := scanner.Text()
		list := rxRelaxed.FindAllString(line, -1)

		for _, url := range list {
			if slices.Contains(urls, url) == false {
				urls = append(urls, url)
			}
		}
	}

	if err = scanner.Err(); err != nil {
		log.Println(err)
		return
	}

	if len(urls) > 0 {
		for _, url := range urls {
			wg.Add(1)
			go downloadFile(url, &wg)
		}
	} else {
		log.Fatal("No se encontraron URLs")
	}

	wg.Wait()
}

func downloadFile(url string, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("Descargando: ", url)

	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Println("Error: ", res.Status)
		return
	}

	// Create a unique filename for each URL
	contentType := res.Header.Get("Content-Type")
	ext, err := mime.ExtensionsByType(contentType)
	if err != nil {
		log.Println(err)
		return
	}

	var newUrl string
	if strings.Contains(url, "http://") {
		newUrl = strings.Replace(url, "http://", "", -1)
	} else if strings.Contains(url, "https://") {
		newUrl = strings.Replace(url, "https://", "", -1)
	}

	fileName := "downloads/" + newUrl + ext[0]
	file, err := os.Create(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	// Open the downloaded file and scan it for URLs
	downloadedFile, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer downloadedFile.Close()
}

//func downloadFile(url string, wg *sync.WaitGroup) {
//	defer wg.Done()
//
//	fmt.Println("Descargando ", url)
//
//	file, err := os.Create("downloads/" + url)
//	if err != nil {
//		log.Println(err)
//		return
//	}
//	defer file.Close()
//
//	res, err := http.Get(url)
//	if err != nil {
//		log.Println(err)
//		return
//	}
//	defer res.Body.Close()
//
//	if res.StatusCode != http.StatusOK {
//		log.Println("Error: ", res.Status)
//		return
//	}
//
//	_, err = io.Copy(file, res.Body)
//	if err != nil {
//		log.Println(err)
//		return
//	}
//}
