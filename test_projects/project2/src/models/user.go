package models

// User represents a system user
type User struct {
    ID   int
    Name string
}

// FIXME: Add validation
func (u *User) Validate() bool {
    return true
}
