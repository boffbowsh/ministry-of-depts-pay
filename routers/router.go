package routers

import (
	"ministry-of-depts-pay/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
		beego.Router("/departments", &controllers.DepartmentController{})
		beego.Router("/departments/:id", &controllers.DepartmentController{})
}
