package scheduler

import (
	"time"

	"github.com/apex/log"
	"github.com/caarlos0/watchub"
	"github.com/caarlos0/watchub/config"
	"github.com/caarlos0/watchub/oauth"
	"github.com/gorilla/sessions"
	"github.com/robfig/cron"
)

var _ watchub.ScheduleSvc = &Scheduler{}

// Scheduler type
type Scheduler struct {
	cron              *cron.Cron
	config            config.Config
	oauth             *oauth.Oauth
	session           sessions.Store
	executions        watchub.ExecutionsSvc
	previousStars     watchub.StargazersSvc
	previousFollowers watchub.FollowersSvc
	currentStars      watchub.StargazersSvc
	currentFollowers  watchub.FollowersSvc
}

// Start the scheduler
func (s *Scheduler) Start() {
	var fn = func() {
		execs, err := s.executions.All()
		if err != nil {
			log.WithError(err).Error("failed to get executions")
			return
		}
		for _, exec := range execs {
			exec := exec
			go s.process(exec)
		}
	}
	s.cron = cron.New()
	if err := s.cron.AddFunc(s.config.Schedule, fn); err != nil {
		log.WithError(err).Fatal("failed to start cron service")
	}
	s.cron.Start()
}

// Stop the scheduler
func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
	s.cron = nil
}

func (s *Scheduler) process(exec watchub.Execution) {
	var start = time.Now()
	defer log.WithField("time_taken", time.Since(start).Seconds()).Info("done")
	var log = log.WithField("id", exec.UserID)
	previousStars, err := s.previousStars.Get(exec)
	if err != nil {
		log.WithError(err).Error("failed")
		return
	}
	previousFollowers, err := s.previousFollowers.Get(exec)
	if err != nil {
		log.WithError(err).Error("failed")
		return
	}
	currentStars, err := s.currentStars.Get(exec)
	if err != nil {
		log.WithError(err).Error("failed")
		return
	}
	currentFollowers, err := s.currentFollowers.Get(exec)
	if err != nil {
		log.WithError(err).Error("failed")
		return
	}
	if err := s.currentStars.Save(exec.UserID, currentStars); err != nil {
		log.WithError(err).Error("failed")
		return
	}
	if err := s.currentFollowers.Save(exec.UserID, currentFollowers); err != nil {
		log.WithError(err).Error("failed")
		return
	}
	// TODO: finish this
	log.Infof("%d %d %d %d", len(previousStars), len(currentStars), len(previousFollowers), len(currentFollowers))
}