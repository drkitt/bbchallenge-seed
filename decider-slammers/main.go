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

func recordsAreEquivalent(movingRight bool, pastRecord *Record, currentRecord *Record) bool {
	offset := 0

	// See whether we've modified the tape squares surrounding the old record
	// position
	for pastRecord.Position+offset >= 0 && pastRecord.Position+offset < len(pastRecord.Tape) {

		// The records are automatically considered equivalent if there's a tape
		// square before the previous record that hasn't been visited since the
		// previous record was broken
		if currentRecord.Tape[pastRecord.Position+offset].LastTimeSeen < pastRecord.Time {
			break
		}

		if currentRecord.Tape[currentRecord.Position+offset].Symbol != pastRecord.Tape[pastRecord.Position+offset].Symbol {
			return false
		}

		if movingRight {
			offset -= 1
		} else {
			offset += 1
		}
	}

	// Otherwise, the records are considered equivalent if we weren't able to
	// find any tape squares that are different between the two records
	return true
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

// Prints the status of the LBA at a specific time step
func getStatus(currentTime int, currentState byte, symbolRead byte, tape []TapePosition, currentPosition int) string {
	return fmt.Sprintf("Current time: %d\n%s",
		currentTime, tapeString(tape, currentPosition, currentState))
}

// Tells you whether the given LBA is a translated cycler, and if it is, what its preperiod and period are
func decide(lba bbc.LBA, tapeLength int) (bool, []int) {
	var tape []TapePosition = make([]TapePosition, tapeLength)
	var periods []int
	// Yo dawg I you like state machines so I put a state machine in your state
	// machine so your decider can switch between well-defined states while your
	// LBA can switch between well-defined states
	searchingForPeriod := true
	movingRight := true
	previousPosition := -1
	currentPosition := 0
	nextPosition := currentPosition
	toWrite := byte(0)
	currentState := byte(1)
	currentTime := 0
	maxPositionSeen := -1
	minPositionSeen := tapeLength
	// When we encounter a new tape square, this maps the current state and
	// symbol read to the contents of the tape at the time of reading
	var records map[byte]map[byte][]Record = make(map[byte]map[byte][]Record)
	previousCycleEndTime := 0

	for currentState > 0 {
		symbolRead := tape[currentPosition].Symbol

		if searchingForPeriod {
			fmt.Println(getStatus(currentTime, currentState, symbolRead, tape, currentPosition))

			// Handle a never-before-seen tape square
			if (movingRight && currentPosition > maxPositionSeen) || (!movingRight && currentPosition < minPositionSeen) {
				fmt.Println("New record")

				var record Record
				record.Tape = make([]TapePosition, tapeLength)
				copy(record.Tape, tape)
				record.Time = currentTime
				record.Position = currentPosition

				if _, ok := records[currentState]; !ok {
					records[currentState] = make(map[byte][]Record)
				}

				// We've encountered this symbol in this state before. Are the
				// nearby tape symbols the same as before?
				if _, ok := records[currentState][symbolRead]; ok {
					for _, previousRecord := range records[currentState][symbolRead] {

						fmt.Println("Comparing records:")
						fmt.Println("\t", tapeString(previousRecord.Tape, previousRecord.Position, currentState))
						fmt.Println("\t", tapeString(record.Tape, record.Position, currentState))

						if recordsAreEquivalent(movingRight, &previousRecord, &record) {
							fmt.Println("oh my god it's a translated cycler")
							constantSection := previousRecord.Time - previousCycleEndTime
							period := currentTime - previousRecord.Time

							periods = append(periods, constantSection, period)

							fmt.Println("Moving to edge of tape...")
							searchingForPeriod = false
						}
					}
				}

				records[currentState][symbolRead] = append(records[currentState][symbolRead], record)

				maxPositionSeen = bbc.MaxI(maxPositionSeen, currentPosition)
			}

			if maxPositionSeen > tapeLength || currentPosition < 0 {
				return false, periods
			}

			tape[currentPosition].Seen = true
			tape[currentPosition].LastTimeSeen = currentTime

			fmt.Println()
		} else {
			// Detect hitting the edge of the tape
			if currentPosition == previousPosition {
				fmt.Println("🥺 hi tape edge")
				fmt.Println(getStatus(currentTime, currentState, symbolRead, tape, currentPosition))
				fmt.Println()

				// Remove old records
				for k := range records {
					delete(records, k)
				}

				previousCycleEndTime = currentTime
				searchingForPeriod = true
				movingRight = !movingRight
			}
		}

		// Take a step
		toWrite, currentState, nextPosition = bbc.LbaStep(lba, tapeLength, symbolRead, currentState, currentPosition, currentTime)

		tape[currentPosition].Symbol = toWrite
		previousPosition = currentPosition
		currentPosition = nextPosition
		currentTime += 1
	}

	fmt.Println("Halted at time", currentTime)

	// Record the steps since the end of the last cycle as a constant section
	periods = append(periods, currentTime-previousCycleEndTime)

	// Did you know?
	// Halting translated cyclers with repeating period 1 are called slammers.
	// For more information, see https://www.youtube.com/watch?v=XYq08kJGp4M

	return (len(periods) > 1), periods
}

func main() {
	var maxConstantTerm int = -1
	var constantChampionIndex uint32 = 0
	var maxLinearTerm int = -1
	var linearChampionIndex uint32 = 0

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
	databaseSize = 10

	for i := 0; i < databaseSize; i += 1 {
		lba, error := bbc.GetMachineI(database, i, false)
		if error != nil {
			fmt.Println("Error: ", error)
		}
		fmt.Println(lba.ToAsciiTable(2))

		if isTranslatedCycler, periods := decide(lba, 30); isTranslatedCycler {
			// Create string for the cost function
			costFunction := ""

			constantTerm := 0
			linearTerm := 0
			for i := 0; i < len(periods); i++ {
				if i%2 == 0 {
					constantTerm += periods[i]
				} else {
					linearTerm += periods[i]
				}
			}
			if linearTerm > 1 {
				costFunction += fmt.Sprintf("%d", linearTerm)
			}
			if linearTerm > 0 {
				costFunction += "t"
			}
			if constantTerm > 0 {
				costFunction += fmt.Sprintf(" + %d", constantTerm)
			}

			fmt.Println()
			fmt.Println("Periods:", periods)
			fmt.Println("Cost function:", costFunction)

			if constantTerm > maxConstantTerm {
				maxConstantTerm = constantTerm
				constantChampionIndex = uint32(i)
			}
			if linearTerm > maxLinearTerm {
				maxLinearTerm = linearTerm
				linearChampionIndex = uint32(i)
			}

			var toWrite [4]byte
			binary.BigEndian.PutUint32(toWrite[0:4], uint32(i))
			outputFile.Write(toWrite[:])
		}

		fmt.Println()
		fmt.Println("----------------------------------------")
		fmt.Println()
	}

	if maxLinearTerm > 0 {
		fmt.Printf("Max linear term: %d (machine %d)\nMax constant term: %d (machine %d)\n", maxLinearTerm, linearChampionIndex, maxConstantTerm, constantChampionIndex)
	} else {
		fmt.Println("No translated cyclers found")
	}
}
