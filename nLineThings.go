// echo 'internal-www-test-server.com' | go run nLineThings.go -stdin
// go run nLineThings.go -out outPutComments.txt internal-www-test-server.com
// for i in `cat urls.txt`; do go run nLineThings.go -out ./outputComments.txt $i; done

// go run nLineThings.go -h 

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

const commentRegex = `(?m)(<!--(.*?)-->)|(/\*([^*]|[\r\n]|(\*+([^*/]|[\r\n])))*\*+/)|(//.*)|(^'.*$)|(^#.*$)`

func main() {
	var outFile string
	useStdin := flag.Bool("stdin", false, "Read URLs from standard input")
	flag.StringVar(&outFile, "out", "", "Output file to append results")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] [URLS...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "OPTIONS:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "URLS:\n")
		fmt.Fprintf(os.Stderr, "  List of URLs to check for comments\n")
	}
	flag.Parse()

	webServers := []string{}
	if *useStdin {

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			webServers = append(webServers, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("Error reading standard input: %v", err)
		}
	} else {

		webServers = flag.Args()
	}

	re := regexp.MustCompile(commentRegex)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var out *os.File
	if outFile != "" {
		var err error
		out, err = os.OpenFile(outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Error opening output file %s: %v", outFile, err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	for _, server := range webServers {
		u, err := url.Parse(server)
		if err != nil {
			log.Printf("Error parsing URL %s: %v", server, err)
			continue
		}

		if u.Scheme == "" {
			u.Scheme = "http"
		}

		resp, err := client.Get(u.String())
		if err != nil {
			log.Printf("Error requesting %s: %v", server, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode < 100 || resp.StatusCode >= 400 {
			log.Printf("Unexpected status code %d for %s", resp.StatusCode, server)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body from %s: %v", server, err)
			continue
		}

		comments := re.FindAllString(string(body), -1)
		fmt.Fprintf(out, "Comments for %s:\n%s\n", server, formatCode(strings.Join(comments, "\n")))
	}
}

func formatCode(code string) string {
	formatted, err := format.Source([]byte(code))
	if err != nil {
		return code
	}
	return string(bytes.TrimSpace(formatted))
}
