package gapc

type TmplData struct {
	CmdLine string
	// Version string
	PackageName    string
	FuncName       string
	PackagePath    string
	Group          string
	Topic          string
	HasRecv        bool
	Dependencies   []Dependency
	DependencyList string
	MsgType        string
	TopicExpr      string
	IsEvent        bool
}

type Flags struct {
	Dir        string
	FuncNames  []string
	FileName   string
	Annotation string
	Group      string
	Separate   bool
	Verbose    bool
	Raw        bool
}

type Dependency struct {
	Name string
	Type string
}
