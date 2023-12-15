package main

import (
	"testing"
)

// Measures the performance of checkPipe with the command ls | wc
func BenchmarkCheckPipe(b *testing.B) {
	for i := 0; i < b.N; i++ {
	   checkPipe("ls | wc")
	}
 } 

// Measures the performance of processString
func BenchmarkProcessString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		args := make([]string, 0)
		argsPipe := make([]string, 0)
		processString("ls -l | grep test_file.txt", &args, &argsPipe)
	}
}

// Benchmark for the shell command echo "I love COS316"
func BenchmarkEchoCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Run the function being benchmarked
		shellCommand([]string{"echo", "I love COS316"})
	}
}

// Benchmark for the processArgs function with ls command
func BenchmarkProcessArgs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Run the function being benchmarked
		processArgs([]string{"ls"})
	}
}

// Benchmark for the processArgsPipe function with ls | wc
func BenchmarkProcessArgsPipe(b *testing.B) {
	bufferArgs := []string{"ls"}
	bufferPipe := []string{"wc"}
	for i := 0; i < b.N; i++ {
	   processArgsPipe(bufferArgs, bufferPipe)
	}
}