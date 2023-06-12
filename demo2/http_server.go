// Run: docker run -d -p 8000:8000 --name $CONT_NAME -t $TAG_NAME
// Check: curl localhost:8000

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Fprintf(os.Stdout, "Listening on port %s\n", port)
	hostname, _ := os.Hostname()
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stdout, "I'm %s\n", hostname)
		fmt.Fprintf(rw, "I'm %s\n", hostname)
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
