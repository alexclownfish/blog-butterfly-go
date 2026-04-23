package services

import (
	"errors"
	"sync"
	"time"
)

var ErrCSDNSyncSessionNotFound = errors.New("csdn sync session not found")

type CSDNSyncSessionStore interface {
	Create(session *CSDNSyncSession) error
	Get(id string) (*CSDNSyncSession, error)
	Update(session *CSDNSyncSession) error
	Delete(id string) error
}

type MemoryCSDNSyncSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]CSDNSyncSession
	now      func() time.Time
}

func NewMemoryCSDNSyncSessionStore() *MemoryCSDNSyncSessionStore {
	return &MemoryCSDNSyncSessionStore{
		sessions: make(map[string]CSDNSyncSession),
		now:      time.Now,
	}
}

func (s *MemoryCSDNSyncSessionStore) Create(session *CSDNSyncSession) error {
	if session == nil || session.ID == "" {
		return errors.New("invalid session")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupExpiredLocked()
	s.sessions[session.ID] = cloneCSDNSyncSession(*session)
	return nil
}

func (s *MemoryCSDNSyncSessionStore) Get(id string) (*CSDNSyncSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupExpiredLocked()

	session, ok := s.sessions[id]
	if !ok {
		return nil, ErrCSDNSyncSessionNotFound
	}
	copied := cloneCSDNSyncSession(session)
	return &copied, nil
}

func (s *MemoryCSDNSyncSessionStore) Update(session *CSDNSyncSession) error {
	if session == nil || session.ID == "" {
		return errors.New("invalid session")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupExpiredLocked()
	if _, ok := s.sessions[session.ID]; !ok {
		return ErrCSDNSyncSessionNotFound
	}
	s.sessions[session.ID] = cloneCSDNSyncSession(*session)
	return nil
}

func (s *MemoryCSDNSyncSessionStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
	return nil
}

func (s *MemoryCSDNSyncSessionStore) cleanupExpiredLocked() {
	now := s.now()
	for id, session := range s.sessions {
		lastUpdatedAt := session.UpdatedAt
		if !session.ExpiresAt.IsZero() && !session.ExpiresAt.After(now) {
			session.Status = CSDNSyncSessionStatusExpired
			session.ErrorMessage = "登录会话已过期，请重新扫码"
			session.UpdatedAt = now
			s.sessions[id] = session
		}
		if session.Status == CSDNSyncSessionStatusExpired && !lastUpdatedAt.After(now.Add(-5*time.Minute)) {
			delete(s.sessions, id)
		}
	}
}

func cloneCSDNSyncSession(session CSDNSyncSession) CSDNSyncSession {
	copied := session
	if session.Articles != nil {
		copied.Articles = append([]CSDNSyncRemoteArticle(nil), session.Articles...)
	}
	return copied
}
