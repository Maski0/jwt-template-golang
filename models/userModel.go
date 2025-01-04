package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID            bson.ObjectID `bson:"_id"`
	FirstName     *string       `json:"first_name" validate:"required, min=2, max=100"`
	LastName      *string       `json:"last_name" validate:"required, min=2, max=100"`
	Password      *string       `json:"password" validate:"required, min=6, max=50"`
	Email         *string       `json:"email" validate:"email, required"`
	Phone         *string       `json:"phone" validate:"required"`
	Token         *string       `json:"token"`
	UserType      *string       `json:"user_type" validate:"required, eq=ADMIN | eq=USER"`
	Refresh_token *string       `json:"refresh_token" validate:"email, required"`
	Created_at    time.Time     `json:"created_at"`
	Updated_at    time.Time     `json:"updated_at"`
	User_id       string        `json:"user_id"`
}
