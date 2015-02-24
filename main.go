package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const url = "http://storage.googleapis.com/vimeo-test/work-at-vimeo.mp4"
const chunksize = 100000
const threads = 30

func main() {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	if resp.Header["Accept-Ranges"][0] == "bytes" {
		var wg sync.WaitGroup
		log.Println("Ranges Supported!")
		log.Println("Content Size:", resp.Header["Content-Length"][0])
		content_size, _ := strconv.Atoi(resp.Header["Content-Length"][0])
		//calculated_chunksize := math.Ceil(float64(content_size) / threads)
		calculated_chunksize := content_size / threads
		log.Println("Chunk Size: ", int(calculated_chunksize))
		var end_byte int
		start_byte := 0
		for i := 0; i < threads; i++ {
			filename := "vimeo.part." + strconv.Itoa(i)
			wg.Add(1)
			//start_byte := i * int(calculated_chunksize)
			end_byte = start_byte + int(calculated_chunksize)
			log.Println("Dispatch ", start_byte, " to ", end_byte)
			go fetchChunk(int64(start_byte), int64(end_byte), url, filename, &wg)
			start_byte = end_byte
		}
		if end_byte < content_size {
			wg.Add(1)
			start_byte = end_byte
			end_byte = content_size
			filename := "vimeo.part." + strconv.Itoa(threads)
			log.Println("Dispatch ", start_byte, " to ", end_byte)
			go fetchChunk(int64(start_byte), int64(end_byte), url, filename, &wg)
		}
		/*
			for i := 0; i < content_size; {
				wg.Add(1)
				end_byte := i + chunksize
				go fetchChunk(int64(i), int64(end_byte), url, outFile, &wg)
				i = end_byte
			}
		*/
		wg.Wait()
		log.Println("Download Complete!")
		log.Println("Building File...")
		outfile, err := os.Create("vimeo_final.mp4")
		if err != nil {
			log.Fatal(err)
			return
		}
		for i := 0; i <= threads; i++ {
			filename := "vimeo.part." + strconv.Itoa(i)
			assembleChunk(filename, outfile)
		}
		return
	}
}

func assembleChunk(filename string, outfile *os.File) {
	chunkFile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer chunkFile.Close()
	creader := bufio.NewReader(chunkFile)
	cwriter := bufio.NewWriter(outfile)
	buffer := make([]byte, 2048)
	for {
		n, err := creader.Read(buffer)
		if err != nil && err != io.EOF {
			log.Fatal(err)
			return
		}
		if n == 0 {
			break
		}
		if _, err := cwriter.Write(buffer); err != nil {
			log.Fatal(err)
			return
		}
	}
	os.Remove(filename)
}

func fetchChunk(start_byte, end_byte int64, url string, filename string, wg *sync.WaitGroup) {
	defer wg.Done()
	//log.Println("Downloading byte ", start_byte)
	client := new(http.Client)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start_byte, end_byte-1))
	log.Println(req.Header)
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
	outfile, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer outfile.Close()
	outfile.Write(body)
	//file.WriteAt(body, start_byte)
	log.Println("Finished Downloading byte ", start_byte)
	return
}
