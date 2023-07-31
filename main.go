package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
)

func main() {
	a := app.New()
	w := a.NewWindow("Songsterr loader")
	w.Resize(fyne.Size{Width: 600, Height: 127})

	label := widget.NewLabel("Get Guitar Pro file from Songsterr by link")
	label.Alignment = fyne.TextAlignCenter

	input := widget.NewEntry()
	input.TextStyle.Bold = true
	input.SetPlaceHolder("Enter the song link")
	input.Resize(fyne.NewSize(w.Canvas().Size().Width*0.9, w.Canvas().Size().Height*0.1))

	submitBtn := widget.NewButton("Save", SaveTabs(input))
	submitBtn.Resize(fyne.NewSize(50, 50))

	w.SetContent(container.NewVBox(
		label,
		input,
		submitBtn,
	))

	w.ShowAndRun()
}

func SaveTabs(inputBar *widget.Entry) func() {
	return func() {
		songsterrSongUrl := inputBar.Text
		resp, err := http.Get(songsterrSongUrl)
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			panic(err)
		}

		scripts := doc.Find("script").Text()

		tabJsn := scripts[strings.Index(scripts, "(document,window);")+len("(document,window);") : strings.Index(scripts, "\n")]

		revId := gjson.Get(tabJsn, "meta.current.revisionId").Int()

		url := fmt.Sprintf("https://www.songsterr.com/a/ra/player/songrevision/%d.json", revId)

		resp, err = http.Get(url)
		if err != nil {
			panic(err)
		}

		res, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		infoJsn := string(res)
		gpFile := gjson.Get(infoJsn, "tab.guitarProTab.attachmentUrl").String()
		songInfo := []string{gjson.Get(infoJsn, "song.artist.name").String(), gjson.Get(infoJsn, "song.title").String()}

		if err := DownloadFile(fmt.Sprintf("%s - %s.gp", songInfo[0], songInfo[1]), gpFile); err != nil {
			panic(err)
		}
	}
}

func DownloadFile(filename string, url string) error {
	resp, err := http.Get(url)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
