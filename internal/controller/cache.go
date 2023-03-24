package controller

import "sync"

type LocalCache struct {
	Cache map[string]interface{}
	Mu    sync.Mutex
}
