//go:generate /home/nadsec/go/bin/fyne bundle -o data.go Icon.png
// - megasena.go: aplicativo para a loteria da Mega-sena (6/60)
// - Autor: Noelson Alves Duarte 
//   (um aprendizado básico de golang + fyne gui toolkit)
// - Data: Mai/2023
// - Facilidades:
//   1: gerar dicas de apostas para a Mega-sena
//   2: aplicar filtros na geração das dicas
//   3: selecionar dicas para as apostas
//   4: arquivar as dicas selecionadas
//   5: recuperar as dicas arquivadas
//   6: conferir os acertos das apostas (dicas selecionadas)
// - OBS: nenhuma verificação de erro é feita
//
// TO DO:
//
package main

import (
	"fmt"
    "strconv"
    "strings"
    "sort"
    "math/rand"
    "time"
    "io"
    "regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/theme"
    "fyne.io/fyne/v2/storage"
)

type megasena struct {
	lblInt *widget.Label
    inpInt *widget.Entry
	lblTam *widget.Label
    inpTam *widget.Entry
	labelc *widget.Label
    inputc *widget.Entry
    labelr *widget.Label
	// saida
    inputr  *widget.Entry
    labelf  *widget.Label
	lbDicas *widget.Label
    sugest  *widget.Entry // MultiLine
	// filtros
	chkCA *widget.Check
	chkPI *widget.Check
	chkMm *widget.Check
	chkGp *widget.Check
    // comandos
	btGerar *widget.Button 
	btAdic  *widget.Button 
	btLimp  *widget.Button
	// janela
	win fyne.Window
}

const maxRep = 1000 // max repeticoes para gerar uma dica filtrada

func (ms *megasena) cmdGerar() {
	var min, max, dpc, nrDezs int
	var dezs []int
	// verifica os dados de entrada
	bt := obtDezCartao (ms.inpTam.Text, &dpc)
	if bt {
		if dpc >= 6 && dpc <= 20 { // faixa de dezenas por aposta no site CEF
			dezenas := make([]int, dpc)
			bi := obtInter (ms.inpInt.Text, &min, &max)
			if bi == false {
				min = 1
				max = 60
			}
			// verificar integridade das dezenas digitadas?
			// bEncont, _ := regexp.MatchString(`[^0-9 ]`, s)
			dezs = str2sli(strings.TrimSpace(ms.inputc.Text))
			nrDezs = len(dezs)
			if nrDezs == 0 {
				// sorteia dezenas com base no intervalo
				if (max-min+1) >= dpc {
					ms.geraPalpite(dezenas, dpc, min, max)
				} else {
					dialog.ShowInformation("Erro", "Intervalo", ms.win)
				}
			} else if nrDezs > dpc {
				// sorteia entre as dezenas digitadas
				if (max-min+1) < nrDezs {
					dialog.ShowInformation("Erro", "Intervalo", ms.win)
				} else {
					ms.geraPalpiteDez(dezenas, dezs, dpc, min, nrDezs)
				}
			} else if nrDezs == dpc {
				// dezenas digitadas sao a propria dica
				dezenas = dezs[:]
			} else if nrDezs < dpc {
				// fixa as dezenas digitadas e sorteia as restantes
				if (max-min+1) < (dpc - nrDezs) {
					dialog.ShowInformation("Erro", "Intervalo", ms.win)
				} else {
					ms.geraPalpiteDezFixa(dezenas, dezs, dpc, min, max)
				}
			}
			// atualiza campo se a dica passar nos filtros
			if dezenas[0] != 0 { // melhor retornar false / true
				sort.Ints(dezenas)
				ms.inputr.SetText(sli2str(dezenas))
			}
		} else {
			dialog.ShowInformation("Erro", "Dezenas por cartão\n de 6 a 20", ms.win)
		}
	} else {
		dialog.ShowInformation("Erro", "Dezenas por cartão", ms.win)
	}
}

func (ms *megasena) cmdAdic() {
	lim := 6
	bb := confereDica(ms.sugest.Text, ms.inputr.Text, lim)
	if bb == false {
		dialog.ShowConfirm("Dezenas Repetidas", "Adicionar?",
			func(value bool) {
				if value == true {
					ms.sugest.SetText(ms.inputr.Text + "\n" + ms.sugest.Text)
				}
			}, ms.win)
	} else {
		ms.sugest.SetText(ms.inputr.Text + "\n" + ms.sugest.Text)
	}
}

func (ms *megasena) cmdLimp() {
	ms.inpInt.SetText("")
	ms.inpTam.SetText("")
	ms.inputc.SetText("")
	ms.inputr.SetText("")
	ms.chkCA.Checked = false
	ms.chkPI.Checked = false
	ms.chkMm.Checked = false
	ms.chkGp.Checked = false
	ms.chkCA.Refresh()
	ms.chkPI.Refresh()
	ms.chkMm.Refresh()
	ms.chkGp.Refresh()
}

func (ms *megasena) cmdSalvar() {
	saveDialog := dialog.NewFileSave(func(arqSai fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ms.win)
			return
		}
		if arqSai == nil { // cancelado
			return
		}
		// grava o arquivo de saida
		arqSai.Write([]byte(strings.TrimSpace(ms.sugest.Text)))
		defer arqSai.Close()
		ms.win.SetTitle(ms.win.Title() + " - " + arqSai.URI().Name())
	}, ms.win)
	// sugestao de nome de arquivo
	saveDialog.SetFileName("megasena.txt")
	saveDialog.Show()
}

func (ms *megasena) cmdAbrir() {
	file_Dialog := dialog.NewFileOpen(
        func(arqEnt fyne.URIReadCloser, err error) {
            // verifica erros
			if err != nil {
				dialog.ShowError(err, ms.win)
				return
			}
			if arqEnt == nil {
				// cancelado
				return
			}
			// le o arquivo de entrada
			if b, err := io.ReadAll(arqEnt); err == nil {
				bEncont, _ := regexp.MatchString(`[^0-9 \n\r\t]`, string(b))
				if bEncont == false {
					ms.sugest.SetText(string(b))
				} else {
					dialog.ShowInformation("Erro", "Arquivo incompatível!", ms.win)
					return
				}
			}
			defer arqEnt.Close()
        }, ms.win)
    // filtro: array strings/extensoes
    file_Dialog.SetFilter(
        storage.NewExtensionFileFilter([]string{".txt"}))
    file_Dialog.Show()
}

func (ms megasena) cmdConferir() {
	// considera as dicas selecionadas como os palpites jogados
	apostas := strings.TrimSpace(ms.sugest.Text)
	bEncont, _ := regexp.MatchString(`[^0-9 \n\r\t]`, apostas)
	if apostas == "" || bEncont == true {
		dialog.ShowInformation("Erro", "Dicas selecionadas", ms.win)
		return
	}
	// monta o dialogo
	lblSort := widget.NewLabel("Dez. sorteadas:")
	entSort := widget.NewEntry()
	lblConf := widget.NewLabel("Nr. de acertos:")
	txtResult := widget.NewMultiLineEntry()
	btnConf := widget.NewButtonWithIcon("Conferir", theme.ConfirmIcon(), func() {
		// dezenas sorteadas (digitos e espacos)
		strSort := strings.TrimSpace(entSort.Text)
		bEncont, _ = regexp.MatchString(`[^0-9 ]`, strSort)
		if strSort == "" || bEncont == true {
			dialog.ShowInformation("Erro", "Dezenas sorteadas", ms.win)
			return
		}
		dzsort := str2sli(strSort)
		if len(dzsort) != 6 {
			dialog.ShowInformation("Erro", "Digite as 6\ndezenas sorteadas", ms.win)
			return
		}
		palpjog := strings.Split(apostas, "\n")
		// acumuladores dos premios
		t6 := 0
		t5 := 0
		t4 := 0
		for _, pj := range palpjog {
			ac6, ac5, ac4 := acertosPalpite(str2sli(pj), dzsort)
			t6 += ac6
			t5 += ac5
			t4 += ac4
		}
		sResult := fmt.Sprintf(
			" Sena(s): %d \n Quina(s): %d \n Quadra(s): %d", t6, t5, t4)
		txtResult.SetText(sResult)
	})
	//btnConf.SetIcon(theme.ConfirmIcon())
	ctConf := container.NewVBox(lblSort, entSort, btnConf, lblConf, txtResult)
	dialog.ShowCustom("Confere Resultado", "Fechar", ctConf, ms.win)
}

func (ms *megasena) makeMenu() {
	menuOpen := fyne.NewMenuItem("Abrir", ms.cmdAbrir)
	menuSave := fyne.NewMenuItem("Salvar", ms.cmdSalvar)
	menuConf := fyne.NewMenuItem("Conferir", ms.cmdConferir)
	menuFile := fyne.NewMenu("Arquivo", menuOpen, menuSave, menuConf)
	mainMenu := fyne.NewMainMenu(menuFile)
	ms.win.SetMainMenu(mainMenu)
}

func (ms *megasena) makeUI(msApp fyne.App) {
	// monta a UI
	// Widgets dos Dados
	ms.lblTam = widget.NewLabel("Dezenas por Cartão:")
    ms.inpTam = widget.NewEntry()
    ms.inpTam.SetPlaceHolder("6 a 20")
	ms.lblInt = widget.NewLabel("Intervalo:")
    ms.inpInt = widget.NewEntry()
	ms.labelc = widget.NewLabel("Dezenas:")
    ms.inputc = widget.NewEntry()
    ms.inputc.SetPlaceHolder("Separadas por espaço")
    // filtros
    ms.labelf = widget.NewLabel("Filtros: ")
    ms.chkCA = widget.NewCheck("C. Aritmética", func(value bool) {})
    ms.chkPI = widget.NewCheck("Par/Impar", func(value bool) {})
    ms.chkMm = widget.NewCheck("Valor Médio", func(value bool) {})
	ms.chkGp = widget.NewCheck("Grupos", func(value bool) {})
    // Widgets de Saida
    ms.labelr = widget.NewLabel("Dica: ")
    ms.inputr = widget.NewEntry()
    ms.lbDicas = widget.NewLabel("Dicas selecionadas: ")
    ms.sugest  = widget.NewMultiLineEntry()
    // Widgets dos Comandos
    ms.btGerar = widget.NewButton("Gerar Dica", ms.cmdGerar)
    ms.btAdic  = widget.NewButton("Selecionar", ms.cmdAdic)
    ms.btLimp  = widget.NewButton("Limpar", ms.cmdLimp)
	// layout
	ctInt := container.NewHBox(ms.lblTam, ms.inpTam, ms.lblInt, ms.inpInt)
    ctDez := container.New(layout.NewFormLayout(),
               ms.labelc, ms.inputc)
    ctFil := container.New(layout.NewGridLayout(2),
               ms.labelf, layout.NewSpacer(), ms.chkCA, ms.chkPI, 
               ms.chkMm, ms.chkGp)
    ctDic := container.New(layout.NewFormLayout(),
               ms.labelr, ms.inputr)
    ctBtn := container.New(layout.NewHBoxLayout(), layout.NewSpacer(),
               ms.btGerar, ms.btAdic, ms.btLimp, layout.NewSpacer())
    ctSai := container.New(layout.NewVBoxLayout(),
               ms.lbDicas, ms.sugest)
	// monta o conteudo
    content := container.New(layout.NewVBoxLayout(),
               ctInt, ctDez, ctFil, ctDic, ctSai, ctBtn)
	// janela principal
	ms.win = msApp.NewWindow("Mega-sena")
	ms.makeMenu()
    ms.win.SetContent(content)
    ms.win.Show()
}

func (ms *megasena) aplicaFiltros(palp []int) bool {
	// retorna true se palpite nao passar nos filtros
	// retorna false em caso contrario
	var bFiltrar bool = false
	var bAC, bPI, bMm, bGp bool = false, false, false, false
	
	if ms.chkCA.Checked == true {
		// calcula complexidade aritmetica
		ac := complexAritmetica(palp)
		bAC = filtraAC(ac,3) // lim = 3
	}
	// par / impar
	if ms.chkPI.Checked == true {
		bPI = filtraParImpar(palp)
	}
	// acima / abaixo valor medio
	if ms.chkMm.Checked == true {
		bMm = filtraMetadeInfSup(palp)
	}
	// dezenas por grupo
	if ms.chkGp.Checked == true {
		bGp = filtraGrupos(palp)
	}
	// se qq dos filtros solicitados falhar
	if bAC || bPI || bMm || bGp {
		bFiltrar = true
	}
	return bFiltrar
}

func (ms *megasena) geraPalpite(palp []int, j, min, max int) {
	// gera a dica ao acaso, escolhendo dezenas do intervalo min-max
	var i, k, nrTentativas int
	
	nrTentativas = 0
	for nrTentativas < maxRep { // const maxRep = 1000
		i = 0
		// loop geração dica
		for i < j {
			k = intRange(min, max) // escolhe a dezena
			if existeInt(palp, k) == false {
				palp[i] = k
				i = i + 1
			}
		}
		// verifica os filtros ativos
		if ms.chkCA.Checked || ms.chkPI.Checked || 
			ms.chkMm.Checked || ms.chkGp.Checked {
			sort.Ints(palp)
			bTemp := ms.aplicaFiltros(palp)
			if  bTemp { // nao passou nos filtros, gera outra dica
				nrTentativas++
				continue
			} else { // passou nos filtros
				break
			}
		} else { // nenhum filtro ativo
			break
		}
	} // loop nrTentativas
	if nrTentativas >= maxRep {
		dialog.ShowInformation("Erro", "Falha ao gerar a dica", ms.win)
		// resetar dezenas[0] para 0
		// ou retornar um valor false ??
		palp[0] = 0
	}
	return
}

func (ms *megasena) geraPalpiteDez(palp []int, dezs []int, j, min, max int) {
	// gera dica usando as dezenas digitadas como fonte
	var i, k, nrTentativas int
	nrTentativas = 0
	for nrTentativas < maxRep { // const maxRep = 1000
		i = 0
		for i < j {
			k = intRange(min, max) // escolhe o indice da dezena
			if existeInt(palp, dezs[k-1]) == false {
				palp[i] = dezs[k-1]
				i = i + 1
			}
		}
		// aplica os filtros ativos
		if ms.chkCA.Checked || ms.chkPI.Checked || 
			ms.chkMm.Checked || ms.chkGp.Checked {
			sort.Ints(palp)
			bTemp := ms.aplicaFiltros(palp)
			if  bTemp { // nao passou nos filtros, gera outra dica
				nrTentativas++
				continue
			} else {	// passou nos filtros
				break
			}
		} else { // nenhum filtro ativo
			break
		}
	} // loop nrTentativas
	if nrTentativas >= maxRep {
		dialog.ShowInformation("Erro", "Falha ao gerar a dica", ms.win)
		// resetar dezenas[0] para 0
		// ou retornar um valor false ??
		palp[0] = 0
	}
}

func (ms *megasena) geraPalpiteDezFixa(palp []int, dezs []int, j, min, max int) {
	// gera palpite usando dezenas fixas, NAO usar filtros nestes casos
	var k int
	copy(palp, dezs)
	i := len(dezs)
	for i < j {
		k = intRange(min, max) // sorteia uma dezena
		if existeInt(palp, k) == false {
			palp[i] = k
			i = i + 1
		}
	}
	if ms.chkCA.Checked || ms.chkPI.Checked || 
		ms.chkMm.Checked || ms.chkGp.Checked {
		// nao filtrar
		dialog.ShowInformation("Atenção", "Filtros não aplicados", ms.win)
	}
}

func acertosPalpite(palp []int, sorteio []int) (int, int, int) {
	var ac6, ac5, ac4 int = 0, 0, 0
	var tacertos int
	n := len(palp)
	tacertos = contaRep(palp, sorteio)
	if tacertos == 6 { 			// sena
		ac6 = acertos(n,6,6,6)
		ac5 = acertos(n,6,6,5)
		ac4 = acertos(n,6,6,4)
	} else if tacertos == 5 { 	// quina
		ac5 = acertos(n,6,5,5)
		ac4 = acertos(n,6,5,4)
	} else if tacertos == 4 { 	// quadra
		ac4 = acertos(n,6,4,4)
	}
	return ac6, ac5, ac4
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

func intRange(min, max int) int {
	// retorna um nr pseudo aleatorio no intervalo min-max (inclusive)
	return rand.Intn(max - min + 1) + min
}

func existeInt(s []int, k int) bool {
	// verifica se um valor k ja existe no vetor (slice)
	b := false
	for _, v := range s {
		if v == k {
			b = true
			break
		}
	}
	return b
}

func obtInter (s string, min, max *int) bool {
	// obtem o intervalo min-max digitado
	bret := true
	bEncont, _ := regexp.MatchString(`[^0-9 ]`, s)
	if bEncont == true {
		bret = false
	} else {
		interv := strings.Fields(s)
		if len(interv) == 2 {
			*min, _ = strconv.Atoi(interv[0])
			*max, _ = strconv.Atoi(interv[1])
		} else {
			bret = false
		}
	}
	return bret
}
	
func obtDezCartao (s string, dpc *int) bool {
	// obtem o nr de dezenas por cartao
	bret := true
	bEncont, _ := regexp.MatchString(`[^0-9 ]`, s)
	if bEncont == true {
		bret = false
	} else {
		dezpc := strings.Fields(s)
		if len(dezpc) == 1 {
			*dpc, _ = strconv.Atoi(dezpc[0])
		} else {
			bret = false
		}
	}
	return bret
}

func sli2str(sint []int) string {
	// converte []int para string
	var sstr string
	sstr = ""
	for _, v := range sint {
		sstr = sstr + fmt.Sprintf("%02d ", v)
	}
	return sstr
}

func str2sli(s string) []int {
	// converte uma string (dezenas separadas por espaco) para []int
	var err error
	slstr := strings.Fields(s)
	slint := make([]int, len(slstr))
	for i, v := range slstr {
		slint[i], err = strconv.Atoi(v)
		if err == nil {
			// panic
		}
	}
	return slint
}

func filtraMetadeInfSup(palpite []int) bool {
	// - filtro para evitar que as 6 dezenas de uma aposta da
	// megasena estejam todas abaixo ou acima do valor medio.
	// - Nao usar em apostas cercando mais de 10 dezenas, pois 6 ou 
	// mais dezenas estarão abaixo ou acima do valor medio.
	// retorna true se nao passar, se passar no teste retorna false
	var minf, msup, meio, lim int
	minf = 0
	msup = 0
	meio = 30 // valor medio
	lim = 6 // maximo de 5 dezenas abaixo ou acima do valor medio (30)
	ret := true
	if len(palpite) > 10 { // nao usar em palpites com + de 10 dezenas
		return false
	}
	for _, v := range(palpite) {
		if v < meio {
			minf += 1
		} else {
			msup += 1
		}
	}
	if minf < lim && msup < lim {
		ret = false
	}
	return ret
}

func filtraParImpar(palpite []int) bool {
	// - filtro para evitar que as 6 dezenas de uma aposta da
	// megasena sejam todas pares ou impares.
	// - Nao usar em apostas cercando mais de 10 dezenas,
	// pois 6 ou mais destas dezenas serão pares / impares.
	// retorna true se nao passar, se passar no teste retorna false
	var mpar, mimp, lim int
	mpar = 0
	mimp = 0
	lim = 6 // maximo de 5 dezenas pares ou impares
	ret := true
	if len(palpite) > 10 { // nao usar em palpites com + de 10 dezenas
		return false
	}
	for _, v := range(palpite) {
		if v % 2 == 0 {
			mpar += 1
		} else {
			mimp += 1
		}
	}
	if mpar < lim && mimp < lim {
		ret = false
	}
	return ret
}

func filtraGrupos(palpite []int) bool {
	// - filtro para evitar que 4 dezenas de uma aposta da
	// megasena sejam todas de um mesmo grupo
	// - Nao usar em apostas cercando mais de 12 (15) dezenas
	// retorna true se nao passar, se passar no teste retorna false
	// Dezenas agrupadas:
	// Grupo01: dezenas de 01-09
	// Grupo02: dezenas de 10-19
	// Grupo03: dezenas de 20-29
	// Grupo04: dezenas de 30-39
	// Grupo05: dezenas de 40-49
	// Grupo06: dezenas de 50-59
	var grp, lim int
	lim = 3 // maximo de 3 dezenas num grupo
	grp = 6 // grupo 01-09 / 10-19 / 20-29 / 30-39 / 40-49 / 50-59
	ret := false
	grupos := make([]int, grp)
	if len(palpite) > 12 { // nao usar em palpites com + de 12 dezenas
		return false
	}
	for _, v := range(palpite) {
		if v < 10 {
			grupos[0] += 1
		} else if v < 20 {
			grupos[1] += 1
		} else if v < 30 {
			grupos[2] += 1
		} else if v < 40 {
			grupos[3] += 1
		} else if v < 50 {
			grupos[4] += 1
		} else if v < 60 { // 60 esta fora
			grupos[5] += 1
		}
	}
	for _, vg := range (grupos) {
		if vg > lim {
			ret = true
			break
		}
	}
	return ret
}

func complexAritmetica(palpite []int) int {
	var j, tmp int
	r := len(palpite)
	tt := nrcomb(r, 2)
	d := make([]int, tt)
	i := r - 1 // indice inicial = 0
	t := 0 // nr diferencas positivas
	difpos := 0
	for i > 0 {
		j = i - 1
		for j >= 0 {
			tmp = palpite[i] - palpite[j]
			if existeInt(d, tmp) == false {
				d[t] = tmp
				t = t + 1
				difpos = difpos + 1
			}
			j = j - 1
		}
		i = i - 1
	}
	ac := difpos - r + 1
	return ac
}

func filtraAC (ac, lim int) bool {
	if ac > lim {
		return false
	} else {
		return true
	}
}

/*
 * How to Win More, by Norbert Henze & Hans Riedwyl, CRC Press 1998
 * Soma dos numeros da combinacao aleatoria numa loteria r/s:
 * sum ~= (r*(s + 1)/2)
 * p.ex.: para uma loteria 6/49 sum = 150, entao uma aposta
 *        deve ter AC > 5 e soma > 173 (?)
*/
func filtraSoma(dez []int, r int, s int) bool {
	soma := r * (s + 1) / 2
	somaDez := 0
	for _, v := range(dez) {
		somaDez += v
	}
	//fmt.Println("Soma Num: ", soma, "Soma Dez: ", somaDez)
	if somaDez > soma {
		return false
	} else {
		return true
	}
}

func confereDica(strDicas string, strdica string, lim int) bool {
	// verifica quantas dezenas de um palpite existem nos
	// palpites ja selecionados, se o valor for maior do 
	// que limite (lim) retorna false, senao retorna true
	var bret bool
	var sli []int
	var rep int
	bret = true
	cjDicas := strings.Split(strDicas, "\n")
	dica := str2sli(strdica)
	for _, ss := range cjDicas {
		sli = str2sli(ss)
		rep = contaRep(sli, dica)
		if rep >= lim {
			bret = false
			break
		}
	}
	return bret
}

func contaRep(v1 []int, v2 []int) int {
	// retorna o nr de dezenas de []v2 repetidas em []v1
	var irep, e1, e2 int
	irep = 0
	for _, e1 = range v1 {
		for _, e2 = range v2 {
			if e2 == e1 {
				irep = irep + 1
				break
			}
		}
	}
	return irep
}

func main() {
	// cria app
	msApp := app.New()
	msApp.SetIcon(resourceIconPng)
    // alimenta o gerador pseudo-randomico
	rand.Seed(time.Now().UnixNano())
	
	ms := &megasena{}
	ms.makeUI(msApp)
	msApp.Run()
}
