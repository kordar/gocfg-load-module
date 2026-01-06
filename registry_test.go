package gocfgmodule_test

import (
	"fmt"
	"testing"

	gocfgmodule "github.com/kordar/gocfg-load-module"
)

type DBModule struct{}

func (d *DBModule) Name() string { return "db" }

func (d *DBModule) BeforeLoad() {
	fmt.Println("db before load")
}

func (d *DBModule) Load(cfg interface{}) {
	fmt.Println("db load", cfg)
}

func (d *DBModule) AfterLoad() {
	fmt.Println("db after load")
}

func (d *DBModule) Close() {
	fmt.Println("db close")
}

// ----------------------------------------
type CacheModule struct{}

func (d *CacheModule) Name() string { return "cache" }

func (d *CacheModule) BeforeLoad() {
	fmt.Println("cache before load")
}

func (d *CacheModule) Load(cfg interface{}) {
	fmt.Println("cache load", cfg)
}

func (d *CacheModule) AfterLoad() {
	fmt.Println("cache after load")
}

func (d *CacheModule) Close() {
	fmt.Println("cache close")
}

// ---------------------------
type HTTPModule struct{}

func (d *HTTPModule) Name() string { return "http" }

func (d *HTTPModule) BeforeLoad() {
	fmt.Println("http before load")
}

func (d *HTTPModule) Load(cfg interface{}) {
	fmt.Println("http load", cfg)
}

func (d *HTTPModule) AfterLoad() {
	fmt.Println("http after load")
}

func (d *HTTPModule) Close() {
	fmt.Println("http close")
}

func TestMain(t *testing.T) {
	cfg := gocfgmodule.New()

	cfg.Register(&DBModule{})
	cfg.Register(&CacheModule{}, "db")
	cfg.RegisterRequired(&HTTPModule{}, "db", "cache")

	cfg.RefreshDepends(nil)

	cfg.ResolveAll(map[string]interface{}{
		"db": map[string]string{"dsn": "..."},
	})

	defer cfg.Destroy()

}
