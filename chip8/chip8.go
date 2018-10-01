package chip8

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"math"
	"math/rand"
	"os"
	"time"
)

type Chip8 struct {
	window     *sdl.Window
	renderer   *sdl.Renderer
	keyPressed map[rune]bool

	display    [][]bool
	memory     []uint8
	opcode     uint16
	pc         uint16
	registers  []uint8
	mar        uint16
	stack      []uint16
	sp         uint8
	delayTimer uint8
	soundTimer uint8

	draw bool
}

func New() *Chip8 {
	memory := make([]uint8, MemorySize)
	for font := 0; font < len(Fonts); font++ {
		for fontPixel := 0; fontPixel < len(Fonts[font]); fontPixel++ {
			memory[font*len(Fonts[font])+fontPixel] = Fonts[font][fontPixel]
		}
	}

	display := make([][]bool, DisplayWidth)
	for i := range display {
		display[i] = make([]bool, DisplayHeight)
	}

	return &Chip8{
		keyPressed: make(map[rune]bool),
		display:    display,
		memory:     memory,
		pc:         RomOffset,
		registers:  make([]uint8, RegistersSize),
		stack:      make([]uint16, StackSize),
	}
}

func (chip8 *Chip8) Load(game *os.File) error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	initialWindowWidth, initialWindowHeight := int32(800), int32(600)

	window, err := sdl.CreateWindow("Chip 8",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		initialWindowWidth, initialWindowHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}

	chip8.window = window
	chip8.renderer = renderer

	defer game.Close()
	buf := make([]byte, 4096)

	_, err = game.Read(buf)
	if err != nil {
		return err
	}

	for i := RomOffset; i < uint16(len(chip8.memory)); i++ {
		chip8.memory[i] = uint8(buf[i-RomOffset])
	}

	return nil
}

func (chip8 *Chip8) CpuCycle() bool {
	chip8.opcode = uint16(chip8.memory[chip8.pc])<<8 | uint16(chip8.memory[chip8.pc+1])

	ret := chip8.executeOpCode()
	time.Sleep(time.Microsecond * 1200)
	return ret
}

func (chip8 *Chip8) executeOpCode() bool {
	instruction := chip8.opcode & 0xF000

	switch instruction {
	case 0x0000:
		switch chip8.opcode & 0x00FF {
		case 0x00E0:
			for x := 0; x < len(chip8.display); x++ {
				for y := 0; y < len(chip8.display[x]); y++ {
					chip8.display[x][y] = false
				}
			}

			chip8.draw = true
			chip8.pc += 2
			break

		case 0x00EE:
			chip8.sp--
			chip8.pc = chip8.stack[chip8.sp] + 2
			break

		default:
			fmt.Printf("unknown instruction: %016X\n", chip8.opcode)
			return false
		}
		break

	case 0x1000:
		chip8.pc = chip8.opcode & 0x0FFF
		break

	case 0x2000:
		chip8.stack[chip8.sp] = chip8.pc
		chip8.sp++

		chip8.pc = chip8.opcode & 0x0FFF
		break

	case 0x3000:
		register := chip8.opcode & 0x0F00 >> 8
		value := chip8.opcode & 0x00FF
		if chip8.registers[register] == uint8(value) {
			chip8.pc += 4
		} else {
			chip8.pc += 2
		}
		break

	case 0x4000:
		register := chip8.opcode & 0x0F00 >> 8
		value := chip8.opcode & 0x00FF
		if chip8.registers[register] != uint8(value) {
			chip8.pc += 4
		} else {
			chip8.pc += 2
		}
		break

	case 0x5000:
		registerA := chip8.opcode & 0x0F00 >> 8
		registerB := chip8.opcode & 0x00F0 >> 4
		if chip8.registers[registerA] == chip8.registers[registerB] {
			chip8.pc += 4
		} else {
			chip8.pc += 2
		}
		break

	case 0x6000:
		register := chip8.opcode & 0x0F00 >> 8
		value := chip8.opcode & 0x00FF
		chip8.registers[register] = uint8(value)

		chip8.pc += 2
		break

	case 0x7000:
		register := chip8.opcode & 0x0F00 >> 8
		add := chip8.opcode & 0x00FF
		chip8.registers[register] += uint8(add)

		chip8.pc += 2
		break

	case 0x8000:
		switch chip8.opcode & 0x000F {
		case 0x0000:
			registerA := chip8.opcode & 0x0F00 >> 8
			registerB := chip8.opcode & 0x00F0 >> 4
			chip8.registers[registerA] = chip8.registers[registerB]

			chip8.pc += 2
			break

		case 0x0001:
			registerA := chip8.opcode & 0x0F00 >> 8
			registerB := chip8.opcode & 0x00F0 >> 4
			chip8.registers[registerA] |= chip8.registers[registerB]

			chip8.pc += 2
			break

		case 0x0002:
			registerA := chip8.opcode & 0x0F00 >> 8
			registerB := chip8.opcode & 0x00F0 >> 4
			chip8.registers[registerA] &= chip8.registers[registerB]

			chip8.pc += 2
			break

		case 0x0003:
			registerA := chip8.opcode & 0x0F00 >> 8
			registerB := chip8.opcode & 0x00F0 >> 4
			chip8.registers[registerA] ^= chip8.registers[registerB]

			chip8.pc += 2
			break

		case 0x0004:
			registerA := chip8.opcode & 0x0F00 >> 8
			registerB := chip8.opcode & 0x00F0 >> 4

			if chip8.registers[registerB] > 0xFF-chip8.registers[registerA] {
				chip8.registers[0xF] = 1
			} else {
				chip8.registers[0xF] = 0
			}
			chip8.registers[registerA] += chip8.registers[registerB]

			chip8.pc += 2
			break

		case 0x0005:
			registerA := chip8.opcode & 0x0F00 >> 8
			registerB := chip8.opcode & 0x00F0 >> 4

			if chip8.registers[registerB] > chip8.registers[registerA] {
				chip8.registers[0xF] = 0
			} else {
				chip8.registers[0xF] = 1
			}
			chip8.registers[registerA] -= chip8.registers[registerB]

			chip8.pc += 2
			break

		case 0x0006:
			register := chip8.opcode & 0x0F00 >> 8
			chip8.registers[0xF] = chip8.registers[register] & 1
			chip8.registers[register] >>= 1

			chip8.pc += 2
			break

		case 0x0007:
			registerA := chip8.opcode & 0x0F00 >> 8
			registerB := chip8.opcode & 0x00F0 >> 4

			if chip8.registers[registerA] > chip8.registers[registerB] {
				chip8.registers[0xF] = 0
			} else {
				chip8.registers[0xF] = 1
			}
			chip8.registers[registerA] = chip8.registers[registerB] - chip8.registers[registerA]

			chip8.pc += 2
			break

		case 0x000E:
			register := chip8.opcode & 0x0F00 >> 8
			chip8.registers[0xF] = chip8.registers[register] >> 7
			chip8.registers[register] <<= 1

			chip8.pc += 2
			break

		default:
			fmt.Printf("unknown instruction: %016X\n", chip8.opcode)
			return false
		}
		break

	case 0x9000:
		registerA := chip8.opcode & 0x0F00 >> 8
		registerB := chip8.opcode & 0x00F0 >> 4
		if chip8.registers[registerA] != chip8.registers[registerB] {
			chip8.pc += 4
		} else {
			chip8.pc += 2
		}
		break

	case 0xA000:
		chip8.mar = chip8.opcode & 0x0FFF

		chip8.pc += 2
		break

	case 0xC000:
		register := chip8.opcode & 0x0F00 >> 8
		chip8.registers[register] = uint8(uint16(rand.Intn(256)) & (chip8.opcode & 0x00FF))

		chip8.pc += 2
		break

	case 0xD000:
		x := chip8.registers[chip8.opcode&0x0F00>>8]
		y := chip8.registers[chip8.opcode&0x00F0>>4]
		height := uint8(chip8.opcode & 0x000F)

		chip8.registers[0xF] = 0
		for yHeight := uint8(0); yHeight < height; yHeight++ {
			pixel := chip8.memory[chip8.mar+uint16(yHeight)]
			for xWidth := uint8(0); xWidth < 8; xWidth++ {
				if (pixel>>(7-xWidth))&1 == 1 {
					actualWidth := int(x + xWidth)
					if actualWidth >= len(chip8.display) {
						actualWidth = int(x+xWidth) % len(chip8.display)
					}

					actualHeight := int(y + yHeight)
					if actualHeight >= len(chip8.display[actualWidth]) {
						actualHeight = int(y+yHeight) % len(chip8.display[actualWidth])
					}

					if chip8.display[actualWidth][actualHeight] {
						chip8.registers[0xF] = 1
					}
					chip8.display[actualWidth][actualHeight] = !chip8.display[actualWidth][actualHeight]
				}
			}
		}

		chip8.draw = true
		chip8.pc += 2
		break

	case 0xE000:
		switch chip8.opcode & 0x00FF {
		case 0x009E:
			register := chip8.opcode & 0x0F00 >> 8
			if chip8.isKeyPressed(KeyMapping[chip8.registers[register]]) {
				chip8.pc += 4
			} else {
				chip8.pc += 2
			}
			break

		case 0x00A1:
			register := chip8.opcode & 0x0F00 >> 8
			if !chip8.isKeyPressed(KeyMapping[chip8.registers[register]]) {
				chip8.pc += 4
			} else {
				chip8.pc += 2
			}
			break

		default:
			fmt.Printf("unknown instruction: %016X\n", chip8.opcode)
			return false
		}
		break

	case 0xF000:
		switch chip8.opcode & 0x00FF {
		case 0x007:
			register := chip8.opcode & 0x0F00 >> 8
			chip8.registers[register] = chip8.delayTimer

			chip8.pc += 2
			break

		case 0x00A:
			register := chip8.opcode & 0x0F00 >> 8
			pressed := false
			for index, key := range KeyMapping {
				if chip8.isKeyPressed(key) {
					chip8.registers[register] = uint8(index)
					chip8.pc += 2
					pressed = true
					break
				}
			}
			if !pressed {
				return true
			}
			break

		case 0x0015:
			register := chip8.opcode & 0x0F00 >> 8
			chip8.delayTimer = chip8.registers[register]

			chip8.pc += 2
			break

		case 0x0018:
			register := chip8.opcode & 0x0F00 >> 8
			chip8.soundTimer = chip8.registers[register]

			chip8.pc += 2
			break

		case 0x001E:
			register := chip8.opcode & 0x0F00 >> 8

			if chip8.mar+uint16(chip8.registers[register]) > 0xFFF {
				chip8.registers[0xF] = 1
			} else {
				chip8.registers[0xF] = 0
			}
			chip8.mar += uint16(chip8.registers[register])

			chip8.pc += 2
			break

		case 0x0029:
			register := chip8.opcode & 0x0F00 >> 8
			character := uint16(chip8.registers[register])
			chip8.mar = character * uint16(len(Fonts[character]))

			chip8.pc += 2
			break

		case 0x0033:
			register := chip8.opcode & 0x0F00 >> 8
			value := chip8.registers[register]
			chip8.memory[chip8.mar] = uint8(value / 100)
			chip8.memory[chip8.mar+1] = uint8(value / 10 % 10)
			chip8.memory[chip8.mar+2] = uint8(value % 10)

			chip8.pc += 2
			break

		case 0x0055:
			value := chip8.opcode & 0x0F00 >> 8
			for i := uint16(0); i <= value; i++ {
				chip8.memory[chip8.mar+i] = chip8.registers[i]
			}

			chip8.mar += value + 1
			chip8.pc += 2
			break

		case 0x0065:
			value := chip8.opcode & 0x0F00 >> 8
			for i := uint16(0); i <= value; i++ {
				chip8.registers[i] = chip8.memory[chip8.mar+i]
			}

			chip8.mar += value + 1
			chip8.pc += 2
			break

		default:
			fmt.Printf("unknown instruction: %016X\n", chip8.opcode)
			return false
		}
		break

	default:
		fmt.Printf("unknown instruction: %016X\n", chip8.opcode)
		return false
	}

	if chip8.delayTimer > 0 {
		chip8.delayTimer--
	}

	if chip8.soundTimer > 0 {
		chip8.soundTimer--
	}

	return true
}

func (chip8 *Chip8) PollEvents() bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.KeyboardEvent:
			keyEvent := event.(*sdl.KeyboardEvent)
			chip8.keyPressed[rune(keyEvent.Keysym.Sym)] = keyEvent.Type == sdl.KEYDOWN
			break
		case *sdl.QuitEvent:
			return false
		}
	}

	if chip8.draw {
		chip8.draw = false

		width, height := chip8.window.GetSize()

		xPixelSize, yPixelSize := float64(width)/DisplayWidth, float64(height)/DisplayHeight

		chip8.renderer.Clear()
		chip8.renderer.SetDrawColor(0, 0, 0, 255)
		chip8.renderer.FillRect(&sdl.Rect{W: width, H: height})

		chip8.renderer.SetDrawColor(255, 255, 255, 255)
		for x := 0; x < len(chip8.display); x++ {
			for y := 0; y < len(chip8.display[x]); y++ {
				if chip8.display[x][y] {
					chip8.renderer.FillRect(&sdl.Rect{
						X: int32(math.Round(float64(x) * xPixelSize)),
						Y: int32(math.Round(float64(y) * yPixelSize)),
						W: int32(math.Round(xPixelSize)),
						H: int32(math.Round(yPixelSize)),
					})
				}
			}
		}

		chip8.renderer.Present()
	}

	return true
}

func (chip8 *Chip8) isKeyPressed(key rune) bool {
	pressed, ok := chip8.keyPressed[key]
	return ok && pressed
}

func (chip8 *Chip8) Destroy() {
	chip8.renderer.Destroy()
	chip8.window.Destroy()
	sdl.Quit()
}
