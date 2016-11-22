package main

import (
	"fmt"
	"os"
	"sort"
)

var eoS error

const unique = true
const parallel = false
const simplifyCount = 0

func init() {
	eoS = fmt.Errorf("End of string")
}

type word struct {
	value string
}

type letval struct {
	let        rune
	val        int
	components []*letval
}

func (lv *letval) computeValue() {
	if lv.components == nil {
		return
	}
	lv.val = 0
	for _, lvc := range lv.components {
		lv.val += lvc.val
	}
	for lv.val > 9 {
		lv.val %= 10
	}
}

func (lv *letval) String() string {
	return fmt.Sprintf("{%c, %d}", lv.let, lv.val)
}

func (lv *letval) deepcopy() *letval {
	return &letval{let: lv.let, val: lv.val}
}

func (w word) String() string {
	return w.value
}

func newWord(s string) word {
	return word{value: s}
}

func (w word) column(i int) (rune, error) {
	if len(w.value) > i {
		return rune(w.value[len(w.value)-1-i]), nil
	}
	return '0', eoS
}

func (w word) val(chars map[rune]*letval) int64 {
	v := int64(0)
	for _, r := range w.value {
		v = v * 10
		lv := chars[r]
		v += int64(lv.val)
	}
	return v
}

func simplify(sol map[rune]*letval, words []word, sum word, index int) {
	target, _ := sum.column(index)
	lv := sol[target]
	if lv.components != nil {
		fmt.Printf("cannot simplfy %c again\n", target)
		return
	}
	for _, w := range words {
		r, _ := w.column(index)
		lv.components = append(lv.components, sol[r])
	}
	fmt.Printf("%c is made from %v\n", target, lv.components)
	// now target has components
}

func runeAccount(ru map[rune]int, w word) {
	for _, r := range w.value {
		c, ok := ru[r]
		if !ok {
			ru[r] = 1
		} else {
			ru[r] = c + 1
		}
	}
}

func valid(sol map[rune]*letval) bool {
	// compute all simplified
	for _, lv := range sol {
		lv.computeValue()
	}

	// check for unique
	if !unique {
		return true
	}
	used := make(map[int]bool)
	for _, lv := range sol {
		if _, ok := used[lv.val]; ok {
			return false
		}
		used[lv.val] = true
	}
	return true
}

func generateNext(seq []*letval, sol map[rune]*letval) bool {
	for {
		for i, v := range seq {
			if v.val == 9 {
				if i == len(seq)-1 {
					return false
				}
				v.val = 0
			} else {
				v.val++
				if valid(sol) {
					return true
				}
				break
			}
		}
	}
}

func copysol(sol map[rune]*letval) map[rune]*letval {
	res := make(map[rune]*letval)
	for k, v := range sol {
		res[k] = v.deepcopy()
	}
	return res
}

func generate(ru map[rune]int, words []word, sum word) {
	// build the list and the map
	sol := make(map[rune]*letval)
	seq := make([]*letval, 0, 12)
	for k := range ru {
		lv := &letval{let: k, val: 0}
		sol[k] = lv
	}

	// try to reduce the complexity
	for i := 0; i < simplifyCount; i++ {
		simplify(sol, words, sum, i)
	}

	// build the seq from what's left
	for _, lv := range sol {
		if lv.components == nil {
			seq = append(seq, lv)
		}
	}

	// This is the evaluation closure
	ef := func(s []*letval, m map[rune]*letval) {
		if evaluate(words, sum, m) {
			printSolV(s, m, words, sum)
		}
	}

	// special first start condition for unique
	if !valid(sol) {
		if !generateNext(seq, sol) {
			return
		}
	}
	// start off with known good
	for {
		if parallel {
			// If we're doing this in parallel then we need to copy the solutions
			s := copysol(sol)
			go ef(seq, s)
		} else {
			// no need to copy since it's one at a time
			ef(seq, sol)
		}
		if !generateNext(seq, sol) {
			break
		}
	}
}

func printSol(seq []*letval, sol map[rune]*letval, words []word, sum word) {
	f := make([]string, 0, 12)
	for k := range sol {
		f = append(f, string(k))
	}
	sort.Strings(f)

	fmt.Printf("SOL: ")
	for _, s := range f {
		fmt.Printf("%s", sol[rune(s[0])])
	}
	fmt.Printf("\n")
}

func printSolV(seq []*letval, sol map[rune]*letval, words []word, sum word) {
	for _, w := range words {
		indent := len(sum.value) - len(w.value)
		for i := 0; i < indent; i++ {
			fmt.Printf("      ")
		}
		for _, r := range w.value {
			fmt.Printf("%s", sol[r])
		}
		fmt.Printf("\n")
		//fmt.Printf("[%d]\n", w.val(sol))
	}
	for i := 0; i < len(sum.value); i++ {
		fmt.Printf("------")
	}
	fmt.Printf("\n")
	for _, r := range sum.value {
		fmt.Printf("%s", sol[r])
	}
	fmt.Printf("\n")
	//fmt.Printf("[%d]\n", sum.val(sol))
	fmt.Printf("\n")
	printSol(seq, sol, words, sum)
	fmt.Printf("\n")
}

func evaluate(words []word, sum word, sol map[rune]*letval) bool {
	wsum := int64(0)
	for _, w := range words {
		wsum += w.val(sol)
	}
	if sum.val(sol) == wsum {
		return true
	}
	return false
}

func main() {
	var words []word
	var sum word
	runeUse := make(map[rune]int)

	// parse
	for _, a := range os.Args[1 : len(os.Args)-1] {
		words = append(words, newWord(a))
	}
	sum = newWord(os.Args[len(os.Args)-1])

	// account
	for _, a := range words {
		runeAccount(runeUse, a)
	}
	runeAccount(runeUse, sum)

	// generate
	generate(runeUse, words, sum)
}
