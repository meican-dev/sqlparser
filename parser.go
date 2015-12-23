package sqlparser

import (
	"fmt"
	"io"
	"strconv"
)

// Column describe column detail information
type Column struct {
	Name     string
	Type     string
	Size     int
	Default  interface{}
	Comment  string
	Nullable bool
	AutoIncr bool
}

// Constraint holds foreign key constraint
type Constraint struct {
	Index      string
	ForeignKey string
	TableName  string
	ColumnName string
}

// Table is table schema
type Table struct {
	Name        string
	Columns     map[string]*Column
	PrimaryKey  string
	UniqueKeys  map[string]string
	Keys        map[string]string // index -> column_name
	Constraints map[string]*Constraint
	Extras      map[string]string
}

// Schema stores table name and its schema
type Schema map[string]*Table

// Parser stores parser state
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token
		lit string
		n   int
	}
}

// Type holds SQL datatype token and its literal representation
var Type map[Token]string

func init() {
	Type = make(map[Token]string)
	Type[BIT] = "bit"
	Type[TINYINT] = "tinyint"
	Type[SMALLINT] = "smallint"
	Type[INT] = "int"
	Type[BIGINT] = "bigint"
	Type[FLOAT] = "float"
	Type[DOUBLE] = "double"
	Type[VARCHAR] = "varchar"
	Type[LONGTEXT] = "longtext"
	Type[MEDIUMTEXT] = "mediumtext"
	Type[DATE] = "date"
	Type[TIME] = "time"
	Type[DATETIME] = "datetime"
	Type[TIMESTAMP] = "timestamp"
}

// NewParser returns a new parser for given reader
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

func (p *Parser) scan() (tok Token, lit string) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}
	tok, lit = p.s.Scan()
	p.buf.tok, p.buf.lit = tok, lit
	return
}

func (p *Parser) unscan() {
	p.buf.n = 1
}

func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS || tok == ANNOTATION {
		tok, lit = p.scan()
	}
	return
}

func (p *Parser) scanIdent() (tok Token, lit string) {
	tok, lit = p.scanIgnoreWhitespace()
	if tok != IDENT {
		return ILLEGAL, lit
	}
	return tok, lit
}

func (p *Parser) scanType() (string, int, error) {
	tok, lit := p.scanIgnoreWhitespace()
	if tok >= BIT && tok <= TIMESTAMP {
		tok1, lit1 := p.scanIgnoreWhitespace()
		if tok1 != OPEN_PAREN {
			p.unscan()
			return Type[tok], 0, nil
		}
		tok2, lit2 := p.scanIgnoreWhitespace()
		tok3, lit3 := p.scanIgnoreWhitespace()
		if tok2 != SIZE || tok3 != CLOSE_PAREN {
			return "", 0, fmt.Errorf("found %q, expected type(integer)", lit+lit1+lit2+lit3)
		}
		size, _ := strconv.Atoi(lit2)
		return Type[tok], size, nil
	}
	return "", 0, fmt.Errorf("found %q, expected type", lit)
}

func (p *Parser) scanDefault() (string, error) {
	tok, lit := p.scanIgnoreWhitespace()
	if tok != DEFAULT {
		return "", fmt.Errorf("found %q, expected DEFAULT", lit)
	}
	tok, lit = p.scanIgnoreWhitespace()
	switch tok {
	case NULL:
		return "null", nil
	case CURRENT_TIMESTAMP:
		return "current_timestamp", nil
	case STRING:
		return lit, nil
	}
	return "", fmt.Errorf("found %q, expected NULL or value", lit)
}

func (p *Parser) scanColumn() (*Column, error) {
	var column = &Column{}
	tok, lit := p.scanIdent()
	if tok != IDENT {
		return nil, fmt.Errorf("found %q, expected ident", lit)
	}
	column.Name = lit
	t, s, err := p.scanType()
	if err != nil {
		return nil, err
	}
	column.Type = t
	column.Size = s

	for {
		tok, lit = p.scanIgnoreWhitespace()
		switch tok {
		case DEFAULT:
			p.unscan()
			val, err := p.scanDefault()
			if err != nil {
				return nil, err
			}
			column.Default = val
			column.Nullable = val == "null"
		case NULL:
			column.Nullable = true
		case NOT:
			tok1, lit1 := p.scanIgnoreWhitespace()
			if tok1 != NULL {
				return nil, fmt.Errorf("found %q, expected NULL", lit1)
			}
			column.Nullable = false
		case COMMENT:
			if tok1, lit1 := p.scanIgnoreWhitespace(); tok1 == STRING {
				column.Comment = lit1
			} else {
				return nil, fmt.Errorf("found %q, expected 'comment'", lit1)
			}
		case AUTO_INCREMENT:
			column.AutoIncr = true
		case COMMA, CLOSE_PAREN:
			p.unscan()
			return column, nil
		case EOF:
			return nil, fmt.Errorf("unexpected EOF")
		default:
			return nil, fmt.Errorf("found %q, expected column constraint", lit)
		}
	}
}

func (p *Parser) scanPrimaryKey() (string, error) {
	tok1, lit1 := p.scanIgnoreWhitespace()
	tok2, lit2 := p.scanIgnoreWhitespace()
	if tok1 != PRIMARY || tok2 != KEY {
		return "", fmt.Errorf("found %q, expected PRIMARY KEY", lit1+lit2)
	}
	tok, lit := p.scanIgnoreWhitespace()
	if tok == OPEN_PAREN {
		p.unscan()
		tok, lit = p.scanParenIdent()
		if tok != IDENT {
			return "", fmt.Errorf("found %q, expected ident", lit)
		}
		return lit, nil
	}
	tok, lit = p.scanIdent()
	if tok != IDENT {
		return "", fmt.Errorf("found %q, expected ident", lit)
	}
	return lit, nil
}

func (p *Parser) scanParenIdent() (Token, string) {
	tok, lit := p.scanIgnoreWhitespace()
	if tok != OPEN_PAREN {
		return ILLEGAL, lit
	}
	tok, lit = p.scanIgnoreWhitespace()
	if tok == IDENT {
		tok1, lit1 := p.scanIgnoreWhitespace()
		if tok1 != CLOSE_PAREN {
			return ILLEGAL, lit + lit1
		}
		return tok, lit
	}
	return ILLEGAL, ""
}

func (p *Parser) scanKey() (string, string, error) {
	var index, column string
	tok, lit := p.scanIgnoreWhitespace()
	if tok != KEY {
		return "", "", fmt.Errorf("found %q, expected KEY", lit)
	}
	// parse index
	tok, lit = p.scanIgnoreWhitespace()
	if tok == IDENT {
		index = lit
	} else {
		return "", "", fmt.Errorf("found %q, expected index", lit)
	}
	// parse column
	tok, lit = p.scanIgnoreWhitespace()
	if tok == IDENT {
		column = lit
	} else if tok == OPEN_PAREN {
		p.unscan()
		tok, lit = p.scanParenIdent()
		if tok != IDENT {
			return "", "", fmt.Errorf("found %q, expected ", lit)
		}
		column = lit
	} else {
		return "", "", fmt.Errorf("found %q, expected ident", lit)
	}
	return index, column, nil
}

func (p *Parser) scanConstraint() (*Constraint, error) {
	var constraint = &Constraint{}
	tok, lit := p.scanIgnoreWhitespace()
	if tok != CONSTRAINT {
		return nil, fmt.Errorf("found %q, expected CONSTRAINT", lit)
	}
	tok, lit = p.scanIdent()
	if tok != IDENT {
		return nil, fmt.Errorf("found %q, expected ident", lit)
	}
	constraint.Index = lit
	tok1, lit1 := p.scanIgnoreWhitespace()
	tok2, lit2 := p.scanIgnoreWhitespace()
	if tok1 != FOREIGN || tok2 != KEY {
		return nil, fmt.Errorf("found %q, expected FOREIGN KEY", lit1+lit2)
	}
	tok, lit = p.scanParenIdent()
	if tok != IDENT {
		return nil, fmt.Errorf("found %q, expected ident", lit)
	}
	constraint.ForeignKey = lit
	tok, lit = p.scanIgnoreWhitespace()
	if tok != REFERENCES {
		return nil, fmt.Errorf("found %q, expected REFERENCES", lit)
	}
	tok, lit = p.scanIdent()
	if tok != IDENT {
		return nil, fmt.Errorf("found %q, expected `table_name`", lit)
	}
	constraint.TableName = lit
	tok, lit = p.scanParenIdent()
	if tok != IDENT {
		return nil, fmt.Errorf("found %q, expected (`column_name`)", lit)
	}
	constraint.ColumnName = lit
	return constraint, nil
}

func (p *Parser) scanKV() (string, string, error) {
	tok, lit := p.scanIgnoreWhitespace()
	tok1, lit1 := p.scanIgnoreWhitespace()
	tok2, lit2 := p.scanIgnoreWhitespace()
	if (tok != IDENT && tok != AUTO_INCREMENT) || tok1 != EQUAL || (tok2 != IDENT && tok2 != STRING && tok2 != SIZE) {
		return "", "", fmt.Errorf("found %q, expected key=value", lit+lit1+lit2)
	}
	return lit, lit2, nil
}

func (p *Parser) scanExtra() (map[string]string, error) {
	extras := make(map[string]string)
	for {
		if tok, _ := p.scanIgnoreWhitespace(); tok != SEMI_COLON {
			if tok != DEFAULT {
				p.unscan()
			}
			k, v, err := p.scanKV()
			if err != nil {
				return nil, err
			}
			extras[k] = v
		} else {
			p.unscan()
			break
		}
	}
	return extras, nil
}

// parse one table
func (p *Parser) parse() (*Table, error) {
	table := &Table{
		Columns:     make(map[string]*Column),
		UniqueKeys:  make(map[string]string),
		Keys:        make(map[string]string),
		Constraints: make(map[string]*Constraint),
		Extras:      make(map[string]string),
	}
	for {
		if tok, lit := p.scanIgnoreWhitespace(); tok == DROP || tok == LOCK || tok == UNLOCK {
			for { // ignore drop, lock and unlock statement
				if tok, _ := p.scanIgnoreWhitespace(); tok == SEMI_COLON {
					break
				} else if tok == EOF {
					return nil, nil
				}
			}
		} else if tok == SEMI_COLON || tok == ANNOTATION {
			continue
		} else if tok == CREATE {
			break
		} else if tok == EOF {
			return nil, nil
		} else {
			return nil, fmt.Errorf("unexpected %v: %q", tok, lit)
		}
	}
	if tok, lit := p.scanIgnoreWhitespace(); tok != TABLE {
		return nil, fmt.Errorf("found CREATE %q, expected CREATE TABLE", lit)
	}

	// scan table name
	if tok, lit := p.scanIdent(); tok == IDENT {
		table.Name = lit
	} else {
		return nil, fmt.Errorf("found CREATE TABLE %d %q, expected CREATE TABLE `ident`", tok, lit)
	}

	// scan columns
	if tok, lit := p.scanIgnoreWhitespace(); tok != OPEN_PAREN {
		return nil, fmt.Errorf("found %q, expected (", lit)
	}

	for {
		tok, lit := p.scanIgnoreWhitespace()
		switch tok {
		case IDENT:
			p.unscan()
			col, err := p.scanColumn()
			if err != nil {
				return nil, err
			}
			table.Columns[col.Name] = col
		case PRIMARY:
			p.unscan()
			key, err := p.scanPrimaryKey()
			if err != nil {
				return nil, err
			}
			table.PrimaryKey = key
		case UNIQUE:
			k, v, err := p.scanKey()
			if err != nil {
				return nil, err
			}
			table.UniqueKeys[k] = v
		case KEY:
			p.unscan()
			index, col, err := p.scanKey()
			if err != nil {
				return nil, err
			}
			table.Keys[index] = col
		case CONSTRAINT:
			p.unscan()
			cos, err := p.scanConstraint()
			if err != nil {
				return nil, err
			}
			table.Constraints[cos.ForeignKey] = cos
		case CLOSE_PAREN:
			tok, lit = p.scanIgnoreWhitespace()
			if tok != SEMI_COLON {
				p.unscan()
				extras, err := p.scanExtra()
				if err != nil {
					return nil, err
				}
				table.Extras = extras
			}
			return table, nil
		case COMMA:
			continue
		case SEMI_COLON:
			return table, nil
		default:
			return nil, fmt.Errorf("found %q, expected ident or primary or unique or key or constraint", lit)
		}
	}
}

// Parse returns parsed table schema and an error
func (p *Parser) Parse() (Schema, error) {
	schema := make(Schema)
	for {
		table, err := p.parse()
		if err != nil {
			return schema, err // return already parsed tables and error
		}
		if table == nil { // parse done
			break
		}
		schema[table.Name] = table
	}
	return schema, nil
}
