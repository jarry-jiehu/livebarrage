package comet

import (
	"strings"
	"sync"
)

type ClientManager struct {
	lock *sync.RWMutex
	mp   map[string]*Client
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		lock: new(sync.RWMutex),
		mp:   make(map[string]*Client),
	}
}

func (this *ClientManager) get(k string) *Client {
	this.lock.RLock()
	r, ok := this.mp[k]
	if !ok {
		r = nil
	}
	this.lock.RUnlock()
	return r
}

func (this *ClientManager) getAll(liveId string) []*Client {
	clients := []*Client{}
	this.lock.RLock()
	for _, client := range this.mp {
		if client.status == CLIENT_CONNECTED || client.status == CLIENT_READY {
			clients = append(clients, client)
		}
	}
	this.lock.RUnlock()
	return clients
}

func (this *ClientManager) getByLiveId(liveId string) []*Client {
	clients := []*Client{}
	this.lock.RLock()
	for _, client := range this.mp {
		if client.status == CLIENT_CONNECTED || client.status == CLIENT_READY {
			if strings.Compare(client.liveId, liveId) == 0 {
				clients = append(clients, client)
			}
		}
	}
	this.lock.RUnlock()
	return clients
}

func (this *ClientManager) Set(k string, v *Client) bool {
	this.lock.Lock()
	r := true
	if val, ok := this.mp[k]; !ok {
		this.mp[k] = v
	} else if val != v {
		this.mp[k] = v
	} else {
		r = false
	}
	this.lock.Unlock()
	return r
}

func (this *ClientManager) Remove(k string) {
	this.lock.Lock()
	delete(this.mp, k)
	this.lock.Unlock()
}

func (this *ClientManager) Size() int {
	this.lock.RLock()
	r := len(this.mp)
	this.lock.RUnlock()
	return r
}
