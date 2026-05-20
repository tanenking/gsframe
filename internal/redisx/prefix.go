package redisx

type Prefix string

func (p Prefix) String() string {
	return string(p)
}

func (p Prefix) ColonStr() string {
	n := len(p)
	if n == 0 {
		return ""
	}
	str := p.String()
	if str[n-1] == ':' {
		return str
	}
	return str + ":"
}

func (p Prefix) MakeKey(key string) string {
	return p.ColonStr() + key
}
