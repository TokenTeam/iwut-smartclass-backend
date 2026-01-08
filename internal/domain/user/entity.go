package user

// User 用户实体
type User struct {
	Account  string
	ID       int
	Phone    string
	TenantID int
}

// ReversePhone 反转手机号
func (u *User) ReversePhone() string {
	runes := []rune(u.Phone)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
