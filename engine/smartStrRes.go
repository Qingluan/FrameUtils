package engine

import (
	"regexp"
)

var (
	// SmartExtractIP extract ip
	SmartExtractIP = regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])`)
	// SmartExtractURL extract url from string
	SmartExtractURL = regexp.MustCompile(`(https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]`)
	// SmartExtractPair extract pair like a:b a=b
	SmartExtractPair = regexp.MustCompile(`([\w\"\']+[\t\ ]*[:=](?:[\t\ ]*(?:[0-9]+|\".+?\")))`)
	// SmartExtractJSON extract json from string
	SmartExtractJSON = JSONExtract
)

// Pair [2]int
type Pair [2]int

/*JSONExtract function
this function will extract json in mass string
*/
func JSONExtract(raw string) (jstr []string) {
	// st := time.Now()
	// defer func() { log.Println(time.Now().Sub(st)) }()
	length := len(raw)

	// fieldTp := 0 // 0 is key, 1 is value, 2 is array value
	// oneField := ""

	start := 0
	end := 0

	for {
		// isValid := true
		dictGraph := []int{}
		dictGraphMark := []bool{}
		quoteStack := false
		notParseMark := false
		lastOp := ' '
		isStart := false

		// tmpStart := 0
	SCAN:
		for i, c := range raw[start:] {
			// i += start
			end++
			if notParseMark {
				notParseMark = false
				// oneField += string(c)
				continue
			} else if quoteStack {
				if c == '\\' {
					notParseMark = true
					continue
				}
				if c == '"' {
					quoteStack = false
				}
				// log.Println("ff :x \"", i, " | ", raw[tmpStart:i])
				// lastOp = c
				continue
			}

			switch c {
			case '{':
				if !isStart && len(dictGraph) == 0 {
					isStart = true
					// tmpStart = start + i
					// log.Println("====================== sep =========================\n[start]:", tmpStart)
				}
				dictGraph = append(dictGraph, start+i)
				dictGraphMark = append(dictGraphMark, true)

				lastOp = c
			case '}':
				if len(dictGraph) == 0 {
					continue
				}
				dictGraph = append(dictGraph, start+i)
				dictGraphMark = append(dictGraphMark, false)
				if lastOp == ',' {
					// isValid = false
					dictGraph = dictGraph[:len(dictGraph)-1]
					dictGraphMark = dictGraphMark[:len(dictGraphMark)-1]
					// log.Println("'}'break in :", string(lastOp), i, "\n", raw[tmpStart:start+i])
					break SCAN
				} else if lastOp == '{' {

					dictGraph = dictGraph[:len(dictGraph)-1]
					dictGraphMark = dictGraphMark[:len(dictGraphMark)-1]
					// log.Println("'}'break in :", string(lastOp), i, "\n", raw[tmpStart:start+i])
					break SCAN
				}
				lastOp = c
			case '"':

				if len(dictGraph) == 0 {
					continue
				}
				quoteStack = true
				if lastOp == c {
					// isValid = false
					// log.Println("'\"' break in :", string(lastOp), i, "\n", raw[tmpStart:start+i])
					break SCAN
				}
				lastOp = c
			case '\\':

				if len(dictGraph) == 0 {
					continue
				}
				if !notParseMark {
					notParseMark = true
				}
				lastOp = c
			case ':':

				if len(dictGraph) == 0 {
					continue
				}
				if lastOp != '"' {
					// isValid = false

					// log.Println("':' break in :", string(lastOp), i, "\n", raw[tmpStart:start+i])
					break SCAN
				}
				lastOp = c
			case ',':

				if len(dictGraph) == 0 {
					continue
				}
				if !quoteStack {

					if lastOp == c {
						// isValid = false

						// log.Println("',' break in :", string(lastOp), i, "\n", raw[tmpStart:i+start])

						break SCAN
					}
					lastOp = c
				}
			}
		}
		// log.Println("carg:", len(dictGraph), len(dictGraphMark))
		if len(dictGraph) == 0 {
			if end >= length-1 {
				break
			}
			// log.Println(end)
			continue
		}

	RECALL:
		for {
			if len(dictGraphMark) == 0 {
				break
			}
			if dictGraphMark[len(dictGraphMark)-1] {
				dictGraph = dictGraph[:len(dictGraph)-1]
				dictGraphMark = dictGraphMark[:len(dictGraphMark)-1]
			} else {
				break RECALL
			}
		}
		start = end
		if len(dictGraph) == 0 {
			if end >= length-1 {
				break
			}
			// log.Println(end)
			continue
		}

		// {{}{{{}{}{}}}
		outerStack := []int{}
		pairs := []Pair{}
		lastPair := Pair{-1, -1}
		for to := len(dictGraph) - 1; to >= 0; to-- {

			switch dictGraphMark[to] {
			case false:
				outerStack = append(outerStack, dictGraph[to])
			case true:
				tmpStart := dictGraph[to]
				if len(outerStack) > 0 {

					tmpEnd := outerStack[len(outerStack)-1]

					// log.Println("p:", tmpStart, tmpEnd, "last:", lastPair)
					if tmpStart < lastPair[0] && tmpEnd > lastPair[1] {
						lastPair[1] = outerStack[len(outerStack)-1]
						lastPair[0] = dictGraph[to]

					} else {
						if lastPair[0] == -1 {
							lastPair[1] = tmpEnd
							lastPair[0] = tmpStart
						} else {
							pairs = append(pairs, lastPair)
							lastPair = Pair{tmpStart, tmpEnd}
						}

					}

					outerStack = outerStack[:len(outerStack)-1]

				}
			}
		}
		pairs = append(pairs, lastPair)
		if lastPair[0] == -1 || lastPair[1] == -1 {
			if end >= length {
				break
			} else {
				continue
			}
		}
		for _, p := range pairs {
			jstr = append(jstr, raw[p[0]:p[1]+1])
			// log.Println(p, "Good:---------------\n", raw[p[0]:p[1]+1], "\n--------------")
		}
	}
	return
}
