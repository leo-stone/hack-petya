package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/handcraftsman/GeneticGo"
	"github.com/willf/bitset"
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

type (
	petya_matrix struct {
		Const0  uint16
		Key0    uint16
		Key2    uint16
		Key4    uint16
		Key6    uint16
		Const2  uint16
		Nounce0 uint16
		Nounce2 uint16
		Counter uint32
		Const4  uint16
		Key8    uint16
		Key10   uint16
		Key12   uint16
		Key14   uint16
		Const6  uint16
	}

	test_key_data struct {
		m_org *petya_matrix
		ow    []uint16
		oq    []uint64
		m_cpy *petya_matrix
		cw    []uint16
		cq    []uint64
	}
)

var (
	key           = []int{0, 0, 0, 0, 0, 0, 0, 0}
	alpha         = make_key_alphabet()
	map_alpha     = make(map[byte]uint16)
	nounce        []byte
	target_words  *[8][16]uint16
	rate_counter  uint64
	target_bitset *bitset.BitSet
)

func try_genetic_approach() {

	solver := new(genetic.Solver)
	solver.MaxSecondsToRunWithoutImprovement = 20 // you decide
	solver.LowerFitnessesAreBetter = true         // you decide

	m_template := make_petya_matrix(nounce)

	getFitness := func(candidate string) int {
		if len(candidate) != 8 {
			panic("Somthing is wrong with that candiate: " + candidate)
		}

		b := []byte(candidate)

		m := m_template.clone()

		m.Key0 = map_alpha[b[0]]
		m.Key2 = map_alpha[b[1]]
		m.Key4 = map_alpha[b[2]]
		m.Key6 = map_alpha[b[3]]
		m.Key8 = map_alpha[b[4]]
		m.Key10 = map_alpha[b[5]]
		m.Key12 = map_alpha[b[6]]
		m.Key14 = map_alpha[b[7]]

		mw := m.words()

		c := m.clone()
		c.shuffle()

		cw := c.words()
		for i, w := range mw {
			cw[i] += w
		}

		return int(bitset.From(c.qwords()).SymmetricDifferenceCardinality(target_bitset))

	}

	// create a display function
	display := func(genes string) {

		key := genesToKey(genes)

		println(key, "score:", getFitness(genes), " (lower is better)") // provide some output to the user if desired
	}

	// each gene is a single character
	geneSet := "123456789abcdefghijkmnopqrstuvwxABCDEFGHJKLMNPQRSTUVWX" // you decide the set of valid genes
	numberOfGenesInAChromosome := 8                                     // you decide

	solver.NumberOfConcurrentEvolvers = 1 // you decide, defaults to 1
	solver.MaxProcs = 1                   // you decide, defaults to 1

	solver.MaxRoundsWithoutImprovement = 1 // you decide
	// you decide
	//solver.PrintDiagnosticInfo = true
	numberOfChromosomes := 1 // you decide
	var result = solver.GetBest(getFitness, display, geneSet, numberOfChromosomes, numberOfGenesInAChromosome)
	for getFitness(result) > 0 {

		result = solver.With(result).GetBest(getFitness, display, geneSet, numberOfChromosomes, numberOfGenesInAChromosome)

	}
	fmt.Println("Your key is: ", genesToKey(result))
	return
	/*
					bestPossibleFitness := 0                // you decide
		maxNumberOfChromosomes := 1
				var result = solver.GetBestUsingHillClimbing(getFitness, display, geneSet, maxNumberOfChromosomes, numberOfGenesInAChromosome, bestPossibleFitness)
				fmt.Println(result)
	*/
}

func genesToKey(genes string) string {

	key := ""
	for _, v := range genes {
		key += string([]rune{v, 'x'})
	}
	return key
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	var err error
	target_words = new([8][16]uint16)

	//the load encrypted sector 0x37
	src, err := ioutil.ReadFile("src.txt")
	if err != nil {
		panic(err)
	}

	//b==target_words

	b := (*[256]byte)(unsafe.Pointer(target_words))

	for i := 0; i < 256; i = i + 2 {

		b[i] = src[i*2] ^ 0x37
		b[i+1] = src[i*2+1] ^ 0x37

		if false {
			//this is to check the algorithm works correctly
			//given the decryption routine the key 111111111111111
			//it'll decrypt it to this content
			//the algorithm should succed with key 1111111111111111, or any key like 1_1_1_1_1_1_1_1_
			target, err := ioutil.ReadFile("target_key.txt")
			if err != nil {
				panic(err)
			}

			b[i] = src[i*2] ^ target[i*2]
			b[i+1] = src[i*2+1] ^ target[i*2+1]
		}

	}
	qw := (*[32]uint64)(unsafe.Pointer(target_words))

	target_bitset = bitset.From(qw[:4])

	fmt.Println(hex.Dump(b[:]))
	fmt.Println(target_words[0])

	nounce, err = ioutil.ReadFile("nonce.txt")

	if err != nil {
		panic(err)
	}

	hex.Dump(nounce)

	try_genetic_approach()

	return

	for i := 0; i < 24*24*24; i++ {
		go check_loop(uint64(i))
	}

	s := time.Now()
	for {
		time.Sleep(5 * time.Second)
		spent := time.Since(s)

		keys_checked := atomic.LoadUint64(&rate_counter)

		rate := keys_checked / uint64(spent/time.Second)
		ptotal := float64(keys_checked) * 100.0 / math.Pow(float64(len(alpha)), 8)

		fmt.Printf("keys/sec: %d progress: %f\n", rate, ptotal)
		time.Sleep(time.Second * 10)
	}
	return
}

func check_loop(start uint64) {
	num_alpha := uint64(len(alpha))
	counter := [8]uint64{}

	s := start * (num_alpha * num_alpha * num_alpha * num_alpha * num_alpha * num_alpha * num_alpha * num_alpha) / 24
	fmt.Printf("start %X\n", s)

	m := make_petya_matrix(nounce)
	pointers := []*uint16{&m.Key0, &m.Key2, &m.Key4, &m.Key6, &m.Key8, &m.Key10, &m.Key12, &m.Key14}

	for i := 0; i < 8; i++ {

		r := s % num_alpha
		s = s / num_alpha

		*pointers[i] = alpha[r]

	}

	//SOME REST OF A FIRST RANDOM ORDER APPROACH
	//time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))
	//rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	/*
		m.Key0 = alpha[0]
		m.Key2 = alpha[0]
		m.Key4 = alpha[0]
		m.Key6 = alpha[0]
		m.Key8 = alpha[0]
		m.Key10 = alpha[0]
		m.Key12 = alpha[0]
		m.Key14 = alpha[0]
	*/

	tkd := test_key_data{}
	tkd.m_org = m
	tkd.ow = m.words()
	tkd.oq = m.qwords()

	m_copy := petya_matrix{}
	tkd.m_cpy = &m_copy
	tkd.cw = m_copy.words()
	tkd.cq = m_copy.qwords()
	for {

		if test_key(&tkd) {
			fmt.Println("found:", m.plain_key())
			os.Exit(0)
			return
		}
		atomic.AddUint64(&rate_counter, 1)

		if m.Counter > 0 {
			fmt.Println(m.plain_key(), m.Counter)
		}
		for j := 0; j < 8; j++ {

			a := counter[j] + 1
			if a == num_alpha {
				counter[j] = 0
				*pointers[j] = alpha[0]
			} else {
				counter[j] = a
				*pointers[j] = alpha[a]
				break
			}

		}

		//*pointers[i%8] = alpha[rnd.Intn(int(num_alpha))] //ALSO REST OF RANDOM ORDER APPROACH
		m.Counter = 0

	}

}

func test_key(tkd *test_key_data) bool {

	//fmt.Println(len(ow))
	cw := tkd.cw
	ow := tkd.ow
	cq := tkd.cq
	oq := tkd.oq
	m_cpy := tkd.m_cpy

	for _, t := range target_words {
		copy(cq, oq)
		m_cpy.shuffle()
		//m := org_m.clone()
		//m.shuffle()

		//s := m.words()
		for j, tw := range t {
			//fmt.Printf("%0X %0X ", s[j]+ow[j], tw)
			if cw[j]+ow[j] != tw {
				return false
			}
		}
		//fmt.Println("")
		tkd.m_org.Counter++
		//fmt.Println(i)

	}

	//fmt.Println("shuff:\n" + hex.Dump(m.bytes()))
	return true
}

func make_petya_matrix(nounce []byte) *petya_matrix {
	// c=0x6578 7061 6e64 2033 322d 6279 7465 206b  "expand 32-byte k"

	r := petya_matrix{}

	r.Const0 = 0x7865
	r.Const2 = 0x646e
	r.Const4 = 0x2d32
	r.Const6 = 0x6574

	r.Nounce0 = uint16(nounce[0]) | uint16(nounce[1])<<8
	r.Nounce2 = uint16(nounce[4]) | uint16(nounce[5])<<8
	return &r
}

func make_key_alphabet() []uint16 {

	s := []byte("123456789abcdefghijkmnopqrstuvwxABCDEFGHJKLMNPQRSTUVWX")

	r := make([]uint16, len(s))
	for i, v := range s {

		r[i] = uint16(v<<1)<<8 | uint16(v+'z')
		map_alpha[v] = r[i]

	}

	return r

}

func (this *petya_matrix) bytes() []byte {

	h := reflect.SliceHeader{}
	h.Cap = int(unsafe.Sizeof(*this))
	h.Len = h.Cap
	h.Data = uintptr(unsafe.Pointer(this))

	return *(*[]byte)(unsafe.Pointer(&h))

}

func (this *petya_matrix) words() []uint16 {

	h := reflect.SliceHeader{}
	h.Cap = int(unsafe.Sizeof(*this)) >> 1
	h.Len = h.Cap
	h.Data = uintptr(unsafe.Pointer(this))

	return *(*[]uint16)(unsafe.Pointer(&h))

}

func (this *petya_matrix) qwords() []uint64 {

	h := reflect.SliceHeader{}
	h.Cap = int(unsafe.Sizeof(*this)) >> 3
	h.Len = h.Cap
	h.Data = uintptr(unsafe.Pointer(this))

	return *(*[]uint64)(unsafe.Pointer(&h))

}

func (this *petya_matrix) clone() *petya_matrix {

	r := petya_matrix{}
	rb := r.bytes()

	copy(rb, this.bytes())
	return &r

}

func (this *petya_matrix) plain_key() string {

	return string([]byte{

		byte(this.Key0 >> 9),
		'x',
		byte(this.Key2 >> 9),
		'x',
		byte(this.Key4 >> 9),
		'x',
		byte(this.Key6 >> 9),
		'x',
		byte(this.Key8 >> 9),
		'x',
		byte(this.Key10 >> 9),
		'x',
		byte(this.Key12 >> 9),
		'x',
		byte(this.Key14 >> 9),
		'x'})

}

func (this *petya_matrix) shuffle() {

	me := this.words()

	for i := 0; i < 10; i++ {

		u := uint32(me[0] + me[12])
		me[4] ^= uint16(u<<7 | u>>(32-7))
		u = uint32(me[4] + me[0])
		me[8] ^= uint16(u<<9 | u>>(32-9))
		u = uint32(me[8] + me[4])
		me[12] ^= uint16(u<<13 | u>>(32-13))
		u = uint32(me[12] + me[8])
		me[0] ^= uint16(u<<18 | u>>(32-18))

		u = uint32(me[5] + me[1])
		me[9] ^= uint16(u<<7 | u>>(32-7))
		u = uint32(me[9] + me[5])
		me[13] ^= uint16(u<<9 | u>>(32-9))
		u = uint32(me[13] + me[9])
		me[1] ^= uint16(u<<13 | u>>(32-13))
		u = uint32(me[1] + me[13])
		me[5] ^= uint16(u<<18 | u>>(32-18))

		u = uint32(me[10] + me[6])
		me[14] ^= uint16(u<<7 | u>>(32-7))
		u = uint32(me[14] + me[10])
		me[2] ^= uint16(u<<9 | u>>(32-9))
		u = uint32(me[2] + me[14])
		me[6] ^= uint16(u<<13 | u>>(32-13))
		u = uint32(me[6] + me[2])
		me[10] ^= uint16(u<<18 | u>>(32-18))

		u = uint32(me[15] + me[11])
		me[3] ^= uint16(u<<7 | u>>(32-7))
		u = uint32(me[3] + me[15])
		me[7] ^= uint16(u<<9 | u>>(32-9))
		u = uint32(me[7] + me[3])
		me[11] ^= uint16(u<<13 | u>>(32-13))
		u = uint32(me[11] + me[7])
		me[15] ^= uint16(u<<18 | u>>(32-18))

		u = uint32(me[0] + me[3])
		me[1] ^= uint16(u<<7 | u>>(32-7))
		u = uint32(me[1] + me[0])
		me[2] ^= uint16(u<<9 | u>>(32-9))
		u = uint32(me[2] + me[1])
		me[3] ^= uint16(u<<13 | u>>(32-13))
		u = uint32(me[3] + me[2])
		/*#*/ me[0] ^= uint16(u<<18 | u>>(32-18))

		u = uint32(me[5] + me[4])
		/*#*/ me[6] ^= uint16(u<<7 | u>>(32-7))
		u = uint32(me[6] + me[5])
		/*#*/ me[7] ^= uint16(u<<9 | u>>(32-9))
		u = uint32(me[7] + me[6])
		me[4] ^= uint16(u<<13 | u>>(32-13))
		u = uint32(me[4] + me[7])
		/*#*/ me[5] ^= uint16(u<<18 | u>>(32-18))

		u = uint32(me[10] + me[9])
		me[11] ^= uint16(u<<7 | u>>(32-7))
		u = uint32(me[11] + me[10])
		/*#*/ me[8] ^= uint16(u<<9 | u>>(32-9))
		u = uint32(me[8] + me[11])
		/*#*/ me[9] ^= uint16(u<<13 | u>>(32-13))
		u = uint32(me[9] + me[8])
		/*#*/ me[10] ^= uint16(u<<18 | u>>(32-18))

		u = uint32(me[15] + me[14])
		me[12] ^= uint16(u<<7 | u>>(32-7))
		u = uint32(me[12] + me[15])
		me[13] ^= uint16(u<<9 | u>>(32-9))
		u = uint32(me[13] + me[12])
		me[14] ^= uint16(u<<13 | u>>(32-13))
		u = uint32(me[14] + me[13])
		/*#*/ me[15] ^= uint16(u<<18 | u>>(32-18))

	}

	/*
		Const0  uint16 me0!
		Key0    uint16 me1
		Key2    uint16 me2
		Key4    uint16 me3
		Key6    uint16 me4
		Const2  uint16 me5!
		Nounce0 uint16 me6!
		Nounce2 uint16 me7!
		Counter uint32 me8! me9!
		Const4  uint16 me10!
		Key8    uint16 me11
		Key10   uint16 me12
		Key12   uint16 me13
		Key14   uint16 me14
		Const6  uint16 me15!
	*/

}

func petya_key(k string) *[16]byte {
	r := [16]byte{}
	for i, b := range []byte(k) {
		r[i*2] = b + byte('z')
		r[i*2+1] = b + b
	}
	return &r

}
