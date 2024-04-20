package main

func exampleNullObject() {
	println("exampleNullObject")

	// 查询到一个存在的用户
	u := queryUser(1)
	//println(u.Id(), u.Name(), u.Gender())
	assertEqual(u.IsNull(), false)

	// 查询到一个不存在的用户
	u = queryUser(0)
	//println(u.Id(), u.Name(), u.Gender())
	assertEqual(u.IsNull(), true)
}

// UserAPI 定义了用户接口
type UserAPI interface {
	Name() string
	Id() int
	Gender() string
	IsNull() bool // 可提供一个方法判断是否为有效用户
}

// NullUserImpl 实现了用户接口
var _ UserAPI = (*NullUserImpl)(nil)

type NullUserImpl struct {
}

func (n *NullUserImpl) Id() int {
	return 0
}

func (n *NullUserImpl) Name() string {
	return "NullUser"
}

func (n *NullUserImpl) Gender() string {
	return "unknown"
}

func (n *NullUserImpl) IsNull() bool {
	return true
}

func NewNullUser() UserAPI {
	return &NullUserImpl{}
}

// NormalUser 是一个存在的用户，也实现了用户接口
type NormalUser struct {
	id     int
	name   string
	gender string
}

func (n *NormalUser) Id() int {
	return n.id
}

func (n *NormalUser) Name() string {
	return n.name
}

func (n *NormalUser) Gender() string {
	return n.gender
}

func (n *NormalUser) IsNull() bool {
	return false
}

// QueryUser 查询用户信息函数
func queryUser(uid int) UserAPI {
	if uid > 0 {
		return &NormalUser{
			id:     uid,
			name:   "some name",
			gender: "some gender",
		}
	}
	return NewNullUser()
}
