package reflector

import (
	"fmt"
	"sync"
	"time"

	"github.com/lbryio/lbry.go/errors"
	"github.com/lbryio/lbry.go/stop"

	log "github.com/sirupsen/logrus"
)

type stats struct {
	mu      *sync.Mutex
	blobs   int
	streams int
	errors  map[string]int

	logger  *log.Logger
	logFreq time.Duration
	grp     *stop.Group
}

func newStatLogger(logger *log.Logger, logFreq time.Duration, parentGrp *stop.Group) *stats {
	return &stats{
		mu:      &sync.Mutex{},
		grp:     stop.New(parentGrp),
		logger:  logger,
		logFreq: logFreq,
		errors:  make(map[string]int),
	}
}

func (s *stats) Start() {
	s.grp.Add(1)
	go func() {
		defer s.grp.Done()
		s.runSlackLogger()
	}()
}

func (s *stats) Shutdown() {
	s.grp.StopAndWait()
}

func (s *stats) AddBlob() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blobs += 1
}
func (s *stats) AddStream() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streams += 1
}
func (s *stats) AddError(e error) {
	if e == nil {
		return
	}
	err := errors.Wrap(e, 0)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors[err.TypeName()] += 1
}

func (s *stats) runSlackLogger() {
	t := time.NewTicker(s.logFreq)
	for {
		select {
		case <-s.grp.Ch():
			return
		case <-t.C:
			s.log()
		}
	}
}

func (s *stats) log() {
	s.mu.Lock()
	blobs, streams := s.blobs, s.streams
	s.blobs, s.streams = 0, 0
	errStr := ""
	for name, count := range s.errors {
		errStr += fmt.Sprintf("%d %s, ", count, name)
		delete(s.errors, name)
	}
	s.mu.Unlock()
	s.logger.Printf(
		"Stats: %d blobs, %d streams, errors: %s",
		blobs, streams, errStr[:len(errStr)-2], // trim last comma and space
	)
}
