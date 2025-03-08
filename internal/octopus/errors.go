package octopus

import "errors"

// This error is returned if we have received a "Too many requests" response from the API.
var ErrTooManyRequests = errors.New("Too many requests")
// This error is returned if we did not make an API request, because we previously
// got a "Too many requests" response.
var ErrSkippingRequest = errors.New("Skipping API request because too many requests")
