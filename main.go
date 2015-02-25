package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const fileChunk = 8192

func main() {
	url := flag.String("url", "http://storage.googleapis.com/vimeo-test/work-at-vimeo.mp4", "URL for download")
	threads := flag.Int("threads", 100, "Number of threads to download with")
	flag.Parse()
	defer timeTrack(time.Now(), "Full download")
	resp, err := http.Get(*url)
	if err != nil {
		fmt.Println(err)
		return
	}
	content_size, _ := strconv.Atoi(resp.Header["Content-Length"][0])
	if resp.Header["Accept-Ranges"][0] == "bytes" {
		var wg sync.WaitGroup
		log.Println("Ranges Supported!")
		log.Println("Content Size:", resp.Header["Content-Length"][0])
		calculated_chunksize := int(content_size / *threads)
		log.Println("Chunk Size: ", int(calculated_chunksize))
		var end_byte int
		start_byte := 0
		chunks := 0
		for i := 0; i < *threads; i++ {
			filename := "vimeo.part." + strconv.Itoa(i)
			wg.Add(1)
			//start_byte := i * int(calculated_chunksize)
			end_byte = start_byte + int(calculated_chunksize)
			log.Println("Dispatch ", start_byte, " to ", end_byte)
			go fetchChunk(int64(start_byte), int64(end_byte), *url, filename, &wg)
			start_byte = end_byte
			chunks++
		}
		if end_byte < content_size {
			wg.Add(1)
			start_byte = end_byte
			end_byte = content_size
			filename := "vimeo.part." + strconv.Itoa(chunks)
			log.Println("Dispatch ", start_byte, " to ", end_byte)
			go fetchChunk(int64(start_byte), int64(end_byte), *url, filename, &wg)
			chunks++
		}
		wg.Wait()
		log.Println("Download Complete!")
		log.Println("Building File...")
		outfile, err := os.Create("vimeo_final.mp4")
		defer outfile.Close()
		if err != nil {
			log.Fatal(err)
			return
		}
		for i := 0; i < chunks; i++ {
			filename := "vimeo.part." + strconv.Itoa(i)
			assembleChunk(filename, outfile)
		}
		//Verify file size
		filestats, err := outfile.Stat()
		if err != nil {
			log.Fatal(err)
			return
		}
		actual_filesize := filestats.Size()
		if actual_filesize != int64(content_size) {
			log.Fatal("Actual Size: ", actual_filesize, "\nExpected: ", content_size)
			return
		}
		//Verify Md5
		if len(resp.Header["X-Goog-Hash"]) > 1 {
			content_md5, err := base64.StdEncoding.DecodeString(resp.Header["X-Goog-Hash"][1][4:])
			if err != nil {
				log.Fatal(err)
				return
			}
			if err != nil {
				log.Fatal(err)
				return
			}
			barray, _ := ioutil.ReadFile("vimeo_final.mp4")
			computed_hash := md5.Sum(barray)
			computed_slice := computed_hash[0:]
			if bytes.Compare(computed_slice, content_md5) != 0 {

				log.Fatal("WARNING: MD5 Sums don't match")
				return
			}
			log.Println("File MD5 Matches!")
		}
		log.Println("File Build Complete!")
		return
	}
	log.Println("Range Download unsupported")
	log.Println("Beginning full download...")
	fetchChunk(0, int64(content_size), *url, "no-range-vimeo.mp4", nil)
	log.Println("Download Complete")
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
	buffer := make([]byte, 32768)
	//buffer := make([]byte, 65536)
	for {
		n, err := creader.Read(buffer)
		if err != nil && err != io.EOF {
			log.Fatal(err)
			return
		}
		if err == io.EOF {
			break
		}
		if n == 0 {
			break
		}
		if _, err := cwriter.Write(buffer[:n]); err != nil {
			log.Fatal(err)
			return
		}
	}
	if err := cwriter.Flush(); err != nil {
		log.Fatal(err)
		return
	}
	os.Remove(filename)
}

func fetchChunk(start_byte, end_byte int64, url string, filename string, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	client := new(http.Client)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start_byte, end_byte-1))
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer res.Body.Close()
	//body, err := ioutil.ReadAll(res.Body)
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
	io.Copy(outfile, res.Body)
	log.Println("Finished Downloading byte ", start_byte)
	return
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
