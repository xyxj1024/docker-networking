package main

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Backend struct {
	proxy       *httputil.ReverseProxy
	containerId string
}

type Registry struct {
	backendStore atomic.Value
}

func (r *Registry) Init() {
	r.backendStore.Store([]Backend{})
}

func (r *Registry) Add(id, addr string) {
	u, _ := url.Parse(addr)

	r.backendStore.Swap(append(r.GetAllBackend(), Backend{
		proxy:       httputil.NewSingleHostReverseProxy(u),
		containerId: id,
	}))
}

func (r *Registry) GetBackendByContainerId(id string) (Backend, bool) {
	for _, b := range r.GetAllBackend() {
		if b.containerId == id {
			return b, true
		}
	}
	return Backend{}, false
}

func (r *Registry) GetBackendByIndex(idx int) Backend {
	return r.GetAllBackend()[idx]
}

func (r *Registry) RemoveBackendByContainerId(id string) {
	var backend []Backend
	for _, b := range r.GetAllBackend() {
		if b.containerId == id {
			continue
		}
		backend = append(backend, b)
	}
	r.backendStore.Store(backend)
}

func (r *Registry) RemoveAllBackend() {
	r.backendStore.Store([]Backend{})
}

func (r *Registry) GetAllBackend() []Backend {
	return r.backendStore.Load().([]Backend)
}

func (r *Registry) Len() int {
	return len(r.GetAllBackend())
}

func (r *Registry) List() {
	backend := r.GetAllBackend()
	for i := range backend {
		fmt.Println(backend[i].containerId)
	}
}
