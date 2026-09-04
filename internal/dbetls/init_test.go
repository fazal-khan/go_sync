package dbetls

import (
	"encoding/xml"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/fazal-khan/go_sync/internal/core"
)

const validXML = `<?xml version="1.0" encoding="UTF-8"?>
<config>
    <table name="test_table">
        <database_name>mydb</database_name>
        <id-column>id</id-column>
        <tracking-column>updated_at</tracking-column>
        <batch-size>100</batch-size>
        <use-skip>false</use-skip>
        <max-records>500</max-records>
        <wait-ms>200</wait-ms>
        <cron>*/5 * * * *</cron>
        <query><![CDATA[SELECT id FROM mydb WHERE updated_at > :last_ts]]></query>
        <filter>
            <mutate row="_row">
                <copy-value from="id" to="_id" in="root" />
                <remove-fields ignore-case="true">
                    <field>id</field>
                </remove-fields>
                <add-fields>
                    <field key="type" value="A" />
                </add-fields>
                <lowercase-fields for="value" case-sentitive="false">
                    <field>name</field>
                </lowercase-fields>
            </mutate>
        </filter>
        <output name="mytable" type="api">
            <target>elasticsearch</target>
            <url>http://example.com/api/v1/mytable</url>
            <auth user="basic_user" password="basic_pass" />
            <tls verify="false" enabled="false" />
            <skip-output>false</skip-output>
            <result bulk="true"><![CDATA[{"index":{"_index":"mytable"}}]]></result>
        </output>
    </table>
</config>`

const xmlWithEnvVars = `<?xml version="1.0" encoding="UTF-8"?>
<config>
    <table name="env_table">
        <database_name>envdb</database_name>
        <id-column>id</id-column>
        <tracking-column>updated_at</tracking-column>
        <batch-size>50</batch-size>
        <use-skip>false</use-skip>
        <max-records>100</max-records>
        <wait-ms>100</wait-ms>
        <cron>*/1 * * * *</cron>
        <query><![CDATA[SELECT * FROM envdb]]></query>
        <filter>
            <mutate row="_row" />
        </filter>
        <output name="env_out" type="api">
            <target>elasticsearch</target>
            <url>http://${env.API_HOST}:9200/${env.INDEX_NAME}</url>
            <auth user="${env.ES_USER}" password="${env.ES_PASS}" />
            <tls verify="true" enabled="true" />
            <skip-output>false</skip-output>
            <result bulk="true"><![CDATA[{"index":{"_index":"${env.INDEX_NAME}","_id":"${row.id}"}}]]></result>
        </output>
    </table>
</config>`

const invalidXML = `<?xml version="1.0" encoding="UTF-8"?>
<config>
    <table name="bad">
        <unclosed_tag>
    </table>
</config>`

func TestDecodeValidXML(t *testing.T) {
	var config Config
	err := xml.Unmarshal([]byte(validXML), &config)
	if err != nil {
		t.Fatalf("failed to decode valid XML: %v", err)
	}

	if len(config.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(config.Tables))
	}

	tbl := config.Tables[0]
	if tbl.DatabaseName != "mydb" {
		t.Errorf("DatabaseName = %q, want %q", tbl.DatabaseName, "mydb")
	}
	if tbl.IDColumn != "id" {
		t.Errorf("IDColumn = %q, want %q", tbl.IDColumn, "id")
	}
	if tbl.TrackingColumn != "updated_at" {
		t.Errorf("TrackingColumn = %q, want %q", tbl.TrackingColumn, "updated_at")
	}
	if tbl.BatchSize != "100" {
		t.Errorf("BatchSize = %q, want %q", tbl.BatchSize, "100")
	}
	if tbl.Cron != "*/5 * * * *" {
		t.Errorf("Cron = %q, want %q", tbl.Cron, "*/5 * * * *")
	}
	if tbl.Query.Cdata == "" {
		t.Error("Query.Cdata should not be empty")
	}
	if tbl.Output.URL != "http://example.com/api/v1/mytable" {
		t.Errorf("Output.URL = %q, want %q", tbl.Output.URL, "http://example.com/api/v1/mytable")
	}
	if tbl.Output.Auth.User != "basic_user" {
		t.Errorf("Output.Auth.User = %q, want %q", tbl.Output.Auth.User, "basic_user")
	}
	if tbl.Output.Auth.Password != "basic_pass" {
		t.Errorf("Output.Auth.Password = %q, want %q", tbl.Output.Auth.Password, "basic_pass")
	}
	if tbl.Output.Target != "elasticsearch" {
		t.Errorf("Output.Target = %q, want %q", tbl.Output.Target, "elasticsearch")
	}
	if tbl.Output.Type != "api" {
		t.Errorf("Output.Type = %q, want %q", tbl.Output.Type, "api")
	}
	if tbl.Output.Name != "mytable" {
		t.Errorf("Output.Name = %q, want %q", tbl.Output.Name, "mytable")
	}
}

func TestDecodeFilterFields(t *testing.T) {
	var config Config
	err := xml.Unmarshal([]byte(validXML), &config)
	if err != nil {
		t.Fatalf("failed to decode XML: %v", err)
	}

	mutate := config.Tables[0].Filter.Mutate
	if mutate.Row != "_row" {
		t.Errorf("Mutate.Row = %q, want %q", mutate.Row, "_row")
	}
	if len(mutate.CopyValue) != 1 {
		t.Fatalf("expected 1 CopyValue, got %d", len(mutate.CopyValue))
	}
	if mutate.CopyValue[0].From != "id" || mutate.CopyValue[0].To != "_id" {
		t.Errorf("CopyValue[0] = from=%q to=%q, want from=id to=_id",
			mutate.CopyValue[0].From, mutate.CopyValue[0].To)
	}
	if mutate.CopyValue[0].In != "root" {
		t.Errorf("CopyValue[0].In = %q, want %q", mutate.CopyValue[0].In, "root")
	}
	if len(mutate.RemoveFields.Field) != 1 || mutate.RemoveFields.Field[0] != "id" {
		t.Errorf("RemoveFields = %v, want [id]", mutate.RemoveFields.Field)
	}
	if mutate.RemoveFields.IgnoreCase != "true" {
		t.Errorf("RemoveFields.IgnoreCase = %q, want %q", mutate.RemoveFields.IgnoreCase, "true")
	}
	if len(mutate.AddFields.Field) != 1 || mutate.AddFields.Field[0].Key != "type" {
		t.Errorf("AddFields = %v, want [{type A}]", mutate.AddFields.Field)
	}
	if len(mutate.LowercaseFields.Field) != 1 || mutate.LowercaseFields.Field[0] != "name" {
		t.Errorf("LowercaseFields = %v, want [name]", mutate.LowercaseFields.Field)
	}
}

func TestDecodeInvalidXML(t *testing.T) {
	var config Config
	err := xml.Unmarshal([]byte(invalidXML), &config)
	if err == nil {
		t.Error("expected error for invalid XML, got nil")
	}
}

func TestEnvSubstitution(t *testing.T) {
	t.Setenv("API_HOST", "search.example.com")
	t.Setenv("INDEX_NAME", "my_index")
	t.Setenv("ES_USER", "elastic")
	t.Setenv("ES_PASS", "secret123")

	var config Config
	err := xml.Unmarshal([]byte(xmlWithEnvVars), &config)
	if err != nil {
		t.Fatalf("failed to decode XML: %v", err)
	}

	// Apply env substitution
	substituteConfigEnvVars(&config)

	tbl := config.Tables[0]

	// URL should have env vars replaced but row.id untouched
	wantURL := "http://search.example.com:9200/my_index"
	if tbl.Output.URL != wantURL {
		t.Errorf("Output.URL = %q, want %q", tbl.Output.URL, wantURL)
	}
	if tbl.Output.Auth.User != "elastic" {
		t.Errorf("Output.Auth.User = %q, want %q", tbl.Output.Auth.User, "elastic")
	}
	if tbl.Output.Auth.Password != "secret123" {
		t.Errorf("Output.Auth.Password = %q, want %q", tbl.Output.Auth.Password, "secret123")
	}

	// Result.Cdata should have ${env.INDEX_NAME} replaced but ${row.id} untouched
	wantResult := `{"index":{"_index":"my_index","_id":"${row.id}"}}`
	if tbl.Output.Result.Cdata != wantResult {
		t.Errorf("Output.Result.Cdata = %q, want %q", tbl.Output.Result.Cdata, wantResult)
	}

	// Query.Cdata should be untouched (no env vars in it)
	wantQuery := "SELECT * FROM envdb"
	if tbl.Query.Cdata != wantQuery {
		t.Errorf("Query.Cdata = %q, want %q", tbl.Query.Cdata, wantQuery)
	}
}

func TestInitFromFile(t *testing.T) {
	// Write a test XML file to a temp directory
	tmpDir := t.TempDir()
	xmlFile := filepath.Join(tmpDir, "test_config.xml")
	err := os.WriteFile(xmlFile, []byte(validXML), 0644)
	if err != nil {
		t.Fatalf("failed to write temp XML file: %v", err)
	}

	ctx := &core.AppCtx{Logger: noopLogger()}

	config, err := Init(ctx, xmlFile)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if len(config.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(config.Tables))
	}
	if config.Tables[0].DatabaseName != "mydb" {
		t.Errorf("DatabaseName = %q, want %q", config.Tables[0].DatabaseName, "mydb")
	}
}

func TestInitFromFileNotFound(t *testing.T) {
	ctx := &core.AppCtx{Logger: noopLogger()}

	_, err := Init(ctx, "/nonexistent/path/config.xml")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestInitFromFileInvalidXML(t *testing.T) {
	tmpDir := t.TempDir()
	xmlFile := filepath.Join(tmpDir, "bad_config.xml")
	err := os.WriteFile(xmlFile, []byte(invalidXML), 0644)
	if err != nil {
		t.Fatalf("failed to write temp XML file: %v", err)
	}

	ctx := &core.AppCtx{Logger: noopLogger()}

	_, err = Init(ctx, xmlFile)
	if err == nil {
		t.Error("expected error for invalid XML, got nil")
	}
}

func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
}
