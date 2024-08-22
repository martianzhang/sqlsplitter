package sqlsplitter

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// SQLSplitter sql splitter config struct
type SQLSplitter struct {
	// config value
	FileName string
	Reader   *bufio.Reader
	Dialect  Dialect
	Escape   rune

	// dynamic value
	Error    error
	LastLine uint64

	delimiter string
	part      bool
	lastSQL   string
	lineBuf   []byte
	inQuote   rune
	inComment bool
	isEscape  bool
}

// New create net SQLSplitter
func New(filename string, dialect Dialect) (*SQLSplitter, error) {
	var reader *bufio.Reader

	file, err := os.Open(filename)
	if err != nil {
		reader = bufio.NewReader(strings.NewReader(filename))
	} else {
		reader = bufio.NewReader(file)
	}

	var escape rune
	switch dialect {
	default:
		escape = '\\'
	}

	return &SQLSplitter{
		FileName:  filename,
		Reader:    reader,
		Dialect:   dialect,
		delimiter: ";",
		Escape:    escape,
	}, nil
}

// Next read quries line by line and check if has next
func (s *SQLSplitter) Next() bool {
	for {
		s.readline()
		if len(s.lineBuf) == 0 && s.lastSQL == "" && s.Error == io.EOF {
			s.Error = nil
			return false
		}
		s.splitline()
		if s.part {
			continue
		}
		if s.lastSQL == "" {
			return false
		}
		return true
	}
}

// Scan return current sql
func (s *SQLSplitter) Scan() string {
	sql := s.lastSQL
	s.lastSQL = ""
	return sql
}

// Statements return all sqls in one time, huge memory cost if large file
func (s *SQLSplitter) Statements() ([]string, error) {
	var statements []string
	for s.Next() {
		statements = append(statements, s.Scan())
	}
	return statements, s.Error
}

func (s *SQLSplitter) readline() {
	// lineBuf has not used content
	if len(s.lineBuf) > 0 {
		return
	}
	// The End
	if s.Error == io.EOF {
		return
	}

	line, isPrefix, err := s.Reader.ReadLine()
	if err != nil {
		s.Error = err
	}
	s.lineBuf = append(s.lineBuf, line...)
	for isPrefix {
		var buf []byte
		buf, isPrefix, err = s.Reader.ReadLine()
		if err != nil {
			s.Error = err
		}
		s.lineBuf = append(s.lineBuf, buf...)
	}
	s.LastLine++
}

func (s *SQLSplitter) splitline() {
	switch s.Dialect {
	default:
		s.mysqlSplitter()
	}
}
