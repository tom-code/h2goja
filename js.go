package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/appscode/go/ioutil"
	"github.com/dop251/goja"
	"golang.org/x/net/http2"
)


type Param struct {
	Url string
	Data []byte
	Status int
	DataOut []byte
}

type VM struct {
	vm *goja.Runtime
	requestF goja.Callable
	http http.Client
	mutex sync.Mutex
}

func (vm *VM)Call(param *Param) {
	vm.mutex.Lock()
	vm.requestF(nil, vm.vm.ToValue(param))
	vm.mutex.Unlock()
}

func vmCreate() (*VM, error) {
	vmo := &VM{}

	vm := goja.New()

	vm.Set("log", func(c goja.FunctionCall) goja.Value {
		if len(c.Arguments) < 1 {
			log.Println("not enough arguments for log")
			return goja.Null()
		}
		log.Printf("[log] %v\n", c.Arguments[0])
		return goja.Null()
	})

	vm.Set("h2r", func(c goja.FunctionCall) goja.Value {
		if len(c.Arguments) < 2 {
			log.Println("not enough arguments for h2r")
			return goja.Null()
		}
		url := c.Arguments[0].String()
		callback, ok := goja.AssertFunction(c.Arguments[1])
		if !ok {
			log.Println("h2r argument error")
			return goja.Null()
		}
		go func() {
			resp, err := vmo.http.Get(url)
			if err != nil {
				log.Println(err)
			}
			vmo.mutex.Lock()
			callback(nil, vm.ToValue(resp))
			vmo.mutex.Unlock()
		}()
		return goja.Null()
	})

	scriptData, err := ioutil.ReadFile("script.js")
	if err != nil {
		log.Println("can't open script")
		return vmo, fmt.Errorf("can't open script %s", err.Error())
	}
	_, err = vm.RunString(scriptData)
	if err != nil {
		log.Println(err)
	}

	var ok bool
	reqObj := vm.GlobalObject().Get("request")
	vmo.requestF, ok = goja.AssertFunction(reqObj)
	if !ok {
		return vmo, fmt.Errorf("can't locate request function")
	}

	vmo.vm = vm

	vmo.http = http.Client{
		Timeout: 10*time.Second,
	}
	vmo.http.Transport = &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.DialTimeout(netw, addr, 10*time.Second)
	  }}
	return vmo, nil
}
