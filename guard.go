package main

import (
    "time"
    "os/exec"
    "fmt"
    "strings"
    "strconv"
    "os"
    "sync"
)


type Status struct {
	List     map[string]string
	ListLock sync.RWMutex
}

var S = &Status{
	List: make(map[string]string),
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
      cmd.Start()
      pidint := cmd.Process.Pid
      pid := strconv.Itoa(pidint)
      fmt.Printf("process:%-15s    pid:%-5s\n",file,pid)
      S.List[file] = pid
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


func Check(){

    for {
            time.Sleep(time.Second * 1)

            for file := range S.List {
                S.ListLock.Lock()
                pid := S.List[file]
                _, err := os.Stat("/proc/"+pid)
                    if err == nil {
                        S.ListLock.Unlock()
                        continue
                    }
                    if os.IsNotExist(err) {
                        S.ListLock.Unlock()
                        fmt.Println(file + " is down ! restarting...")
                        go startup(file)
                        fmt.Println(file + " is recover successfully!")
                    }
         }
      }
}


func main() {
       
    process()
    Check()

}
