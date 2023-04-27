//go:generate /home/nadsec/go/bin/fyne bundle -o tbicons.go ../img/Comb.png
//go:generate /home/nadsec/go/bin/fyne bundle -append -o tbicons.go ../img/Perm.png
//go:generate /home/nadsec/go/bin/fyne bundle -append -o tbicons.go ../img/Icon.png
// calcomb.go - calcula nCr ou nPr, dados n e r
// Fyne GUI toolkit
// OBS: Nenhum controle de erro !
//
package main

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
)

type ccombui struct {
	// toolbar
	toolBar *widget.Toolbar
	// widgets
	labeln *widget.Label
    inputn *widget.Entry
    labelr *widget.Label
    inputr *widget.Entry
    labres *widget.Label
    inpres *widget.Entry
    btcomb *widget.Button
    btperm *widget.Button
    window fyne.Window
}

func (c *ccombui) calcComb() {
    nn, _ := strconv.ParseInt(c.inputn.Text, 0, 64)
    rr, _ := strconv.ParseInt(c.inputr.Text, 0, 64)
    if nn < rr || nn <= 0 || rr <= 1 {
		dialog.ShowInformation("Erro", "Valores N ou R", c.window)
		return
	}
    ncr := nrcomb(nn, rr)
    c.labres.Text = "nCr"
    c.inpres.Text = strconv.FormatInt(ncr, 10)
    c.labres.Refresh()
    c.inpres.Refresh()
}

func (c *ccombui) calcPerm() {
    nn, _ := strconv.ParseInt(c.inputn.Text, 0, 64)
    rr, _ := strconv.ParseInt(c.inputr.Text, 0, 64)
    if nn < rr || nn <= 0 || rr <= 0 {
		dialog.ShowInformation("Erro", "Valores N ou R", c.window)
		return
	}
    npr := nrperm(nn, rr)
    c.labres.Text = "nPr"
    c.inpres.Text = strconv.FormatInt(npr, 10)
    c.labres.Refresh()
    c.inpres.Refresh()
}

/*
func (c *ccombui) makeMenu(win fyne.Window) {
	menuCalc := fyne.NewMenuItem("Calcular", c.calcComb)
	menuFile := fyne.NewMenu("File", menuCalc)
	mainMenu := fyne.NewMainMenu(menuFile)
	win.SetMainMenu(mainMenu)
}
*/

func (c *ccombui) makeTb() {
    c.toolBar = widget.NewToolbar(
            // usa ícones estáticos e do pkg theme
            widget.NewToolbarAction(resourceCombPng, c.calcComb),
            widget.NewToolbarAction(resourcePermPng, c.calcPerm),
            widget.NewToolbarSeparator(),
            widget.NewToolbarAction(theme.HelpIcon(), func() {
                dialog.ShowInformation("Ajuda", "Digite os valores N e R\n e comande nCr ou nPr", c.window)
            }),
    )
}

func (c *ccombui) makeUI(app fyne.App) {
	c.labeln = widget.NewLabel("Digite N:")
    c.inputn = widget.NewEntry()
    c.labelr = widget.NewLabel("Digite R:")
    c.inputr = widget.NewEntry()
    c.labres = widget.NewLabel("n C r:")
    c.inpres = widget.NewEntry()
    c.btcomb = widget.NewButton("nCr", c.calcComb)
    c.btperm = widget.NewButton("nPr", c.calcPerm)

    ctd := container.New(
            layout.NewFormLayout(),
            c.labeln, c.inputn, 
            c.labelr, c.inputr, 
            c.labres, c.inpres)
    ctc := container.NewHBox(layout.NewSpacer(), c.btcomb, c.btperm, layout.NewSpacer() )
    
    c.window = app.NewWindow("Combinatória")
    c.makeTb() // monta a toolbar
    c.window.SetContent(container.NewVBox(c.toolBar, ctd, ctc))
    c.window.Show()
}

func nrcomb(n, m int64) int64 {
    var i int64
    cnm := int64(1)

    if m * 2 > n {
        m = n - m
    }
    for i = 1; i < m; n, i = n-1, i+1 {
        cnm = cnm * n / i
    }
    return cnm
}

func nrperm(n, m int64) int64 {
    var i int64
    npr := n

	for i = 1; i < m; i++ {
		npr = npr * (n-i)
	}
    return npr
}

func main() {
	myApp := app.New()
	myApp.SetIcon(resourceIconPng)
	c := &ccombui{}
	c.makeUI(myApp)
	myApp.Run()
}
