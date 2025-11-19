package model

import "time"

type Code struct {
	Language   string
	SourceCode string
	Stdin 	   string
}

type Container struct {
	Image  string
	ID     string
	Status string
}

type Result struct {
	Status       string        
	Error        string        
	Success      string        
	TimeTaken    time.Duration 
	MemoryTaken  int          
}

type CodeContainer interface {
	Execute(code *Code) (*Result, error)
}
