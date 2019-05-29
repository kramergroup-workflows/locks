package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
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
queryParameter models the parameeters send as query strings to the API server as
part of CRUD calls.
*/
type queryParameter struct {
	Name  string
	Value string
}

/*
LockAPI represents the Lock API endpoint
*/
type LockAPI struct {
	endpoint string
}

// NewLockAPI creates a new lock API instance
func NewLockAPI(endpoint string) lock.API {
	return &LockAPI{
		endpoint: endpoint,
	}
}

/*
DeleteLock deletes the lock with id
*/
func (a *LockAPI) Delete(id string) error {

	params := []queryParameter{
		queryParameter{Name: "id", Value: id},
	}

	_, err := a.crud("DELETE", nil, params)
	if err != nil {
		fmt.Printf("ERROR DELETE: %s", err)
	}
	return err
}

/*
Release changes the status of a lock to released.
*/
func (a *LockAPI) Release(id string) error {

	params := []queryParameter{
		queryParameter{Name: "id", Value: id},
	}

	_, err := a.crud("PATCH", nil, params)
	if err != nil {
		fmt.Printf("ERROR PATCH: %s", err)
	}
	return err
}

/*
Create registers a new lock with the API server, sets its status to "locked"
and returns the created lock
*/
func (a *LockAPI) Create(workflow string, namespace string) (lock.Lock, error) {

	lock := lock.Lock{
		Workflow:  workflow,
		Namespace: namespace,
	}

	locJSON, err := json.Marshal(lock)

	res, err := a.crud("POST", locJSON, nil)
	if err != nil {
		fmt.Printf("ERROR POST: %s", err)
		return lock, err
	}

	if err := json.Unmarshal(res, &lock); err != nil {
		log.Printf("ERROR ParseGetResponse: %s", err)
		return lock, err
	}

	return lock, nil

}

/*
Get obtains the lock with id
*/
func (a *LockAPI) Get(id string) (lock.Lock, error) {

	var lock lock.Lock

	params := []queryParameter{
		queryParameter{Name: "id", Value: id},
	}

	body, err := a.crud("GET", nil, params)
	if err != nil {
		fmt.Printf("ERROR PATCH: %s", err)
		return lock, err
	}

	err = json.Unmarshal([]byte(body), &lock)
	return lock, err
}

/*
crud is the main communication routine to send requests to the API server and collects
the response in a []byte array.

The method returns an error if the returned status code differs from 200.
*/
func (a *LockAPI) crud(method string, body []byte, queryParameters []queryParameter) ([]byte, error) {

	url, err := url.Parse(a.endpoint)
	if err != nil {
		log.Fatal("URLMalformated: ", err)
		return nil, err
	}

	q := url.Query()
	for _, param := range queryParameters {
		q.Set(param.Name, param.Value)
	}
	url.RawQuery = q.Encode()

	// Build the request
	req, err := http.NewRequest(method, url.String(), bytes.NewBuffer(body))
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		return body, err
	}

	return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)

}
