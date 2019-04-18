package tasks

import (
	"cig-exchange-libs/models"
	"time"
)

func getDurationTillNoon() time.Duration {

	now := time.Now()
	noon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	if now.After(noon) {
		noon = noon.Add(24 * time.Hour)
	}
	duration := noon.Sub(now)
	return duration
}

func invitationExpirationTask() {

	for {
		// sleep till noon
		time.Sleep(getDurationTillNoon())
		models.DeleteExpiredInvitations()
	}
}

// ScheduleTasks starts goroutines for sheduled tasks
func ScheduleTasks() {

	go invitationExpirationTask()
}
