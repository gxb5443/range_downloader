package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

const url = "http://storage.googleapis.com/vimeo-test/work-at-vimeo.mp4"
const chunksize = 30000

func main() {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	outFile, err := os.Create("vimeofile.mp4")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer outFile.Close()
	if resp.Header["Accept-Ranges"][0] == "bytes" {
		log.Println("Ranges Supported!")
		log.Println("Content Size:", resp.Header["Content-Length"][0])
		//content_size, _ := strconv.ParseInt(resp.Header["Content-Length"][0], 0, 16)
		content_size, _ := strconv.Atoi(resp.Header["Content-Length"][0])
		for i := 0; i < content_size; {
			end_byte := i + chunksize
			go fetchChunk(int64(i), int64(end_byte), url, outFile)
			i = end_byte
		}
		return
	}
}

func fetchChunk(start_byte, end_byte int64, url string, file *os.File) {
	log.Println("Downloading byte ", start_byte)
	client := new(http.Client)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	req.Header.Set("Range: ", fmt.Sprintf("bytes=%d-%d", start_byte, end_byte-1))
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		return
	}
	file.WriteAt(body, start_byte)
	return
}
