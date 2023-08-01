package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(rMain())
}

func rMain() int {
	var dSize = flag.Uint("s", 65536, "size of data array")
	var iFile = flag.String("i", "", "brainfuck file to run")
	var info = flag.Bool("v", false, "output info about run after successfully completed")

	flag.Parse()

	var err error

	var ip int
	var inst []byte

	var dp int
	var data []byte

	var eof bool
	var bLevel int
	//giving it a position of a bracket it returns the corresponding bracket position
	//var jumpMap map[int]int
	var jumpMap []int
	var instCount uint64

	if *iFile != "" {
		inst, err = os.ReadFile(*iFile)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "couldn't read file:", err)
			return 1
		}

		//"remove" invalid instructions and comments
		var vPos int
		var comment bool
		for i := 0; i < len(inst); i++ {
			switch inst[i] {
			case '>', '<', '+', '-', '.', ',', '[', ']', '/':
				if !comment {
					if inst[i] == '/' {
						if i < len(inst)-1 && inst[i+1] == '/' {
							comment = true
						}
						//swallow / if it's not part of a pair
					} else {
						inst[vPos] = inst[i]
						vPos++
					}
				}
			case '\n':
				comment = false
			}
		}
		inst = inst[:vPos]
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "no input file specified")
		return 1
	}

	//generate jump map
	//todo: optimize
	//jumpMap = make(map[int]int)
	jumpMap = make([]int, len(inst))
	{
		for i := 0; i < len(inst); i++ {
			switch inst[i] {
			case '[':
				ip = i
				bLevel = 0
				for {
					if ip >= len(inst) {
						_, _ = fmt.Fprintln(os.Stderr, "unterminated [; reached end of instructions.")
						return 1
					}
					switch inst[ip] {
					case '[':
						bLevel++
					case ']':
						bLevel--
						if bLevel <= 0 {
							jumpMap[i] = ip
							goto next
						}
					}
					ip++
				}
			case ']':
				ip = i
				bLevel = 0
				for {
					if ip >= len(inst) {
						_, _ = fmt.Fprintln(os.Stderr, "unterminated ]; reached beginning of instructions.")
						return 1
					}
					switch inst[ip] {
					case '[':
						bLevel++
						if bLevel >= 0 {
							jumpMap[i] = ip
							goto next
						}
					case ']':
						bLevel--
					}
					ip--
				}
			}
		next:
		}
		ip = 0
	}

	data = make([]byte, *dSize)

	for ip < len(inst) {
		instCount++
		switch inst[ip] {
		case '>':
			dp++
		case '<':
			dp--
		case '+', '-', '.', ',', '[', ']':
			if dp < 0 || dp > len(data) {
				_, _ = fmt.Fprintln(os.Stderr, "illegal data access:", dp)
				return 1
			} else {
				switch inst[ip] {
				case '+':
					data[dp]++
				case '-':
					data[dp]--
				case '.':
					//fmt.Println doesn't need error checking usually so none here either
					_, _ = os.Stdout.Write([]byte{data[dp]})
				case ',':
					if eof {
						_, _ = fmt.Fprintln(os.Stderr, "waiting on input that won't come; exiting.")
						return 1
					}
					_, err = os.Stdin.Read(data[dp : dp+1])
					switch err {
					case nil:
						break
					case io.EOF:
						eof = true
					default:
						_, _ = fmt.Fprintln(os.Stderr, "unknown error reading:", err)
						return 1
					}
				case '[':
					if data[dp] == 0 {
						ip = jumpMap[ip]
					}
				case ']':
					if data[dp] != 0 {
						ip = jumpMap[ip]
					}
				}
			}
		default:
			_, _ = fmt.Fprintln(os.Stderr, "invalid instruction:", inst[ip])
			return 1
		}

		ip++
	}

	if *info {
		fmt.Println("instructions ran:", instCount)
	}

	return 0
}
