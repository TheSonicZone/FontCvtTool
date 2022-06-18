package main

//------------------------------------------------------------------------------------------------------------
//
//								FontCvtTool
//        Converts output from Segger's FontCvt to a binary (FLASH memory) format or the format
//        used by my legacy library GraphLCD
//------------------------------------------------------------------------------------------------------------

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var VersionString = "1.0.0.1"
var inputfile = ""
var outputfile = ""
var options = ""
var validFontCvtSrcFile bool = false
var validSeggerFile bool = false
var fontNameFromSrc = ""

// Function to read the file in lines.. rather than seeking...
//-------------------------------------------------------------
func readLines(filename string) ([]string, error) {
	var lines []string
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return lines, err
	}
	buf := bytes.NewBuffer(file)
	for {
		line, err := buf.ReadString('\n')
		if len(line) == 0 {
			if err != nil {
				if err == io.EOF {
					break
				}
				return lines, err
			}
		}
		lines = append(lines, line)
		if err != nil && err != io.EOF {
			return lines, err
		}
	}
	return lines, nil
}

//-----------------------------------------------------------------------------------------------
// Main Function
//------------------------------------------------------------------------------------------------
func main() {

	// Store program arguments... if given, if not, args will have no length but will NOT be = nil!
	args := os.Args[1:]

	fmt.Println(" ")
	fmt.Println("+----------------------------------------------------------------------+")
	fmt.Println("|                             FontCvtTool                              |")
	fmt.Println("|   Conversion utility for working with Segger FontCvt in standalone   |")
	fmt.Println("|                              By Sonic2k                              |")
	fmt.Println("+----------------------------------------------------------------------+")
	fmt.Println("  ")

	if len(args) == 0 {
		fmt.Println(" [ERROR] Invalid program arguments, try fontcvttool --help")
		return // Terminate program
	}

	// Parse program arguments here before the program is run, the --help command has highest priority
	// if we find the --help argument anywhere we display that and ignore all other command-line args
	for n := 0; n < len(args); n++ {
		if args[n] == "--help" {
			fmt.Println("   Command-line arguments help:")
			fmt.Println(" ")
			fmt.Println("   -graphlcd         - Output a .c file that works with GraphLCD library")
			fmt.Println("                       for microcontrollers")
			fmt.Println("   -graphlcd_compat  - Output a .c file that is compatible with")
			fmt.Println("                       GraphLCD (legacy)")
			fmt.Println("   -bin              - Write a binary file that can be used with")
			fmt.Println("                       an EPROM programmer or GraphLCD's")
			fmt.Println("                       embedded font loader")
			fmt.Println("   --help            - This help text")
			fmt.Println("   --version         - Program version")
			return
		}
	}
	// Parse program arguments here for --version which has the same priority as --help
	// if we find the --version argument anywhere we display that and ignore all the other command-line args
	for n := 0; n < len(args); n++ {
		if args[n] == "--version" {
			fmt.Println("  FontCvtTool - conversion tool for Segger's FontCvt ")
			fmt.Println("                by Sonic2k     Version: " + VersionString)
			return
		}
	}

	// General parsing of command line arguments, if anything is wrong we display the generic format line
	// First parameter: -infile command-line option
	if len(args) == 3 {
		// We have all the necessary arguments.... parse

		if strings.Contains(args[0], "-infile") {
			f := strings.Split(args[0], "=")
			inputfile = f[1]
		} else {
			fmt.Println(" [ERROR] Invalid argument(s) \n\r Usage: fontcvttool -infile=<filename> -outfile=<filename> -option")
			return
		}

		if strings.Contains(args[1], "-outfile") {
			f := strings.Split(args[1], "=")
			outputfile = f[1]
		} else {
			fmt.Println(" [ERROR] Invalid argument(s) \n\r Usage: fontcvttool -infile=<filename> -outfile=<filename> -option")
			return
		}
		options = args[2]
	} else {
		fmt.Println(" [ERROR] Invalid argument(s) \n\r Usage: fontcvttool -infile=<filename> -outfile=<filename> -option")
		return
	}

	// Check all arguments parsed and act as directed
	fmt.Println("         ********** DEBUG INFO *******************")
	fmt.Println("               infile argument = " + inputfile)
	fmt.Println("               outfile argument = " + outputfile)
	fmt.Println("               option argument = " + options)
	fmt.Println("         *****************************************")
	fmt.Println("       ")

	// Open up the file for reading...
	lines, err := readLines(inputfile)
	if err != nil {
		fmt.Println(" [ERROR] Unable to open/read given file (" + inputfile + ")")
		return
	}
	// File was opened... now begin to parse it
	fmt.Println(fmt.Sprintf("  File Opened- number of lines = %d", len(lines)))
	fmt.Println(" Parsing input file...")

	// Perform checks for file validity by looping through the first 50-100 lines to look for
	// stuff that indicates its a valid file
	for n := 0; n < 100; n++ {
		if strings.Contains(lines[n], "Font converter for emWin") {
			validFontCvtSrcFile = true
			fmt.Println("   Valid FontCvt source...            [OK]")
		}
		if strings.Contains(lines[n], "Segger Microcontroller") {
			validSeggerFile = true
			fmt.Println("   Valid Segger \"c\" file format...    [OK]")
		}
		if strings.Contains(lines[n], "* Source file:") {
			s := strings.Split(lines[n], ": ")
			t := strings.Split(s[1], ".c")
			fmt.Println("   Original Source Filename: " + t[0] + ".c")
		}

	}
	if validFontCvtSrcFile == true && validSeggerFile == true {

	} else {
		fmt.Println(" [ERROR] Invalid file formatting (" + inputfile + ")")
		return
	}

	// This loop will now step through each line of the file and parse out character bitmap data as well as
	// the other parameters

	var current_line = 0
	for {

		if current_line >= len(lines) {
			fmt.Println("    End of file reached...")
			break
		}

		if strings.Contains(lines[current_line], "GUI_CONST_STORAGE unsigned char acGUI") {
			// the line is the start of a const char array..
			p1 := strings.Split(lines[current_line], "acGUI_")
			p2 := strings.Split(p1[1], "_")
			p3 := strings.Split(p2[1], "[")

			unicode, error := strconv.ParseInt(p3[0], 16, 32)
			if error != nil {
				fmt.Println("  [FATAL] - Unable to decode Unicode Hex from source file... Operation Aborted!")
			}
			// Unicode character code is read, now advance to the next line and examine the pixel data
			// We figure out how many rows and columns the character fits in
			fmt.Println(fmt.Sprintf("   Unicode Char Code: %04X", unicode))

			current_line++ // Advance to next line

			for {
				var pattern string = ""
				q1 := strings.Split(lines[current_line], "  ") // Split.. left part = spaces, right part = bit pattern
				q2 := strings.SplitAfter(q1[1], ",")

				for n := 0; n < len(q2)-1; n++ {
					pattern += strings.Trim(q2[n], ",")
					pattern += " "
				}
				fmt.Println("  Row Bit Pattern: " + pattern)
				current_line++
				if strings.Contains(lines[current_line], "};") {
					break
				}
			}

			//fmt.Println(fmt.Sprintf("    └ Vert. Col = %d   Bit Patt: %s", len(q2)-1, pattern))

		}
		current_line++
	}
	fmt.Println("Out of parsing loop... ")
}
