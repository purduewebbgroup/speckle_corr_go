package corr

import "fmt"
import "log"
import "math"
import "runtime"
import "sync"
import "time"

type CorrelationCalculator struct {
	InputJobChannel chan CorrelationJob
	OutputResults   chan CorrelationResult
	// InputJobChannel and OutputResults as channel of type CorrelationJob and type CorrelationResult


	scanGroupDiff int

	baseDirectory string
	counter       chan string
	counts        map[string]int64

	correlationResult map[string]CorrelationSum
	// variable correlationResult is a map of string keys to CorrelationSum type variable values

	wg        sync.WaitGroup
	wgResults sync.WaitGroup
	// A WaitGroup waits for a collection of goroutines to finish. 
}

type CorrelationSum struct {
	sum   float64
	count int
}

type CorrelationJob struct {
	ImageA *SpeckleImage
	ImageB *SpeckleImage
}

type CorrelationResult struct {
	Correlation float64

	sources []string
	offsets Position
}

func (c *CorrelationCalculator) correlationCounterWorker() {

	ticker := time.Tick(5 * time.Second)

	for {
		select {
		case key, ok := <-c.counter:

			if !ok {
				return
			}

			c.counts[key]++

		case <-ticker:
			log.Println(c)
		}
	}
// count how many work has done?
}

func Initialize(baseDirectory string, numProcessors int) *CorrelationCalculator {
	c := &CorrelationCalculator{}
    // pointer pass variable to function
	c.scanGroupDiff = 0

	c.baseDirectory = baseDirectory
	c.Initialize(numProcessors)

	return c

}

func (c *CorrelationCalculator) Start(files []SpeckleImage) {

	count := 0
	countSame := 0

	for i := 0; i < len(files); i++ {
		for j := 0; j < len(files); j++ {

			if i == j {
				countSame++
			}
			count++

			c.Calculate(&files[i], &files[j])
			// feed all combo files in two to CorrelationJob channel

		}

	}

	c.NoMoreJobs()

	log.Println("Corr Count ", count)
	log.Println("Waiting for correlations to complete")

	c.Wait()

	return

}

func (c *CorrelationCalculator) Initialize(numWorkers int) {

	// we don't know the order in which jobs will be assigned/processed, make the
	// buffer large enough to keep things busy

	runtime.GOMAXPROCS(numWorkers + 4) //sets the maximum number of CPUs
    
	c.InputJobChannel = make(chan CorrelationJob, numWorkers*5) 
	c.OutputResults = make(chan CorrelationResult, numWorkers*5)
    //numWorkers*5 is the buffer capacity
	
	c.counter = make(chan string, 5)

	c.counts = make(map[string]int64)

	c.correlationResult = make(map[string]CorrelationSum) // sum and count

	go c.CorrelationResultReader()
	go c.correlationCounterWorker()

	for i := 0; i < numWorkers; i++ {

		go c.CorrelationWorker()

	}
    // launch one instant of worker per CPU meaning getting ready?
	log.Println("Done Setting Up Correlation Workers/Reader")

}

func (c *CorrelationCalculator) CorrelationWorker() {

	c.wg.Add(1)
	defer c.wg.Done()

	for {

		job, ok := <-c.InputJobChannel

		if !ok {
			return
		}

		o := Position{}

		o.X = job.ImageA.Position.X - job.ImageB.Position.X
		o.Y = job.ImageA.Position.Y - job.ImageB.Position.Y
		o.Z = job.ImageA.Position.Z - job.ImageB.Position.Z

		// drop the job if the x offset is less than zero

		if o.X < 0 {
			c.counter <- "dropped o[1] < 0"
			continue
		}

		// drop the upper half of dx = 0
		if o.X == 0 && o.Y > 0 {
			c.counter <- "dropped o[1] == 0, o[0] > 0"
			continue
		}

		if math.Abs(float64(job.ImageA.ScanGroupId-job.ImageB.ScanGroupId)) > float64(c.scanGroupDiff) {
			c.counter <- "dropped scan group diff"
			continue
		}

		// do something with the job :)
		imageA := job.ImageA.GetNormalizedImage(c.counter)
		imageB := job.ImageB.GetNormalizedImage(c.counter)
        // v is the value, i is the index
		sum := ImagePixel(0)
		for i, v := range imageA {

			sum += v * imageB[i]

		}
    // calculate correlation in this step at this position(one pixel) only
		corr := float64(sum) / float64(len(imageA))

		c.OutputResults <- CorrelationResult{Correlation: corr, offsets: o}
    // CorrelationResult contains value and delta x y positions
	}

}

func (c *CorrelationCalculator) Wait() {

	time.Sleep(1 * time.Second)

	c.wg.Wait()

	// no more results will be sent, so close that channel

	close(c.OutputResults)
	log.Println("Closed Results Channel")

	// wait for the result reader to finish
	c.wgResults.Wait()

	log.Println("Done Waiting")

	return
}

func (c *CorrelationCalculator) Calculate(f1 *SpeckleImage, f2 *SpeckleImage) {

	job := CorrelationJob{f1, f2}

	c.InputJobChannel <- job

	c.counter <- "job submitted"

}

func (c *CorrelationCalculator) NoMoreJobs() {
	close(c.InputJobChannel)
}

func (c *CorrelationCalculator) CorrelationResultReader() {
	c.wgResults.Add(1)
	defer c.wgResults.Done()

	for {

		result, ok := <-c.OutputResults

		if !ok {
			c.WriteToFile()
			log.Println("Result Reader Closing")
			return
		}

		key := getKey(result.offsets)

		a := c.correlationResult[key]
		a.count++
		a.sum += result.Correlation
        // adding data points at the same correlation points? delta x delta y
		c.correlationResult[key] = a

		c.counter <- "correlation result received"

	}

}

func getKey(offsets Position) string {

	key := fmt.Sprintf("%08d %08d", offsets.Y, offsets.X)

	return key
}

func explodeKey(key string) Position {

	offsets := Position{}

	_, err := fmt.Sscanf(key, "%08d %08d", &offsets.Y, &offsets.X)

	if err != nil {
		log.Fatal("Unable to explode key: ", err.Error(), " ", key)
	}

	return offsets
}

// for the stringer interface
func (c *CorrelationCalculator) String() string {

	return fmt.Sprintf("%+v\n", c.counts)
}
