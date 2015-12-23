package sqlparser

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// Scanner wrapps a buffer reader
type Scanner struct {
	r *bufio.Reader
}

// Token represents a token
type Token int

const (
	// special token
	ILLEGAL Token = iota
	EOF
	ANNOTATION
	WS // space, tab and newline

	STRING
	IDENT // table_name, index, column_name, engine_name, charset_name

	COMMA
	BACKTICK
	SEMI_COLON
	OPEN_PAREN
	CLOSE_PAREN

	// data type
	SIZE // an integer indicate datatype size
	BIT
	TINYINT
	SMALLINT
	INT
	BIGINT
	FLOAT
	DOUBLE
	LONGTEXT
	MEDIUMTEXT
	VARCHAR
	DATE
	TIME
	DATETIME
	TIMESTAMP

	// SQL keywords
	DROP
	LOCK
	UNLOCK
	TABLES
	WRITE
	IF
	EXISTS
	EQUAL
	CREATE
	TABLE
	DEFAULT
	NOT
	NULL
	COMMENT
	KEY
	UNIQUE
	CONSTRAINT
	PRIMARY
	FOREIGN
	REFERENCES
	AUTO_INCREMENT
	CURRENT_TIMESTAMP
)

var (
	eof = rune(0)
)

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

func isString(ch rune) bool {
	return ch == '\''
}

// NewScanner returns a new scanner for the given reader
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
}

func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return WS, buf.String()
}

func (s *Scanner) scanDigit() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isDigit(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	return SIZE, buf.String()
}

func (s *Scanner) scanString() (tok Token, lit string) {
	var buf bytes.Buffer
	ch := s.read()
	readString := func(c rune) {
		for {
			if ch := s.read(); ch == c {
				break
			} else {
				_, _ = buf.WriteRune(ch)
			}
		}
	}
	switch ch {
	case '`':
		tok = IDENT
		readString('`')
	case '\'':
		tok = STRING
		readString('\'')
	default:
		return ILLEGAL, string(ch)
	}
	return tok, buf.String()
}

func (s *Scanner) scanInlineComment() (tok Token, lit string) {
	for {
		if ch := s.read(); ch == eof {
			return ILLEGAL, ""
		} else if ch == '*' {
			if c := s.read(); c == '/' {
				break
			}
		}
	}
	return ANNOTATION, ""
}

func (s *Scanner) scanIdent() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) && ch != '_' {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}
	switch strings.ToUpper(buf.String()) {
	case "DROP":
		return DROP, buf.String()
	case "IF":
		return IF, buf.String()
	case "EXISTS":
		return EXISTS, buf.String()
	case "LOCK":
		return LOCK, buf.String()
	case "UNLOCK":
		return UNLOCK, buf.String()
	case "TABLES":
		return TABLES, buf.String()
	case "WRITE":
		return WRITE, buf.String()
	case "CREATE":
		return CREATE, buf.String()
	case "TABLE":
		return TABLE, buf.String()
	case "NOT":
		return NOT, buf.String()
	case "NULL":
		return NULL, buf.String()
	case "DEFAULT":
		return DEFAULT, buf.String()
	case "COMMENT":
		return COMMENT, buf.String()
	case "KEY":
		return KEY, buf.String()
	case "UNIQUE":
		return UNIQUE, buf.String()
	case "CONSTRAINT":
		return CONSTRAINT, buf.String()
	case "PRIMARY":
		return PRIMARY, buf.String()
	case "FOREIGN":
		return FOREIGN, buf.String()
	case "REFERENCES":
		return REFERENCES, buf.String()
	case "AUTO_INCREMENT":
		return AUTO_INCREMENT, buf.String()
	case "CURRENT_TIMESTAMP":
		return CURRENT_TIMESTAMP, buf.String()
	case "BIT":
		return BIT, buf.String()
	case "TINYINT":
		return TINYINT, buf.String()
	case "SMALLINT":
		return SMALLINT, buf.String()
	case "INT":
		return INT, buf.String()
	case "BIGINT":
		return BIGINT, buf.String()
	case "FLOAT":
		return FLOAT, buf.String()
	case "DOUBLE":
		return DOUBLE, buf.String()
	case "VARCHAR":
		return VARCHAR, buf.String()
	case "LONGTEXT":
		return LONGTEXT, buf.String()
	case "MEDIUMTEXT":
		return MEDIUMTEXT, buf.String()
	case "DATE":
		return DATE, buf.String()
	case "TIME":
		return TIME, buf.String()
	case "DATETIME":
		return DATETIME, buf.String()
	case "TIMESTAMP":
		return TIMESTAMP, buf.String()
	default:
		return IDENT, buf.String()
	}
}

// Scan method scans one token, returns a token and its literal string
func (s *Scanner) Scan() (tok Token, lit string) {
	ch := s.read()

	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		s.unread()
		return s.scanIdent()
	} else if isDigit(ch) {
		s.unread()
		return s.scanDigit()
	} else if ch == '\'' || ch == '`' {
		s.unread()
		return s.scanString()
	} else if ch == '/' {
		if c := s.read(); c == '*' {
			s.unread()
			s.unread()
			return s.scanInlineComment()
		}
		s.unread()
		return ILLEGAL, string(ch)
	}

	switch ch {
	case eof:
		return EOF, "EOF"
	case ',':
		return COMMA, ","
	case '(':
		return OPEN_PAREN, "("
	case ')':
		return CLOSE_PAREN, ")"
	case ';':
		return SEMI_COLON, ";"
	case '=':
		return EQUAL, "="
	case '-':
		if c := s.read(); c == '-' { // comment
			for {
				if c := s.read(); c == '\n' {
					return ANNOTATION, ""
				}
			}
		}
		return ILLEGAL, string(ch)
	default:
		return ILLEGAL, string(ch)
	}
}
