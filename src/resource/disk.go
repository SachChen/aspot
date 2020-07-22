package resource

import (
	"bufio"
	"io"
	"os"
	"strings"
	"syscall"
)

var A []string

type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

type Diskslice struct {
	Path    string
	Total   int64
	Used    float64
	UsePerc float64
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

// disk usage of path/disk
func DiskUsage(path string) (disk DiskStatus) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}

func Fstab() []string {
	var A []string
	f, err := os.Open("/proc/mounts")
	if nil == err {
		buff := bufio.NewReader(f)
		for {
			st, err := buff.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			}
			line := strings.Trim(st, "\n")
			arr := strings.Split(line, " ")
			check := strings.Contains(arr[0], "/dev/")
			if check {
				tmpcheck := strings.Contains(arr[0], "tmpfs")
				if tmpcheck {
					continue
				} else {
					etcheck := strings.Contains(arr[1], "/etc/")
					if etcheck {
						continue
					} else {
						shmheck := strings.Contains(arr[1], "/dev/shm")
						if shmheck {
							continue
						} else {
							A = append(A, arr[1])
						}
					}

				}
			} else {
				continue
			}
		}
		f.Close()
	}
	return A
}

func Dcount() []Diskslice {
	Paths := Fstab()
	//var  i  string
	var M []Diskslice
	for _, v := range Paths {
		//i = v
		//var M []Diskslice
		d := Diskslice{}
		//d := new(m)
		disk := DiskUsage(v)
		diskall := float64(disk.All) / float64(MB)
		//diskfree := float64(disk.Free)/float64(MB)
		diskused := float64(disk.Used) / float64(MB)
		dfpercent := float64(diskused / diskall)
		//fmt.Printf("%s %s %s %s %s %s %s %.2f%%\n","Path:",v," Volume:",diskall," Used:",diskused, " Avalible:",dfpercent*100)
		//fmt.Println("Path:",v," Total:",int64(diskall+0.5),"GB"," Used:",float64(int64((diskused+0.05)*10))/10,"GB", " Use%:",float64(int64((dfpercent*100+0.05)*10))/10)
		d.Path = v
		d.Total = int64(diskall + 0.5)
		d.Used = float64(int64((diskused+0.05)*10)) / 10
		d.UsePerc = float64(int64((dfpercent*100+0.05)*10)) / 10
		M = append(M, d)

	}
	return M
}
