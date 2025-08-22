package service

import (
	"context"
	"fmt"
	"github.com/asynccnu/ccnubox-be/be-grade/crawler"
	"github.com/asynccnu/ccnubox-be/be-grade/events/producer"
	"github.com/asynccnu/ccnubox-be/be-grade/events/topic"
	"golang.org/x/sync/singleflight"
	"time"

	gradev1 "github.com/asynccnu/ccnubox-be/be-api/gen/proto/grade/v1"
	userv1 "github.com/asynccnu/ccnubox-be/be-api/gen/proto/user/v1"
	"github.com/asynccnu/ccnubox-be/be-grade/domain"
	"github.com/asynccnu/ccnubox-be/be-grade/pkg/errorx"
	"github.com/asynccnu/ccnubox-be/be-grade/pkg/logger"
	"github.com/asynccnu/ccnubox-be/be-grade/repository/dao"
	"github.com/asynccnu/ccnubox-be/be-grade/repository/model"
)

var (
	ErrGetGrade = func(err error) error {
		return errorx.New(gradev1.ErrorGetGradeError("获取成绩失败"), "data", err)
	}
)

type GradeService interface {
	GetGradeByTerm(ctx context.Context, req *domain.GetGradeByTermReq) ([]domain.Grade, error)
	GetGradeScore(ctx context.Context, studentId string) ([]domain.TypeOfGradeScore, error)
	GetUpdateScore(ctx context.Context, studentId string) ([]domain.Grade, error)
	UpdateDetailScore(ctx context.Context, need domain.NeedDetailGrade) error
}

type gradeService struct {
	userClient userv1.UserServiceClient
	producer   producer.Producer
	gradeDAO   dao.GradeDAO
	l          logger.Logger
	sf         singleflight.Group
}

func NewGradeService(
	producer producer.Producer,
	gradeDAO dao.GradeDAO,
	l logger.Logger,
	userClient userv1.UserServiceClient,
) GradeService {
	return &gradeService{
		producer:   producer,
		gradeDAO:   gradeDAO,
		l:          l,
		userClient: userClient,
	}
}

func (s *gradeService) GetGradeByTerm(ctx context.Context, req *domain.GetGradeByTermReq) ([]domain.Grade, error) {
	var (
		grades    []model.Grade
		fetchdata FetchGrades
		err       error
	)

	if req.Refresh {

		//如果要求强制更新的话就需要去拉取远程数据
		fetchdata, err = s.fetchGradesWithSingleFlight(ctx, req.StudentID)
		if err != nil || len(fetchdata.final) == 0 {
			s.l.Warn("从ccnu获取成绩失败!", logger.Error(err))
			//拉取失败本地作为兜底
			grades, err = s.gradeDAO.FindGrades(ctx, req.StudentID, 0, 0)
			if err != nil {
				return nil, err
			}
			return modelConvDomainAndFilter(grades, req.Terms, req.Kcxzmcs), nil
		}

		grades = fetchdata.final
		return modelConvDomainAndFilter(grades, req.Terms, req.Kcxzmcs), nil

	} else {

		//如果成功直接返回结果
		grades, err = s.gradeDAO.FindGrades(ctx, req.StudentID, 0, 0)
		if err != nil || len(grades) == 0 {
			//失败尝试从远程拉取
			fetchdata, err = s.fetchGradesWithSingleFlight(ctx, req.StudentID)
			if err != nil {
				s.l.Warn("从ccnu获取成绩失败!", logger.Error(err))
				return nil, ErrGetGrade(err)
			}
			grades = fetchdata.final

			return modelConvDomainAndFilter(grades, req.Terms, req.Kcxzmcs), nil
		}

		//异步更新结果
		go func() {
			fetchdata, err = s.fetchGradesWithSingleFlight(context.Background(), req.StudentID)
			if err != nil {
				s.l.Warn("从ccnu获取成绩失败!", logger.Error(err))
			}
		}()

		return modelConvDomainAndFilter(grades, req.Terms, req.Kcxzmcs), nil
	}
}

func (s *gradeService) GetGradeScore(ctx context.Context, studentId string) ([]domain.TypeOfGradeScore, error) {
	//如果成功直接返回结果
	grades, err := s.gradeDAO.FindGrades(ctx, studentId, 0, 0)
	if err != nil || len(grades) == 0 {
		//失败尝试从远程拉取
		fetchdata, err := s.fetchGradesWithSingleFlight(ctx, studentId)
		if err != nil || len(fetchdata.final) == 0 {
			s.l.Warn("从ccnu获取成绩失败!", logger.Error(err))
			return nil, ErrGetGrade(err)
		}
		return aggregateGradeScore(fetchdata.final), nil
	}

	//异步更新结果
	go func() {
		fetchdata, err := s.fetchGradesWithSingleFlight(context.Background(), studentId)
		if err != nil || len(fetchdata.final) == 0 {
			s.l.Warn("从ccnu获取成绩失败!", logger.Error(err))
		}
	}()

	return aggregateGradeScore(grades), nil
}

func (s *gradeService) GetUpdateScore(ctx context.Context, studentId string) ([]domain.Grade, error) {
	grades, err := s.fetchGradesWithSingleFlight(ctx, studentId)
	if err != nil || len(grades.update) == 0 {
		return nil, ErrGetGrade(err)
	}
	return modelConvDomain(grades.update), nil
}

func (s *gradeService) UpdateDetailScore(ctx context.Context, need domain.NeedDetailGrade) error {
	ug, err := s.newUGWithCookie(ctx, need.StudentID)
	if err != nil {
		return err
	}

	grades := need.Grades
	for i, grade := range grades {
		detail, err := ug.GetDetail(ctx, grade.StudentId, grade.JxbId, grade.KcId, grade.Cj)
		if err == crawler.COOKIE_TIMEOUT {
			ug, err = s.newUGWithCookie(ctx, need.StudentID)
			if err != nil {
				return err
			}
			detail, err = ug.GetDetail(ctx, grade.StudentId, grade.JxbId, grade.KcId, grade.Cj)
		}

		if err != nil {
			s.l.Warn(fmt.Sprintf("获取详细分数失败! 学号:%s,教学班id:%s,课程id:%s,总分:%f", grade.StudentId, grade.JxbId, grade.KcId, grade.Cj), logger.Error(err))
			continue
		}
		grade.RegularGradePercent = detail.Cjxm1bl
		grade.RegularGrade = detail.Cjxm1
		grade.FinalGradePercent = detail.Cjxm3bl
		grade.FinalGrade = detail.Cjxm3
		grades[i] = grade
	}

	_, err = s.gradeDAO.BatchInsertOrUpdate(ctx, grades, true)
	if err != nil {
		return err
	}

	return nil
}

func (s *gradeService) fetchGradesWithSingleFlight(ctx context.Context, studentId string) (FetchGrades, error) {
	key := studentId

	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		start := time.Now()
		ug, err := s.newUGWithCookie(ctx, studentId)
		if err != nil {
			return nil, fmt.Errorf("创建带cookie的ug实例失败:%w", err)
		}

		remote, err := ug.GetGrade(ctx, 0, 0, 300)
		if err == crawler.COOKIE_TIMEOUT {
			ug, err = s.newUGWithCookie(ctx, studentId)
			if err != nil {
				return nil, fmt.Errorf("创建带cookie的ug实例失败:%w", err)
			}
			remote, err = ug.GetGrade(ctx, 0, 0, 300)
			if err != nil {
				return nil, err
			}
		}

		s.l.Info("获取成绩耗时", logger.String("耗时", time.Since(start).String()))
		grades := aggregateGrade(remote)
		// 插入并更新数据,这里不比较详细数据,因为更新会导致问题
		update, err := s.gradeDAO.BatchInsertOrUpdate(ctx, grades, false)
		if err != nil {
			return nil, err
		}

		for _, g := range update {
			s.l.Info("更新成绩成功", logger.String("studentId", g.StudentId), logger.String("课程", g.Kcmc))
		}

		// 读取数据库,已经有的数据要使用已经存在的(因为平时成绩已经获取到了)
		final, err := s.gradeDAO.FindGrades(ctx, studentId, 0, 0)
		if err != nil {
			return nil, err
		}

		var needDetailgrades []model.Grade
		for _, g := range final {
			if g.RegularGradePercent == RegularGradePercentMSG && g.FinalGradePercent == FinalGradePercentMAG {
				needDetailgrades = append(needDetailgrades, g)
			}
		}

		err = s.producer.SendMessage(topic.GradeDetailEvent, domain.NeedDetailGrade{
			StudentID: studentId,
			Grades:    needDetailgrades,
		})
		if err != nil {
			return nil, err
		}

		fetchGrades := FetchGrades{
			update: update,
			final:  final,
		}

		return fetchGrades, err
	})

	fetchGrades, ok := result.(FetchGrades)
	if !ok {
		s.l.Warn("类型断言失败", logger.Error(err))
	}

	return fetchGrades, err

}

func (s *gradeService) newUGWithCookie(ctx context.Context, studentId string) (*crawler.UnderGrad, error) {
	cookieResp, err := s.userClient.GetCookie(ctx, &userv1.GetCookieRequest{StudentId: studentId})
	if err != nil {
		return &crawler.UnderGrad{}, err
	}

	grad, err := crawler.NewUnderGrad(
		crawler.NewCrawlerClientWithCookieJar(
			30*time.Second,
			crawler.NewJarWithCookie(crawler.PG_URL, cookieResp.Cookie),
		),
	)
	if err != nil {
		return &crawler.UnderGrad{}, fmt.Errorf("创建ug爬虫实例失败:%w", err)
	}
	return grad, nil
}
