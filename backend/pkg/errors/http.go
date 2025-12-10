package errors

// import "net/http"

// func HTTPStatus(code int) int {
// 	switch {
// 	case code >= 2000 && code < 2100:
// 		if code == 2000 {
// 			return http.StatusConflict
// 		}
// 		if code == 2001 {
// 			return http.StatusBadRequest
// 		}
// 		if code == 2005 {
// 			return http.StatusNotFound
// 		}
// 		return http.StatusServiceUnavailable
// 	case code >= 2100 && code < 2200:
// 		if code == 2104 {
// 			return http.StatusConflict
// 		}
// 		return http.StatusUnauthorized
// 	case code == 1002:
// 		return http.StatusNotFound
// 	case code == 1003:
// 		return http.StatusUnauthorized
// 	case code == 1004:
// 		return http.StatusForbidden
// 	case code == 1005:
// 		return http.StatusConflict
// 	case code == 1006:
// 		return http.StatusTooManyRequests
// 	case code == 1001:
// 		return http.StatusBadRequest
// 	default:
// 		return http.StatusInternalServerError
// 	}
// }
