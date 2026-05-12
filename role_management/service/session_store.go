package service

import (
	"crypto/rand"
	"encoding/base64"
	"role-management/entity"
	"sync"
	"time"

	"gorm.io/gorm"
)

type Session struct {
	UserID    uint64
	ExpiresAt time.Time
}

type SessionStore interface {
	Create(userID uint64, ttl time.Duration) (string, Session)
	Get(token string) (Session, bool)
	Delete(token string)
	DeleteByUser(userID uint64) int64
	CleanupExpired() int64
}

type inMemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

type dbSessionStore struct {
	db *gorm.DB
}

func NewInMemorySessionStore() SessionStore {
	return &inMemorySessionStore{sessions: map[string]Session{}}
}

func NewDBSessionStore(db *gorm.DB) SessionStore {
	return &dbSessionStore{db: db}
}

func (s *inMemorySessionStore) Create(userID uint64, ttl time.Duration) (string, Session) {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	token := base64.RawURLEncoding.EncodeToString(buf)
	session := Session{UserID: userID, ExpiresAt: time.Now().Add(ttl)}

	s.mu.Lock()
	s.sessions[token] = session
	s.mu.Unlock()

	return token, session
}

func (s *inMemorySessionStore) Get(token string) (Session, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok {
		return Session{}, false
	}
	if time.Now().After(sess.ExpiresAt) {
		s.Delete(token)
		return Session{}, false
	}
	return sess, true
}

func (s *inMemorySessionStore) Delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func (s *inMemorySessionStore) DeleteByUser(userID uint64) int64 {
	var deleted int64
	s.mu.Lock()
	for token, session := range s.sessions {
		if session.UserID == userID {
			delete(s.sessions, token)
			deleted++
		}
	}
	s.mu.Unlock()
	return deleted
}

func (s *inMemorySessionStore) CleanupExpired() int64 {
	now := time.Now()
	var deleted int64
	s.mu.Lock()
	for token, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, token)
			deleted++
		}
	}
	s.mu.Unlock()
	return deleted
}

func (s *dbSessionStore) Create(userID uint64, ttl time.Duration) (string, Session) {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	token := base64.RawURLEncoding.EncodeToString(buf)
	session := Session{UserID: userID, ExpiresAt: time.Now().Add(ttl)}

	_ = s.db.Create(&entity.SessionToken{Token: token, UserID: userID, ExpiresAt: session.ExpiresAt}).Error
	return token, session
}

func (s *dbSessionStore) Get(token string) (Session, bool) {
	var row entity.SessionToken
	if err := s.db.Where("token = ?", token).First(&row).Error; err != nil {
		return Session{}, false
	}
	if time.Now().After(row.ExpiresAt) {
		s.Delete(token)
		return Session{}, false
	}
	return Session{UserID: row.UserID, ExpiresAt: row.ExpiresAt}, true
}

func (s *dbSessionStore) Delete(token string) {
	_ = s.db.Where("token = ?", token).Delete(&entity.SessionToken{}).Error
}

func (s *dbSessionStore) DeleteByUser(userID uint64) int64 {
	result := s.db.Where("user_id = ?", userID).Delete(&entity.SessionToken{})
	if result.Error != nil {
		return 0
	}
	return result.RowsAffected
}

func (s *dbSessionStore) CleanupExpired() int64 {
	result := s.db.Where("expires_at <= ?", time.Now()).Delete(&entity.SessionToken{})
	if result.Error != nil {
		return 0
	}
	return result.RowsAffected
}
