package main

import (
	"testing"
)

func TestValidateFileName1(t *testing.T) {
	actual := validateFileName("0.1.png", 2)
	expected := false
	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func TestValidateFileName2(t *testing.T) {
	actual := validateFileName("test.png", 2)
	expected := false
	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func TestValidateFileName3(t *testing.T) {
	actual := validateFileName("3.gif", 2)
	expected := false
	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func TestValidateFileName4(t *testing.T) {
	actual := validateFileName("0.png", 2)
	expected := false
	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func TestValidateFileName5(t *testing.T) {
	actual := validateFileName("2.png", 2)
	expected := true
	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func TestValidateFileName6(t *testing.T) {
	actual := validateFileName("3.png", 2)
	expected := true
	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func TestGetIndex(t *testing.T) {
	actual := getIndex("0.png")
	expected := 0
	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func TestFilterFileNames(t *testing.T) {
	result := filterFileNames([]string{"0.png", "1.png", "2.png", "test.txt"}, 1)

	actual0 := len(result)
	expected0 := 2
	if actual0 != expected0 {
		t.Errorf("got: %v\nwant: %v", actual0, expected0)
	}

	actual1 := result[0]
	expected1 := "1.png"
	if actual1 != expected1 {
		t.Errorf("got: %v\nwant: %v", actual1, expected1)
	}

	actual2 := result[1]
	expected2 := "2.png"
	if actual2 != expected2 {
		t.Errorf("got: %v\nwant: %v", actual2, expected2)
	}
}

func TestNumericalSortFileNames(t *testing.T) {
	fileNames := []string{"11.png", "9.png", "10.png"}
	numericalSortFileNames(fileNames)

	actual0 := fileNames[0]
	expected0 := "9.png"
	if actual0 != expected0 {
		t.Errorf("got: %v\nwant: %v", actual0, expected0)
	}

	actual1 := fileNames[1]
	expected1 := "10.png"
	if actual1 != expected1 {
		t.Errorf("got: %v\nwant: %v", actual1, expected1)
	}

	actual2 := fileNames[2]
	expected2 := "11.png"
	if actual2 != expected2 {
		t.Errorf("got: %v\nwant: %v", actual2, expected2)
	}
}
