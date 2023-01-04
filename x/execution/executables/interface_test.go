package executables_test

import (
	"fmt"
	"kwil/x/execution/executables"
	"kwil/x/execution/mocks"
	"testing"
)

func Test_DBInterface(t *testing.T) {
	// create a db interface from a mock database
	intfc, err := executables.FromDatabase(&mocks.Db1)
	if err != nil {
		t.Errorf("failed to create db interface: %v", err)
	}

	accessable := intfc.CanExecute("0xbennan", "insert1")
	if !accessable {
		t.Errorf("default role expected to be able to execute insert1")
	}

	accessable = intfc.CanExecute("0xbennan", "insert2")
	if accessable {
		t.Errorf("default role not expected to be able to execute insert2")
	}

	accessable = intfc.CanExecute("0xabc", "insert2")
	if !accessable {
		t.Errorf("expected to be able to execute insert2")
	}

	// try to prepare insert1
	wallet := "0xbennan"
	params, err := intfc.Prepare("insert1", wallet, mocks.Insert1Inputs)
	if err != nil {
		t.Errorf("failed to prepare insert1: %v", err)
	}

	if fmt.Sprint(params[0]) != wallet {
		t.Errorf("expected wallet to be first parameter")
	}

	if fmt.Sprint(params[1]) != "421" {
		t.Errorf("expected 421 to be second parameter. got %v", params[1])
	}

	// try to prepare insert2
	wallet = "0xabc"
	params, err = intfc.Prepare("insert2", wallet, mocks.Insert2Inputs)
	if err != nil {
		t.Errorf("failed to prepare insert2: %v", err)
	}

	if fmt.Sprint(params[0]) != wallet {
		t.Errorf("expected wallet to be first parameter")
	}

	if fmt.Sprint(params[1]) != "true" {
		t.Errorf("expected true to be second parameter. got %v", params[1])
	}

	// try to prepare update1
	wallet = "0xbennan"
	params, err = intfc.Prepare("update1", wallet, mocks.Update1Inputs)
	if err != nil {
		t.Errorf("failed to prepare update1: %v", err)
	}

	if params[0] != "0xbennan" || params[2] != "0xbennan" {
		t.Errorf("expected wallet to be first and third parameter")
	}

	if fmt.Sprint(params[1]) != "421" {
		t.Errorf("expected true to be second parameter. got %v", params[1])
	}

	// try to prepare update2
	wallet = "0xabc"
	params, err = intfc.Prepare("update2", wallet, mocks.Update2Inputs)
	if err != nil {
		t.Errorf("failed to prepare update2: %v", err)
	}

	if fmt.Sprint(params[0]) != "0xabc" {
		t.Errorf("expected wallet to be first parameter")
	}
	if fmt.Sprint(params[1]) != "true" || fmt.Sprint(params[2]) != "true" {
		t.Errorf("expected true to be second and third parameter. second: %v, third: %v", params[1], params[2])
	}

	// try to prepare delete1
	wallet = "0xbennan"
	params, err = intfc.Prepare("delete1", wallet, mocks.Delete1Inputs)
	if err != nil {
		t.Errorf("failed to prepare delete1: %v", err)
	}

	if fmt.Sprint(params[0]) != "0xbennan" {
		t.Errorf("expected wallet to be first parameter")
	}

	// try to prepare delete2
	wallet = "0xabc"
	params, err = intfc.Prepare("delete2", wallet, mocks.Delete2Inputs)
	if err != nil {
		t.Errorf("failed to prepare delete2: %v", err)
	}

	if fmt.Sprint(params[0]) != "true" {
		t.Errorf("expected true to be second parameter. got %v", params[1])
	}
}
