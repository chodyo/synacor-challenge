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
			// set: 1 a b
			//   set register <a> to the value of <b>
		case push:
			// push: 2 a
			//   push <a> onto the stack
		case pop:
			// pop: 3 a
			//   remove the top element from the stack and write it into <a>; empty stack = error
		case eq:
			// eq: 4 a b c
			//   set <a> to 1 if <b> is equal to <c>; set it to 0 otherwise
		case gt:
			// gt: 5 a b c
			//   set <a> to 1 if <b> is greater than <c>; set it to 0 otherwise
		case jt:
			// jt: 7 a b
			//   if <a> is nonzero, jump to <b>
		case jf:
			// jf: 8 a b
			//   if <a> is zero, jump to <b>
		case add:
			// add: 9 a b c
			//   assign into <a> the sum of <b> and <c> (modulo 32768)
		case mult:
			// mult: 10 a b c
			//   store into <a> the product of <b> and <c> (modulo 32768)
		case mod:
			// mod: 11 a b c
			//   store into <a> the remainder of <b> divided by <c>
		case and:
			// and: 12 a b c
			//   stores into <a> the bitwise and of <b> and <c>
		case or:
			// or: 13 a b c
			//   stores into <a> the bitwise or of <b> and <c>
		case not:
			// not: 14 a b
			//   stores 15-bit bitwise inverse of <b> in <a>
		case rmem:
			// rmem: 15 a b
			//   read memory at address <b> and write it to <a>
		case wmem:
			// wmem: 16 a b
			//   write the value from <b> into memory at address <a>
		case call:
			// call: 17 a
			//   write the address of the next instruction to the stack and jump to <a>
		case ret:
			// ret: 18
			//   remove the top element from the stack and jump to it; empty stack = halt
		case in:
			// in: 20 a
			//   read a character from the terminal and write its ascii code to <a>; it can be assumed that once input starts, it will continue until a newline is encountered; this means that you can safely read whole lines from the keyboard and trust that they will be fully read
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
