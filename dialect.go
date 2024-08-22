package sqlsplitter

import (
	"regexp"
	"strings"
)

// Dialect database dialect
type Dialect int

const (
	Default Dialect = iota
	MySQL
	PostgreSQL
	SQLite
	MSSQL
	Oracle
)

func (d Dialect) String() string {
	switch d {
	case MySQL:
		return "MySQL"
	case PostgreSQL:
		return "PostgreSQL"
	case SQLite:
		return "SQLite"
	case MSSQL:
		return "MSSQL"
	case Oracle:
		return "Oracle"
	default:
		return "Default"
	}
}

func (s *SQLSplitter) mysqlSplitter() {
	var sb strings.Builder
	lineSize := len(s.lineBuf)
	var line = make([]byte, lineSize)
	copy(line, s.lineBuf)

	// check delimiter update
	re := regexp.MustCompile(`(?i)^\s*delimiter\s*(.*)$`)
	if s.lastSQL == "" {
		if matches := re.FindStringSubmatch(string(line)); len(matches) > 1 {
			s.delimiter = matches[1]
			s.lastSQL = string(line)
			s.lineBuf = []byte{}
			s.part = false
			return
		}
	}
	delimiterLen := len(s.delimiter)

	if s.part {
		if lineSize == 0 && s.Error != nil {
			s.part = false
			return
		}
	}

	for i, char := range line {
		// parser quoted string
		if s.inQuote != 0 {
			// escape char in quoted string
			if s.isEscape {
				sb.WriteByte(char)
				s.isEscape = false
				continue
			}
			if char == byte(s.Escape) {
				s.isEscape = true
				sb.WriteByte(char)
				continue
			}

			// quote end
			if char == byte(s.inQuote) {
				s.inQuote = 0
			}
			sb.WriteByte(char)
			continue
		} else if char == '\'' || char == '"' || char == '`' {
			// quote start
			s.inQuote = rune(char)
		}

		// parser comment string
		if s.inComment {
			if i+2 <= lineSize && string(line[i:i+2]) == "*/" {
				s.inComment = false
				i += 1 // skip over "*/"
			}
			sb.WriteByte(char)
			s.part = true
			continue
		} else if i+2 <= lineSize && string(line[i:i+2]) == "/*" {
			s.inComment = true
			s.part = true
			i += 1 // skip over "/*"
		} else if (i+2 <= lineSize && string(line[i:i+2]) == "--") ||
			(string(line[i]) == "#") {
			sb.WriteString(string(line[i:]))
			s.part = true
			break // skip single line comment
		}

		// delimiter separator
		if !s.inComment && s.inQuote == 0 {
			if delimiterLen > 0 && i+delimiterLen <= lineSize && string(line[i:i+delimiterLen]) == s.delimiter {
				sb.WriteString(string(line[i : i+delimiterLen]))
				if s.lastSQL != "" {
					s.lastSQL += "\n" + sb.String()
				} else {
					s.lastSQL = sb.String()
				}
				s.part = false
				s.lineBuf = line[i+delimiterLen:]
				return
			}
		}
		sb.WriteByte(char)
	}
	if s.lastSQL != "" {
		s.lastSQL += "\n" + sb.String()
	} else {
		s.lastSQL = sb.String()
	}
	s.lineBuf = []byte{}
	s.part = true
}
