package corr

import "encoding/binary"
import "fmt"
import "log"
import "os"

// bufio.Scanner

func (c *CorrelationCalculator) WriteToFile() {

	filename := fmt.Sprintf("%s%c%s", c.baseDirectory, os.PathSeparator, "correlation.bin")

	log.Println("Saving results to", filename)

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)

	if err != nil {
		log.Fatal("error : os.OpenFile : " + err.Error())
		os.Exit(1)
	}

	defer file.Close()

	log.Println("Writing Results to File")

	for key, CorrSum := range c.correlationResult {

		if CorrSum.count == 0 {
			log.Fatal("Zero count in corr sum.")
		}

		value := float64(CorrSum.sum) / float64(CorrSum.count)

		offsets := explodeKey(key)

		log.Println(offsets, value, CorrSum.sum, CorrSum.count)

		err = binary.Write(file, binary.LittleEndian, &offsets.Y)

		if err != nil {
			log.Fatal("Writing offsets", err)
		}

		err = binary.Write(file, binary.LittleEndian, &offsets.X)

		if err != nil {
			log.Fatal("Writing offsets", err)
		}

		err = binary.Write(file, binary.LittleEndian, &value)

		if err != nil {
			log.Fatal("Write To File", err)
		}

	}

	log.Println("Wrote Results File")
}
