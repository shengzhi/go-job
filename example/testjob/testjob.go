package testjob

import (
	"fmt"
	"time"
)

type job struct{}

func (j *job) Name() string { return "TestJob" }

func (j *job) Desc() string { return "Job测试" }

func (j *job) AllowConcurrent() bool { return false }

func (j *job) Run() error {
	time.Sleep(time.Second * 8)
	fmt.Println("TestJob Run")
	return nil
}

var Job = &job{}
