package chip8

import (
	"testing"
)

func assert(t *testing.T, a, b interface{}) {
	if a != b {
		t.Fatalf("%v != %v", a, b)
	}
}

func testCpuCycle(opcodes ...uint16) *Chip8 {
	emulator := New()
	for _, opcode := range opcodes {
		emulator.opcode = opcode
		emulator.executeOpCode()
	}
	return emulator
}

func TestChip8_CpuCycle(t *testing.T) {
	emulator := testCpuCycle(0x00E0)
	for x := 0; x < len(emulator.display); x++ {
		for y := 0; y < len(emulator.display[x]); y++ {
			if emulator.display[x][y] {
				t.Fatal("Display was not cleared")
			}
		}
	}
	assert(t, emulator.draw, true)

	emulator = testCpuCycle(0x1111)
	assert(t, emulator.pc, uint16(0x111))

	emulator = testCpuCycle(0x2122)
	assert(t, emulator.pc, uint16(0x122))
	assert(t, emulator.stack[emulator.sp-1], RomOffset)

	emulator = testCpuCycle(0x6011, 0x3011)
	assert(t, emulator.registers[0], uint8(0x11))
	assert(t, emulator.pc, RomOffset+6)

	emulator = testCpuCycle(0xA555, 0x747C, 0xF433)
	assert(t, emulator.mar, uint16(0x555))
	assert(t, emulator.registers[4], uint8(0x7C))
	assert(t, emulator.memory[0x555], uint8(1))
	assert(t, emulator.memory[0x555+1], uint8(2))
	assert(t, emulator.memory[0x555+2], uint8(4))
}
