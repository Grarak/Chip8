package main

import (
	"fmt"
	"github.com/Grarak/Chip8/chip8"
	"os"
	"runtime"
)

func printUsage() {
	fmt.Println("usage:", os.Args[0], "<path to game file>")
}

func main() {
	runtime.LockOSThread()

	if len(os.Args) != 2 {
		printUsage()
		os.Exit(1)
	}

	game, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0600)
	if err != nil {
		println("Can't open", os.Args[1])
		os.Exit(1)
	}

	emulator := chip8.New()
	err = emulator.Load(game)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		if !emulator.CpuCycle() || !emulator.PollEvents() {
			break
		}
	}

	emulator.Destroy()
}
