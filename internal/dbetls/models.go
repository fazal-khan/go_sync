package dbetls

type Config struct {
	Table Table `xml:"table"`
}

type Table struct {
	DatabaseName   string `xml:"database_name"`
	IDColumn       string `xml:"id-column"`
	TrackingColumn string `xml:"tracking-column"`
	BatchSize      string `xml:"batch-size"`
	UseSkip        string `xml:"use-skip"`
	MaxRecords     string `xml:"max-records"`
	WaitMS         string `xml:"wait-ms"`
	Cron           string `xml:"cron"`
	Query          Query  `xml:"query"`
	Filter         Filter `xml:"filter"`
	Output         Output `xml:"output"`
	Name           string `xml:"_name"`
}

type Filter struct {
	Mutate Mutate `xml:"mutate"`
}

type Mutate struct {
	CopyValue       []CopyValue     `xml:"copy-value"`
	RemoveFields    RemoveFields    `xml:"remove-fields"`
	AddFields       AddFields       `xml:"add-fields"`
	LowercaseFields LowercaseFields `xml:"lowercase-fields"`
	Row             string          `xml:"_row"`
}

type AddFields struct {
	Field []Field `xml:"field"`
}

type Field struct {
	Key   string `xml:"_key"`
	Value string `xml:"_value"`
}

type CopyValue struct {
	From string  `xml:"_from"`
	To   string  `xml:"_to"`
	In   *string `xml:"_in,omitempty"`
}

type LowercaseFields struct {
	Field         []string `xml:"field"`
	For           string   `xml:"_for"`
	CaseSentitive string   `xml:"_case-sentitive"`
}

type RemoveFields struct {
	Field      []string `xml:"field"`
	IgnoreCase string   `xml:"_ignore-case"`
}

type Output struct {
	Target     string `xml:"target"`
	URL        string `xml:"url"`
	Auth       Auth   `xml:"auth"`
	TLS        TLS    `xml:"tls"`
	SkipOutput string `xml:"skip-output"`
	Result     Result `xml:"result"`
	Name       string `xml:"_name"`
	Type       string `xml:"_type"`
}

type Auth struct {
	User     string `xml:"_user"`
	Password string `xml:"_password"`
}

type Result struct {
	Bulk  string `xml:"_bulk"`
	Cdata string `xml:"__cdata"`
}

type TLS struct {
	Verify  string `xml:"_verify"`
	Enabled string `xml:"_enabled"`
}

type Query struct {
	Cdata string `xml:"__cdata"`
}
