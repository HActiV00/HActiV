package routers

import (
	"server/controllers"
	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	// Dashboard 관련 API 라우팅
	beego.Router("/api/dashboard", &controllers.DashboardController{}, "get:Get;post:Post")

	// WebSocket 연결을 위한 라우트 추가
	beego.Router("/ws", &controllers.DashboardController{}, "get:WebSocketHandler")
	beego.Router("/health", &controllers.HealthController{})
	
	// 정적 파일 서빙 (옵션)
	beego.SetStaticPath("/static", "static")
}

