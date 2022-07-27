package scanner

import (
	"github.com/cyevgeniy/pldoc/token"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"
)

const eof = -1

type Scanner struct {
	file *token.File
	src  []byte

	ch         rune
	offset     int
	rdOffset   int
	lineOffset int
}

func (s *Scanner) Init(file *token.File, src []byte) {
	s.file = file
	s.src = src
	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
}

func (s *Scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.file.AddLine(s.offset)
		}
		r, w := utf8.DecodeRune(s.src[s.rdOffset:])

		if r == utf8.RuneError && w == 1 {
			log.Fatal("illegal UTF-8 encoding")
		}

		s.rdOffset += w
		s.ch = r

	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.file.AddLine(s.offset)
		}
		s.ch = eof
	}
}

func stripCR(b []byte, comment bool) []byte {
	c := make([]byte, len(b))
	i := 0
	for j, ch := range b {
		// In a /*-style comment, don't strip \r from *\r/ (incl.
		// sequences of \r from *\r\r...\r/) since the resulting
		// */ would terminate the comment too early unless the \r
		// is immediately following the opening /* in which case
		// it's ok because /*/ is not closed yet (issue #11151).
		if ch != '\r' || comment && i > len("/*") && c[i-1] == '*' && j+1 < len(b) && b[j+1] == '/' {
			c[i] = ch
			i++
		}
	}
	return c[:i]
}

func (s *Scanner) scanComment() string {
	// initial '/' or '-' already consumed; s.ch == '*' || s.ch == '-'
	offs := s.offset - 1 // position of initial '/'
	next := -1           // position immediately following the comment; < 0 means invalid comment
	numCR := 0

	if s.ch == '-' {
		// -- -style comment
		// (the final '\n' is not considered part of the comment)
		s.next()
		for s.ch != '\n' && s.ch >= 0 {
			if s.ch == '\r' {
				numCR++
			}
			s.next()
		}
		// if we are at '\n', the position following the comment is afterwards
		next = s.offset
		if s.ch == '\n' {
			next++
		}
		goto exit
	}

	/*-style comment */
	s.next()
	for s.ch >= 0 {
		ch := s.ch
		if ch == '\r' {
			numCR++
		}
		s.next()
		if ch == '*' && s.ch == '/' {
			s.next()
			next = s.offset
			goto exit
		}
	}

	log.Fatal("comment not terminated")

exit:
	lit := s.src[offs:s.offset]

	// On Windows, a (//-comment) line may end in "\r\n".
	// Remove the final '\r' before analyzing the text for
	// line directives (matching the compiler). Remove any
	// other '\r' afterwards (matching the pre-existing be-
	// havior of the scanner).
	if numCR > 0 && len(lit) >= 2 && lit[1] == '/' && lit[len(lit)-1] == '\r' {
		lit = lit[:len(lit)-1]
		numCR--
	}

	if numCR > 0 {
		lit = stripCR(lit, lit[1] == '*')
	}

	return string(lit)
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\r' || s.ch == '\n' {
		s.next()
	}
}

func lower(ch rune) rune { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter

func isLetter(ch rune) bool {
	return 'a' <= lower(ch) && lower(ch) <= 'z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }

func isDigit(ch rune) bool {
	return isDecimal(ch) || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

func (s *Scanner) scanIdentifier() string {
	offs := s.offset

	for isDigit(s.ch) || isLetter(s.ch) {
		s.next()
	}

	return string(s.src[offs:s.offset])

}

func (s *Scanner) scanString() string {
	offs := s.offset
	// 'Hello, World'
	for s.ch != '\'' {
		s.next()
	}

	// We don't want to include closing ' symbol
	endOffs := s.offset

	// Skip last ' symbol
	s.next()

	return string(s.src[offs:endOffs])
}

func (s *Scanner) scanNumber() string {
	offs := s.offset

	for isDigit(s.ch) {
		s.next()
	}

	return string(s.src[offs:s.offset])
}

func (s *Scanner) peek() byte {
	if s.rdOffset < len(s.src) {
		return s.src[s.rdOffset]
	}

	return 0
}

func (s *Scanner) Scan() (pos token.Pos, tok token.Token, lit string) {
	s.skipWhitespace()

	pos = s.file.Pos(s.offset)

	switch ch := s.ch; {
	case isLetter(ch):
		lit = strings.ToLower(s.scanIdentifier())
		tok = token.Lookup(lit)
	case isDigit(ch):
		lit = s.scanNumber()
		tok = token.NUMBER
	default:
		s.next()
		switch ch {
		case eof:
			tok = token.EOF
		case '\'':
			tok = token.STRING
			lit = s.scanString()
		case '.':
			tok = token.DOT
			lit = "."
		case '/':
			if s.ch == '*' {
				tok = token.COMMENT
				lit = s.scanComment()
			} else {
				tok = token.DIV
				lit = "/"
			}
		case '+':
			tok = token.ADD
			lit = "+"
		case '-':
			if s.ch == '-' {
				tok = token.COMMENT
				lit = s.scanComment()
			} else {
				tok = token.SUB
				lit = "-"
			}
		case '*':
			if s.ch == '*' {
				tok = token.EXP
				lit = "**"
				s.next()
			} else {
				tok = token.MUL
				lit = "*"
			}
		case '%':
			tok = token.REM
			lit = "%"
		case '=':
			tok = token.EQL
			lit = "="
		case '|':
			if s.ch == '|' {
				tok = token.CON
				lit = "||"
				s.next()
			} else {
				log.Fatal("Expecting another one |")
			}
		case '(':
			tok = token.LPAREN
			lit = "("
		case ')':
			tok = token.RPAREN
			lit = ")"
		case '[':
			tok = token.LBRACK
			lit = "["
		case ']':
			tok = token.RBRACK
			lit = "]"
		case ',':
			tok = token.COMMA
			lit = ","
		case ';':
			tok = token.SEMICOLON
			lit = ";"
		case '>':
			if s.ch == '=' {
				tok = token.GEQ
				lit = ">="
				s.next()
			} else {
				tok = token.GRT
				lit = ">"
			}
		case '<':
			if s.ch == '>' {
				tok = token.NEQ
				lit = "<>"
				s.next()
			} else if s.ch == '=' {
				tok = token.LEQ
				lit = "<="
				s.next()
			} else {
				tok = token.LSS
				lit = "<"
			}
		case ':':
			if s.ch == '=' {
				tok = token.ASSIGN
				lit = ":="
				s.next()
			} else {
				tok = token.COLON
				lit = ":"
			}
		case '"':
			tok = token.DQUOTE
			lit = "\""
		case '$':
			tok = token.DOLLAR
			lit = "$"
		}
	}

	return
}
