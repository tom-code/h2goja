package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var vm *VM

func startServer() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		log.Println("got request " + uri)
		indata, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err == nil {
			log.Println(string(indata))
		} else {
			log.Println(err)
		}
		p := Param {
			Url: r.RequestURI,
			Data: indata,
			Status: 500,
		}
		vm.Call(&p)
		w.WriteHeader(p.Status)
		if len(p.DataOut) > 0 {
			w.Write(p.DataOut)
		}
	})
	h2s := &http2.Server{}
	h1s := &http.Server{
		Addr:    ":8080",
		Handler: h2c.NewHandler(handler, h2s),
	}

	log.Println("starting server")
	err := h1s.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

func main() {
	var err error
	vm, err = vmCreate()
	if err != nil {
		panic(err)
	}
	startServer()
}