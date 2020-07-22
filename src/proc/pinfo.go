package proc

import "sync"

type Pinfo struct {
	Pid       int
	CPid      int
	Status    bool
	Logsize   int
	Logfiles  int
	Alart     string
	Logapi    string
	Logserver string
	Topic     string
	WashMode  string
	Version   string
	Startup   string
	Dure      int
	Retry     int
}

type Status struct {
	List     map[string]*Pinfo
	ListLock sync.RWMutex
}

var S = &Status{
	List: make(map[string]*Pinfo),
}
