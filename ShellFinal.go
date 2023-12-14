package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var isPipe int
var isBackground int
var input string
var stopIt bool
var wg sync.WaitGroup
var numPipes int
var buffer string

func main() {
	// collect the arguments and pipe arguments
	var bufferArgs []string
	var bufferPipe []string

	for {
		fmt.Printf("User@%s:~ ", os.Getenv("USER"))
		// read in commands
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		buffer = scanner.Text()
		flag := processString(buffer, &bufferArgs, &bufferPipe)
		isBackground = 0
		//  Piping
		//ecfmt.Println(flag)
		if flag == 1 {
			isPipe = 1
			if isBackgroundTask(bufferPipe) {
				isBackground = 1
				go processArgsPipe(buffer)
			} else {
				processArgsPipe(buffer)
			}
			// && function
		} else if flag == 4 {
			isPipe = 0
			if isBackgroundTask(bufferPipe) {
				isBackground = 1
				go processANDAND(bufferArgs, bufferPipe)
			} else {
				processANDAND(bufferArgs, bufferPipe)
			}
			// regular command
		} else if flag == 2 {
			isPipe = 0
			if isBackgroundTask(bufferArgs) {
				isBackground = 1
				go runInBackground(processArgs, bufferArgs)
			} else {
				processArgs(bufferArgs)
			}
			// Check for background execution and run our implementation
		} else if flag == 3 {
			isPipe = 0
			if isBackgroundTask(bufferArgs) {
				isBackground = 1
				go shellCommand(bufferArgs)
			} else {
				shellCommand(bufferArgs)
			}
			// run system call
		} else if flag == 5 {
			isPipe = 0
			isBackground = 1
			if isBackgroundTask(bufferArgs) {
				isBackground = 1
				runInBackground(processArgs, bufferArgs)

			} else {
				processArgs(bufferArgs)
			}

			//time.Sleep(3 * time.Second)
			//stopIt = true
			//wg.Wait()

		} else if flag == 6 {
			continue
		}
		fmt.Printf("\n")
	}
}

// see if there is a pipe
func checkPipe(str string) bool {
	for i := 0; i < len(str); i++ {
		if str[i] == '|' {
			return true
		}
	}
	return false
}

// seperate by space
func parseSpace(str string) []string {

	var rows []string
	rows = strings.Split(str, " ")

	return rows
}

// seperate by pipes
func parsePipe(str string) []string {
	var rows []string
	rows = strings.Split(str, "|")
	numPipes = strings.Count(str, "|")
	return rows
}

// divide the strings of input
func processString(str string, args *[]string, argspipe *[]string) int {

	check := checkPipe(str)
	parsePipe := parsePipe(str)
	var result int
	s := strings.TrimSpace(str)
	if len(s) == 0 || s == " " || s == "\n" {
		return 6
	} else if check {
		*args = parseSpace(strings.TrimSpace(parsePipe[0]))
		*argspipe = parseSpace(strings.TrimSpace(parsePipe[1]))
		result = 1
		return result
	} else if strings.Contains(str, "&&") {
		*args = parseSpace(strings.TrimSpace(strings.Split(str, "&&")[0]))
		*argspipe = parseSpace(strings.TrimSpace(strings.Split(str, "&&")[1]))

		result = 4
		return result

	} else {
		*args = parseSpace(str)
		result = 2
	}

	if checkShellCommand(*args) {
		result = 3
	} else if !checkShellCommand(*args) {
		result = 5
	}
	return result

}

// run commands we have in our implementation
func shellCommand(buffer []string) int {
	success := 1
	//fmt.Println("run")
	if isBackgroundTask(buffer) {
		buffer = buffer[:len(buffer)-1]
	}

	switch buffer[0] {

	case "cd":
		err := os.Chdir(buffer[1])
		if err != nil {
			success = 0
			log.Print(err)
		}
		return success

	case "mkdir":
		err := os.Mkdir(buffer[1], 0755)
		if err != nil {
			log.Print(err)
			success = 0
		}
		return success
	case "rename":
		err := os.Rename(buffer[1], buffer[2])
		if err != nil {
			log.Print(os.LinkError{})
			success = 0
		}
		return success
	case "remove":
		for i := 1; i < len(buffer); i++ {
			err := os.Remove(buffer[i])
			if err != nil {
				log.Print(os.PathError{})
				success = 0
			}

		}

		return success
	case "getpid":
		fmt.Printf("%d", os.Getpid())
		return success
	case "pwd":
		if len(buffer) > 1 && buffer[1] == ">" {
			if len(buffer) > 2 {
				outputFile := buffer[2]
				file, err := os.Create(outputFile)
				defer file.Close()
				if err != nil {
					log.Print(err)
					return 0
				}

				// Get the current working directory and write it to the output file
				mydir, err := os.Getwd()
				if err != nil {
					fmt.Println(err)
					return 0
				}

				_, err = fmt.Fprintln(file, mydir)
				if err != nil {
					log.Print(err)
					return 0
				}
			} else {
				fmt.Println("Usage: pwd > OUTPUT_FILE")
			}
		} else {
			// If no redirection, simply print the current working directory to the console
			mydir, err := os.Getwd()
			if err != nil {
				fmt.Println(err)
				return 0
			}
			fmt.Println(mydir)
		}
		return success
	case "setenv":
		if len(buffer) < 3 {
			fmt.Println("Usage: setenv VARIABLE VALUE")
		} else {
			err := os.Setenv(buffer[1], buffer[2])
			if err != nil {
				log.Print(err)
				success = 0
			}
		}
		return success
	case "getenv":
		if len(buffer) < 2 {
			fmt.Println("Usage: getenv VARIABLE")
		} else {
			varName := buffer[1]
			varValue := os.Getenv(varName)
			if varValue == "" {
				fmt.Printf("Environment variable '%s' not set\n", varName)
			} else {
				fmt.Printf("%s=%s\n", varName, varValue)
			}
		}
		return success
	case "unsetenv":
		if len(buffer) < 2 {
			fmt.Println("Usage: unsetenv VARIABLE")
		} else {
			err := os.Unsetenv(buffer[1])
			if err != nil {
				log.Print(err)
				success = 0
			}
		}
		return success
	case "echo":
		if len(buffer) > 2 && buffer[2] == ">" {
			if len(buffer) > 3 {
				outputFile := buffer[3]
				file, err := os.Create(outputFile)
				defer file.Close()
				if err != nil {
					log.Print(err)
					return 0
				}

				// Write the echoed message to the output file
				_, err = fmt.Fprintln(file, strings.Join(buffer[1:], " "))
				if err != nil {
					log.Print(err)
					return 0
				}
			} else {
				fmt.Println("Usage: echo MESSAGE > OUTPUT_FILE")
			}
		} else {
			// If no redirection, simply echo the message to the console
			fmt.Println(strings.Join(buffer[1:], " "))
		}
		return success
	case "ls":
		dirPath := "."
		if len(buffer) > 1 {
			dirPath = buffer[1]
		}

		files, err := os.ReadDir(dirPath)
		if err != nil {
			log.Println(err)
			return 0
		}

		for _, file := range files {

			fmt.Println(file.Name())
		}

		return success
	case "cat":
		if len(buffer) < 2 {
			fmt.Println("Usage: cat FILE [> OUTPUT_FILE] [>> APPEND_OUTPUT_FILE] [> OUTPUT]")
			return 0
		}
		inputFile := ""
		outputFile := ""
		appendFile := false
		if len(buffer) == 2 {

			catContent, err := os.ReadFile(buffer[1])
			if err != nil {
				return 0
			}
			fmt.Printf("%s", catContent)
			return success
		}

		// Check for input redirection
		if len(buffer) > 1 && buffer[1] == "<" {
			if len(buffer) > 2 {
				inputFile = buffer[2]
			} else {
				fmt.Println("Usage: cat FILE < INPUT_FILE")
				return 0
			}
		}
		if len(buffer) > 1 && buffer[1] == ">" {
			if len(buffer) > 2 {
				outputFile = buffer[2]
			} else {
				fmt.Println("Usage: cat FILE < INPUT_FILE")
				return 0
			}
		}

		// Check for output redirection
		if len(buffer) > 3 {
			if buffer[2] == ">" {
				inputFile = buffer[1]
				outputFile = buffer[3]
			}
			if buffer[2] == ">>" {
				inputFile = buffer[1]
				outputFile = buffer[3]
				appendFile = true
			}

		}
		if len(buffer) > 4 {
			if buffer[1] == "<" && buffer[3] == ">" {
				inputFile = buffer[2]
				outputFile = buffer[4]
			}
			if buffer[1] == ">" && buffer[3] == "<" {
				inputFile = buffer[4]
				outputFile = buffer[2]
				//appendFile = true
			}
		}

		// Perform the actual cat operation based on input and output redirection
		catContent, err := os.ReadFile(inputFile)
		if err != nil {
			log.Print(err)
			return 0
		}

		if outputFile != "" {
			fileMode := os.O_WRONLY | os.O_CREATE

			// Use append mode for ">>"
			if appendFile {
				fileMode |= os.O_APPEND
			} else {
				fileMode |= os.O_TRUNC
			}

			// Open or create the file
			file, err := os.OpenFile(outputFile, fileMode, 0644)
			defer file.Close()
			if err != nil {
				log.Print(err)
				return 0
			}

			// Write the cat content to the output file
			_, err = file.Write(catContent)
			if err != nil {
				log.Print(err)
				return 0
			}
			return success
		} else {
			// If no output redirection, simply print the cat content to the console
			fmt.Printf("%s", catContent)
		}

		return success
	case "exit":
		fmt.Println("Goodbye!")
		os.Exit(0)
		return success

	}
	return 0

}

// Check if it is a command we have
func checkShellCommand(buffer []string) bool {

	command := []string{"cd", "pwd", "echo", "mkdir", "rename", "remove", "getpid", "cat", "setenv", "getenv", "unsetenv", "ls", "exit"}
	for i := 0; i < len(command); i++ {
		if buffer[0] == command[i] {
			if len(buffer) == 1 {
				return true
			}
			if len(buffer) > 1 {
				if strings.Contains(buffer[1], "-") {
					return false
				}
			}

			return true

		}
	}
	return false
}

// run Pipe in background
func runInBackgroundPipe(fn func([]string, []string), args []string, pipeArgs []string) {
	go fn(args, pipeArgs)
}
func processANDAND(buffer []string, bufferPipe []string) {
	if isBackgroundTask(bufferPipe) {
		bufferPipe = bufferPipe[:len(bufferPipe)-1]
	}
	if isPipe == 0 {
		if checkShellCommand(buffer) {
			suc := shellCommand(buffer)
			if suc == 1 && checkShellCommand(bufferPipe) {
				shellCommand(bufferPipe)
			} else if (suc == 1) && isPipe == 1 {
				bufferPipe = append(bufferPipe, input)
				processArgs(bufferPipe)
			} else if suc == 1 {
				processArgs(bufferPipe)
			}
		} else {
			processArgs(buffer)
			if checkShellCommand(bufferPipe) {
				shellCommand(bufferPipe)
			} else {
				processArgs(bufferPipe)
			}

		}
		return
	}
}

// process Pipe
func processArgsPipe(buffer string) {
	pipes := strings.Split(buffer, "|")
	commands := make([][]string, len(pipes))
	for i, pipe := range pipes {
		commands[i] = parseSpace(strings.TrimSpace(pipe))
	}

	var finalResult string
	var previousOutput []byte

	for i := 0; i < len(commands); i++ {
		cmd := exec.Command(commands[i][0], commands[i][1:]...)

		if i > 0 {
			cmd.Stdin = bytes.NewBuffer(previousOutput)
		}

		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb

		err := cmd.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		finalResult = strings.TrimSpace(outb.String())
		previousOutput = outb.Bytes()

		if i == len(commands)-1 {
			// Print the final result only after the last command
			fmt.Println("Final Result:", finalResult)
		}
	}

}

// runInBackground runs a function in the background
func runInBackground(fn func([]string), args []string) {
	go fn(args)
}

// isBackgroundTask checks if the command should run in the background
func isBackgroundTask(args []string) bool {
	return len(args) > 0 && args[len(args)-1] == "&"
}

// process command that goes to system
func processArgs(buffer []string) {
	if isBackgroundTask(buffer) {
		buffer = buffer[:len(buffer)-1]
	}
	//wg.Add(1)

	//defer wg.Done()
	//fmt.Printf("Command is: %s\n ", buffer[0])
	var cmd *exec.Cmd

	if len(buffer) > 1 {
		cmd = exec.Command(buffer[0], buffer[1:]...)
		//fmt.Println(buffer[1])
	} else {
		cmd = exec.Command(buffer[0])
	}

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	_ = cmd.Run()

	fmt.Println(outb.String())

	// If not a background task, wait for the command to finish
	if !isBackgroundTask(buffer) || isBackground == 0 {
		cmd.Wait()
	}

}
