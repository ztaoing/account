package services

import (
	"github.com/shopspring/decimal"
	"github.com/ztaoing/infra/base"
	"time"
)

var InterfaceAccountService AccountService

//用于对外暴露账户应用服务的唯一暴露点
func GetAccountService() AccountService {
	//检查传入的变量是否为nil,避免因为没有初始化造成的错误
	base.Check(InterfaceAccountService)
	return InterfaceAccountService
}

type AccountService interface {
	CreateAccount(dto AccountCreateDTO) (*AccountDTO, error)     //创建账户接口
	Transfer(dto AccountTransferDTO) (TransferedStatus, error)   //转账接口
	StoreValue(dto AccountTransferDTO) (TransferedStatus, error) //储值接口
	GetEnvelopeAccountByUserId(userId string) *AccountDTO        //查询账户接口
	GetAccount(accountNum string) *AccountDTO                    //查询账户信息
}

//CreateAccount 入参
type AccountCreateDTO struct {
	UserId       string //用户id
	Username     string //用户名称
	AccountName  string //账户名称
	AccountType  int    //账户的类型
	CurrencyCode string //货币类型编码：CNY人民币，EUR欧元，USD美元 。。。
	Amount       string //金额 使用string，防止在传递的过程中丢失精度
}

//CreateAccount 出参
type AccountDTO struct {
	AccountNum   string          //账户编号,账户唯一标识
	AccountName  string          //账户名称,用来说明账户的简短描述,账户对应的名称或者命名，比如xxx积分、xxx零钱
	AccountType  int             //账户类型，用来区分不同类型的账户：积分账户、会员卡账户、钱包账户、红包账户
	CurrencyCode string          //货币类型编码：CNY人民币，EUR欧元，USD美元 。。。
	UserId       string          //用户编号, 账户所属用户
	Username     string          //用户名称
	Balance      decimal.Decimal //账户可用余额
	Status       int             //账户状态，账户状态：0账户初始化，1启用，2停用
	CreatedAt    time.Time       //创建时间
	UpdatedAt    time.Time       //更新时间
}

//Transfer接口的入参
//--转账的入参
type AccountTransferDTO struct {
	TradeNum    string           //交易的编号
	TradeBody   TradePaticipator //交易的主体
	TradeTarget TradePaticipator //交易的目标
	AmountStr   string
	Amount      decimal.Decimal //交易金额
	ChangeType  ChangeType      //交易变化的类型
	ChangeFlag  ChangeFlag      //交易变化的标识
	Desc        string          //交易描述
}

//--交易的参与者
type TradePaticipator struct {
	AccountNum string //交易变化
	UserId     string //交易的id
	Username   string //交易的名称
}

//账户流水
type AccountLogDTO struct {
	LogNum           string          //流水编号 全局不重复字符或数字，唯一性标识
	TradeNum         string          //交易单号 全局不重复字符或数字，唯一性标识
	AccountNum       string          //账户编号 账户ID
	TargetAccountNum string          //账户编号 账户ID
	UserId           string          //用户编号
	Username         string          //用户名称
	TargetUserId     string          //目标用户编号
	TargetUsername   string          //目标用户名称
	Amount           decimal.Decimal //交易金额,该交易涉及的金额
	Balance          decimal.Decimal //交易后余额,该交易后的余额
	ChangeType       ChangeType      //流水交易类型，0 创建账户，>0 为收入类型，<0 为支出类型，自定义
	ChangeFlag       ChangeFlag      //交易变化标识：-1 出账 1为进账，枚举
	Status           int             //交易状态：
	Decs             string          //交易描述
	CreatedAt        time.Time       //创建时间
}
