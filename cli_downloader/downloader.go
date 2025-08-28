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

func GetRequest(url string) (*http.Response, error) {
	//sending the get request to the outlined url
	r, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error %v with GET request", err)

	}
	return r, nil

}
func CheckStatusCode(r *http.Response) error {
	// Check status code to see if it is an ok
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %v", r.StatusCode)
	}
	return nil
}
func CreateDirectory(s string) error {
	if _, err := os.Stat(s); os.IsNotExist(err) {
		fmt.Printf("Directory does not exist: %v\n", err)
		if err = os.MkdirAll(s, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}
func SavingFileNAme() (string, error) {
	// saving the downloaded file with a name which will be imputed by the user
	filename := bufio.NewReader(os.Stdin)
	fmt.Println("Save as: ")
	saveAs, err := filename.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in writing file name %v\n", err)
	}
	so := strings.TrimSpace(saveAs)
	return so, nil
}
func MakingFilePAth(n string, so string) string {
	// making the path that will have the saved file
	// where n is directory and s is file name
	filepaths := filepath.Join(n, so)
	return filepaths
}
func CreatingFile(so string) *os.File {
	// creating the file itself
	file, err := os.Create(so)
	if err != nil {
		fmt.Printf("Error in creating file: %v", err)
	}
	return file
}
func main() {
	//godotenv for loading environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error in accessing .env file")
	}
	// load environment variable
	dir := os.Getenv("Local_Storage")
	url_downloader := os.Getenv("Download_url")
	//create a directory to store the files
	if err := CreateDirectory(dir); err != nil {
		log.Fatal(err)
	}
	//make http get request
	resp, err := GetRequest(url_downloader)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	//check status
	if err := CheckStatusCode(resp); err != nil {
		log.Fatal(err)
	}
	//ask user to save
	filename, err := SavingFileNAme()
	if err != nil {
		log.Fatal(err)
	}
	//create filepath
	file := MakingFilePAth(dir, filename)
	//show the progress bar
	f := CreatingFile(file)
	defer f.Close()
	//save the file
	// Handling the progress bar
	count := 1000
	// create and start new bar
	//bar := pb.StartNew(count)
	bar := pb.New64(resp.ContentLength).SetRefreshRate(time.Second).Start()
	for i := 0; i < count; i++ {
		bar.Increment()
		time.Sleep(time.Millisecond)
	}
	// refresh info every second (default 200ms)
	bar.SetRefreshRate(time.Second)
	bar.Set(pb.Bytes, true) // Show as bytes instead of raw numbers
	// Create proxy reader
	reader := bar.NewProxyReader(resp.Body)
	// Finish progress bar
	bar.Finish()
	//copying the response body(download file itself which is a stream of bytes) into the created file
	_, err = io.Copy(f, reader)
	if err != nil {
		fmt.Printf("error in downloading file %v", err)
	}
	fmt.Println("Download Completed")
}
