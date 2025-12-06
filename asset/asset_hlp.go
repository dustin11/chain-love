package asset

func Restore() {
	//go-bindata -o=asset/asset.go -pkg=asset conf/... docs/... statics/...
	dirs := []string{"asset/conf", "docs", "asset/statics"} // 设置需要释放的目录
	isSuccess := true
	for _, dir := range dirs {
		// 解压dir目录到当前目录
		if err := RestoreAssets("./", dir); err != nil {
			isSuccess = false
			break
		}
	}
	if !isSuccess {
		//for _, dir := range dirs {
		//	os.RemoveAll(filepath.Join("./", dir))
		//}
	}
}
