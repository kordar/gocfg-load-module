package gocfgmodule

type SingletonLoad func(moduleName string, id string, cfg map[string]string)
type SingletonLoadI func(moduleName string, id string, cfg map[string]interface{})
