package common

import (
	"net/http"
)

// HXRedirect sets the HX-Redirect header to trigger a client-side redirect.
func HXRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Redirect", url)
}

// HXTrigger sets the HX-Trigger header to trigger client-side events.
func HXTrigger(w http.ResponseWriter, event string) {
	w.Header().Set("HX-Trigger", event)
}

// HXReswap sets the HX-Reswap header.
func HXReswap(w http.ResponseWriter, swap string) {
	w.Header().Set("HX-Reswap", swap)
}
