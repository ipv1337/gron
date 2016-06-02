package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
)

const (
	exitOK = iota
	exitInvalidArgs
)

func main() {
	flag.Parse()

	filename := flag.Arg(0)
	if filename == "" {
		os.Exit(exitInvalidArgs)
	}

	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		todo(err)
	}

	// The 'JSON' be an object, array or scalar,
	// so the best we can do for now is an empty interface type
	var top interface{}
	err = json.Unmarshal(raw, &top)
	if err != nil {
		todo(err)
	}

	ss, err := makeStatements("json", top)
	if err != nil {
		todo(err)
	}

	sort.Sort(ss)

	for _, s := range ss.statements {
		fmt.Println(s)
	}
}

func keyMustBeQuoted(s string) bool {
	r := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
	return !r.MatchString(s)
}

func makePrefix(prev string, next interface{}) (string, error) {
	switch v := next.(type) {
	case int:
		return fmt.Sprintf("%s[%d]", prev, v), nil
	case string:
		if keyMustBeQuoted(v) {
			return fmt.Sprintf("%s[%s]", prev, escape(v)), nil
		} else {
			return fmt.Sprintf("%s.%s", prev, v), nil
		}
	default:
		return "", fmt.Errorf("could not form prefix for %#v", next)
	}
}

type statementGroup struct {
	statements []string
}

func (s *statementGroup) Add(prefix, value string) {
	s.statements = append(s.statements, fmt.Sprintf("%s = %s;", prefix, value))
}

func (s *statementGroup) AddGroup(g *statementGroup) {
	s.statements = append(s.statements, g.statements...)
}

func (s *statementGroup) Len() int {
	return len(s.statements)
}

func (s *statementGroup) Swap(i, j int) {
	s.statements[i], s.statements[j] = s.statements[j], s.statements[i]
}

func (s *statementGroup) Less(i, j int) bool {
	return s.statements[i] < s.statements[j]
}

func (s *statementGroup) Contains(search string) bool {
	for _, i := range s.statements {
		if i == search {
			return true
		}
	}
	return false
}

func makeStatements(prefix string, v interface{}) (*statementGroup, error) {
	ss := &statementGroup{make([]string, 0)}

	switch vv := v.(type) {

	case map[string]interface{}:
		// It's an object
		ss.Add(prefix, "{}")

		for k, sub := range vv {
			newPrefix, err := makePrefix(prefix, k)
			if err != nil {
				return ss, err
			}
			extra, err := makeStatements(newPrefix, sub)
			if err != nil {
				return ss, err
			}
			ss.AddGroup(extra)
		}

	case []interface{}:
		// It's an array
		ss.Add(prefix, "[]")

		for k, sub := range vv {
			newPrefix, err := makePrefix(prefix, k)
			if err != nil {
				return ss, err
			}
			extra, err := makeStatements(newPrefix, sub)
			if err != nil {
				return ss, err
			}
			ss.AddGroup(extra)
		}

	case float64:
		ss.Add(prefix, escape(vv))

	case string:
		ss.Add(prefix, escape(vv))

	case bool:
		ss.Add(prefix, fmt.Sprintf("%t", vv))

	case nil:
		ss.Add(prefix, "null")
	}

	return ss, nil
}

func escape(s interface{}) string {
	// I'm pretty sure it's safe to ignore this error
	// ...maybe. I'll work something into this I promise
	out, _ := json.Marshal(s)
	return string(out)
}

func todo(err error) {
	log.Fatalf("TODO: %s", err)
}
