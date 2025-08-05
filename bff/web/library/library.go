package library

import (
	libraryv1 "github.com/asynccnu/ccnubox-be/be-api/gen/proto/library/v1"
	"github.com/asynccnu/ccnubox-be/bff/errs"
	"github.com/asynccnu/ccnubox-be/bff/pkg/ginx"
	"github.com/asynccnu/ccnubox-be/bff/web"
	"github.com/asynccnu/ccnubox-be/bff/web/ijwt"
	"github.com/gin-gonic/gin"
)

type LibraryHandler struct {
	LibraryClient  libraryv1.LibraryClient // 注入 grpc 服务
	Administrators map[string]struct{}
}

func NewLibraryHandler(client libraryv1.LibraryClient, admins map[string]struct{}) *LibraryHandler {
	return &LibraryHandler{
		LibraryClient:  client,
		Administrators: admins,
	}
}

func (h *LibraryHandler) RegisterRoutes(s *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	sg := s.Group("/library")
	sg.GET("/get_seat", authMiddleware, ginx.WrapClaims(h.GetSeatInfos))
	sg.POST("/reserve_seat", authMiddleware, ginx.WrapClaimsAndReq(h.ReserveSeat))
	sg.GET("/get_seat_records", authMiddleware, ginx.WrapClaims(h.GetSeatRecord))
	sg.GET("/get_history_records", authMiddleware, ginx.WrapClaims(h.GetHistory))
	sg.GET("/get_credit_points", authMiddleware, ginx.WrapClaims(h.GetCreditPoint))
	sg.POST("/get_discussion", authMiddleware, ginx.WrapClaimsAndReq(h.GetDiscussion))
	sg.POST("/search_user", authMiddleware, ginx.WrapClaimsAndReq(h.SearchUser))
	sg.POST("/reserve_discussion", authMiddleware, ginx.WrapClaimsAndReq(h.ReserveDiscussion))
	sg.POST("/cancel_reserve", authMiddleware, ginx.WrapClaimsAndReq(h.CancelReserve))
}

// GetSeatInfos 获取图书馆座位信息
// @Summary 获取图书馆座位信息
// @Description 默认获取当天图书馆座位信息
// @Tags library
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} web.Response{data=GetSeatResponse} "成功返回图书馆座位信息"
// @Failure 500 {object} web.Response "系统异常，获取失败"
// @Router /library/get_seat [get]
func (h *LibraryHandler) GetSeatInfos(ctx *gin.Context, uc ijwt.UserClaims) (web.Response, error) {

	res, err := h.LibraryClient.GetSeat(ctx, &libraryv1.GetSeatRequest{
		StuId: uc.StudentId,
	})

	if err != nil {
		return web.Response{}, errs.GET_SEAT_ERROR(err)
	}

	var roomList []Room

	for _, room := range res.RoomSeats {
		var seatList []Seat

		for _, seat := range room.Seats {
			var timeSlots []TimeSlot
			for _, ts := range seat.Ts {
				timeSlots = append(timeSlots, TimeSlot{
					Start:  ts.Start,
					End:    ts.End,
					State:  ts.State,
					Owner:  ts.Owner,
					Occupy: ts.Occupy,
				})
			}

			seatList = append(seatList, Seat{
				LabName:   seat.LabName,
				KindName:  seat.KindName,
				DevID:     seat.DevId,
				DevName:   seat.DevName,
				TimeSlots: timeSlots,
			})
		}

		roomList = append(roomList, Room{
			RoomID: room.RoomId,
			Seats:  seatList,
		})
	}

	resp := GetSeatResponse{
		Rooms: roomList,
	}

	return web.Response{
		Msg:  "Success",
		Data: resp,
	}, nil
}

// ReserveSeat 预约图书馆座位
// @Summary 预约图书馆座位
// @Description 预约图书馆座位
// @Tags library
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param request body ReserveSeatRequest true "预约座位的请求参数"
// @Success 200 {object} web.Response "成功返回预约成功"
// @Failure 500 {object} web.Response "系统异常，预约失败"
// @Router /library/reserve_seat [post]
func (h *LibraryHandler) ReserveSeat(ctx *gin.Context, req ReserveSeatRequest, uc ijwt.UserClaims) (web.Response, error) {
	_, err := h.LibraryClient.ReserveSeat(ctx, &libraryv1.ReserveSeatRequest{
		DevId: req.DevID,
		Start: req.Start,
		End:   req.End,
		StuId: uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.RESERVE_SEAT_ERROR(err)
	}

	return web.Response{
		Msg: "Success",
	}, nil
}

// GetSeatRecord 获取未来预约
// @Summary 获取未来预约
// @Description 获取即将到来的预约
// @Tags library
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} web.Response{data=GetSeatRecordResponse} "成功返回即将到来的预约"
// @Failure 500 {object} web.Response "系统异常，获取失败"
// @Router /library/get_seat_records [get]
func (h *LibraryHandler) GetSeatRecord(ctx *gin.Context, uc ijwt.UserClaims) (web.Response, error) {
	res, err := h.LibraryClient.GetSeatRecord(ctx, &libraryv1.GetSeatRecordRequest{
		StuId: uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.GET_SEAT_RECORD_ERROR(err)
	}

	var respRecords = make([]Record, 0, len(res.Record))
	for _, record := range res.Record {
		respRecords = append(respRecords, Record{
			ID:       record.Id,
			Owner:    record.Owner,
			Start:    record.Start,
			End:      record.End,
			TimeDesc: record.TimeDesc,
			States:   record.States,
			DevName:  record.DevName,
			RoomID:   record.RoomId,
			RoomName: record.RoomName,
			LabName:  record.LabName,
		})
	}

	resp := GetSeatRecordResponse{
		Records: respRecords,
	}

	return web.Response{
		Msg:  "Success",
		Data: resp,
	}, nil
}

// GetHistory 获取历史预约记录
// @Summary 获取历史预约记录
// @Description 获取1年内的预约记录和三个月内的取消记录
// @Tags library
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} web.Response{data=GetHistoryResponse} "成功返回历史预约记录"
// @Failure 500 {object} web.Response "系统异常，获取失败"
// @Router /library/get_history_records [get]
func (h *LibraryHandler) GetHistory(ctx *gin.Context, uc ijwt.UserClaims) (web.Response, error) {
	res, err := h.LibraryClient.GetHistory(ctx, &libraryv1.GetHistoryRequest{
		StuId: uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.GET_HISTORY_ERROR(err)
	}

	var HistoryRecords = make([]History, 0, len(res.History))
	for _, history := range res.History {
		HistoryRecords = append(HistoryRecords, History{
			Place:      history.Place,
			Floor:      history.Floor,
			Status:     history.Status,
			Date:       history.Date,
			SubmitTime: history.SubmitTime,
		})
	}

	resp := GetHistoryResponse{
		Histories: HistoryRecords,
	}

	return web.Response{
		Msg:  "Success",
		Data: resp,
	}, nil
}

// GetCreditPoint 获取信誉分
// @Summary 获取信誉分
// @Description 获取信誉分及扣分记录
// @Tags library
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} web.Response{data=GetCreditPointResponse} "成功返回信誉分"
// @Failure 500 {object} web.Response "系统异常，获取失败"
// @Router /library/get_credit_points [get]
func (h *LibraryHandler) GetCreditPoint(ctx *gin.Context, uc ijwt.UserClaims) (web.Response, error) {
	res, err := h.LibraryClient.GetCreditPoint(ctx, &libraryv1.GetCreditPointRequest{
		StuId: uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.GET_CREDIT_POINTS_ERROR(err)
	}

	summary := CreditSummary{
		System: res.CreditSummary.System,
		Remain: res.CreditSummary.Remain,
		Total:  res.CreditSummary.Total,
	}

	var records []CreditRecord
	for _, record := range res.CreditRecord {
		records = append(records, CreditRecord{
			Title:    record.Title,
			Subtitle: record.Subtitle,
			Location: record.Location,
		})
	}

	resp := GetCreditPointResponse{
		CreditPoints: CreditPoints{
			Summary: summary,
			Records: records,
		},
	}

	return web.Response{
		Msg:  "Success",
		Data: resp,
	}, nil
}

// GetDiscussion 获取图书馆研讨间信息
// @Summary 获取图书馆研讨间信息
// @Description 传入时间获取图书馆研讨间信息
// @Tags library
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param request body GetDiscussionRequest true "获取研讨间信息的请求参数"
// @Success 200 {object} web.Response{data=GetDiscussionResponse} "成功返回图书馆研讨间信息"
// @Failure 500 {object} web.Response "系统异常，获取失败"
// @Router /library/get_discussion [post]
func (h *LibraryHandler) GetDiscussion(ctx *gin.Context, req GetDiscussionRequest, uc ijwt.UserClaims) (web.Response, error) {
	res, err := h.LibraryClient.GetDiscussion(ctx, &libraryv1.GetDiscussionRequest{
		ClassId: req.ClassID,
		Date:    req.Date,
		StuId:   uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.GET_DISCUSSION_ERROR(err)
	}

	var discussions []Discussion
	for _, d := range res.Discussions {
		var ts []DiscussionTS
		for _, t := range d.TS {
			ts = append(ts, DiscussionTS{
				Start:  t.Start,
				End:    t.End,
				State:  t.State,
				Title:  t.Title,
				Owner:  t.Owner,
				Occupy: t.Occupy,
			})
		}
		discussions = append(discussions, Discussion{
			LabID:    d.LabId,
			LabName:  d.LabName,
			KindID:   d.KindId,
			KindName: d.KindName,
			DevID:    d.DevId,
			DevName:  d.DevName,
			TS:       ts,
		})
	}

	resp := GetDiscussionResponse{
		Discussions: discussions,
	}

	return web.Response{
		Msg:  "Success",
		Data: resp,
	}, nil
}

// SearchUser 搜索学生ID
// @Summary 搜索学生ID
// @Description 传入学生学号获取对应的学生ID
// @Tags library
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param request query SearchUserRequest true "搜索学生ID的请求参数"
// @Success 200 {object} web.Response{data=SearchUserResponse} "成功返回学生的ID"
// @Failure 500 {object} web.Response "系统异常，获取失败"
// @Router /library/search_user [post]
func (h *LibraryHandler) SearchUser(ctx *gin.Context, req SearchUserRequest, uc ijwt.UserClaims) (web.Response, error) {
	res, err := h.LibraryClient.SearchUser(ctx, &libraryv1.SearchUserRequest{
		StudentId: req.StudentID,
		StuId:     uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.SEARCH_USER_ERROR(err)
	}

	resp := SearchUserResponse{
		Search: Search{
			ID:    res.Id,
			Pid:   res.Pid,
			Name:  res.Name,
			Label: res.Label,
		},
	}

	return web.Response{
		Msg:  "Success",
		Data: resp,
	}, nil
}

// ReserveDiscussion 预约研讨间
// @Summary 预约研讨间
// @Description 传入学生ID,时间,主题等预约研讨间
// @Tags library
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param request body ReserveDiscussionRequest true "预约研讨间所需要的参数"
// @Success 200 {object} web.Response "成功返回预约研讨间成功"
// @Failure 500 {object} web.Response "系统异常，获取失败"
// @Router /library/reserve_discussion [post]
func (h *LibraryHandler) ReserveDiscussion(ctx *gin.Context, req ReserveDiscussionRequest, uc ijwt.UserClaims) (web.Response, error) {
	_, err := h.LibraryClient.ReserveDiscussion(ctx, &libraryv1.ReserveDiscussionRequest{
		DevId:  req.DevID,
		LabId:  req.LabID,
		KindId: req.KindID,
		Title:  req.Title,
		Start:  req.Start,
		End:    req.End,
		List:   req.List,
		StuId:  uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.RESERVE_DISCUSSION_ERROR(err)
	}

	return web.Response{
		Msg: "Success",
	}, nil
}

// CancelReserve 取消预约
// @Summary 取消预约
// @Description 取消预约
// @Tags library
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param request query CancelReserveRequest true "取消预约所需要的参数"
// @Success 200 {object} web.Response "成功返回取消预约成功"
// @Failure 500 {object} web.Response "系统异常，获取失败"
// @Router /library/cancel_reserve [post]
func (h *LibraryHandler) CancelReserve(ctx *gin.Context, req CancelReserveRequest, uc ijwt.UserClaims) (web.Response, error) {
	_, err := h.LibraryClient.CancelReserve(ctx, &libraryv1.CancelReserveRequest{
		Id:    req.ID,
		StuId: uc.StudentId,
	})
	if err != nil {
		return web.Response{}, errs.CANCEL_DISCUSSION_ERROR(err)
	}

	return web.Response{
		Msg: "Success",
	}, nil
}
