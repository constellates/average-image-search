package main

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"strconv"
)

// structs
type Image struct {
	ContextLink     string `json:"contextLink"`
	Height          int    `json:"height"`
	Width           int    `json:"width"`
	ByteSize        int    `json:"byteSize"`
	ThumbnailLink   string `json:"thumbnailLink"`
	ThumbnailHeight int    `json:"thumbnailHeight"`
	ThumbnailWidth  int    `json:"thumbnailWidth"`
}

type Item struct {
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	HtmlTitle   string `json:"htmlTitle"`
	Link        string `json:"link"`
	DisplayLink string `json:"displayLink"`
	Snippet     string `json:"snippet"`
	HtmlSnippet string `json:"htmlSnippet"`
	Mime        string `json:"mime"`
	Image       Image  `json:"image"`
}

type GoogleImagesResponse struct {
	Items []Item `json:"items"`
}

func getImages(body []byte) (*GoogleImagesResponse, error) {
	var s = new(GoogleImagesResponse)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("whoops:", err)
	}
	return s, err
}

func downloadFile(term string, uri string) {
	// parse filename
	u, err := url.Parse(uri)
	if err != nil {
		log.Fatal(err)
	}
	var fileName = strings.Split(u.Path, "/")[len(strings.Split(u.Path, "/"))-1]

	if strings.Contains(fileName, ".jpg") {
		// create empty file
		out, err := os.Create("temp/" + term + "/" + fileName)
		if err != nil {
			fmt.Println("whoops:", err)
		}
		defer out.Close()

		// get file
		resp, err := http.Get(uri)
		if err != nil {
			fmt.Println("whoops:", err)
		}
		defer resp.Body.Close()

		// save file
		n, err := io.Copy(out, resp.Body)
		fmt.Println(n)
		if err != nil {
			fmt.Println("whoops:", err)
		}

		// resize image
		normalize("temp/" + term + "/" + fileName)
	}
}

func requestImages(term string, count int) []string {
	var i = 1
	var images []string

	for len(images) < count {
		resp, err := http.Get("https://www.googleapis.com/customsearch/v1?key=" + os.Getenv("SearchApiKey") + "&cx=" + os.Getenv("SearchEngineId") + "&q=" + url.QueryEscape(term) + "&imgSize=large&num=10&searchType=image&start=" + strconv.Itoa(i))
		if err != nil {
			// panic(err.Error())
			continue
		}

		// extract body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// panic(err.Error())
			continue
		}

		// unmarshall json
		s, err := getImages([]byte(body))

		var items = s.Items
		for _, item := range items {
			var uri = item.Link
			if strings.HasSuffix(uri, ".jpg") {
				images = append(images, uri)
			}
		}
		i = i + 10
	}

	return images
}

func main() {

	// get input term
	var input []string = append(os.Args[:0], os.Args[1:]...)
	var term = strings.Join(input, " ")

	// make temp dir for term
	os.MkdirAll("temp/"+term, 0777)

	// make google images request
	images := requestImages(term, 5)

	// download images
	for _, url := range images {
		downloadFile(term, url)
	}

	size := image.Point{800, 800}
	execute("temp/"+term, "output.jpg", size)

	// remove source data
	os.Remove("temp/" + term)
}
