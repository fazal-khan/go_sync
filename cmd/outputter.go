package main

import (
	"github.com/fazal-khan/go_sync/internal/dbetls"
)

type noOpOutputter struct{}

func (n *noOpOutputter) Output(records []map[string]any, cfg dbetls.Output) error {
	return nil
}

func newOutputter() dbetls.Outputter {
	return &noOpOutputter{}
}
