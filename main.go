package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const url = "http://storage.googleapis.com/vimeo-test/work-at-vimeo.mp4"
const chunksize = 100000
const threads = 1000

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
		calculated_chunksize := int(content_size / threads)
		log.Println("Chunk Size: ", int(calculated_chunksize))
		var end_byte int
		start_byte := 0
		chunks := 0
		for i := 0; i < threads; i++ {
			filename := "vimeo.part." + strconv.Itoa(i)
			wg.Add(1)
			//start_byte := i * int(calculated_chunksize)
			end_byte = start_byte + int(calculated_chunksize)
			log.Println("Dispatch ", start_byte, " to ", end_byte)
			go fetchChunk(int64(start_byte), int64(end_byte), url, filename, &wg)
			start_byte = end_byte
			chunks++
		}
		if end_byte < content_size {
			wg.Add(1)
			start_byte = end_byte
			end_byte = content_size
			filename := "vimeo.part." + strconv.Itoa(chunks)
			log.Println("Dispatch ", start_byte, " to ", end_byte)
			go fetchChunk(int64(start_byte), int64(end_byte), url, filename, &wg)
			chunks++
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
		defer outfile.Close()
		if err != nil {
			log.Fatal(err)
			return
		}
		for i := 0; i < chunks; i++ {
			filename := "vimeo.part." + strconv.Itoa(i)
			assembleChunk(filename, outfile)
		}
		//Veify file
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
		/*
			filesize := outfile.Stat().Size()
			blocks := uint64(math.Ceil(float64(filesize)/float64(8192)))
			hash := md5.New()
			for i:=uint64(0); i<blocks; i++ {
				blocksize := int(math.Min(filechunk, float64(filesize-int64(i*filechunk))))
				buf := make([] byte, blocksize)
				file.Read(buf)
				io.WriteString(hash, string(buf))   // append into the hash
			}
		*/
		log.Println("File Build Complete!")
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
	buffer := make([]byte, 4096)
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
	defer wg.Done()
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
	//outfile.Write(body)
	//io.Copy(outfile, body)
	io.Copy(outfile, res.Body)
	log.Println("Finished Downloading byte ", start_byte)
	return
}
