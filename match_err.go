package match

//数据不存在
type NotExist string

func (n NotExist) Error() string {
	if n == "" {
		return "data not exist"
	}

	return string(n)
}

//验证err类型
func IsNotExist(err error) bool {
	if _, ok := err.(NotExist); ok {
		return true
	}

	return false
}

//数据价格为o
type PriceIsZero string

func (z PriceIsZero) Error() string {
	if z == "" {
		return "price is zero"
	}

	return string(z)
}

//验证价格为0的错误
func IsZero(err error) bool {
	if _, ok := err.(PriceIsZero); ok {
		return true
	}

	return false
}

type DataNil string

func (d DataNil) Error() string {
	if d == "" {
		return "is nil"
	}

	return string(d)
}

func IsNil(err error) bool {
	if _, ok := err.(DataNil); ok {
		return true
	}

	return false
}
