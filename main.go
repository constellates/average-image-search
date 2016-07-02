package main

import (
	"log"
    "io"
    "io/ioutil"
	"net/http"
	"net/url"
	"os"
	"fmt"
	"encoding/json"
	"strings"

	"image"
	"image/color"
	"image/jpeg"
	// "flag"
	// "regexp"
	"path"

	"github.com/nfnt/resize"
    "strconv"
)

// configuration
type Configuration struct {
    SearchApiKey     string `json:"searchApiKey"`
    SearchEngineId     string `json:"searchEngineId"`
}

// structs
type Image struct {
	ContextLink     string `json:"contextLink"`
    Height          int `json:"height"`
    Width           int `json:"width"`
    ByteSize        int `json:"byteSize"`
    ThumbnailLink   string `json:"thumbnailLink"`
    ThumbnailHeight int `json:"thumbnailHeight"`
    ThumbnailWidth  int `json:"thumbnailWidth"`
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
    Image       Image `json:"image"`
}

type GoogleImagesResponse struct {
	Items []Item `json:"items"`
}

func getImages(body []byte) (*GoogleImagesResponse, error) {
    var s = new(GoogleImagesResponse)
    err := json.Unmarshal(body, &s)
    if(err != nil){
        fmt.Println("whoops:", err)
    }
    return s, err
}

func downloadFile(term string, uri string) () {
	// parse filename
	u, err := url.Parse(uri)
	if err != nil {
		log.Fatal(err)
	}
	var fileName = strings.Split(u.Path, "/")[len(strings.Split(u.Path, "/")) - 1]

	if strings.Contains(fileName, ".jpg") {
		// create empty file
		out, err := os.Create("temp/" + term + "/" + fileName)
		if(err != nil){
	        fmt.Println("whoops:", err)
	    }
		defer out.Close()

		// get file
		resp, err := http.Get(uri)
		if(err != nil){
	        fmt.Println("whoops:", err)
	    }
		defer resp.Body.Close()

		// save file
		n, err := io.Copy(out, resp.Body)
		fmt.Println(n)
		if(err != nil){
	        fmt.Println("whoops:", err)
	    }

	    // resize image
	    normalize("temp/" + term + "/" + fileName)
	}
}

func requestImages(term string, count int) []string {

	// configuration
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
	  fmt.Println("error reading configuration:", err)
	}

	var i = 1
	var images []string

	for len(images) < count {
		resp, err := http.Get("https://www.googleapis.com/customsearch/v1?key=" + configuration.SearchApiKey + "&cx=" + configuration.SearchEngineId + "&q=" + url.QueryEscape(term) + "&imgSize=large&num=10&searchType=image&start=" + strconv.Itoa(i))
		if err != nil {
			panic(err.Error())
		}

		// extract body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
		    panic(err.Error())
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
	os.MkdirAll("temp/" + term, 0777)

	// make google images request
	images := requestImages(term, 30)

	// download images
	for _, url := range images {
    	downloadFile(term, url)
    }

    size := image.Point{500, 500}
    execute("temp/" + term, "output.jpg", size)

    // remove source data
    os.Remove("temp/" + term)
}


func normalize(filePath string) () {
	 // open "test.jpg"
    file, err := os.Open(filePath)
    if err != nil {
        log.Fatal(err)
    }

    // decode jpeg into image.Image
    img, err := jpeg.Decode(file)
    if err != nil {
        // disregard images we can't decode
        return
    }
    file.Close()

    // resize to width 1000 using Lanczos resampling
    // and preserve aspect ratio
    m := resize.Resize(500, 500, img, resize.Lanczos3)

    out, err := os.Create(filePath)
    if err != nil {
        log.Fatal(err)
    }
    defer out.Close()

    // write new image to file
    jpeg.Encode(out, m, nil)

}




//averageImage
type Values struct {
	forSize     image.Point
	count       uint32
	// uint32 enough space for 16777216 images
	red_total   []uint32
	green_total []uint32
	blue_total  []uint32
}

func execute(folder string, outputImage string, size image.Point) {
	files, err := ioutil.ReadDir(folder)
	handleError(err)

	xy := size.X * size.Y
	value := Values{size, 0, make([]uint32, xy), make([]uint32, xy), make([]uint32, xy)}

	i := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		log.Printf("Reading image %s", file.Name())
		if readImg(&value, path.Join(folder, file.Name())) {
			i++
		}
	}

	log.Printf("Read %d images", i)
	writeAverageImage(&value, outputImage)
}

func readImg(values *Values, imageFile string) bool {
	file, err := os.Open(imageFile)
	handleError(err)
	defer file.Close()
	image, _, err := image.Decode(file)

	size := image.Bounds().Size()
	if !size.Eq(values.forSize) {
		log.Printf("Image '%s' with %d x %d doesn't have the correct size (%d x %d)",
			imageFile, size.X, size.Y, values.forSize.X, values.forSize.Y)
		return false
	}

	values.count++

	var arrayCounter uint32 = 0
	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			r, g, b, _ := image.At(x, y).RGBA()
			values.red_total[arrayCounter] += r
			values.green_total[arrayCounter] += g
			values.blue_total[arrayCounter] += b
			arrayCounter++
		}
	}

	return true
}

func writeAverageImage(values *Values, resultImage string) {
	if values.count == 0 {
		log.Println("No data for result image")
		return
	}

	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, values.forSize})

	var arrayCounter uint32 = 0
	for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
		for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
			r := uint8(values.red_total[arrayCounter] / values.count / 256)
			g := uint8(values.green_total[arrayCounter] / values.count / 256)
			b := uint8(values.blue_total[arrayCounter] / values.count / 256)
			img.Set(x, y, color.RGBA{r, g, b, 0xff})
			arrayCounter++
		}
	}

	file, err := os.Create(resultImage)
	handleError(err)
	defer file.Close()

	jpeg.Encode(file, img, &jpeg.Options{95})
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
