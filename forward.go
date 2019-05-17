package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

func handleURL(w http.ResponseWriter, r *http.Request) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	type Config map[string]string
	urls := Config{
		"test": "https://jsonplaceholder.typicode.com/todos/1",
	}
	filename := "config.properties"
	if len(filename) == 0 {
		log.Fatal("No config")
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Can't open config")
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')

		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				urls[key] = value
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Error getting config")
		}
	}
	src := r.RemoteAddr
	dst := r.Host + r.RequestURI
	path := strings.Split(r.RequestURI, "/")
	fwdDst, ok := urls[path[1]]
	if !ok {
		log.Println("Entry not found, please check config.properties")
		return
	}
	method := r.Method
	fwdURL, err := url.Parse(fwdDst)
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(fwdURL)
	newPath := strings.SplitAfter(r.RequestURI, path[1])
	newHost := strings.SplitAfter(fwdDst, "//")
	r.URL.Path = newPath[1]
	r.URL.Host = newHost[1]
	r.URL.Scheme = "https"
	r.Host = newHost[1]
	log.Printf("%s forwarding %s from %s to %s%s", src, method, dst, newHost[1], newPath[1])
	proxy.ServeHTTP(w, r)
}

func main() {
	rtr := mux.NewRouter()
	rtr.PathPrefix("/").HandlerFunc(handleURL)
	http.Handle("/", rtr)
	log.Fatal(http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", nil))
}
