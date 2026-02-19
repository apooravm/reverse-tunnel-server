package utils

import (
	"sync"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

var (
	VHM = VH_Manager{
		vHosts: make(map[string]*echo.Echo),
	}
)

type VH_Manager struct {
	vHosts map[string]*echo.Echo
	mu     sync.RWMutex
}

func (vhm *VH_Manager) Add_Host(hostname string) {
	vhm.mu.Lock()
	defer vhm.mu.Unlock()

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	vhm.vHosts[hostname+".localhost:8000"] = e
}

func (vhm *VH_Manager) Get_Host(hostname string) *echo.Echo {
	vhm.mu.Lock()
	defer vhm.mu.Unlock()

	return vhm.vHosts[hostname+".localhost:8000"]
}
