package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	spinhttp "github.com/spinframework/spin-go-sdk/v3/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Hello World!")

		// Reading a mounted file
		dat, err := os.ReadFile("test.data")
		if err != nil {
			log.Print("Error calling os.ReadFile()", err)
			return
		}
		if string(dat) != "This is used for testing" {
			log.Print("Error calling os.ReadFile(). Files contents don't match")
			log.Print("Contents: ", string(dat))
			return
		}
	})
}

func main() {}
