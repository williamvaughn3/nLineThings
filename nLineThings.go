//echo "www.google.com" | go run nLineThings.go --stdin
//go run nLineThings.go "www.google.com"

package main

import (
	"bufio"
	"bytes"
	"flag"
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
	useStdin := flag.Bool("stdin", false, "Read URLs from standard input")
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
		log.Printf("Comments for %s:\n%s", server, formatCode(strings.Join(comments, "\n")))
	}
}
func formatCode(code string) string {
	formatted, err := format.Source([]byte(code))
	if err != nil {
		return code
	}
	return string(bytes.TrimSpace(formatted))
}
