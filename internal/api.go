package agent

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/kramergroup-workflows/lock-agent/api"
)

/*
Poller periodically polls the lock API for
released locks
*/
type Poller struct {
	endpoint string
	period   time.Duration
	stop     chan bool
}

/*
New creates a new Poller
*/
func NewPoller(api LockAPI, period time.Duration) Poller {

	return Poller{
		endpoint: api.endpoint,
		period:   period,
		stop:     nil,
	}

}

/*
Start the pooler. This call will activate the Poller
and periodically poll for released locks
*/
func (p *Poller) Start(callback func(lock.Lock)) {

	// First check if a poll thread is running and
	// return
	if p.stop != nil {
		return
	}

	p.stop = make(chan bool)

	go func() {
		for {
			p.poll(callback)
			select {
			case <-time.After(p.period):
			case <-p.stop:
				p.stop = nil
				return
			}
		}
	}()

}

/*
Stop stops polling for released locks
*/
func (p *Poller) Stop() {

	p.stop <- true

}

/*
poll implements the polling logic
*/
func (p *Poller) poll(callback func(lock.Lock)) {

	url, err := url.Parse(p.endpoint)
	if err != nil {
		log.Print("URLMalformated: ", err)
		return
	}

	q := url.Query()
	q.Set("status", "released")
	url.RawQuery = q.Encode()

	// Build the request
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		log.Print("NewRequest: ", err)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("Do: ", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print("Read: ", err)
			return
		}
		locks := make([]lock.Lock, 0)
		json.Unmarshal(body, &locks)

		for _, lock := range locks {
			callback(lock)
		}
	} else {
		log.Printf("Poll: API returned error - Response code: %d", resp.StatusCode)
	}

}

// -----------------------------------------------------------------------------

/*
LockAPI represents the Lock API endpoint
*/
type LockAPI struct {
	endpoint string
}

// NewLockAPI creates a new lock API instance
func NewLockAPI(endpoint string) LockAPI {
	return LockAPI{
		endpoint: endpoint,
	}
}

/*
DeleteLock deletes the lock with id
*/
func (a *LockAPI) DeleteLock(id string) {

	url, err := url.Parse(a.endpoint)
	if err != nil {
		log.Print("URLMalformated: ", err)
		return
	}

	q := url.Query()
	q.Set("id", id)
	url.RawQuery = q.Encode()

	// Build the request
	req, err := http.NewRequest("DELETE", url.String(), nil)
	if err != nil {
		log.Print("NewRequest: ", err)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("Do: ", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Print("DELETE Response error")
	}

}
