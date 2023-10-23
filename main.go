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
	memory    []op
	registers [8]op
	stack     []op

	// memory is a 15 bit address space so ptr should never go above maxValue
	ptr op

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

func main() {
	// _ = writeTestFile("test.bin", []op{9, 32768, 32769, 4, 19, 32768})

	// readFileToMemory("test.bin")
	readFileToMemory("challenge.bin")

	code := exec()
	if code <= 0 {
		fmt.Println("============== out ==============")
		fmt.Printf("%s\n", outbuf) // could get fancy and put this in a defer, would need to watch out because os.Exit is going to skip defer
		os.Exit(code)
	}
}

func writeTestFile(filename string, littleEndianBytes []op) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, b := range littleEndianBytes {
		left := byte(b)
		right := byte(b >> 8)
		_, _ = f.Write([]byte{left, right})
	}

	return nil
}

func readFileToMemory(filename string) {
	// 16 bit architecture (op)
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
		value := op(b[1])
		value <<= 8
		// low byte
		value += op(b[0])

		memory = append(memory, value)
	}
}

func exec() int {
	for o := readAndIncrement(); ptr >= minValue && int(ptr) < len(memory); o = readAndIncrement() {
		switch o {
		case halt:
			// halt: 0
			//   stop execution and terminate the program
			return 0
		case set:
			// set: 1 a b
			//   set register <a> to the value of <b>
			a := readAddrAndIncrement()
			b := readAndIncrement()
			setAddr(a, b)
		case push:
			// push: 2 a
			//   push <a> onto the stack
			a := readAndIncrement()
			stack = append(stack, a)
		case pop:
			// pop: 3 a
			//   remove the top element from the stack and write it into <a>; empty stack = error
			a := readAddrAndIncrement()
			if len(stack) == 0 {
				return 0
			}
			i := len(stack) - 1
			val := stack[i]
			stack = stack[0:i]
			setAddr(a, val)
		case eq:
			// eq: 4 a b c
			//   set <a> to 1 if <b> is equal to <c>; set it to 0 otherwise
			a := readAddrAndIncrement()
			b := readAndIncrement()
			c := readAndIncrement()
			val := op(0)
			if b == c {
				val = 1
			}
			setAddr(a, val)
		case gt:
			// gt: 5 a b c
			//   set <a> to 1 if <b> is greater than <c>; set it to 0 otherwise
			a := readAddrAndIncrement()
			b := readAndIncrement()
			c := readAndIncrement()
			val := op(0)
			if b > c {
				val = 1
			}
			setAddr(a, val)
		case jmp:
			// jmp: 6 a
			//   jump to <a>
			a := readAndIncrement()
			jump(a)
		case jt:
			// jt: 7 a b
			//   if <a> is nonzero, jump to <b>
			a := readAndIncrement()
			b := readAndIncrement()
			if a != 0 {
				jump(b)
			}
		case jf:
			// jf: 8 a b
			//   if <a> is zero, jump to <b>
			a := readAndIncrement()
			b := readAndIncrement()
			if a == 0 {
				jump(b)
			}
		case add:
			// add: 9 a b c
			//   assign into <a> the sum of <b> and <c> (modulo 32768)
			a := readAddrAndIncrement()
			b := readAndIncrement()
			c := readAndIncrement()
			sum := (int(b) + int(c)) % r0
			setAddr(a, op(sum))
		case mult:
			// mult: 10 a b c
			//   store into <a> the product of <b> and <c> (modulo 32768)
			a := readAddrAndIncrement()
			b := readAndIncrement()
			c := readAndIncrement()
			product := (int(b) * int(c)) % r0
			setAddr(a, op(product))
		case mod:
			// mod: 11 a b c
			//   store into <a> the remainder of <b> divided by <c>
			a := readAddrAndIncrement()
			b := readAndIncrement()
			c := readAndIncrement()
			mod := b % c
			setAddr(a, mod)
		case and:
			// and: 12 a b c
			//   stores into <a> the bitwise and of <b> and <c>
			a := readAddrAndIncrement()
			b := readAndIncrement()
			c := readAndIncrement()
			and := (int(b) & int(c)) % r0
			setAddr(a, op(and))
		case or:
			// or: 13 a b c
			//   stores into <a> the bitwise or of <b> and <c>
			a := readAddrAndIncrement()
			b := readAndIncrement()
			c := readAndIncrement()
			or := (int(b) | int(c)) % r0
			setAddr(a, op(or))
		case not:
			// not: 14 a b
			//   stores 15-bit bitwise inverse of <b> in <a>
			a := readAddrAndIncrement()
			b := readAndIncrement()
			not := (0xFFFF ^ b) % r0
			setAddr(a, not)
		case rmem:
			// rmem: 15 a b
			//   read memory at address <b> and write it to <a>
			a := readAddrAndIncrement()
			b := readAndIncrement()
			val := memory[b]
			setAddr(a, val)
		case wmem:
			// wmem: 16 a b
			//   write the value from <b> into memory at address <a>
			a := readAndIncrement()
			b := readAndIncrement()
			memory[a] = b
		case call:
			// call: 17 a
			//   write the address of the next instruction to the stack and jump to <a>
			a := readAndIncrement()
			stack = append(stack, ptr)
			jump(a)
		case ret:
			// ret: 18
			//   remove the top element from the stack and jump to it; empty stack = halt
			if len(stack) == 0 {
				return 0
			}
			i := len(stack) - 1
			val := stack[i] % r0
			stack = stack[0:i]
			jump(val)
		case out:
			// out: 19 a
			//   write the character represented by ascii code <a> to the terminal
			a := readAndIncrement()
			outbuf = append(outbuf, byte(a))
		case noop:
			// noop: 21
			//   no operation
			break
		case in:
			// in: 20 a
			//   read a character from the terminal and write its ascii code to <a>; it can be assumed that once input starts, it will continue until a newline is encountered; this means that you can safely read whole lines from the keyboard and trust that they will be fully read
			fmt.Println("not implemented", o)
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

func readAddr() op {
	o := memory[ptr]
	fmt.Printf("%d: %s\n", ptr, o)
	return o
}

func readAddrAndIncrement() op {
	o := readAddr()
	ptr++
	return o
}

func read() op {
	if 0 > ptr || ptr > r7 {
		return halt
	}

	o := readAddr()

	if r0 <= o && o <= r7 {
		o = registers[o%r0]
	}
	return o
}

func readAndIncrement() op {
	o := read()
	ptr++
	return o
}

func jump(jmpTo op) {
	ptr = jmpTo
}

func setAddr(r op, val op) {
	r %= r0
	registers[r] = val
}

func (o op) String() string {
	if halt <= o && o <= noop {
		return fmt.Sprintf("%s (%d)", ops[o], o)
	}

	if r0 <= o && o <= r7 {
		return fmt.Sprintf("%d -> r%d <%d>", o, o%r0, registers[o%r0])
	}

	return fmt.Sprintf("%sop(%d)", "%!", o)
}
