package idgen

type Resource interface {
	IDPrefix() string
}

type resource struct{}

func (r resource) IDPrefix() string { return "" }

type Request struct{}

func (r Request) IDPrefix() string { return "req" }

type Canvas struct{}

func (r Canvas) IDPrefix() string { return "cnv" }
