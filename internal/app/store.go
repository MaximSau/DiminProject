package app

import (
	"pulsefit/internal/models"
	"pulsefit/internal/repository"
)

type User = models.User
type Profile = models.Profile
type Reminder = models.Reminder
type WeightEntry = models.WeightEntry
type SessionEntry = models.SessionEntry
type Checkin = models.Checkin
type Progress = models.Progress
type ChatMessage = models.ChatMessage
type Store = repository.SQLiteRepository

func OpenStore(path string) (*Store, error) {
	return repository.Open(path)
}
