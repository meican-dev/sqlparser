# sqlparser

A Golang library to parse table schema dumped from MySQL database.

## Usage:
1. Dump table structure from mysql database

  ```shell
  mysqldump -d -hlocalhost -uroot -p mydb > table_schema.sql
  ```

2. Parse table schema in go program

  ```go
  package main
  
  import (
  	"fmt"
  	"log"
  	"os"
  
  	"github.com/meican-dev/sqlparser"
  )
  
  func main() {
  	schemaFile := "table_schema.sql"
  	file, err := os.Open(schemaFile)
  	if err != nil {
  		log.Fatalln(err)
  	}
  	parser := sqlparser.NewParser(file)
  	schema, err := parser.Parse()
  	if err != nil {
  		log.Fatalln(err)
  	}
  	fmt.Printf("table schema: %v\n", schema)
  }
  ```


*Warning*: This library is primarily used to parse dumped table schema
 which is well-formed SQL statement. It's not robust to parse any SQL
 statement
