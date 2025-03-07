//go:build ignore

package main

import (
	"fmt"
	"os"
)

const (
	modk = 1 << 15
	modm = modk - 1
	regc = 8
)

const (
	op_halt = iota
	op_set
	op_push
	op_pop
	op_eq
	op_gt
	op_jmp
	op_jt
	op_jf
	op_add
	op_mult
	op_mod
	op_and
	op_or
	op_not
	op_rmem
	op_wmem
	op_call
	op_ret
	op_out
	op_in
	op_noop
)

func machine(prog []uint16) {
	// trick: make registers just be after the mem array
	mem := [modk + regc]uint16{}
	stk := []uint16{}
	pi := uint16(0)
	copy(mem[:], prog)
	getval := func(n uint16) uint16 {
		if n < modk {
			return n // get literal
		}
		return mem[n] // get register value
	}
	dbgmd := false
	for {
		if dbgmd {
			switch mem[pi] {
			case op_call:
				fmt.Fprintf(os.Stderr, " / mem[%v] = %v .. %v\n", pi, mem[pi], mem[pi+1:pi+4])
			}
		}
		ins := mem[pi]
		pi++
		switch ins {
		case op_halt:
			return
		case op_set:
			a, b := mem[pi], getval(mem[pi+1])
			pi += 2
			mem[a] = b
		case op_push:
			a := getval(mem[pi])
			pi += 1
			stk = append(stk, a)
		case op_pop:
			a := mem[pi]
			pi += 1
			mem[a] = stk[len(stk)-1]
			stk = stk[:len(stk)-1]
		case op_eq:
			a, b, c := mem[pi], getval(mem[pi+1]), getval(mem[pi+2])
			pi += 3
			if b == c {
				mem[a] = 1
			} else {
				mem[a] = 0
			}
		case op_gt:
			a, b, c := mem[pi], getval(mem[pi+1]), getval(mem[pi+2])
			pi += 3
			if b > c {
				mem[a] = 1
			} else {
				mem[a] = 0
			}
		case op_jmp:
			a := getval(mem[pi])
			pi += 1
			pi = a
		case op_jt:
			a, b := getval(mem[pi]), getval(mem[pi+1])
			pi += 2
			if a != 0 {
				pi = b
			}
		case op_jf:
			a, b := getval(mem[pi]), getval(mem[pi+1])
			pi += 2
			if a == 0 {
				pi = b
			}
		case op_add:
			a, b, c := mem[pi], getval(mem[pi+1]), getval(mem[pi+2])
			pi += 3
			mem[a] = b + c
			mem[a] &= modm // overflow
		case op_mult:
			a, b, c := mem[pi], getval(mem[pi+1]), getval(mem[pi+2])
			pi += 3
			mem[a] = b * c
			mem[a] &= modm // overflow
		case op_mod:
			a, b, c := mem[pi], getval(mem[pi+1]), getval(mem[pi+2])
			pi += 3
			mem[a] = b % c
		case op_and:
			a, b, c := mem[pi], getval(mem[pi+1]), getval(mem[pi+2])
			pi += 3
			mem[a] = b & c
		case op_or:
			a, b, c := mem[pi], getval(mem[pi+1]), getval(mem[pi+2])
			pi += 3
			mem[a] = b | c
		case op_not:
			a, b := mem[pi], getval(mem[pi+1])
			pi += 2
			mem[a] = ^b
			mem[a] &= modm // overflow
		case op_rmem:
			a, b := mem[pi], getval(mem[pi+1])
			pi += 2
			mem[a] = mem[b]
		case op_wmem:
			a, b := getval(mem[pi]), getval(mem[pi+1])
			pi += 2
			mem[a] = b
		case op_call:
			a := getval(mem[pi])
			pi += 1
			stk = append(stk, pi)
			pi = a
		case op_ret:
			pi = stk[len(stk)-1]
			stk = stk[:len(stk)-1]
		case op_out:
			a := getval(mem[pi])
			pi += 1
			fmt.Printf("%c", a)
		case op_in:
			a := mem[pi]
			pi += 1
			var c rune
			fmt.Scanf("%c", &c)
			if c == '%' {
				fmt.Scanf("%s", nil)
				dbgmd = !dbgmd
			}
			if c == '+' {
				fmt.Scanf("%s", nil)
				mem[len(mem)-1] = 1
			}
			if c == '\r' { // windows hack to fix CR
				fmt.Scanf("%c", &c)
				c = '\n'
			}
			mem[a] = uint16(c)
		case op_noop:
		default:
			panic(fmt.Sprintf("unimplemented instr: mem[%v] = %v", pi, ins))
		}
	}
}

func bytesTouint16(b []byte) []uint16 {
	r := make([]uint16, len(b)/2)
	for i := 0; i < len(b); i += 2 {
		r[i/2] = (uint16(b[i+1]) << 8) | uint16(b[i]) // low endian binary format
	}
	return r
}

func main() {
	progf := os.Args[1]
	prog, err := os.ReadFile(progf)
	if err != nil {
		panic(err)
	}
	machine(bytesTouint16(prog))
}
