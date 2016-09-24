package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/guregu/kami"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"

	"github.com/netlify/gocommerce/models"
	tu "github.com/netlify/gocommerce/testutils"
)

// -------------------------------------------------------------------------------------------------------------------
// LIST
// -------------------------------------------------------------------------------------------------------------------

func TestOrderQueryForAllOrdersAsTheUser(t *testing.T) {
	db := db(t)
	assert := assert.New(t)

	config := testConfig()
	ctx := testContext(token(tu.TestUser.ID, tu.TestUser.Email, nil), config)
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "https://not-real", nil)

	api := NewAPI(config, db, nil)
	api.OrderList(ctx, recorder, req)

	orders := extractOrders(t, 200, recorder)
	assert.Equal(2, len(orders))

	for _, o := range orders {
		switch o.ID {
		case tu.FirstOrder.ID:
			validateOrder(t, tu.FirstOrder, &o)
			validateAddress(t, tu.FirstOrder.BillingAddress, o.BillingAddress)
			validateAddress(t, tu.FirstOrder.ShippingAddress, o.ShippingAddress)
		case tu.SecondOrder.ID:
			validateOrder(t, tu.SecondOrder, &o)
			validateAddress(t, tu.SecondOrder.BillingAddress, o.BillingAddress)
			validateAddress(t, tu.SecondOrder.ShippingAddress, o.ShippingAddress)
		default:
			assert.Fail(fmt.Sprintf("unexpected order: %+v\n", o))
		}
	}
}

func TestOrderQueryForAllOrdersAsAdmin(t *testing.T) {
	db := db(t)
	assert := assert.New(t)

	config := testConfig()
	config.JWT.AdminGroupName = "admin"
	ctx := testContext(token("admin-yo", "admin@wayneindustries.com", &[]string{"admin"}), config)
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlWithUserID, nil)

	api := NewAPI(config, db, nil)
	api.OrderList(ctx, recorder, req)
	orders := extractOrders(t, 200, recorder)

	assert.Equal(2, len(orders))
	for _, o := range orders {
		switch o.ID {
		case tu.FirstOrder.ID:
			validateOrder(t, tu.FirstOrder, &o)
			validateAddress(t, tu.FirstOrder.BillingAddress, o.BillingAddress)
			validateAddress(t, tu.FirstOrder.ShippingAddress, o.ShippingAddress)
		case tu.SecondOrder.ID:
			validateOrder(t, tu.SecondOrder, &o)
			validateAddress(t, tu.SecondOrder.BillingAddress, o.BillingAddress)
			validateAddress(t, tu.SecondOrder.ShippingAddress, o.ShippingAddress)
		default:
			assert.Fail(fmt.Sprintf("unexpected order: %+v\n", o))
		}
	}
}

func TestOrderQueryForAllOrdersAsStranger(t *testing.T) {
	db := db(t)
	assert := assert.New(t)

	config := testConfig()
	ctx := testContext(token("stranger", "stranger-danger@wayneindustries.com", nil), config)
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "https://not-real", nil)

	api := NewAPI(config, db, nil)
	api.OrderList(ctx, recorder, req)
	assert.Equal(200, recorder.Code)
	assert.Equal("[]\n", recorder.Body.String())
}

func TestOrderQueryForAllOrdersNotWithAdminRights(t *testing.T) {
	db := db(t)
	assert := assert.New(t)
	config := testConfig()
	ctx := testContext(token("stranger", "stranger-danger@wayneindustries.com", nil), config)

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlWithUserID, nil)

	api := NewAPI(config, db, nil)
	api.OrderList(ctx, recorder, req)
	assert.Equal(400, recorder.Code)
	validateError(t, 400, recorder)
}

func TestOrderQueryForAllOrdersWithNoToken(t *testing.T) {
	assert := assert.New(t)
	config := testConfig()
	ctx := testContext(nil, config)

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlWithUserID, nil)

	api := NewAPI(config, nil, nil)
	api.OrderList(ctx, recorder, req)
	assert.Equal(401, recorder.Code)
	validateError(t, 401, recorder)
}

// -------------------------------------------------------------------------------------------------------------------
// VIEW
// -------------------------------------------------------------------------------------------------------------------

func TestOrderQueryForAnOrderAsTheUser(t *testing.T) {
	db := db(t)
	config := testConfig()
	ctx := testContext(token(tu.TestUser.ID, "marp@wayneindustries.com", nil), config)

	// have to add it to the context ~ it isn't from the params
	ctx = kami.SetParam(ctx, "id", tu.FirstOrder.ID)

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "https://not-real/"+tu.FirstOrder.ID, nil)

	api := NewAPI(config, db, nil)
	api.OrderView(ctx, recorder, req)
	order := extractOrder(t, 200, recorder)
	validateOrder(t, tu.FirstOrder, order)
	validateAddress(t, tu.FirstOrder.BillingAddress, order.BillingAddress)
	validateAddress(t, tu.FirstOrder.ShippingAddress, order.ShippingAddress)
}

func TestOrderQueryForAnOrderAsAnAdmin(t *testing.T) {
	db := db(t)
	config := testConfig()
	config.JWT.AdminGroupName = "admin"
	ctx := testContext(token("admin-yo", "admin@wayneindustries.com", &[]string{"admin"}), config)

	// have to add it to the context ~ it isn't from the params
	ctx = kami.SetParam(ctx, "id", tu.FirstOrder.ID)

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlForFirstOrder, nil)

	api := NewAPI(config, db, nil)
	api.OrderView(ctx, recorder, req)
	order := extractOrder(t, 200, recorder)
	validateOrder(t, tu.FirstOrder, order)
	validateAddress(t, tu.FirstOrder.BillingAddress, order.BillingAddress)
	validateAddress(t, tu.FirstOrder.ShippingAddress, order.ShippingAddress)
}

func TestOrderQueryForAnOrderAsAStranger(t *testing.T) {
	db := db(t)
	assert := assert.New(t)
	config := testConfig()
	ctx := testContext(token("stranger", "stranger-danger@wayneindustries.com", nil), config)

	// have to add it to the context ~ it isn't from the params
	ctx = kami.SetParam(ctx, "id", tu.FirstOrder.ID)

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", urlForFirstOrder, nil)

	api := NewAPI(config, db, nil)
	api.OrderView(ctx, recorder, req)
	assert.Equal(401, recorder.Code)
	validateError(t, 401, recorder)
}

func TestOrderQueryForAMissingOrder(t *testing.T) {
	db := db(t)
	assert := assert.New(t)
	config := testConfig()
	ctx := testContext(token("stranger", "stranger-danger@wayneindustries.com", nil), config)

	// have to add it to the context ~ it isn't from the params
	ctx = kami.SetParam(ctx, "id", "does-not-exist")

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "https://not-real/does-not-exist", nil)

	api := NewAPI(config, db, nil)
	api.OrderView(ctx, recorder, req)
	assert.Equal(404, recorder.Code)
	validateError(t, 404, recorder)
}

func TestOrderQueryForAnOrderWithNoToken(t *testing.T) {
	assert := assert.New(t)
	config := testConfig()
	ctx := testContext(nil, config)

	// have to add it to the context ~ it isn't from the params
	ctx = kami.SetParam(ctx, "id", "does-not-exist")

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "https://not-real/does-not-exist", nil)

	// use nil for DB b/c it should *NEVER* be called
	api := NewAPI(config, nil, nil)
	api.OrderView(ctx, recorder, req)
	assert.Equal(401, recorder.Code)
	validateError(t, 401, recorder)
}

// -------------------------------------------------------------------------------------------------------------------
// CREATE
// -------------------------------------------------------------------------------------------------------------------
// TODO vvvv ~ need to make it verifiable
//func TestCreateAnOrderAsAnExistingUser(t *testing.T) {
//	assert := assert.New(t)
//	orderRequest := &OrderParams{
//		SessionID: "test-session",
//		LineItems: []*OrderLineItem{&OrderLineItem{
//			SKU:      "nananana",
//			Path:     "/fashion/utility-belt",
//			Quantity: 1,
//		}},
//		BillingAddress: &testAddress,
//		ShippingAddress: &models.Address{
//			LastName: "robin",
//			Address1: "123456 circus lane",
//			Country:  "dcland",
//			City:     "gotham",
//			Zip:      "234789",
//		},
//	}
//
//	bs, err := json.Marshal(orderRequest)
//	if !assert.NoError(t, err) {
//		assert.FailNow(t, "setup failure")
//	}
//
//	ctx := testContext(token(tu.TestUser.ID, tu.TestUser.Email, nil))
//	recorder := httptest.NewRecorder()
//	req, err := http.NewRequest("PUT", "https://not-real/orders", bytes.NewReader(bs))
//
//	api := NewAPI(config, db, nil)
//	api.OrderCreate(ctx, recorder, req)
//	assert.Equal(200, recorder.Code)
//
//	//ret := new(models.Order)
//	ret := make(map[string]interface{})
//	err = json.Unmarshal(recorder.Body.Bytes(), ret)
//	assert.NoError(err)
//
//	fmt.Printf("%+v\n", ret)
//}

// --------------------------------------------------------------------------------------------------------------------
// Create ~ email logic
// --------------------------------------------------------------------------------------------------------------------
func TestOrderSetUserIDLogic_AnonymousUser(t *testing.T) {
	assert := assert.New(t)
	simpleOrder := models.NewOrder("session", "params@email.com", "usd")
	err := setOrderEmail(nil, simpleOrder, nil, testLogger)
	assert.Nil(err)
	assert.Equal("params@email.com", simpleOrder.Email)
}

func TestOrderSetUserIDLogic_AnonymousUserWithNoEmail(t *testing.T) {
	assert := assert.New(t)
	simpleOrder := models.NewOrder("session", "", "usd")
	err := setOrderEmail(nil, simpleOrder, nil, testLogger)
	if !assert.Error(err) {
		assert.Equal(400, err.Code)
	}
}

func TestOrderSetUserIDLogic_NewUserNoEmailOnRequest(t *testing.T) {
	validateNewUserEmail(
		t,
		models.NewOrder("session", "", "usd"),
		token("alfred", "alfred@wayne.com", nil).Claims.(*JWTClaims),
		"alfred@wayne.com",
		"alfred@wayne.com",
	)
}

func TestOrderSetUserIDLogic_NewUserNoEmailOnClaim(t *testing.T) {
	validateNewUserEmail(
		t,
		models.NewOrder("session", "joker@wayne.com", "usd"),
		token("alfred", "", nil).Claims.(*JWTClaims),
		"",
		"joker@wayne.com",
	)
}

func TestOrderSetUserIDLogic_NewUserAllTheEmails(t *testing.T) {
	validateNewUserEmail(
		t,
		models.NewOrder("session", "joker@wayne.com", "usd"),
		token("alfred", "alfred@wayne.com", nil).Claims.(*JWTClaims),
		"alfred@wayne.com",
		"joker@wayne.com",
	)
}

func TestOrderSetUserIDLogic_NewUserNoEmails(t *testing.T) {
	db := db(t)
	assert := assert.New(t)
	simpleOrder := models.NewOrder("session", "", "usd")
	claims := token("alfred", "", nil).Claims.(*JWTClaims)
	err := setOrderEmail(db, simpleOrder, claims, testLogger)
	if assert.Error(err) {
		assert.Equal(400, err.Code)
	}
}

func TestOrderSetUserIDLogic_KnownUserClaimsOnRequest(t *testing.T) {
	validateExistingUserEmail(
		t,
		db(t),
		models.NewOrder("session", "joker@wayne.com", "usd"),
		token(tu.TestUser.ID, "", nil).Claims.(*JWTClaims),
		"joker@wayne.com",
	)
}

func TestOrderSetUserIDLogic_KnownUserClaimsOnClaim(t *testing.T) {
	validateExistingUserEmail(
		t,
		db(t),
		models.NewOrder("session", "", "usd"),
		token(tu.TestUser.ID, tu.TestUser.Email, nil).Claims.(*JWTClaims),
		tu.TestUser.Email,
	)
}

func TestOrderSetUserIDLogic_KnownUserAllTheEmail(t *testing.T) {
	validateExistingUserEmail(
		t,
		db(t),
		models.NewOrder("session", "joker@wayne.com", "usd"),
		token(tu.TestUser.ID, tu.TestUser.Email, nil).Claims.(*JWTClaims),
		"joker@wayne.com",
	)
}

func TestOrderSetUserIDLogic_KnownUserNoEmail(t *testing.T) {
	validateExistingUserEmail(
		t,
		db(t),
		models.NewOrder("session", "", "usd"),
		token(tu.TestUser.ID, "", nil).Claims.(*JWTClaims),
		tu.TestUser.Email,
	)
}

// --------------------------------------------------------------------------------------------------------------------
// UPDATE
// --------------------------------------------------------------------------------------------------------------------
func TestOrderUpdateFields(t *testing.T) {
	db := db(t)
	defer db.Save(tu.FirstOrder)
	assert := assert.New(t)

	recorder := runUpdate(t, db, tu.FirstOrder, &OrderParams{
		Email:    "mrfreeze@dc.com",
		Currency: "monopoly-dollars",
	})
	rspOrder := extractOrder(t, 200, recorder)

	saved := new(models.Order)
	rsp := db.First(saved, "id = ?", tu.FirstOrder.ID)
	assert.False(rsp.RecordNotFound())

	assert.Equal("mrfreeze@dc.com", rspOrder.Email)
	assert.Equal("monopoly-dollars", saved.Currency)

	// did it get persisted to the db
	assert.Equal("mrfreeze@dc.com", saved.Email)
	assert.Equal("monopoly-dollars", saved.Currency)
	validateOrder(t, saved, rspOrder)

	// should be the only field that has changed ~ check it
	saved.Email = tu.FirstOrder.Email
	saved.Currency = tu.FirstOrder.Currency
	validateOrder(t, tu.FirstOrder, saved)
}

func TestOrderUpdateAddress_ExistingAddress(t *testing.T) {
	db := db(t)
	defer db.Save(tu.FirstOrder)
	assert := assert.New(t)

	newAddr := tu.GetTestAddress()
	newAddr.ID = "new-addr"
	newAddr.UserID = tu.FirstOrder.UserID
	rsp := db.Create(newAddr)
	defer db.Unscoped().Delete(newAddr)

	recorder := runUpdate(t, db, tu.FirstOrder, &OrderParams{
		BillingAddressID: newAddr.ID,
	})
	rspOrder := extractOrder(t, 200, recorder)

	saved := new(models.Order)
	rsp = db.First(saved, "id = ?", tu.FirstOrder.ID)
	assert.False(rsp.RecordNotFound())

	// now we load the addresses
	assert.Equal(saved.BillingAddressID, rspOrder.BillingAddressID)

	savedAddr := &models.Address{ID: saved.BillingAddressID}
	rsp = db.First(savedAddr)
	assert.False(rsp.RecordNotFound())
	defer db.Unscoped().Delete(savedAddr)

	validateAddress(t, *newAddr, *savedAddr)
}

func TestOrderUpdateAddress_NewAddress(t *testing.T) {
	db := db(t)
	defer db.Save(tu.FirstOrder)
	assert := assert.New(t)

	paramsAddress := tu.GetTestAddress()
	recorder := runUpdate(t, db, tu.FirstOrder, &OrderParams{
		// should create a new address associated with the order's user
		ShippingAddress: paramsAddress,
	})
	rspOrder := extractOrder(t, 200, recorder)

	saved := new(models.Order)
	rsp := db.First(saved, "id = ?", tu.FirstOrder.ID)
	assert.False(rsp.RecordNotFound())

	// now we load the addresses
	assert.Equal(saved.ShippingAddressID, rspOrder.ShippingAddressID)

	savedAddr := &models.Address{ID: saved.ShippingAddressID}
	rsp = db.First(savedAddr)
	assert.False(rsp.RecordNotFound())
	defer db.Unscoped().Delete(savedAddr)

	validateAddress(t, *paramsAddress, *savedAddr)
}

func TestOrderUpdatePaymentInfoAfterPaid(t *testing.T) {
	db := db(t)
	defer db.Save(tu.FirstOrder)
	db.Model(tu.FirstOrder).Update("payment_state", "paid")

	recorder := runUpdate(t, db, tu.FirstOrder, &OrderParams{
		Currency: "monopoly",
	})
	validateError(t, 400, recorder)
}

func TestOrderUpdateBillingAddressAfterPaid(t *testing.T) {
	db := db(t)
	defer db.Model(tu.FirstOrder).Update("payment_state", "pending")
	db.Model(tu.FirstOrder).Update("payment_state", "paid")

	recorder := runUpdate(t, db, tu.FirstOrder, &OrderParams{
		BillingAddressID: tu.TestAddress.ID,
	})
	validateError(t, 400, recorder)
}

func TestOrderUpdateShippingAfterShipped(t *testing.T) {
	db := db(t)
	defer db.Model(tu.FirstOrder).Update("fulfillment_state", "pending")
	db.Model(tu.FirstOrder).Update("fulfillment_state", "paid")

	recorder := runUpdate(t, db, tu.FirstOrder, &OrderParams{
		ShippingAddressID: tu.TestAddress.ID,
	})
	validateError(t, 400, recorder)
}

func TestOrderUpdateAsNonAdmin(t *testing.T) {
	db := db(t)
	config := testConfig()
	ctx := testContext(token("villian", "villian@wayneindustries.com", nil), config)
	ctx = kami.SetParam(ctx, "id", tu.FirstOrder.ID)

	params := &OrderParams{
		Email:    "mrfreeze@dc.com",
		Currency: "monopoly-dollars",
	}

	updateBody, err := json.Marshal(params)
	if !assert.NoError(t, err) {
		assert.FailNow(t, "Failed to setup test "+err.Error())
	}

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", urlWithUserID, bytes.NewReader(updateBody))

	api := NewAPI(config, db, nil)
	api.OrderUpdate(ctx, recorder, req)
	validateError(t, 401, recorder)
}

func TestOrderUpdateWithNoCreds(t *testing.T) {
	db := db(t)
	config := testConfig()
	ctx := testContext(nil, config)
	ctx = kami.SetParam(ctx, "id", tu.FirstOrder.ID)

	params := &OrderParams{
		Email:    "mrfreeze@dc.com",
		Currency: "monopoly-dollars",
	}

	updateBody, err := json.Marshal(params)
	if !assert.NoError(t, err) {
		assert.FailNow(t, "Failed to setup test "+err.Error())
	}

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", urlForFirstOrder, bytes.NewReader(updateBody))

	api := NewAPI(config, db, nil)
	api.OrderUpdate(ctx, recorder, req)
	validateError(t, 401, recorder)
}

func TestOrderUpdateWithNewData(t *testing.T) {
	db := db(t)
	assert := assert.New(t)
	defer db.Where("order_id = ?", tu.FirstOrder.ID).Delete(&models.Data{})

	op := &OrderParams{
		Data: map[string]interface{}{
			"thing":       1,
			"red":         "fish",
			"other thing": 3.4,
			"exists":      true,
		},
	}
	recorder := runUpdate(t, db, tu.FirstOrder, op)
	returnedOrder := extractOrder(t, 200, recorder)

	// TODO test that the returned order contains the data we expect
	_ = returnedOrder

	datas := []models.Data{}
	db.Where("order_id = ?", tu.FirstOrder.ID).Find(&datas)
	assert.Len(datas, 4)
	for _, datum := range datas {
		switch datum.Key {
		case "thing":
			assert.Equal(models.NumberType, datum.Type)
			assert.EqualValues(1, datum.NumericValue)
		case "red":
			assert.Equal(models.StringType, datum.Type)
			assert.Equal("fish", datum.StringValue)
		case "other thing":
			assert.Equal(models.NumberType, datum.Type)
			assert.EqualValues(3.4, datum.NumericValue)
		case "exists":
			assert.Equal(models.BoolType, datum.Type)
			assert.Equal(true, datum.BoolValue)
		}
	}
}

func TestOrderUpdateWithBadData(t *testing.T) {
	db := db(t)
	defer db.Where("order_id = ?", tu.FirstOrder.ID).Delete(&models.Data{})

	op := &OrderParams{
		Data: map[string]interface{}{
			"array": []int{4},
		},
	}
	recorder := runUpdate(t, db, tu.FirstOrder, op)
	validateError(t, 400, recorder)
}

// -------------------------------------------------------------------------------------------------------------------
// HELPERS
// -------------------------------------------------------------------------------------------------------------------

func extractOrders(t *testing.T, code int, recorder *httptest.ResponseRecorder) []models.Order {
	assert.Equal(t, code, recorder.Code)
	orders := []models.Order{}
	err := json.NewDecoder(recorder.Body).Decode(&orders)
	assert.Nil(t, err)
	return orders
}

func extractOrder(t *testing.T, code int, recorder *httptest.ResponseRecorder) *models.Order {
	assert.Equal(t, code, recorder.Code)
	order := new(models.Order)
	err := json.NewDecoder(recorder.Body).Decode(order)
	assert.Nil(t, err)
	return order
}

// -------------------------------------------------------------------------------------------------------------------
// VALIDATORS
// -------------------------------------------------------------------------------------------------------------------

func runUpdate(t *testing.T, db *gorm.DB, order *models.Order, params *OrderParams) *httptest.ResponseRecorder {
	config := testConfig()
	config.JWT.AdminGroupName = "admin"
	ctx := testContext(token("admin-yo", "admin@wayneindustries.com", &[]string{"admin"}), config)
	ctx = kami.SetParam(ctx, "id", order.ID)

	updateBody, err := json.Marshal(params)
	if !assert.NoError(t, err) {
		assert.FailNow(t, "Failed to setup test "+err.Error())
	}

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("https://not-real/%s", order.ID), bytes.NewReader(updateBody))

	NewAPI(config, db, nil).OrderUpdate(ctx, recorder, req)
	return recorder
}

func validateOrder(t *testing.T, expected, actual *models.Order) {
	assert := assert.New(t)

	// all the stock fields
	assert.Equal(expected.ID, actual.ID)
	assert.Equal(expected.UserID, actual.UserID)
	assert.Equal(expected.Email, actual.Email)
	assert.Equal(expected.Currency, actual.Currency)
	assert.Equal(expected.Taxes, actual.Taxes)
	assert.Equal(expected.Shipping, actual.Shipping)
	assert.Equal(expected.SubTotal, actual.SubTotal)
	assert.Equal(expected.Total, actual.Total)
	assert.Equal(expected.PaymentState, actual.PaymentState)
	assert.Equal(expected.FulfillmentState, actual.FulfillmentState)
	assert.Equal(expected.State, actual.State)
	assert.Equal(expected.ShippingAddressID, actual.ShippingAddressID)
	assert.Equal(expected.BillingAddressID, actual.BillingAddressID)
	assert.Equal(expected.CreatedAt.Unix(), actual.CreatedAt.Unix())
	assert.Equal(expected.UpdatedAt.Unix(), actual.UpdatedAt.Unix())
	assert.Equal(expected.VATNumber, actual.VATNumber)

	// we don't return the actual user
	assert.Nil(actual.User)

	for _, exp := range expected.LineItems {
		found := false
		for _, act := range expected.LineItems {
			if act.ID == exp.ID {
				found = true
				assert.Equal(exp, act)
			}
		}
		assert.True(found, fmt.Sprintf("Failed to find line item: %d", exp.ID))
	}
}

func validateLineItem(t *testing.T, expected *models.LineItem, actual *models.LineItem) {
	assert := assert.New(t)

	assert.Equal(expected.ID, actual.ID)
	assert.Equal(expected.Title, actual.Title)
	assert.Equal(expected.SKU, actual.SKU)
	assert.Equal(expected.Type, actual.Type)
	assert.Equal(expected.Description, actual.Description)
	assert.Equal(expected.VAT, actual.VAT)
	assert.Equal(expected.Path, actual.Path)
	assert.Equal(expected.Price, actual.Price)
	assert.Equal(expected.Quantity, actual.Quantity)
}

func validateNewUserEmail(t *testing.T, order *models.Order, claims *JWTClaims, expectedUserEmail, expectedOrderEmail string) {
	db := db(t)
	assert := assert.New(t)
	result := db.First(new(models.User), "id = ?", claims.ID)
	if !result.RecordNotFound() {
		assert.FailNow("Unclean test env -- user exists with ID " + claims.ID)
	}

	err := setOrderEmail(db, order, claims, testLogger)
	if assert.NoError(err) {
		user := new(models.User)
		result = db.First(user, "id = ?", claims.ID)
		assert.False(result.RecordNotFound())
		assert.Equal(claims.ID, user.ID)
		assert.Equal(claims.ID, order.UserID)
		assert.Equal(expectedOrderEmail, order.Email)
		assert.Equal(expectedUserEmail, user.Email)

		db.Unscoped().Delete(user)
		t.Logf("Deleted user %s", claims.ID)
	}
}

func validateExistingUserEmail(t *testing.T, db *gorm.DB, order *models.Order, claims *JWTClaims, expectedOrderEmail string) {
	assert := assert.New(t)
	err := setOrderEmail(db, order, claims, testLogger)
	if assert.NoError(err) {
		assert.Equal(claims.ID, order.UserID)
		assert.Equal(expectedOrderEmail, order.Email)
	}
}