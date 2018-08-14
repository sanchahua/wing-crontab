package utility

import ()

type PathPattern struct {
	path  string
	level []string
}

func NewPathPattern(path string) *PathPattern {
	p := &PathPattern{
		path:  path,
		level: make([]string, 0),
	}
	p.parse()

	return p
}

func (p *PathPattern) Depth() int {

	return len(p.level)
}

func (p *PathPattern) LevelName(i int) string {
	if i >= p.Depth() {
		return ""
	}

	return p.level[i]
}

func (p *PathPattern) parse() error {

	const dash = '/'

	var lastP = 0
	for i := 0; i < len(p.path); {

		var c byte

		// Skip dashes
		for i < len(p.path) {
			c = p.path[i]
			if c != dash {
				break
			}

			i++
		}
		if i == len(p.path) {
			break
		}
		lastP = i

		// Next level
		for i < len(p.path) {

			c = p.path[i]
			if c == dash {
				break
			}

			i++
		}

		newLevel := p.path[lastP:i]
		p.level = append(p.level, newLevel)
	}

	return nil
}

func (p *PathPattern) Format(fromLevel int) string {
	s := ""
	for i, v := range p.level {
		if i < fromLevel {
			continue
		}
		s += "/"
		s += v
	}
	return s
}
