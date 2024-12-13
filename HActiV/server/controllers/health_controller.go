package controllers

import (
    beego "github.com/beego/beego/v2/server/web"
)

type HealthController struct {
    beego.Controller
}

func (c *HealthController) Get() {
    c.Data["json"] = map[string]string{"status": "ok"}
    c.ServeJSON()
}
