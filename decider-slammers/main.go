// Determines whether a halting LBA is a translated cycler

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

// Represents a single square on the tape along with some metadata
type TapePosition struct {
	Symbol       byte
	LastTimeSeen int
	Seen         bool
}

// Keeps track of tape contents when the machine reaches a tape sqaure that it
// hasn't reached before
type Record struct {
	Tape     []TapePosition
	Time     int
	Position int
}

// Takes a number that represents a state internally and returns the letter
// that represents that state in printing
func stateToLetter(state byte) rune {
	return rune(state + 'A' - 1)
}

// Returns a string representation of the tape
func tapeString(tape []TapePosition, currentPosition int, currentState byte) string {
	var result string = ""

	for i, position := range tape {
		if i == currentPosition {
			result += fmt.Sprintf("%c[%d]", stateToLetter(currentState), position.Symbol)
		} else {
			result += fmt.Sprintf(" %d ", position.Symbol)
		}
	}

	return result
}

func decide(lba bbc.LBA, tapeLength int) bool {
	var tape []TapePosition = make([]TapePosition, tapeLength)
	currentPosition := 0
	//nextPosition := currentPosition
	//toWrite := byte(0)
	currentState := byte(1)
	currentTime := 0
	//maxPositionSeen := 0

	// When we encounter a new tape square, this maps the current state and
	// symbol read to the contents of the tape at the time of reading
	//var records map[byte]map[byte][]Record = make(map[byte]map[byte][]Record)

	for currentState > 0 {
		symbolRead := tape[currentPosition].Symbol
		fmt.Printf("Current time: %d\nCurrent state: %c\nSymbol read: %d\nTape:\n%s\n",
			currentTime, stateToLetter(currentState), symbolRead, tapeString(tape, currentPosition, currentState))
		currentState = 0
	}

	// Did you know? Halting translated cyclers that enter their post-period
	// the first time they encounter a tape edge are called slammers. For more
	// information, see https://www.youtube.com/watch?v=XYq08kJGp4M

	// up next: the slammers
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

	// Oh man what happened here?
	databaseSize = 1

	for i := 0; i < databaseSize; i += 1 {
		lba, error := bbc.GetMachineI(database, i, false)
		if error != nil {
			fmt.Println("Error: ", error)
		}
		fmt.Println(lba.ToAsciiTable(2))

		if decide(lba, 30) {
			var toWrite [4]byte
			binary.BigEndian.PutUint32(toWrite[0:4], uint32(i))
			outputFile.Write(toWrite[:])
		}
	}
}
