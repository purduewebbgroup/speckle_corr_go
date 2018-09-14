package corr

import "log"
import "math"

// Returns a normalized image, such that the mean is zero and the standard deviation is one.
func (s *SpeckleImage) GetNormalizedImage(counter chan string) []ImagePixel {

	s.normalizedLock.Lock()
	defer s.normalizedLock.Unlock()

	if len(s.normalizedImage) > 0 {
		return s.normalizedImage
	}

	counter <- "calculate normalized"

	s.normalize(counter)

	return s.normalizedImage

}

// Normalize should  only be called from GetNormalized.
func (s *SpeckleImage) normalize(counter chan string) {

	//s.normalizedLock.Lock()

	image := s.GetImage(counter)

	mean := image.Mean()
	stdDev := image.StdDev()

	s.normalizedImage = make([]ImagePixel, len(image))

	log.Printf("Mean: %0.3f, StdDev: %0.3f, Contrast Ratio: %0.3f\n", mean, stdDev, stdDev/mean)

	for i, v := range image {

		// which normalization to use? (I - <I>) / <I> or (I - <I>) / stddev(I)
		// with zero mean circular Gaussian field statistics they are equivalent
		if false {
			s.normalizedImage[i] = (v - mean) / stdDev
		} else {
			s.normalizedImage[i] = (v - mean) / mean
		}

	}

	//s.normalizedLock.Unlock()
	return

}

// Mean calculates the mean intensity of a given image.
func (image Image) Mean() ImagePixel {

	counter := ImagePixel(0.0)
	N := ImagePixel(0)
	for i, v := range image {
		// log.Printf("% 2d %g\n", i, v)
		counter += v

		N = ImagePixel(i + 1)

		if i >= 20-1 {
			// break
		}
	}

	//N := ImagePixel(len(s.image))
	mean := ImagePixel(counter) / N

	return mean
}

// Calculates the standerd deviation of a given image.
func (image Image) StdDev() ImagePixel {

	counter := 0.0
    // _ ,means omitting unused variable
	for _, v := range image {
		counter += float64(v * v)
	}

	mean := image.Mean()

	N := float64(len([]ImagePixel(image)))
	variance := counter/N - float64(mean*mean)

	return ImagePixel(math.Sqrt(variance))
}
