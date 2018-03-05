package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/yanyiwu/gojieba"
)

type freqMap map[string]int
type markovChain map[string]freqMap

var (
	file        string
	mChain      markovChain
	startOfWord map[string]struct{}
)

func main() {
	pflag.StringVar(&file, "file", "article.txt", "File to be read")
	pflag.Parse()

	j := gojieba.NewJieba()
	defer j.Free()
	fp, err := os.Open(file)
	if err != nil {
		fmt.Println("Cannot read file:", err)
		return
	}
	defer fp.Close()
	sc := bufio.NewScanner(fp)
	b := bytes.Buffer{}
	for sc.Scan() {
		b.WriteString(strings.TrimSpace(sc.Text()))
		b.WriteString(" ")
	}
	mChain = make(markovChain)
	startOfWord = make(map[string]struct{})
	words := j.Cut(b.String(), true)
	for i := range words {
		if i == len(words)-1 {
			break
		}
		if !strings.ContainsAny(words[i], "。，「」《》 ‧") {
			// Not the end of words
			if _, exists := mChain[words[i]]; exists {
				wMap := mChain[words[i]]
				if _, swExist := wMap[words[i+1]]; swExist {
					wMap[words[i+1]]++
				} else {
					wMap[words[i+1]] = 1
				}
			} else {
				wMap := make(freqMap)
				wMap[words[i+1]] = 1
				mChain[words[i]] = wMap

			}
		} else {
			startOfWord[words[i+1]] = struct{}{}
		}
	}

	/* for w, fMap := range mChain {
		fmt.Println(w + ":")
		for w, freq := range fMap {
			fmt.Printf("%v: %v, ", w, freq)
		}
		fmt.Println("\n")
	} */

	fmt.Printf("Summary: %v words\n", len(mChain))

	for i := 0; i < 7; i++ {
		w := pickFirstWord(startOfWord)
		fmt.Println(generateSentence(w, 5, 15))
	}
}

func pickFirstWord(m map[string]struct{}) (firstWord string) {
	for w := range m {
		return w
	}
	return
}

func pickNextWord(cWord string, m markovChain, tryToEnd bool) (nextWord string, ended bool) {
	subMap := m[cWord]
	if _, exists := m["。"]; tryToEnd && exists {
		return " ", true
	}
	for subW := range subMap {
		if !strings.ContainsAny(subW, "。「」《》 ‧\n") {
			return subW, false
		}
	}

	return "", false
}

func generateSentence(startingWord string, minLen, maxLen int) string {
	bs := strings.Builder{}
	cWord := startingWord
	end := false
	i := 0
	bs.WriteString(cWord)
	for !end && i < maxLen {
		cWord, end = pickNextWord(cWord, mChain, i >= minLen)
		bs.WriteString(cWord)
		if cWord == "。" {
			break
		}
		i++
	}
	return bs.String()
}
