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
		Name:                 "Test Buyer",
		Email:                "buyer@test.com",
		Password:             "password123",
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
		Name:                 "Test Seller",
		Email:                "seller@test.com",
		Password:             "password123",
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


func (s *MarketplaceTestSuite) Test_02b_ShopReadsAndUpdates() {
	var shopID string
	s.db.Raw("SELECT id FROM shops LIMIT 1").Scan(&shopID)

	// GET /shops/:id
	req := httptest.NewRequest(http.MethodGet, "/api/v1/shops/"+shopID, nil)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// PUT /shops/:id
	updateReq := dto.UpdateShopRequest{
		Name:    "Updated Shop Name",
		Address: "456 Updated Street",
	}
	body, _ := json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/shops/"+shopID, bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.SellerToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)
}

func (s *MarketplaceTestSuite) Test_03_Admin_CreateCategory() {
	// Register Admin
	adminReq := dto.RegisterRequest{
		Name:                 "Admin User",
		Email:                "admin@test.com",
		Password:             "password123",
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


func (s *MarketplaceTestSuite) Test_03b_CategoryReadsAndDelete() {
	// GET /categories
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// Create dummy category to delete
	catReq := dto.CreateCategoryRequest{Name: "To Be Deleted"}
	body, _ := json.Marshal(catReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.AdminToken)
	res = s.executeRequest(req)
	
	var responseMap map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &responseMap)
	dummyCatID := responseMap["data"].(map[string]interface{})["id"].(string)

	// DELETE /categories/:id
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/categories/"+dummyCatID, nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.AdminToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)
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
	// their JWT might still say "buyer".
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


func (s *MarketplaceTestSuite) Test_05b_CartUpdatesAndDelete() {
	// Buyer adds dummy product to cart
	cartReq := dto.AddToCartRequest{
		ProductID: uuid.MustParse(s.ProductID),
		Quantity:  1,
	}
	body, _ := json.Marshal(cartReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/carts", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusCreated, res)

	// Fetch new cart item id
	req = httptest.NewRequest(http.MethodGet, "/api/v1/carts", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
	res = s.executeRequest(req)
	var responseMap map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &responseMap)
	items := responseMap["data"].(map[string]interface{})["items"].([]interface{})
	var dummyCartID string
	for _, item := range items {
		i := item.(map[string]interface{})
		if i["id"].(string) != s.CartID { // Skip the checked out one
			dummyCartID = i["id"].(string)
		}
	}

	// PUT /carts/:id
	updateReq := dto.UpdateCartRequest{Quantity: 5}
	body, _ = json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/carts/"+dummyCartID, bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// DELETE /carts/:id
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/carts/"+dummyCartID, nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)
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


func (s *MarketplaceTestSuite) Test_06b_OrderAndPaymentReads() {
	// GET /orders
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// GET /orders/:id
	req = httptest.NewRequest(http.MethodGet, "/api/v1/orders/"+s.OrderID, nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// GET /payments/:orderId
	req = httptest.NewRequest(http.MethodGet, "/api/v1/payments/"+s.OrderID, nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
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



func (s *MarketplaceTestSuite) Test_07b_ReviewReadsAndDelete() {
	// GET /products/:id/reviews
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/"+s.ProductID+"/reviews", nil)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	var reviewID string
	s.db.Raw("SELECT id FROM reviews LIMIT 1").Scan(&reviewID)

	// DELETE /reviews/:id
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/"+reviewID, nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)
}

func (s *MarketplaceTestSuite) Test_08_DeleteProduct() {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/products/"+s.ProductID, nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.SellerToken)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)
}

func (s *MarketplaceTestSuite) Test_09_Auth_Me_And_Logout() {
	// 1. GET /api/v1/auth/me
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.BuyerToken)
	res := s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// Fetch the active refresh token from DB manually for Logout test (since it was hidden in token pair)
	var refreshToken string
	s.db.Raw("SELECT token FROM refresh_tokens LIMIT 1").Scan(&refreshToken)

	
	// POST /api/v1/auth/refresh
	refreshReq := dto.RefreshRequest{RefreshToken: refreshToken}
	body, _ := json.Marshal(refreshReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// Need to fetch the NEW refresh token for logout
	s.db.Raw("SELECT token FROM refresh_tokens LIMIT 1").Scan(&refreshToken)

	// 2. POST /api/v1/auth/logout
	logoutReq := dto.RefreshRequest{
		RefreshToken: refreshToken,
	}
	body, _ = json.Marshal(logoutReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res = s.executeRequest(req)
	s.assertStatus(http.StatusOK, res)

	// Verify it was deleted
	var count int64
	s.db.Table("refresh_tokens").Where("token = ?", refreshToken).Count(&count)
	s.Require().Equal(int64(0), count)
}

func TestMarketplaceSuite(t *testing.T) {
	suite.Run(t, new(MarketplaceTestSuite))
}
