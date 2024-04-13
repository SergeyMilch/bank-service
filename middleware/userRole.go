package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func UserRoleMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        role := c.Request().Header.Get("User-Role")
        if role == "" || (role != "admin" && role != "client") {
            // Если роль не указана или указана некорректно
            return echo.NewHTTPError(http.StatusForbidden, "access denied")
        }
        c.Set("role", role)
        return next(c)
    }
}
