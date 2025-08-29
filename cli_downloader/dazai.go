package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
)

//	type Job struct {
//		Url []string
//
// Status int
// }
type Job []string

type Result struct {
	Success bool
	Err     error
	Bytes   float64
}

// worker function
func worker(jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for urls := range jobs {
		//make http get request
		resps, err := GetArrayRequest(urls)
		if err != nil {
			results <- Result{Success: false, Err: err}
			continue
		}

		// Process each response individually
		for i, resp := range resps {
			// Close response body when we're done with it
			defer resp.Body.Close()

			//ask user to save
			filename, err := SavingMultipleFileName()
			if err != nil {
				results <- Result{Success: false, Err: err}
				continue
			}

			//Get the present working directory and have access to it
			dir, err := WorkingDirectory()
			if err != nil {
				results <- Result{Success: false, Err: err}
				continue
			}

			//create filepath
			file := MakingFilePAth(dir, filename)

			//show the progress bar
			f := CreatingFile(file)

			// Create progress bar based on actual content length
			var bar *pb.ProgressBar
			if resp.ContentLength > 0 {
				bar = pb.New64(resp.ContentLength).SetRefreshRate(time.Second).Start()
			} else {
				bar = pb.StartNew(1000) // Fallback if content length unknown
			}

			bar.Set(pb.Bytes, true) // Show as bytes instead of raw numbers
			reader := bar.NewProxyReader(resp.Body)

			//copying the response body(download file itself which is a stream of bytes) into the created file
			_, err = io.Copy(f, reader)
			if err != nil {
				fmt.Printf("error in downloading file %v", err)
				f.Close()
				results <- Result{Success: false, Err: err}
				continue
			}

			// Close file and finish progress bar
			f.Close()
			bar.Finish()

			fmt.Printf("Download %d completed\n", i+1)
		}

		results <- Result{
			Success: true,
			Err:     nil,
		}
	}
}

// results
func collectResults(results <-chan Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for result := range results {
		fmt.Printf("%v, %.1f", result.Err, result.Bytes)
	}
}

// dispatcher to coordinate everything
func dispatcher(jobCount int, workerCount int) {

	urls := os.Args[2:]

	// Check if we have URLs to process
	if len(urls) == 0 {
		fmt.Println("No URLs provided. Please provide URLs as command line arguments.")
		return
	}

	jobs := make(chan Job, jobCount)
	results := make(chan Result, jobCount)

	var wg sync.WaitGroup

	// start workers
	wg.Add(workerCount)
	for w := 1; w <= workerCount; w++ {
		go worker(jobs, results, &wg)
	}

	// start collecting results
	var resultsWg sync.WaitGroup
	resultsWg.Add(1)
	go collectResults(results, &resultsWg)

	// Distribute jobs and wait for completion
	for j := 1; j <= jobCount; j++ {
		jobs <- urls
	}
	close(jobs)

	wg.Wait()
	//close(results)

	// Ensure results are collected
	resultsWg.Wait()

}

// function to handle the add subcommands
func HandleAdd(addCmd *flag.FlagSet, addUrl *string) {
	if *addUrl == "" {
		fmt.Println("use --all to handle multiple files")
		addCmd.PrintDefaults()
		os.Exit(2)
	} else {
		//dispatcher(100, 7)

		const jobCount = 100  //Total number of jobs to process
		const workerCount = 7 //Total number of workers to do the jobs

		fmt.Println("Starting..... ")
		dispatcher(jobCount, workerCount)

	}
}
func GetArrayRequest(urls []string) ([]*http.Response, error) {
	var results []*http.Response
	var failedUrls []string

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second, // 30 second timeout
	}

	for _, url := range urls {
		r, err := client.Get(url)
		if err != nil {
			fmt.Printf("error fetching %s: %v\n", url, err)
			failedUrls = append(failedUrls, url)
			continue
		}

		// Check if the response is successful
		if err := CheckStatusCode(r); err != nil {
			fmt.Printf("HTTP error for %s: %v\n", url, err)
			r.Body.Close() // Close the response body
			failedUrls = append(failedUrls, url)
			continue
		}

		results = append(results, r)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("all requests failed. Failed URLs: %v", failedUrls)
	}

	if len(failedUrls) > 0 {
		fmt.Printf("Warning: %d URLs failed, but proceeding with %d successful downloads\n", len(failedUrls), len(results))
	}

	return results, nil
}
func SavingMultipleFileName() (string, error) {
	// saving the downloaded file with a name which will be imputed by the user
	filename := bufio.NewReader(os.Stdin)
	fmt.Print("Save as: ")
	saveAs, err := filename.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in writing file name %v\n", err)
		return "", err
	}
	so := strings.TrimSpace(saveAs)
	return so, nil
}
func GetRequest(url_downloader string) (*http.Response, error) {
	//sending the get request to the outlined url
	r, err := http.Get(url_downloader)
	if err != nil {
		return nil, fmt.Errorf("error %v with GET request", err)
	}
	return r, nil
}
func CheckStatusCode(r *http.Response) error {
	// Check status code to see if it is an ok
	// Accept 200 (OK), 301/302 (redirects), and 206 (partial content)
	if r.StatusCode != http.StatusOK &&
		r.StatusCode != http.StatusMovedPermanently &&
		r.StatusCode != http.StatusFound &&
		r.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("unexpected status code %v", r.StatusCode)
	}
	return nil
}

/*func CreateDirectory(s string) error {
	if _, err := os.Stat(s); os.IsNotExist(err) {
		fmt.Printf("Directory does not exist: %v\n", err)
		if err = os.MkdirAll(s, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}*/

func WorkingDirectory() (string, error) {
	n, err := os.Getwd()
	if err != nil {
		fmt.Printf("cannot access present working directory %v", err)
	}
	return n, err
}

func SavingFileNAme() (string, error) {
	// saving the downloaded file with a name which will be imputed by the user
	filename := bufio.NewReader(os.Stdin)
	fmt.Print("Save as: ")
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
		log.Fatal(err)
	}
	//return file

	return file
}

// function to handle the get subcommands
func HandleGet(getCmd *flag.FlagSet, getUrl *string) {

	if *getUrl == "" {
		fmt.Println("use -u to fetch a download url")
		getCmd.PrintDefaults()
		os.Exit(2)

	} else {

		//make http get request
		resp, err := GetRequest(*getUrl)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		//check status
		err = CheckStatusCode(resp)
		if err != nil {
			fmt.Printf("HTTP Error: %v\n", err)
			return
		}
		//ask user to save
		filename, err := SavingFileNAme()
		if err != nil {
			log.Fatal(err)
		}

		//Get the present working directory and have access to it
		dir, err := WorkingDirectory()
		if err != nil {
			log.Fatal(err)
		}
		//create filepath
		file := MakingFilePAth(dir, filename)
		//show the progress bar
		f := CreatingFile(file)
		defer f.Close()
		// Create progress bar based on actual content length
		var bar *pb.ProgressBar
		if resp.ContentLength > 0 {
			bar = pb.New64(resp.ContentLength).SetRefreshRate(time.Millisecond).Start()
		} else {
			bar = pb.StartNew(1000) // Fallback if content length unknown
		}

		bar.Set(pb.Bytes, true) // Show as bytes instead of raw numbers
		reader := bar.NewProxyReader(resp.Body)

		//copying the response body(download file itself which is a stream of bytes) into the created file
		_, err = io.Copy(f, reader)
		if err != nil {
			fmt.Printf("error in downloading file %v", err)
			return
		}
		fmt.Println("Download Completed")
	}

}

func main() {

	//dazai get subcommands
	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getUrl := getCmd.String("u", "", "Download the inputed url")
	//dazai add subcommands
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addUrl := addCmd.String("all", "", "Download multiple files at the same time")

	//validaton of commands
	if len(os.Args) < 2 {
		fmt.Println("Expected the 'get' or the 'add' subcommand")
		fmt.Println("Usage:")
		fmt.Println("  go run downloader.go get -u <URL>")
		fmt.Println("  go run downloader.go add --all <URL1> <URL2> ...")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "get":
		// Only parse args if we have enough arguments
		if len(os.Args) < 3 {
			fmt.Println("get command requires a URL. Use: go run downloader.go get -u <URL>")
			os.Exit(4)
		}
		getCmd.Parse(os.Args[2:])
		HandleGet(getCmd, getUrl)
	case "add":
		// Only parse args if we have enough arguments
		if len(os.Args) < 3 {
			fmt.Println("add command requires URLs. Use: go run downloader.go add --all <URL1> <URL2> ...")
			os.Exit(1)
		}
		addCmd.Parse(os.Args[2:])
		HandleAdd(addCmd, addUrl)
	default:
		println("Don't Understand that command")
		fmt.Println("Available commands: get, add")
	}

}
