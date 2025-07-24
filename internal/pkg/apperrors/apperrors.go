package apperrors

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func ApiErrorResponse(eCtx echo.Context, errInput error, statusCode int, message string) error {
	// There really isn't a reason for this function to ever be called with errInput == nil,
	// but we're covering the case for completeness.
	if errInput == nil {
		log.Debug("cannot build an error response from a nil error")
		return eCtx.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	var serviceErr *ServiceError
	if errors.As(errInput, &serviceErr) {
		log.Errorf("service error occurred: %s", serviceErr.Err)
		return eCtx.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	var dbErr *DBError
	if errors.As(errInput, &dbErr) {
		if errors.Is(dbErr, DBErrorNotFound) {
			log.Warn("database resource not found")
			return eCtx.JSON(http.StatusNotFound, map[string]string{"error": "Resource Not Found"})
		}

		log.Errorf("database error occurred: %s", dbErr.Err)
		return eCtx.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	var validationErr validator.ValidationErrors
	if errors.As(errInput, &validationErr) {
		log.Errorf("unprocessable entity: %s", errInput.Error())

		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]string{
			"error": validationErr.Error(),
		})
	}

	var httpErr *HTTPError
	if errors.As(errInput, &httpErr) {
		log.Errorf("HTTP error occurred: %s", httpErr.Err)
		return eCtx.JSON(httpErr.StatusCode, map[string]string{"error": httpErr.Message})
	}

	log.Errorf("uncaught error: %s", errInput.Error())
	return eCtx.JSON(http.StatusInternalServerError, map[string]string{
		"error":   "Internal Server Error",
		"message": message,
	})
}
