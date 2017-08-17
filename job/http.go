package job

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (m *JobManager) listenHTTP() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		jobs := m.AllJobs()
		data, _ := json.Marshal(jobs)
		w.Write(data)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	})
	http.ListenAndServe(fmt.Sprintf(":%d", m.httpPort), nil)
}
