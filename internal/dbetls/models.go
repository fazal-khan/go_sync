package dbetls

type Config struct {
	Table []Table `xml:"table"`
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
	Name           string `xml:"name,attr"`
}

type Filter struct {
	Mutate Mutate `xml:"mutate"`
}

type Mutate struct {
	CopyValue       []CopyValue     `xml:"copy-value"`
	RemoveFields    RemoveFields    `xml:"remove-fields"`
	AddFields       AddFields       `xml:"add-fields"`
	LowercaseFields LowercaseFields `xml:"lowercase-fields"`
	Row             string          `xml:"row,attr"`
}

type AddFields struct {
	Field []Field `xml:"field"`
}

type Field struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

type CopyValue struct {
	From string `xml:"from,attr"`
	To   string `xml:"to,attr"`
	In   string `xml:"in,attr"`
}

type LowercaseFields struct {
	Field         []string `xml:"field"`
	For           string   `xml:"for,attr"`
	CaseSentitive string   `xml:"case-sentitive,attr"`
}

type RemoveFields struct {
	Field      []string `xml:"field"`
	IgnoreCase string   `xml:"ignore-case,attr"`
}

type Output struct {
	Target     string `xml:"target"`
	URL        string `xml:"url"`
	Auth       Auth   `xml:"auth"`
	TLS        TLS    `xml:"tls"`
	SkipOutput string `xml:"skip-output"`
	Result     Result `xml:"result"`
	Name       string `xml:"name,attr"`
	Type       string `xml:"type,attr"`
}

type Auth struct {
	User     string `xml:"user,attr"`
	Password string `xml:"password,attr"`
}

type Result struct {
	Bulk  string `xml:"bulk,attr"`
	Cdata string `xml:",chardata"`
}

type TLS struct {
	Verify  string `xml:"verify,attr"`
	Enabled string `xml:"enabled,attr"`
}

type Query struct {
	Cdata string `xml:",chardata"`
}
