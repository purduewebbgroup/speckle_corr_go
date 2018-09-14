package corr

import "io/ioutil"
import "log"
import "os"
import "sync"

type ImagePixel float32

type SpeckleFile struct {
	Filename      string
	BaseDirectory string
	Position      Position
	ScanGroupId   int
}

type Position struct {
	X int32
	Y int32
	Z int32
}

type SpeckleImage struct {
	Filename        string
	Position        Position
	ScanGroupId     int
	backgroundImage *[]ImagePixel

	image     []ImagePixel
	imageLock sync.Mutex

	normalizedImage []ImagePixel
	normalizedLock  sync.Mutex
}

func GetBackground(BaseDirectory string) []ImagePixel {

	backgroundFilename1 := BaseDirectory + "/background_1"
	backgroundFilename2 := BaseDirectory + "/background_2"

	background := LoadBackgroundImage(backgroundFilename1, backgroundFilename2)

	return background
}

func LoadBackgroundImage(f1 string, f2 string) []ImagePixel {

	image1 := loadImageFile(f1)
	image2 := loadImageFile(f2)

	background := make([]ImagePixel, 1392*1040)

	for i := range image1 {
		background[i] = (image1[i] + image2[i]) / 2
	}

	return background
}

func loadImageFile(filename string) []ImagePixel {

	raw, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	image := make([]ImagePixel, 2048*2048)

	if 2048*2048 != len(raw)/2 {
		log.Fatal(2048*2048, "!=", len(raw)/2)
	}

	// ignore the last two bytes
	// last two bytes should be EOF

	for i, v := range raw {

		if i >= 1392*1040*2 {
			break
		}

		multiplier := int32(1)

		if i%2 == 1 {
			multiplier = 1 << 8
		}

		image[i/2] += ImagePixel(int32(v) * multiplier)
	}

	return image
}

// Wait until first use to load
func LoadImage(speckleFile SpeckleFile, background []ImagePixel) SpeckleImage {

	image := SpeckleImage{}
	image.Filename = speckleFile.BaseDirectory + string(os.PathSeparator) + speckleFile.Filename
	image.Position = speckleFile.Position
	image.ScanGroupId = speckleFile.ScanGroupId

	image.backgroundImage = &background

	return image
}

type Image []ImagePixel

func (s *SpeckleImage) GetImage(counter chan string) Image {

	s.imageLock.Lock()

	if len(s.image) > 0 {
		s.imageLock.Unlock()
		return s.image

	}

	counter <- "load image"

	image := loadImageFile(s.Filename)

	// do we really care if we have a NxN array to return
	// or a Mx1 array where M=N*N

	// take off the background image
	for i := 0; i < len(image); i++ {

		image[i] = image[i] - (*s.backgroundImage)[i]

	}

	s.imageLock.Unlock()

	// cache the image?
	// s.image = &image

	return image
}
