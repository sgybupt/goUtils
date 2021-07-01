package elemWatch

type InFileInfoInter interface {
	GetToken() string
}

type OutFileInfoInter interface {
	InFileInfoInter
	GetChange() interface{}
}

type ElemInfo struct {
	token string
}

func NewElemInfo(p string) ElemInfo {
	return ElemInfo{
		token: p,
	}
}

func NewOutElemInfo(p string, s interface{}) OutElemInfo {
	return OutElemInfo{
		path:        p,
		changeInter: s,
	}
}

func (f ElemInfo) GetToken() string {
	return f.token
}

type OutElemInfo struct {
	path        string
	changeInter interface{}
}

func (f OutElemInfo) GetToken() string {
	return f.path
}

func (f OutElemInfo) GetChange() interface{} {
	return f.changeInter
}
