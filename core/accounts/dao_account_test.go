package accounts

import (
	"database/sql"
	"github.com/segmentio/ksuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/tietang/dbx"
	"github.com/ztaoing/infra/base"
	_ "go1234.cn/account/testx"
	"testing"
)

func TestAccountDao_GetOne(t *testing.T) {
	err := base.Tx(func(runner *dbx.TxRunner) error {
		dao := &AccountDao{runner: runner}
		Convey("通过编号查询账户数据", t, func() {
			a := &Account{
				Balance:     decimal.NewFromFloat(100),
				Status:      1,
				AccountNo:   ksuid.New().String(),
				AccountName: "测试-查询账户信息",
				UserId:      ksuid.New().Next().String(),
				Username:    sql.NullString{String: "测试用户", Valid: true},
			}
			id, err := dao.Insert(a)
			//验证结果
			So(id, ShouldBeGreaterThan, 0)
			So(err, ShouldBeNil)

			new_a := dao.GetOne(a.AccountNo)
			//验证结果
			So(new_a, ShouldNotBeNil)
			So(new_a.Balance.String(), ShouldEqual, a.Balance.String())
			So(new_a.CreateAt, ShouldNotBeNil)
			So(new_a.UpdateAt, ShouldNotBeNil)
		})
		return nil
	})
	if err != nil {
		logrus.Error(err)
	}
}
