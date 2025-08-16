// go Colly example
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/zshanhui/gejiezhipin/cli"
	gejie "github.com/zshanhui/gejiezhipin/gejielib"
)

// const zhipinJobDetailPageEx = "https://www.zhipin.com/job_detail/2d635ccda89345b01HF72Nm1EVpZ.html"
const ClassName_jobDescription = "#main > div.job-box > div > div.job-detail > div:nth-child(1) > div.job-sec-text"

// func handleMsakuraCommand() {
// 	if len(os.Args) < 3 {
// 		fmt.Println("Usage: --msakura <target_url> [selector]")
// 		return
// 	}

// 	target := os.Args[2]
// 	selector := "div.plyr__video-wrapper > video"
// 	if len(os.Args) > 3 {
// 		selector = os.Args[3]
// 	}
// 	if err := sakuram.DownloadVideoFromUrl(target, selector, false); err != nil {
// 		fmt.Printf("msakura error: %v\n", err)
// 		os.Exit(1)
// 	}
// }

func main() {
	cli.Execute()

	if len(os.Args) > 1 && os.Args[1] == "--zhipin" {
		gejie.RunZhipin("https://www.zhipin.com/job_detail/b6840d4438ff55c41n1609S-FFVT.html", true)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "--meli" {
		maxItems := 10 // default value
		// Check for --max-items flag
		for i, arg := range os.Args {
			if arg == "--max-items" && i+1 < len(os.Args) {
				if parsed, err := strconv.Atoi(os.Args[i+1]); err == nil {
					maxItems = parsed
				} else {
					fmt.Printf("Invalid maxItems value: %s. Using default value 10.\n", os.Args[i+1])
				}
				break
			}
		}
		gejie.RunMeliSearch(nil, int8(maxItems))
		return
	}
}
