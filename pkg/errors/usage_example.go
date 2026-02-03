package errors

import (
	"database/sql"
	"fmt"
	"net/http"
)

// 1. Validation Failures (Frontend Form Error)
func RegisterUserHandler(w http.ResponseWriter, r *http.Request) error {
	// Assume we parsed the body...
	username := r.FormValue("username")
	email := r.FormValue("email")

	// Perform validation
	fieldErrors := make(map[string]string)
	if len(username) < 3 {
		fieldErrors["username"] = "Kullanıcı adı en az 3 karakter olmalı."
	}
	if email == "" {
		fieldErrors["email"] = "Email adresi zorunludur."
	}

	// If validation fails, return structured 422 error
	if len(fieldErrors) > 0 {
		return ValidationError("Kayıt formu geçersiz.", fieldErrors)
	}

	return nil
}

// 2. Database/Internal Errors (Secure Handling)
func GetOrderHandler(w http.ResponseWriter, r *http.Request) error {
	orderID := r.URL.Query().Get("id")

	// Simulate Database Call
	err := databaseQuery(orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Case 2a: Not Found (Low severity)
			// User sees: "Sipariş bulunamadı"
			// Log: WARN
			return NotFound("Sipariş sistemi üzerinde bulunamadı.", err)
		}

		// Case 2b: DB Connection Failed (High severity)
		// User sees: "Bir hata oluştu, lütfen daha sonra tekrar deneyin."
		// Log: ERROR with StackTrace and "sql: connection refused"
		return Internal(fmt.Errorf("database query failed: %w", err))
	}

	w.Write([]byte(`{"order_id": 123}`))
	return nil // Success
}

// 3. Business Logic / Panic
func RiskyHandler(w http.ResponseWriter, r *http.Request) error {
	// If this panics, our middleware will catch it, log the stack,
	// and return a refined 500 error to the user.
	var list []string
	fmt.Println(list[0]) // Panic: index out of range

	return nil
}

// --- Mock Database Helper ---
func databaseQuery(id string) error {
	if id == "404" {
		return sql.ErrNoRows
	}
	if id == "500" {
		return fmt.Errorf("connection refused")
	}
	return nil
}

// --- Setup ---
// ExampleMain demonstrates how to use the error handling system
// Note: This is an example function, not a real main() since we're in the errors package
func ExampleMain() {
	mux := http.NewServeMux()

	// Apply wrapper to all handlers
	mux.HandleFunc("/register", ErrorWrapper(RegisterUserHandler))
	mux.HandleFunc("/order", ErrorWrapper(GetOrderHandler))
	mux.HandleFunc("/risky", ErrorWrapper(RiskyHandler))

	http.ListenAndServe(":8080", mux)
}
