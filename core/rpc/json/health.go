package jsonrpc

import "encoding/json"

// HealthResponse is the service-wide aggregate health response. For the shape
// of an individual service's health response, see the analogous struct in their
// respective packages.
//
// The Alive field indicates that the HTTP server is responsive, and will always
// be true. This provides assurance that the response is from the kwild HTTP
// server rather than some other server.
//
// The Healthy field is true if *all* registered services report that they are
// healthy. Details and any additional information from the services are
// included in the Services map (marshalled as a JSON object with structure
// determined by the service).
//
// For example, if the HTTP server has registered JSON-RPC services for the
// "user", "admin", and "function" services:
//
//	{
//	    "kwil_alive": true,
//	    "healthy": false,
//	    "services": [
//	        "user": {
//	            "chain_id": "chain_asdf",
//	            "height": 1234,
//	            "block_age": 12341322,
//				...
//	        },
//	        "admin": {
//	            "is_validator": true,
//	        },
//	        "function": "alive", // not all services have much to say
//	    ]
//	}
type HealthResponse struct {
	// Alive will always be true in a response from the Kwil HTTP server. This
	// ensures that a health check is actually receiving a response from kwild
	// and not some other page, proxy, or misconfigured handler.
	Alive bool `json:"kwil_alive"`

	// Healthy indicates if all services are healthy, as reported by the
	// services themselves.
	Healthy bool `json:"healthy"`

	// Services provides the details of each service's health. Unmarshalling of
	// the responses should be done according to the health response defined by
	// the individual services.
	Services map[string]json.RawMessage `json:"services"` // shape of each service response may be different
}
