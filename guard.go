package main

import (
    "time"
    "os/exec"
    "fmt"
    "strings"
    "strconv"
    "os"
    "sync"
    "syscall"
    "net/http"
)


type Status struct {
	List     map[string]map[int]bool
	ListLock sync.RWMutex
}

var S = &Status{
	List: make(map[string]map[int]bool),
}


func ls(path string) ([]byte,error) {
	return exec.Command("ls",path).Output()
}


func guard() ([]string) {
    out,err := ls("scripts")
    if err != nil {
        panic(err)
}
    output := strings.Split(string(out), "\n")
    return output

}

func pwd() string{
    pwd,err := exec.Command("pwd").Output()
    if err != nil {
	panic(err)
    }
    pout := string(pwd)
    return  pout
}



func startup(file string) {

      S.ListLock.Lock()
      filepath := ("scripts/"+file)
      cmd := exec.Command(filepath)
      //cmd.Start()
      cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
      cmd.Start()
      //pidint := cmd.Process.Pid
      //pid := strconv.Itoa(pidint)
      fmt.Printf("Process:%-15s    Pid:%-5d\n",file,cmd.Process.Pid)
      S.List[file] = map[int]bool{cmd.Process.Pid: true}
      S.ListLock.Unlock()
      cmd.Wait()
}


func process() {

    files := guard()
        for i,_ := range files{
                if files[i] == "" {
                        continue
                }
                go startup(files[i])
        }
}


func killpid(file string) {
    S.ListLock.Lock()
    defer S.ListLock.Unlock()
    a := S.List[file]
    for pid,_ := range a{
    //pid,_ := strconv.Atoi(k)
    err := syscall.Kill(-pid, syscall.SIGKILL)
    if err != nil {
        fmt.Println("kill process team is failed, err:", err)
    }
    fmt.Println("kill process team is ok")
    S.List[file][pid] = false
    }

}

func Check(){

	for {
		time.Sleep(time.Second * 1)
		S.ListLock.Lock()
		for file := range S.List {
			for pid := range S.List[file] {
				if S.List[file][pid] == true {
                                        pidstr := strconv.Itoa(pid)
					_, err := os.Stat("/proc/"+pidstr)
						if err == nil {
							continue
						}
					if os.IsNotExist(err) {
						fmt.Println(file + " is down ! restarting...")
							go startup(file)
							fmt.Println(file + " is recover successfully!")
					}
				}
				continue

			}
		}
		S.ListLock.Unlock()
	}
}


func downprocess(w http.ResponseWriter, r *http.Request) {
    //r.ParseForm()
    file := r.FormValue("file")
    killpid(file)
    fmt.Fprintln(w, file + " has been shutdown !")
}


func Getinfo(w http.ResponseWriter, r *http.Request){ 
    fmt.Fprintln(w ,S.List)
    
}

func main() {
    process()
    time.Sleep(time.Second * 1)
    go Check()
    http.HandleFunc("/down", downprocess)
    http.HandleFunc("/getinfo", Getinfo)
    err := http.ListenAndServe("0.0.0.0:8010",nil)
    fmt.Println(err)

}
