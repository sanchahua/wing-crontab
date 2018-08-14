package utility

type ScopeParser struct {
	ServerName string
	MethodName string
}

func NewScopeParser(path string) *ScopeParser {
	parser := &ScopeParser{}

	p := NewPathPattern(path)
	if p.Depth() == 0 {
		return parser
	}

	parser.ServerName = p.LevelName(0)
	parser.MethodName = p.Format(1)

	return parser
}

func (p *ScopeParser) Format() string {
	return "/" + p.ServerName + p.MethodName
}
