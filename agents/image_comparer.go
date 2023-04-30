package agents

import (
	"image/png"
	"os"

	"github.com/corona10/goimagehash"
)

func (a *URLScreenshotter) Compare(image1Path string, image2Path string) int {
	a.session.Out.Debug("Comparing: Image1: %s | Image2: %s\n", image1Path, image2Path)
	image1File, err := os.Open(image1Path)
	if err != nil {
		a.session.Out.Error(err.Error())
	}
	defer image1File.Close()
	image2File, err := os.Open(image2Path)
	if err != nil {
		a.session.Out.Error(err.Error())
	}
	defer image2File.Close()
	img1, _ := png.Decode(image1File)
	img2, _ := png.Decode(image2File)
	hash1, _ := goimagehash.DifferenceHash(img1)
	hash2, _ := goimagehash.DifferenceHash(img2)
	distance, _ := hash1.Distance(hash2)
	return distance
}
