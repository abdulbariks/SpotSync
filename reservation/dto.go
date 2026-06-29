package reservation

type CreateRequest struct {
	ZoneID       uint   `json:"zone_id" validate:"required"`
	LicensePlate string `json:"license_plate" validate:"required,min=2,max=15"`
}

type ReservationResponse struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`
	ZoneID       uint   `json:"zone_id"`
	LicensePlate string `json:"license_plate"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type MyReservationResponse struct {
	ID          uint `json:"id"`
	LicensePlate string `json:"license_plate"`
	Status      string `json:"status"`
	Zone struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"zone"`
	CreatedAt string `json:"created_at"`
}