//
// tab-cef.go
// gera a tabela de premios da megasena para N dezenas
// jogadas (cercadas), com N variando de 6 a 20
// Autor: Noelson Alves Duarte
//
package main

import (
	"fmt"
)

func main () {
	
	var ss, s5, s4, qui, qui4, qua int
	
	fmt.Println("Tabela de prêmios da Mega-sena da CEF")
	fmt.Println("=====================================")
	fmt.Println("Dez \t| Sena \tQui \tQua \t| Quina Qua \t | Quadra")
	for n :=6; n <=20; n++ {
	   // calcula sena, quina, quadra
	   ss = acertos(n,6,6,6) 		// sena
	   s5 = acertos(n,6,6,5) 		// quinas c/ sena
	   s4 = acertos(n,6,6,4) 		// quadras c/ sena
	   qui = acertos(n,6,5,5)		// quina
	   qui4 = acertos(n,6,5,4)		// quadras c/ quina
	   qua = acertos(n,6,4,4)		// quadras
	   fmt.Println(n, " \t| ", ss, " \t", s5, " \t", s4,
			" \t| ", qui, " \t", qui4, " \t |", qua)
	}	
}

func nrcomb(n, m int) int {
	// retorna o nr de combinacoes (n C m)
    var i int
    cnm := int(1)

    if m * 2 > n {
        m = n - m
    }
    for i = 1; i <= m; n, i = n-1, i+1 {
        cnm = cnm * n / i
    }
    return cnm
}

func acertos(m, p, a, k int) int {
	// Usa a formula: C(m-a, p-k) * C(a, k)
	//    (se p-k > m-a, acertos = 0)
	// formula testada para a megasena com m de 6 a 20,
	//    p = 6 e a,k variando de 6 a 4
	// m: nr de dezenas jogadas (cercadas)
	// p: nr de dezenas sorteadas (6 p/ megasena)
	// a: total de dezenas sorteadas existentes no conjunto m
	// k: premios pagos (sena=6, quina=5, quadra=4)
	var tacert int = 0
	if m - a >= p - k {
		tacert = nrcomb(m-a, p-k) * nrcomb (a, k)
	}
	return tacert
}
