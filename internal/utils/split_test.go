package utils

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"
)

func TestSplitBySeparator(t *testing.T) {
	data := []byte("Hello,World,Test")
	separator := []byte(",")
	splitFunc := SplitBySeparator(separator)

	advance, token, err := splitFunc(data, false)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if advance != 6 { // "Hello," is 6 bytes
		t.Errorf("Expected advance to be 6, got %d", advance)
	}

	expectedToken := []byte("Hello")
	if !reflect.DeepEqual(token, expectedToken) {
		t.Errorf("Expected token to be %v, got %v", expectedToken, token)
	}
}
func TestSplitInScanner(t *testing.T) {
	data := []byte("Hello,World,Test")
	separator := []byte(",")

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(SplitBySeparator(separator))

	scanner.Scan()
	if scanner.Text() != "Hello" {
		t.Errorf("Expected token to be Hello, got %v", scanner.Text())
	}

	scanner.Scan()
	if scanner.Text() != "World" {
		t.Errorf("Expected token to be World, got %v", scanner.Text())
	}
	scanner.Scan()
	if scanner.Text() != "Test" {
		t.Errorf("Expected token to be Test, got %v", scanner.Text())
	}
	if scanner.Scan() {
		t.Errorf("Expected scanner to be done")
	}
}

func TestSplitInScannerTrailingSeparator(t *testing.T) {
	data := []byte("Hello,World,Test,")
	separator := []byte(",")

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(SplitBySeparator(separator))

	scanner.Scan()
	if scanner.Text() != "Hello" {
		t.Errorf("Expected token to be Hello, got %v", scanner.Text())
	}

	scanner.Scan()
	if scanner.Text() != "World" {
		t.Errorf("Expected token to be World, got %v", scanner.Text())
	}
	scanner.Scan()
	if scanner.Text() != "Test" {
		t.Errorf("Expected token to be Test, got %v", scanner.Text())
	}
	if scanner.Scan() {
		t.Errorf("Expected scanner to be done")
	}
}
