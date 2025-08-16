package service

import (
	"context"
	"crypto/rsa"
	"errors"
	ccnuv1 "github.com/asynccnu/ccnubox-be/be-api/gen/proto/ccnu/v1"
	"github.com/asynccnu/ccnubox-be/be-ccnu/crawler"
	"github.com/asynccnu/ccnubox-be/be-ccnu/pkg/errorx"
	"github.com/asynccnu/ccnubox-be/be-ccnu/tool"
)

// 定义错误,这里将kratos的error作为一个重要部分传入,此处的错误并不直接在service中去捕获,而是选择在更底层的爬虫去捕获,因为爬虫的错误处理非常复杂
var (
	CCNUSERVER_ERROR = func(err error) error {
		return errorx.New(ccnuv1.ErrorCcnuserverError("ccnu服务器错误"), "ccnuServer", err)
	}

	Invalid_SidOrPwd_ERROR = func(err error) error {
		return errorx.New(ccnuv1.ErrorInvalidSidOrPwd("账号密码错误"), "user", err)
	}

	SYSTEM_ERROR = func(err error) error {
		return errorx.New(ccnuv1.ErrorSystemError("系统内部错误"), "system", err)
	}
)

func (c *ccnuService) GetXKCookie(ctx context.Context, studentId string, password string) (string, error) {

	//初始化client
	var (
		ug                  = crawler.NewUnderGrad(crawler.NewCrawlerClient(c.timeout))
		isInCorrectPASSWORD = false //用于判断是否是账号密码错误
	)

	params, err := tool.Retry(func() (*crawler.AccountRequestParams, error) {
		return ug.GetParamsFromHtml(ctx)
	})
	if err != nil {
		return "", err
	}

	//此处比较特殊由于账号密码错误是必然无效的请求,应当直接返回
	_, err = tool.Retry(func() (string, error) {
		err := ug.LoginCCNUPassport(ctx, studentId, password, params)
		if errors.Is(err, crawler.INCorrectPASSWORD) {
			// 标识账号密码错误,强制结束
			isInCorrectPASSWORD = true
			return "", nil
		}
		return "", err
	})
	//如果密码有误
	if isInCorrectPASSWORD {
		return "", Invalid_SidOrPwd_ERROR(errors.New("账号密码错误"))
	}
	//如果存在错误
	if err != nil {
		return "", err
	}

	_, err = tool.Retry(func() (string, error) {
		err := ug.LoginUnderGradSystem(ctx)
		if err != nil {
			return "", err
		}
		return "", nil
	})
	if err != nil {
		return "", err
	}

	cookie, err := ug.GetCookieFromUnderGradSystem()
	if err != nil {
		return "", err
	}

	return cookie, nil
}

func (c *ccnuService) LoginCCNU(ctx context.Context, studentId string, password string) (bool, error) {
	// TODO 抽象成函数
	if len(studentId) > 4 && studentId[4] == '1' {
		//研究生
		var (
			pg                  = crawler.NewPostGraduate(crawler.NewCrawlerClient(c.timeout))
			isInCorrectPASSWORD = false //用于判断是否是账号密码错误

		)
		pubkey, err := tool.Retry(func() (*rsa.PublicKey, error) {
			return pg.FetchPublicKey(ctx)
		})
		if err != nil {
			return false, err
		}

		_, err = tool.Retry(func() (string, error) {
			err := pg.LoginPostgraduateSystem(ctx, studentId, password, pubkey)
			if errors.Is(err, crawler.INCorrectPASSWORD) {
				// 标识账号密码错误,强制结束
				isInCorrectPASSWORD = true
				return "", nil
			}
			return "", err
		})
		//如果密码有误
		if isInCorrectPASSWORD {
			return false, Invalid_SidOrPwd_ERROR(errors.New("账号密码错误"))
		}
		//如果存在错误
		if err != nil {
			return false, err
		}
		return true, nil

	} else if len(studentId) > 4 && studentId[4] == '2' {
		//本科生
		var (
			ug                  = crawler.NewUnderGrad(crawler.NewCrawlerClient(c.timeout))
			isInCorrectPASSWORD = false //用于判断是否是账号密码错误
		)

		params, err := tool.Retry(func() (*crawler.AccountRequestParams, error) {
			return ug.GetParamsFromHtml(ctx)
		})
		if err != nil {
			return false, err
		}

		//此处比较特殊由于账号密码错误是必然无效的请求,应当直接返回
		_, err = tool.Retry(func() (string, error) {
			err := ug.LoginCCNUPassport(ctx, studentId, password, params)
			if errors.Is(err, crawler.INCorrectPASSWORD) {
				// 标识账号密码错误,强制结束
				isInCorrectPASSWORD = true
				return "", nil
			}
			return "", err
		})
		//如果密码有误
		if isInCorrectPASSWORD {
			return false, Invalid_SidOrPwd_ERROR(errors.New("账号密码错误"))
		}
		//如果存在错误
		if err != nil {
			return false, err
		}
		return true, nil

	} else {
		return false, Invalid_SidOrPwd_ERROR(errors.New("账号密码错误"))
	}

}
