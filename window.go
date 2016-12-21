package trainerKit

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

type Window struct {
	WindowName string
	ptrProc    uintptr
	ptrWindow  uintptr
	ptrThread  uintptr
	ExeName    string
}

type PointerModel struct {
	Offsets      []uint32
	BaseAddress  uint32
	value        []byte
	pointerLevel int
	dmaAddress   uintptr
	DllName      string
}

func (win *Window) Open() (err error) {
	var mainptr uintptr = 0
	if win.WindowName != "" {
		ptrWindow := FindWindow("", win.WindowName)
		if ptrWindow == 0 {
			err = errors.New("FindWindow failed")
			return
		}
		win.ptrWindow = ptrWindow

		res, ptrThread := GetWindowThreadProcessId(ptrWindow)
		if res == 0 {
			err = errors.New("GetWindowThreadProcessId failed")
			return
		}
		win.ptrThread = ptrThread
		mainptr = ptrThread
	} else {
		buff, _ := processes()
		for _, element := range buff {
			if element != nil {
				if strings.ToLower(win.ExeName) == strings.ToLower(element.Executable()) {
					mainptr = uintptr(element.Pid())
					break
				}
			}
		}

		if mainptr == 0 {
			err = errors.New(fmt.Sprintf("FindProcPtr failed %s", win.ExeName))
			return
		}
	}

	ptrProc := OpenProcess(All, 0, mainptr)
	if ptrProc == 0 {
		err = errors.New("OpenProcess failed")
		return
	}
	win.ptrProc = ptrProc

	err = nil
	return
}

func (win *Window) Write(pointer PointerModel) (err error) {
	err = nil
	return
}

func (pointerM *PointerModel) FindDmaAddress(win Window) {
	pointer := pointerM.BaseAddress
	if pointerM.DllName != "" {
		pointer += uint32(FindModule(win.ptrProc, pointerM.DllName))
		println(pointer)
	}
	pointerLevel := len(pointerM.Offsets)
	hProcHandle := win.ptrProc
	var pTemp uint32 = 0

	var buffer []byte
	var pointerAddr uintptr = 0
	for i := 0; i < pointerLevel; i++ {
		if i == 0 {
			_, buffer, _ = ReadProcessMemory(hProcHandle, uintptr(pointer), 4)
		}
		pTemp = binary.LittleEndian.Uint32(buffer[:])
		pointerAddr = uintptr(pTemp + pointerM.Offsets[i])
		_, buffer, _ = ReadProcessMemory(hProcHandle, pointerAddr, 4)
	}
	pointerM.dmaAddress = pointerAddr
	return
}

func (pointerM *PointerModel) Read(win Window) uint32 {
	_, buffer, _ := ReadProcessMemory(win.ptrProc, uintptr(pointerM.dmaAddress), 4)
	return binary.LittleEndian.Uint32(buffer[:])
}

func (pointerM *PointerModel) Write(win Window, value []byte) {
	WriteProcessMemory(win.ptrProc, uintptr(pointerM.dmaAddress), value, 4)
	return
}
