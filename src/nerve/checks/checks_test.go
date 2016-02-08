package checks_test

import (
	"nerve/checks"
	"testing"
)

func TestCreateCheck(t *testing.T) {
	var emptyStringSlice []string
	//Create a Check with an unvalied type, will default to TCP theoriticaly
	check, err := checks.CreateCheck("NainPorteQuoi","1.2.3.4","apely",1234,42,false,"","","","",emptyStringSlice)
	if err != nil {
		t.Fatal("Unable to create NainPorteQuoi check with error: ",err)
	}
	if check.GetType() != "TCP" {
		t.Error("NainPorteQuoi: nvalid Check Type, expected TCP, got ",check.GetType())
	}
	//Create a TCP Check and verify it
	check, err = checks.CreateCheck("tcp","1.2.3.4","apely",1234,42,false,"","","","",emptyStringSlice)
	if err != nil {
		t.Fatal("Unable to create TCP check with error: ",err)
	}
	if check.GetType() != "TCP" {
		t.Error("TCP: Invalid Check Type, expected TCP, got ",check.GetType())
	}
}
