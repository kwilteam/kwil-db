package dto

type EstimateRequest struct {
	Payload []byte `json:"payload"`
}

func (e *EstimateRequest) Validate() error {

	return nil
}
