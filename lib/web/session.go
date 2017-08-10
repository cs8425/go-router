// Modify from https://github.com/astaxie/beego/tree/master/session
// "github.com/astaxie/beego/session"
//
// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

import (
	"container/list"
	"crypto/rand"
	"encoding/base64"
	"errors"
//	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Manager define the session config
type Manager struct {
	CookieName              string
	EnableSetCookie         bool
	Gclifetime              time.Duration
	Maxlifetime             time.Duration
	DisableHTTPOnly         bool   // disable HTTPOnly
	Secure                  bool   // force TLS
	CookieLifeTime          time.Duration // cookie Expires
	Domain                  string
	SessionIDLength         int64  // default = 18
	EnableSidInHTTPHeader   bool
	SessionNameInHTTPHeader string

	end         chan struct{}
	lock        sync.RWMutex             // locker
	sessions    map[string]*list.Element // map in memory
	list        *list.List               // for gc
}

// NewManager Create new Manager.
func NewManager() (*Manager, error) {
	return &Manager{
		CookieName: "sid",
		EnableSetCookie: true,
		Gclifetime: 5 * time.Minute,
		Maxlifetime: 30 * time.Minute,
		CookieLifeTime: 24 * time.Hour,
		SessionIDLength: 18,
		list: list.New(),
		sessions: make(map[string]*list.Element),
	}, nil
}

func NewPlugin(mng *Manager, tag string) (Session, error) {
	return &PluginSession{
		mng: mng,
		plugTag: tag,
	}, nil
}

func (mng *Manager) NewPlugin(tag string) (Session, error) {
	return &PluginSession{
		mng: mng,
		plugTag: tag,
	}, nil
}

// getSid retrieves session identifier from HTTP Request.
// First try to retrieve id by reading from cookie, session cookie name is configurable,
// if not exist, then retrieve id from querying parameters.
//
// error is not nil when there is anything wrong.
// sid is empty when need to generate a new session id
// otherwise return an valid session id.
func (manager *Manager) getSid(r *http.Request) (string, error) {
	cookie, errs := r.Cookie(manager.CookieName)
	if errs != nil || cookie.Value == "" {
		var sid string

		// if not found in Cookie / param, then read it from request headers
		if manager.EnableSidInHTTPHeader && sid == "" {
			sids, isFound := r.Header[manager.SessionNameInHTTPHeader]
			if isFound && len(sids) != 0 {
				return sids[0], nil
			}
		}

		return sid, nil
	}

	// HTTP Request contains cookie for sessionid info.
	return url.QueryUnescape(cookie.Value)
}

// SessionStart generate or read the session id from http request.
// if session id exists, return SessionStore with this id.
func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (session Store, err error) {
	sid, errs := manager.getSid(r)
	if errs != nil {
		return nil, errs
	}

	if sid != "" && manager.sessionExist(sid) {
		return manager.sessionRead(sid)
	}

	// Generate a new session
	sid, errs = manager.sessionID()
	if errs != nil {
		return nil, errs
	}

	session, err = manager.sessionRead(sid)
	if err != nil {
		return nil, err
	}
	cookie := &http.Cookie{
		Name:     manager.CookieName,
		Value:    sid,
		Path:     "/",
		HttpOnly: !manager.DisableHTTPOnly,
//		Secure:   manager.isSecure(r),
		Domain:   manager.Domain,
	}
	if manager.CookieLifeTime > 0 {
		cookie.MaxAge = int(manager.CookieLifeTime.Seconds())
		cookie.Expires = time.Now().Add(manager.CookieLifeTime)
	}
	if manager.EnableSetCookie {
		http.SetCookie(w, cookie)
	}
	r.AddCookie(cookie)

	if manager.EnableSidInHTTPHeader {
		r.Header.Set(manager.SessionNameInHTTPHeader, sid)
		w.Header().Set(manager.SessionNameInHTTPHeader, sid)
	}

	return
}

// SessionDestroy Destroy session by its id in http request cookie.
func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	if manager.EnableSidInHTTPHeader {
		r.Header.Del(manager.SessionNameInHTTPHeader)
		w.Header().Del(manager.SessionNameInHTTPHeader)
	}

	cookie, err := r.Cookie(manager.CookieName)
	if err != nil || cookie.Value == "" {
		return
	}

	sid, _ := url.QueryUnescape(cookie.Value)
	manager.sessionDestroy(sid)

	if manager.EnableSetCookie {
		expiration := time.Now()
		cookie = &http.Cookie{Name: manager.CookieName,
			Path:     "/",
			HttpOnly: !manager.DisableHTTPOnly,
			Expires:  expiration,
			MaxAge:   -1}

		http.SetCookie(w, cookie)
	}
}

// SessionRegenerateID Regenerate a session id for this SessionStore who's id is saving in http request.
func (manager *Manager) SessionRegenerateID(w http.ResponseWriter, r *http.Request) (session Store, err error) {
	sid, err := manager.sessionID()
	if err != nil {
		return
	}
	cookie, err := r.Cookie(manager.CookieName)
	if err != nil || cookie.Value == "" {
		//delete old cookie
		session, _ = manager.sessionRead(sid)
		cookie = &http.Cookie{Name: manager.CookieName,
			Value:    sid,
			Path:     "/",
			HttpOnly: !manager.DisableHTTPOnly,
//			Secure:   manager.isSecure(r),
			Domain:   manager.Domain,
		}
	} else {
		oldsid, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.sessionRegenerate(oldsid, sid)
		cookie.Value = sid
		cookie.HttpOnly = true
		cookie.Path = "/"
	}
	if manager.CookieLifeTime > 0 {
		cookie.MaxAge = int(manager.CookieLifeTime.Seconds())
		cookie.Expires = time.Now().Add(manager.CookieLifeTime)
	}
	if manager.EnableSetCookie {
		http.SetCookie(w, cookie)
	}
	r.AddCookie(cookie)

	if manager.EnableSidInHTTPHeader {
		r.Header.Set(manager.SessionNameInHTTPHeader, sid)
		w.Header().Set(manager.SessionNameInHTTPHeader, sid)
	}

	return
}

func (manager *Manager) sessionID() (string, error) {
	b := make([]byte, manager.SessionIDLength)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		return "", errors.New("Could not successfully read from the system CSPRNG")
	}
	return base64.URLEncoding.EncodeToString(b), nil
}


// sessionRead get memory session store by sid
func (pder *Manager) sessionRead(sid string) (Store, error) {
	pder.lock.RLock()
	if element, ok := pder.sessions[sid]; ok {
		go pder.sessionUpdate(sid)
		pder.lock.RUnlock()
		return element.Value.(*MemSessionStore), nil
	}
	pder.lock.RUnlock()
	pder.lock.Lock()
	newsess := newMemSessionStore(sid)
	element := pder.list.PushFront(newsess)
	pder.sessions[sid] = element
	pder.lock.Unlock()
	return newsess, nil
}

// sessionExist check session store exist in memory session by sid
func (pder *Manager) sessionExist(sid string) bool {
	pder.lock.RLock()
	defer pder.lock.RUnlock()
	if _, ok := pder.sessions[sid]; ok {
		return true
	}
	return false
}

// sessionRegenerate generate new sid for session store in memory session
func (pder *Manager) sessionRegenerate(oldsid, sid string) (Store, error) {
	pder.lock.RLock()
	if element, ok := pder.sessions[oldsid]; ok {
		go pder.sessionUpdate(oldsid)
		pder.lock.RUnlock()
		pder.lock.Lock()
		element.Value.(*MemSessionStore).sid = sid
		pder.sessions[sid] = element
		delete(pder.sessions, oldsid)
		pder.lock.Unlock()
		return element.Value.(*MemSessionStore), nil
	}
	pder.lock.RUnlock()
	pder.lock.Lock()
	newsess := newMemSessionStore(sid)
	element := pder.list.PushFront(newsess)
	pder.sessions[sid] = element
	pder.lock.Unlock()
	return newsess, nil
}

// SessionDestroy delete session store in memory session by id
func (pder *Manager) sessionDestroy(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	if element, ok := pder.sessions[sid]; ok {
		delete(pder.sessions, sid)
		pder.list.Remove(element)
		return nil
	}
	return nil
}

// SessionGC clean expired session stores in memory session
func (pder *Manager) SessionGC() {
	tick := time.NewTicker(pder.Gclifetime)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
//			fmt.Println("[SessionForceGC]start", pder.sessionAll())
			pder.SessionForceGC()
//			fmt.Println("[SessionForceGC]end", pder.sessionAll())
		case <-pder.end:
			return
		}
	}

}

func (pder *Manager) StopGC()  {
	select {
	case <- pder.end:
	default:
		close(pder.end)
	}
}

// SessionGC clean expired session stores in memory session
func (pder *Manager) SessionForceGC() {
	pder.lock.RLock()
	for {
		element := pder.list.Back()
		if element == nil {
			break
		}
		if element.Value.(*MemSessionStore).timeAccessed.Add(pder.Maxlifetime).After(time.Now()) {
			pder.lock.RUnlock()
			pder.lock.Lock()
			pder.list.Remove(element)
			delete(pder.sessions, element.Value.(*MemSessionStore).sid)
			pder.lock.Unlock()
			pder.lock.RLock()
		} else {
			break
		}
	}
	pder.lock.RUnlock()
}

// sessionAll get count number of memory session
func (pder *Manager) sessionAll() int {
	return pder.list.Len()
}

// sessionUpdate expand time of session store by id in memory session
func (pder *Manager) sessionUpdate(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	if element, ok := pder.sessions[sid]; ok {
		element.Value.(*MemSessionStore).timeAccessed = time.Now()
		pder.list.MoveToFront(element)
		return nil
	}
	return nil
}

// Session for plugin
type Session interface {
	Start(w http.ResponseWriter, r *http.Request) (session Store, err error)
	Destroy(w http.ResponseWriter, r *http.Request)
	RegenerateID(w http.ResponseWriter, r *http.Request) (session Store, err error)
}

type PluginSession struct {
	mng             *Manager
	plugTag         string   // main system = "/"
}

func (sess *PluginSession) Start(w http.ResponseWriter, r *http.Request) (Store, error) {
	return sess.getPlugin(sess.mng.SessionStart(w, r))
}

func (sess *PluginSession) Destroy(w http.ResponseWriter, r *http.Request) {
	sess.mng.SessionDestroy(w, r)
}

func (sess *PluginSession) RegenerateID(w http.ResponseWriter, r *http.Request) (Store, error) {
	return sess.getPlugin(sess.mng.SessionRegenerateID(w, r))
}

func (sess *PluginSession) getPlugin(basestore Store, err error) (Store, error) {
	if err!= nil {
		return nil, err
	}

	ps, ok := basestore.Get(sess.plugTag)
	if ok {
		return ps.(Store), nil
	}

	store := newMemSessionStore(basestore.SessionID())
	store.base = basestore

	if sess.plugTag != "/" {
		mainstore, ok := basestore.Get("/")
		if !ok {
//			fmt.Println("[new main session]", sess.plugTag, basestore.SessionID())
			mainstore = newMemSessionStore(basestore.SessionID())
			mainstore.(*MemSessionStore).base = basestore
			mainstore.(*MemSessionStore).mainstore = mainstore.(Store)
			basestore.Set("/", mainstore)
		}
		store.mainstore = mainstore.(Store)
	} else {
//		fmt.Println("[init main session]", sess.plugTag, basestore.SessionID())
		store.mainstore = store
	}

	basestore.Set(sess.plugTag, store)

	return store, nil
}


// Store contains all data for one session process with specific id.
type Store interface {
	Set(key string, value interface{}) (error)     //set session value
	Get(key string) (interface{}, bool)      //get session value
	Delete(key string) (error)         //delete session value
	SessionID() (string)                    //back current sessionID
	SessionRelease() // release the resource & save data to provider & return the data
	Flush() (error)                         //delete all data

	IsLogin() (bool)                          //current had Login?
	IsAdmin() (bool)                          //current Has Admin Permission?
//	GetCSRF() (string)                        //create CSRF token
}

// MemSessionStore memory session store.
// it saved sessions in a map in memory.
type MemSessionStore struct {
	sid          string                      //session id
	timeAccessed time.Time                   //last access time
	value        map[string]interface{} //session store
	mainstore    Store  // point to system session store
	base         Store  // point to root store
	lock         sync.RWMutex
}

// New memory session
func newMemSessionStore(sid string) (*MemSessionStore) {
	return &MemSessionStore{
		sid: sid,
		timeAccessed: time.Now(),
		value: make(map[string]interface{}),
	}
}

// Set value to memory session
func (st *MemSessionStore) Set(key string, value interface{}) (error) {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.value[key] = value
	return nil
}

// Get value from memory session by key
func (st *MemSessionStore) Get(key string) (interface{}, bool) {
	st.lock.RLock()
	defer st.lock.RUnlock()
	if v, ok := st.value[key]; ok {
		return v, true
	}
	return nil, false
}

// Delete in memory session by key
func (st *MemSessionStore) Delete(key string) (error) {
	st.lock.Lock()
	defer st.lock.Unlock()
	delete(st.value, key)
	return nil
}

// Flush clear all values in memory session
func (st *MemSessionStore) Flush() (error) {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.value = make(map[string]interface{})
	return nil
}

// SessionID get this id of memory session store
func (st *MemSessionStore) SessionID() (string) {
	return st.sid
}

// SessionRelease Implement method, no used.
func (st *MemSessionStore) SessionRelease() {
}

func (st *MemSessionStore) IsLogin() (bool) {
	status, ok := st.mainstore.Get("isLogin")
	if ok && status.(bool) {
		return true
	}
	return false
}

func (st *MemSessionStore) IsAdmin() (bool) {
	status, ok := st.mainstore.Get("isAdmin")
	if ok && status.(bool) {
		return true
	}
	return false
}

