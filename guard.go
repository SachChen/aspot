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
    "io/ioutil"
    //"log"
)



type Info struct {

    Status     bool
    Pid        int
    Logsize    int
    Logfiles   int
    Cron       time
    alart      bool

}




type Status struct {
	List     map[string]bool
	ListLock sync.RWMutex
}

var S = &Status{
	List: make(map[string]bool),
}


type PidInfo struct {
       List      map[string]int
       ListLock  sync.RWMutex
}

var P = &PidInfo{
        List: make(map[string]int),
}


func ls(path string) ([]byte,error) {
	return exec.Command("ls",path).Output()
}


func guard() ([]string) {
    out,err := ls("autolaunch")
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


func PathExists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
    return false, err
}


func startup(file string) {

      S.ListLock.Lock()
      P.ListLock.Lock()
      filepatsh := ("autolaunch/"+file)
      cmd := exec.Command(filepatsh)
      cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
      cmd.Start()
      //pidint := cmd.Process.Pid
      //pid := strconv.Itoa(pidint)
      fmt.Printf("Process:%-15s    Pid:%-5d\n",file,cmd.Process.Pid)
      //S.List[file] = map[int]bool{cmd.Process.Pid: true}
      S.List[file] = true
      P.List[file] = cmd.Process.Pid
      S.ListLock.Unlock()
      P.ListLock.Unlock()
      cmd.Wait()
}


func addpro(file string)  {

      //添加重试次数以及失败时间,达到重启次数上限后放弃重启操作,并执行相应动作
      i := 0
      for {
          i++
          if i < 3 {
              filepatsh := ("scripts/"+file)
              cmd := exec.Command(filepatsh)
              cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
              cmd.Start()
              fmt.Printf("Process:%-15s    Pid:%-5d\n",file,cmd.Process.Pid)
              pid := cmd.Process.Pid
              time.Sleep(time.Second * 5)
              pidstr := strconv.Itoa(pid)
              S.ListLock.Lock()
              P.ListLock.Lock()
              //m := syscall.Kill(pid,0)
              //fmt.Println(m)
              f,_ := os.Open("/proc/"+pidstr+"/cmdline")       //通过查看proc的cmdline判断是否产生僵尸进程，如果cmdline为空则是僵尸进，否则不是。
              i,_ := ioutil.ReadAll(f)
              m := string(i)
              
              if m != "" {
              // if m == nil
                  fmt.Println(m)
                  fmt.Println("launch process of "+file+" successfully !")
                  S.List[file] = true
                  P.List[file] = cmd.Process.Pid
                  S.ListLock.Unlock()
                  P.ListLock.Unlock()
                  //break
                  cmd.Wait()
                  break
              }
              if  m == "" {
                  fmt.Println("launch process of "+file+" filed , retrying . . .")
                  cmd.Wait()
                  //continue
                  S.List[file] = false
                  P.List[file] = cmd.Process.Pid
                  S.ListLock.Unlock()
                  P.ListLock.Unlock()
                  continue
              }
          } else {
              S.ListLock.Lock()
              P.ListLock.Lock()
              fmt.Println("launch process of "+file+" faild "+strconv.Itoa(i)+" times, droped !")
              S.List[file] = false
              S.ListLock.Unlock()
              P.ListLock.Unlock()
              break
          }
          
      }      

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
    P.ListLock.Lock()
    defer S.ListLock.Unlock()
    defer P.ListLock.Unlock()
    pid := P.List[file]
    err := syscall.Kill(-pid, syscall.SIGKILL)
    if err != nil {
        fmt.Println("kill process of "+file+" is failed, err:", err)
    } 
    for {
        time.Sleep(time.Second * 1)
        pidstr := strconv.Itoa(pid)
        _, err := os.Stat("/proc/"+pidstr)
        if err == nil {
            continue
        }
        if os.IsNotExist(err) {
            fmt.Println("kill process of "+file+" successfully !")
            break
        } 

    }
    S.List[file] = false

}

func Check(){

	for {
		time.Sleep(time.Second * 1)
		S.ListLock.Lock()
                P.ListLock.Lock()
		for file := range S.List {
		    pid := P.List[file]
			if S.List[file] == true {
                            pidstr := strconv.Itoa(pid)
			    _, err := os.Stat("/proc/"+pidstr)
			    if err == nil {
				continue
			    }
			    if os.IsNotExist(err) {
				fmt.Println(file + " is down ! restarting...")
				go addpro(file)
				fmt.Println(file + " is recover successfully!")
			    }
			}
				continue
	        }
		S.ListLock.Unlock()
                P.ListLock.Unlock()
	}
}


func downprocess(w http.ResponseWriter, r *http.Request) {
    //r.ParseForm()
    file := r.FormValue("file")
    killpid(file)
    fmt.Fprintln(w, file + " has been shutdown !")
}

func launchprocess(w http.ResponseWriter, r *http.Request) {
    file := r.FormValue("file")
    go addpro(file)
    fmt.Fprintln(w, file + " has been launched !")
    

}

func getstatus(w http.ResponseWriter, r *http.Request){
    file := r.FormValue("file")
    fmt.Fprintln(w ,S.List[file])
    
}


func getallstatus(w http.ResponseWriter, r *http.Request){
    fmt.Fprintln(w ,S.List)

}



func allshutdown(w http.ResponseWriter, r *http.Request){
    for file := range S.List{
        killpid(file)
        fmt.Fprintln(w, file + " has been shutdown !")

    }
}


func main() {
    process()
    time.Sleep(time.Second * 1)
    go Check()
    http.HandleFunc("/down", downprocess)
    http.HandleFunc("/alldown", allshutdown)
    http.HandleFunc("/status", getstatus)
    http.HandleFunc("/allstatus", getallstatus)
    http.HandleFunc("/launch", launchprocess)
    err := http.ListenAndServe("0.0.0.0:8010",nil)
    fmt.Println(err)

}
