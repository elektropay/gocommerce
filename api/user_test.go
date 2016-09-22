package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/guregu/kami"
	tu "github.com/netlify/gocommerce/testutils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"github.com/netlify/gocommerce/conf"
	"github.com/netlify/gocommerce/models"
)

func TestUsersQueryForAllUsersAsStranger(t *testing.T) {
	config := testConfig()
	config.JWT.AdminGroupName = "admin"
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlWithUserID, nil)
	ctx := testContext(token("magical-unicorn", "", nil), config)

	NewAPI(config, db, nil).GetAllUsers(ctx, recorder, req)
	validateError(t, 401, recorder)
}

func TestUsersQueryForAllUserWithParams(t *testing.T) {
	toDie := models.User{
		ID:    "villian",
		Email: "twoface@dc.com",
	}
	db.Create(&toDie)
	defer db.Unscoped().Delete(&toDie)

	config := testConfig()
	config.JWT.AdminGroupName = "admin"
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://junk?email=twoface@dc.com", nil)
	ctx := testContext(token("magical-unicorn", "", &[]string{"admin"}), config)

	NewAPI(config, db, nil).GetAllUsers(ctx, recorder, req)

	users := []models.User{}
	extractPayload(t, 200, recorder, &users)
	assert.Equal(t, 1, len(users))
	assert.Equal(t, "villian", users[0].ID)
}

func TestUsersQueryForAllUsers(t *testing.T) {
	toDie := models.User{
		ID:    "villian",
		Email: "twoface@dc.com",
	}
	db.Create(&toDie)
	defer db.Unscoped().Delete(&toDie)

	config := testConfig()
	config.JWT.AdminGroupName = "admin"
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlWithUserID, nil)
	ctx := testContext(token("magical-unicorn", "", &[]string{"admin"}), config)

	NewAPI(config, db, nil).GetAllUsers(ctx, recorder, req)

	users := []models.User{}
	extractPayload(t, 200, recorder, &users)

	for _, u := range users {
		switch u.ID {
		case toDie.ID:
			assert.Equal(t, "twoface@dc.com", u.Email)
		case tu.TestUser.ID:
			assert.Equal(t, tu.TestUser.Email, u.Email)
		default:
			assert.Fail(t, "unexpected user %v\n", u)
		}
	}
}

//func TestUsersQueryForDeletedUser(t *testing.T) {
//	toDie := models.User{
//		ID:    "def-should-not-exist",
//		Email: "twoface@dc.com",
//	}
//	db.Create(&toDie)
//	db.Delete(&toDie) // soft delete
//	defer db.Unscoped().Delete(&toDie)
//
//	recorder := httptest.NewRecorder()
//	req, _ := http.NewRequest("GET", urlWithUserID, nil)
//
//	config := testConfig()
//	ctx := testContext(token(toDie.ID, toDie.Email, nil), config)
//	ctx = kami.SetParam(ctx, "user_id", toDie.ID)
//
//	api := NewAPI(config, db, nil)
//	api.GetSingleUser(ctx, recorder, req)
//	validateError(t, 404, recorder)
//}

func TestUsersQueryForUserAsUser(t *testing.T) {
	config := testConfig()
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlWithUserID, nil)

	ctx := testContext(token(tu.TestUser.ID, tu.TestUser.Email, nil), config)
	ctx = kami.SetParam(ctx, "user_id", tu.TestUser.ID)

	api := NewAPI(config, db, nil)
	api.GetSingleUser(ctx, recorder, req)
	user := new(models.User)
	extractPayload(t, 200, recorder, user)

	validateUser(t, &tu.TestUser, user)
}

func TestUsersQueryForUserAsStranger(t *testing.T) {
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlWithUserID, nil)

	config := testConfig()
	ctx := testContext(token("magical-unicorn", "", nil), config)
	ctx = kami.SetParam(ctx, "user_id", tu.TestUser.ID)

	api := NewAPI(config, db, nil)
	api.GetSingleUser(ctx, recorder, req)
	validateError(t, 401, recorder)
}

func TestUsersQueryForUserAsAdmin(t *testing.T) {
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlWithUserID, nil)

	config := testConfig()
	config.JWT.AdminGroupName = "admin"
	ctx := testContext(token("magical-unicorn", "", &[]string{"admin"}), config)
	ctx = kami.SetParam(ctx, "user_id", tu.TestUser.ID)

	NewAPI(config, db, nil).GetSingleUser(ctx, recorder, req)

	user := new(models.User)
	extractPayload(t, 200, recorder, user)
	validateUser(t, &tu.TestUser, user)
}

func TestUsersQueryForAllAddressesAsAdmin(t *testing.T) {
	second := tu.GetTestAddress()
	second.ID = "second"
	second.UserID = tu.TestUser.ID
	assert.True(t, second.Valid())
	db.Create(&second)
	defer db.Unscoped().Delete(&second)

	config := testConfig()
	config.JWT.AdminGroupName = "admin"
	ctx := testContext(token("magical-unicorn", "", &[]string{"admin"}), config)
	addrs := queryForAddresses(t, ctx, config, tu.TestUser.ID)
	assert.Equal(t, 2, len(addrs))
	for _, a := range addrs {
		assert.True(t, a.Valid(), fmt.Sprintf("invalid addr: %+v", a))
		switch a.ID {
		case second.ID:
			validateAddress(t, *second, a)
		case tu.TestAddress.ID:
			validateAddress(t, tu.TestAddress, a)
		default:
			assert.Fail(t, fmt.Sprintf("Unexpected address: %+v", a))
		}
	}
}

func TestUsersQueryForAllAddressesAsUser(t *testing.T) {
	config := testConfig()
	ctx := testContext(token(tu.TestUser.ID, "", nil), config)
	addrs := queryForAddresses(t, ctx, config, tu.TestUser.ID)
	assert.Equal(t, 1, len(addrs))
	validateAddress(t, tu.TestAddress, addrs[0])
}

func TestUsersQueryForAllAddressesAsStranger(t *testing.T) {
	config := testConfig()
	ctx := testContext(token("stranger-danger", "", nil), config)
	ctx = kami.SetParam(ctx, "user_id", tu.TestUser.ID)
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlWithUserID, nil)

	NewAPI(config, db, nil).GetAllAddresses(ctx, recorder, req)
	validateError(t, 401, recorder)
}

func TestUsersQueryForAllAddressesNoAddresses(t *testing.T) {
	u := models.User{
		ID:    "temporary",
		Email: "junk@junk.com",
	}
	db.Create(u)
	defer db.Unscoped().Delete(u)

	config := testConfig()
	ctx := testContext(token(u.ID, "", nil), config)
	addrs := queryForAddresses(t, ctx, config, u.ID)
	assert.Equal(t, 0, len(addrs))
}

func TestUsersQueryForAllAddressesMissingUser(t *testing.T) {
	config := testConfig()
	ctx := testContext(token("dne", "", nil), config)
	ctx = kami.SetParam(ctx, "user_id", "dne")
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlWithUserID, nil)

	NewAPI(config, db, nil).GetAllAddresses(ctx, recorder, req)
	validateError(t, 404, recorder)
}

func TestUserQueryForSingleAddressAsUser(t *testing.T) {
	config := testConfig()
	ctx := testContext(token(tu.TestUser.ID, "", nil), config)

	ctx = kami.SetParam(ctx, "user_id", tu.TestUser.ID)
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlWithUserID, nil)

	NewAPI(config, db, nil).GetSingleAddress(ctx, recorder, req)

	addr := new(models.Address)
	extractPayload(t, 200, recorder, addr)
	validateAddress(t, tu.TestAddress, *addr)
}

func queryForAddresses(t *testing.T, ctx context.Context, config *conf.Configuration, id string) []models.Address {
	ctx = kami.SetParam(ctx, "user_id", id)
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlWithUserID, nil)

	NewAPI(config, db, nil).GetAllAddresses(ctx, recorder, req)
	addrs := []models.Address{}
	extractPayload(t, 200, recorder, &addrs)
	return addrs
}
