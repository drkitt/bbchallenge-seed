package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"

	bbc "github.com/drkitt/bbchallenge-go"
)

// Where to find the file that contains the halting machines
const DATABASE_PATH = "./run_2025-01-14_12-25-37_halting"

// Visual demonstration of how slammers work: https://www.youtube.com/watch?v=XYq08kJGp4M
func decide(lba bbc.LBA) bool {
	// wait wait what's an lba
	return false
}

func main() {
	database, error := os.ReadFile(DATABASE_PATH)

	if error != nil {
		fmt.Println(error)
		os.Exit(-1)
	}

	fmt.Println("Hi 🥺 :3")

	// The BBChallenge deciders subtracted 1 from this quantity to account for
	// the 30-byte header, but we don't have a header so we don't do that
	databaseSize := (len(database) / 30)
	fmt.Println(databaseSize)

	// Create output file
	outputFile, error := os.OpenFile("output/"+bbc.GetRunName(),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if error != nil {
		log.Fatal(error)
	}
	defer outputFile.Close()

	// Not gonna add multithreading until it gets annoyingly slow 😤
	for i := 0; i < databaseSize; i += 1 {
		lba, error := bbc.GetMachineI(database, i, false)
		if error != nil {
			fmt.Println("Error: ", error)
		}
		fmt.Println(lba.ToAsciiTable(2))

		if decide(lba) {
			var toWrite [4]byte
			binary.BigEndian.PutUint32(toWrite[0:4], uint32(i))
			outputFile.Write(toWrite[:])
		}
	}
}
