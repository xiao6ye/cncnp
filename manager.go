// Package zcnp
// @Title  zcnp
// @Description  zcnp
// @Author  zxx1224@gmail.com  2021/8/11 3:43 下午
// @Update  zxx1224@gmail.com  2021/8/11 3:43 下午
package zcnp

import (
	"log"
	"sync"
)

const sessionMapNum = 32

type Manager struct {
	sessionMaps [sessionMapNum]sessionMap
	disposeOnce sync.Once
	disposeWait sync.WaitGroup
}

type sessionMap struct {
	sync.RWMutex
	sessions map[uint64]*Session
	disposed bool
}

// GetSessions 获取Sessions
func (sm *sessionMap) GetSessions() map[uint64]*Session {
	sessions := sm.sessions
	return sessions
}

func NewManager() *Manager {
	manager := &Manager{}
	for i := 0; i < len(manager.sessionMaps); i++ {
		manager.sessionMaps[i].sessions = make(map[uint64]*Session)
	}
	return manager
}

// Dispose 关闭所有会话
func (manager *Manager) Dispose() {
	manager.disposeOnce.Do(func() {
		for i := 0; i < sessionMapNum; i++ {
			smap := &manager.sessionMaps[i]
			smap.Lock()
			smap.disposed = true
			for _, session := range smap.sessions {
				session.Close()
			}
			smap.Unlock()
		}
		manager.disposeWait.Wait()
	})
}

// NewSession 新建会话
func (manager *Manager) NewSession(codec Codec, sendChanSize int) *Session {
	session := newSession(manager, codec, sendChanSize)
	manager.putSession(session)
	return session
}

// GetSession 从会话池获取会话Session
func (manager *Manager) GetSession(sessionID uint64) *Session {
	smap := &manager.sessionMaps[sessionID%sessionMapNum]
	smap.RLock()
	defer smap.RUnlock()

	session, _ := smap.sessions[sessionID]
	return session
}

// GetSessionMaps 获取SessionMaps
func (manager *Manager) GetSessionMaps() [sessionMapNum]sessionMap {
	sMaps := manager.sessionMaps
	return sMaps
}

// 放入会话Map
func (manager *Manager) putSession(session *Session) {
	log.Println("新增TCP会话。。。。。", session.id)
	smap := &manager.sessionMaps[session.id%sessionMapNum]

	smap.Lock()
	defer smap.Unlock()

	if smap.disposed {
		session.Close()
		return
	}

	smap.sessions[session.id] = session
	manager.disposeWait.Add(1)
}

// 会话结束后，删除管理Map 里面的会话句柄
func (manager *Manager) delSession(session *Session) {
	smap := &manager.sessionMaps[session.id%sessionMapNum]

	smap.Lock()
	defer smap.Unlock()

	if _, ok := smap.sessions[session.id]; ok {
		delete(smap.sessions, session.id)
		manager.disposeWait.Done()
	}
}
