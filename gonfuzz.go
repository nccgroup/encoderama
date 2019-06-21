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
)

// Encoding scheme definition - this makes it a little easier to add new ones
type encodingScheme struct {
	params    []string
	desc      string
	isDefault bool
	enabled   bool
	method    func(string) string
}

func main() {

	// Define encoding schemes
	var encodingSchemes = []encodingScheme{
		encodingScheme{
			params:    []string{"p", "plain"},
			desc:      "No encoding: '&' --> '&'",
			isDefault: true,
			enabled:   false,
			method:    plainText,
		},
		encodingScheme{
			params:    []string{"uau", "urlallupper"},
			desc:      "URL encode and force upper case: '/' --> '%2F'",
			isDefault: false,
			enabled:   false,
			method:    urlEncodeAllUppercase,
		},
		encodingScheme{
			params:    []string{"ual", "urlalllower"},
			desc:      "No encode and force lower case: '/' --> '%2f'",
			isDefault: false,
			enabled:   false,
			method:    urlEncodeAllLowercase,
		},
		encodingScheme{
			params:    []string{"dual", "doubleurlalllower"},
			desc:      "Double URL encode and force lower case: '/' --> '%25%2f'",
			isDefault: false,
			enabled:   false,
			method:    doubleUrlEncodeAllLowercase,
		},
		encodingScheme{
			params:    []string{"duau", "doubleurlallupper"},
			desc:      "Double URL encode and force upper case: '/' --> '%25%2F'",
			isDefault: false,
			enabled:   false,
			method:    doubleUrlEncodeAllUppercase,
		},
		encodingScheme{
			params:    []string{"u", "url"},
			desc:      "URL encode key characters: ' ' --> '+'",
			isDefault: true,
			enabled:   false,
			method:    urlEncodeSpecialOnly,
		},
		encodingScheme{
			params:    []string{"du", "doubleurl"},
			desc:      "Double URL encode key characters: '+' --> '%2B'",
			isDefault: false,
			enabled:   false,
			method:    doubleUrlEncodeSpecialOnly,
		},
		encodingScheme{
			params:    []string{"had", "htmlalldec"},
			desc:      "HTML encode all characters with decimal notation: '&' --> '&#38;'",
			isDefault: false,
			enabled:   false,
			method:    htmlEncodeAllAsDecimal,
		},
		encodingScheme{
			params:    []string{"dhad", "doublehtmlalldec"},
			desc:      "Double HTML encode all characters with decimal notation: '&' --> '&#38;&#35;&#51;&#56;&#59;'",
			isDefault: false,
			enabled:   false,
			method:    doubleHtmlEncodeAllAsDecimal,
		},
		encodingScheme{
			params:    []string{"hahl", "htmlallhexlower"},
			desc:      "HTML encode all characters with hex notation, force lower case: '+' --> '&#x2b;'",
			isDefault: false,
			enabled:   false,
			method:    htmlEncodeAllAsHexLowercase,
		},
		encodingScheme{
			params:    []string{"hahu", "htmlallhexupper"},
			desc:      "HTML encode all characters with hex notation, force upper case: '+' --> '&#x2B;'",
			isDefault: false,
			enabled:   false,
			method:    htmlEncodeAllAsHexUppercase,
		},
		encodingScheme{
			params:    []string{"dhahl", "doublehtmlallhexlower"},
			desc:      "Double HTML encode all characters with hex notation, force lower case: '+' --> '&#x26;&#x23;&#x78;&#x32;&#x42;&#x3b;'",
			isDefault: false,
			enabled:   false,
			method:    doubleHtmlEncodeAllAsHexLowercase,
		},
		encodingScheme{
			params:    []string{"dhahu", "doublehtmlallhexupper"},
			desc:      "Double HTML encode all characters with hex notation, force upper case: '+' --> '&#x26;&#x23;&#x78;&#x32;&#x42;&#x3B;'",
			isDefault: false,
			enabled:   false,
			method:    doubleHtmlEncodeAllAsHexUppercase,
		},
		encodingScheme{
			params:    []string{"h", "html"},
			desc:      "HTML encode key characters: '&' --> '&amp;'",
			isDefault: true,
			enabled:   false,
			method:    htmlEncodeSpecialOnly,
		},
		encodingScheme{
			params:    []string{"dh", "doublehtml"},
			desc:      "Double HTML encode key characters: '&' --> '&amp;amp;'",
			isDefault: false,
			enabled:   false,
			method:    doubleHtmlEncodeSpecialOnly,
		},
	}

	incrementalPtr := flag.Bool("i", false, "For each input string, incrementally encode each character and print out the string")
	stdinPtr := flag.Bool("stdin", false, "Read input strings from stdin")
	inputFilePtr := flag.String("f", "", "File to read input strings from")
	outputFilePtr := flag.String("o", "", "File to output results to")
	encodingSchemesPtr := flag.String("e", generateDefaultEncodingSchemes(encodingSchemes), "Comma seperated list of encoding schemes to use\n"+generateEncodingSchemeDocs(encodingSchemes))

	flag.Parse()

	// Workout if more than one mode has been selected
	modeCount := 0

	if *stdinPtr {
		modeCount++
	}

	if *inputFilePtr != "" {
		modeCount++
	}

	if flag.NArg() > 0 {
		modeCount++
	}

	if modeCount > 1 {
		log.Fatal("Pass a string either by: strings seperated by spaces on the command line, from an input file with -f or via stdin with -stdin")
	}

	// Check stdin, maybe the user didn't pass the arg for it..
	//FIXME if doing echo test | ./gonfuzz it prints out test at the top
	stdin := os.Stdin
	stdinStat, _ := stdin.Stat()
	stdinSize := stdinStat.Size()

	if stdinSize > 0 {
		*stdinPtr = true
	}

	// If no input modes were provided or stdin not detected, quit and show usage
	if flag.NArg() == 0 && *stdinPtr == false && *inputFilePtr == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Validate encoding schemes provided actually exist
	if *encodingSchemesPtr != "" {
		encodingSchemesEntered := strings.Split(*encodingSchemesPtr, ",")
		errors := ""

		// for each entered encoding scheme
		for _, encSchemeEnt := range encodingSchemesEntered {
			found := false
			// check each schemes params to see if it matches
			for _, encScheme := range encodingSchemes {
				for _, encSchemeParam := range encScheme.params {
					if encSchemeParam == encSchemeEnt {
						found = true
						break
					}
				}
			}
			if !found {
				errors = errors + " " + encSchemeEnt
			}
		}

		if errors != "" {
			log.Fatal("The following encoding schemes could not be matched:" + errors)
		}
	}

	// Temp target list for pulling in candidates
	tmpTargetList := make([]string, 0)

	// For target strings passed on the command line
	if flag.NArg() > 0 {
		for i := 0; i < flag.NArg(); i++ {
			tmpTargetList = append(tmpTargetList, flag.Arg(i))
		}
	}

	targetList := make([]string, 0)

	// If reading from a file or stdin get the candidates into memory
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

	// Generate incremental lists based off of input, then merge
	if *incrementalPtr {
		for i := 0; i < len(tmpTargetList); i++ {
			tmpList := incrementalStringGenerator(tmpTargetList[i])
			for l := 0; l < len(tmpList); l++ {
				targetList = append(targetList, tmpList[l])
			}
		}
	} else {
		targetList = tmpTargetList
	}

	// Set encoding schemes
	encodingSchemesSelected := strings.Split(*encodingSchemesPtr, ",")

	for i := 0; i < len(encodingSchemes); i++ {
		// If no encoding schemes were provided, the encodingSchemes struct
		// already specifies the defaults and they will be set inside the
		// encodingSchemesSelected variable, we just need to enable as if they
		// were passed on the command line
		if Contains(encodingSchemesSelected, encodingSchemes[i].params) {
			encodingSchemes[i].enabled = true
		}
	}

	// If an output file was defined:
	var outputFile *os.File
	var outputFileErr error
	if *outputFilePtr != "" {
		outputFile, outputFileErr = os.Create(*outputFilePtr)

		if outputFileErr != nil {
			log.Fatal(outputFileErr)
		}
		defer fmt.Printf("Outputing to: %s\n", *outputFilePtr)
	}

	// Process candidates and output
	for e := 0; e < len(encodingSchemes); e++ {
		if encodingSchemes[e].enabled {
			for t := 0; t < len(targetList); t++ {
				if *outputFilePtr != "" {
					outputFile.Write([]byte(encodingSchemes[e].method(targetList[t]) + "\n"))
				} else {
					fmt.Println(encodingSchemes[e].method(targetList[t]))
				}
			}
		}
	}
}

// Plaintext funtion, needs to be in the same format s other encoding schemes
// to match the required func(string) string
func plainText(input string) string { return input }

// URL Encoding
func urlEncodeAllUppercase(input string) string {
	return urlEncode(input, false, false)
}

func urlEncodeAllLowercase(input string) string {
	return urlEncode(input, true, false)
}

func doubleUrlEncodeAllUppercase(input string) string {
	return urlEncode(input, false, true)
}

func doubleUrlEncodeAllLowercase(input string) string {
	return urlEncode(input, true, true)
}

func urlEncode(input string, lowercase bool, doubleUrlEncode bool) string {
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

// Golangs built in lib only encodes key characters
func doubleUrlEncodeSpecialOnly(input string) string {
	return url.QueryEscape(url.QueryEscape(input))
}

func urlEncodeSpecialOnly(input string) string {
	return url.QueryEscape(input)
}

// Split by runes in order to extract specific unicode chars
func incrementalStringGenerator(input string) []string {

	incrementalStringList := make([]string, 0)
	first := true

	for _, r := range []rune(input) {
		if first {
			incrementalStringList = append(incrementalStringList, fmt.Sprintf("%s", string(r)))
		} else {
			incrementalStringList = append(incrementalStringList, fmt.Sprintf("%s%s", incrementalStringList[len(incrementalStringList)-1], string(r)))
		}
		first = false
	}

	return incrementalStringList
}

// HTML encoding
func doubleHtmlEncodeAllAsDecimal(input string) string {
	return htmlEncodeAllAsDecimal(htmlEncodeAllAsDecimal(input))
}

func htmlEncodeAllAsDecimal(input string) string {
	outputString := ""

	r := []rune(input)
	for i := 0; i < len(r); i++ {
		outputString = outputString + fmt.Sprintf("&#%d;", r[i])
	}
	return outputString
}
func htmlEncodeAllAsHexLowercase(input string) string {
	return htmlEncodeAllAsHex(input, true, false)
}

func htmlEncodeAllAsHexUppercase(input string) string {
	return htmlEncodeAllAsHex(input, false, false)
}

func doubleHtmlEncodeAllAsHexLowercase(input string) string {
	return htmlEncodeAllAsHex(input, true, true)
}

func doubleHtmlEncodeAllAsHexUppercase(input string) string {
	return htmlEncodeAllAsHex(input, false, true)
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

// Golangs built in lib only encodes key characters
func doubleHtmlEncodeSpecialOnly(input string) string {
	return htmlEncodeSpecialOnly(htmlEncodeSpecialOnly(input))
}

func htmlEncodeSpecialOnly(input string) string {
	return html.EscapeString(input)
}

// Any of these in these?
func Contains(a []string, x []string) bool {
	for _, y := range x {
		for _, n := range a {
			if y == n {
				return true
			}
		}
	}
	return false
}

// Used to build flag description for encoding scheme parameter docs
func generateEncodingSchemeDocs(encodingSchemes []encodingScheme) string {
	docs := ""

	for i := 0; i < len(encodingSchemes); i++ {
		params := strings.Join(encodingSchemes[i].params, "/")
		docs = docs + fmt.Sprintf("    \t%s - %s\n", params, encodingSchemes[i].desc)
	}
	return docs
}

// Generate comma seperated list of default encoding schemes to use
func generateDefaultEncodingSchemes(encodingSchemes []encodingScheme) string {
	defaultList := ""
	first := true

	for i := 0; i < len(encodingSchemes); i++ {
		if encodingSchemes[i].isDefault == true {
			// Deal with trailing comma
			if first {
				first = false
				defaultList = defaultList + encodingSchemes[i].params[0]
			} else {
				defaultList = defaultList + "," + encodingSchemes[i].params[0]
			}
		}
	}
	return defaultList
}
