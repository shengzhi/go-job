package job

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	"github.com/robfig/cron"
	"github.com/shengzhi/util/snowflake"
)

// IJob job接口
type IJob interface {
	// Name Job名称
	Name() string
	// Desc Job 描述
	Desc() string
	// Run 执行Job
	Run() error
	// AllowConcurrent 是否允许多个job实例并发执行
	AllowConcurrent() bool
}

type RunStatus uint8

func (s RunStatus) String() string {
	switch s {
	case Ready:
		return "Ready(First)"
	case Running:
		return "Running"
	case Sleeping:
		return "Sleeping"
	default:
		return "Unknown"
	}
}

const (
	Ready RunStatus = iota
	Running
	Sleeping
)

type jobWrapper struct {
	job             IJob
	schedule        cron.Schedule
	runHis          *RunHisList
	prevExecTime    time.Time
	status          RunStatus
	allowConcurrent bool
}

func (jw *jobWrapper) Next() time.Time {
	return jw.schedule.Next(time.Now())
}

var jobIDGen, _ = snowflake.NewIdWorker(int64(os.Getpid()))

func (jw *jobWrapper) Run() {
	instanceid, _ := jobIDGen.NextId()
	log.Printf("%s - %d begin run \r\n", jw.job.Name(), instanceid)
	if !jw.allowConcurrent && jw.status == Running {
		log.Println("not allow concurrent")
		return
	}
	start := time.Now()
	jw.status = Running
	defer func() { jw.status = Sleeping }()
	err := jw.job.Run()
	end := time.Now()
	his := RunHistory{
		ID:        instanceid,
		StartTime: start, CompletedTime: end,
		Elased: int64(end.Sub(start).Seconds()),
	}
	if err != nil {
		his.Result = "Failed"
		his.ErrMsg = err.Error()
	} else {
		his.Result = "Success"
	}
	jw.runHis.add(his)
	jw.prevExecTime = start
	log.Printf("%s - %d end run \r\n", jw.job.Name(), instanceid)
}

type OptionFunc func(jm *JobManager)

// JobManager Job 管理器
type JobManager struct {
	c              *cron.Cron
	path           string // 插件目录
	jobs           map[string]*jobWrapper
	scheduleConf   map[string]string // 计划配置
	latestLoadTime time.Time         // 最近一次Load Job的时间
	location       *time.Location
	httpPort       int
}

// NewJobManager creates job manager
func NewJobManager(options ...OptionFunc) *JobManager {
	jm := &JobManager{
		scheduleConf: make(map[string]string),
		jobs:         make(map[string]*jobWrapper),
		location:     time.Local,
		httpPort:     8080,
	}
	for _, fn := range options {
		fn(jm)
	}
	jm.c = cron.NewWithLocation(jm.location)
	return jm
}

// SetTimeZone 设置时区
func (m *JobManager) setTimeZone(name string) {
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Fatalln(err)
	}
	m.location = loc
}

// Stop 停止Job
func (m *JobManager) Stop() { m.c.Stop() }

// Start 启动Job管理程序
func (m *JobManager) Start() error {
	if err := m.addJobs(); err != nil {
		return err
	}
	go m.listenHTTP()
	m.c.Start()
	select {}
}

// AddJob 增加Job
func (m *JobManager) AddJob(spec string, cmd IJob) error {
	schedule, err := cron.Parse(spec)
	if err != nil {
		return err
	}
	jw := &jobWrapper{
		job:             cmd,
		schedule:        schedule,
		runHis:          NewRunHisList(10),
		allowConcurrent: cmd.AllowConcurrent(),
	}
	m.jobs[strings.ToLower(cmd.Name())] = jw
	m.c.Schedule(jw.schedule, jw)
	return nil
}

func (m *JobManager) addJobs() error {
	jobs := m.loadJob()
	for _, cmd := range jobs {
		jobName := strings.ToLower(cmd.Name())
		if _, has := m.jobs[jobName]; has {
			continue
		}
		spec, has := m.scheduleConf[jobName]
		if !has {
			log.Println("Not specify schedule for job ", jobName)
			continue
		}
		schedule, err := cron.Parse(spec)
		if err != nil {
			return err
		}
		jw := &jobWrapper{
			job:             cmd,
			schedule:        schedule,
			runHis:          NewRunHisList(10),
			allowConcurrent: cmd.AllowConcurrent(),
		}
		m.jobs[jobName] = jw
		m.c.Schedule(jw.schedule, jw)
	}
	m.latestLoadTime = time.Now()
	return nil
}

func (m *JobManager) loadJob() []IJob {
	files, _ := filepath.Glob(fmt.Sprintf("%s/*.so", m.path))
	jobs := make([]IJob, 0, len(files))
	for _, file := range files {
		p, err := plugin.Open(file)
		if err != nil {
			log.Printf("load plugin file %s failed,error: %v\r\n", file, err)
			continue
		}
		fi, err := os.Stat(file)
		if err != nil {
			continue
		}
		if fi.ModTime().Before(m.latestLoadTime) {
			continue
		}
		j, err := p.Lookup("Job")
		if err != nil {
			continue
		}
		if ijob, ok := j.(IJob); ok {
			jobs = append(jobs, ijob)
		}
	}
	return jobs
}

type jobInfo struct {
	Name, Desc             string
	PrevExecTime, NextTime string
	Status                 string
	His                    []RunHistory
}

// AllJobs 返回所有Job信息
func (m *JobManager) AllJobs() []jobInfo {
	result := make([]jobInfo, 0, len(m.jobs))
	for _, v := range m.jobs {
		info := jobInfo{
			Name:     v.job.Name(),
			Desc:     v.job.Desc(),
			NextTime: v.Next().Format("2006-01-02 15:04:05"),
			Status:   v.status.String(),
			His:      v.runHis.His(),
		}
		if !v.prevExecTime.IsZero() {
			info.PrevExecTime = v.prevExecTime.Format("2006-01-02 15:04:05")
		}
		result = append(result, info)
	}
	return result
}
