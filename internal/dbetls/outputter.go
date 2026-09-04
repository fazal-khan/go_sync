package dbetls

type Outputter interface {
	Output(records []map[string]any, cfg Output) error
}