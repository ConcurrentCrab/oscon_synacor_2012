//go:build ignore

package main

import (
	"fmt"
	"os"
)

func slicesMap[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

type lineScan struct {
	b *[]byte
}

func (s *lineScan) Scan(state fmt.ScanState, verb rune) error {
	t, err := state.Token(false, func(r rune) bool { return r != '\n' })
	if err != nil || len(t) == 0 {
		return err
	}
	if t[len(t)-1] == '\r' { // Windows workaround
		t = t[:len(t)-1]
	}
	*s.b = append(*s.b, t...)
	return nil
}

func readLine() string {
	var b []byte
	fmt.Scan(&lineScan{&b})
	return string(b)
}

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

var opnames = map[uint16]string{
	op_halt: "op_halt",
	op_set:  "op_set",
	op_push: "op_push",
	op_pop:  "op_pop",
	op_eq:   "op_eq",
	op_gt:   "op_gt",
	op_jmp:  "op_jmp",
	op_jt:   "op_jt",
	op_jf:   "op_jf",
	op_add:  "op_add",
	op_mult: "op_mult",
	op_mod:  "op_mod",
	op_and:  "op_and",
	op_or:   "op_or",
	op_not:  "op_not",
	op_rmem: "op_rmem",
	op_wmem: "op_wmem",
	op_call: "op_call",
	op_ret:  "op_ret",
	op_out:  "op_out",
	op_in:   "op_in",
	op_noop: "op_noop",
}

var opargs = map[uint16]uint16{
	op_halt: 0,
	op_set:  2,
	op_push: 1,
	op_pop:  1,
	op_eq:   3,
	op_gt:   3,
	op_jmp:  1,
	op_jt:   2,
	op_jf:   2,
	op_add:  3,
	op_mult: 3,
	op_mod:  3,
	op_and:  3,
	op_or:   3,
	op_not:  2,
	op_rmem: 2,
	op_wmem: 2,
	op_call: 1,
	op_ret:  0,
	op_out:  1,
	op_in:   1,
	op_noop: 0,
}

type state struct {
	// trick: make registers just be after the mem array
	mem [modk + regc]uint16
	stk []uint16
	pi  uint16
	fin bool
}

func machine(st *state) {
	getval := func(n uint16) uint16 {
		if n < modk {
			return n // get literal
		}
		return st.mem[n] // get register value
	}
	if st.fin {
		return
	}
	ins := st.mem[st.pi]
	st.pi++
	switch ins {
	case op_halt:
		st.fin = true
	case op_set:
		a, b := st.mem[st.pi], getval(st.mem[st.pi+1])
		st.pi += 2
		st.mem[a] = b
	case op_push:
		a := getval(st.mem[st.pi])
		st.pi += 1
		st.stk = append(st.stk, a)
	case op_pop:
		a := st.mem[st.pi]
		st.pi += 1
		st.mem[a] = st.stk[len(st.stk)-1]
		st.stk = st.stk[:len(st.stk)-1]
	case op_eq:
		a, b, c := st.mem[st.pi], getval(st.mem[st.pi+1]), getval(st.mem[st.pi+2])
		st.pi += 3
		if b == c {
			st.mem[a] = 1
		} else {
			st.mem[a] = 0
		}
	case op_gt:
		a, b, c := st.mem[st.pi], getval(st.mem[st.pi+1]), getval(st.mem[st.pi+2])
		st.pi += 3
		if b > c {
			st.mem[a] = 1
		} else {
			st.mem[a] = 0
		}
	case op_jmp:
		a := getval(st.mem[st.pi])
		st.pi += 1
		st.pi = a
	case op_jt:
		a, b := getval(st.mem[st.pi]), getval(st.mem[st.pi+1])
		st.pi += 2
		if a != 0 {
			st.pi = b
		}
	case op_jf:
		a, b := getval(st.mem[st.pi]), getval(st.mem[st.pi+1])
		st.pi += 2
		if a == 0 {
			st.pi = b
		}
	case op_add:
		a, b, c := st.mem[st.pi], getval(st.mem[st.pi+1]), getval(st.mem[st.pi+2])
		st.pi += 3
		st.mem[a] = b + c
		st.mem[a] &= modm // overflow
	case op_mult:
		a, b, c := st.mem[st.pi], getval(st.mem[st.pi+1]), getval(st.mem[st.pi+2])
		st.pi += 3
		st.mem[a] = b * c
		st.mem[a] &= modm // overflow
	case op_mod:
		a, b, c := st.mem[st.pi], getval(st.mem[st.pi+1]), getval(st.mem[st.pi+2])
		st.pi += 3
		st.mem[a] = b % c
	case op_and:
		a, b, c := st.mem[st.pi], getval(st.mem[st.pi+1]), getval(st.mem[st.pi+2])
		st.pi += 3
		st.mem[a] = b & c
	case op_or:
		a, b, c := st.mem[st.pi], getval(st.mem[st.pi+1]), getval(st.mem[st.pi+2])
		st.pi += 3
		st.mem[a] = b | c
	case op_not:
		a, b := st.mem[st.pi], getval(st.mem[st.pi+1])
		st.pi += 2
		st.mem[a] = ^b
		st.mem[a] &= modm // overflow
	case op_rmem:
		a, b := st.mem[st.pi], getval(st.mem[st.pi+1])
		st.pi += 2
		st.mem[a] = st.mem[b]
	case op_wmem:
		a, b := getval(st.mem[st.pi]), getval(st.mem[st.pi+1])
		st.pi += 2
		st.mem[a] = b
	case op_call:
		a := getval(st.mem[st.pi])
		st.pi += 1
		st.stk = append(st.stk, st.pi)
		st.pi = a
	case op_ret:
		st.pi = st.stk[len(st.stk)-1]
		st.stk = st.stk[:len(st.stk)-1]
	case op_out:
		a := getval(st.mem[st.pi])
		st.pi += 1
		fmt.Printf("%c", a)
	case op_in:
		a := st.mem[st.pi]
		st.pi += 1
		var c rune
		fmt.Scanf("%c", &c)
		if c == '\r' { // windows hack to fix CR
			fmt.Scanf("%c", &c)
			c = '\n'
		}
		st.mem[a] = uint16(c)
	case op_noop:
	default:
		panic(fmt.Sprintf("unimplemented instr: mem[%v] = %v", st.pi, ins))
	}
}

func debugger(st *state) {
	argsString := func(s []uint16) []string {
		return slicesMap(s, func(e uint16) string {
			if e < modk {
				return fmt.Sprintf("%d", e)
			}
			return fmt.Sprintf("<%d>", e-modk)
		})
	}
	doBreakDef := func() bool {
		switch st.mem[st.pi] {
		case op_out:
			return false
		}
		return true
	}
	doBreak := doBreakDef
	for !st.fin {
		if doBreak() {
			cmd := readLine()
			switch cmd {
			case "w": // where
				fmt.Printf("pi: %v\nstk: %v\nregs: %v\n", st.pi, st.stk, st.mem[modk:])
				continue
			case "d": // disassemble
				pid := st.pi
			disloop:
				for {
					pii := st.mem[pid]
					pia := opargs[pii]
					fmt.Printf("%v: %v\n", opnames[pii], argsString(st.mem[pid+1:pid+1+pia]))
					pid += 1 + pia
					switch pii {
					case op_halt, op_ret:
						break disloop
					}
				}
				continue
			case "cf": // continue function
				cs := 1
				doBreak = func() bool {
					cso := cs
					switch st.mem[st.pi] {
					case op_call:
						cs += 1
					case op_ret:
						cs -= 1
					}
					if cso == 0 {
						doBreak = doBreakDef
						return true
					}
					return false
				}
				continue
			case "":
			}
		}
		machine(st)
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
	st := state{}
	copy(st.mem[:], bytesTouint16(prog))
	debugger(&st)
}
