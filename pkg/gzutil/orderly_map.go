package gzutil

// 有序 Map 严格按照Append的顺序执行, 先进先出, 同名会被覆盖
type OrderlyMap struct {
	funcMap  map[string]func()
	nameList []string
}

func NewOrderlyMap() *OrderlyMap {
	return &OrderlyMap{
		funcMap:  make(map[string]func()),
		nameList: make([]string, 0),
	}
}

func (self *OrderlyMap) Append(funcName string, value func()) {
	if _, exists := self.funcMap[funcName]; !exists {
		self.nameList = append(self.nameList, funcName)
	}
	self.funcMap[funcName] = value
}

func (self *OrderlyMap) Foreach() {
	if self == nil || len(self.nameList) == 0 {
		return
	}

	for _, funcName := range self.nameList {
		self.funcMap[funcName]()
	}
}
