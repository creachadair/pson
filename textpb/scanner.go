package textpb

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Token int

const (
	None     Token = iota
	Name           // field name or enumerator
	True           // the "true" constant
	False          // the "false" constant
	TypeName       // extension or Any type name
	Colon          // name:value separator
	String         // quoted string
	Number         // int or float
	LeftA          // left angle bracket
	RightA         // right angle bracket
	LeftC          // left curly bracker
	RightC         // right curly bracket
	Comma          // comma
	Semi           // semicolon

	// These are whitespace characters.
	whiteSpace = " \t\r\n"

	// These are delimiters for a name-like token.
	nameDelim = whiteSpace + `<>{}:'",;`
)

func (t Token) String() string { return tokenString[t] }

// IsValue reports whether t is a value token.
func (t Token) IsValue() bool {
	return t == Name || t == True || t == False || t == TypeName || t == String || t == Number
}

var selfToken = map[rune]Token{
	':': Colon,
	'<': LeftA,
	'>': RightA,
	'{': LeftC,
	'}': RightC,
	',': Comma,
	';': Semi,
}

var escapeCode = map[rune]rune{'r': '\r', 'n': '\n', 't': '\t', '\\': '\\'}

var tokenString = map[Token]string{
	None:     "<none>",
	Name:     "NAME",
	True:     "true",
	False:    "false",
	TypeName: "TYPE",
	Colon:    `":"`,
	String:   "STRING",
	Number:   "NUMBER",
	LeftA:    `"<"`,
	RightA:   `">"`,
	LeftC:    `"{"`,
	RightC:   `"}"`,
	Comma:    `","`,
	Semi:     `";"`,
}

var isNumber = regexp.MustCompile(`^-?(\d+(\.\d*)?|\.\d+)([eE][-+]?\d+)?[fF]?$`)
var isName = regexp.MustCompile(`(?i)^[_a-z][_a-z0-9]*$`)

func isSpace(c rune) bool { return strings.IndexRune(whiteSpace, c) >= 0 }
func isDelim(c rune) bool { return strings.IndexRune(nameDelim, c) >= 0 }

// NewScanner returns a scanner that consumes data from r.
func NewScanner(r io.Reader) *Scanner { return &Scanner{r: bufio.NewReader(r)} }

// A Scanner returns tokens from a text-format protobuf message.
type Scanner struct {
	r        *bufio.Reader
	tok      Token // current token type
	pos, end int   // byte offset in input
	lnum     int   // line number (0-based)
	err      error // error from previous operation
	cur      bytes.Buffer
}

// Token returns the type of the current token.
func (s *Scanner) Token() Token { return s.tok }

// Err returns the error from the most recent call to Next.
func (s *Scanner) Err() error { return s.err }

// Pos returns the byte offset of the start of the current token.
func (s *Scanner) Pos() int { return s.pos }

// End returns the byte offset just past the end of the current token.
func (s *Scanner) End() int { return s.end }

// Line returns the line number of the start of the current token (1-based).
func (s *Scanner) Line() int { return s.lnum + 1 }

// Text returns the text of the current token.
func (s *Scanner) Text() string { return s.cur.String() }

func (s *Scanner) ok(tok Token) bool   { s.tok = tok; return true }
func (s *Scanner) fail(err error) bool { s.err = err; return false }

// Next advances the scanner to the next token and reports whether any token
// was found. The Err method reports whether there was an error. When the input
// is exhausted, Err returns io.EOF.
func (s *Scanner) Next() bool {
	if s.err != nil {
		return false
	}
	s.tok = None
	s.pos = s.end
	s.cur.Reset()

	c, err := s.skipSpace()
	if err != nil {
		return s.fail(err)
	}
	if c == '"' || c == '\'' {
		return s.quotedString(c)
	} else if c == '[' {
		return s.typeName()
	}
	s.cur.WriteRune(c)

	// Check for self-delimiting tokens.
	if t, ok := selfToken[c]; ok {
		return s.ok(t)
	} else {
		return s.nameLike(c)
	}
}

// nameLike scans names, numbers, and Boolean constants.
func (s *Scanner) nameLike(init rune) bool {
	for {
		c, n, err := s.r.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return s.fail(err)
		}

		if isDelim(c) {
			s.r.UnreadRune()
			break
		}
		s.cur.WriteRune(c)
		s.end += n
	}
	cur := s.cur.String()
	if cur == "true" {
		return s.ok(True)
	} else if cur == "false" {
		return s.ok(False)
	} else if isNumber.MatchString(cur) {
		return s.ok(Number)
	} else if isName.MatchString(cur) {
		return s.ok(Name)
	}
	return s.fail(fmt.Errorf("invalid token %q", cur))
}

// typeName scans a string bounded by square brackets, assuming the leading
// bracket has already been read. On success the token text excludes the
// brackets.
func (s *Scanner) typeName() bool {
	for {
		c, n, err := s.r.ReadRune()
		if err != nil {
			return s.fail(err)
		}
		s.end += n
		if c == ']' {
			return s.ok(TypeName)
		} else if isDelim(c) {
			return s.fail(fmt.Errorf("unexpected %q in type name", c))
		}
		s.cur.WriteRune(c)
	}
}

// quotedString scans a string bounded by quote, assuming the leading quote has
// already been read. On success the token text excludes the quotes and escape
// sequences have been folded out.
func (s *Scanner) quotedString(quote rune) bool {
	esc := false
	for {
		c, n, err := s.r.ReadRune()
		if err == io.EOF {
			return s.fail(fmt.Errorf("missing %q in string", quote))
		} else if err != nil {
			return s.fail(err)
		}
		s.end += n
		if c == '\r' || c == '\n' {
			return s.fail(fmt.Errorf("unexpected %q in string", c))
		} else if esc {
			esc = false
			if sub, ok := escapeCode[c]; ok {
				s.cur.WriteRune(sub)
				continue
			} else if c != quote {
				s.cur.WriteRune('\\')
			}
		} else if c == '\\' {
			esc = true
			continue
		} else if c == quote {
			return s.ok(String)
		}
		s.cur.WriteRune(c)
	}
}

// skipSpace discards whitespace and returns the first non-space byte.
func (s *Scanner) skipSpace() (rune, error) {
	for {
		c, n, err := s.r.ReadRune()
		if err != nil {
			return 0, err
		}
		s.end += n
		if c == '\n' {
			s.lnum++
		}
		if c == '#' {
			for c != '\n' {
				c, n, err = s.r.ReadRune()
				if err != nil {
					return 0, err
				}
				s.end += n
			}
			s.lnum++
		} else if !isSpace(c) {
			s.pos = s.end
			return c, nil
		}
	}
}
