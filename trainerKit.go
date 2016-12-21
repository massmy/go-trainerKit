package trainerKit

import (
	"fmt"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

//     internal class WinApiHandler
//     {
//         [DllImport("user32.dll", CharSet = CharSet.Auto, SetLastError = true)]
//         public static extern IntPtr FindWindow(string lpClassName, string lpWindowName);

//         [DllImport("user32.dll")]
//         public static extern uint GetWindowThreadProcessId(IntPtr hWnd, IntPtr ProcessId);

//         [DllImport("user32.dll", SetLastError = true)]
//         public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out uint lpdwProcessId);

//         [DllImport("kernel32.dll")]
//         public static extern IntPtr OpenProcess(int dwDesiredAccess, bool bInheritHandle, int dwProcessId);

//         [DllImport("kernel32.dll", SetLastError = true)]
//         public static extern bool ReadProcessMemory(
//             IntPtr hProcess,
//             IntPtr lpBaseAddress,
//             [Out] byte[] lpBuffer,
//             int dwSize,
//             out IntPtr lpNumberOfBytesRead
//         );

//         [DllImport("kernel32.dll", SetLastError = true)]
//         public static extern bool WriteProcessMemory(IntPtr hProcess, IntPtr lpBaseAddress, byte[] lpBuffer, uint nSize, out UIntPtr lpNumberOfBytesWritten);
//     }

//     [Flags]
//     public enum ProcessAccessFlags : uint
//     {
//         All = 0x001F0FFF,
//         Terminate = 0x00000001,
//         CreateThread = 0x00000002,
//         VMOperation = 0x00000008,
//         VMRead = 0x00000010,
//         VMWrite = 0x00000020,
//         DupHandle = 0x00000040,
//         SetInformation = 0x00000200,
//         QueryInformation = 0x00000400,
//         Synchronize = 0x00100000
// }

var (
	kernel32            = syscall.MustLoadDLL("kernel32.dll")
	openProcess         = kernel32.MustFindProc("OpenProcess")
	readProcessMemory   = kernel32.MustFindProc("ReadProcessMemory")
	writeProcessMemory  = kernel32.MustFindProc("WriteProcessMemory")
	psapi               = syscall.MustLoadDLL("Psapi.dll")
	enumProcessModules  = psapi.MustFindProc("EnumProcessModulesEx")
	getModuleFileNameEx = psapi.MustFindProc("GetModuleFileNameExW")
	getModuleBaseName   = psapi.MustFindProc("GetModuleBaseNameW")
	enumProcesses       = psapi.MustFindProc("EnumProcesses")

	// getModuleHandleEx   = kernel32.MustFindProc("GetModuleHandleExW")

	user32                   = syscall.MustLoadDLL("user32.dll")
	findWindow               = user32.MustFindProc("FindWindowW")
	getWindowThreadProcessId = user32.MustFindProc("GetWindowThreadProcessId")
	// messageBox, _ = syscall.GetProcAddress(user32, "MessageBoxW")
)

const (
	All              = 0x001F0FFF
	Terminate        = 0x00000001
	CreateThread     = 0x00000002
	VMOperation      = 0x00000008
	VMRead           = 0x00000010
	VMWrite          = 0x00000020
	DupHandle        = 0x00000040
	SetInformation   = 0x00000200
	QueryInformation = 0x00000400
	Synchronize      = 0x00100000
	LIST_MODULES_ALL = 0x03
)

func abort(funcname string, err error) {
	fmt.Sprintf("%s failed: %v", funcname, err)
	// panic(fmt.Sprintf("%s failed: %v", funcname, err))
}
func strToPtr(str string) (res uintptr) {
	if str == "" {
		res = uintptr(0)
	} else {
		res = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(str)))
	}
	return
}

// func freeLib() {
// 	defer syscall.FreeLibrary(kernel32)
// 	defer syscall.FreeLibrary(user32)
// }

func OpenProcess(dwDesiredAccess int, bInheritHandle int, dwProcessId uintptr) (result uintptr) {
	//     uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
	// uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(caption))),
	// var nargs uintptr = 3
	ret, _, callErr := openProcess.Call(
		uintptr(dwDesiredAccess),
		uintptr(bInheritHandle),
		dwProcessId)
	if callErr != nil {
		abort("Call OpenProcess", callErr)
	}
	result = uintptr(ret)
	return
}

func FindWindow(lpClassName string, lpWindowName string) (result uintptr) {
	ret, _, callErr := findWindow.Call(
		strToPtr(lpClassName),
		strToPtr(lpWindowName))
	if callErr != nil {
		abort("Call FindWindow", callErr)
	}
	result = uintptr(ret)
	return
}

func GetWindowThreadProcessId(window uintptr) (result uintptr, handle uintptr) {
	ret, _, callErr := getWindowThreadProcessId.Call(
		window,
		uintptr(unsafe.Pointer(&handle)))
	if callErr != nil {
		abort("Call FindWindow", callErr)
	}
	result = uintptr(ret)
	return
}

func WriteProcessMemory(hProcess uintptr, lpBaseAddress uintptr, lpBuffer []byte, nSize int) (result uintptr, lpNumberOfBytesWritten uintptr) {
	ret, _, callErr := writeProcessMemory.Call(
		hProcess,
		lpBaseAddress,
		uintptr(unsafe.Pointer(&lpBuffer[0])),
		uintptr(nSize),
		uintptr(unsafe.Pointer(&lpNumberOfBytesWritten)))

	if callErr != nil {
		abort("Call WriteProcessMemory", callErr)
	}
	result = uintptr(ret)
	return
}

func ReadProcessMemory(hProcess uintptr, lpBaseAddress uintptr, nSize int) (result uintptr, lpBuffer []byte, lpNumberOfBytesWritten uintptr) {
	lpBuffer = make([]byte, nSize)
	ret, _, callErr := readProcessMemory.Call(
		hProcess,
		lpBaseAddress,
		uintptr(unsafe.Pointer(&lpBuffer[0])),
		uintptr(nSize),
		uintptr(unsafe.Pointer(&lpNumberOfBytesWritten)))

	// fmt.Println(lpBuffer)
	if callErr != nil {
		abort("Call ReadProcessMemory", callErr)
	}
	result = uintptr(ret)
	return
}

func EnumProcessModules(handle uintptr) {
	fmt.Println(handle)

	modules := make([]uintptr, 2049)
	var needed int
	enumProcessModules.Call(
		handle,
		uintptr(unsafe.Pointer(&modules[0])),
		uintptr(2048),
		uintptr(unsafe.Pointer(&needed)),
		uintptr(LIST_MODULES_ALL),
	)
	fmt.Println(needed)
	for i := 0; i < needed; i++ {
		if modules[i] != 0 {
			fmt.Print(modules[i])
			_, name := GetModuleBaseName(handle, modules[i])
			fmt.Println(name)
		}
	}
}

func FindModule(handle uintptr, moduleName string) (retHandle uintptr) {
	fmt.Println(handle)

	modules := make([]uintptr, 2049)
	var needed int
	enumProcessModules.Call(
		handle,
		uintptr(unsafe.Pointer(&modules[0])),
		uintptr(2048),
		uintptr(unsafe.Pointer(&needed)),
		uintptr(LIST_MODULES_ALL),
	)
	retHandle = 0
	for i := 0; i < needed; i++ {
		if modules[i] != 0 {
			// fmt.Print(modules[i])
			_, name := GetModuleBaseName(handle, modules[i])
			if strings.Contains(strings.ToLower(name), strings.ToLower(moduleName)) {
				retHandle = modules[i]
				return
			}
		}
	}
	return
}

func UintptrToString(v uintptr) string {
	if v == 0 {
		return ""
	}

	return syscall.UTF16ToString((*[1 << 29]uint16)(unsafe.Pointer(v))[0:])
}

func GetModuleFileNameEx(handle uintptr, handleMod uintptr) (result uintptr, res string) {
	buff := make([]uint16, syscall.MAX_PATH)
	var resptr uintptr = uintptr(unsafe.Pointer(&buff[0]))
	ret, _, callErr := getModuleFileNameEx.Call(
		handle,
		handleMod,
		resptr,
		uintptr(syscall.MAX_PATH))

	if callErr != nil || uint32(ret) == 0 {
		abort("Call GetModuleFileNameEx", callErr)
	}
	result = uintptr(ret)

	res = string(utf16.Decode(buff[0:uint32(ret)]))

	return
}

func GetModuleBaseName(handle uintptr, handleMod uintptr) (result uintptr, res string) {
	buff := make([]uint16, syscall.MAX_PATH)
	var resptr uintptr = uintptr(unsafe.Pointer(&buff[0]))
	ret, _, callErr := getModuleBaseName.Call(
		handle,
		handleMod,
		resptr,
		uintptr(syscall.MAX_PATH))

	if callErr != nil {
		abort("Call GetModuleFileNameEx", callErr)
	}

	result = uintptr(ret)
	res = string(utf16.Decode(buff[0:uint32(ret)]))

	return
}

func EnumProcesses() (result uintptr, res []uint16) {
	buff := make([]uint16, 1024)
	var resptr uintptr = uintptr(unsafe.Pointer(&buff[0]))
	var needed int

	ret, _, callErr := enumProcesses.Call(
		resptr,
		unsafe.Sizeof(buff),
		uintptr(unsafe.Pointer(&needed)))

	if callErr != nil {
		abort("Call EnumProcesses", callErr)
	}
	print(needed)
	result = uintptr(ret)
	res = buff
	return
}

func GetName(handle uintptr) (name string) {
	ptr := OpenProcess(QueryInformation|
		VMRead,
		0, handle)
	EnumProcessModules(ptr)
	name = ""
	return
}
