package structures_test

import (
	"testing"

	"kwil/x/deposits_old/structures"

	"github.com/stretchr/testify/assert"
)

func TestWithdrawTracker(t *testing.T) {
	wt := structures.NewWithdrawalTracker()

	// add withdrawals
	wt.Insert(newMockWithdrawal(100, "n1"))
	wt.Insert(newMockWithdrawal(200, "n2"))
	wt.Insert(newMockWithdrawal(50, "n3"))
	wt.Insert(newMockWithdrawal(150, "n4"))
	wt.Insert(newMockWithdrawal(250, "n5"))

	// check the size
	ws := wt.PopExpired(100)
	assert.Equal(t, 2, len(ws))

	// assert that these no longer exist here
	ws = wt.PopExpired(100)
	assert.Equal(t, 0, len(ws))

	// testing for nil memory reference when polling highest number
	ws = wt.PopExpired(250)
	assert.Equal(t, 3, len(ws))

	wt.Insert(newMockWithdrawal(100, "n1"))
	wt.Insert(newMockWithdrawal(200, "n2"))
	wt.Insert(newMockWithdrawal(50, "n3"))
	wt.Insert(newMockWithdrawal(150, "n4"))
	wt.Insert(newMockWithdrawal(250, "n5"))

	// now try removing by nonce
	wt.RemoveByNonce("n4")
	wth := wt.GetByNonce("n4")
	assert.Nil(t, wth)

	// we should still have n1, n2, n3, n5
	wth = wt.GetByNonce("n1")
	assert.NotNil(t, wth)
	wth = wt.GetByNonce("n2")
	assert.NotNil(t, wth)
	wth = wt.GetByNonce("n3")
	assert.NotNil(t, wth)
	wth = wt.GetByNonce("n5")
	assert.NotNil(t, wth)

	// now garbage collect
	wt.RunGC()

	// we should still have n1, n2, n3, n5
	wth = wt.GetByNonce("n1")
	assert.NotNil(t, wth)
	wth = wt.GetByNonce("n2")
	assert.NotNil(t, wth)
	wth = wt.GetByNonce("n3")
	assert.NotNil(t, wth)
	wth = wt.GetByNonce("n5")
	assert.NotNil(t, wth)
}

func TestPop(t *testing.T) {
	// this is already tested in the function above but its failing elsewhere so this is a sanity check
	wt := structures.NewWithdrawalTracker()
	wt.Insert(newMockWithdrawal(100, "n1"))
	wt.Insert(newMockWithdrawal(200, "n2"))

	ws := wt.PopExpired(100)
	assert.Equal(t, 1, len(ws))

	node := wt.GetByNonce("n1")
	assert.Nil(t, node)
}

func TestRemoveNonexistentNonce(t *testing.T) {
	wt := structures.NewWithdrawalTracker()
	wt.Insert(newMockWithdrawal(100, "n1"))
	wt.Insert(newMockWithdrawal(200, "n2"))

	wt.RemoveByNonce("n3")
}
