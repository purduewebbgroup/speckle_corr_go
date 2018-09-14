package main

import "corr"
import "flag"
import "fmt"
import "io/ioutil"
import "log"

import "time"

func main() {

	numProcessors := 20

	log.Println("Finding Files")

	baseDirectory := flag.String("dir", "/home/yara/webb/newmanj/speckle_data/20140211_000", "The directory where the speckle images are located")
	flag.Parse()

	//  baseDirectory := "/home/yara/webb/newmanj/speckle_data/20140210_001"

	files := FindAndLoadFiles(*baseDirectory)

	log.Printf("%d files found\n", len(files))

	corr := corr.Initialize(*baseDirectory, numProcessors)

	startTime := time.Now()

	corr.Start(files)

	log.Println(corr)

	log.Println("Run Time ", time.Since(startTime))

}

func FindAndLoadFiles(BaseDirectory string) []corr.SpeckleImage {

	files := FindFiles(BaseDirectory)
  // files contain set of index and coordinates
	images := []corr.SpeckleImage{}

	background := corr.GetBackground(BaseDirectory)

	for _, v := range files {
		speckleImage := corr.LoadImage(v, background)

		images = append(images, speckleImage)
	}

	return images

}

func FindFiles(BaseDirectory string) []corr.SpeckleFile {

	files, err := ioutil.ReadDir(BaseDirectory)

	if err != nil {
		log.Fatal(err)
	}

	speckleFiles := []corr.SpeckleFile{}

	count := 0

	// could potentially use path/filepath.Glob(pattern string) instead

	for _, file := range files {

		if file.IsDir() {
			continue
		}
    // skip any directory
		// check to see if the filename matches the standard format
		// o_002_000002_000015
		// o_%03d_%06d_%06d
		// number lengths not guaranteed

		var offsetA, offsetB int
		var index int

		n, err := fmt.Sscanf(file.Name(), "o_%d_%d_%d", &index, &offsetA, &offsetB)

		if err != nil {
			// log.Println("FindFiles ", err, file.Name(), n)

		} else if n == 3 {
			count++
			speckleFile := corr.SpeckleFile{Filename: file.Name(),
				BaseDirectory: BaseDirectory,
				Position:      corr.Position{Y: int32(offsetA), X: int32(offsetB)},
				ScanGroupId:   index}
			speckleFiles = append(speckleFiles, speckleFile)

		}

	}

	return speckleFiles
	// return struct SpeckleFile

}
