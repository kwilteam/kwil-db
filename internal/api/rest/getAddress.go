package rest

import (
	"net/http"
)

// GetAddress handler for getting the address of the node
func (h *Handler) GetAddress(w http.ResponseWriter, r *http.Request) {
	// Get the address from the context
	address := "0x995d95245698212D4Af52c8031F614C3D3127994" // TODO: This should be the address of the node

	// Return the address
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(address))
}
