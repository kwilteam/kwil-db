package convert

type convert struct {
	Clean clean
}

var Convert = convert{
	Clean: clean{},
}

func (c *convert) CleanDB()
