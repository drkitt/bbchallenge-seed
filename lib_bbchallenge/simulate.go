// Here we simulate TMs in Go
package bbchallenge

// #cgo CFLAGS: -g -Wall -O3
// #include "simulate.h"
import "C"
import (
	"fmt"
	"strconv"

	tabulate "github.com/rgeoghegan/tabulate"
)

// We currently work with machines that have at most MAX_STATES states
const MAX_STATES = 5

// Name of halting state
const H = 6

const R = 0
const L = 1

type HaltStatus byte

const (
	HALT HaltStatus = iota
	NO_HALT
	UNDECIDED_TIME
	UNDECIDED_SPACE
)

var BBtUpperBound int

func MaxI(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func MinI(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// We are considering <= 5-state 2-symbol TMs
// The TM:
//
// +---+-----------+-----+
// | - |     0     |  1  |
// +---+-----------+-----+
// | A | 1RB       | 1RH |
// | B | 1LB       | 0RC |
// | C | 1LC       | 1LA |
// | D | undefined | 1RA |
// +---+-----------+-----+
//
// Is encoded by the array:
// 1, 0, 2, 1, 1, 6, 1, 1, 2, 0, 0, 3, 1, 1, 3  1, 1, 1, 0, 0, 0, 1, 0, 1
// 1, R, B, 1, R, H, 1, L, B, 0, R, C, 1, L, C, 1, L, A, -, -, -, 1, R, A

type TM [2 * MAX_STATES * 3]byte

func tmTransitionToStr(b1 byte, b2 byte, b3 byte) (toRet string) {

	if b3 == 0 {
		return "???"
	}

	toRet = strconv.Itoa(int(b1))

	if b2 == 0 {
		toRet += "R"
	} else {
		toRet += "L"
	}

	toRet += string(rune(int('A') + int(b3) - 1))

	return toRet
}

func (tm TM) ToAsciiTable(nbStates byte) (toRet string) {

	var table [][]string

	for i := byte(0); i < nbStates; i += 1 {

		table = append(table, []string{string(rune(int('A') + int(i))),
			tmTransitionToStr(tm[6*i], tm[6*i+1], tm[6*i+2]),
			tmTransitionToStr(tm[6*i+3], tm[6*i+4], tm[6*i+5])})
	}

	layout := &tabulate.Layout{Headers: []string{"-", "0", "1"}, Format: tabulate.SimpleFormat}
	asText, _ := tabulate.Tabulate(
		table, layout,
	)

	return asText
}

// Simulates the input TM from blank input
// and state 1.
// Returns undetermined, state, read with:
// - halting status (HaltStatus)
// - state (byte): State number of undetermined transition if reached
// - read (byte): Read symbol of undetermined transition if reached
// - steps count
// - space count
func simulate(tm TM, limitTime int, limitSpace int) (HaltStatus, byte, byte, int, int) {
	var tape = make([]byte, limitSpace)

	max_pos := 0
	min_pos := limitSpace - 1
	curr_head := 0

	var curr_state byte = 1

	steps_count := 0

	var state_seen [MAX_STATES]bool
	var nbStateSeen byte

	var read byte

	for curr_state != H {

		if !state_seen[curr_state-1] {
			nbStateSeen += 1
		}
		state_seen[curr_state-1] = true

		if steps_count > BBtUpperBound {
			return NO_HALT, 0, 0, steps_count, max_pos - min_pos + 1
		}
		if steps_count > limitTime {
			return UNDECIDED_TIME, 0, 0, steps_count, max_pos - min_pos + 1
		}
		// We no longer use the space limit, since that's built into the model now.

		min_pos = MinI(min_pos, curr_head)
		max_pos = MaxI(max_pos, curr_head)

		read = tape[curr_head]

		tm_transition := 6*(curr_state-1) + 3*read
		write := tm[tm_transition]
		move := tm[tm_transition+1]
		next_state := tm[tm_transition+2]

		// undefined transition
		if next_state == 0 {
			return HALT, curr_state, read,
				steps_count + 1, max_pos - min_pos + 1
		}

		tape[curr_head] = write

		// Prevent tape head from moving beyond the edges of the tape
		if move == R && curr_head < limitSpace-1 {
			curr_head += 1
		} else if move == L && curr_head > 0 {
			curr_head -= 1
		}

		steps_count += 1
		curr_state = next_state
	}

	return HALT, H, read,
		steps_count, max_pos - min_pos + 1
}

// Wrapper for the C simulation code in order to have same API as Go code
func simulate_C_wrapper(tm TM, limitTime int, limitSpace int) (HaltStatus, byte, byte, int, int) {
	end_state := C.uchar(0)
	read := C.uchar(0)
	steps_count := C.int(0)
	space_count := C.int(0)

	halt_status := C.simulate((*C.uchar)(&tm[0]), C.int(limitTime), C.int(limitSpace), &end_state, &read, &steps_count, &space_count)

	return HaltStatus(halt_status), byte(end_state), byte(read), int(steps_count), int(space_count)
}

// Useful for debugging
func printTM(nbStates byte, tm TM) {
	for i := 0; i < int(nbStates); i += 1 {
		for j := 0; j <= 1; j += 1 {
			fmt.Printf("%d%d%d ", tm[6*i+3*j], tm[6*i+3*j+1], tm[6*i+3*j+2])
		}
		fmt.Print("\n")
	}
	fmt.Println()
}
