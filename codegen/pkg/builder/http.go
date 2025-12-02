package builder

import (
	"fmt"
	"net/http"
)

func httpMethod(method string) string {
	switch method {
	case http.MethodGet:
		return "http.MethodGet"
	case http.MethodPost:
		return "http.MethodPost"
	case http.MethodPut:
		return "http.MethodPut"
	case http.MethodPatch:
		return "http.MethodPatch"
	case http.MethodDelete:
		return "http.MethodDelete"
	default:
		return method
	}
}

func httpStatusCode(code int) string {
	switch code {
	case http.StatusContinue:
		return "http.StatusContinue"
	case http.StatusSwitchingProtocols:
		return "http.StatusSwitchingProtocols"
	case http.StatusProcessing:
		return "http.StatusProcessing"
	case http.StatusEarlyHints:
		return "http.StatusEarlyHints"
	case http.StatusOK:
		return "http.StatusOK"
	case http.StatusCreated:
		return "http.StatusCreated"
	case http.StatusAccepted:
		return "http.StatusAccepted"
	case http.StatusNonAuthoritativeInfo:
		return "http.StatusNonAuthoritativeInfo"
	case http.StatusNoContent:
		return "http.StatusNoContent"
	case http.StatusResetContent:
		return "http.StatusResetContent"
	case http.StatusPartialContent:
		return "http.StatusPartialContent"
	case http.StatusMultiStatus:
		return "http.StatusMultiStatus"
	case http.StatusAlreadyReported:
		return "http.StatusAlreadyReported"
	case http.StatusIMUsed:
		return "http.StatusIMUsed"
	case http.StatusMultipleChoices:
		return "http.StatusMultipleChoices"
	case http.StatusMovedPermanently:
		return "http.StatusMovedPermanently"
	case http.StatusFound:
		return "http.StatusFound"
	case http.StatusSeeOther:
		return "http.StatusSeeOther"
	case http.StatusNotModified:
		return "http.StatusNotModified"
	case http.StatusUseProxy:
		return "http.StatusUseProxy"
	case http.StatusTemporaryRedirect:
		return "http.StatusTemporaryRedirect"
	case http.StatusPermanentRedirect:
		return "http.StatusPermanentRedirect"
	case http.StatusBadRequest:
		return "http.StatusBadRequest"
	case http.StatusUnauthorized:
		return "http.StatusUnauthorized"
	case http.StatusPaymentRequired:
		return "http.StatusPaymentRequired"
	case http.StatusForbidden:
		return "http.StatusForbidden"
	case http.StatusNotFound:
		return "http.StatusNotFound"
	case http.StatusMethodNotAllowed:
		return "http.StatusMethodNotAllowed"
	case http.StatusNotAcceptable:
		return "http.StatusNotAcceptable"
	case http.StatusProxyAuthRequired:
		return "http.StatusProxyAuthRequired"
	case http.StatusRequestTimeout:
		return "http.StatusRequestTimeout"
	case http.StatusConflict:
		return "http.StatusConflict"
	case http.StatusGone:
		return "http.StatusGone"
	case http.StatusLengthRequired:
		return "http.StatusLengthRequired"
	case http.StatusPreconditionFailed:
		return "http.StatusPreconditionFailed"
	case http.StatusRequestEntityTooLarge:
		return "http.StatusRequestEntityTooLarge"
	case http.StatusRequestURITooLong:
		return "http.StatusRequestURITooLong"
	case http.StatusUnsupportedMediaType:
		return "http.StatusUnsupportedMediaType"
	case http.StatusRequestedRangeNotSatisfiable:
		return "http.StatusRequestedRangeNotSatisfiable"
	case http.StatusExpectationFailed:
		return "http.StatusExpectationFailed"
	case http.StatusTeapot:
		return "http.StatusTeapot"
	case http.StatusMisdirectedRequest:
		return "http.StatusMisdirectedRequest"
	case http.StatusUnprocessableEntity:
		return "http.StatusUnprocessableEntity"
	case http.StatusLocked:
		return "http.StatusLocked"
	case http.StatusFailedDependency:
		return "http.StatusFailedDependency"
	case http.StatusTooEarly:
		return "http.StatusTooEarly"
	case http.StatusUpgradeRequired:
		return "http.StatusUpgradeRequired"
	case http.StatusPreconditionRequired:
		return "http.StatusPreconditionRequired"
	case http.StatusTooManyRequests:
		return "http.StatusTooManyRequests"
	case http.StatusRequestHeaderFieldsTooLarge:
		return "http.StatusRequestHeaderFieldsTooLarge"
	case http.StatusUnavailableForLegalReasons:
		return "http.StatusUnavailableForLegalReasons"
	case http.StatusInternalServerError:
		return "http.StatusInternalServerError"
	case http.StatusNotImplemented:
		return "http.StatusNotImplemented"
	case http.StatusBadGateway:
		return "http.StatusBadGateway"
	case http.StatusServiceUnavailable:
		return "http.StatusServiceUnavailable"
	case http.StatusGatewayTimeout:
		return "http.StatusGatewayTimeout"
	case http.StatusHTTPVersionNotSupported:
		return "http.StatusHTTPVersionNotSupported"
	case http.StatusVariantAlsoNegotiates:
		return "http.StatusVariantAlsoNegotiates"
	case http.StatusInsufficientStorage:
		return "http.StatusInsufficientStorage"
	case http.StatusLoopDetected:
		return "http.StatusLoopDetected"
	case http.StatusNotExtended:
		return "http.StatusNotExtended"
	case http.StatusNetworkAuthenticationRequired:
		return "http.StatusNetworkAuthenticationRequired"
	default:
		panic(fmt.Errorf("invalid status code %d", code))
	}
}
