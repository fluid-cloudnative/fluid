## gohive
	克隆自 https://github.com/derekgr/hivething
    更改了TCLIService, 在inf/目录下，并修改了相应的调用。支持Hive1.2+版本
	支持thrift 0.11版本
## Usage
	package main

	import (
		"fmt"
	
		"github.com/dazheng/gohive"
	)
	
	func main() {
		//	conn, err := gohive.Connect("127.0.0.1:10000", gohive.DefaultOptions) // 无用户名、密码
		conn, err := gohive.ConnectWithUser("127.0.0.1:10000", "username", "password", gohive.DefaultOptions) // 需要用户名、密码
		if err != nil {
			fmt.Errorf("Connect error %v", err)
		}
	
		_, err = conn.Exec("create table if not exists t(c1 int)")
		_, err = conn.Exec(" insert into default.t values(1), (2)")
		if err != nil {
			fmt.Errorf("Connection.Exec error: %v", err)
		}
		rs, err := conn.Query("select c1 from t limit 10")
		if err != nil {
			fmt.Errorf("Connection.Query error: %v", err)
		}
		var c1 int
		for rs.Next() {
			rs.Scan(&c1)
			fmt.Println(c1)
		}
		conn.Close()
	}
