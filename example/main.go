package main

import (
	"go-job/example/testjob"
	"go-job/job"
	"log"
)

func main() {
	jm := job.NewJobManager(job.SetTimeZone("Asia/Bangkok"))
	jm.AddJob("0/3 * * * * ? ", testjob.Job)

	log.Fatalln(jm.Start())
}
