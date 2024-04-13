package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SergeyMilch/bank-service/internal/mocks"
	"github.com/SergeyMilch/bank-service/middleware"
	"github.com/SergeyMilch/bank-service/pkg/models"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// Тест на обработку неверного JSON
func TestHandleBankRequest_InvalidJSON(t *testing.T) {
    logger := zaptest.NewLogger(t)
    defer logger.Sync()

    e := echo.New()
    request := []byte(`{"id":1, "sum":bad}`)

    req := httptest.NewRequest(http.MethodPost, "/bank", bytes.NewBuffer(request))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockDBService := mocks.NewMockDBService(ctrl)
    handler := NewBankHandler(mockDBService, logger)

    err := handler.HandleBankRequest(c)
    if err != nil {
        httpError, ok := err.(*echo.HTTPError)
        if ok {
            c.JSON(httpError.Code, httpError.Message)
        } else {
            t.Errorf("Expected HTTPError, got %v", err)
        }
    }

    // Проверяем, что логгер был вызван с ожидаемым сообщением
    logger.Check(zap.ErrorLevel, "Error binding request").Write()
    // Проверяем, что статус код в rec установлен корректно
    assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// Тест на проверку роли
func TestHandleBankRequest_InvalidRole(t *testing.T) {
    e := echo.New()
    e.Use(middleware.UserRoleMiddleware) // Добавляем middleware, который проверяет роли

    request := []byte(`{"id":1, "sum":100}`)

    req := httptest.NewRequest(http.MethodPost, "/bank", bytes.NewBuffer(request))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Role", "guest") // Недопустимая роль

    rec := httptest.NewRecorder()

    // Обрабатываем запрос через Echo, который использует middleware
    e.ServeHTTP(rec, req)

    // Проверяем, что статус код и сообщение в ответе установлены корректно
    assert.Equal(t, http.StatusForbidden, rec.Code)
    assert.Contains(t, rec.Body.String(), "access denied")  // Проверяем текст сообщения
}

// Тест на недостаточные средства
func TestHandleBankRequest_InsufficientFunds(t *testing.T) {
    logger := zaptest.NewLogger(t)
    defer logger.Sync()

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockDBService := mocks.NewMockDBService(ctrl)
    mockTx := mocks.NewMockTransaction(ctrl)

    mockDBService.EXPECT().Begin(gomock.Any()).Return(mockTx, nil)
    mockDBService.EXPECT().GetAccount(gomock.Any(), mockTx, 1).Return(&models.Account{ID: 1, Balance: 100}, nil)
    // Не ожидаем вызов UpdateAccount, потому что баланс недостаточен
    mockTx.EXPECT().Rollback(gomock.Any()).Return(nil)

    e := echo.New()
    req := httptest.NewRequest(http.MethodPost, "/bank", bytes.NewBufferString(`{"id":1, "sum":150}`))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Role", "admin")

    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    c.Set("role", "admin")

    h := NewBankHandler(mockDBService, logger)
    err := h.HandleBankRequest(c)
    
    // Проверяем, что логгер был вызван с ожидаемым сообщением
    logger.Check(zap.ErrorLevel, "Error binding request").Write()
    // Ожидаем ошибку и проверяем ее
    if assert.Error(t, err, "Expected an error for insufficient funds") {
        httpError, ok := err.(*echo.HTTPError)
        if assert.True(t, ok, "Error should be of type *echo.HTTPError") {
            assert.Equal(t, http.StatusForbidden, httpError.Code, "Expected HTTP status 403 for insufficient funds")
            assert.Contains(t, httpError.Message.(string), "Insufficient funds", "Expected error message to contain 'Insufficient funds'")
        }
    }
}


