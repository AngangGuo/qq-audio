package main

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const usage = `
This program will download audio from https://mp.weixin.qq.com/mp

Usage:
1. Open the page with the audio from web browser
2. Copy the link of the page from address bar
3. Paste the link into link.txt and save
4. Run this program to download the audio

`

func main() {
	linkfile := "link.txt"
	var pageURL string
	content, err := ioutil.ReadFile(linkfile)
	if err != nil {
		if err == os.ErrExist {
			fmt.Printf("can't find %s \n", linkfile)
			return
		}
		fmt.Printf("read %s failed\n", linkfile)
		fmt.Print(usage)
		fmt.Println(err)
		return
	}

	pageURL = string(content)
	if !strings.HasPrefix(pageURL, "https://mp.weixin.qq.com/mp/") {
		fmt.Print(usage)
		return
	}

	fmt.Println("Processing...")
	url := launcher.New().
		Headless(false).
		Launch()

	browser := rod.New().ControlURL(url).Connect()
	defer browser.Close()

	maxLen := 60
	if len(pageURL) < maxLen {
		maxLen = len(pageURL)
	}
	fmt.Printf("Go to page %s ...\n", pageURL[:maxLen-1])

	// pageURL := `https://mp.weixin.qq.com/mp/audio?t=pages/audio_detail&scene=1&__biz=MzUxOTEwOTk1OA==&mid=2247484982&idx=1&sn=b8741c3d24ad335f3b6c5a04f175985b&voice_id=MzUxOTEwOTk1OF8yMjQ3NDg0OTc5&_wxindex_=0&uin=MTMyMDk2NDQ0Ng%3D%3D&key=0acde1ff6de13ef41bbc791a74bee0b671ddc512e76725a64df554c17af92fdc7d253de9fed1ac31e3443d1989efa41c5f8857b509d203c843b66cbbb05a933e171c03e3037c7efd22138dd3e5b8cc8e&devicetype=Windows+10+x64&version=62090529&lang=en&ascene=1&pass_ticket=dbjbGtV6MnpzxVPqIOUBOnNbYh34DilvS0pgaeO71sJk%2Frr67n9bv6AJYTkCFaIw`
	page := browser.Page(pageURL).WindowMinimize()
	wait := page.WaitRequestIdle()
	wait()

	fmt.Println("\nChecking the audio file...")

	// play
	page.Timeout(30 * time.Second).Element("#voice_play > em").Click()
	// time.Sleep(7* time.Second)
	audioTitle := strings.TrimSpace(page.Element("#voice_frame > strong").Text())
	audioURL := page.Timeout(30 * time.Second).Element("body > audio").Property("src").String()
	fmt.Println("Audio URL:\n", audioURL)

	// stop
	page.Element("#voice_play > em").Click()

	audioName := audioTitle + ".mp3"
	if _, err := os.Stat(audioName); err == nil {
		rand := strconv.FormatInt(time.Now().UnixNano(), 16)
		audioName = fmt.Sprintf("%s - %s.mp3", audioTitle, rand)
	}
	fmt.Print("\nStart downloading to file:\n" + audioName + " ")
	f, err := os.Create(audioName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	response, err := http.Get(audioURL)
	if err != nil {
		fmt.Printf("Connecting to %s failed. Please try again.\n", audioURL)
		// fmt.Println(err)
		return
	}
	defer response.Body.Close()

	start := time.Now()

	// progress indicator
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-time.After(time.Second):
				fmt.Print(".")
			}
		}
	}()

	// copy audio to file
	_, err = io.Copy(f, response.Body)
	if err != nil {
		fmt.Printf("Downloading failed. Please try again later.\n")
		// fmt.Println(err)
		return
	}

	// exit progress indicator
	done <- true

	fmt.Printf("\n\nTotal time: %v s\n", int(time.Now().Sub(start).Seconds()))
	fmt.Println("Done.\n")
}
