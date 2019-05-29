package lock

import "time"

/*
Lock represents a simple lock structure
*/
type Lock struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	Workflow   string    `json:"workflow"`
	Namespace  string    `json:"namespace"`
	Created    time.Time `json:"created"`
	LastChange time.Time `json:"lastChange"`
}

/*
API provides methods to interact with the lock API server
*/
type API interface {
	Delete(id string) error
	Release(id string) error
	Create(workflow string, namespace string) (Lock, error)
	Get(id string) (Lock, error)
}
