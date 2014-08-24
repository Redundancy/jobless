package jobless

type ChainedVariableStore struct {
	Parent    *ChainedVariableStore
	Variables map[string]string
}

func NewChainedVariableStore(parent *ChainedVariableStore) *ChainedVariableStore {
	return &ChainedVariableStore{
		Parent:    parent,
		Variables: make(map[string]string),
	}
}

func (s *ChainedVariableStore) Get(key string) (string, bool) {
	value, hasLocal := s.Variables[key]

	if hasLocal {
		return value, true
	} else if s.Parent != nil {
		return s.Parent.Get(key)
	} else {
		return "", false
	}
}
