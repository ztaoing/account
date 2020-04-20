package accounts

import (
	"errors"
	"github.com/shopspring/decimal"
	"github.com/ztaoing/infra/base"
	"go1234.cn/account/services"
	"sync"
)

//应用服务层
var _ services.AccountService = new(accountService)

//避免重复实例化
var once sync.Once

func init() {
	once.Do(func() {
		services.InterfaceAccountService = new(accountService)
	})
}

type accountService struct {
}

func (a *accountService) CreateAccount(dto services.AccountCreateDTO) (*services.AccountDTO, error) {
	domain := accountDomain{}
	//验证输入参数
	if err := base.ValidateStructs(&dto); err != nil {
		return nil, err
	}
	amount, err := decimal.NewFromString(dto.Amount)
	//执行账户创建的业务逻辑
	account := services.AccountDTO{
		AccountName:  dto.AccountName,
		AccountType:  dto.AccountType,
		CurrencyCode: dto.CurrencyCode,
		UserId:       dto.UserId,
		Username:     dto.Username,
		Balance:      amount,
		Status:       1,
	}
	accountDto, err := domain.CreateAccount(account)
	return accountDto, err
}

//转账接口
func (a *accountService) Transfer(dto services.AccountTransferDTO) (services.TransferedStatus, error) {
	//验证参数
	domain := accountDomain{}
	//验证输入参数
	if err := base.ValidateStructs(&dto); err != nil {
		return services.TransferedStatusFailure, err
	}

	//转换数据
	amount, err := decimal.NewFromString(dto.AmountStr)
	if err != nil {
		return services.TransferedStatusFailure, err
	}
	//转换成功
	dto.Amount = amount
	//验证change_flag
	if dto.ChangeFlag == services.FlagTransferOut {
		if dto.ChangeType > 0 {
			return services.TransferedStatusFailure, errors.New("如果changeFlag为支出，那么changeType必须小于0")
		}
	} else {
		//changeFlag为输入
		if dto.ChangeType < 0 {
			return services.TransferedStatusFailure, errors.New("如果changeFlag为输入，那么changeType必须大于0")
		}
	}
	//完成验证，执行转账
	status, err := domain.Transfer(dto)
	return status, err
}

//储值接口

func (a *accountService) StoreValue(dto services.AccountTransferDTO) (services.TransferedStatus, error) {
	dto.TradeTarget = dto.TradeBody
	dto.ChangeFlag = services.FlagTransferIn //储值
	dto.ChangeType = services.AccountStoreValue
	return a.Transfer(dto)
}

//查询红包账户接口
func (a *accountService) GetEnvelopeAccountByUserId(userId string) *services.AccountDTO {
	domain := accountDomain{}
	account := domain.GetEnvelopeAccountByUserId(userId)
	return account
}

//查询账户信息
func (a *accountService) GetAccount(accountNo string) *services.AccountDTO {
	domain := accountDomain{}
	return domain.GetAccount(accountNo)
}
