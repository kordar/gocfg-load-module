package starter

type LoadModule interface {
	Name() string
	Load(data interface{})
}

var (
	modules = map[string]LoadModule{}
)

func RegModule(m LoadModule) {
	modules[m.Name()] = m
}

func Resolve(name string, setting interface{}) {
	if modules[name] != nil {
		modules[name].Load(setting)
	}
}
