package accounts

import (
	"github.com/shopspring/decimal"
	"go1234.cn/account/services"
	"time"
)

type AccountLog struct {
	Id              int64               `db:"id,omitempty"`
	LogNo           string              `db:"log_no,unique"`       //流水编号， 全局唯一的字符或者数字
	TradeNo         string              `db:"trade_no"`            //交易单号 ，全局唯一的字符或者数字
	AccountNo       string              `db:"account_no"`          //账户id
	UserId          string              `db:"user_id"`             //用户id
	Username        string              `db:"username"`            //用户名称
	TargetAccountNo string              `db:"target_account_no"`   //目标账户编号
	TargetUserId    string              `db:"target_user_id"`      //目标用户编号
	TargetUsername  string              `db:"target_username"`     //目标用户名称
	Amount          decimal.Decimal     `db:"amount"`              //交易金额
	Balance         decimal.Decimal     `db:"balance"`             //交易之后的余额
	ChangeType      services.ChangeType `db:"change_type"`         //交易的类型 ，0 创建账户  >0收入 <0 支出
	ChangeFlag      services.ChangeFlag `db:"change_flag"`         //交易变化的标识 ：-1 出账 1进账  枚举
	Status          int                 `db:"status"`              //交易状态
	Decs            string              `db:"desc"`                //交易描述
	CreatedAt       time.Time           `db:"create_at,omitempty"` //创建时间
}

func (po *AccountLog) FromTransferDTO(dto *services.AccountTransferDTO) {
	po.TradeNo = dto.TradeNo
	po.AccountNo = dto.TradeBody.AccountNo
	po.TargetAccountNo = dto.TradeTarget.AccountNo
	po.UserId = dto.TradeBody.UserId
	po.Username = dto.TradeBody.Username
	po.TargetUserId = dto.TradeTarget.UserId
	po.TargetUsername = dto.TradeTarget.Username
	po.Amount = dto.Amount
	po.ChangeType = dto.ChangeType
	po.ChangeFlag = dto.ChangeFlag
	po.Decs = dto.Desc
}

func (po *AccountLog) ToDTO() *services.AccountLogDTO {
	dto := &services.AccountLogDTO{

		TradeNo:         po.TradeNo,
		LogNo:           po.LogNo,
		AccountNo:       po.AccountNo,
		TargetAccountNo: po.TargetAccountNo,
		UserId:          po.UserId,
		Username:        po.Username,
		TargetUserId:    po.TargetUserId,
		TargetUsername:  po.TargetUsername,
		Amount:          po.Amount,
		Balance:         po.Balance,
		ChangeType:      po.ChangeType,
		ChangeFlag:      po.ChangeFlag,
		Status:          po.Status,
		Decs:            po.Decs,
		CreatedAt:       po.CreatedAt,
	}
	return dto
}
