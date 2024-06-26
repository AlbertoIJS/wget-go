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
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	url := "https://doodleipsum.com/"
	downloadFile(url, &wg)
	wg.Wait()
}

func downloadFile(url string, wg *sync.WaitGroup) {
	fmt.Println("Descargando: ", url)

	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Println("Error: ", url, res.Status)
		wg.Done()
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

	var fileName string
	name := strings.ReplaceAll(newUrl, "/", "_")
	if filepath.Ext(name) == "" && len(ext) > 0 {
		fileName = "downloads/" + name
	} else if len(ext) > 0 {
		fileName = "downloads/" + name + ext[0]
	} else {
		fileName = "downloads/" + name
	}

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

	var waitGroup sync.WaitGroup

	urls, err := SearchURLsInFile(fileName)
	fmt.Println("URLs encontradas en "+url+":", urls)

	if err != nil {
		log.Println(err)
		return
	}

	waitGroup.Add(len(urls))
	for _, url = range urls {
		go downloadFile(url, &waitGroup)
	}

	waitGroup.Wait()
	wg.Done()
}

func SearchURLsInFile(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	rxRelaxed := xurls.Strict()

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
		return nil, err
	}

	return urls, nil
}
