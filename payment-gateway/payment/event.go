package payment

import "math/big"

type PaymentCreated struct {
	PaymentId uint64
	Cid       string
	Amount    *big.Int
	ExpiredAt uint64
}

type PaymentConfirmed struct {
	PaymentId uint64
}
