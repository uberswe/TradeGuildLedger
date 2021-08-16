package parser

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"sync"
)

var (
	p ParsedData
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

// Inspired by https://medium.com/swlh/processing-16gb-file-in-seconds-go-lang-3982c235dfa2
func LuaChunkParser(filePath string) (ParsedData, error) {
	p = ParsedData{}
	file, err := os.Open(filePath)

	if err != nil {
		return p, err
	}

	defer file.Close() //close after checking err

	filestat, err := file.Stat()
	if err != nil {
		return p, err
	}

	fileSize := filestat.Size()
	offset := fileSize - 1
	lastLineSize := 0

	for {
		b := make([]byte, 1)
		n, err := file.ReadAt(b, offset)
		if err != nil {
			break
		}
		char := string(b[0])
		if char == "\n" {
			break
		}
		offset--
		lastLineSize += n
	}

	lastLine := make([]byte, lastLineSize)
	_, err = file.ReadAt(lastLine, offset+1)

	if err != nil {
		return p, err
	}

	err = Process(file)

	if err != nil {
		return p, err
	}
	return p, nil
}

// Inspired by https://medium.com/swlh/processing-16gb-file-in-seconds-go-lang-3982c235dfa2
func Process(f *os.File) error {
	linesPool := sync.Pool{New: func() interface{} {
		lines := make([]byte, 4000*1024)
		return lines
	}}

	stringPool := sync.Pool{New: func() interface{} {
		lines := ""
		return lines
	}}

	r := bufio.NewReader(f)

	var wg sync.WaitGroup

	for {
		buf := linesPool.Get().([]byte)

		n, err := r.Read(buf)
		buf = buf[:n]

		if n == 0 {
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(err)
				break
			}
			return err
		}

		nextUntillNewline, err := r.ReadBytes('\n')

		if err != io.EOF {
			buf = append(buf, nextUntillNewline...)
		}

		wg.Add(1)
		go func() {
			ProcessChunk(buf, &linesPool, &stringPool)
			wg.Done()
		}()
	}

	wg.Wait()
	return nil
}

// Inspired by https://medium.com/swlh/processing-16gb-file-in-seconds-go-lang-3982c235dfa2
func ProcessChunk(chunk []byte, linesPool *sync.Pool, stringPool *sync.Pool) {
	var wg2 sync.WaitGroup
	mu := &sync.Mutex{}
	sections := stringPool.Get().(string)
	sections = string(chunk)

	linesPool.Put(chunk)

	logsSlice := strings.Split(sections, "\n")

	stringPool.Put(sections)

	chunkSize := 300
	n := len(logsSlice)
	noOfThread := n / chunkSize

	if n%chunkSize != 0 {
		noOfThread++
	}

	for i := 0; i < (noOfThread); i++ {

		wg2.Add(1)
		go func(s int, e int) {
			defer wg2.Done() //to avaoid deadlocks
			if s != 0 || e != len(logsSlice) {
			}
			for i := s; i < e; i++ {
				text := logsSlice[i]
				if len(text) == 0 {
					continue
				}
				if strings.Contains(text, "\"l:") {
					// listing
					s := GetStringInBetween(text, "\"l:", "\",")
					// if not empty
					if s != "" {
						mu.Lock()
						p.Listings = append(p.Listings, s)
						mu.Unlock()
					}
				} else if strings.Contains(text, "\"i:") {
					// item
					s := GetStringInBetween(text, "\"i:", "\",")
					// if not empty
					if s != "" {
						mu.Lock()
						p.Items = append(p.Items, s)
						mu.Unlock()
					}
				} else if strings.Contains(text, "\"g:") {
					// guild
					s := GetStringInBetween(text, "\"g:", "\",")
					// if not empty
					if s != "" {
						mu.Lock()
						p.Guilds = append(p.Guilds, s)
						mu.Unlock()
					}
				} else if strings.Contains(text, "\"s:") {
					// buy
					s := GetStringInBetween(text, "\"s:", "\",")
					// if not empty
					if s != "" {
						mu.Lock()
						p.Buys = append(p.Buys, s)
						mu.Unlock()
					}
				} else if strings.Contains(text, "\"w:") {
					// server
					s := GetStringInBetween(text, "\"w:", "\"")
					// if not empty
					if s != "" {
						mu.Lock()
						p.Server = s
						mu.Unlock()
					}
				} else if strings.Contains(text, "\"r:") {
					// regions
					s := GetStringInBetween(text, "\"r:", "\",")
					// if not empty
					if s != "" {
						mu.Lock()
						p.Regions = append(p.Regions, s)
						mu.Unlock()
					}
				} else if strings.Contains(text, "\"t:") {
					// traits
					s := GetStringInBetween(text, "\"t:", "\",")
					// if not empty
					if s != "" {
						mu.Lock()
						p.Traits = append(p.Traits, s)
						mu.Unlock()
					}
				} else if strings.Contains(text, "[\"tglv\"]") {
					// addon version
					s := GetStringInBetween(text, "[\"tglv\"] = \"", "\"")
					// if not empty
					if s != "" {
						mu.Lock()
						p.Version = s
						mu.Unlock()
					}
				} else if strings.Contains(text, "[\"@") {
					// username
					s := GetStringInBetween(text, "[\"@", "\"")
					// if not empty
					if s != "" {
						mu.Lock()
						p.Username = s
						mu.Unlock()
					}
				}
			}
		}(i*chunkSize, int(math.Min(float64((i+1)*chunkSize), float64(len(logsSlice)))))
	}

	wg2.Wait()
	logsSlice = nil
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
