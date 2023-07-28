package cardinality

type Replacer interface {
	Replace(path string) string
}

type ReplacerList []Replacer

func (m ReplacerList) Apply(path string) string {
	for _, r := range m {
		path = r.Replace(path)
	}

	return path
}
