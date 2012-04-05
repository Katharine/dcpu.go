package dcpu

import (
	"io"
	"os"
)

type word uint16

type register byte

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

type DCPU16 struct {
	Memory    [0x10000]word
	Registers  [8]word // Access them using the register constants above, e.g. Registers[C]
	PC        word
	SP        word
	O         word
	skipping  bool
	cycles    uint
}

func New() *DCPU16 {
	cpu := new(DCPU16)
	cpu.PC = 0
	cpu.SP = 0xFFFF
	return cpu
}

func (this *DCPU16) LoadImage(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return this.LoadStream(f)
}

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

func (this *DCPU16) Run() {
	for {
		this.ExecuteCycle()
	}
}

func (this *DCPU16) Cycles() uint {
	return this.cycles
}

func (this *DCPU16) ExecuteCycle() {
	operation := this.Memory[this.PC]
	this.PC++
	opcode := operation & 0xF
	v1 := operation >> 4 & 0x3F
	v2 := operation >> 10

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
	case 0x1b:
		return &this.SP
	case 0x1c:
		return &this.PC
	case 0x1d:
		return &this.O
	case 0x1e:
		this.cycles++
		value := &this.Memory[this.Memory[this.PC]]
		this.PC++
		return value
	case 0x1f:
		// Can't assign to this, so take a pointer to a useless variable instead of Memory.
		this.cycles++
		value := this.Memory[this.PC]
		this.PC++
		return &value
	}
	panic("Invalid value passed to resolve")
}

func (this *DCPU16) skipValue(what word) {
	if (what >= 0x0f && what < 0x18) || what == 0x1e || what == 0x1f {
		this.PC++
	}
}
