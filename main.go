package main

import (
	_ "ministry-of-depts-pay/routers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/lib/pq"
)

func main() {
	orm.RegisterDataBase("default", "postgres", "postgres://boffbowsh@localhost/ministry-of-depts-pay?sslmode=disable")

	beego.Run()
}
