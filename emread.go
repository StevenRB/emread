/*
This program is meant to address the issue when someone sends you a
.eml file as an attachment. It will process it into an html file
and load it in your default browser.

Need to organize the switch case so that I can determine what OS
is being used to accomodate Windows use of `\` as a directory switch
*/

package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const versionNum = "0.2.1"

func openBrowser(url string) bool {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start() == nil
}

func main() {
	var fileName string

	helpString := `
Emread 2018 ver %s
Converts .eml to .html

usage:  emread [options] <inputFilename>

options:
--help           displays this dialogue
--v              show version information
--o "filename"   specifies the output filename
--s              suppresses the automatic browser launch
--d              deletes the html file once launched
--t <seconds>    time to delay deletion (helps ensure browser launch)

https://github.com/StevenRB/emread/ 
`

	// Sets up various flags
	helpFlg := flag.Bool("help", false, "displays help")
	hFlg := flag.Bool("h", false, "displays help")
	noBrowse := flag.Bool("s", false, "Suppresses the automatic browser launch")
	delFile := flag.Bool("d", false, "Delete the .html file after loading")
	output := flag.String("o", "blank", "Output filename")
	verFlg := flag.Bool("v", false, "Displays version number")
	timeFlg := flag.Int("t", 2, "Time in seconds to delay deletion")
	flag.Parse()

	// Displays version or helpString if requested
	if *verFlg == true {
		fmt.Println("Emread version", versionNum)
		os.Exit(0)
	}
	if *helpFlg == true || *hFlg == true {
		fmt.Printf(helpString, versionNum)
		os.Exit(0)
	}

	// If emread is run without an input file, this will display the help string
	if flag.NArg() < 1 {
		fmt.Printf(helpString, versionNum)
		os.Exit(0)
	} else {
		fileName = flag.Arg(0)
	}

	// Splits filename, grabs cwd, and creates filename with absolute path
	var name string
	pwd, _ := os.Getwd()
	q := string(filepath.Base(fileName))
	t := strings.Split(q, ".")
	if *output != "blank" {
		name = string(*output)
	} else {
		name = string(t[0]) + ".html"
	}
	if t[1] != "eml" {
		fmt.Println("You must specify an .eml file for processing")
		os.Exit(1)
	}
	newFile := string(pwd + "/" + name)

	// Read in the contents of the .eml
	rawData, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("There was an error reading the file:", err)
		os.Exit(1)
	}

	// Extract the base64. The metadata is extracted,
	// rejoined, and html tags added for readability
	temp := strings.Split(string(rawData), "\n\n")
	payload := temp[2]
	a := strings.Split(temp[0], ";")
	b := a[1 : len(a)-1]
	meta := strings.Join(b, "<br>")
	meta = fmt.Sprintf("<html>%s</html><br><br>", meta)

	decode, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		fmt.Println("There was an unrecoverable error decoding the email:", err)
		os.Exit(1)
	}

	// Creates new file
	fo, err := os.Create(newFile)
	if err != nil {
		fmt.Println("Error creating new file:", err)
		os.Exit(1)
	}

	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	// write to new file
	if _, err := fo.Write([]byte(meta)); err != nil {
		fmt.Println("Error writing to new file:", err)
		os.Exit(1)
	}

	if _, err := fo.Write(decode); err != nil {
		fmt.Println("Error writing to new file:", err)
		os.Exit(1)
	}

	// Controls whether to launch the browser or not

	if *noBrowse == false {
		openBrowser(newFile)
	}

	// The sleep is because launching the browser takes a moment
	// If you don't sleep, it will delete before it even launches
	if *delFile == true {
		time.Sleep(time.Duration(*timeFlg) * time.Second)
		os.Remove(newFile)
		fmt.Println("Success!")
		os.Exit(0)
	} else {
		fmt.Printf("Success! Email contents written to %s.html\n", t[0])
		os.Exit(0)
	}
}
