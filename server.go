package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

const url = "https://boards.greenhouse.io/vimeo/jobs/42976#.VLaurYrF9TM"

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

func main() {
	http.HandleFunc("/", downloadHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
