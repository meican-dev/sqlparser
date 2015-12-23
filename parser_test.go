package sqlparser

import (
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	sqlStmt := "--this is a comment\nDROP TABLE IF EXISTS `user`;\n/* this is another comment */;\nCREATE TABLE `user` (\n  `id` bigint(20) NOT NULL AUTO_INCREMENT,\n  `username` varchar(20) DEFAULT NULL\n) ENGINE=InnoDB DEFAULT CHARSET=utf8;"
	expected := make(Schema)
	columns := make(map[string]*Column)
	columns["id"] = &Column{
		Name:    "id",
		Type:    "bigint",
		Size:    20,
		Default: "null",
	}
	columns["username"] = &Column{
		Name:    "username",
		Type:    "varchar",
		Size:    20,
		Default: "null",
	}
	extras := make(map[string]string)
	extras["engine"] = "InnoDB"
	extras["charset"] = "utf8"
	expected["user"] = &Table{
		Name:    "user",
		Columns: columns,
		Extras:  extras,
	}
	p := NewParser(strings.NewReader(sqlStmt))
	schema, err := p.Parse()
	if err != nil {
		t.Error(err)
	}
	if len(schema) != 1 {
		t.Errorf("expected one table, found %d", len(schema))
	}
	user := schema["user"]
	if user == nil {
		t.Errorf("expected table user, but not found")
	}
	if len(user.Columns) != 2 {
		t.Errorf("expected 2 columns, found %d", len(user.Columns))
	}

	// parsing sql file
	// f, err := os.Open("table_schema.sql")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// p := NewParser(f)
	// schema, err := p.Parse()
	// if err != nil {
	// 	t.Errorf(err.Error())
	// }
	// fmt.Printf("%v\n", schema)
}
