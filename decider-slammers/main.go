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

	// See whether we've modified the tape squares behind the old record
	// position
	for pastRecord.Position+offset >= 0 && pastRecord.Position+offset < len(pastRecord.Tape) {

		// The records are automatically considered equivalent if there's a tape
		// square behind the previous record that hasn't been visited since the
		// previous record was broken
		if currentRecord.Tape[pastRecord.Position+offset].LastTimeSeen < pastRecord.Time {
			break
		}

		if currentRecord.Tape[currentRecord.Position+offset].Symbol != pastRecord.Tape[pastRecord.Position+offset].Symbol {
			return false
		}

		// (the meaning of "behind" depends on which direction the tape head is moving in)
		if movingRight {
			offset -= 1
		} else {
			offset += 1
		}
	}

	// Now see if the tape squares ahead of the records are equivalent
	offset = 0
	for currentRecord.Position+offset < len(currentRecord.Tape) && currentRecord.Position+offset >= 0 {
		if currentRecord.Tape[currentRecord.Position+offset].Symbol != pastRecord.Tape[pastRecord.Position+offset].Symbol {
			return false
		}

		if movingRight {
			offset += 1
		} else {
			offset -= 1
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

// Tells you whether the given LBA is a translated cycler, and if it is, the coefficient and constant of its cost function. (Translated cyclers run in linear time)
func decide(lba bbc.LBA, tapeLength int) (bool, int, int) {
	var tape []TapePosition = make([]TapePosition, tapeLength)
	// The max/min position seen by the machine so far in each state
	var maxPositionSeen map[byte]int = make(map[byte]int)
	var minPositionSeen map[byte]int = make(map[byte]int)
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
	// When we encounter a new tape square, this maps the current state and
	// symbol read to the contents of the tape at the time of reading
	var records map[byte]map[byte][]Record = make(map[byte]map[byte][]Record)
	previousCycleEndTime := 0
	// Used to construct the machine's cost function
	coefficient, constant := 0, 0

	for currentState > 0 {
		symbolRead := tape[currentPosition].Symbol

		// Detect hitting the edge of the tape
		if currentPosition == previousPosition {
			fmt.Println("🥺 hi tape edge")

			// Remove old records
			for k := range records {
				delete(records, k)
			}
			for k := range maxPositionSeen {
				delete(maxPositionSeen, k)
			}
			for k := range minPositionSeen {
				delete(minPositionSeen, k)
			}

			// If we had already detected a cycle and were just waiting to reach
			// the edge, prepare to detect the next cycle
			if !searchingForPeriod {
				previousCycleEndTime = currentTime
				movingRight = !movingRight
			}
			searchingForPeriod = true
		}

		if searchingForPeriod {
			// Handle a never-before-seen tape square
			if _, ok := maxPositionSeen[currentState]; !ok {
				maxPositionSeen[currentState] = -1
			}
			if _, ok := minPositionSeen[currentState]; !ok {
				minPositionSeen[currentState] = tapeLength
			}
			if (movingRight && currentPosition > maxPositionSeen[currentState]) || (!movingRight && currentPosition < minPositionSeen[currentState]) {
				fmt.Println("New record")
				fmt.Println(getStatus(currentTime, currentState, symbolRead, tape, currentPosition))

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
							period := currentTime - previousRecord.Time
							constantSection := previousRecord.Time - previousCycleEndTime
							fmt.Printf("oh my god it's a translated cycler (preperiod: %d, period: %d)\n", constantSection, period)

							// Go doesn't have an absolute value function for integers :/
							distanceTraveledInPeriod := previousRecord.Position - record.Position
							if distanceTraveledInPeriod < 0 {
								distanceTraveledInPeriod = -distanceTraveledInPeriod
							}

							coefficient += period / distanceTraveledInPeriod
							constant += constantSection
							// Adjust the contant (again) to avoid overcounting.
							// Specifically, take into account that the machine
							// may not have traversed the entire tape while in
							// its cycle; if it was already partway through the
							// tape when the cycle began, we subtract from the
							// constant the distance it traveled before the
							// cycle began.
							distanceTraveledInConstantSection := 0
							if movingRight {
								distanceTraveledInConstantSection = previousRecord.Position
							} else {
								distanceTraveledInConstantSection = (tapeLength - 1) - previousRecord.Position
							}
							constant -= (period / distanceTraveledInPeriod) * distanceTraveledInConstantSection

							// It gets worse: Since the tape head starts at the
							// first tape square, we should actually multiply
							// the coefficient by (tapeLength - 1) to account
							// for the fact that the machine doesn't have to
							// move to that square.
							constant -= period / distanceTraveledInPeriod
							// And then we add 1 to to account for when the tape
							// head bonks against the edge of the tape.
							constant += 1

							fmt.Println("Moving to edge of tape...")
							searchingForPeriod = false
						}
					}
				}

				records[currentState][symbolRead] = append(records[currentState][symbolRead], record)

				maxPositionSeen[currentState] = bbc.MaxI(maxPositionSeen[currentState], currentPosition)
				minPositionSeen[currentState] = bbc.MinI(minPositionSeen[currentState], currentPosition)

				fmt.Println()
			}

			if maxPositionSeen[currentState] > tapeLength || currentPosition < 0 {
				fmt.Println("Not a translated cycler")
				return false, -1, -1
			}

			tape[currentPosition].LastTimeSeen = currentTime
		}

		// Take a step
		toWrite, currentState, nextPosition = bbc.LbaStep(lba, tapeLength, symbolRead, currentState, currentPosition, currentTime)

		tape[currentPosition].Symbol = toWrite
		previousPosition = currentPosition
		currentPosition = nextPosition
		currentTime += 1
	}

	// Record the steps since the end of the last cycle as a constant section
	constant += currentTime - previousCycleEndTime

	fmt.Printf("Halted at time %d (coefficient: %d, constant: %d)\n", currentTime, coefficient, constant)

	// Did you know?
	// Halting translated cyclers with repeating period 1 are called slammers.
	// For more information, see https://www.youtube.com/watch?v=XYq08kJGp4M

	if currentTime != coefficient*tapeLength+constant {
		log.Printf("❗️ Warning: the machine did not halt in the expected time (cost function gives runtime of %d, but the machine halted at time %d)\n", coefficient*tapeLength+constant, currentTime)
	}

	return coefficient > 0 || constant > 0, coefficient, constant
}

func main() {
	var maxConstantTerm int = -1
	var constantChampionIndex uint32 = 0
	var maxCoefficient int = -1
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
	databaseSize = 128

	for i := 0; i < databaseSize; i += 1 {
		lba, error := bbc.GetMachineI(database, i, false)
		if error != nil {
			fmt.Println("Error: ", error)
		}
		fmt.Println("Machine", i)
		fmt.Println(lba.ToAsciiTable(2))

		if isTranslatedCycler, coefficient, constantTerm := decide(lba, 13); isTranslatedCycler {
			// Create string for the cost function
			costFunction := ""
			if coefficient > 1 {
				costFunction += fmt.Sprintf("%d", coefficient)
			}
			if coefficient > 0 {
				costFunction += "t"
			}
			if constantTerm > 0 {
				costFunction += fmt.Sprintf(" + %d", constantTerm)
			}

			fmt.Println()
			fmt.Println("Cost function:", costFunction)

			if constantTerm > maxConstantTerm {
				maxConstantTerm = constantTerm
				constantChampionIndex = uint32(i)
			}
			if coefficient > maxCoefficient {
				maxCoefficient = coefficient
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

	if maxCoefficient > 0 {
		fmt.Printf("Max coefficient: %d (machine %d)\nMax constant term: %d (machine %d)\n", maxCoefficient, linearChampionIndex, maxConstantTerm, constantChampionIndex)
	} else {
		fmt.Println("No translated cyclers found")
	}
}
