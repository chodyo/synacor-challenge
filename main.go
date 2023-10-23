package main

import (
	"fmt"
	"os"
)

type op int

const (
	halt op = iota
	set
	push
	pop
	eq
	gt
	jmp
	jt
	jf
	add
	mult
	mod
	and
	or
	not
	rmem
	wmem
	call
	ret
	out
	in
	noop
)

const (
	r0 = iota + 32768
	r1
	r2
	r3
	r4
	r5
	r6
	r7
)

const (
	minValue = 0
	maxValue = 32775
)

var (
	memory    []uint16
	registers []uint16
	stack     []uint16

	// memory is a 15 bit address space so ptr should never go above maxValue
	ptr uint16

	// print buffer
	outbuf []byte

	// for nice strings
	ops = [...]string{
		"halt",
		"set",
		"push",
		"pop",
		"eq",
		"gt",
		"jmp",
		"jt",
		"jf",
		"add",
		"mult",
		"mod",
		"and",
		"or",
		"not",
		"rmem",
		"wmem",
		"call",
		"ret",
		"out",
		"in",
		"noop",
	}
)

func init() {
	registers = make([]uint16, 8)
}

func main() {
	readFileToMemory("challenge.bin")

	code := exec()
	if code <= 0 {
		fmt.Printf("%s\n", outbuf) // could get fancy and put this in a defer, would need to watch out because os.Exit is going to skip defer
		os.Exit(code)
	}
}

func readFileToMemory(filename string) {
	// 16 bit architecture (uint16)
	// little-endian pair (low byte, high byte)

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for {
		b := make([]byte, 2)
		_, err = f.ReadAt(b, int64(len(memory)*2))
		if err != nil {
			break
		}

		// high byte
		value := uint16(b[1])
		value <<= 8
		// low byte
		value += uint16(b[0])

		memory = append(memory, value)
	}
}

func exec() int {
	for o := readAndIncrement(); ptr >= minValue && ptr <= maxValue; o = readAndIncrement() {
		switch o {
		case halt:
			// halt: 0
			//   stop execution and terminate the program
			return 0
		case jmp:
			// jmp: 6 a
			//   jump to <a>
			jmpTo := read()
			ptr = uint16(jmpTo)
		case out:
			// out: 19 a
			//   write the character represented by ascii code <a> to the terminal
			c := readAndIncrement()
			outbuf = append(outbuf, byte(c))
		case noop:
			// noop: 21
			//   no operation
			break
		case set:
		case push:
		case pop:
		case eq:
		case gt:
		case jt:
		case jf:
		case add:
		case mult:
		case mod:
		case and:
		case or:
		case not:
		case rmem:
		case wmem:
		case call:
		case ret:
		case in:
			fmt.Println("not implemented")
			return -1
		default:
			fmt.Println("don't know what to do with this number:", o)
			fmt.Println("previous op:", op(memory[ptr-2]), ptr-2)
			fmt.Println("next op:", op(memory[ptr]), ptr)
			return -2
		}
	}

	return -110002 // arbitrary
}

func read() op {
	return op(memory[ptr])
}

func readAndIncrement() op {
	o := op(memory[ptr])
	ptr++
	return o
}

func (o op) String() string {
	if halt <= o && o <= noop {
		return ops[o]
	}
	return fmt.Sprintf("%sop(%d)", "%!", o)
}
