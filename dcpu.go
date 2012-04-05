package dcpu

import (
	"io"
	"os"
)

type word uint16

type register byte

// These register constants can be used to access the DCPU16's registers.
const (
	A register = iota
	B
	C
	X
	Y
	Z
	I
	J
)

// An emulator for Notch's DCPU16… thing. See http://0x10c.com/doc/dcpu-16.txt for the spec.
// Memory, Registers, PC, SP and O are public in order to provide some ability to access and
// manipulate state.
type DCPU16 struct {
	Memory    [0x10000]word // 0x10000 words of memory!
	Registers [8]word       // Access them using the register constants above, e.g. Registers[C]
	PC        word          // Program counter; tracks currently executing instruction
	SP        word          // Stack pointer; points at the bottom of the downward-growing stack
	O         word          // Overflow; indicates the impact of assorted arithmetic operations
	skipping  bool          // True when the next instruction is to be skipped, false otherwise
	cycles    uint          // Number of "cycles" executed.
}

// Creates a new DCPU16.
func New() *DCPU16 {
	cpu := new(DCPU16)
	cpu.PC = 0
	cpu.SP = 0xFFFF
	return cpu
}

// Loads a memory image from a file on disk. Memory images are assumed to start from address
// zero. Anything beyond the loaded image will contain its old contents.
func (this *DCPU16) LoadImage(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return this.LoadStream(f)
}

// Loads a memory image from an io.Reader. Memory images are assumed to start from address
// zero. Anything beyond the loaded image will contain its old contents.
func (this *DCPU16) LoadStream(f io.Reader) error {
	buffer := make([]byte, 0x20000)

	if _, err := f.Read(buffer); err != nil {
		return err
	}

	for i := 0; i < 0x10000; i++ {
		this.Memory[i] = word(buffer[i*2])<<8 | word(buffer[i*2+1])
	}
	return nil
}

// Runs the machine indefinitely. You can also call ExecuteCycle repeatedly yourself to step through.
func (this *DCPU16) Run() {
	for {
		this.ExecuteCycle()
	}
}

// Returns the number of Notch-decreed "cycles" executed so far.
func (this *DCPU16) Cycles() uint {
	return this.cycles
}

// Executes one instruction (or skips over one instruction, if necessary). Can be called in a loop
// to execute a program. See also: Run().
func (this *DCPU16) ExecuteCycle() {
	operation := this.Memory[this.PC]
	this.PC++
	opcode := operation & 0xF
	v1 := operation >> 4 & 0x3F
	v2 := operation >> 10

	// Zero is for extended opcodes.
	if opcode != 0 {
		if !this.skipping {
			a := this.resolve(v1)
			b := this.resolve(v2)
			basicOpcodes[opcode](this, a, b) // A must be a pointer because it's altered.
		} else {
			this.skipValue(v1)
			this.skipValue(v2)
			this.skipping = false
		}
	} else {
		opcode = v1
		if !this.skipping {
			a := this.resolve(v2)
			extendedOpcodes[opcode](this, a)
		} else {
			this.skipValue(v2)
			this.skipping = false
		}
	}
}

// Given a 6-bit "value", determines what the value should actually refer to and returns
// a pointer to it.
//
// Why a pointer? Because the spec says that we need to modify the values it returns.
// However, some values are immediate and modifications to them should be silently discarded
// (or just make no sense). For those we declare a local variable and return a pointer to that.
// Go makes sure this stays in memory for us.
func (this *DCPU16) resolve(what word) *word {
	switch {
	case what < 0x08: // Register
		return &this.Registers[what]
	case what < 0x0f: // [Register]
		this.cycles++
		return &this.Memory[this.Registers[what-0x08]]
	case what < 0x18: // [Register + word]
		this.cycles++
		value := &this.Memory[this.Registers[what-0x0f]+this.Memory[this.PC]]
		this.PC++
		return value
	case what > 0x1f: // Immediate byte
		// Have to use a variable because we return a pointer.
		immediate := what - 0x20
		return &immediate
	}
	switch what {
	case 0x18: // Pop
		value := &this.Memory[this.SP]
		this.SP++
		return value
	case 0x19: // Peek
		return &this.Memory[this.SP]
	case 0x1a: // Push
		this.SP--
		return &this.Memory[this.SP]
	case 0x1b: // SP
		return &this.SP
	case 0x1c: // PC
		return &this.PC
	case 0x1d: // O
		return &this.O
	case 0x1e: // [address]
		this.cycles++
		value := &this.Memory[this.Memory[this.PC]]
		this.PC++
		return value
	case 0x1f: // Literal (immutable) number from memory
		// Can't assign to this, so take a pointer to a useless variable instead of Memory.
		this.cycles++
		value := this.Memory[this.PC]
		this.PC++
		return &value
	}
	panic("Invalid value passed to resolve")
}

// Increments the PC if a value would expect us to read a word from memory.
func (this *DCPU16) skipValue(what word) {
	if (what >= 0x0f && what < 0x18) || what == 0x1e || what == 0x1f {
		this.PC++
	}
}
