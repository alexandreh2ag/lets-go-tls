package types

type Requesters map[string]Requester

func (p Requesters) Get(key string) Requester {
	if requester, ok := p[key]; ok {
		return requester
	}
	return nil
}

type Requester interface {
	ID() string
	Fetch() ([]*DomainRequest, error)
}
