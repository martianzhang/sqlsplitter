package sqlsplitter

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test splitter
var ts *SQLSplitter

var testInput = "select 0;\n" + // simple sql
	"select 1;select 2;\n" + // single line multi sql
	" select 3;\n" + // add space prefix
	"	select 4;\n" + // add tab prefix
	"\n" + // empty line, will ignore
	"select 'single ; quote';\n" + // single quote
	"select `back ; tick`;\n" + // back tick
	`select "double ; quote";` + // double quote
	`select "escape double \"; quote";` + // escape double quote
	`select 'escape single \'; quote';` + // escape single quote
	`select 'escape delimiter \;';` + // escape delimiter
	`select 'mixed quote";';` + // mixed quote
	`select "mixed quote';";` + // mixed quote
	"select `mixed quote';`;\n" + // mixed quote
	"delimiter ;;\n" + // update delimiter
	"select 1;select 2;;\n" + // use new delimiter
	"delimiter ;\n" + // update delimiter
	"select\n1;\n" + // multi line sql
	"select \n\n1;\n" + // empty line in sql
	"# comment\nselect 1;\n" + // single line comment
	"-- comment\nselect 1;\n" + // single line comment
	"select /* comment */1;\n" + // single line comment
	"-- comment\n\nselect 1;\n" + // single line comment with empty line
	"select /*comment\n*/1;\n" + // multi line comment
	"select /*comment;*/1;\n" + // delimiter in comment
	"-- comment ;\nselect 1;\n" + // delimiter in comment
	"--\nselect 1;\n" + // single line comment
	"select 5" // not end with delimiter

var testWant = []string{
	"select 0;", // single sql
	"select 1;",
	"select 2;",                         // single line multi sql
	" select 3;",                        // add space prefix
	"	select 4;",                        // add tab prefix
	"select 'single ; quote';",          // single quote
	"select `back ; tick`;",             // back tick
	`select "double ; quote";`,          // double quote
	`select "escape double \"; quote";`, // escape double quote
	`select 'escape single \'; quote';`, // escape single quote
	`select 'escape delimiter \;';`,     // escape delimiter
	`select 'mixed quote";';`,           // mixed quote
	`select "mixed quote';";`,           // mixed quote
	"select `mixed quote';`;",           // mixed quote
	"delimiter ;;",                      // update delimiter
	"select 1;select 2;;",               // use new delimiter
	"delimiter ;",                       // update delimiter
	"select\n1;",                        // multi line sql
	"select \n\n1;",                     // empty line in sql
	"# comment\nselect 1;",              // single line comment
	"-- comment\nselect 1;",             // single line comment
	"select /* comment */1;",            // single line comment
	"-- comment\n\nselect 1;",           // single line comment with empty line
	"select /*comment\n*/1;",            // multi line comment
	"select /*comment;*/1;",             // delimiter in comment
	"-- comment ;\nselect 1;",           // delimiter in comment
	"--\nselect 1;",                     // single line comment
	"select 5",                          // then end without delimiter
}

func init() {
	var err error
	ts, err = New(testInput, Default)
	if err != nil {
		fmt.Println(err)
	}
}

func TestStatements(t *testing.T) {
	stmts, err := ts.Statements()
	assert.Nil(t, err)
	for i := range testWant {
		assert.Equal(t, testWant[i], stmts[i])
	}
}

func TestNext(t *testing.T) {
	var i int
	for ts.Next() {
		assert.Equal(t, testWant[i], ts.Scan())
		i++
	}
	assert.Nil(t, ts.Error)
}

func BenchmarkStatements(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ts.Statements()
	}
}

func Test_largesql(t *testing.T) {
	query := "select '" + strings.Repeat("a", 1<<20*10) + "';" // ~10MB
	t1, err := New(query, Default)
	assert.Nil(t, err)
	for t1.Next() {
		assert.Equal(t, len(t1.Scan()), len(query))
	}
}

func Test_readline(t *testing.T) {
	data := struct {
		Query string
		Want  string
	}{
		Query: "line1\nline2",
		Want:  "line1",
	}

	t1, err := New(data.Query, Default)
	assert.Nil(t, err)
	t1.readline()
	assert.Equal(t, data.Want, string(t1.lineBuf))
}

func Test_next(t *testing.T) {
	data := struct {
		Query string
		Want  string
	}{
		Query: " ",
		Want:  " ",
	}

	t1, err := New(data.Query, Default)
	assert.Nil(t, err)
	t1.Next()
	assert.Equal(t, data.Want, string(t1.lastSQL))
}

func TestDialectString(t *testing.T) {
	// Test case 1: Test the default dialect
	dialect1 := Default
	assert.Equal(t, "Default", dialect1.String())

	// Test case 2: Test the MySQL dialect
	dialect2 := MySQL
	assert.Equal(t, "MySQL", dialect2.String())

	// Test case 3: Test the PostgreSQL dialect
	dialect3 := PostgreSQL
	assert.Equal(t, "PostgreSQL", dialect3.String())

	// Test case 4: Test the SQLite dialect
	dialect4 := SQLite
	assert.Equal(t, "SQLite", dialect4.String())

	// Test case 5: Test the MSSQL dialect
	dialect5 := MSSQL
	assert.Equal(t, "MSSQL", dialect5.String())

	// Test case 6: Test the Oracle dialect
	dialect6 := Oracle
	assert.Equal(t, "Oracle", dialect6.String())
}

func Test_readline2(t *testing.T) {
	tf, err := randomTestFile()
	if err != nil {
		panic(err)
	}
	defer func() {
		// Remove the file after the test
		os.Remove(tf)
	}()

	data := struct {
		Query string
		Want  string
	}{
		Query: tf,
		Want:  "",
	}

	t1, err := New(data.Query, Default)
	assert.Nil(t, err)
	t1.Next()
	assert.Equal(t, data.Want, string(t1.lastSQL))
}

func randomTestFile() (string, error) {
	err := os.MkdirAll("fixture", os.ModePerm)
	if err != nil {
		return "", err
	}
	tf, err := os.CreateTemp("fixture", "test.")
	if err != nil {
		return "", err
	}
	return tf.Name(), nil
}
