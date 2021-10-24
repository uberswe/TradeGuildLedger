package parser

import (
	"bufio"
	"os"
	"strings"
)

type ParsedData struct {
	Listings []string
	Items    []string
	Guilds   []string
	Buys     []string
	Regions  []string
	Traits   []string
	Server   string
	Version  string
	Username string
}

func LuaChunkParser(filePath string) (ParsedData, error) {
	p := ParsedData{}
	file, err := os.Open(filePath)

	if err != nil {
		return p, err
	}

	defer file.Close() //close after checking err

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		processLine(scanner.Text(), &p)
	}

	if err := scanner.Err(); err != nil {
		return p, err
	}

	if err != nil {
		return p, err
	}
	return p, nil
}

func processLine(line string, p *ParsedData) {
	if len(line) == 0 {
		return
	}
	line = strings.TrimSpace(line)
	if strings.Contains(line, "\"l:") {
		// listing
		s := GetStringInBetween(line, "\"l:", "\",")
		// if not empty
		if s != "" {
			p.Listings = append(p.Listings, s)
		}
	} else if strings.Contains(line, "\"i:") {
		// item
		s := GetStringInBetween(line, "\"i:", "\",")
		// if not empty
		if s != "" {
			p.Items = append(p.Items, s)
		}
	} else if strings.Contains(line, "\"g:") {
		// guild
		s := GetStringInBetween(line, "\"g:", "\",")
		// if not empty
		if s != "" {
			p.Guilds = append(p.Guilds, s)
		}
	} else if strings.Contains(line, "\"s:") {
		// buy
		s := GetStringInBetween(line, "\"s:", "\",")
		// if not empty
		if s != "" {
			p.Buys = append(p.Buys, s)
		}
	} else if strings.Contains(line, "\"w:") {
		// server
		s := GetStringInBetween(line, "\"w:", "\"")
		// if not empty
		if s != "" {
			p.Server = s
		}
	} else if strings.Contains(line, "\"r:") {
		// regions
		s := GetStringInBetween(line, "\"r:", "\",")
		// if not empty
		if s != "" {
			p.Regions = append(p.Regions, s)
		}
	} else if strings.Contains(line, "\"t:") {
		// traits
		s := GetStringInBetween(line, "\"t:", "\",")
		// if not empty
		if s != "" {
			p.Traits = append(p.Traits, s)
		}
	} else if strings.Contains(line, "[\"tglv\"]") {
		// addon version
		s := GetStringInBetween(line, "[\"tglv\"] = \"", "\"")
		// if not empty
		if s != "" {
			p.Version = s
		}
	} else if strings.Contains(line, "[\"@") {
		// username
		s := GetStringInBetween(line, "[\"@", "\"")
		// if not empty
		if s != "" {
			p.Username = s
		}
	}
}

// https://stackoverflow.com/a/42331558/1260548
func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return
	}
	return str[s : s+e]
}
