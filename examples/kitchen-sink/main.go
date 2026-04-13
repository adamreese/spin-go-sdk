package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	spinhttp "github.com/spinframework/spin-go-sdk/v3/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Hello World!")

		// Test for panics using the logger
		log.Print("Hello logger")

		// Reading all environment variables
		envs := os.Environ()
		fmt.Printf("envs: %s\n", envs)

		// Reading a named environment variable
		foo := os.Getenv("FOO")
		fmt.Printf("FOO=%s\n", foo)

		// List files
		files, err := os.ReadDir(".")
		if err != nil {
			log.Printf("Error reading dir: %#v\n", err)
			return
		}
		if len(files) != 1 || files[0].Name() != "test.data" {
			fmt.Printf("Files don't match: %#v\n", files)
			return
		}

		// Reading a missing file
		_, err = os.ReadFile("nope")
		if !errors.Is(err, fs.ErrNotExist) {
			log.Printf("Should have path error, got: %#v\n", err)
			return
		}

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

		// Writing a file

		// if err := os.WriteFile("test.data", []byte("some random data"), 0644); err != nil {
		// 	log.Print("Error calling os.WriteFile(): ", err)
		// 	return
		// }

		// Should I be able to write to a file without mounting anything to
		// the guest?
		// err := os.WriteFile("file.data", []byte("some random data"), 0644)
		// if err != nil {
		// 	log.Print("Error calling os.WriteFile()", err)
		// 	return
		// }

	})
}

func main() {}
