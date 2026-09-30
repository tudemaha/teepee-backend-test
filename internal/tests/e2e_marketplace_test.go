package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"github.com/google/uuid"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"gorm.io/gorm"
)

type MarketplaceTestSuite struct {
	suite.Suite
	e  *echo.Echo
	db *gorm.DB

	// State preserved across tests
	BuyerToken  string
	SellerToken string
	AdminToken  string
	CategoryID  string
	ProductID   string
	CartID      string
	OrderID     string
	PaymentID   string
}

func (s *MarketplaceTestSuite) SetupSuite() {
	s.e, s.db = SetupTestServer()
}

// executeRequest is a helper to run http requests against the Echo server
func (s *MarketplaceTestSuite) executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.e.ServeHTTP(rr, req)
	return rr
}

func (s *MarketplaceTestSuite) assertStatus(expected int, res *httptest.ResponseRecorder) {
	if res.Code != expected {
		s.T().Fatalf("Expected status %d but got %d. Body: %s", expected, res.Code, res.Body.String())
	}
}

func (s *MarketplaceTestSuite) Test_01_Auth_RegisterAndLogin() {
	// Register Buyer
	buyerReq := dto.RegisterRequest{
		Name:     "Test Buyer",
		Email:    "buyer@test.com",
		Password: "password123",
		PasswordConfirmation: "password123",
	}
	body, _ := json.Marshal(buyerReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusCreated, res)

	var responseMap map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &responseMap)
	data := responseMap["data"].(map[string]interface{})
	s.BuyerToken = data["access_token"].(string)

	// Register Seller
	sellerReq := dto.RegisterRequest{
		Name:     "Test Seller",
		Email:    "seller@test.com",
		Password: "password123",
		PasswordConfirmation: "password123",
	}
	body, _ = json.Marshal(sellerReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusCreated, res)

	json.Unmarshal(res.Body.Bytes(), &responseMap)
	data = responseMap["data"].(map[string]interface{})
	s.SellerToken = data["access_token"].(string)
}
func (s *MarketplaceTestSuite) Test_02_CreateShop() {
	shopReq := dto.CreateShopRequest{
		Name:    "Test Shop",
		Address: "123 Test Street",
	}
	body, _ := json.Marshal(shopReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shops", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.SellerToken)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusCreated, res)
}

func (s *MarketplaceTestSuite) Test_03_Admin_CreateCategory() {
	// Register Admin
	adminReq := dto.RegisterRequest{
		Name:     "Admin User",
		Email:    "admin@test.com",
		Password: "password123",
		PasswordConfirmation: "password123",
	}
	body, _ := json.Marshal(adminReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusCreated, res)

	// Manually upgrade user to Admin in DB for testing
	s.db.Exec("UPDATE users SET role = 'admin' WHERE email = 'admin@test.com'")

	// Login again to get a fresh token with 'admin' role
	time.Sleep(1 * time.Second)
	loginReq := dto.LoginRequest{
		Email:    "admin@test.com",
		Password: "password123",
	}
	body, _ = json.Marshal(loginReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res = s.executeRequest(req)

	var responseMap map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &responseMap)
	data := responseMap["data"].(map[string]interface{})
	s.AdminToken = data["access_token"].(string)


	// Create Category
	catReq := dto.CreateCategoryRequest{
		Name: "Electronics",
	}
	body, _ = json.Marshal(catReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.AdminToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusCreated, res)

	json.Unmarshal(res.Body.Bytes(), &responseMap)
	catData := responseMap["data"].(map[string]interface{})
	s.CategoryID = catData["id"].(string)
}

func (s *MarketplaceTestSuite) Test_04_CreateProduct() {
	// The seller creates a product, using the SellerToken (which should now have Seller role!)
	prodReq := dto.CreateProductRequest{
		Name:        "Smartphone",
		Description: "Latest model",
		Price:       99900,
		Stock:       10,
		CategoryIDs: []uuid.UUID{uuid.MustParse(s.CategoryID)},
	}
	body, _ := json.Marshal(prodReq)
	
	// Because role is cached in the JWT from Login, and they created a shop *after* login,
	// their JWT might still say "buyer". Wait, if the role was "buyer" during login, it won't update in the JWT until they re-login.
	// Let's re-login the seller to get the upgraded "seller" JWT token!
	time.Sleep(1 * time.Second)
	loginReq := dto.LoginRequest{
		Email:    "seller@test.com",
		Password: "password123",
	}
	loginBody, _ := json.Marshal(loginReq)
	loginReqHTTP := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	loginReqHTTP.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	loginRes := s.executeRequest(loginReqHTTP)
	var loginMap map[string]interface{}
	json.Unmarshal(loginRes.Body.Bytes(), &loginMap)
	s.SellerToken = loginMap["data"].(map[string]interface{})["access_token"].(string)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.SellerToken)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusCreated, res)

	var responseMap map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &responseMap)
	prodData := responseMap["data"].(map[string]interface{})
	s.ProductID = prodData["id"].(string)
}

func (s *MarketplaceTestSuite) Test_04b_ProductUpdatesAndReads() {
	// 1. Update Stock
	stockReq := dto.UpdateProductStockRequest{Stock: 50}
	body, _ := json.Marshal(stockReq)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/"+s.ProductID+"/stock", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.SellerToken)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// 2. Update Availability
	avail := false
	availReq := dto.UpdateProductAvailabilityRequest{Available: &avail}
	body, _ = json.Marshal(availReq)
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/products/"+s.ProductID+"/availability", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.SellerToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// Revert availability to true so checkout test doesn't fail
	avail = true
	availReq = dto.UpdateProductAvailabilityRequest{Available: &avail}
	body, _ = json.Marshal(availReq)
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/products/"+s.ProductID+"/availability", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.SellerToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// 3. Full Update (PUT)
	putReq := dto.UpdateProductRequest{
		Name:        "Updated Smartphone",
		Description: "Updated description",
		Price:       88800,
		Stock:       50,
		CategoryIDs: []uuid.UUID{uuid.MustParse(s.CategoryID)},
	}
	body, _ = json.Marshal(putReq)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/products/"+s.ProductID, bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.SellerToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// 4. GET Product by ID
	req = httptest.NewRequest(http.MethodGet, "/api/v1/products/"+s.ProductID, nil)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// 5. GET All Products
	req = httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)
}

func (s *MarketplaceTestSuite) Test_05_AddToCartAndCheckout() {
	// Buyer adds product to cart
	cartReq := dto.AddToCartRequest{
		ProductID: uuid.MustParse(s.ProductID),
		Quantity:  2,
	}
	body, _ := json.Marshal(cartReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/carts", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusCreated, res)
	
	// Fetch Cart to get the CartID
	req = httptest.NewRequest(http.MethodGet, "/api/v1/carts", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	var responseMap map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &responseMap)
	cartData := responseMap["data"].(map[string]interface{})
	items := cartData["items"].([]interface{})
	firstItem := items[0].(map[string]interface{})
	s.CartID = firstItem["id"].(string)

	// Checkout
	checkoutReq := dto.CheckoutRequest{
		CartItemIDs:     []uuid.UUID{uuid.MustParse(s.CartID)},
		ShippingAddress: "Buyer Address 123",
	}
	body, _ = json.Marshal(checkoutReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/orders/checkout", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusCreated, res)

	json.Unmarshal(res.Body.Bytes(), &responseMap)
	orderData := responseMap["data"].(map[string]interface{})
	s.OrderID = orderData["id"].(string)
}

func (s *MarketplaceTestSuite) Test_06_PayAndCompleteOrder() {
	// Buyer pays for the order
	payReq := dto.CreatePaymentRequest{
		OrderID: uuid.MustParse(s.OrderID),
		Method:  "qris",
	}
	body, _ := json.Marshal(payReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusCreated, res)

	var responseMap map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &responseMap)
	payData := responseMap["data"].(map[string]interface{})
	s.PaymentID = payData["id"].(string)

	// Admin confirms payment
	updatePayReq := dto.UpdatePaymentStatusRequest{
		Status: "paid",
	}
	body, _ = json.Marshal(updatePayReq)
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/payments/"+s.PaymentID+"/status", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.AdminToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// Admin updates order to 'completed'
	updateOrdReq := dto.UpdateOrderStatusRequest{
		Status: "completed",
	}
	body, _ = json.Marshal(updateOrdReq)
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/orders/"+s.OrderID+"/status", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.AdminToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)
}

func (s *MarketplaceTestSuite) Test_07_ReviewProduct() {
	// Buyer reviews the completed product
	reviewReq := dto.CreateReviewRequest{
		ProductID: uuid.MustParse(s.ProductID),
		Rating:    5,
		Comment:   "Excellent smartphone!",
	}
	body, _ := json.Marshal(reviewReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusCreated, res)
}

func (s *MarketplaceTestSuite) Test_08_DeleteProduct() {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/products/"+s.ProductID, nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.SellerToken)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)
}

func TestMarketplaceSuite(t *testing.T) {
	suite.Run(t, new(MarketplaceTestSuite))
}
