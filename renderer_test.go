package imgrenderer

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
    "io/ioutil"
    "image"
    _ "image/png"
	"testing"
    "bytes"
)

func testRenderImage(t *testing.T){
    postData, err := ioutil.ReadFile("good_matrix.json")
    assert.NoError(t, err)
    goodImageBytes, _ := ioutil.ReadFile("good_image.png")
    assert.NoError(t, err)
    goodImage, _, err := image.Decode(bytes.NewBuffer(goodImageBytes))
    assert.NoError(t, err)

    req, err := http.NewRequest("POST", "/renderImage", bytes.NewBuffer(postData))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	renderImage(w, req)

    assert.Equal(t, 200, w.Code)

    m, _, err := image.Decode(w.Body)
    assert.NoError(t, err)

    assert.Equal(t, goodImage.Bounds(), m.Bounds())

    for x := goodImage.Bounds().Max.X; x >= 0; x-- {
        for y := goodImage.Bounds().Max.Y; y >= 0; y-- {
            assert.Equal(t, goodImage.At(x,y), m.At(x,y))
        }
    }
}
