package settings

type Setting struct {
	Key   string `db:"key"`
	Value string `db:"value"`
}
