package handler

import (
	"net/http"

	"github.com/SergeyMilch/bank-service/internal/db"
	"github.com/SergeyMilch/bank-service/pkg/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type BankHandler struct {
    DB     db.DBService
    Logger *zap.Logger
}

func NewBankHandler(dbService db.DBService, logger *zap.Logger) *BankHandler {
    return &BankHandler{DB: dbService, Logger: logger}
}

func (h *BankHandler) HandleBankRequest(c echo.Context) error {
    var request models.Account
    if err := c.Bind(&request); err != nil {
        h.Logger.Error("Error binding request", zap.Error(err))
        return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON")
    }

    if request.Balance < 0 {
        h.Logger.Info("Invalid request: negative values are not allowed")
        return echo.NewHTTPError(http.StatusBadRequest, "Negative values are not allowed")
    }

    role, ok := c.Get("role").(string)
    if !ok || (role != "admin" && role != "client") {
        h.Logger.Warn("Access denied", zap.String("role", role))
        return echo.NewHTTPError(http.StatusForbidden, "access denied")
    }

    ctx := c.Request().Context()
    tx, err := h.DB.Begin(ctx)
    if err != nil {
        h.Logger.Error("Failed to begin transaction", zap.Error(err))
        return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction")
    }
    defer tx.Rollback(ctx)

    account, err := h.DB.GetAccount(ctx, tx, 1)
    if err != nil {
        h.Logger.Error("Database error", zap.Error(err))
        return echo.NewHTTPError(http.StatusInternalServerError, "database error")
    }

    change := request.Balance
    if role == "admin" {
        // Администраторы снимают деньги
        if account.Balance < change {
            h.Logger.Warn("Insufficient funds", zap.Float64("attempted", change), zap.Float64("balance", account.Balance))
            return echo.NewHTTPError(http.StatusForbidden, "Insufficient funds")
        }
        change = -change
    }

    account.Balance += change
    err = h.DB.UpdateAccount(ctx, tx, account.ID, account.Balance)
    if err != nil {
        h.Logger.Error("Failed to update account", zap.Error(err))
        return echo.NewHTTPError(http.StatusInternalServerError, "failed to update account")
    }

    if err = tx.Commit(ctx); err != nil {
        h.Logger.Error("Failed to commit transaction", zap.Error(err))
        return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit transaction")
    }

    h.Logger.Info("Transaction successful", zap.String("role", role), zap.Float64("change", change), zap.Float64("new_balance", account.Balance))
    return c.JSON(http.StatusOK, account)
}