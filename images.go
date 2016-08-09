package main

import (
	"log"
    "io/ioutil"
	"os"
	// "fmt"

	"image"
	"image/color"
	"image/jpeg"
	"path"
	"image/draw"

	"github.com/nfnt/resize"
)

func normalize(filePath string) () {
	 // open "test.jpg"
    file, err := os.Open(filePath)
    if err != nil {
        // log.Fatal(err)
        return
    }

    // decode jpeg into image.Image
    img, err := jpeg.Decode(file)
    if err != nil {
        // disregard images we can't decode
        return
    }
    file.Close()

    // resize to width 800 using Lanczos resampling
    // and preserve aspect ratio
    var resizeX = 800
    var resizeY = 800
    if img.Bounds().Max.X > img.Bounds().Max.Y {
	    resizeY = 0
	} else {
	    resizeX = 0
	}
    m := resize.Resize(uint(resizeX), uint(resizeY), img, resize.Lanczos3)

    // create black square
    n := image.NewRGBA(image.Rect(0, 0, 800, 800))
	black := color.RGBA{0, 0, 0, 255}
	draw.Draw(n, n.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)

	// calculate top left image point for centering
	var imageStartX = 0
    var imageStartY = 0
    if m.Bounds().Max.X > m.Bounds().Max.Y {
	    imageStartX = 0
	    imageStartY = -(800 - m.Bounds().Max.Y) / 2
	} else {
		imageStartX = -(800 - m.Bounds().Max.X) / 2
	    imageStartY = 0
	}

	// superimpose image on background
    draw.Draw(n, n.Bounds(), m, image.Point{imageStartX,imageStartY}, draw.Src)

    out, err := os.Create(filePath)
    if err != nil {
        log.Fatal(err)
    }
    defer out.Close()

    // write new image to file
    jpeg.Encode(out, n, nil)

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
		return
		// log.Fatal(err)
	}
}