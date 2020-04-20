package accounts

import (
	"context"
	"errors"
	"github.com/segmentio/ksuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/tietang/dbx"
	"github.com/ztaoing/account/services"
	"github.com/ztaoing/infra/base"
)

//领域模型是有状态的，每次使用时都要实例化
//只能在accounts包中使用
type accountDomain struct {
	account    Account
	accountLog AccountLog
}

func NewAccountDomain() *accountDomain {
	return new(accountDomain)
}

//创建流水记录
func (domain *accountDomain) creatAccountLog() {
	domain.accountLog = AccountLog{}
	domain.createAccountLogNo()
	domain.accountLog.TradeNo = domain.accountLog.AccountNo

	//流水中的交易主体信息
	domain.accountLog.AccountNo = domain.account.AccountNo
	domain.accountLog.UserId = domain.account.UserId
	domain.accountLog.Username = domain.account.Username.String

	//交易对象信息
	domain.accountLog.TargetAccountNo = domain.account.AccountNo
	domain.accountLog.TargetUserId = domain.account.UserId
	domain.accountLog.TargetUsername = domain.account.Username.String

	//交易金额
	domain.accountLog.Amount = domain.account.Balance  //交易的金额
	domain.accountLog.Balance = domain.account.Balance //交易之后的余额

	//交易变化属性
	domain.accountLog.Decs = "创建账户"
	domain.accountLog.ChangeType = services.AccountCreated
	domain.accountLog.ChangeFlag = services.FlagAccountCreated
}

//创建账户
func (domain *accountDomain) CreateAccount(dto services.AccountDTO) (*services.AccountDTO, error) {
	//创建账户持久化对象
	domain.account = Account{}
	domain.account.FromDTO(&dto) //转换
	domain.createAccountNo()
	domain.account.Username.Valid = true //为true时才向数据库写入

	//创建账户流水的持久化对象
	domain.createAccountLogNo()
	accountDao := AccountDao{}
	accountLogDao := AccountLogDao{}
	var rdto *services.AccountDTO
	//快捷的事务函数，返回为非nil则会回滚
	err := base.Tx(func(runner *dbx.TxRunner) error {
		accountDao.runner = runner
		accountLogDao.runner = runner
		//插入账户数据，然后插入流水数据
		id, err := accountDao.Insert(&domain.account)
		if err != nil {
			return nil
		}
		if id <= 0 {
			return errors.New("创建账户失败")
		}
		//插入流水数据
		id, err = accountLogDao.Insert(&domain.accountLog)
		if err != nil {
			return nil
		}
		if id <= 0 {
			return errors.New("创建流水失败")
		}
		//通过账户编号查出账户信息
		domain.account = *accountDao.GetOne(domain.account.AccountNo)
		return nil
	})
	//转换成DTO对象
	rdto = domain.account.ToDTO()
	return rdto, err
}

//创建流水logNo
func (domain *accountDomain) createAccountLogNo() {
	//暂时使用ksuid生成ID
	//后期需要使用优化成分布式ID
	//全局唯一的ID
	domain.accountLog.LogNo = ksuid.New().Next().String()
}

//生成账户accountNo
func (domain *accountDomain) createAccountNo() {
	domain.account.AccountNo = ksuid.New().Next().String()
}

//转账业务
func (a *accountDomain) Transfer(dto services.AccountTransferDTO) (status services.TransferedStatus, err error) {
	base.Tx(func(runner *dbx.TxRunner) error {
		//把事务绑定到上下文对象汇总
		ctx := base.WithValueContext(context.Background(), runner)
		status, err := a.TransferWithContext(ctx, dto)
		return err
	})
}

//必须在base.Tx事务块里运行，不能单独运行，因为此方法中没有单独的事务
func (a *accountDomain) TransferWithContext(ctx context.Context, dto services.AccountTransferDTO) (status services.TransferedStatus, err error) {
	//如果交易变化是支出，则amount是负数
	amount := dto.Amount
	if dto.ChangeFlag == services.FlagTransferOut {
		amount = amount.Mul(decimal.NewFromFloat(-1))
	}
	//创建账户流水记录
	a.accountLog = AccountLog{}
	a.accountLog.FromTransferDTO(&dto)
	//检查余额是否足够，更新余额
	//在上下文中传递runner，使多个方法在一个事务中执行
	//ExecuteContext并没有开启事务，事务是从上下文context传递过来的
	err = base.ExecuteContext(ctx, func(runner *dbx.TxRunner) error {
		accountDao := AccountDao{runner: runner}
		accountLogDao := AccountLogDao{runner: runner}
		//更新余额时，检查余额
		rows, err := accountDao.UpdateBalance(dto.TradeBody.AccountNo, amount)
		if err != nil {
			status = services.TransferedStatusFailure
			return err
		}
		//只在扣减时才校验余额是否足够
		if rows <= 0 && dto.ChangeFlag == services.FlagTransferOut {
			//余额不足
			status = services.TransferedStatusSufficientFunds
			return errors.New("余额不足，扣减失败")
		}
		//执行成功
		//写入流水记录
		account := accountDao.GetOne(dto.TradeBody.AccountNo)
		if account == nil {
			//没有查询到记录
			return errors.New("未查询到记录，账户出错")
		}
		a.account = *account //取指针的值
		a.accountLog.Balance = a.account.Balance
		id, err := accountLogDao.Insert(&a.accountLog)
		if err != nil || id <= 0 {
			status = services.TransferedStatusFailure
			return errors.New("账户流水创建失败")
		}

		return nil
	})
	if err != nil {
		logrus.Error(err)
	}
	status = services.TransferedStatusSuccess
	return status, err
}

//根据账户编号查询账户信息
func (a *accountDomain) GetAccount(accountNo string) *services.AccountDTO {
	accountDao := AccountDao{}
	var account *Account
	err := base.Tx(func(runner *dbx.TxRunner) error {
		accountDao.runner = runner
		account = accountDao.GetOne(accountNo)
		return nil
	})
	if err != nil {
		return nil
	}
	if account == nil {
		return nil
	}
	return account.ToDTO()
}

//根据用户ID查询红包账户信息
func (a *accountDomain) GetEnvelopeAccountByUserId(userId string) *services.AccountDTO {
	accountDao := AccountDao{}
	var account *Account
	err := base.Tx(func(runner *dbx.TxRunner) error {
		accountDao.runner = runner
		account = accountDao.GetByUserId(userId, int(services.EnvelopeAccountType))
		return nil
	})
	if err != nil {
		return nil
	}
	if account == nil {
		return nil
	}
	return account.ToDTO()
}

//根据流水ID账户流水查询
func (a *accountDomain) GetAccountLog(logNo string) *services.AccountLogDTO {
	accountLogDao := AccountLogDao{}
	var log *AccountLog
	err := base.Tx(func(runner *dbx.TxRunner) error {
		accountLogDao.runner = runner
		log = accountLogDao.GetOne(logNo)
		return nil
	})
	if err != nil {
		logrus.Error(err)
		return nil
	}
	if log == nil {
		return nil
	}
	return log.ToDTO()
}

//根据交易编号查询账户流水
func (a *accountDomain) GetAccountLogByTradeNo(tradeNo string) *services.AccountLogDTO {
	accountLogDao := AccountLogDao{}
	var log *AccountLog
	err := base.Tx(func(runner *dbx.TxRunner) error {
		accountLogDao.runner = runner
		log = accountLogDao.GetByTradeNo(tradeNo)
		return nil
	})
	if err != nil {
		logrus.Error(err)
		return nil
	}
	if log == nil {
		return nil
	}
	return log.ToDTO()
}
