package elemwatch

type ElemInter interface {
	GetToken() string
}

type ElemInfo struct {
	token string
}

func NewElemInfo(p string) ElemInfo {
	return ElemInfo{
		token: p,
	}
}

func (f ElemInfo) GetToken() string {
	return f.token
}
