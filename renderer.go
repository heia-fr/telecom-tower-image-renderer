package imgrenderer

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
    "image"
    "image/color"
	"image/png"
	"image/draw"
	"net/http"
    "bytes"
    "strconv"
    "math"
    "errors"
)

const (
    DEFAULT_PIXSIZE = 4
)

var STROKE_COLOR color.RGBA = color.RGBA{50, 50, 50, 185}

type Stripe []uint32

type Matrix struct {
	Rows    int    `json:"rows"`
	Columns int    `json:"columns"`
	Bitmap  Stripe `json:"bitmap"`
}

// GetPixel returns the uint32-encoded color of the matrix
// at index [x,y]
func (m *Matrix) GetPixel(x, y int) uint32 {
	if y < 0 || y >= m.Rows {
		panic("y out of bound")
	}
	if x < 0 || x >= m.Columns {
		panic("x out of bound")
	}
	return m.Bitmap[x*m.Rows+y]
}


func init() {
    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/renderImage", renderImage)
    router.HandleFunc("/renderRealistic", renderRealisticImage)
    http.Handle("/", router)
    //http.ListenAndServe(":8080", nil)
}

// renderImage renders a "normal" image
func renderImage(w http.ResponseWriter, r *http.Request) {

    // get optional  query parameter
    pixSize := DEFAULT_PIXSIZE
    err := getPixSizeParam(r, &pixSize)
    if err != nil {
        http.Error(w, err.Error(), 400)
    }

    // get matrix
    var matrix Matrix
	err = decodeMatrix(r, &matrix)

	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), 400)
		return
	}

    m := generateImage(&matrix, pixSize)
    writeImage(w, m)
}

// renderImage renders a realistic image
func renderRealisticImage(w http.ResponseWriter, r *http.Request) {

    // get optional  query parameter
    pixSize := DEFAULT_PIXSIZE
    err := getPixSizeParam(r, &pixSize)
    if err != nil {
        http.Error(w, err.Error(), 400)
    }

    // get matrix
    var matrix Matrix
    err = decodeMatrix(r, &matrix)

	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), 400)
		return
	}

    m := generateRealisticImage(&matrix, pixSize)
    writeImage(w, m)
}

// getPixSizeParam parses the optional pixSize parameter
// if not given, the pixSize parameter is left intact
// if given but incorrect, it will return an error
func getPixSizeParam(r *http.Request, pixSize *int) error {
    param := r.FormValue("pixSize")

    if param != "" {
        size, err := strconv.Atoi(param)
        if err != nil || size <= 0 || size >= 30 {
            return errors.New(fmt.Sprintf("Invalid  parameter (should be an int between 1 and 30): %s", param))
        }
        *pixSize = size
    }

    return nil
}

// decodeMatrix parses the body of the request and converts the
// json to a matrix struct. If the json is incorrect, it will return an error
func decodeMatrix(r *http.Request, matrix *Matrix) (err error){
    d := json.NewDecoder(r.Body)
    defer r.Body.Close()
    err = d.Decode(&matrix)
    return
}

// writeImage encodes an image 'img' in jpeg format and writes it into ResponseWriter.
func writeImage(w http.ResponseWriter, img *image.RGBA) {

    buffer := new(bytes.Buffer)
    if err := png.Encode(buffer, img); err != nil {
        http.Error(w, fmt.Sprintf("Unable to encode resulting image: %v", err), 400)
    }

    w.Header().Set("Content-Type", "image/png")
    w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))

    if _, err := w.Write(buffer.Bytes()); err != nil {
        http.Error(w, fmt.Sprintf("Unable to write resulting image: %v", err), 400)
    }
}

// toRGB converts a uint32 encoded color to an RGBA color
func toRGB(c uint32) color.RGBA {
    return color.RGBA{
        uint8(c>>16),
        uint8(c>>8),
        uint8(c),
        255,
    }
}


// -------------- realistic image

func generateRealisticImage(matrix *Matrix, pixSize int) (m *image.RGBA){

    // ensure pixSize if odd, so we can center our circles
    if pixSize%2 == 0 {
        pixSize++
    }

    w, h := matrix.Columns, matrix.Rows
    radius := int(pixSize / 2)

    m = image.NewRGBA(image.Rect(0, 0, w*pixSize, h*pixSize))
    draw.Draw(m, m.Bounds(), &image.Uniform{color.RGBA{0,0,0,255}}, image.ZP, draw.Src)
    for x := 0; x < w; x++ {
        for y := 0; y < h; y++ {
            encodedColor := matrix.GetPixel(x,y)
            color := toRGB(encodedColor)
            fillCircle(m, &color, x*pixSize+radius, y*pixSize+radius, radius)
            drawCircle(m, &STROKE_COLOR, x*pixSize+radius, y*pixSize+radius, radius)
        }
    }

    return
}


func drawCircle(m *image.RGBA, c *color.RGBA, centerX, centerY, radius int){

    l := int(float64(radius) * math.Cos(math.Pi/4))

    for x := 0; x <= l; x++ {
        y := int(math.Sqrt(float64(radius*radius) - float64(x*x)))
        m.Set(centerX + x, centerY + y,c)
        m.Set(centerX + x, centerY - y,c)
        m.Set(centerX - x, centerY + y,c)
        m.Set(centerX - x, centerY - y,c)

        m.Set(centerX + y, centerY + x,c)
        m.Set(centerX + y, centerY - x,c)
        m.Set(centerX - y, centerY + x,c)
        m.Set(centerX - y, centerY - x,c)
    }
}


func fillCircle(img *image.RGBA, c *color.RGBA, centerX, centerY, radius int){
    radSquare := radius * radius
    for x := -radius; x <= radius; x++ {
        for y := -radius; y <= radius; y++ {
            if (x*x) + (y*y) <= radSquare {
                img.Set(centerX + x, centerY + y, c);
            }
        }
    }
}

// -------------- "normal" image

func generateImage(matrix *Matrix, pixSize int) *image.RGBA {

    var w, h int = matrix.Columns, matrix.Rows

    m := image.NewRGBA(image.Rect(0, 0, w*pixSize, h*pixSize))
    for x := 0; x < w; x++ {
        for y := 0; y < h; y++ {
            u := matrix.GetPixel(x,y)
            c := toRGB(u)
            setOnePixel(m, &c, x*pixSize, y*pixSize, pixSize)
        }
    }

    return m
}

func setOnePixel(m *image.RGBA, c *color.RGBA, offsetX, offsetY, pixSize int){
    for x := 0; x < pixSize; x++ {
        for y := 0; y < pixSize; y++ {
            m.Set(offsetX+x, offsetY+y, c)
        }
    }
}
