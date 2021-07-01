package fileWatch

type InFileInfoInter interface {
	GetFullPath() string
}

type OutFileInfoInter interface {
	InFileInfoInter
	GetSize() int64
}

type FileInfo struct {
	path string
}

func NewFileInfo(p string) FileInfo {
	return FileInfo{
		path: p,
	}
}

func NewOutFileInfo(p string, s int64) OutFileInfo {
	return OutFileInfo{
		path: p,
		size: s,
	}
}

func (f FileInfo) GetFullPath() string {
	return f.path
}

type OutFileInfo struct {
	path string
	size int64
}

func (f OutFileInfo) GetFullPath() string {
	return f.path
}

func (f OutFileInfo) GetSize() int64 {
	return f.size
}
