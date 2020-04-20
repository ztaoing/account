package newResk

import (
	"github.com/ztaoing/infra"
	"github.com/ztaoing/infra/base"
	_ "go1234.cn/account/apis/web"
	_ "go1234.cn/account/core/accounts"
)

func init() {

	infra.Register(&base.PropsStarter{})
	//infra.Register(&base.DbxStarter{})
	infra.Register(&base.ValidatorStart{})
	infra.Register(&base.IrisSveverStarter{})
	infra.Register(&infra.WebApiStart{})
	infra.Register(&base.EurekaStarter{})
	//infra.Register(&base.HookStarter{})
}
