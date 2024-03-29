package service

import (
	"exercicio5/shared"
	"math"
	"sync"
)

type SieveCalc struct{}

func (SieveCalc) InvokeSieveCalc(req shared.Request) []int {
	var r []int

	calcType := req.Type
	rng := req.Rng

	switch calcType {
	case "seq":
		r = SieveCalc{}.sieve(rng)
	case "conc":
		r = SieveCalc{}.concSieve(rng)
	case "blk_conc":
		r = SieveCalc{}.blockConcSieve(rng)
	}

	return r
}

//----------------SEQUENCIAL----------------

func (SieveCalc) sieve(rng int) []int {
	rng++
	var composites = make([]bool, rng)

	var rngRoot = int(math.Sqrt(float64(rng)))

	//marca todos os compostos
	for i := 2; i <= rngRoot; i++ {
		if !composites[i] {
			for j := int(math.Pow(float64(i), 2)); j < rng; j += i {
				composites[j] = true
			}
		}
	}

	//extrai os primos do bit array
	var primes = make([]int, 0, 1)
	for i := 2; i < rng; i++ {
		if !composites[i] {
			primes = append(primes, i)
		}
	}

	return primes
}

//----------------CONCORRENTE----------------

func markDivided(idx int, rng_end int, composites []bool, wg *sync.WaitGroup) {
	defer wg.Done()

	if !composites[idx] {
		for j := idx * 2; j < rng_end; j += idx {
			composites[j] = true
		}
	}
}

func (SieveCalc) concSieve(rng int) []int {
	rng++
	var composites = make([]bool, rng)
	var wg sync.WaitGroup

	//cada thread marca multiplos de um numero
	var rngRoot = int(math.Sqrt(float64(rng)))
	for i := 2; i <= rngRoot; i++ {
		wg.Add(1)
		go markDivided(i, rng, composites, &wg)
	}
	wg.Wait()

	var primes = make([]int, 0, 1)

	//extrai primos do bit array
	for i := 2; i < rng; i++ {
		if !composites[i] {
			primes = append(primes, i)
		}
	}

	return primes
}

//----------------CONCORRENTE MELHORADO----------------

func markBlock(start int, end int, firstPrimes *[]int,
	primes *[]int, wg *sync.WaitGroup, mutex *sync.Mutex) []int {
	defer wg.Done()

	rng := end - start + 1
	endRoot := int(math.Sqrt(float64(end)))

	//pula todos os pares, logo tamanho da array precisa ser so metade to range
	var composites = make([]bool, rng/2)

	for _, prime := range *firstPrimes {

		//todos os compostos  maiores q a raiz quadrada do limite ja estarao marcados
		if prime > endRoot {
			break
		}

		//acha primeiro multiplo de i maior que start
		firstMult := ((start + prime - 1) / prime) * prime

		//se i^2 é maior q o primeiro multiplo é melhor so começar de i^2
		sqrdI := prime * prime
		if firstMult < sqrdI {
			firstMult = sqrdI
		}

		//se primeiro multiplo é par pega o proximo multiplo, que será impar
		if (firstMult & 1) == 0 {
			firstMult += prime
		}

		//marca todos os multiplos de i dentro do intervalo
		for j := firstMult; j <= end; j += prime * 2 {
			composites[(j-start)/2] = true
		}

	}

	//extrai todos os primos do bit array
	var slicePrimes = make([]int, 0, 100)

	//extrai todos os primos do bit array
	for i := 0; i < rng/2; i++ {
		if !composites[i] {
			prime := (start + i*2) + 1
			slicePrimes = append(slicePrimes, prime)
		}
	}

	mutex.Lock()
	*primes = append(*primes, slicePrimes...)
	mutex.Unlock()

	return *primes
}

func (SieveCalc) blockConcSieve(rng int) []int {
	primes := []int{2}
	var wg sync.WaitGroup
	var mutex sync.Mutex

	rngRoot := int(math.Sqrt(float64(rng)))

	var firstPrimes []int

	if rngRoot < 10000 {
		firstPrimes = SieveCalc{}.sieve(rngRoot)
		firstPrimes = firstPrimes[1:]
	} else {
		firstPrimes = SieveCalc{}.blockConcSieve(rngRoot)
		firstPrimes = firstPrimes[1:]
	}

	sliceSize := 256 * 1024 //256K * 8B (int tem 8 bytes) / 2 (pares nao sao considerados)= 1MB por thread

	for start := 2; start <= rng; start += sliceSize {
		var end = start + sliceSize

		if end > rng {
			end = rng
		}

		wg.Add(1)
		go markBlock(start, end, &firstPrimes, &primes, &wg, &mutex)
	}

	wg.Wait()

	return primes
}
