package schema

import "testing"

func Test_checkAllowedCharacters(t *testing.T) {
	ok, err := checkAllowedCharacters("kwil")
	if err != nil {
		t.Errorf(err.Error())
	}
	if !ok {
		t.Errorf("should be ok")
	}

	ok, err = checkAllowedCharacters("kwil_")
	if err != nil {
		t.Errorf(err.Error())
	}
	if !ok {
		t.Errorf("should be ok")
	}

	ok, err = checkAllowedCharacters("_kwil_123")
	if err != nil {
		t.Errorf(err.Error())
	}
	if !ok {
		t.Errorf("should be ok")
	}

	ok, err = checkAllowedCharacters("$123")
	if err != nil {
		t.Errorf(err.Error())
	}
	if ok {
		t.Errorf("should not be ok")
	}

	ok, err = checkAllowedCharacters("123")
	if err != nil {
		t.Errorf(err.Error())
	}
	if ok {
		t.Errorf("should not be ok")
	}

	ok, err = checkAllowedCharacters("Hello$#%^3_")
	if err != nil {
		t.Errorf(err.Error())
	}

	if ok {
		t.Errorf("should not be ok")
	}
}
