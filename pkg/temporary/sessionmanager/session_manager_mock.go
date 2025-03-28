package sessionmanager

import (
	"sync"
)

// SessionManager is a mock implementation of the SessionManager interface.
type SessionManager struct {
	mu sync.Mutex
	// sessions is a map of sessionNonce to a Session
	sessions map[string]PeerSession
	// identityKeyToSessions is a map of peerIdentityKey to a list of sessionNonce's
	identityKeyToSessions map[string][]string
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions:              make(map[string]PeerSession),
		identityKeyToSessions: make(map[string][]string),
	}
}

// AddSession adds a session to the manager, associating it with its sessionNonce and also with its peerIdentityKey.
func (m *SessionManager) AddSession(session PeerSession) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session.SessionNonce != nil {
		m.sessions[*session.SessionNonce] = session
	}

	if session.PeerIdentityKey != nil {
		m.addSessionByIdentityKey(session)
	}
}

// addSessionByIdentityKey adds a session nonce to the manager by associating it with its peerIdentityKey.
// This does NOT overwrite existing sessions for the same peerIdentityKey, allowing multiple concurrent sessions.
func (m *SessionManager) addSessionByIdentityKey(session PeerSession) {
	sessionNonces, exists := m.identityKeyToSessions[*session.PeerIdentityKey]
	if exists {
		// append sessionNonce to existing list
		// at this point we have at least two concurrent sessions for the same peerIdentityKey
		m.identityKeyToSessions[*session.PeerIdentityKey] = append(sessionNonces, *session.SessionNonce)
		return
	}

	// create new list with sessionNonce and assign to peerIdentityKey
	m.identityKeyToSessions[*session.PeerIdentityKey] = []string{*session.SessionNonce}
}

// GetSession retrieves a "best" session based on a given identifier, which can be a sessionNonce or a peerIdentityKey.
func (m *SessionManager) GetSession(identifier string) *PeerSession {
	m.mu.Lock()
	defer m.mu.Unlock()

	// try to get session by sessionNonce
	if session, exists := m.sessions[identifier]; exists {
		return &session
	}

	// check if sessions exists by peerIdentityKey
	sessionNonces, exists := m.identityKeyToSessions[identifier]
	if !exists {
		return nil
	}

	// get the "best" session
	bestSession := m.getBestSession(sessionNonces)

	return bestSession
}

// getBestSession retrieves the "best" session from a list of sessionNonces.
// The "best" session is the most recent one, or the most recent authenticated one if there are multiple.
func (m *SessionManager) getBestSession(sessionNonces []string) *PeerSession {
	var bestSession *PeerSession
	for _, sessionNonce := range sessionNonces {
		session, exists := m.sessions[sessionNonce]
		if !exists {
			continue
		}

		// If no session is selected yet, set the current session
		if bestSession == nil {
			bestSession = &session
			continue
		}

		// If the current session is authenticated and the bestSession is not, update bestSession
		if session.IsAuthenticated && !bestSession.IsAuthenticated {
			bestSession = &session
			continue
		}

		// If both are authenticated or both are not, select the most recent one
		if session.IsAuthenticated == bestSession.IsAuthenticated && session.LastUpdate.After(bestSession.LastUpdate) {
			bestSession = &session
		}
	}
	return bestSession
}

// RemoveSession removes a session from the manager by clearing all associated identifiers.
func (m *SessionManager) RemoveSession(session PeerSession) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session.SessionNonce != nil {
		delete(m.sessions, *session.SessionNonce)
	}

	if session.PeerIdentityKey != nil {
		sessionNonces, exists := m.identityKeyToSessions[*session.PeerIdentityKey]
		if !exists {
			return
		}

		updatedNonces := removeSessionNonce(sessionNonces, *session.SessionNonce)

		// if there are no more sessions for the peerIdentityKey, remove the key
		if len(updatedNonces) == 0 {
			delete(m.identityKeyToSessions, *session.PeerIdentityKey)
			return
		}

		// update the list of sessionNonces for the peerIdentityKey
		m.identityKeyToSessions[*session.PeerIdentityKey] = updatedNonces
	}
}

// HasSession checks if a session exists for a given identifier (either sessionNonce or identityKey).
func (m *SessionManager) HasSession(identifier string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// check if session exists by sessionNonce
	_, exists := m.sessions[identifier]
	if exists {
		return true
	}

	// check if sessions are assigned to peerIdentityKey
	nonces, exists := m.identityKeyToSessions[identifier]
	if !exists {
		return false
	}

	return len(nonces) > 0
}

// UpdateSession updates a session in the manager.
func (m *SessionManager) UpdateSession(session PeerSession) {
	m.AddSession(session)
}

func removeSessionNonce(slice []string, target string) []string {
	newSlice := slice[:0] // Reuse the same slice memory
	for _, str := range slice {
		if str != target {
			newSlice = append(newSlice, str)
		}
	}
	return newSlice
}
