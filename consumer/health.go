package consumer

import (
	"encoding/json"
	"github.com/denismitr/auditbase/model"
	"github.com/pkg/errors"
	"net/http"
	"os"
)

type health struct {
	StatusOK        bool           `json:"statusOk"`
	PersistedEvents int            `json:"persistedEvents"`
	FailedEvents    int            `json:"failedActionCreations"`
	StartedAt       model.JSONTime `json:"startedAt"`
	FailedAt        model.JSONTime `json:"failedAt"`
}

func (c *Consumer) healthCheck(stopCh <-chan os.Signal) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		c.mu.Lock()
		defer c.mu.Unlock()

		var h health
		h.StatusOK = c.statusOK
		h.PersistedEvents = c.persistedEvents
		h.FailedEvents = c.failedActionCreations
		h.StartedAt = model.JSONTime{Time: c.startedAt}
		h.FailedAt = model.JSONTime{Time: c.failedAt}

		if c.statusOK {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(500)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(h); err != nil {
			c.lg.Error(errors.Wrap(err, "health endpoint failed"))
		}
	})

	c.lg.Debugf("\nStarting healthcheck on port %s", os.Getenv("HEALTH_PORT"))
	err := http.ListenAndServe(":"+os.Getenv("HEALTH_PORT"), nil)
	if err != nil {
		c.lg.Error(errors.Wrap(err, "helthcheck endpoint failed"))
	}
}
