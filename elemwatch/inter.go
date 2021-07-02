package elemwatch

type ElemInter interface {
	GetToken() string
}

type ElemInfo struct {
	Token string
}

func NewElemInfo(p string) ElemInfo {
	return ElemInfo{
		Token: p,
	}
}

func (f ElemInfo) GetToken() string {
	return f.Token
}
