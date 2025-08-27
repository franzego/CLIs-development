package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/lpernett/godotenv"
)

func main() {

	//godotenv for loading environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error in accessing .env file")
	}
	//sending the get request
	URL_downloader := os.Getenv("Download_url")
	res, err := http.Get(URL_downloader)
	if err != nil {
		fmt.Printf("error with GET request to specified url %v", err)
		return
	}
	defer res.Body.Close()
	// Check status code
	if res.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code %v\n", res.StatusCode)
		return
	}

	//create a directory to store the files
	dir := os.Getenv("Local_Storage")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Directory does not exist: %v\n", err)
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	// saving the downloaded file with a name which will be imputed by the user
	filename := bufio.NewReader(os.Stdin)
	fmt.Println("Save as: ")
	saveAs, err := filename.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in writing file name %v\n", err)
		return
	}
	name := strings.TrimSpace(saveAs)

	// making the path that will have the saved file
	filepath := filepath.Join(dir, name)

	// creating the file itself
	file, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
	} else {
		defer file.Close()
	}
	// Handling the progress bar
	count := 10000
	// create and start new bar
	bar := pb.StartNew(count)

	for i := 0; i < count; i++ {
		bar.Increment()
		time.Sleep(time.Millisecond)
	}

	// refresh info every second (default 200ms)
	bar.SetRefreshRate(time.Second)
	bar.Set(pb.Bytes, true) // Show as bytes instead of raw numbers

	// Create proxy reader
	reader := bar.NewProxyReader(res.Body)

	//copying the response body(download file itself which is a stream of bytes) into the created file
	_, err = io.Copy(file, reader)
	if err != nil {
		fmt.Printf("error in downloading file %v", err)
	}
	// Finish progress bar
	bar.Finish()
}
