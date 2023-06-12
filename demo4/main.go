package main

import (
	"log"
	"net/http"
)

func main() {
	registry := Registry{}
	registry.Init()

	client, err := NewDockerClient()
	if err != nil {
		panic(err)
	}

	registrar := Registrar{serviceRegistry: &registry, dockerClient: client}
	if err = registrar.Init(); err != nil {
		panic(err)
	}
	go registrar.Observe()

	log.Println("Handle service registry queries...")
	app := Application{serviceRegistry: &registry, requestCount: 0}
	http.HandleFunc("/", app.Handle)

	log.Fatalln(http.ListenAndServe(":3000", nil))
}
