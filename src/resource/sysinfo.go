package resource

import (
	"os"
	"runtime"
	"syscall"
	"time"
)

type Sys struct {
	Hostname string
	Os       string
	Release  string
	Arch     string
	Systime  string
	Zone     string
}

func int8ToStr(arr []int8) string {
	b := make([]byte, 0, len(arr))
	for _, v := range arr {
		if v == 0x00 {
			break
		}
		b = append(b, byte(v))
	}
	return string(b)
}

var info map[string]Sys

func Sysinfo() (info Sys) {
	name, err := os.Hostname()
	if err == nil {
		//fmt.Println(name)
		info.Hostname = name
	} else {
		info.Hostname = "localhost"
	}
	//fmt.Println(info.Hostname)
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err == nil {
		info.Os = runtime.GOOS
		info.Release = int8ToStr(uname.Release[:])
		info.Arch = runtime.GOARCH
	}

	info.Systime = time.Now().Format("2006-01-02 15:04:05")
	info.Zone, _ = time.Now().Zone()
	return
}
