package domain

import "time"

const (
	EventCustomerRegistered  = "customer.registered"
	EventCustomerDeactivated = "customer.deactivated"
)

type CustomerRegistered struct {
	CustomerID string
	Name       string
	Email      string
}

type CustomerDeactivated struct {
	CustomerID    string
	DeactivatedAt time.Time
}
