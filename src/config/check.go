package config

import (
	"bufio"
	"guard/src/proc"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func linex(file, line string) {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return
	}

	if line[0] == '#' {
		return
	}

	if strings.HasPrefix(line, "Logsize=") {
		kv := strings.Split(line, "=")
		val, _ := strconv.Atoi(kv[1])
		proc.S.ListLock.Lock()
		proc.S.List[file].Logsize = val
		proc.S.ListLock.Unlock()
		//fmt.Println(line, kv)
	} else if strings.HasPrefix(line, "Logfiles=") {
		kv := strings.Split(line, "=")
		val, _ := strconv.Atoi(kv[1])
		proc.S.ListLock.Lock()
		proc.S.List[file].Logfiles = val
		proc.S.ListLock.Unlock()
	} else if strings.HasPrefix(line, "Alart=") {
		kv := strings.Split(line, "=")
		val := kv[1]
		proc.S.ListLock.Lock()
		proc.S.List[file].Alart = val
		proc.S.ListLock.Unlock()
	} else if strings.HasPrefix(line, "Logapi=") {
		kv := strings.Split(line, "=")
		val := kv[1]
		proc.S.ListLock.Lock()
		proc.S.List[file].Logapi = val
		proc.S.ListLock.Unlock()
	} else if strings.HasPrefix(line, "Logserver=") {
		kv := strings.Split(line, "=")
		val := kv[1]
		proc.S.ListLock.Lock()
		proc.S.List[file].Logserver = val
		proc.S.ListLock.Unlock()
	} else if strings.HasPrefix(line, "Topic=") {
		kv := strings.Split(line, "=")
		val := kv[1]
		proc.S.ListLock.Lock()
		proc.S.List[file].Topic = val
		proc.S.ListLock.Unlock()
	} else if strings.HasPrefix(line, "WashMode=") {
		kv := strings.Split(line, "=")
		val := kv[1]
		proc.S.ListLock.Lock()
		proc.S.List[file].WashMode = val
		proc.S.ListLock.Unlock()
	} else if strings.HasPrefix(line, "Version=") {
		kv := strings.Split(line, "=")
		val := kv[1]
		if strings.Contains(val, "$") {
			proc.S.ListLock.Lock()
			cmd := exec.Command("/bin/bash", "-c", "echo "+val)
			bytes, _ := cmd.Output()
			proc.S.List[file].Version = strings.Replace(string(bytes), "\n", "", -1)
			proc.S.ListLock.Unlock()
		} else {
			proc.S.ListLock.Lock()
			proc.S.List[file].Version = val
			proc.S.ListLock.Unlock()
		}
	} else if strings.HasPrefix(line, "Dure=") {
		kv := strings.Split(line, "=")
		val, _ := strconv.Atoi(kv[1])
		proc.S.ListLock.Lock()
		proc.S.List[file].Dure = val
		proc.S.ListLock.Unlock()
	} else if strings.HasPrefix(line, "Retry=") {
		kv := strings.Split(line, "=")
		val, _ := strconv.Atoi(kv[1])
		proc.S.ListLock.Lock()
		proc.S.List[file].Retry = val
		proc.S.ListLock.Unlock()
	} /*else {
		fmt.Println(line)
	}*/
}

//Rconfig 读取守护脚本的配置参数，并写入map
func Rconfig(file, path string) {
	f, err := os.Open(path + "/bin/daemon/" + file)
	defer f.Close()
	if nil == err {
		buff := bufio.NewReader(f)
		for {
			line, err := buff.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			}
			linex(file, line)
		}
	}
}
