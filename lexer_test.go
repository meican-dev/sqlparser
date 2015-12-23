package sqlparser

import (
	"fmt"
	"strings"
	"testing"
)

func Test_Lexer(t *testing.T) {
	sqlStmt := "--this is a comment\nDROP TABLE IF EXISTS `user`;\n/* this is an inline comment */\nCREATE TABLE `user` (\n  `id` bigint(20) NOT NULL AUTO_INCREMENT,\n  `username` varchar(20) DEFAULT NULL\n) ENGINE=InnoDB DEFAULT CHARSET=utf8;"
	fmt.Printf("%q\n", sqlStmt)
	s := NewScanner(strings.NewReader(sqlStmt))
	expectedTokens := []Token{
		ANNOTATION,
		DROP, WS, TABLE, WS, IF, WS, EXISTS, WS, IDENT, SEMI_COLON, WS,
		ANNOTATION, WS,
		CREATE, WS, TABLE, WS, IDENT, WS, OPEN_PAREN, WS,
		IDENT, WS, BIGINT, OPEN_PAREN, SIZE, CLOSE_PAREN, WS, NOT, WS, NULL, WS, AUTO_INCREMENT, COMMA, WS,
		IDENT, WS, VARCHAR, OPEN_PAREN, SIZE, CLOSE_PAREN, WS, DEFAULT, WS, NULL, WS,
		CLOSE_PAREN, WS, IDENT, EQUAL, IDENT, WS, DEFAULT, WS, IDENT, EQUAL, IDENT, SEMI_COLON,
	}
	var tokens []Token
	for {
		if tok, lit := s.Scan(); tok != EOF {
			fmt.Printf("%v: %v\n", tok, lit)
			tokens = append(tokens, tok)
		} else {
			break
		}
	}
	if len(tokens) != len(expectedTokens) {
		t.Errorf("tokens length mismatch, expected %d, found %d\n", len(expectedTokens), len(tokens))
	}
	for i := 0; i < len(tokens); i++ {
		if tokens[i] != expectedTokens[i] {
			t.Errorf("expected: %v found: %v", expectedTokens[i], tokens[i])
		}
	}

	// scan sql file
	// f, err := os.Open("table_schema.sql")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// s := NewScanner(f)
}
