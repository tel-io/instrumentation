package http

type CardinalityGrouper func(path string) string

type CardinalityGrouperList []CardinalityGrouper

func (m CardinalityGrouperList) Apply(path string) string {
	for _, grp := range m {
		path = grp(path)
	}

	return path
}

func WithCardinalityGroupers(list []CardinalityGrouper) Option {
	return optionFunc(func(c *config) {
		c.groupers = list
	})
}
