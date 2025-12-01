package caoscrape_test

import (
	"fmt"
	"io"
	"log"

	"rag/caoscrape"
)

func Example() {
	// Create a new client
	client := caoscrape.NewClient()

	// Search for documents with JC number 3180200
	jc := 3180200
	urls, err := client.Search(&jc)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d documents\n", len(urls))
	for i, url := range urls {
		fmt.Printf("%d. %s\n", i+1, url)
	}

	// Download the first document if available
	if len(urls) > 0 {
		reader, err := client.DownloadDocument(urls[0])
		if err != nil {
			log.Fatal(err)
		}

		// Read some bytes to verify download
		buf := make([]byte, 100)
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		fmt.Printf("Downloaded %d bytes\n", n)
	}
}

func ExampleSearchAll() {
	// Create a new client
	client := caoscrape.NewClient()

	// Search for all documents (nil JC parameter)
	urls, err := client.Search(nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d total documents\n", len(urls))
}
