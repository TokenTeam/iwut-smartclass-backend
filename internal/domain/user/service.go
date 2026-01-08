package user

// ExternalService 外部用户服务接口
type ExternalService interface {
	// GetUserInfo 获取用户信息
	GetUserInfo(token string) (*User, error)
}
