package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type Application struct {
	requestCount    uint64
	serviceRegistry *Registry
}

func (a *Application) Handle(w http.ResponseWriter, req *http.Request) {
	atomic.AddUint64(&a.requestCount, 1)

	if a.serviceRegistry.Len() == 0 {
		w.Write([]byte(`No backend entry in the service registry`))
		return
	}

	idx := int(atomic.LoadUint64(&a.requestCount) % uint64(a.serviceRegistry.Len()))
	fmt.Printf("Request routing to instance %d\n", idx)

	b := a.serviceRegistry.GetBackendByIndex(idx)
	b.proxy.ServeHTTP(w, req)
}
