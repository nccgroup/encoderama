package main

import (
	"bufio"
	"flag"
	"fmt"
	"html"
	"log"
	"net/url"
	"os"
	"strings"
	"unicode"
)

func main() {

	incrementalPtr := flag.Bool("i", false, "For each input string, incrementally encode each character and print out the string")
	stdinPtr := flag.Bool("stdin", false, "Read input strings from stdin")
	inputFilePtr := flag.String("f", "", "File to read input strings from")

	flag.Parse()

	if flag.NArg() == 0 && *stdinPtr == false && *inputFilePtr == "" {
		flag.Usage()
		os.Exit(1)
	}

	// If we got this far it means that the target string has been passed at
	// the command line, if stdin or inputfile have also been passed, throw an
	// error
	fmt.Printf("%d\n", flag.NArg())
	if *stdinPtr || *inputFilePtr != "" {
		log.Fatal("Pass a string either by: an single string on the command line, from an input file with -f or via stdin with -stdin")
	}

	// The final arg should be a single input string if stdin or file input
	// isn't used
	input := flag.Arg(flag.NArg() - 1)

	//TODO ensure unicode support
	if isASCII(input) {
		fmt.Println("Input string contains non-ascii chars... issues may appear")
	}

	targetList := make([]string, 0)

	// If providing an input string
	if *inputFilePtr == "" && !*stdinPtr {
		if *incrementalPtr {
			targetList = incrementalStringGenerator(input)
		} else {
			targetList = append(targetList, input)
		}
	}

	// If reading from a file or stdin get the candidates into memory

	tmpTargetList := make([]string, 0)
	if *stdinPtr {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
			tmpTargetList = append(tmpTargetList, scanner.Text())
		}
	}

	if *inputFilePtr != "" {
		file, err := os.Open(*inputFilePtr)

		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
			tmpTargetList = append(tmpTargetList, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

	// TODO must be a more efficient way of doing this
	if *incrementalPtr {
		for i := 0; i < len(tmpTargetList); i++ {
			tmpList := incrementalStringGenerator(tmpTargetList[i])
			for l := 0; l < len(tmpList); l++ {
				targetList = append(targetList, tmpList[i])
			}
		}
	} else {
		targetList = tmpTargetList
	}

	for i := 0; i < len(targetList); i++ {

		//TODO try to deduplicate?
		// URL encoding:
		// URL Encode - lowercase
		fmt.Printf("%s\n", urlEncodeAll(targetList[i], true, false))

		// URL Encode - uppercase
		fmt.Printf("%s\n", urlEncodeAll(targetList[i], false, false))

		// Double URL Encode - lowercase
		fmt.Printf("%s\n", urlEncodeAll(targetList[i], true, true))

		// Double URL Encode - uppercase
		fmt.Printf("%s\n", urlEncodeAll(targetList[i], false, true))

		// URL Encode Required chars only
		fmt.Printf("%s\n", urlEncodeRequired(targetList[i], false))

		// Double URL Encode Required chars only
		fmt.Printf("%s\n", urlEncodeRequired(targetList[i], true))

		// HTML Encoding:

		// HTML encode characters in decimal format
		fmt.Printf("%s\n", htmlEncodeAllAsDecimal(targetList[i], false))

		// Double HTML characters in decimal format
		//fmt.Printf("%s\n", htmlEncodeAllAsDecimal(targetList[i], true))

		// HTML encode characters in hex format - lowercase
		fmt.Printf("%s\n", htmlEncodeAllAsHex(targetList[i], true, false))

		// HTML encode characters in hex format - uppercase
		fmt.Printf("%s\n", htmlEncodeAllAsHex(targetList[i], false, false))

		// Double HTML characters in hex format - lowercase
		//fmt.Printf("%s\n", htmlEncodeAllAsHex(targetList[i], true, true))

		// Double HTML characters in hex format - uppercase
		//fmt.Printf("%s\n", htmlEncodeAllAsHex(targetList[i], false, true))

		// HTML encode required characters only
		fmt.Printf("%s\n", htmlEncodeRequired(targetList[i], false))

		// Double HTML encode required characters only
		//fmt.Printf("%s\n", htmlEncodeRequired(targetList[i], true))
	}
}

func htmlEncodeAllAsDecimal(input string, doubleHtmlEncode bool) string {
	outputString := ""

	r := []rune(input)
	for i := 0; i < len(r); i++ {
		outputString = outputString + fmt.Sprintf("&#%d;", r[i])
		if doubleHtmlEncode {
			outputString = htmlEncodeAllAsDecimal(outputString, false)
		}
	}
	return outputString
}

func htmlEncodeAllAsHex(input string, lowercase bool, doubleHtmlEncode bool) string {
	outputString := ""

	r := []rune(input)
	for i := 0; i < len(r); i++ {
		outputString = outputString + fmt.Sprintf("&#x%X;", r[i])
		if doubleHtmlEncode {
			outputString = htmlEncodeAllAsHex(outputString, lowercase, false)
		}
	}
	if lowercase {
		return strings.ToLower(outputString)
	}
	return outputString
}

func htmlEncodeRequired(input string, doubleHtmlEncode bool) string {
	outputString := html.EscapeString(input)

	if doubleHtmlEncode {
		outputString = html.EscapeString(outputString)
	}

	return outputString
}

func urlEncodeAll(input string, lowercase bool, doubleUrlEncode bool) string {
	outputString := ""

	r := []rune(input)
	for i := 0; i < len(r); i++ {
		if doubleUrlEncode {
			outputString = outputString + fmt.Sprintf("%%25%%%X", r[i])
		} else {
			outputString = outputString + fmt.Sprintf("%%%X", r[i])
		}
	}
	if lowercase {
		return strings.ToLower(outputString)
	}
	return outputString
}

func urlEncodeRequired(input string, doubleUrlEncode bool) string {

	outputString := url.QueryEscape(input)

	if doubleUrlEncode {
		outputString = url.QueryEscape(outputString)
	}

	return outputString
}

//TODO modify this to work on runes rather than splitting the string
func incrementalStringGenerator(input string) []string {

	incrementalStringList := make([]string, 0)
	// Build the string up one character at a time
	for i := 0; i < len(input); i++ {
		incrementalStringList = append(incrementalStringList, input[0:i+1])
	}
	return incrementalStringList
}

func isASCII(input string) bool {
	for i := 0; i < len(input); i++ {
		if input[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
