package trainerKit

import (
	"encoding/binary"
	"testing"
)

func TestFindWindow(t *testing.T) {
	win := Window{
		ExeName: "Freelancer.exe",
		// windowName: "Calculator",
	}
	ptr := PointerModel{
		Offsets: []uint32{(0x54), (0x0), (0x4), (0x1C0), (0x31C)},
		DllName: "server.dll",
	}
	if ptr.dmaAddress == 0 {
		ptr.dmaAddress = 0
	}
	err := win.Open()
	if err != nil {
		panic(err)
	}
	print(win.ptrProc)

	ptr.FindDmaAddress(win)
	print("dmaAddress: ")
	println(ptr.dmaAddress)
	println(ptr.Read(win))
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, 20000)
	ptr.Write(win, bs)
}
