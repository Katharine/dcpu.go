package dcpu

var basicOpcodes = [0x10]func(cpu *DCPU16, a, b *word){
	nil,

	// SET
	func(cpu *DCPU16, a *word, b *word) {
        cpu.cycles++
		*a = *b
	},

	// ADD
	func(cpu *DCPU16, a *word, b *word) {
        cpu.cycles += 2
		total := int32(*a + *b)
		if total > 0xFFFF {
			cpu.O = 0x0001
		}
		*a = word(total & 0xFFFF)
	},

	// SUB
	func(cpu *DCPU16, a, b *word) {
        cpu.cycles += 2
		total := int32(*a - *b)
		if total < 0 {
			cpu.O = 0xFFFF
		}
		*a = word(total & 0xFFFF)
	},

	// MUL
	func(cpu *DCPU16, a, b *word) {
        cpu.cycles += 2
		total := uint32(*a * *b)
		cpu.O = word(total >> 16)
		*a = word(total & 0xFFFF)
	},

	// DIV
	func(cpu *DCPU16, a, b *word) {
        cpu.cycles += 3
		if *b == 0 {
			*a = 0
			cpu.O = 0
		} else {
			cpu.O = word((uint32(*a) << 16) / uint32(*b))
			*a = *a / *b
		}
	},

	// MOD
	func(cpu *DCPU16, a, b *word) {
        cpu.cycles += 3
		if *b == 0 {
			*a = 0
			cpu.O = 0
		} else {
			*a = *a % *b
		}
	},

	// SHL
	func(cpu *DCPU16, a, b *word) {
        cpu.cycles += 2
		cpu.O = word(((uint32(*a) << *b) >> 16) & 0xFFFF)
		*a = *a << *b
	},

	// SHR
	func(cpu *DCPU16, a, b *word) {
        cpu.cycles += 2
		cpu.O = word(((uint32(*a) << 16) >> *b) & 0xFFFF)
		*a = *a >> *b
	},

	// AND
	func(cpu *DCPU16, a, b *word) {
        cpu.cycles++
		*a = *a & *b
	},

	// BOR
	func(cpu *DCPU16, a, b *word) {
        cpu.cycles++
		*a = *a | *b
	},

	// XOR
	func(cpu *DCPU16, a, b *word) {
        cpu.cycles++
		*a = *a ^ *b
	},

	// IFE
	func(cpu *DCPU16, a, b *word) {
        cpu.cycles += 2
		cpu.skipping = !(*a == *b)
        if cpu.skipping {
            cpu.cycles++
        }
	},

	// IFN
	func(cpu *DCPU16, a, b *word) {
        cpu.cycles += 2
		cpu.skipping = !(*a != *b)
        if cpu.skipping {
            cpu.cycles++
        }
	},

	// IFG
	func(cpu *DCPU16, a, b *word) {
        cpu.cycles += 2
		cpu.skipping = !(*a > *b)
        if cpu.skipping {
            cpu.cycles++
        }
	},

	// IFB
	func(cpu *DCPU16, a, b *word) {
        cpu.cycles += 2
		cpu.skipping = !(*a&*b != 0)
        if cpu.skipping {
            cpu.cycles++
        }
	},
}

var extendedOpcodes = map[word]func(*DCPU16, *word) {
    0x01: func(cpu *DCPU16, a *word) {
        cpu.cycles += 2
        cpu.SP--
        cpu.Memory[cpu.SP] = cpu.PC
        cpu.PC = *a
    },
}
