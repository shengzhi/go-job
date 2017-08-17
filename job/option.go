package job

import "strings"

func SetTimeZone(zone string) OptionFunc {
	return func(jm *JobManager) {
		jm.setTimeZone(zone)
	}
}

func SetPluginPath(path string) OptionFunc {
	return func(jm *JobManager) {
		jm.path = path
	}
}

func SetHTTPPort(port int) OptionFunc {
	return func(jm *JobManager) {
		jm.httpPort = port
	}
}

// SetSchedule 设置Job计划
func SetSchedule(jobName, spec string) OptionFunc {
	return func(jm *JobManager) {
		jm.scheduleConf[strings.ToLower(jobName)] = spec
	}
}

// SetScheduleBatch 批量设置Job计划
func SetScheduleBatch(batch map[string]string) OptionFunc {
	return func(jm *JobManager) {
		for k, v := range batch {
			jm.scheduleConf[k] = v
		}
	}
}
