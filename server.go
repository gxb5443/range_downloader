package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

const url = "http://storage.googleapis.com/vimeo-test/work-at-vimeo.mp4"
const filename = "work-at-vimeo.mp4"

func downloadFromLocal(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("File not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer file.Close()
	http.ServeContent(w, r, filename, time.Time{}, file)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	timeout := time.Duration(5) * time.Second
	transport := &http.Transport{
		ResponseHeaderTimeout: timeout,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, timeout)
		},
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Disposition", "attachment; filename=work-for-vimeo.mp4")
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
	io.Copy(w, resp.Body)
}

func rangeDownloader(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	if resp.Header["Accept-Ranges"][0] == "bytes" {
		//Start parallel downloads
		//Runs Get requests with
		log.Println("Ranges!")
		log.Println("Content Size:", resp.Header["Content-Length"][0])
		return
	}
}

/*
func fetchChunk(start_byte, end_byte int64, url string) ([]byte, error) {
	//Set range header as start-end/total
	client := new(http.Client)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return nil, err
}

func main() {
	http.HandleFunc("/", downloadHandler)
	http.HandleFunc("/range", rangeDownloader)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
*/
